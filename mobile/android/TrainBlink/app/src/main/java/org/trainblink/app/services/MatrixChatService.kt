package org.trainblink.app.services

import android.content.Context
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext
import org.trainblink.app.network.ApiService
import org.trainblink.app.network.MatrixEphemeralRoom
import java.util.Date
import java.util.UUID
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

/**
 * Matrix Chat Service for ephemeral messaging
 * STUB IMPLEMENTATION - Matrix SDK dependency temporarily removed
 */
class MatrixChatService(
    private val context: Context,
    private val userId: UUID,
    private val homeserverUrl: String = "https://matrix.trainblink.org"
) {
    
    private val apiService: ApiService by lazy {
        ApiService.create(context, userId)
    }

    // Coroutine scope
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    // State flows
    private val _isInitialized = MutableStateFlow(false)
    val isInitialized: StateFlow<Boolean> = _isInitialized.asStateFlow()

    private val _isLoggedIn = MutableStateFlow(false)
    val isLoggedIn: StateFlow<Boolean> = _isLoggedIn.asStateFlow()

    private val _activeRooms = MutableStateFlow<List<MatrixEphemeralRoom>>(emptyList())
    val activeRooms: StateFlow<List<MatrixEphemeralRoom>> = _activeRooms.asStateFlow()

    private val _messages = MutableStateFlow<List<MessageItem>>(emptyList())
    val messages: StateFlow<List<MessageItem>> = _messages.asStateFlow()
    
    private val _roomMessages = MutableStateFlow<Map<String, List<org.trainblink.app.ui.ChatMessage>>>(emptyMap())
    private val roomMessages: StateFlow<Map<String, List<org.trainblink.app.ui.ChatMessage>>> = _roomMessages.asStateFlow()

    // MARK: - STUB IMPLEMENTATIONS

    /**
     * Initialize Matrix SDK - STUB
     */
    fun initialize() {
        // TODO: Implement when Matrix SDK dependency is restored
        _isInitialized.value = true
    }

    /**
     * Login with Matrix credentials - STUB
     */
    suspend fun login(matrixUserId: String, accessToken: String) = withContext(Dispatchers.IO) {
        // TODO: Implement when Matrix SDK dependency is restored
        _isLoggedIn.value = true
    }

    /**
     * Logout from Matrix - STUB
     */
    suspend fun logout() = withContext(Dispatchers.IO) {
        // TODO: Implement when Matrix SDK dependency is restored
        _isLoggedIn.value = false
        _isInitialized.value = false
    }

    /**
     * Load active ephemeral rooms from backend - STUB
     */
    suspend fun loadActiveRooms() = withContext(Dispatchers.IO) {
        // TODO: Implement when Matrix SDK dependency is restored
        // For now, provide some mock data
        val mockRooms = listOf(
            MatrixEphemeralRoom(
                id = UUID.randomUUID(),
                roomId = "!room1:matrix.trainblink.org",
                trip1Id = UUID.randomUUID(),
                trip2Id = UUID.randomUUID(),
                anonymousId1 = "traveler_${userId.toString().take(8)}",
                anonymousId2 = "traveler_${UUID.randomUUID().toString().take(8)}",
                mlsGroupId = null,
                expiresAt = Date(System.currentTimeMillis() + 3600000), // 1 hour from now
                messageCount = 3,
                lastMessageAt = Date(System.currentTimeMillis() - 180000), // 3 minutes ago
                createdAt = Date(System.currentTimeMillis() - 600000) // 10 minutes ago
            ),
            MatrixEphemeralRoom(
                id = UUID.randomUUID(),
                roomId = "!room2:matrix.trainblink.org",
                trip1Id = UUID.randomUUID(),
                trip2Id = UUID.randomUUID(),
                anonymousId1 = "traveler_${userId.toString().take(8)}",
                anonymousId2 = "traveler_${UUID.randomUUID().toString().take(8)}",
                mlsGroupId = null,
                expiresAt = Date(System.currentTimeMillis() + 7200000), // 2 hours from now
                messageCount = 0,
                lastMessageAt = null,
                createdAt = Date(System.currentTimeMillis() - 120000) // 2 minutes ago
            )
        )
        _activeRooms.value = mockRooms
    }

    /**
     * Join a Matrix room by ID - STUB
     */
    suspend fun joinRoom(roomId: String): Any = withContext(Dispatchers.IO) {
        // TODO: Implement when Matrix SDK dependency is restored
        throw MatrixException("Matrix SDK not available")
    }

    /**
     * Open a chat room - STUB
     */
    suspend fun openRoom(ephemeralRoom: MatrixEphemeralRoom) = withContext(Dispatchers.IO) {
        // TODO: Implement when Matrix SDK dependency is restored
    }

    /**
     * Close current room - STUB
     */
    fun closeRoom() {
        // TODO: Implement when Matrix SDK dependency is restored
        _messages.value = emptyList()
    }

    /**
     * Extend room lifetime - STUB
     */
    suspend fun extendRoomLifetime(room: MatrixEphemeralRoom, hours: Int) = withContext(Dispatchers.IO) {
        // TODO: Implement when Matrix SDK dependency is restored
    }

    /**
     * Send a text message - STUB
     */
    suspend fun sendMessage(text: String) = withContext(Dispatchers.IO) {
        // TODO: Implement when Matrix SDK dependency is restored
    }
    
    /**
     * Send a message to a specific room - STUB
     */
    suspend fun sendMessage(roomId: String, text: String) = withContext(Dispatchers.IO) {
        // TODO: Implement when Matrix SDK dependency is restored
        // For now, add a mock message to the room
        val currentMessages = _roomMessages.value[roomId] ?: emptyList()
        val newMessage = org.trainblink.app.ui.ChatMessage(
            id = UUID.randomUUID().toString(),
            text = text,
            timestamp = Date(),
            isOwn = true
        )
        _roomMessages.value = _roomMessages.value.toMutableMap().apply {
            put(roomId, currentMessages + newMessage)
        }
    }
    
    /**
     * Get messages for a specific room - STUB
     */
    fun getMessagesForRoom(roomId: String): StateFlow<List<org.trainblink.app.ui.ChatMessage>> {
        return MutableStateFlow(_roomMessages.value[roomId] ?: emptyList()).asStateFlow()
    }
    
    /**
     * Load messages for a specific room - STUB
     */
    suspend fun loadMessagesForRoom(roomId: String) = withContext(Dispatchers.IO) {
        // TODO: Implement when Matrix SDK dependency is restored
        // For now, add some mock messages
        if (_roomMessages.value[roomId] == null) {
            val mockMessages = listOf(
                org.trainblink.app.ui.ChatMessage(
                    id = "1",
                    text = "Hey! I noticed we're on the same train route 🚂",
                    timestamp = Date(System.currentTimeMillis() - 300000), // 5 minutes ago
                    isOwn = false
                ),
                org.trainblink.app.ui.ChatMessage(
                    id = "2",
                    text = "Welcome to TrainBlink! This is an encrypted ephemeral chat room.",
                    timestamp = Date(System.currentTimeMillis() - 240000), // 4 minutes ago
                    isOwn = false
                )
            )
            _roomMessages.value = _roomMessages.value.toMutableMap().apply {
                put(roomId, mockMessages)
            }
        }
    }
}

// MARK: - Models

/**
 * Message item for UI display
 */
data class MessageItem(
    val id: String,
    val senderId: String,
    val text: String,
    val timestamp: Date,
    val isSent: Boolean,
    val isMine: Boolean
) {
    val timeFormatted: String
        get() {
            val formatter = java.text.SimpleDateFormat("HH:mm", java.util.Locale.getDefault())
            return formatter.format(timestamp)
        }

}

/**
 * Queued message for offline delivery
 */
data class QueuedMessage(
    val id: UUID,
    val roomId: String,
    val text: String,
    val timestamp: Date
)

// MARK: - Exceptions

class MatrixException(message: String) : Exception(message)