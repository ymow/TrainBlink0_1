//
//  BLEDiscoveryManager.swift
//  TrainBlink
//
//  Manages BLE broadcasting and scanning for nearby travelers
//

import Foundation
import CoreBluetooth
import Combine

/// BLE Discovery Manager for proximity-based user discovery
class BLEDiscoveryManager: NSObject, ObservableObject {

    // MARK: - Published Properties

    @Published var discoveredUsers: [String: DiscoveredUser] = [:]
    @Published var isScanning: Bool = false
    @Published var isBroadcasting: Bool = false
    @Published var discoveryEnabled: Bool = true

    // MARK: - Private Properties

    private var centralManager: CBCentralManager!
    private var peripheralManager: CBPeripheralManager!

    private var currentTrip: Trip?
    private var discoveryTimer: Timer?
    private var cleanupTimer: Timer?

    // TrainBlink BLE Service UUID (matches backend)
    private let serviceUUID = CBUUID(string: "TB000000-0000-1000-8000-00805F9B34FB")
    private let characteristicUUID = CBUUID(string: "TB000001-0000-1000-8000-00805F9B34FB")

    // Broadcasting characteristic
    private var advertisingCharacteristic: CBMutableCharacteristic?

    // Analytics service
    private let apiService: APIService

    // Callback for discovery events
    var onDiscoveryEvent: ((DiscoveredUser) -> Void)?

    // Store RSSI values from didDiscover callback (keyed by peripheral UUID)
    private var peripheralRSSI: [UUID: Int] = [:]

    // MARK: - Initialization

    init(apiService: APIService) {
        self.apiService = apiService
        super.init()

        // Initialize Bluetooth managers
        centralManager = CBCentralManager(delegate: self, queue: nil)
        peripheralManager = CBPeripheralManager(delegate: self, queue: nil)

        // Setup cleanup timer (remove stale discoveries every 10 seconds)
        cleanupTimer = Timer.scheduledTimer(withTimeInterval: 10.0, repeats: true) { [weak self] _ in
            self?.cleanupStaleDiscoveries()
        }
    }

    deinit {
        stopDiscovery()
        cleanupTimer?.invalidate()
        discoveryTimer?.invalidate()
    }

    // MARK: - Public Methods

    /// Start BLE discovery for a trip
    func startDiscovery(for trip: Trip) {
        guard discoveryEnabled else {
            print("[BLE] Discovery disabled")
            return
        }

        currentTrip = trip

        // Start broadcasting our anonymous ID
        startBroadcasting(anonymousId: trip.bleAnonymousId)

        // Start scanning for nearby users
        startScanning()

        print("[BLE] Discovery started for trip: \(trip.route)")
    }

    /// Stop BLE discovery
    func stopDiscovery() {
        stopBroadcasting()
        stopScanning()
        discoveredUsers.removeAll()
        currentTrip = nil

        print("[BLE] Discovery stopped")
    }

    /// Toggle discovery enabled state
    func setDiscoveryEnabled(_ enabled: Bool) {
        discoveryEnabled = enabled

        if !enabled {
            stopDiscovery()
        } else if let trip = currentTrip {
            startDiscovery(for: trip)
        }
    }

    /// Get fresh discoveries (within last 30 seconds)
    func getFreshDiscoveries() -> [DiscoveredUser] {
        return discoveredUsers.values.filter { $0.isFresh }.sorted { $0.rssi > $1.rssi }
    }

    // MARK: - Private Methods - Broadcasting

    private func startBroadcasting(anonymousId: String) {
        guard peripheralManager.state == .poweredOn else {
            print("[BLE] Peripheral manager not ready")
            return
        }

        // Create characteristic with anonymous ID
        let characteristic = CBMutableCharacteristic(
            type: characteristicUUID,
            properties: [.read, .notify],
            value: anonymousId.data(using: .utf8),
            permissions: [.readable]
        )
        advertisingCharacteristic = characteristic

        // Create service
        let service = CBMutableService(type: serviceUUID, primary: true)
        service.characteristics = [characteristic]

        // Add service to peripheral manager
        peripheralManager.removeAllServices()
        peripheralManager.add(service)

        // Start advertising
        peripheralManager.startAdvertising([
            CBAdvertisementDataServiceUUIDsKey: [serviceUUID],
            CBAdvertisementDataLocalNameKey: "TrainBlink"
        ])

        isBroadcasting = true
        print("[BLE] Broadcasting started with ID: \(anonymousId)")
    }

    private func stopBroadcasting() {
        peripheralManager.stopAdvertising()
        peripheralManager.removeAllServices()
        isBroadcasting = false
        print("[BLE] Broadcasting stopped")
    }

    // MARK: - Private Methods - Scanning

    private func startScanning() {
        guard centralManager.state == .poweredOn else {
            print("[BLE] Central manager not ready")
            return
        }

        centralManager.scanForPeripherals(
            withServices: [serviceUUID],
            options: [CBCentralManagerScanOptionAllowDuplicatesKey: true]
        )

        isScanning = true
        print("[BLE] Scanning started")
    }

    private func stopScanning() {
        centralManager.stopScan()
        isScanning = false
        print("[BLE] Scanning stopped")
    }

    private func handleDiscoveredPeripheral(_ peripheral: CBPeripheral, rssi: Int, advertisementData: [String: Any]) {
        // Connect to read anonymous ID
        centralManager.connect(peripheral, options: nil)
    }

    private func processDiscovery(anonymousId: String, rssi: Int) {
        guard let trip = currentTrip else { return }

        // Don't discover ourselves
        guard anonymousId != trip.bleAnonymousId else { return }

        let distance = DistanceEstimate(rssi: rssi)
        let now = Date()

        // Update or create discovered user
        if var existing = discoveredUsers[anonymousId] {
            existing.rssi = rssi
            existing.distance = distance
            existing.lastSeen = now
            discoveredUsers[anonymousId] = existing
        } else {
            let discovered = DiscoveredUser(
                id: anonymousId,
                rssi: rssi,
                distance: distance,
                discoveredAt: now,
                lastSeen: now
            )
            discoveredUsers[anonymousId] = discovered

            // Trigger callback for new discovery
            onDiscoveryEvent?(discovered)

            // Log to backend analytics (async, fire and forget)
            logDiscoveryToBackend(anonymousId: anonymousId, rssi: rssi, distance: distance, route: trip.route)
        }
    }

    private func cleanupStaleDiscoveries() {
        let staleThreshold: TimeInterval = 30 // 30 seconds
        let now = Date()

        discoveredUsers = discoveredUsers.filter { _, user in
            now.timeIntervalSince(user.lastSeen) < staleThreshold
        }
    }

    // MARK: - Analytics

    private func logDiscoveryToBackend(anonymousId: String, rssi: Int, distance: DistanceEstimate, route: String) {
        let request = LogDiscoveryRequest(
            tripRoute: route,
            discoveredUserAnonymousId: anonymousId,
            distanceEstimate: distance.rawValue,
            rssi: rssi,
            discovererAgeRange: nil,
            discoveredAgeRange: nil,
            discovererGender: nil,
            discoveredGender: nil
        )

        Task {
            do {
                try await apiService.logDiscovery(request: request)
                print("[BLE] Discovery logged to backend: \(anonymousId)")
            } catch {
                print("[BLE] Failed to log discovery: \(error.localizedDescription)")
            }
        }
    }
}

// MARK: - CBCentralManagerDelegate

extension BLEDiscoveryManager: CBCentralManagerDelegate {

    func centralManagerDidUpdateState(_ central: CBCentralManager) {
        switch central.state {
        case .poweredOn:
            print("[BLE] Central manager powered on")
            if let trip = currentTrip, discoveryEnabled {
                startScanning()
            }
        case .poweredOff:
            print("[BLE] Central manager powered off")
            isScanning = false
        case .unauthorized:
            print("[BLE] Bluetooth unauthorized")
        case .unsupported:
            print("[BLE] Bluetooth unsupported")
        default:
            print("[BLE] Central manager state: \(central.state.rawValue)")
        }
    }

    func centralManager(_ central: CBCentralManager, didDiscover peripheral: CBPeripheral, advertisementData: [String: Any], rssi RSSI: NSNumber) {
        let rssiValue = RSSI.intValue

        // Filter out weak signals (beyond 100m)
        guard rssiValue > -100 else { return }

        // Store RSSI value for this peripheral (will be used when characteristic is read)
        peripheralRSSI[peripheral.identifier] = rssiValue

        // Store peripheral for connection
        peripheral.delegate = self
        central.connect(peripheral, options: nil)
    }

    func centralManager(_ central: CBCentralManager, didConnect peripheral: CBPeripheral) {
        print("[BLE] Connected to peripheral: \(peripheral.identifier)")
        peripheral.discoverServices([serviceUUID])
    }

    func centralManager(_ central: CBCentralManager, didFailToConnect peripheral: CBPeripheral, error: Error?) {
        print("[BLE] Failed to connect to peripheral: \(error?.localizedDescription ?? "unknown")")
    }

    func centralManager(_ central: CBCentralManager, didDisconnectPeripheral peripheral: CBPeripheral, error: Error?) {
        // Disconnection is normal after reading characteristic
    }
}

// MARK: - CBPeripheralDelegate

extension BLEDiscoveryManager: CBPeripheralDelegate {

    func peripheral(_ peripheral: CBPeripheral, didDiscoverServices error: Error?) {
        guard let services = peripheral.services else { return }

        for service in services {
            if service.uuid == serviceUUID {
                peripheral.discoverCharacteristics([characteristicUUID], for: service)
            }
        }
    }

    func peripheral(_ peripheral: CBPeripheral, didDiscoverCharacteristicsFor service: CBService, error: Error?) {
        guard let characteristics = service.characteristics else { return }

        for characteristic in characteristics {
            if characteristic.uuid == characteristicUUID {
                peripheral.readValue(for: characteristic)
            }
        }
    }

    func peripheral(_ peripheral: CBPeripheral, didUpdateValueFor characteristic: CBCharacteristic, error: Error?) {
        guard let data = characteristic.value,
              let anonymousId = String(data: data, encoding: .utf8) else {
            return
        }

        // Retrieve RSSI value stored from didDiscover callback
        guard let rssi = peripheralRSSI[peripheral.identifier] else {
            print("[BLE] Warning: No RSSI value found for peripheral \(peripheral.identifier)")
            return
        }

        // Process the discovery
        processDiscovery(anonymousId: anonymousId, rssi: rssi)

        // Clean up RSSI value after use
        peripheralRSSI.removeValue(forKey: peripheral.identifier)

        // Disconnect after reading
        centralManager.cancelPeripheralConnection(peripheral)
    }
}

// MARK: - CBPeripheralManagerDelegate

extension BLEDiscoveryManager: CBPeripheralManagerDelegate {

    func peripheralManagerDidUpdateState(_ peripheral: CBPeripheralManager) {
        switch peripheral.state {
        case .poweredOn:
            print("[BLE] Peripheral manager powered on")
            if let trip = currentTrip, discoveryEnabled {
                startBroadcasting(anonymousId: trip.bleAnonymousId)
            }
        case .poweredOff:
            print("[BLE] Peripheral manager powered off")
            isBroadcasting = false
        case .unauthorized:
            print("[BLE] Bluetooth peripheral unauthorized")
        case .unsupported:
            print("[BLE] Bluetooth peripheral unsupported")
        default:
            print("[BLE] Peripheral manager state: \(peripheral.state.rawValue)")
        }
    }

    func peripheralManager(_ peripheral: CBPeripheralManager, didAdd service: CBService, error: Error?) {
        if let error = error {
            print("[BLE] Failed to add service: \(error.localizedDescription)")
        } else {
            print("[BLE] Service added successfully")
        }
    }

    func peripheralManagerDidStartAdvertising(_ peripheral: CBPeripheralManager, error: Error?) {
        if let error = error {
            print("[BLE] Failed to start advertising: \(error.localizedDescription)")
            isBroadcasting = false
        } else {
            print("[BLE] Advertising started successfully")
            isBroadcasting = true
        }
    }

    func peripheralManager(_ peripheral: CBPeripheralManager, didReceiveRead request: CBATTRequest) {
        // Respond to read requests for our characteristic
        if request.characteristic.uuid == characteristicUUID,
           let value = advertisingCharacteristic?.value {
            request.value = value
            peripheral.respond(to: request, withResult: .success)
        } else {
            peripheral.respond(to: request, withResult: .attributeNotFound)
        }
    }
}

// MARK: - Helper Extension

// Note: RSSI is captured in didDiscover callback and stored in peripheralRSSI dictionary
