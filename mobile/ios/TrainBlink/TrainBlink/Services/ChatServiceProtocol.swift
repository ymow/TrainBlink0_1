//
//  ChatServiceProtocol.swift
//  TrainBlink
//
//  Protocol defining chat service interface (stream-chat-swift pattern)
//  Enables dependency injection and testability
//

import Foundation
import Combine

/// Protocol defining chat service interface following stream-chat-swift layered architecture
/// Provides progressive disclosure: simple interface with complex implementation hidden
protocol ChatServiceProtocol: ObservableObject {
    // MARK: - State Properties

    /// Whether the service has been initialized
    var isInitialized: Bool { get }

    /// Whether the user is logged in to Matrix
    var isLoggedIn: Bool { get }

    /// List of active ephemeral chat rooms
    var activeRooms: [MatrixEphemeralRoom] { get }

    /// Messages in the currently open room
    var messages: [MessageItem] { get }

    // MARK: - Reactive Publishers

    /// Combine publisher for active rooms (reactive state management)
    var activeRoomsPublisher: AnyPublisher<[MatrixEphemeralRoom], Never> { get }

    /// Combine publisher for messages (reactive state management)
    var messagesPublisher: AnyPublisher<[MessageItem], Never> { get }

    // MARK: - Lifecycle Methods

    /// Login to Matrix with provided credentials
    /// - Parameters:
    ///   - matrixUserId: Matrix user ID (e.g., "@user:matrix.trainblink.org")
    ///   - accessToken: Matrix access token
    func login(matrixUserId: String, accessToken: String) async throws

    /// Logout from Matrix and clean up session
    func logout() async

    // MARK: - Room Management

    /// Load list of active ephemeral DM rooms from backend
    func loadActiveRooms() async throws

    /// Open a specific room and start listening for messages
    /// - Parameter room: The ephemeral room to open
    func openRoom(_ room: MatrixEphemeralRoom) async throws

    /// Close the currently open room and stop listening
    func closeRoom()

    /// Extend the lifetime of a room
    /// - Parameters:
    ///   - room: The room to extend
    ///   - hours: Number of hours to extend (1-24)
    func extendRoomLifetime(_ room: MatrixEphemeralRoom, hours: Int) async throws

    // MARK: - Messaging

    /// Send a text message in the currently open room
    /// - Parameter text: Message content to send
    func sendMessage(_ text: String) async throws

    /// Load message history for the current room
    /// - Parameter limit: Maximum number of messages to load
    func loadMessageHistory(limit: Int) async throws
}
