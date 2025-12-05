//
//  WebSocketManager.swift
//  TrainBlink
//
//  WebSocket client for real-time messaging
//

import Foundation
import Combine

/// WebSocket Connection Status
enum ConnectionStatus {
    case disconnected
    case connecting
    case connected
    case error(Error)
}

/// WebSocket Manager
class WebSocketManager: ObservableObject {
    
    // MARK: - Properties
    
    @Published var connectionStatus: ConnectionStatus = .disconnected
    @Published var messages: [WebSocketMessage] = []
    
    private var webSocketTask: URLSessionWebSocketTask?
    private let session = URLSession(configuration: .default)
    private var pingTimer: Timer?
    private var isConnected = false
    
    private let baseURL: String
    private var jwtToken: String?
    private var stationId: String?
    
    // Reconnection logic
    private var reconnectionAttempts = 0
    private let maxReconnectionAttempts = 5
    private var isIntentionalDisconnect = false
    
    // MARK: - Initialization
    
    init(baseURL: String = "ws://localhost:8080/ws") {
        self.baseURL = baseURL
    }
    
    // MARK: - Configuration
    
    func configure(token: String, stationId: String) {
        self.jwtToken = token
        self.stationId = stationId
    }
    
    // MARK: - Connection Management
    
    func connect() {
        guard let token = jwtToken, let stationId = stationId else {
            print("WebSocket: Missing configuration")
            return
        }
        
        isIntentionalDisconnect = false
        connectionStatus = .connecting
        
        // Construct URL with query params
        var components = URLComponents(string: baseURL)
        components?.queryItems = [
            URLQueryItem(name: "token", value: token),
            URLQueryItem(name: "station_id", value: stationId)
        ]
        
        guard let url = components?.url else {
            print("WebSocket: Invalid URL")
            connectionStatus = .error(URLError(.badURL))
            return
        }
        
        let request = URLRequest(url: url)
        webSocketTask = session.webSocketTask(with: request)
        webSocketTask?.resume()
        
        receiveMessage()
        startPingTimer()
        
        isConnected = true
        connectionStatus = .connected
        reconnectionAttempts = 0
        print("WebSocket: Connecting to \(url)")
    }
    
    func disconnect() {
        isIntentionalDisconnect = true
        stopPingTimer()
        webSocketTask?.cancel(with: .goingAway, reason: nil)
        webSocketTask = nil
        isConnected = false
        connectionStatus = .disconnected
        print("WebSocket: Disconnected")
    }
    
    // MARK: - Sending Messages
    
    func send(message: ClientMessage) {
        guard isConnected else {
            print("WebSocket: Not connected")
            return
        }
        
        do {
            let data = try JSONEncoder().encode(message)
            guard let jsonString = String(data: data, encoding: .utf8) else { return }
            
            let wsMessage = URLSessionWebSocketTask.Message.string(jsonString)
            webSocketTask?.send(wsMessage) { error in
                if let error = error {
                    print("WebSocket: Send error: \(error)")
                }
            }
        } catch {
            print("WebSocket: Encoding error: \(error)")
        }
    }
    
    func sendMessage(content: String, to recipientId: String? = nil) {
        let message = ClientMessage(
            type: .message,
            to: recipientId,
            content: content
        )
        send(message: message)
    }
    
    // MARK: - Receiving Messages
    
    private func receiveMessage() {
        webSocketTask?.receive { [weak self] result in
            guard let self = self else { return }
            
            switch result {
            case .success(let message):
                switch message {
                case .string(let text):
                    self.handleMessage(text)
                case .data(let data):
                    if let text = String(data: data, encoding: .utf8) {
                        self.handleMessage(text)
                    }
                @unknown default:
                    break
                }
                
                // Continue receiving messages
                self.receiveMessage()
                
            case .failure(let error):
                print("WebSocket: Receive error: \(error)")
                self.handleDisconnection(error: error)
            }
        }
    }
    
    private func handleMessage(_ text: String) {
        guard let data = text.data(using: .utf8) else { return }
        
        do {
            let message = try JSONDecoder().decode(WebSocketMessage.self, from: data)
            
            DispatchQueue.main.async {
                self.messages.append(message)
                
                // Handle specific message types
                switch message.type {
                case .ping:
                    self.sendPong()
                default:
                    break
                }
            }
        } catch {
            print("WebSocket: Decode error: \(error)")
        }
    }
    
    private func sendPong() {
        let pong = ClientMessage(type: .pong) // Assuming .pong is a valid client message type or re-using .ping
        send(message: pong)
    }
    
    // MARK: - Reconnection
    
    private func handleDisconnection(error: Error) {
        DispatchQueue.main.async {
            self.isConnected = false
            self.connectionStatus = .error(error)
            
            if !self.isIntentionalDisconnect {
                self.attemptReconnect()
            }
        }
    }
    
    private func attemptReconnect() {
        guard reconnectionAttempts < maxReconnectionAttempts else {
            print("WebSocket: Max reconnection attempts reached")
            return
        }
        
        reconnectionAttempts += 1
        let delay = Double(reconnectionAttempts) * 1.0 // Simple linear backoff
        
        print("WebSocket: Reconnecting in \(delay)s (Attempt \(reconnectionAttempts))")
        
        DispatchQueue.main.asyncAfter(deadline: .now() + delay) {
            self.connect()
        }
    }
    
    // MARK: - Heartbeat
    
    private func startPingTimer() {
        stopPingTimer()
        pingTimer = Timer.scheduledTimer(withTimeInterval: 30.0, repeats: true) { [weak self] _ in
            self?.sendPing()
        }
    }
    
    private func stopPingTimer() {
        pingTimer?.invalidate()
        pingTimer = nil
    }
    
    private func sendPing() {
        webSocketTask?.sendPing { error in
            if let error = error {
                print("WebSocket: Ping failed: \(error)")
                // Connection might be dead, trigger reconnect?
            }
        }
    }
}

// MARK: - Models

struct WebSocketMessage: Codable, Identifiable {
    let id: String
    let type: MessageType
    let from: String?
    let to: String?
    let stationId: String?
    let content: String?
    let timestamp: Date
    
    enum CodingKeys: String, CodingKey {
        case id, type, from, to, content, timestamp
        case stationId = "station_id"
    }
}

struct ClientMessage: Codable {
    let type: MessageType
    let to: String?
    let content: String?
    
    init(type: MessageType, to: String? = nil, content: String? = nil) {
        self.type = type
        self.to = to
        self.content = content
    }
}

enum MessageType: String, Codable {
    case message
    case broadcast
    case typing
    case presence
    case join
    case leave
    case ping
    case pong
    case error
    case ack
}
