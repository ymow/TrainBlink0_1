//
//  ChatView.swift
//  TrainBlink
//
//  Chat UI with message bubbles and room management
//

import SwiftUI

struct ChatView: View {

    let ephemeralRoom: MatrixEphemeralRoom
    @ObservedObject var chatManager: MatrixChatManager

    @State private var messageText: String = ""
    @State private var showingExtendLifetime = false
    @State private var showError = false
    @State private var errorMessage: String?

    @FocusState private var isInputFocused: Bool

    var body: some View {
        VStack(spacing: 0) {
            // Room Expiry Header
            RoomExpiryHeader(room: ephemeralRoom, onExtend: {
                showingExtendLifetime = true
            })

            Divider()

            // Messages List
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(chatManager.messages) { message in
                            MessageBubble(message: message)
                                .id(message.id)
                        }
                    }
                    .padding()
                }
                .onChange(of: chatManager.messages.count) { _ in
                    // Scroll to bottom when new message arrives
                    if let lastMessage = chatManager.messages.last {
                        withAnimation {
                            proxy.scrollTo(lastMessage.id, anchor: .bottom)
                        }
                    }
                }
            }

            Divider()

            // Message Input
            MessageInputView(
                text: $messageText,
                isFocused: $isInputFocused,
                onSend: sendMessage
            )
        }
        .navigationTitle("Anonymous Chat")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                RoomInfoButton(room: ephemeralRoom)
            }
        }
        .task {
            await openRoom()
        }
        .onDisappear {
            chatManager.closeRoom()
        }
        .sheet(isPresented: $showingExtendLifetime) {
            ExtendLifetimeView(
                room: ephemeralRoom,
                chatManager: chatManager,
                isPresented: $showingExtendLifetime
            )
        }
        .alert("Error", isPresented: $showError) {
            Button("OK", role: .cancel) {}
        } message: {
            Text(errorMessage ?? "An error occurred")
        }
    }

    private func openRoom() async {
        do {
            try await chatManager.openRoom(ephemeralRoom: ephemeralRoom)
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }

    private func sendMessage() {
        guard !messageText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            return
        }

        let text = messageText
        messageText = ""

        Task {
            do {
                try await chatManager.sendMessage(text: text)
            } catch {
                errorMessage = error.localizedDescription
                showError = true
            }
        }
    }
}

// MARK: - Room Expiry Header

struct RoomExpiryHeader: View {
    let room: MatrixEphemeralRoom
    let onExtend: () -> Void

    var body: some View {
        HStack(spacing: 12) {
            // Expiry Icon
            Image(systemName: room.isExpiringSoon ? "exclamationmark.triangle.fill" : "timer")
                .foregroundColor(room.isExpiringSoon ? .orange : .blue)

            VStack(alignment: .leading, spacing: 2) {
                Text("Ephemeral Chat")
                    .font(.caption)
                    .foregroundColor(.secondary)

                Text("Expires in \(room.timeRemainingFormatted)")
                    .font(.subheadline)
                    .fontWeight(.medium)
                    .foregroundColor(room.isExpiringSoon ? .orange : .primary)
            }

            Spacer()

            if room.isExpiringSoon {
                Button(action: onExtend) {
                    Text("Extend")
                        .font(.caption)
                        .fontWeight(.semibold)
                        .foregroundColor(.white)
                        .padding(.horizontal, 12)
                        .padding(.vertical, 6)
                        .background(Color.blue)
                        .cornerRadius(8)
                }
            }
        }
        .padding()
        .background(room.isExpiringSoon ? Color.orange.opacity(0.1) : Color(.systemGroupedBackground))
    }
}

// MARK: - Message Bubble

struct MessageBubble: View {
    let message: MessageItem

    var body: some View {
        HStack {
            if message.isMine {
                Spacer(minLength: 60)
            }

            VStack(alignment: message.isMine ? .trailing : .leading, spacing: 4) {
                // Message text
                Text(message.text)
                    .font(.body)
                    .foregroundColor(message.isMine ? .white : .primary)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 8)
                    .background(message.isMine ? Color.blue : Color(.systemGray5))
                    .cornerRadius(16)

                // Timestamp and status
                HStack(spacing: 4) {
                    Text(message.timeFormatted)
                        .font(.caption2)
                        .foregroundColor(.secondary)

                    if message.isMine {
                        if message.isSent {
                            Image(systemName: "checkmark")
                                .font(.caption2)
                                .foregroundColor(.blue)
                        } else {
                            ProgressView()
                                .scaleEffect(0.5)
                        }
                    }
                }
                .padding(.horizontal, 4)
            }

            if !message.isMine {
                Spacer(minLength: 60)
            }
        }
    }
}

// MARK: - Message Input

struct MessageInputView: View {
    @Binding var text: String
    var isFocused: FocusState<Bool>.Binding
    let onSend: () -> Void

    var body: some View {
        HStack(spacing: 12) {
            // Text field
            TextField("Message", text: $text, axis: .vertical)
                .textFieldStyle(.plain)
                .padding(.horizontal, 12)
                .padding(.vertical, 8)
                .background(Color(.systemGray6))
                .cornerRadius(20)
                .lineLimit(1...5)
                .focused(isFocused)
                .onSubmit(onSend)

            // Send button
            Button(action: onSend) {
                Image(systemName: "arrow.up.circle.fill")
                    .font(.system(size: 32))
                    .foregroundColor(text.isEmpty ? .gray : .blue)
            }
            .disabled(text.isEmpty)
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
        .background(Color(.systemBackground))
    }
}

// MARK: - Room Info Button

struct RoomInfoButton: View {
    let room: MatrixEphemeralRoom

    @State private var showingInfo = false

    var body: some View {
        Button(action: { showingInfo = true }) {
            Image(systemName: "info.circle")
        }
        .sheet(isPresented: $showingInfo) {
            RoomInfoView(room: room, isPresented: $showingInfo)
        }
    }
}

// MARK: - Room Info View

struct RoomInfoView: View {
    let room: MatrixEphemeralRoom
    @Binding var isPresented: Bool

    var body: some View {
        NavigationView {
            List {
                Section("Chat Information") {
                    InfoRow(label: "Room ID", value: room.roomId)
                    InfoRow(label: "Messages", value: "\(room.messageCount)")
                    InfoRow(label: "Created", value: room.createdAt.formatted(date: .abbreviated, time: .shortened))
                }

                Section("Ephemeral Settings") {
                    InfoRow(label: "Expires At", value: room.expiresAt.formatted(date: .abbreviated, time: .shortened))
                    InfoRow(label: "Time Remaining", value: room.timeRemainingFormatted)
                }

                Section("Privacy & Security") {
                    PrivacyInfoRow(icon: "lock.shield.fill", title: "End-to-End Encrypted", description: "Messages are encrypted with MLS protocol")
                    PrivacyInfoRow(icon: "timer", title: "Auto-Delete", description: "All messages deleted when trip ends")
                    PrivacyInfoRow(icon: "eye.slash.fill", title: "Anonymous", description: "No personal information shared")
                }

                if let mlsGroupId = room.mlsGroupId {
                    Section("Encryption Details") {
                        InfoRow(label: "MLS Group ID", value: mlsGroupId)
                    }
                }
            }
            .navigationTitle("Chat Info")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") {
                        isPresented = false
                    }
                }
            }
        }
    }
}

struct InfoRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label)
                .foregroundColor(.secondary)
            Spacer()
            Text(value)
                .foregroundColor(.primary)
        }
    }
}

struct PrivacyInfoRow: View {
    let icon: String
    let title: String
    let description: String

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .foregroundColor(.blue)
                .frame(width: 24)

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.body)
                    .fontWeight(.medium)
                Text(description)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
    }
}

// MARK: - Extend Lifetime View

struct ExtendLifetimeView: View {
    let room: MatrixEphemeralRoom
    @ObservedObject var chatManager: MatrixChatManager
    @Binding var isPresented: Bool

    @State private var selectedHours: Int = 1
    @State private var isExtending = false
    @State private var showError = false
    @State private var errorMessage: String?

    let hourOptions = [1, 2, 4, 8, 12, 24]

    var body: some View {
        NavigationView {
            VStack(spacing: 24) {
                // Icon
                Image(systemName: "clock.arrow.circlepath")
                    .font(.system(size: 60))
                    .foregroundColor(.blue)
                    .padding(.top, 32)

                // Current expiry
                VStack(spacing: 8) {
                    Text("Current Expiry")
                        .font(.headline)
                        .foregroundColor(.secondary)

                    Text(room.expiresAt.formatted(date: .abbreviated, time: .shortened))
                        .font(.title3)
                        .fontWeight(.semibold)

                    Text("(\(room.timeRemainingFormatted) remaining)")
                        .font(.subheadline)
                        .foregroundColor(.orange)
                }
                .padding()
                .frame(maxWidth: .infinity)
                .background(Color(.systemGroupedBackground))
                .cornerRadius(12)
                .padding(.horizontal)

                // Hour selector
                VStack(alignment: .leading, spacing: 12) {
                    Text("Extend by:")
                        .font(.headline)
                        .padding(.horizontal)

                    Picker("Hours", selection: $selectedHours) {
                        ForEach(hourOptions, id: \.self) { hours in
                            Text("\(hours) hour\(hours > 1 ? "s" : "")")
                                .tag(hours)
                        }
                    }
                    .pickerStyle(.segmented)
                    .padding(.horizontal)
                }

                // New expiry preview
                VStack(spacing: 8) {
                    Text("New Expiry")
                        .font(.headline)
                        .foregroundColor(.secondary)

                    Text(newExpiryDate.formatted(date: .abbreviated, time: .shortened))
                        .font(.title3)
                        .fontWeight(.semibold)
                        .foregroundColor(.blue)
                }
                .padding()
                .frame(maxWidth: .infinity)
                .background(Color.blue.opacity(0.1))
                .cornerRadius(12)
                .padding(.horizontal)

                Spacer()

                // Extend button
                Button(action: extendLifetime) {
                    HStack {
                        if isExtending {
                            ProgressView()
                                .progressViewStyle(CircularProgressViewStyle(tint: .white))
                        } else {
                            Text("Extend Lifetime")
                        }
                    }
                    .font(.headline)
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .frame(height: 56)
                    .background(Color.blue)
                    .cornerRadius(16)
                }
                .disabled(isExtending)
                .padding()
            }
            .navigationTitle("Extend Chat")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        isPresented = false
                    }
                }
            }
            .alert("Error", isPresented: $showError) {
                Button("OK", role: .cancel) {}
            } message: {
                Text(errorMessage ?? "Failed to extend lifetime")
            }
        }
    }

    private var newExpiryDate: Date {
        return room.expiresAt.addingTimeInterval(TimeInterval(selectedHours * 3600))
    }

    private func extendLifetime() {
        Task {
            isExtending = true
            defer { isExtending = false }

            do {
                try await chatManager.extendRoomLifetime(room: room, hours: selectedHours)
                isPresented = false
            } catch {
                errorMessage = error.localizedDescription
                showError = true
            }
        }
    }
}

#Preview {
    NavigationView {
        ChatView(
            ephemeralRoom: MatrixEphemeralRoom(
                id: UUID(),
                roomId: "!test:matrix.org",
                trip1Id: UUID(),
                trip2Id: UUID(),
                anonymousId1: "TB_abc123",
                anonymousId2: "TB_def456",
                mlsGroupId: nil,
                expiresAt: Date().addingTimeInterval(7200),
                messageCount: 5,
                lastMessageAt: Date(),
                createdAt: Date().addingTimeInterval(-3600)
            ),
            chatManager: MatrixChatManager(apiService: APIService())
        )
    }
}
