//
//  NearbyTravelersView.swift
//  TrainBlink
//
//  Nearby travelers discovery UI
//

import SwiftUI

struct NearbyTravelersView: View {

    @ObservedObject var bleManager: BLEDiscoveryManager
    @ObservedObject var apiService: APIService
    @ObservedObject var chatManager: MatrixChatManager

    @State private var selectedUser: DiscoveredUser?
    @State private var showingChatCreation = false
    @State private var createdRoom: MatrixEphemeralRoom?
    @State private var showingChat = false

    var freshDiscoveries: [DiscoveredUser] {
        bleManager.getFreshDiscoveries()
    }

    var body: some View {
        VStack(spacing: 0) {
            // Header Status
            DiscoveryStatusHeader(
                isScanning: bleManager.isScanning,
                isBroadcasting: bleManager.isBroadcasting,
                discoveryCount: freshDiscoveries.count
            )

            // Travelers List
            if freshDiscoveries.isEmpty {
                EmptyDiscoveriesView()
            } else {
                List(freshDiscoveries) { user in
                    DiscoveredUserRow(user: user)
                        .onTapGesture {
                            selectedUser = user
                            showingChatCreation = true
                        }
                }
                .listStyle(.plain)
            }
        }
        .navigationTitle("Nearby Travelers")
        .navigationBarTitleDisplayMode(.inline)
        .sheet(isPresented: $showingChatCreation) {
            if let user = selectedUser {
                CreateChatView(
                    discoveredUser: user,
                    isPresented: $showingChatCreation,
                    apiService: apiService,
                    chatManager: chatManager,
                    onRoomCreated: { room in
                        createdRoom = room
                        showingChat = true
                    }
                )
            }
        }
        .background(
            NavigationLink(
                destination: createdRoom.map { room in
                    ChatView(ephemeralRoom: room, chatManager: chatManager)
                },
                isActive: $showingChat,
                label: { EmptyView() }
            )
        )
    }
}

// MARK: - Discovery Status Header

struct DiscoveryStatusHeader: View {
    let isScanning: Bool
    let isBroadcasting: Bool
    let discoveryCount: Int

    var body: some View {
        VStack(spacing: 12) {
            HStack(spacing: 16) {
                // Scanning Status
                HStack(spacing: 8) {
                    Circle()
                        .fill(isScanning ? Color.green : Color.gray)
                        .frame(width: 8, height: 8)
                    Text("Scanning")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }

                // Broadcasting Status
                HStack(spacing: 8) {
                    Circle()
                        .fill(isBroadcasting ? Color.blue : Color.gray)
                        .frame(width: 8, height: 8)
                    Text("Broadcasting")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }

                Spacer()

                // Discovery Count
                Text("\(discoveryCount) nearby")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            .padding(.horizontal)

            if isScanning {
                HStack(spacing: 4) {
                    Image(systemName: "antenna.radiowaves.left.and.right")
                        .font(.caption)
                        .foregroundColor(.blue)
                    Text("Discovering nearby travelers...")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
            }
        }
        .padding(.vertical, 12)
        .background(Color(.systemGroupedBackground))
    }
}

// MARK: - Discovered User Row

struct DiscoveredUserRow: View {
    let user: DiscoveredUser

    var body: some View {
        HStack(spacing: 16) {
            // Distance Icon
            ZStack {
                Circle()
                    .fill(distanceColor.opacity(0.2))
                    .frame(width: 56, height: 56)

                Text(user.distance.icon)
                    .font(.system(size: 28))
            }

            // User Info
            VStack(alignment: .leading, spacing: 4) {
                Text("Traveler")
                    .font(.headline)

                Text(user.distance.shortDescription)
                    .font(.subheadline)
                    .foregroundColor(.secondary)

                HStack(spacing: 8) {
                    Image(systemName: "antenna.radiowaves.left.and.right")
                        .font(.caption)
                        .foregroundColor(.secondary)

                    Text("RSSI: \(user.rssi) dBm")
                        .font(.caption)
                        .foregroundColor(.secondary)

                    if user.isFresh {
                        Circle()
                            .fill(Color.green)
                            .frame(width: 6, height: 6)
                        Text("Active")
                            .font(.caption)
                            .foregroundColor(.green)
                    }
                }
            }

            Spacer()

            // Chat Button
            Image(systemName: "message.circle.fill")
                .font(.system(size: 32))
                .foregroundColor(.blue)
        }
        .padding(.vertical, 8)
    }

    private var distanceColor: Color {
        switch user.distance {
        case .veryClose: return .green
        case .close: return Color(red: 0.5, green: 0.8, blue: 0.3)
        case .medium: return .yellow
        case .far: return .orange
        case .veryFar: return .red
        }
    }
}

// MARK: - Empty Discoveries View

struct EmptyDiscoveriesView: View {
    var body: some View {
        VStack(spacing: 20) {
            Spacer()

            Image(systemName: "antenna.radiowaves.left.and.right.slash")
                .font(.system(size: 60))
                .foregroundColor(.gray)

            VStack(spacing: 8) {
                Text("No Travelers Nearby")
                    .font(.title3)
                    .fontWeight(.semibold)

                Text("Make sure Bluetooth is enabled and you're on the same train as other TrainBlink users")
                    .font(.body)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 32)
            }

            Spacer()
        }
    }
}

// MARK: - Create Chat View

struct CreateChatView: View {
    let discoveredUser: DiscoveredUser
    @Binding var isPresented: Bool
    let onRoomCreated: (MatrixEphemeralRoom) -> Void

    @ObservedObject var apiService: APIService
    @ObservedObject var chatManager: MatrixChatManager
    @State private var isCreatingChat = false
    @State private var showError = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationView {
            VStack(spacing: 24) {
                // User Preview
                VStack(spacing: 16) {
                    ZStack {
                        Circle()
                            .fill(Color.blue.opacity(0.2))
                            .frame(width: 100, height: 100)

                        Text(discoveredUser.distance.icon)
                            .font(.system(size: 50))
                    }

                    Text("Start Anonymous Chat")
                        .font(.title2)
                        .fontWeight(.bold)

                    VStack(spacing: 8) {
                        HStack {
                            Text("Distance:")
                                .foregroundColor(.secondary)
                            Spacer()
                            Text(discoveredUser.distance.shortDescription)
                                .fontWeight(.medium)
                        }

                        HStack {
                            Text("Signal Strength:")
                                .foregroundColor(.secondary)
                            Spacer()
                            Text("\(discoveredUser.rssi) dBm")
                                .fontWeight(.medium)
                        }

                        HStack {
                            Text("Last Seen:")
                                .foregroundColor(.secondary)
                            Spacer()
                            Text("Just now")
                                .fontWeight(.medium)
                        }
                    }
                    .padding()
                    .background(Color(.systemGroupedBackground))
                    .cornerRadius(12)
                }
                .padding()

                // Privacy Notice
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        Image(systemName: "lock.shield.fill")
                            .foregroundColor(.blue)
                        Text("Privacy & Security")
                            .font(.headline)
                    }

                    VStack(alignment: .leading, spacing: 8) {
                        PrivacyFeature(icon: "eye.slash.fill", text: "Fully anonymous - no names or profiles")
                        PrivacyFeature(icon: "timer", text: "Messages auto-delete after trip ends")
                        PrivacyFeature(icon: "lock.fill", text: "End-to-end encrypted with MLS")
                    }
                }
                .padding()
                .background(Color(.systemGroupedBackground))
                .cornerRadius(12)
                .padding(.horizontal)

                Spacer()

                // Start Chat Button
                Button(action: createChat) {
                    HStack {
                        if isCreatingChat {
                            ProgressView()
                                .progressViewStyle(CircularProgressViewStyle(tint: .white))
                        } else {
                            Image(systemName: "message.fill")
                            Text("Start Chat")
                        }
                    }
                    .font(.headline)
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .frame(height: 56)
                    .background(Color.blue)
                    .cornerRadius(16)
                }
                .disabled(isCreatingChat)
                .padding()
            }
            .navigationTitle("New Chat")
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
                Text(errorMessage ?? "Failed to create chat")
            }
        }
    }

    private func createChat() {
        Task {
            isCreatingChat = true

            do {
                // Create ephemeral DM room
                let room = try await apiService.createEphemeralDM(
                    discoveredUserBLEID: discoveredUser.id
                )

                // Open room in chat manager
                try await chatManager.openRoom(room)

                // Dismiss sheet
                isPresented = false

                // Navigate to chat view
                onRoomCreated(room)

                print("[CreateChatView] Successfully created and opened room: \(room.roomId)")
            } catch {
                errorMessage = error.localizedDescription
                showError = true
                print("[CreateChatView] Failed to create chat: \(error)")
            }

            isCreatingChat = false
        }
    }
}

// MARK: - Privacy Feature Row

struct PrivacyFeature: View {
    let icon: String
    let text: String

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .foregroundColor(.blue)
                .frame(width: 20)

            Text(text)
                .font(.subheadline)
                .foregroundColor(.secondary)
        }
    }
}

// MARK: - Create Chat View Model removed - now using apiService and chatManager directly

#Preview {
    NavigationView {
        let api = APIService()
        NearbyTravelersView(
            bleManager: BLEDiscoveryManager(apiService: api),
            apiService: api,
            chatManager: MatrixChatManager(apiService: api)
        )
    }
}
