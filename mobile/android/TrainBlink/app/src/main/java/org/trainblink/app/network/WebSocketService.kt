package org.trainblink.app.network

import android.content.Context
import android.util.Log
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import okhttp3.*
import org.json.JSONObject
import java.util.concurrent.TimeUnit
import com.google.gson.Gson
import com.google.gson.GsonBuilder
import com.google.gson.annotations.SerializedName
import java.util.Date

enum class ConnectionStatus {
    DISCONNECTED,
    CONNECTING,
    CONNECTED,
    ERROR
}

/**
 * WebSocket Service for real-time messaging
 */
class WebSocketService(private val context: Context) {

    private val client = OkHttpClient.Builder()
        .readTimeout(0, TimeUnit.MILLISECONDS) // Disable timeout for long-lived connection
        .build()

    private val gson: Gson = GsonBuilder()
        .setDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'")
        .create()

    private var webSocket: WebSocket? = null
    private var jwtToken: String? = null
    private var stationId: String? = null
    private var isIntentionalDisconnect = false
    private var reconnectionAttempts = 0
    private val maxReconnectionAttempts = 5

    // State
    private val _connectionStatus = MutableStateFlow(ConnectionStatus.DISCONNECTED)
    val connectionStatus: StateFlow<ConnectionStatus> = _connectionStatus

    private val _messages = MutableStateFlow<List<WebSocketMessage>>(emptyList())
    val messages: StateFlow<List<WebSocketMessage>> = _messages

    private val scope = CoroutineScope(Dispatchers.IO)

    // Configuration
    fun configure(token: String, stationId: String) {
        this.jwtToken = token
        this.stationId = stationId
    }

    // Connection
    fun connect() {
        val token = jwtToken
        val station = stationId

        if (token == null || station == null) {
            Log.e(TAG, "Missing configuration")
            return
        }

        isIntentionalDisconnect = false
        _connectionStatus.value = ConnectionStatus.CONNECTING

        val request = Request.Builder()
            .url("ws://10.0.2.2:8080/ws?token=$token&station_id=$station")
            .build()

        webSocket = client.newWebSocket(request, createWebSocketListener())
    }

    fun disconnect() {
        isIntentionalDisconnect = true
        webSocket?.close(1000, "User disconnected")
        webSocket = null
        _connectionStatus.value = ConnectionStatus.DISCONNECTED
    }

    // Sending
    fun sendMessage(content: String, to: String? = null) {
        val message = ClientMessage(
            type = "message",
            to = to,
            content = content
        )
        val json = gson.toJson(message)
        webSocket?.send(json)
    }

    private fun createWebSocketListener() = object : WebSocketListener() {
        override fun onOpen(webSocket: WebSocket, response: Response) {
            Log.d(TAG, "WebSocket Connected")
            _connectionStatus.value = ConnectionStatus.CONNECTED
            reconnectionAttempts = 0
        }

        override fun onMessage(webSocket: WebSocket, text: String) {
            Log.d(TAG, "Received: $text")
            try {
                val message = gson.fromJson(text, WebSocketMessage::class.java)
                handleMessage(message)
            } catch (e: Exception) {
                Log.e(TAG, "Error parsing message", e)
            }
        }

        override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
            Log.d(TAG, "Closing: $reason")
            _connectionStatus.value = ConnectionStatus.DISCONNECTED
        }

        override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
            Log.e(TAG, "WebSocket Failure", t)
            _connectionStatus.value = ConnectionStatus.ERROR
            handleDisconnection()
        }
    }

    private fun handleMessage(message: WebSocketMessage) {
        scope.launch {
            if (message.type == "ping") {
                // Respond with pong if server initiates ping (optional, server usually pings)
                return@launch
            }
            val currentList = _messages.value.toMutableList()
            currentList.add(message)
            _messages.value = currentList
        }
    }

    private fun handleDisconnection() {
        if (!isIntentionalDisconnect && reconnectionAttempts < maxReconnectionAttempts) {
            reconnectionAttempts++
            val delayMs = reconnectionAttempts * 1000L
            Log.d(TAG, "Reconnecting in ${delayMs}ms...")
            scope.launch {
                delay(delayMs)
                connect()
            }
        }
    }

    companion object {
        private const val TAG = "WebSocketService"
    }
}

// Models
data class WebSocketMessage(
    val id: String,
    val type: String,
    val from: String?,
    val to: String?,
    @SerializedName("station_id") val stationId: String?,
    val content: String?,
    val timestamp: Date
)

data class ClientMessage(
    val type: String,
    val to: String?,
    val content: String?
)
