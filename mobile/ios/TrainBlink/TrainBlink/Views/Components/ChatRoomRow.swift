//
//  ChatRoomRow.swift
//  TrainBlink
//
//  Chat room row component following stream-chat-swift patterns
//  Displays room info with avatar, expiry, and message count
//

import SwiftUI

/// Chat room row displaying room information in list view
/// Pattern: Avatar + room info + timestamp + badge
struct ChatRoomRow: View {
    let room: MatrixEphemeralRoom

    var body: some View {
        HStack(spacing: 12) {
            // Avatar
            Circle()
                .fill(Color.blue.opacity(0.2))
                .frame(width: 48, height: 48)
                .overlay(
                    Text("🚄")
                        .font(.title2)
                )
                .accessibilityLabel("Anonymous traveler avatar")

            // Room info
            VStack(alignment: .leading, spacing: 4) {
                // Title and timestamp
                HStack {
                    Text("Anonymous Traveler")
                        .font(.headline)
                        .foregroundColor(.primary)

                    Spacer()

                    if let lastMessageTime = room.lastMessageTimeFormatted {
                        Text(lastMessageTime)
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }
                }

                // Expiry info and message count
                HStack(spacing: 8) {
                    // Expiry indicator
                    Image(systemName: "clock")
                        .font(.caption)
                        .foregroundColor(room.isExpiringSoon ? .red : .orange)

                    Text("Expires in \(room.timeRemainingFormatted)")
                        .font(.subheadline)
                        .foregroundColor(.secondary)

                    Spacer()

                    // Message count badge
                    if room.messageCount > 0 {
                        Text("\(room.messageCount)")
                            .font(.caption)
                            .fontWeight(.semibold)
                            .foregroundColor(.white)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(Color.blue)
                            .clipShape(Capsule())
                            .accessibilityLabel("\(room.messageCount) messages")
                    }
                }
            }
        }
        .padding(.vertical, 8)
        .contentShape(Rectangle())
    }
}

// MARK: - Preview

#Preview("Normal Room") {
    List {
        ChatRoomRow(room: MatrixEphemeralRoom(
            id: UUID(),
            roomId: "!test1:matrix.org",
            trip1Id: UUID(),
            trip2Id: UUID(),
            anonymousId1: "TB_abc",
            anonymousId2: "TB_xyz",
            mlsGroupId: nil,
            expiresAt: Date().addingTimeInterval(7200), // 2 hours
            messageCount: 5,
            lastMessageAt: Date().addingTimeInterval(-300), // 5 min ago
            createdAt: Date()
        ))
    }
}

#Preview("Expiring Soon") {
    List {
        ChatRoomRow(room: MatrixEphemeralRoom(
            id: UUID(),
            roomId: "!test2:matrix.org",
            trip1Id: UUID(),
            trip2Id: UUID(),
            anonymousId1: "TB_def",
            anonymousId2: "TB_uvw",
            mlsGroupId: nil,
            expiresAt: Date().addingTimeInterval(1800), // 30 minutes
            messageCount: 12,
            lastMessageAt: Date().addingTimeInterval(-60), // 1 min ago
            createdAt: Date()
        ))
    }
}

#Preview("Multiple Rooms") {
    List {
        ChatRoomRow(room: MatrixEphemeralRoom(
            id: UUID(),
            roomId: "!test1:matrix.org",
            trip1Id: UUID(),
            trip2Id: UUID(),
            anonymousId1: "TB_abc",
            anonymousId2: "TB_xyz",
            mlsGroupId: nil,
            expiresAt: Date().addingTimeInterval(7200),
            messageCount: 5,
            lastMessageAt: Date().addingTimeInterval(-300),
            createdAt: Date()
        ))

        ChatRoomRow(room: MatrixEphemeralRoom(
            id: UUID(),
            roomId: "!test2:matrix.org",
            trip1Id: UUID(),
            trip2Id: UUID(),
            anonymousId1: "TB_def",
            anonymousId2: "TB_uvw",
            mlsGroupId: nil,
            expiresAt: Date().addingTimeInterval(1800),
            messageCount: 0,
            lastMessageAt: nil,
            createdAt: Date()
        ))

        ChatRoomRow(room: MatrixEphemeralRoom(
            id: UUID(),
            roomId: "!test3:matrix.org",
            trip1Id: UUID(),
            trip2Id: UUID(),
            anonymousId1: "TB_ghi",
            anonymousId2: "TB_rst",
            mlsGroupId: nil,
            expiresAt: Date().addingTimeInterval(10800),
            messageCount: 24,
            lastMessageAt: Date().addingTimeInterval(-3600), // 1 hour ago
            createdAt: Date()
        ))
    }
}
