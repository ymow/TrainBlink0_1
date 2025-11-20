//
//  Trip.swift
//  TrainBlink
//
//  Core trip model for tracking user journeys
//

import Foundation

/// Trip status enumeration
enum TripStatus: String, Codable {
    case active = "active"
    case ended = "ended"
    case cancelled = "cancelled"
}

/// Trip model representing a user's journey
struct Trip: Codable, Identifiable {
    let id: UUID
    let userId: UUID
    let route: String
    let trainNumber: String?
    let departureTime: Date
    let estimatedArrival: Date
    var actualEndTime: Date?
    var discoveryEnabled: Bool
    let bleAnonymousId: String
    var status: TripStatus
    let createdAt: Date
    let updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case route
        case trainNumber = "train_number"
        case departureTime = "departure_time"
        case estimatedArrival = "estimated_arrival"
        case actualEndTime = "actual_end_time"
        case discoveryEnabled = "discovery_enabled"
        case bleAnonymousId = "ble_anonymous_id"
        case status
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    /// Check if trip is currently active
    var isActive: Bool {
        return status == .active && Date() < estimatedArrival
    }

    /// Calculate trip duration
    var duration: TimeInterval {
        return estimatedArrival.timeIntervalSince(departureTime)
    }

    /// Calculate remaining time
    var remainingTime: TimeInterval {
        return max(0, estimatedArrival.timeIntervalSince(Date()))
    }

    /// Format remaining time as string
    var remainingTimeFormatted: String {
        let remaining = remainingTime
        let hours = Int(remaining) / 3600
        let minutes = Int(remaining) % 3600 / 60

        if hours > 0 {
            return "\(hours)h \(minutes)m remaining"
        } else {
            return "\(minutes)m remaining"
        }
    }
}

/// Request payload for starting a new trip
struct StartTripRequest: Codable {
    let route: String
    let trainNumber: String?
    let departureTime: Date
    let estimatedArrival: Date

    enum CodingKeys: String, CodingKey {
        case route
        case trainNumber = "train_number"
        case departureTime = "departure_time"
        case estimatedArrival = "estimated_arrival"
    }
}

/// Request payload for ending a trip
struct EndTripRequest: Codable {
    let tripId: UUID

    enum CodingKeys: String, CodingKey {
        case tripId = "trip_id"
    }
}

/// API response wrapper
struct TripResponse: Codable {
    let status: String
    let data: Trip?
    let error: String?
}

struct TripsListResponse: Codable {
    let status: String
    let data: TripsData?
    let error: String?

    struct TripsData: Codable {
        let trips: [Trip]
        let total: Int
        let limit: Int
        let offset: Int
    }
}
