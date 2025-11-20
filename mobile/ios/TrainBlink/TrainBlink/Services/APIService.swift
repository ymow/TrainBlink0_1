//
//  APIService.swift
//  TrainBlink
//
//  HTTP API client for TrainBlink backend
//

import Foundation
import Alamofire

/// API Service for backend communication
class APIService {

    // MARK: - Configuration

    private let baseURL: String
    private let session: Session
    private var userId: UUID?
    private var jwtToken: String?

    // MARK: - Initialization

    init(baseURL: String = "http://localhost:8080") {
        self.baseURL = baseURL
        self.session = Session()
    }

    /// Set user ID for authenticated requests
    func setUserId(_ userId: UUID) {
        self.userId = userId
    }

    /// Set JWT token for production authentication
    func setJWTToken(_ token: String) {
        self.jwtToken = token
    }

    // MARK: - Headers

    private var headers: HTTPHeaders {
        var headers = HTTPHeaders()

        // Development: Use X-User-ID header
        if let userId = userId {
            headers.add(name: "X-User-ID", value: userId.uuidString)
        }

        // Production: Use JWT token
        if let token = jwtToken {
            headers.add(name: "Authorization", value: "Bearer \(token)")
        }

        headers.add(name: "Content-Type", value: "application/json")
        return headers
    }

    // MARK: - Trip Endpoints

    /// Start a new trip
    func startTrip(request: StartTripRequest) async throws -> Trip {
        let response = try await session
            .request("\(baseURL)/api/v1/trips/start", method: .post, parameters: request, encoder: JSONParameterEncoder.default, headers: headers)
            .validate()
            .serializingDecodable(TripResponse.self)
            .value

        guard let trip = response.data else {
            throw APIError.invalidResponse
        }

        return trip
    }

    /// End current trip
    func endTrip(tripId: UUID) async throws {
        let request = EndTripRequest(tripId: tripId)

        try await session
            .request("\(baseURL)/api/v1/trips/end", method: .post, parameters: request, encoder: JSONParameterEncoder.default, headers: headers)
            .validate()
            .serializingDecodable(TripResponse.self)
            .value
    }

    /// Get active trip
    func getActiveTrip() async throws -> Trip? {
        let response = try await session
            .request("\(baseURL)/api/v1/trips/active", method: .get, headers: headers)
            .validate()
            .serializingDecodable(TripResponse.self)
            .value

        return response.data
    }

    /// Get user trips
    func getUserTrips(limit: Int = 10, offset: Int = 0) async throws -> [Trip] {
        let response = try await session
            .request("\(baseURL)/api/v1/trips", method: .get, parameters: ["limit": limit, "offset": offset], headers: headers)
            .validate()
            .serializingDecodable(TripsListResponse.self)
            .value

        return response.data?.trips ?? []
    }

    /// Cancel trip
    func cancelTrip(tripId: UUID) async throws {
        try await session
            .request("\(baseURL)/api/v1/trips/\(tripId.uuidString)", method: .delete, headers: headers)
            .validate()
            .serializingDecodable(TripResponse.self)
            .value
    }

    /// Update discovery enabled
    func updateDiscoveryEnabled(tripId: UUID, enabled: Bool) async throws {
        try await session
            .request("\(baseURL)/api/v1/trips/\(tripId.uuidString)/discovery", method: .patch, parameters: ["discovery_enabled": enabled], encoder: JSONParameterEncoder.default, headers: headers)
            .validate()
            .serializingDecodable(TripResponse.self)
            .value
    }

    // MARK: - Discovery Endpoints

    /// Log discovery event
    func logDiscovery(request: LogDiscoveryRequest) async throws {
        try await session
            .request("\(baseURL)/api/v1/discoveries/log", method: .post, parameters: request, encoder: JSONParameterEncoder.default, headers: headers)
            .validate()
            .serializingData()
            .value
    }

    /// Get discovery stats
    func getDiscoveryStats() async throws -> DiscoveryStatsResponse.DiscoveryStats? {
        let response = try await session
            .request("\(baseURL)/api/v1/discoveries/stats", method: .get, headers: headers)
            .validate()
            .serializingDecodable(DiscoveryStatsResponse.self)
            .value

        return response.data
    }

    // MARK: - Matrix Endpoints

    /// Create ephemeral DM
    func createEphemeralDM(discoveredUserBLEID: String, mlsGroupID: String? = nil) async throws -> MatrixEphemeralRoom {
        let params: [String: Any] = [
            "discovered_user_ble_id": discoveredUserBLEID,
            "mls_group_id": mlsGroupID as Any
        ]

        let response = try await session
            .request("\(baseURL)/api/v1/matrix/dm/create", method: .post, parameters: params, encoding: JSONEncoding.default, headers: headers)
            .validate()
            .serializingDecodable(MatrixRoomResponse.self)
            .value

        guard let room = response.data else {
            throw APIError.invalidResponse
        }

        return room
    }

    /// Get active ephemeral DMs
    func getActiveEphemeralDMs() async throws -> [MatrixEphemeralRoom] {
        let response = try await session
            .request("\(baseURL)/api/v1/matrix/dm/active", method: .get, headers: headers)
            .validate()
            .serializingDecodable(MatrixRoomsResponse.self)
            .value

        return response.data?.rooms ?? []
    }

    /// Extend room lifetime
    func extendRoomLifetime(roomId: UUID, extensionHours: Int) async throws {
        try await session
            .request("\(baseURL)/api/v1/matrix/dm/\(roomId.uuidString)/extend", method: .patch, parameters: ["extension_hours": extensionHours], encoder: JSONParameterEncoder.default, headers: headers)
            .validate()
            .serializingData()
            .value
    }
}

// MARK: - Matrix Response Models

struct MatrixEphemeralRoom: Codable, Identifiable {
    let id: UUID
    let roomId: String
    let trip1Id: UUID
    let trip2Id: UUID
    let anonymousId1: String
    let anonymousId2: String
    let mlsGroupId: String?
    let expiresAt: Date
    let messageCount: Int
    let lastMessageAt: Date?
    let createdAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case roomId = "room_id"
        case trip1Id = "trip1_id"
        case trip2Id = "trip2_id"
        case anonymousId1 = "anonymous_id_1"
        case anonymousId2 = "anonymous_id_2"
        case mlsGroupId = "mls_group_id"
        case expiresAt = "expires_at"
        case messageCount = "message_count"
        case lastMessageAt = "last_message_at"
        case createdAt = "created_at"
    }

    /// Time remaining until expiry
    var timeRemaining: TimeInterval {
        return max(0, expiresAt.timeIntervalSince(Date()))
    }

    /// Check if room is expiring soon (within 1 hour)
    var isExpiringSoon: Bool {
        return timeRemaining < 3600
    }

    /// Format time remaining
    var timeRemainingFormatted: String {
        let hours = Int(timeRemaining) / 3600
        let minutes = Int(timeRemaining) % 3600 / 60

        if hours > 0 {
            return "\(hours)h \(minutes)m"
        } else if minutes > 0 {
            return "\(minutes)m"
        } else {
            return "Expiring soon"
        }
    }
}

struct MatrixRoomResponse: Codable {
    let status: String
    let data: MatrixEphemeralRoom?
}

struct MatrixRoomsResponse: Codable {
    let status: String
    let data: RoomsData?

    struct RoomsData: Codable {
        let rooms: [MatrixEphemeralRoom]
        let total: Int
    }
}

// MARK: - Error Handling

enum APIError: Error, LocalizedError {
    case invalidResponse
    case networkError(String)
    case unauthorized
    case serverError(Int)

    var errorDescription: String? {
        switch self {
        case .invalidResponse:
            return "Invalid response from server"
        case .networkError(let message):
            return "Network error: \(message)"
        case .unauthorized:
            return "Unauthorized request"
        case .serverError(let code):
            return "Server error: \(code)"
        }
    }
}
