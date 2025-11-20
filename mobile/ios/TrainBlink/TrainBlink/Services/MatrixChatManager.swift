//
//  MatrixChatManager.swift
//  TrainBlink
//
//  Matrix SDK integration for ephemeral messaging
//

import Foundation
import MatrixSDK
import Combine

/// Matrix Chat Manager for ephemeral DM rooms
class MatrixChatManager: ObservableObject {

    // MARK: - Published Properties

    @Published var isInitialized: Bool = false
    @Published var isLoggedIn: Bool = false
    @Published var activeRooms: [MatrixEphemeralRoom] = []
    @Published var currentRoom: MXRoom?
    @Published var messages: [MessageItem] = []

    // MARK: - Private Properties

    private var mxRestClient: MXRestClient?
    private var mxSession: MXSession?
    private var mxRoom: MXRoom?

    private let homeserverURL: String
    private let apiService: APIService

    // Message queue for offline support
    private var messageQueue: [QueuedMessage] = []
    private var isProcessingQueue = false

    // Cancellables
    private var cancellables = Set<AnyCancellable>()

    // Timer for room expiry checks
    private var expiryTimer: Timer?

    // MARK: - Initialization

    init(homeserverURL: String = AppConfig.matrixHomeserverURL, apiService: APIService) {
        self.homeserverURL = homeserverURL
        self.apiService = apiService
    }

    deinit {
        expiryTimer?.invalidate()
    }

    // MARK: - Authentication

    /// Initialize Matrix client and login with credentials
    func login(matrixUserId: String, accessToken: String) async throws {
        guard let url = URL(string: homeserverURL) else {
            throw MatrixError.invalidHomeserver
        }

        // Create REST client
        let credentials = MXCredentials(
            homeServer: homeserverURL,
            userId: matrixUserId,
            accessToken: accessToken
        )

        mxRestClient = MXRestClient(credentials: credentials, unrecognizedCertificateHandler: nil)

        // Create session
        guard let restClient = mxRestClient else {
            throw MatrixError.initializationFailed
        }

        let session = MXSession(matrixRestClient: restClient)
        mxSession = session

        // Start session
        try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<Void, Error>) in
            session.start { response in
                switch response {
                case .success:
                    self.isInitialized = true
                    self.isLoggedIn = true
                    self.setupSessionListeners()
                    self.startExpiryTimer()
                    continuation.resume()
                case .failure(let error):
                    continuation.resume(throwing: error)
                }
            }
        }

        print("[Matrix] Logged in as \(matrixUserId)")
    }

    /// Logout from Matrix
    func logout() async {
        guard let session = mxSession else { return }

        await withCheckedContinuation { continuation in
            session.logout { _ in
                continuation.resume()
            }
        }

        mxSession = nil
        mxRestClient = nil
        isLoggedIn = false
        isInitialized = false
        expiryTimer?.invalidate()

        print("[Matrix] Logged out")
    }

    // MARK: - Room Management

    /// Load active ephemeral rooms from backend
    func loadActiveRooms() async throws {
        let rooms = try await apiService.getActiveEphemeralDMs()

        await MainActor.run {
            self.activeRooms = rooms
        }

        print("[Matrix] Loaded \(rooms.count) active rooms")
    }

    /// Join a Matrix room by ID
    func joinRoom(roomId: String) async throws -> MXRoom {
        guard let session = mxSession else {
            throw MatrixError.notLoggedIn
        }

        return try await withCheckedThrowingContinuation { continuation in
            session.joinRoom(roomId) { response in
                switch response {
                case .success(let room):
                    continuation.resume(returning: room)
                case .failure(let error):
                    continuation.resume(throwing: error)
                }
            }
        }
    }

    /// Open a chat room
    func openRoom(ephemeralRoom: MatrixEphemeralRoom) async throws {
        guard let session = mxSession else {
            throw MatrixError.notLoggedIn
        }

        // Get or join room
        var room = session.room(withRoomId: ephemeralRoom.roomId)
        if room == nil {
            room = try await joinRoom(roomId: ephemeralRoom.roomId)
        }

        guard let room = room else {
            throw MatrixError.roomNotFound
        }

        mxRoom = room
        currentRoom = room

        // Load messages
        await loadMessages(room: room)

        // Listen for new messages
        setupRoomListeners(room: room)

        print("[Matrix] Opened room: \(ephemeralRoom.roomId)")
    }

    /// Close current room
    func closeRoom() {
        mxRoom = nil
        currentRoom = nil
        messages = []
        print("[Matrix] Closed room")
    }

    /// Extend room lifetime
    func extendRoomLifetime(room: MatrixEphemeralRoom, hours: Int) async throws {
        try await apiService.extendRoomLifetime(roomId: room.id, extensionHours: hours)

        // Reload rooms to get updated expiry
        try await loadActiveRooms()

        print("[Matrix] Extended room lifetime by \(hours) hours")
    }

    // MARK: - Messaging

    /// Send a text message
    func sendMessage(text: String) async throws {
        guard let room = mxRoom else {
            throw MatrixError.noActiveRoom
        }

        // Check if online
        if !isOnline {
            // Queue message for later
            queueMessage(text: text, roomId: room.roomId)
            return
        }

        // Create local echo immediately
        let localMessage = MessageItem(
            id: UUID().uuidString,
            senderId: mxSession?.myUserId ?? "",
            text: text,
            timestamp: Date(),
            isSent: false,
            isMine: true
        )

        await MainActor.run {
            self.messages.append(localMessage)
        }

        // Send to Matrix
        try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<Void, Error>) in
            room.sendTextMessage(text, threadId: nil, localEcho: nil) { response in
                switch response {
                case .success(let eventId):
                    print("[Matrix] Message sent: \(eventId ?? "unknown")")

                    // Update local echo
                    Task { @MainActor in
                        if let index = self.messages.firstIndex(where: { $0.id == localMessage.id }) {
                            self.messages[index].isSent = true
                            if let eventId = eventId {
                                self.messages[index].id = eventId
                            }
                        }
                    }

                    continuation.resume()
                case .failure(let error):
                    // Remove failed local echo
                    Task { @MainActor in
                        self.messages.removeAll { $0.id == localMessage.id }
                    }
                    continuation.resume(throwing: error)
                }
            }
        }
    }

    /// Load message history
    private func loadMessages(room: MXRoom) async {
        let timeline = room.timelineLaggedWindow()

        await withCheckedContinuation { continuation in
            timeline.paginate(50, direction: .backwards, onlyFromStore: false) { _ in
                continuation.resume()
            } failure: { _ in
                continuation.resume()
            }
        }

        // Convert MXEvents to MessageItems
        let events = timeline.events.compactMap { event -> MessageItem? in
            guard event.eventType == .roomMessage,
                  let content = event.content,
                  let msgType = content["msgtype"] as? String,
                  msgType == "m.text",
                  let body = content["body"] as? String else {
                return nil
            }

            return MessageItem(
                id: event.eventId ?? UUID().uuidString,
                senderId: event.sender ?? "",
                text: body,
                timestamp: Date(timeIntervalSince1970: TimeInterval(event.originServerTs) / 1000),
                isSent: true,
                isMine: event.sender == mxSession?.myUserId
            )
        }

        await MainActor.run {
            self.messages = events.reversed()
        }
    }

    // MARK: - Offline Message Queue

    private func queueMessage(text: String, roomId: String) {
        let queued = QueuedMessage(
            id: UUID(),
            roomId: roomId,
            text: text,
            timestamp: Date()
        )
        messageQueue.append(queued)
        print("[Matrix] Message queued for offline delivery: \(text)")
    }

    private func processMessageQueue() async {
        guard !isProcessingQueue, isOnline else { return }

        isProcessingQueue = true
        defer { isProcessingQueue = false }

        for message in messageQueue {
            do {
                // Get room
                guard let session = mxSession,
                      let room = session.room(withRoomId: message.roomId) else {
                    continue
                }

                // Send message
                try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<Void, Error>) in
                    room.sendTextMessage(message.text, threadId: nil, localEcho: nil) { response in
                        switch response {
                        case .success:
                            continuation.resume()
                        case .failure(let error):
                            continuation.resume(throwing: error)
                        }
                    }
                }

                // Remove from queue
                messageQueue.removeAll { $0.id == message.id }
                print("[Matrix] Queued message sent: \(message.text)")
            } catch {
                print("[Matrix] Failed to send queued message: \(error.localizedDescription)")
            }
        }
    }

    // MARK: - Listeners

    private func setupSessionListeners() {
        guard let session = mxSession else { return }

        // Listen for sync state changes
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleSyncStateChange),
            name: .mxSessionStateDidChange,
            object: session
        )
    }

    private func setupRoomListeners(room: MXRoom) {
        // Listen for new messages
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handleNewMessage(_:)),
            name: .mxRoomDidFlushData,
            object: room
        )
    }

    @objc private func handleSyncStateChange() {
        print("[Matrix] Sync state changed")

        // Try to process queued messages
        Task {
            await processMessageQueue()
        }
    }

    @objc private func handleNewMessage(_ notification: Notification) {
        guard let room = notification.object as? MXRoom,
              room.roomId == mxRoom?.roomId else {
            return
        }

        // Reload messages
        Task {
            await loadMessages(room: room)
        }
    }

    // MARK: - Room Expiry Monitoring

    private func startExpiryTimer() {
        expiryTimer = Timer.scheduledTimer(withTimeInterval: 60.0, repeats: true) { [weak self] _ in
            Task {
                try? await self?.loadActiveRooms()
            }
        }
    }

    // MARK: - Helpers

    private var isOnline: Bool {
        return mxSession?.state == .running
    }
}

// MARK: - Models

/// Message item for UI display
struct MessageItem: Identifiable, Equatable {
    var id: String
    let senderId: String
    let text: String
    let timestamp: Date
    var isSent: Bool
    let isMine: Bool

    var timeFormatted: String {
        let formatter = DateFormatter()
        formatter.timeStyle = .short
        return formatter.string(from: timestamp)
    }
}

/// Queued message for offline delivery
struct QueuedMessage: Identifiable {
    let id: UUID
    let roomId: String
    let text: String
    let timestamp: Date
}

// MARK: - Errors

enum MatrixError: Error, LocalizedError {
    case invalidHomeserver
    case initializationFailed
    case notLoggedIn
    case roomNotFound
    case noActiveRoom
    case messageSendFailed

    var errorDescription: String? {
        switch self {
        case .invalidHomeserver:
            return "Invalid Matrix homeserver URL"
        case .initializationFailed:
            return "Failed to initialize Matrix session"
        case .notLoggedIn:
            return "Not logged in to Matrix"
        case .roomNotFound:
            return "Room not found"
        case .noActiveRoom:
            return "No active chat room"
        case .messageSendFailed:
            return "Failed to send message"
        }
    }
}
