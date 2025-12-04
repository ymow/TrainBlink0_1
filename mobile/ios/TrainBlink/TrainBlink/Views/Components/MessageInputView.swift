//
//  MessageInputView.swift
//  TrainBlink
//
//  Message input component following stream-chat-swift patterns
//  Text field with send button and multiline support
//

import SwiftUI

/// Message input bar with text field and send button
/// Pattern: iOS Messages-style input with disabled state management
struct MessageInputView: View {
    @Binding var text: String
    @FocusState.Binding var isFocused: Bool
    let onSend: () -> Void

    var body: some View {
        HStack(alignment: .bottom, spacing: 12) {
            // Text field
            TextField("Type a message...", text: $text, axis: .vertical)
                .textFieldStyle(.roundedBorder)
                .focused($isFocused)
                .lineLimit(1...5)
                .accessibilityLabel("Message text field")
                .accessibilityHint("Type your message here")

            // Send button
            Button(action: handleSend) {
                Image(systemName: "arrow.up.circle.fill")
                    .font(.title2)
                    .foregroundColor(canSend ? .blue : .gray)
            }
            .disabled(!canSend)
            .accessibilityLabel("Send message")
            .accessibilityHint(canSend ? "Send your message" : "Type a message first")
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(Color(.systemBackground))
    }

    // MARK: - Private Methods

    private var canSend: Bool {
        !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private func handleSend() {
        guard canSend else { return }
        onSend()
    }
}

// MARK: - Preview

#Preview {
    VStack {
        Spacer()

        // Sample messages
        VStack(spacing: 12) {
            Text("Message 1")
                .padding()
                .background(Color.blue)
                .foregroundColor(.white)
                .cornerRadius(16)
                .frame(maxWidth: .infinity, alignment: .trailing)

            Text("Message 2")
                .padding()
                .background(Color(.systemGray5))
                .cornerRadius(16)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding()

        Divider()

        // Input view
        MessageInputView(
            text: .constant("Hello!"),
            isFocused: .constant(false),
            onSend: { print("Send tapped") }
        )
    }
}

#Preview("Empty") {
    MessageInputView(
        text: .constant(""),
        isFocused: .constant(false),
        onSend: { print("Send tapped") }
    )
}

#Preview("Focused") {
    MessageInputView(
        text: .constant("Typing a message..."),
        isFocused: .constant(true),
        onSend: { print("Send tapped") }
    )
}
