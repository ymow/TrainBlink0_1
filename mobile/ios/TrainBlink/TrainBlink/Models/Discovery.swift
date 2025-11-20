//
//  Discovery.swift
//  TrainBlink
//
//  BLE discovery and nearby traveler models
//

import Foundation

/// Distance estimation based on RSSI signal strength
enum DistanceEstimate: String, Codable {
    case veryClose = "Very Close (0-2m)"
    case close = "Close (2-10m)"
    case medium = "Medium (10-30m)"
    case far = "Far (30-50m)"
    case veryFar = "Very Far (50-100m)"

    /// Initialize from RSSI value
    init(rssi: Int) {
        switch rssi {
        case -50...0:
            self = .veryClose
        case -70...(-51):
            self = .close
        case -85...(-71):
            self = .medium
        case -95...(-86):
            self = .far
        default:
            self = .veryFar
        }
    }

    /// Get emoji icon for distance
    var icon: String {
        switch self {
        case .veryClose: return "👤"
        case .close: return "🚶"
        case .medium: return "🚶‍♂️"
        case .far: return "🏃"
        case .veryFar: return "🏃‍♂️"
        }
    }

    /// Get color for distance indicator
    var color: String {
        switch self {
        case .veryClose: return "green"
        case .close: return "lightGreen"
        case .medium: return "yellow"
        case .far: return "orange"
        case .veryFar: return "red"
        }
    }

    /// Get short description
    var shortDescription: String {
        switch self {
        case .veryClose: return "Very Close"
        case .close: return "Close"
        case .medium: return "Medium"
        case .far: return "Far"
        case .veryFar: return "Very Far"
        }
    }
}

/// Discovered user via BLE
struct DiscoveredUser: Identifiable {
    let id: String // BLE Anonymous ID (TB_xxxxx)
    let rssi: Int
    let distance: DistanceEstimate
    let discoveredAt: Date
    var lastSeen: Date

    /// Check if discovery is fresh (within last 30 seconds)
    var isFresh: Bool {
        return Date().timeIntervalSince(lastSeen) < 30
    }

    /// Time since last seen
    var timeSinceLastSeen: TimeInterval {
        return Date().timeIntervalSince(lastSeen)
    }
}

/// Request payload for logging a discovery event
struct LogDiscoveryRequest: Codable {
    let tripRoute: String
    let discoveredUserAnonymousId: String
    let distanceEstimate: String
    let rssi: Int
    let discovererAgeRange: String?
    let discoveredAgeRange: String?
    let discovererGender: String?
    let discoveredGender: String?

    enum CodingKeys: String, CodingKey {
        case tripRoute = "trip_route"
        case discoveredUserAnonymousId = "discovered_user_anonymous_id"
        case distanceEstimate = "distance_estimate"
        case rssi
        case discovererAgeRange = "discoverer_age_range"
        case discoveredAgeRange = "discovered_age_range"
        case discovererGender = "discoverer_gender"
        case discoveredGender = "discovered_gender"
    }
}

/// Discovery analytics response
struct DiscoveryStatsResponse: Codable {
    let status: String
    let data: DiscoveryStats?

    struct DiscoveryStats: Codable {
        let totalDiscoveries: Int
        let uniqueUsers: Int
        let averageRssi: Double
        let mostCommonDistance: String

        enum CodingKeys: String, CodingKey {
            case totalDiscoveries = "total_discoveries"
            case uniqueUsers = "unique_users"
            case averageRssi = "average_rssi"
            case mostCommonDistance = "most_common_distance"
        }
    }
}
