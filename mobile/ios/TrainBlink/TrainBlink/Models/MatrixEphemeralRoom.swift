//
//  MatrixEphemeralRoom.swift
//  TrainBlink
//
//  Model for ephemeral Matrix DM rooms
//  Extracted from APIService for better organization
//

import Foundation

/// Represents an ephemeral direct message room with automatic expiry
struct MatrixEphemeralRoom: Identifiable, Codable, Equatable {
    // MARK: - Properties

    let id: UUID
    let roomId: String
    let trip1Id: UUID
    let trip2Id: UUID
    let anonymousId1: String
    let anonymousId2: String
    let mlsGroupId: String?
    let expiresAt: Date
    let messageCount: Int
    let lastMessageAt: Date?
    let createdAt: Date

    // MARK: - Coding Keys

    enum CodingKeys: String, CodingKey {
        case id
        case roomId = "room_id"
        case trip1Id = "trip1_id"
        case trip2Id = "trip2_id"
        case anonymousId1 = "anonymous_id_1"
        case anonymousId2 = "anonymous_id_2"
        case mlsGroupId = "mls_group_id"
        case expiresAt = "expires_at"
        case messageCount = "message_count"
        case lastMessageAt = "last_message_at"
        case createdAt = "created_at"
    }

    // MARK: - Computed Properties (Android pattern)

    /// Time remaining until room expires
    var timeRemaining: TimeInterval {
        max(0, expiresAt.timeIntervalSinceNow)
    }

    /// Check if room is expiring soon (within 1 hour)
    var isExpiringSoon: Bool {
        timeRemaining < 3600
    }

    /// Formatted time remaining string
    var timeRemainingFormatted: String {
        let hours = Int(timeRemaining) / 3600
        let minutes = Int(timeRemaining) % 3600 / 60

        if hours > 0 {
            return "\(hours)h \(minutes)m"
        } else if minutes > 0 {
            return "\(minutes)m"
        } else {
            return "Expiring soon"
        }
    }

    /// Relative time format for last message ("2m ago", "1h ago")
    var lastMessageTimeFormatted: String? {
        guard let lastMessageAt = lastMessageAt else { return nil }

        let now = Date()
        let interval = now.timeIntervalSince(lastMessageAt)

        if interval < 60 {
            return "just now"
        } else if interval < 3600 {
            let minutes = Int(interval / 60)
            return "\(minutes)m ago"
        } else if interval < 86400 {
            let hours = Int(interval / 3600)
            return "\(hours)h ago"
        } else {
            let days = Int(interval / 86400)
            return "\(days)d ago"
        }
    }
}

// MARK: - Response Models

struct MatrixRoomResponse: Codable {
    let status: String
    let data: MatrixEphemeralRoom?
}

struct MatrixRoomsResponse: Codable {
    let status: String
    let data: [MatrixEphemeralRoom]
}
