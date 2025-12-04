//
//  MessageItem.swift
//  TrainBlink
//
//  Model for chat messages in UI
//  Extracted from MatrixChatManager for better organization
//

import Foundation

/// Represents a chat message for UI display
struct MessageItem: Identifiable, Equatable {
    // MARK: - Properties

    let id: String
    let senderId: String
    let text: String
    let timestamp: Date
    var isSent: Bool  // Mutable for local echo updates
    let isMine: Bool

    // MARK: - Computed Properties

    /// Formatted time string (e.g., "2:30 PM")
    var timeFormatted: String {
        let formatter = DateFormatter()
        formatter.timeStyle = .short
        return formatter.string(from: timestamp)
    }

    /// Whether this message is pending (not yet sent to server)
    var isPending: Bool {
        !isSent
    }

    /// Delivery status description
    var statusDescription: String {
        if isPending {
            return "Sending..."
        } else if isSent {
            return "Sent"
        } else {
            return "Delivered"
        }
    }
}

// MARK: - Queued Message

/// Queued message for offline delivery
struct QueuedMessage: Codable {
    let id: String
    let text: String
    let timestamp: Date
    let roomId: String
}
