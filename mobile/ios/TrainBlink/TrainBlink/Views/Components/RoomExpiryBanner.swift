//
//  RoomExpiryBanner.swift
//  TrainBlink
//
//  Room expiry banner component following stream-chat-swift patterns
//  Color-coded warning system with countdown timer
//

import SwiftUI

/// Banner displaying room expiry information with visual warning system
/// Pattern: Color-coded warning (red for urgent, orange for normal)
struct RoomExpiryBanner: View {
    let room: MatrixEphemeralRoom
    let onExtend: () -> Void

    @State private var timeRemaining: String = ""
    @State private var updateTimer: Timer?

    var body: some View {
        HStack(spacing: 12) {
            // Clock icon
            Image(systemName: "clock.fill")
                .font(.subheadline)
                .foregroundColor(iconColor)

            // Expiry text
            Text("Expires in \(timeRemaining)")
                .font(.subheadline)
                .foregroundColor(textColor)

            Spacer()

            // Extend button
            Button(action: onExtend) {
                Text("Extend")
                    .font(.subheadline.bold())
                    .foregroundColor(.blue)
            }
            .accessibilityLabel("Extend room lifetime")
            .accessibilityHint("Tap to add more time to this chat room")
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(backgroundColor)
        .onAppear {
            startTimer()
        }
        .onDisappear {
            stopTimer()
        }
    }

    // MARK: - Styling

    private var backgroundColor: Color {
        room.isExpiringSoon
            ? Color.red.opacity(0.1)
            : Color(.systemGray6)
    }

    private var textColor: Color {
        room.isExpiringSoon ? .red : .secondary
    }

    private var iconColor: Color {
        room.isExpiringSoon ? .red : .orange
    }

    // MARK: - Timer Management

    private func startTimer() {
        updateTimeRemaining()

        // Update every 60 seconds
        updateTimer = Timer.scheduledTimer(withTimeInterval: 60, repeats: true) { _ in
            updateTimeRemaining()
        }
    }

    private func stopTimer() {
        updateTimer?.invalidate()
        updateTimer = nil
    }

    private func updateTimeRemaining() {
        timeRemaining = room.timeRemainingFormatted
    }
}

// MARK: - Preview

#Preview("Normal Expiry") {
    VStack(spacing: 0) {
        RoomExpiryBanner(
            room: MatrixEphemeralRoom(
                id: UUID(),
                roomId: "!test:matrix.org",
                trip1Id: UUID(),
                trip2Id: UUID(),
                anonymousId1: "TB_abc",
                anonymousId2: "TB_xyz",
                mlsGroupId: nil,
                expiresAt: Date().addingTimeInterval(7200), // 2 hours
                messageCount: 5,
                lastMessageAt: Date(),
                createdAt: Date()
            ),
            onExtend: { print("Extend tapped") }
        )

        Divider()

        Text("Rest of chat view")
            .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

#Preview("Expiring Soon") {
    VStack(spacing: 0) {
        RoomExpiryBanner(
            room: MatrixEphemeralRoom(
                id: UUID(),
                roomId: "!test:matrix.org",
                trip1Id: UUID(),
                trip2Id: UUID(),
                anonymousId1: "TB_abc",
                anonymousId2: "TB_xyz",
                mlsGroupId: nil,
                expiresAt: Date().addingTimeInterval(1800), // 30 minutes
                messageCount: 5,
                lastMessageAt: Date(),
                createdAt: Date()
            ),
            onExtend: { print("Extend tapped") }
        )

        Divider()

        Text("Rest of chat view")
            .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
