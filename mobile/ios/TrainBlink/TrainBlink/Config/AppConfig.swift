//
//  AppConfig.swift
//  TrainBlink
//
//  Application configuration
//

import Foundation

enum AppConfig {

    // MARK: - API Configuration

    static let apiBaseURL: String = {
        #if DEBUG
        return "http://localhost:8080"
        #else
        return "https://api.trainblink.org"
        #endif
    }()

    // MARK: - BLE Configuration

    /// TrainBlink BLE Service UUID
    static let bleServiceUUID = "TB000000-0000-1000-8000-00805F9B34FB"

    /// TrainBlink BLE Characteristic UUID
    static let bleCharacteristicUUID = "TB000001-0000-1000-8000-00805F9B34FB"

    /// BLE scan interval (seconds)
    static let bleScanInterval: TimeInterval = 1.0

    /// Discovery cleanup interval (seconds)
    static let discoveryCleanupInterval: TimeInterval = 10.0

    /// Discovery freshness threshold (seconds)
    static let discoveryFreshnessThreshold: TimeInterval = 30.0

    // MARK: - Matrix Configuration

    /// Matrix homeserver URL
    static let matrixHomeserverURL: String = {
        #if DEBUG
        return "https://matrix.trainblink.org"
        #else
        return "https://matrix.trainblink.org"
        #endif
    }()

    /// Default room lifetime (hours)
    static let defaultRoomLifetimeHours: Int = 24

    /// Room expiry warning threshold (hours)
    static let roomExpiryWarningHours: Int = 1

    // MARK: - Features

    /// Enable BLE discovery
    static let bleDiscoveryEnabled = true

    /// Enable analytics logging
    static let analyticsEnabled = true

    /// Enable debug logging
    static let debugLoggingEnabled: Bool = {
        #if DEBUG
        return true
        #else
        return false
        #endif
    }()
}
