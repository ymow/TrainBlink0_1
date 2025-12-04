import SwiftUI

struct ContentView: View {
    @StateObject private var apiService = APIService()
    @StateObject private var bleManager: BLEDiscoveryManager
    @StateObject private var chatManager: MatrixChatManager
    
    init() {
        let api = APIService()
        _apiService = StateObject(wrappedValue: api)
        _bleManager = StateObject(wrappedValue: BLEDiscoveryManager(apiService: api))
        _chatManager = StateObject(wrappedValue: MatrixChatManager(apiService: api))
    }
    
    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                Text("TrainBlink")
                    .font(.largeTitle)
                    .fontWeight(.bold)
                    .padding()
                
                VStack(spacing: 16) {
                    NavigationLink(destination: NearbyTravelersView(bleManager: bleManager, apiService: apiService, chatManager: chatManager)) {
                        FeatureButton(icon: "antenna.radiowaves.left.and.right", title: "Nearby Travelers", subtitle: "Discover fellow passengers")
                    }
                    
                    NavigationLink(destination: TripSelectionView(apiService: apiService, bleManager: bleManager)) {
                        FeatureButton(icon: "train.side.front.car", title: "Trip Selection", subtitle: "Start your journey")
                    }
                    
                    NavigationLink(destination: ChatListView(apiService: apiService, chatManager: chatManager)) {
                        FeatureButton(icon: "message.circle", title: "Chat", subtitle: "View conversations")
                    }
                }
                .padding(.horizontal)
                
                Spacer()
            }
            .navigationTitle("TrainBlink")
            .navigationBarTitleDisplayMode(.large)
        }
        .onAppear {
            Task {
                // Set a demo user ID for development
                let userId = UUID()
                apiService.setUserId(userId)

                // Login to Matrix
                do {
                    // TODO: Add backend endpoint to get Matrix credentials
                    // For now, use dev credentials with user UUID
                    let matrixUserId = "@\(userId.uuidString)_trainblink:matrix.trainblink.org"
                    let accessToken = "dev_token_\(userId.uuidString)"

                    try await chatManager.login(
                        matrixUserId: matrixUserId,
                        accessToken: accessToken
                    )

                    print("[ContentView] Matrix login successful for user: \(matrixUserId)")
                } catch {
                    print("[ContentView] Matrix login failed: \(error.localizedDescription)")
                    // Don't block app usage if Matrix login fails
                    // User can still browse and discover, just not chat
                }
            }
        }
    }
}

struct FeatureButton: View {
    let icon: String
    let title: String
    let subtitle: String
    
    var body: some View {
        HStack(spacing: 16) {
            Image(systemName: icon)
                .font(.title2)
                .foregroundColor(.blue)
                .frame(width: 40)
            
            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(.headline)
                    .foregroundColor(.primary)
                
                Text(subtitle)
                    .font(.subheadline)
                    .foregroundColor(.secondary)
            }
            
            Spacer()
            
            Image(systemName: "chevron.right")
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(12)
    }
}

#Preview {
    ContentView()
}