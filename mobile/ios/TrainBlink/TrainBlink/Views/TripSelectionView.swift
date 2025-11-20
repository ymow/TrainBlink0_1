//
//  TripSelectionView.swift
//  TrainBlink
//
//  Trip selection and creation UI
//

import SwiftUI

struct TripSelectionView: View {

    @StateObject private var viewModel: TripSelectionViewModel
    @State private var showingTripForm = false

    init(apiService: APIService, bleManager: BLEDiscoveryManager) {
        _viewModel = StateObject(wrappedValue: TripSelectionViewModel(apiService: apiService, bleManager: bleManager))
    }

    var body: some View {
        NavigationView {
            ZStack {
                if let activeTrip = viewModel.activeTrip {
                    ActiveTripView(trip: activeTrip, viewModel: viewModel)
                } else {
                    NoActiveTripView(showingTripForm: $showingTripForm)
                }
            }
            .navigationTitle("TrainBlink")
            .navigationBarTitleDisplayMode(.large)
            .sheet(isPresented: $showingTripForm) {
                StartTripFormView(viewModel: viewModel, isPresented: $showingTripForm)
            }
            .task {
                await viewModel.loadActiveTrip()
            }
            .alert("Error", isPresented: $viewModel.showError) {
                Button("OK", role: .cancel) {}
            } message: {
                Text(viewModel.errorMessage ?? "An error occurred")
            }
        }
    }
}

// MARK: - No Active Trip View

struct NoActiveTripView: View {
    @Binding var showingTripForm: Bool

    var body: some View {
        VStack(spacing: 24) {
            Image(systemName: "train.side.front.car")
                .font(.system(size: 80))
                .foregroundColor(.blue)

            VStack(spacing: 12) {
                Text("No Active Trip")
                    .font(.title)
                    .fontWeight(.bold)

                Text("Start a trip to discover nearby travelers and chat anonymously during your journey")
                    .font(.body)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 32)
            }

            Button(action: {
                showingTripForm = true
            }) {
                Label("Start Trip", systemImage: "plus.circle.fill")
                    .font(.headline)
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .frame(height: 56)
                    .background(Color.blue)
                    .cornerRadius(16)
            }
            .padding(.horizontal, 32)
            .padding(.top, 16)
        }
    }
}

// MARK: - Active Trip View

struct ActiveTripView: View {
    let trip: Trip
    @ObservedObject var viewModel: TripSelectionViewModel

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Trip Header
                VStack(spacing: 12) {
                    Image(systemName: "train.side.front.car")
                        .font(.system(size: 60))
                        .foregroundColor(.blue)

                    Text(trip.route)
                        .font(.title2)
                        .fontWeight(.bold)
                        .multilineTextAlignment(.center)

                    if let trainNumber = trip.trainNumber {
                        Text("Train \(trainNumber)")
                            .font(.subheadline)
                            .foregroundColor(.secondary)
                    }

                    Text(trip.remainingTimeFormatted)
                        .font(.headline)
                        .foregroundColor(.orange)
                }
                .padding()
                .frame(maxWidth: .infinity)
                .background(Color(.systemBackground))
                .cornerRadius(16)
                .shadow(color: Color.black.opacity(0.05), radius: 8, x: 0, y: 2)

                // Trip Details
                VStack(alignment: .leading, spacing: 16) {
                    TripDetailRow(
                        icon: "calendar",
                        title: "Departure",
                        value: trip.departureTime.formatted(date: .abbreviated, time: .shortened)
                    )

                    TripDetailRow(
                        icon: "clock",
                        title: "Estimated Arrival",
                        value: trip.estimatedArrival.formatted(date: .abbreviated, time: .shortened)
                    )

                    TripDetailRow(
                        icon: "person.wave.2",
                        title: "Discovery",
                        value: trip.discoveryEnabled ? "Enabled" : "Disabled",
                        valueColor: trip.discoveryEnabled ? .green : .red
                    )
                }
                .padding()
                .frame(maxWidth: .infinity)
                .background(Color(.systemBackground))
                .cornerRadius(16)
                .shadow(color: Color.black.opacity(0.05), radius: 8, x: 0, y: 2)

                // Discovery Toggle
                Toggle(isOn: Binding(
                    get: { trip.discoveryEnabled },
                    set: { viewModel.updateDiscoveryEnabled($0) }
                )) {
                    HStack {
                        Image(systemName: "antenna.radiowaves.left.and.right")
                            .foregroundColor(.blue)
                        Text("Enable Discovery")
                            .font(.headline)
                    }
                }
                .padding()
                .background(Color(.systemBackground))
                .cornerRadius(16)
                .shadow(color: Color.black.opacity(0.05), radius: 8, x: 0, y: 2)

                // Navigate to Nearby Travelers
                NavigationLink(destination: NearbyTravelersView(bleManager: viewModel.bleManager)) {
                    HStack {
                        Image(systemName: "person.3.fill")
                        Text("View Nearby Travelers")
                        Spacer()
                        Image(systemName: "chevron.right")
                    }
                    .font(.headline)
                    .foregroundColor(.white)
                    .padding()
                    .background(Color.blue)
                    .cornerRadius(16)
                }

                // End Trip Button
                Button(action: {
                    Task {
                        await viewModel.endTrip()
                    }
                }) {
                    Label("End Trip", systemImage: "stop.circle.fill")
                        .font(.headline)
                        .foregroundColor(.white)
                        .frame(maxWidth: .infinity)
                        .frame(height: 56)
                        .background(Color.red)
                        .cornerRadius(16)
                }

                Spacer()
            }
            .padding()
        }
    }
}

// MARK: - Trip Detail Row

struct TripDetailRow: View {
    let icon: String
    let title: String
    let value: String
    var valueColor: Color = .primary

    var body: some View {
        HStack {
            Image(systemName: icon)
                .foregroundColor(.blue)
                .frame(width: 24)

            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(.caption)
                    .foregroundColor(.secondary)
                Text(value)
                    .font(.body)
                    .foregroundColor(valueColor)
            }

            Spacer()
        }
    }
}

// MARK: - Start Trip Form

struct StartTripFormView: View {
    @ObservedObject var viewModel: TripSelectionViewModel
    @Binding var isPresented: Bool

    @State private var route: String = ""
    @State private var trainNumber: String = ""
    @State private var departureTime: Date = Date()
    @State private var estimatedArrival: Date = Date().addingTimeInterval(3600) // 1 hour default

    // Popular routes
    let popularRoutes = [
        "Tokyo → Osaka",
        "Tokyo → Kyoto",
        "Osaka → Kyoto",
        "Tokyo → Nagoya",
        "Paris → Lyon",
        "London → Manchester",
        "New York → Boston",
        "San Francisco → Los Angeles"
    ]

    var body: some View {
        NavigationView {
            Form {
                Section {
                    Picker("Select Route", selection: $route) {
                        Text("Custom Route").tag("")
                        ForEach(popularRoutes, id: \.self) { route in
                            Text(route).tag(route)
                        }
                    }

                    if route.isEmpty {
                        TextField("Enter custom route (e.g., Tokyo → Osaka)", text: $route)
                    }
                }

                Section {
                    TextField("Train Number (optional)", text: $trainNumber)
                }

                Section {
                    DatePicker("Departure Time", selection: $departureTime, in: Date()...)
                    DatePicker("Estimated Arrival", selection: $estimatedArrival, in: departureTime...)
                }

                Section {
                    Button(action: startTrip) {
                        HStack {
                            Spacer()
                            if viewModel.isLoading {
                                ProgressView()
                            } else {
                                Text("Start Trip")
                                    .fontWeight(.semibold)
                            }
                            Spacer()
                        }
                    }
                    .disabled(!isFormValid || viewModel.isLoading)
                }
            }
            .navigationTitle("Start New Trip")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        isPresented = false
                    }
                }
            }
        }
    }

    private var isFormValid: Bool {
        !route.isEmpty && estimatedArrival > departureTime
    }

    private func startTrip() {
        Task {
            let request = StartTripRequest(
                route: route,
                trainNumber: trainNumber.isEmpty ? nil : trainNumber,
                departureTime: departureTime,
                estimatedArrival: estimatedArrival
            )

            let success = await viewModel.startTrip(request: request)
            if success {
                isPresented = false
            }
        }
    }
}

// MARK: - View Model

@MainActor
class TripSelectionViewModel: ObservableObject {

    @Published var activeTrip: Trip?
    @Published var isLoading = false
    @Published var showError = false
    @Published var errorMessage: String?

    private let apiService: APIService
    let bleManager: BLEDiscoveryManager

    init(apiService: APIService, bleManager: BLEDiscoveryManager) {
        self.apiService = apiService
        self.bleManager = bleManager
    }

    func loadActiveTrip() async {
        isLoading = true
        defer { isLoading = false }

        do {
            activeTrip = try await apiService.getActiveTrip()

            // Start BLE discovery if trip is active
            if let trip = activeTrip, trip.isActive && trip.discoveryEnabled {
                bleManager.startDiscovery(for: trip)
            }
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }

    func startTrip(request: StartTripRequest) async -> Bool {
        isLoading = true
        defer { isLoading = false }

        do {
            let trip = try await apiService.startTrip(request: request)
            activeTrip = trip

            // Start BLE discovery
            if trip.discoveryEnabled {
                bleManager.startDiscovery(for: trip)
            }

            return true
        } catch {
            errorMessage = error.localizedDescription
            showError = true
            return false
        }
    }

    func endTrip() async {
        guard let trip = activeTrip else { return }

        isLoading = true
        defer { isLoading = false }

        do {
            try await apiService.endTrip(tripId: trip.id)
            bleManager.stopDiscovery()
            activeTrip = nil
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }

    func updateDiscoveryEnabled(_ enabled: Bool) {
        guard let trip = activeTrip else { return }

        Task {
            do {
                try await apiService.updateDiscoveryEnabled(tripId: trip.id, enabled: enabled)

                // Update local state
                var updatedTrip = trip
                updatedTrip.discoveryEnabled = enabled
                activeTrip = updatedTrip

                // Update BLE discovery
                if enabled {
                    bleManager.startDiscovery(for: updatedTrip)
                } else {
                    bleManager.stopDiscovery()
                }
            } catch {
                errorMessage = error.localizedDescription
                showError = true
            }
        }
    }
}

#Preview {
    TripSelectionView(
        apiService: APIService(),
        bleManager: BLEDiscoveryManager(apiService: APIService())
    )
}
