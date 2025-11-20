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
import org.matrix.android.sdk.api.Matrix
import org.matrix.android.sdk.api.MatrixConfiguration
import org.matrix.android.sdk.api.auth.data.HomeServerConnectionConfig
import org.matrix.android.sdk.api.auth.data.SessionParams
import org.matrix.android.sdk.api.session.Session
import org.matrix.android.sdk.api.session.events.model.toModel
import org.matrix.android.sdk.api.session.room.Room
import org.matrix.android.sdk.api.session.room.model.message.MessageContent
import org.matrix.android.sdk.api.session.room.model.message.MessageTextContent
import org.matrix.android.sdk.api.session.room.timeline.Timeline
import org.matrix.android.sdk.api.session.room.timeline.TimelineEvent
import org.matrix.android.sdk.api.session.room.timeline.TimelineSettings
import org.trainblink.app.network.ApiService
import org.trainblink.app.network.MatrixEphemeralRoom
import java.util.Date
import java.util.UUID
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

/**
 * Matrix Chat Service for ephemeral messaging
 */
class MatrixChatService(
    private val context: Context,
    private val apiService: ApiService,
    private val homeserverUrl: String = "https://matrix.trainblink.org"
) {

    // Coroutine scope
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    // Matrix components
    private var matrix: Matrix? = null
    private var session: Session? = null
    private var currentRoom: Room? = null
    private var timeline: Timeline? = null

    // State flows
    private val _isInitialized = MutableStateFlow(false)
    val isInitialized: StateFlow<Boolean> = _isInitialized.asStateFlow()

    private val _isLoggedIn = MutableStateFlow(false)
    val isLoggedIn: StateFlow<Boolean> = _isLoggedIn.asStateFlow()

    private val _activeRooms = MutableStateFlow<List<MatrixEphemeralRoom>>(emptyList())
    val activeRooms: StateFlow<List<MatrixEphemeralRoom>> = _activeRooms.asStateFlow()

    private val _messages = MutableStateFlow<List<MessageItem>>(emptyList())
    val messages: StateFlow<List<MessageItem>> = _messages.asStateFlow()

    // Message queue for offline support
    private val messageQueue = mutableListOf<QueuedMessage>()
    private var isProcessingQueue = false

    // MARK: - Initialization

    /**
     * Initialize Matrix SDK
     */
    fun initialize() {
        val matrixConfiguration = MatrixConfiguration(
            applicationFlavor = "TrainBlink",
            roomDisplayNameFallbackProvider = { "Anonymous Chat" }
        )

        matrix = Matrix(context, matrixConfiguration)
        _isInitialized.value = true
    }

    /**
     * Login with Matrix credentials
     */
    suspend fun login(matrixUserId: String, accessToken: String) = withContext(Dispatchers.IO) {
        val matrix = this@MatrixChatService.matrix
            ?: throw MatrixException("Matrix not initialized")

        // Create home server config
        val homeServerConfig = HomeServerConnectionConfig.Builder()
            .withHomeServerUri(homeserverUrl)
            .build()

        // Create session params
        val sessionParams = SessionParams(
            credentials = org.matrix.android.sdk.api.auth.data.Credentials(
                userId = matrixUserId,
                accessToken = accessToken,
                homeServer = homeserverUrl,
                deviceId = null
            ),
            homeServerConnectionConfig = homeServerConfig
        )

        // Get or create session
        val session = matrix.authenticationService().createSessionFromSso(
            homeServerConnectionConfig = homeServerConfig,
            credentials = sessionParams.credentials
        )

        this@MatrixChatService.session = session
        session.open()
        session.syncService().startSync(true)

        _isLoggedIn.value = true
    }

    /**
     * Logout from Matrix
     */
    suspend fun logout() = withContext(Dispatchers.IO) {
        session?.signOutService()?.signOut(true)
        session?.close()
        session = null
        _isLoggedIn.value = false
        _isInitialized.value = false
    }

    // MARK: - Room Management

    /**
     * Load active ephemeral rooms from backend
     */
    suspend fun loadActiveRooms() = withContext(Dispatchers.IO) {
        val rooms = apiService.getActiveEphemeralDMs()
        _activeRooms.value = rooms.data?.rooms ?: emptyList()
    }

    /**
     * Join a Matrix room by ID
     */
    suspend fun joinRoom(roomId: String): Room = withContext(Dispatchers.IO) {
        val session = this@MatrixChatService.session
            ?: throw MatrixException("Not logged in")

        suspendCancellableCoroutine { continuation ->
            session.roomService().joinRoom(roomId, null, emptyList(), object : org.matrix.android.sdk.api.MatrixCallback<Unit> {
                override fun onSuccess(data: Unit) {
                    val room = session.roomService().getRoom(roomId)
                    if (room != null) {
                        continuation.resume(room)
                    } else {
                        continuation.resumeWithException(MatrixException("Room not found after join"))
                    }
                }

                override fun onFailure(failure: Throwable) {
                    continuation.resumeWithException(failure)
                }
            })
        }
    }

    /**
     * Open a chat room
     */
    suspend fun openRoom(ephemeralRoom: MatrixEphemeralRoom) = withContext(Dispatchers.IO) {
        val session = this@MatrixChatService.session
            ?: throw MatrixException("Not logged in")

        // Get or join room
        var room = session.roomService().getRoom(ephemeralRoom.roomId)
        if (room == null) {
            room = joinRoom(ephemeralRoom.roomId)
        }

        currentRoom = room

        // Create timeline
        val timelineSettings = TimelineSettings(
            initialSize = 50,
            buildReadReceipts = false
        )

        val timeline = room.timelineService().createTimeline(null, timelineSettings)
        this@MatrixChatService.timeline = timeline

        // Listen for timeline updates
        timeline.addListener(timelineListener)
        timeline.start()

        // Load initial messages
        loadMessages()
    }

    /**
     * Close current room
     */
    fun closeRoom() {
        timeline?.removeListener(timelineListener)
        timeline?.dispose()
        timeline = null
        currentRoom = null
        _messages.value = emptyList()
    }

    /**
     * Extend room lifetime
     */
    suspend fun extendRoomLifetime(room: MatrixEphemeralRoom, hours: Int) = withContext(Dispatchers.IO) {
        apiService.extendRoomLifetime(
            roomId = room.id,
            request = org.trainblink.app.network.ExtendRoomRequest(extensionHours = hours)
        )

        // Reload rooms
        loadActiveRooms()
    }

    // MARK: - Messaging

    /**
     * Send a text message
     */
    suspend fun sendMessage(text: String) = withContext(Dispatchers.IO) {
        val room = currentRoom ?: throw MatrixException("No active room")

        // Check if online
        if (!isOnline) {
            // Queue message for later
            queueMessage(text, room.roomId)
            return@withContext
        }

        // Create local echo
        val localMessage = MessageItem(
            id = UUID.randomUUID().toString(),
            senderId = session?.myUserId ?: "",
            text = text,
            timestamp = Date(),
            isSent = false,
            isMine = true
        )

        withContext(Dispatchers.Main) {
            _messages.value = _messages.value + localMessage
        }

        // Send to Matrix
        try {
            val result = room.sendService().sendTextMessage(text)

            // Update local echo
            withContext(Dispatchers.Main) {
                _messages.value = _messages.value.map {
                    if (it.id == localMessage.id) {
                        it.copy(id = result.eventId.orEmpty(), isSent = true)
                    } else {
                        it
                    }
                }
            }
        } catch (e: Exception) {
            // Remove failed local echo
            withContext(Dispatchers.Main) {
                _messages.value = _messages.value.filterNot { it.id == localMessage.id }
            }
            throw e
        }
    }

    /**
     * Load message history
     */
    private fun loadMessages() {
        val timeline = this.timeline ?: return

        val events = timeline.getSnapshot()
        val messages = events.mapNotNull { event ->
            convertTimelineEventToMessage(event)
        }

        _messages.value = messages.reversed()
    }

    // MARK: - Timeline Listener

    private val timelineListener = object : Timeline.Listener {
        override fun onNewTimelineEvents(eventIds: List<String>) {
            loadMessages()
        }

        override fun onTimelineUpdated(snapshot: List<TimelineEvent>) {
            val messages = snapshot.mapNotNull { event ->
                convertTimelineEventToMessage(event)
            }
            _messages.value = messages.reversed()
        }

        override fun onTimelineFailure(throwable: Throwable) {
            // Handle timeline errors
        }
    }

    // MARK: - Offline Message Queue

    private fun queueMessage(text: String, roomId: String) {
        val queued = QueuedMessage(
            id = UUID.randomUUID(),
            roomId = roomId,
            text = text,
            timestamp = Date()
        )
        messageQueue.add(queued)
    }

    private suspend fun processMessageQueue() {
        if (isProcessingQueue || !isOnline) return

        isProcessingQueue = true

        try {
            messageQueue.toList().forEach { message ->
                try {
                    val room = session?.roomService()?.getRoom(message.roomId)
                    room?.sendService()?.sendTextMessage(message.text)
                    messageQueue.remove(message)
                } catch (e: Exception) {
                    // Keep in queue for retry
                }
            }
        } finally {
            isProcessingQueue = false
        }
    }

    // MARK: - Helpers

    private val isOnline: Boolean
        get() = session?.syncService()?.getSyncState() == org.matrix.android.sdk.api.session.sync.SyncState.RUNNING

    private fun convertTimelineEventToMessage(event: TimelineEvent): MessageItem? {
        val content = event.root.content?.toModel<MessageContent>() as? MessageTextContent
            ?: return null

        return MessageItem(
            id = event.eventId ?: UUID.randomUUID().toString(),
            senderId = event.root.senderId ?: "",
            text = content.body,
            timestamp = Date(event.root.originServerTs ?: 0),
            isSent = true,
            isMine = event.root.senderId == session?.myUserId
        )
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

    fun copy(
        id: String = this.id,
        senderId: String = this.senderId,
        text: String = this.text,
        timestamp: Date = this.timestamp,
        isSent: Boolean = this.isSent,
        isMine: Boolean = this.isMine
    ) = MessageItem(id, senderId, text, timestamp, isSent, isMine)
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
