package org.trainblink.app.services

import android.annotation.SuppressLint
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothGatt
import android.bluetooth.BluetoothGattCallback
import android.bluetooth.BluetoothGattCharacteristic
import android.bluetooth.BluetoothGattServer
import android.bluetooth.BluetoothGattServerCallback
import android.bluetooth.BluetoothGattService
import android.bluetooth.BluetoothManager
import android.bluetooth.BluetoothProfile
import android.bluetooth.le.AdvertiseCallback
import android.bluetooth.le.AdvertiseData
import android.bluetooth.le.AdvertiseSettings
import android.bluetooth.le.BluetoothLeAdvertiser
import android.bluetooth.le.BluetoothLeScanner
import android.bluetooth.le.ScanCallback
import android.bluetooth.le.ScanFilter
import android.bluetooth.le.ScanResult
import android.bluetooth.le.ScanSettings
import android.content.Intent
import android.os.Binder
import android.os.Build
import android.os.Handler
import android.os.IBinder
import android.os.Looper
import android.os.ParcelUuid
import android.util.Log
import androidx.core.app.NotificationCompat
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import org.trainblink.app.models.DiscoveredUser
import org.trainblink.app.models.DistanceEstimate
import org.trainblink.app.models.LogDiscoveryRequest
import org.trainblink.app.models.Trip
import org.trainblink.app.network.ApiService
import java.util.Date
import java.util.UUID

/**
 * BLE Discovery Service - Foreground service for BLE broadcasting and scanning
 */
class BLEDiscoveryService : Service() {

    companion object {
        private const val TAG = "BLEDiscoveryService"
        private const val NOTIFICATION_CHANNEL_ID = "trainblink_discovery"
        private const val NOTIFICATION_ID = 1001

        // TrainBlink BLE Service UUID (matches backend)
        private val SERVICE_UUID = UUID.fromString("TB000000-0000-1000-8000-00805F9B34FB")
        private val CHARACTERISTIC_UUID = UUID.fromString("TB000001-0000-1000-8000-00805F9B34FB")

        // Cleanup interval
        private const val CLEANUP_INTERVAL_MS = 10_000L // 10 seconds
        private const val STALE_THRESHOLD_MS = 30_000L // 30 seconds
    }

    // Binder for activity communication
    private val binder = LocalBinder()

    inner class LocalBinder : Binder() {
        fun getService(): BLEDiscoveryService = this@BLEDiscoveryService
    }

    // Bluetooth components
    private lateinit var bluetoothManager: BluetoothManager
    private lateinit var bluetoothAdapter: BluetoothAdapter
    private var bluetoothLeScanner: BluetoothLeScanner? = null
    private var bluetoothLeAdvertiser: BluetoothLeAdvertiser? = null
    private var gattServer: BluetoothGattServer? = null

    // Current trip
    private var currentTrip: Trip? = null

    // Discovered users
    private val _discoveredUsers = MutableStateFlow<Map<String, DiscoveredUser>>(emptyMap())
    val discoveredUsers: StateFlow<Map<String, DiscoveredUser>> = _discoveredUsers.asStateFlow()

    // State
    private val _isScanning = MutableStateFlow(false)
    val isScanning: StateFlow<Boolean> = _isScanning.asStateFlow()

    private val _isAdvertising = MutableStateFlow(false)
    val isAdvertising: StateFlow<Boolean> = _isAdvertising.asStateFlow()

    // Coroutine scope
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    // Handlers
    private val mainHandler = Handler(Looper.getMainLooper())
    private val cleanupRunnable = object : Runnable {
        override fun run() {
            cleanupStaleDiscoveries()
            mainHandler.postDelayed(this, CLEANUP_INTERVAL_MS)
        }
    }

    // API service
    private lateinit var apiService: ApiService

    // MARK: - Lifecycle

    override fun onCreate() {
        super.onCreate()
        Log.d(TAG, "Service created")

        // Initialize Bluetooth
        bluetoothManager = getSystemService(BLUETOOTH_SERVICE) as BluetoothManager
        bluetoothAdapter = bluetoothManager.adapter
        bluetoothLeScanner = bluetoothAdapter.bluetoothLeScanner
        bluetoothLeAdvertiser = bluetoothAdapter.bluetoothLeAdvertiser

        // Initialize API service
        apiService = ApiService.create(applicationContext)

        // Start cleanup timer
        mainHandler.post(cleanupRunnable)
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.d(TAG, "Service started")

        // Create foreground notification
        startForeground(NOTIFICATION_ID, createNotification())

        return START_STICKY
    }

    override fun onBind(intent: Intent?): IBinder {
        return binder
    }

    override fun onDestroy() {
        Log.d(TAG, "Service destroyed")
        stopDiscovery()
        mainHandler.removeCallbacks(cleanupRunnable)
        serviceScope.cancel()
        super.onDestroy()
    }

    // MARK: - Public Methods

    /**
     * Start BLE discovery for a trip
     */
    @SuppressLint("MissingPermission")
    fun startDiscovery(trip: Trip) {
        Log.d(TAG, "Starting discovery for trip: ${trip.route}")
        currentTrip = trip

        // Start advertising our anonymous ID
        startAdvertising(trip.bleAnonymousId)

        // Start scanning for nearby users
        startScanning()

        // Update notification
        updateNotification("Discovering nearby travelers...")
    }

    /**
     * Stop BLE discovery
     */
    fun stopDiscovery() {
        Log.d(TAG, "Stopping discovery")
        stopAdvertising()
        stopScanning()
        _discoveredUsers.value = emptyMap()
        currentTrip = null
        updateNotification("Discovery stopped")
    }

    /**
     * Get fresh discoveries (within last 30 seconds)
     */
    fun getFreshDiscoveries(): List<DiscoveredUser> {
        return _discoveredUsers.value.values.filter { it.isFresh }.sortedByDescending { it.rssi }
    }

    // MARK: - Advertising (Broadcasting)

    @SuppressLint("MissingPermission")
    private fun startAdvertising(anonymousId: String) {
        if (bluetoothLeAdvertiser == null) {
            Log.e(TAG, "BLE advertiser not available")
            return
        }

        // Create GATT server
        val gattServer = bluetoothManager.openGattServer(this, gattServerCallback)
        this.gattServer = gattServer

        // Create characteristic with anonymous ID
        val characteristic = BluetoothGattCharacteristic(
            CHARACTERISTIC_UUID,
            BluetoothGattCharacteristic.PROPERTY_READ,
            BluetoothGattCharacteristic.PERMISSION_READ
        ).apply {
            value = anonymousId.toByteArray(Charsets.UTF_8)
        }

        // Create service
        val service = BluetoothGattService(SERVICE_UUID, BluetoothGattService.SERVICE_TYPE_PRIMARY).apply {
            addCharacteristic(characteristic)
        }

        // Add service to GATT server
        gattServer.addService(service)

        // Configure advertising settings
        val settings = AdvertiseSettings.Builder()
            .setAdvertiseMode(AdvertiseSettings.ADVERTISE_MODE_LOW_LATENCY)
            .setTxPowerLevel(AdvertiseSettings.ADVERTISE_TX_POWER_HIGH)
            .setConnectable(true)
            .build()

        // Configure advertising data
        val data = AdvertiseData.Builder()
            .setIncludeDeviceName(false)
            .addServiceUuid(ParcelUuid(SERVICE_UUID))
            .build()

        // Start advertising
        bluetoothLeAdvertiser?.startAdvertising(settings, data, advertiseCallback)

        Log.d(TAG, "Advertising started with ID: $anonymousId")
    }

    @SuppressLint("MissingPermission")
    private fun stopAdvertising() {
        bluetoothLeAdvertiser?.stopAdvertising(advertiseCallback)
        gattServer?.close()
        gattServer = null
        _isAdvertising.value = false
        Log.d(TAG, "Advertising stopped")
    }

    private val advertiseCallback = object : AdvertiseCallback() {
        override fun onStartSuccess(settingsInEffect: AdvertiseSettings?) {
            Log.d(TAG, "Advertising started successfully")
            _isAdvertising.value = true
        }

        override fun onStartFailure(errorCode: Int) {
            Log.e(TAG, "Advertising failed: $errorCode")
            _isAdvertising.value = false
        }
    }

    private val gattServerCallback = object : BluetoothGattServerCallback() {
        @SuppressLint("MissingPermission")
        override fun onCharacteristicReadRequest(
            device: BluetoothDevice?,
            requestId: Int,
            offset: Int,
            characteristic: BluetoothGattCharacteristic?
        ) {
            if (characteristic?.uuid == CHARACTERISTIC_UUID) {
                gattServer?.sendResponse(
                    device,
                    requestId,
                    BluetoothGatt.GATT_SUCCESS,
                    offset,
                    characteristic.value
                )
            } else {
                gattServer?.sendResponse(
                    device,
                    requestId,
                    BluetoothGatt.GATT_FAILURE,
                    offset,
                    null
                )
            }
        }
    }

    // MARK: - Scanning

    @SuppressLint("MissingPermission")
    private fun startScanning() {
        if (bluetoothLeScanner == null) {
            Log.e(TAG, "BLE scanner not available")
            return
        }

        // Configure scan filter
        val scanFilter = ScanFilter.Builder()
            .setServiceUuid(ParcelUuid(SERVICE_UUID))
            .build()

        // Configure scan settings
        val scanSettings = ScanSettings.Builder()
            .setScanMode(ScanSettings.SCAN_MODE_LOW_LATENCY)
            .build()

        // Start scanning
        bluetoothLeScanner?.startScan(listOf(scanFilter), scanSettings, scanCallback)
        _isScanning.value = true

        Log.d(TAG, "Scanning started")
    }

    @SuppressLint("MissingPermission")
    private fun stopScanning() {
        bluetoothLeScanner?.stopScan(scanCallback)
        _isScanning.value = false
        Log.d(TAG, "Scanning stopped")
    }

    private val scanCallback = object : ScanCallback() {
        override fun onScanResult(callbackType: Int, result: ScanResult?) {
            result?.let { handleScanResult(it) }
        }

        override fun onBatchScanResults(results: MutableList<ScanResult>?) {
            results?.forEach { handleScanResult(it) }
        }

        override fun onScanFailed(errorCode: Int) {
            Log.e(TAG, "Scan failed: $errorCode")
            _isScanning.value = false
        }
    }

    @SuppressLint("MissingPermission")
    private fun handleScanResult(result: ScanResult) {
        val rssi = result.rssi

        // Filter out weak signals (beyond 100m)
        if (rssi < -100) return

        val device = result.device
        Log.d(TAG, "Discovered device: ${device.address}, RSSI: $rssi")

        // Connect to read anonymous ID
        device.connectGatt(this, false, object : BluetoothGattCallback() {
            override fun onConnectionStateChange(gatt: BluetoothGatt?, status: Int, newState: Int) {
                if (newState == BluetoothProfile.STATE_CONNECTED) {
                    gatt?.discoverServices()
                } else if (newState == BluetoothProfile.STATE_DISCONNECTED) {
                    gatt?.close()
                }
            }

            override fun onServicesDiscovered(gatt: BluetoothGatt?, status: Int) {
                if (status == BluetoothGatt.GATT_SUCCESS) {
                    val service = gatt?.getService(SERVICE_UUID)
                    val characteristic = service?.getCharacteristic(CHARACTERISTIC_UUID)
                    characteristic?.let {
                        gatt.readCharacteristic(it)
                    }
                }
            }

            override fun onCharacteristicRead(
                gatt: BluetoothGatt?,
                characteristic: BluetoothGattCharacteristic?,
                status: Int
            ) {
                if (status == BluetoothGatt.GATT_SUCCESS && characteristic?.uuid == CHARACTERISTIC_UUID) {
                    val anonymousId = characteristic.value?.toString(Charsets.UTF_8)
                    anonymousId?.let {
                        processDiscovery(it, rssi)
                    }
                }
                gatt?.disconnect()
            }
        })
    }

    // MARK: - Discovery Processing

    private fun processDiscovery(anonymousId: String, rssi: Int) {
        val trip = currentTrip ?: return

        // Don't discover ourselves
        if (anonymousId == trip.bleAnonymousId) return

        val distance = DistanceEstimate.fromRssi(rssi)
        val now = Date()

        Log.d(TAG, "Processed discovery: $anonymousId, RSSI: $rssi, Distance: ${distance.shortDescription}")

        // Update or create discovered user
        val updatedUsers = _discoveredUsers.value.toMutableMap()

        if (updatedUsers.containsKey(anonymousId)) {
            updatedUsers[anonymousId]?.update(rssi)
        } else {
            val discovered = DiscoveredUser(
                id = anonymousId,
                rssi = rssi,
                distance = distance,
                discoveredAt = now,
                lastSeen = now
            )
            updatedUsers[anonymousId] = discovered

            // Log to backend analytics (async)
            logDiscoveryToBackend(anonymousId, rssi, distance, trip.route)
        }

        _discoveredUsers.value = updatedUsers
    }

    private fun cleanupStaleDiscoveries() {
        val now = Date()
        val updatedUsers = _discoveredUsers.value.filterValues {
            (now.time - it.lastSeen.time) < STALE_THRESHOLD_MS
        }

        if (updatedUsers.size != _discoveredUsers.value.size) {
            _discoveredUsers.value = updatedUsers
            Log.d(TAG, "Cleaned up ${_discoveredUsers.value.size - updatedUsers.size} stale discoveries")
        }
    }

    // MARK: - Analytics

    private fun logDiscoveryToBackend(anonymousId: String, rssi: Int, distance: DistanceEstimate, route: String) {
        val request = LogDiscoveryRequest(
            tripRoute = route,
            discoveredUserAnonymousId = anonymousId,
            distanceEstimate = distance.rawValue,
            rssi = rssi
        )

        serviceScope.launch {
            try {
                apiService.logDiscovery(request)
                Log.d(TAG, "Discovery logged to backend: $anonymousId")
            } catch (e: Exception) {
                Log.e(TAG, "Failed to log discovery: ${e.message}", e)
            }
        }
    }

    // MARK: - Notification

    private fun createNotification(): Notification {
        createNotificationChannel()

        val intent = packageManager.getLaunchIntentForPackage(packageName)
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            intent,
            PendingIntent.FLAG_IMMUTABLE
        )

        return NotificationCompat.Builder(this, NOTIFICATION_CHANNEL_ID)
            .setContentTitle("TrainBlink Discovery")
            .setContentText("Searching for nearby travelers...")
            .setSmallIcon(android.R.drawable.ic_dialog_info)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .build()
    }

    private fun updateNotification(text: String) {
        val notification = NotificationCompat.Builder(this, NOTIFICATION_CHANNEL_ID)
            .setContentTitle("TrainBlink Discovery")
            .setContentText(text)
            .setSmallIcon(android.R.drawable.ic_dialog_info)
            .setOngoing(true)
            .build()

        val notificationManager = getSystemService(NOTIFICATION_SERVICE) as NotificationManager
        notificationManager.notify(NOTIFICATION_ID, notification)
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                NOTIFICATION_CHANNEL_ID,
                "Discovery Service",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "TrainBlink BLE discovery service"
            }

            val notificationManager = getSystemService(NOTIFICATION_SERVICE) as NotificationManager
            notificationManager.createNotificationChannel(channel)
        }
    }
}
