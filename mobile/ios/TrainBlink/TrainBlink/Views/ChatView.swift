//
//  ChatView.swift
//  TrainBlink
//
//  Chat UI with message bubbles and room management
//  Rebuilt following stream-chat-swift component composition pattern
//

import SwiftUI

struct ChatView: View {
    let ephemeralRoom: MatrixEphemeralRoom
    @ObservedObject var chatManager: MatrixChatManager

    @State private var messageText: String = ""
    @State private var showingExtendLifetime = false
    @State private var showingRoomInfo = false
    @State private var showError = false
    @State private var errorMessage: String?
    @FocusState private var isInputFocused: Bool

    var body: some View {
        VStack(spacing: 0) {
            // Room expiry banner
            RoomExpiryBanner(room: ephemeralRoom) {
                showingExtendLifetime = true
            }

            Divider()

            // Messages list
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(chatManager.messages) { message in
                            MessageBubble(message: message)
                                .id(message.id)
                        }
                    }
                    .padding(.vertical)
                }
                .onChange(of: chatManager.messages.count) { _ in
                    scrollToBottom(proxy: proxy)
                }
            }

            Divider()

            // Message input
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
                Button(action: { showingRoomInfo = true }) {
                    Image(systemName: "info.circle")
                }
                .accessibilityLabel("Room info")
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
        .sheet(isPresented: $showingRoomInfo) {
            RoomInfoView(room: ephemeralRoom, isPresented: $showingRoomInfo)
        }
        .alert("Error", isPresented: $showError) {
            Button("OK") { }
        } message: {
            if let errorMessage = errorMessage {
                Text(errorMessage)
            }
        }
    }

    // MARK: - Private Methods

    private func sendMessage() {
        let text = messageText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }

        Task {
            do {
                try await chatManager.sendMessage(text)
                messageText = ""
            } catch {
                errorMessage = error.localizedDescription
                showError = true
            }
        }
    }

    private func openRoom() async {
        do {
            try await chatManager.openRoom(ephemeralRoom)
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }

    private func scrollToBottom(proxy: ScrollViewProxy) {
        if let lastMessage = chatManager.messages.last {
            withAnimation {
                proxy.scrollTo(lastMessage.id, anchor: .bottom)
            }
        }
    }
}

// MARK: - Supporting Views

/// Extend lifetime view for adding more time to chat room
struct ExtendLifetimeView: View {
    let room: MatrixEphemeralRoom
    @ObservedObject var chatManager: MatrixChatManager
    @Binding var isPresented: Bool

    @State private var selectedHours: Int = 1
    @State private var isExtending = false
    @State private var showError = false
    @State private var errorMessage: String?

    private let hourOptions = [1, 2, 4, 8, 12, 24]

    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                // Header
                VStack(spacing: 8) {
                    Image(systemName: "clock.arrow.circlepath")
                        .font(.system(size: 48))
                        .foregroundColor(.orange)

                    Text("Extend Room Lifetime")
                        .font(.title2)
                        .fontWeight(.bold)

                    Text("Add more time to continue chatting")
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                }
                .padding(.top)

                // Hour selection
                VStack(alignment: .leading, spacing: 12) {
                    Text("Select extension duration:")
                        .font(.headline)

                    LazyVGrid(columns: [GridItem(.adaptive(minimum: 80))], spacing: 12) {
                        ForEach(hourOptions, id: \.self) { hours in
                            Button(action: { selectedHours = hours }) {
                                VStack {
                                    Text("\(hours)")
                                        .font(.title3)
                                        .fontWeight(.semibold)
                                    Text(hours == 1 ? "hour" : "hours")
                                        .font(.caption)
                                }
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(selectedHours == hours ? Color.blue : Color(.systemGray6))
                                .foregroundColor(selectedHours == hours ? .white : .primary)
                                .cornerRadius(12)
                            }
                        }
                    }
                }
                .padding()

                Spacer()

                // Extend button
                Button(action: extendLifetime) {
                    if isExtending {
                        ProgressView()
                            .progressViewStyle(.circular)
                            .frame(maxWidth: .infinity)
                    } else {
                        Text("Extend by \(selectedHours) \(selectedHours == 1 ? "hour" : "hours")")
                            .fontWeight(.semibold)
                            .frame(maxWidth: .infinity)
                    }
                }
                .buttonStyle(.borderedProminent)
                .disabled(isExtending)
                .padding()
            }
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("Cancel") {
                        isPresented = false
                    }
                }
            }
            .alert("Error", isPresented: $showError) {
                Button("OK") { }
            } message: {
                if let errorMessage = errorMessage {
                    Text(errorMessage)
                }
            }
        }
    }

    private func extendLifetime() {
        isExtending = true
        Task {
            do {
                try await chatManager.extendRoomLifetime(room, hours: selectedHours)
                isPresented = false
            } catch {
                errorMessage = error.localizedDescription
                showError = true
            }
            isExtending = false
        }
    }
}

/// Room info view showing room details
struct RoomInfoView: View {
    let room: MatrixEphemeralRoom
    @Binding var isPresented: Bool

    var body: some View {
        NavigationView {
            List {
                Section {
                    InfoRow(label: "Room ID", value: room.roomId)
                    InfoRow(label: "Created", value: room.createdAt.formatted())
                    InfoRow(label: "Expires", value: room.expiresAt.formatted())
                    InfoRow(label: "Time Remaining", value: room.timeRemainingFormatted)
                }

                Section("Privacy") {
                    Label {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("Ephemeral Chat")
                                .font(.headline)
                            Text("This chat is temporary and will be automatically deleted when it expires")
                                .font(.caption)
                                .foregroundColor(.secondary)
                        }
                    } icon: {
                        Image(systemName: "eye.slash")
                            .foregroundColor(.blue)
                    }

                    Label {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("Anonymous")
                                .font(.headline)
                            Text("Both participants remain anonymous")
                                .font(.caption)
                                .foregroundColor(.secondary)
                        }
                    } icon: {
                        Image(systemName: "person.fill.questionmark")
                            .foregroundColor(.green)
                    }
                }
            }
            .navigationTitle("Room Info")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
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

// MARK: - Preview

#Preview {
    NavigationView {
        ChatView(
            ephemeralRoom: MatrixEphemeralRoom(
                id: UUID(),
                roomId: "!test:matrix.org",
                trip1Id: UUID(),
                trip2Id: UUID(),
                anonymousId1: "TB_abc",
                anonymousId2: "TB_xyz",
                mlsGroupId: nil,
                expiresAt: Date().addingTimeInterval(7200),
                messageCount: 5,
                lastMessageAt: Date(),
                createdAt: Date()
            ),
            chatManager: MatrixChatManager(
                apiService: APIService()
            )
        )
    }
}
