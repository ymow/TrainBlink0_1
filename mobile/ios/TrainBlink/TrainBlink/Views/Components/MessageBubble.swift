//
//  MessageBubble.swift
//  TrainBlink
//
//  Message bubble component following stream-chat-swift patterns
//  iMessage-style bubbles with status indicators
//

import SwiftUI

/// Message bubble component displaying individual chat messages
/// Pattern: iMessage-style bubbles (blue for sent, gray for received)
struct MessageBubble: View {
    let message: MessageItem

    var body: some View {
        HStack(alignment: .bottom, spacing: 8) {
            if message.isMine {
                Spacer()
            }

            VStack(alignment: message.isMine ? .trailing : .leading, spacing: 4) {
                // Message bubble
                Text(message.text)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 8)
                    .background(bubbleColor)
                    .foregroundColor(textColor)
                    .cornerRadius(16)
                    .textSelection(.enabled)

                // Timestamp + status indicator
                HStack(spacing: 4) {
                    Text(message.timeFormatted)
                        .font(.caption2)
                        .foregroundColor(.secondary)

                    if message.isMine {
                        statusIcon
                    }
                }
            }

            if !message.isMine {
                Spacer()
            }
        }
        .padding(.horizontal)
    }

    // MARK: - Computed Properties

    private var bubbleColor: Color {
        message.isMine ? .blue : Color(.systemGray5)
    }

    private var textColor: Color {
        message.isMine ? .white : .primary
    }

    @ViewBuilder
    private var statusIcon: some View {
        if message.isPending {
            // Pending: show progress indicator
            ProgressView()
                .scaleEffect(0.6)
                .frame(width: 12, height: 12)
        } else if message.isSent {
            // Sent: show checkmark
            Image(systemName: "checkmark")
                .font(.caption2)
                .foregroundColor(.secondary)
        }
    }
}

// MARK: - Preview

#Preview("Sent Message") {
    VStack(spacing: 12) {
        MessageBubble(message: MessageItem(
            id: "1",
            senderId: "me",
            text: "Hello! How's your journey?",
            timestamp: Date(),
            isSent: true,
            isMine: true
        ))

        MessageBubble(message: MessageItem(
            id: "2",
            senderId: "me",
            text: "I'm on the Nozomi heading to Osaka",
            timestamp: Date().addingTimeInterval(-60),
            isSent: false,
            isMine: true
        ))
    }
    .padding()
}

#Preview("Received Message") {
    VStack(spacing: 12) {
        MessageBubble(message: MessageItem(
            id: "1",
            senderId: "other",
            text: "Hi there! My train is doing great, thanks!",
            timestamp: Date(),
            isSent: true,
            isMine: false
        ))

        MessageBubble(message: MessageItem(
            id: "2",
            senderId: "other",
            text: "I'm on the same line, heading to Kyoto",
            timestamp: Date().addingTimeInterval(-120),
            isSent: true,
            isMine: false
        ))
    }
    .padding()
}
