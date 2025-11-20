//
//  ChatListView.swift
//  TrainBlink
//
//  List of active ephemeral chat rooms
//

import SwiftUI

struct ChatListView: View {

    @StateObject private var viewModel: ChatListViewModel
    @ObservedObject var chatManager: MatrixChatManager

    init(apiService: APIService, chatManager: MatrixChatManager) {
        _viewModel = StateObject(wrappedValue: ChatListViewModel(apiService: apiService))
        self.chatManager = chatManager
    }

    var body: some View {
        NavigationView {
            ZStack {
                if viewModel.activeRooms.isEmpty {
                    EmptyChatsView()
                } else {
                    List(viewModel.activeRooms) { room in
                        NavigationLink(destination: ChatView(ephemeralRoom: room, chatManager: chatManager)) {
                            ChatRoomRow(room: room)
                        }
                    }
                    .listStyle(.plain)
                    .refreshable {
                        await viewModel.loadActiveRooms()
                    }
                }
            }
            .navigationTitle("Chats")
            .navigationBarTitleDisplayMode(.large)
            .task {
                await viewModel.loadActiveRooms()
            }
            .alert("Error", isPresented: $viewModel.showError) {
                Button("OK", role: .cancel) {}
            } message: {
                Text(viewModel.errorMessage ?? "An error occurred")
            }
        }
    }
}

// MARK: - Chat Room Row

struct ChatRoomRow: View {
    let room: MatrixEphemeralRoom

    var body: some View {
        HStack(spacing: 12) {
            // Anonymous avatar
            ZStack {
                Circle()
                    .fill(Color.blue.opacity(0.2))
                    .frame(width: 50, height: 50)

                Text("👤")
                    .font(.system(size: 24))
            }

            // Room info
            VStack(alignment: .leading, spacing: 4) {
                Text("Anonymous Traveler")
                    .font(.headline)

                if let lastMessageAt = room.lastMessageAt {
                    Text("Last message \(lastMessageAt, style: .relative) ago")
                        .font(.caption)
                        .foregroundColor(.secondary)
                } else {
                    Text("No messages yet")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
            }

            Spacer()

            VStack(alignment: .trailing, spacing: 4) {
                // Message count badge
                if room.messageCount > 0 {
                    Text("\(room.messageCount)")
                        .font(.caption2)
                        .fontWeight(.semibold)
                        .foregroundColor(.white)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.blue)
                        .cornerRadius(12)
                }

                // Expiry indicator
                HStack(spacing: 4) {
                    Image(systemName: "timer")
                        .font(.caption2)
                    Text(room.timeRemainingFormatted)
                        .font(.caption2)
                }
                .foregroundColor(room.isExpiringSoon ? .orange : .secondary)
            }
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Empty Chats View

struct EmptyChatsView: View {
    var body: some View {
        VStack(spacing: 20) {
            Spacer()

            Image(systemName: "bubble.left.and.bubble.right")
                .font(.system(size: 60))
                .foregroundColor(.gray)

            VStack(spacing: 8) {
                Text("No Active Chats")
                    .font(.title3)
                    .fontWeight(.semibold)

                Text("Discover nearby travelers and start a chat to see your conversations here")
                    .font(.body)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 32)
            }

            Spacer()
        }
    }
}

// MARK: - View Model

@MainActor
class ChatListViewModel: ObservableObject {

    @Published var activeRooms: [MatrixEphemeralRoom] = []
    @Published var isLoading = false
    @Published var showError = false
    @Published var errorMessage: String?

    private let apiService: APIService

    init(apiService: APIService) {
        self.apiService = apiService
    }

    func loadActiveRooms() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let rooms = try await apiService.getActiveEphemeralDMs()
            activeRooms = rooms.sorted { $0.createdAt > $1.createdAt }
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }
}

#Preview {
    ChatListView(
        apiService: APIService(),
        chatManager: MatrixChatManager(apiService: APIService())
    )
}
