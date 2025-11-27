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
                    NavigationLink(destination: NearbyTravelersView(bleManager: bleManager)) {
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
            // Set a demo user ID for development
            apiService.setUserId(UUID())
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