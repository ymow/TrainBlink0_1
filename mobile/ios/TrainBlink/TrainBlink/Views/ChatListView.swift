//
//  ChatListView.swift
//  TrainBlink
//
//  List of active ephemeral chat rooms
//  Rebuilt following stream-chat-swift patterns with ChatRoomRow component
//

import SwiftUI

struct ChatListView: View {
    @ObservedObject var apiService: APIService
    @ObservedObject var chatManager: MatrixChatManager

    @State private var isLoading = false
    @State private var showError = false
    @State private var errorMessage: String?

    var body: some View {
        Group {
            if chatManager.activeRooms.isEmpty && !isLoading {
                emptyStateView
            } else {
                roomListView
            }
        }
        .navigationTitle("Chats")
        .navigationBarTitleDisplayMode(.large)
        .refreshable {
            await loadRooms()
        }
        .task {
            await loadRooms()
        }
        .alert("Error", isPresented: $showError) {
            Button("OK") { }
        } message: {
            if let errorMessage = errorMessage {
                Text(errorMessage)
            }
        }
    }

    // MARK: - Views

    private var roomListView: some View {
        List {
            ForEach(chatManager.activeRooms) { room in
                NavigationLink(destination: ChatView(ephemeralRoom: room, chatManager: chatManager)) {
                    ChatRoomRow(room: room)
                }
                .listRowInsets(EdgeInsets(top: 8, leading: 16, bottom: 8, trailing: 16))
            }
        }
        .listStyle(.plain)
        .overlay {
            if isLoading && chatManager.activeRooms.isEmpty {
                ProgressView("Loading chats...")
            }
        }
    }

    private var emptyStateView: some View {
        VStack(spacing: 20) {
            Image(systemName: "message.circle")
                .font(.system(size: 64))
                .foregroundColor(.secondary)

            VStack(spacing: 8) {
                Text("No Active Chats")
                    .font(.title2)
                    .fontWeight(.semibold)

                Text("Discover nearby travelers\nto start chatting")
                    .font(.subheadline)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
            }
        }
        .padding()
    }

    // MARK: - Private Methods

    private func loadRooms() async {
        isLoading = true
        defer { isLoading = false }

        do {
            try await chatManager.loadActiveRooms()
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }
}

// MARK: - Preview

#Preview("With Rooms") {
    NavigationView {
        ChatListView(
            apiService: APIService(),
            chatManager: {
                let manager = MatrixChatManager(apiService: APIService())
                // Simulate rooms being loaded
                return manager
            }()
        )
    }
}

#Preview("Empty State") {
    NavigationView {
        ChatListView(
            apiService: APIService(),
            chatManager: MatrixChatManager(apiService: APIService())
        )
    }
}
