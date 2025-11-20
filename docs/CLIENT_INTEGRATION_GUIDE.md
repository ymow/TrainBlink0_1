# TrainBlink 客戶端整合指南

**版本**: 1.0  
**最後更新**: 2025-11-20  
**適用平台**: iOS (Swift), Android (Kotlin)

---

## 📖 概述

本文檔提供 iOS 和 Android 客戶端整合 TrainBlink 通訊協議的詳細指南，包括代碼示例、最佳實踐和常見問題解決方案。

---

## 🍎 iOS 客戶端整合 (Swift)

### Phase 0: HTTP REST API

#### 1. 基礎設置

**安裝依賴**:
```swift
// 不需要額外依賴，使用系統 URLSession
```

**網絡層** (`NetworkManager.swift`):
```swift
import Foundation

class NetworkManager {
    static let shared = NetworkManager()
    
    private let baseURL = "http://localhost:8080"
    private var clientID: String?
    
    private init() {}
    
    // MARK: - Generic Request
    
    func request<T: Decodable>(
        endpoint: String,
        method: String = "GET",
        body: Encodable? = nil
    ) async throws -> T {
        guard let url = URL(string: "\(baseURL)\(endpoint)") else {
            throw NetworkError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        if let body = body {
            request.httpBody = try JSONEncoder().encode(body)
        }
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw NetworkError.invalidResponse
        }
        
        guard (200...299).contains(httpResponse.statusCode) else {
            // 解析錯誤響應
            if let errorResponse = try? JSONDecoder().decode(ErrorResponse.self, from: data) {
                throw NetworkError.serverError(errorResponse.message)
            }
            throw NetworkError.httpError(httpResponse.statusCode)
        }
        
        return try JSONDecoder().decode(T.self, from: data)
    }
}

// MARK: - Error Types

enum NetworkError: Error {
    case invalidURL
    case invalidResponse
    case httpError(Int)
    case serverError(String)
    case decodingError
}

struct ErrorResponse: Decodable {
    let error: String
    let message: String
    let code: String
}
```

#### 2. 數據模型

**Client.swift**:
```swift
import Foundation

struct Client: Codable {
    let id: String
    let name: String
    let deviceID: String?
    let connectedAt: Date
    let lastSeen: Date
    
    enum CodingKeys: String, CodingKey {
        case id
        case name
        case deviceID = "device_id"
        case connectedAt = "connected_at"
        case lastSeen = "last_seen"
    }
}

struct RegisterClientRequest: Encodable {
    let name: String
    let deviceID: String?
    
    enum CodingKeys: String, CodingKey {
        case name
        case deviceID = "device_id"
    }
}

struct RegisterClientResponse: Decodable {
    let status: String
    let client: Client
    let message: String
}
```

**Message.swift**:
```swift
import Foundation

struct Message: Codable, Identifiable {
    let id: String
    let from: String
    let to: String
    let text: String
    let timestamp: Date
    let status: String
}

struct SendMessageRequest: Encodable {
    let from: String
    let to: String
    let text: String
}

struct SendMessageResponse: Decodable {
    let status: String
    let message: Message
}

struct UserMessagesResponse: Decodable {
    let userID: String
    let count: Int
    let messages: [Message]
    
    enum CodingKeys: String, CodingKey {
        case userID = "user_id"
        case count
        case messages
    }
}
```

#### 3. API 服務

**TrainBlinkAPI.swift**:
```swift
import Foundation

class TrainBlinkAPI {
    private let network = NetworkManager.shared
    private var currentClient: Client?
    
    // MARK: - Client Management
    
    func registerClient(name: String, deviceID: String? = nil) async throws -> Client {
        let request = RegisterClientRequest(name: name, deviceID: deviceID)
        let response: RegisterClientResponse = try await network.request(
            endpoint: "/api/v1/clients",
            method: "POST",
            body: request
        )
        
        currentClient = response.client
        // 保存到 UserDefaults
        saveClientID(response.client.id)
        
        return response.client
    }
    
    func getClients() async throws -> [Client] {
        struct Response: Decodable {
            let clients: [Client]
        }
        let response: Response = try await network.request(endpoint: "/api/v1/clients")
        return response.clients
    }
    
    // MARK: - Messaging
    
    func sendMessage(to: String, text: String) async throws -> Message {
        guard let clientID = currentClient?.id else {
            throw TrainBlinkError.notRegistered
        }
        
        let request = SendMessageRequest(from: clientID, to: to, text: text)
        let response: SendMessageResponse = try await network.request(
            endpoint: "/api/v1/messages",
            method: "POST",
            body: request
        )
        
        return response.message
    }
    
    func getMessages() async throws -> [Message] {
        guard let clientID = currentClient?.id else {
            throw TrainBlinkError.notRegistered
        }
        
        let response: UserMessagesResponse = try await network.request(
            endpoint: "/api/v1/messages/user?id=\(clientID)"
        )
        
        return response.messages
    }
    
    // MARK: - Persistence
    
    private func saveClientID(_ id: String) {
        UserDefaults.standard.set(id, forKey: "trainblink_client_id")
    }
    
    func loadClientID() -> String? {
        return UserDefaults.standard.string(forKey: "trainblink_client_id")
    }
}

enum TrainBlinkError: Error {
    case notRegistered
}
```

#### 4. SwiftUI 視圖

**ContentView.swift**:
```swift
import SwiftUI

struct ContentView: View {
    @StateObject private var viewModel = ChatViewModel()
    
    var body: some View {
        NavigationView {
            VStack {
                if viewModel.currentClient == nil {
                    // 註冊視圖
                    RegistrationView(viewModel: viewModel)
                } else {
                    // 聊天視圖
                    ChatView(viewModel: viewModel)
                }
            }
            .navigationTitle("TrainBlink")
        }
    }
}

struct RegistrationView: View {
    @ObservedObject var viewModel: ChatViewModel
    @State private var name: String = ""
    
    var body: some View {
        VStack(spacing: 20) {
            TextField("Enter your name", text: $name)
                .textFieldStyle(RoundedBorderTextFieldStyle())
                .padding()
            
            Button("Register") {
                Task {
                    await viewModel.register(name: name)
                }
            }
            .disabled(name.isEmpty)
            .buttonStyle(.borderedProminent)
        }
        .padding()
    }
}

struct ChatView: View {
    @ObservedObject var viewModel: ChatViewModel
    @State private var messageText: String = ""
    
    var body: some View {
        VStack {
            // 消息列表
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 12) {
                    ForEach(viewModel.messages) { message in
                        MessageRow(message: message, currentClientID: viewModel.currentClient?.id)
                    }
                }
                .padding()
            }
            
            // 輸入框
            HStack {
                TextField("Type a message...", text: $messageText)
                    .textFieldStyle(RoundedBorderTextFieldStyle())
                
                Button("Send") {
                    Task {
                        await viewModel.sendMessage(text: messageText)
                        messageText = ""
                    }
                }
                .disabled(messageText.isEmpty)
            }
            .padding()
        }
        .onAppear {
            Task {
                await viewModel.loadMessages()
            }
        }
    }
}

struct MessageRow: View {
    let message: Message
    let currentClientID: String?
    
    var isFromMe: Bool {
        message.from == currentClientID
    }
    
    var body: some View {
        HStack {
            if isFromMe { Spacer() }
            
            VStack(alignment: isFromMe ? .trailing : .leading, spacing: 4) {
                Text(message.text)
                    .padding(10)
                    .background(isFromMe ? Color.blue : Color.gray.opacity(0.2))
                    .foregroundColor(isFromMe ? .white : .primary)
                    .cornerRadius(12)
                
                Text(message.timestamp, style: .time)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            
            if !isFromMe { Spacer() }
        }
    }
}
```

**ChatViewModel.swift**:
```swift
import Foundation
import Combine

@MainActor
class ChatViewModel: ObservableObject {
    @Published var currentClient: Client?
    @Published var messages: [Message] = []
    @Published var peers: [Client] = []
    @Published var isLoading = false
    @Published var errorMessage: String?
    
    private let api = TrainBlinkAPI()
    
    // MARK: - Registration
    
    func register(name: String) async {
        isLoading = true
        defer { isLoading = false }
        
        do {
            let deviceID = UIDevice.current.identifierForVendor?.uuidString
            currentClient = try await api.registerClient(name: name, deviceID: deviceID)
            await loadPeers()
        } catch {
            errorMessage = "Registration failed: \(error.localizedDescription)"
        }
    }
    
    // MARK: - Messaging
    
    func sendMessage(text: String, to: String = "all") async {
        do {
            let message = try await api.sendMessage(to: to, text: text)
            messages.append(message)
        } catch {
            errorMessage = "Failed to send message: \(error.localizedDescription)"
        }
    }
    
    func loadMessages() async {
        do {
            messages = try await api.getMessages()
        } catch {
            errorMessage = "Failed to load messages: \(error.localizedDescription)"
        }
    }
    
    // MARK: - Peers
    
    func loadPeers() async {
        do {
            peers = try await api.getClients()
        } catch {
            errorMessage = "Failed to load peers: \(error.localizedDescription)"
        }
    }
}
```

---

### Phase 1: WebSocket

#### 1. WebSocket 管理器

**安裝依賴**:
```ruby
# Podfile
pod 'Starscream', '~> 4.0'
```

**WebSocketManager.swift**:
```swift
import Foundation
import Starscream

class WebSocketManager: WebSocketDelegate {
    static let shared = WebSocketManager()
    
    private var socket: WebSocket?
    private var isConnected = false
    private var reconnectAttempt = 0
    private let maxReconnectAttempt = 6
    
    var onMessageReceived: ((WSMessage) -> Void)?
    var onConnected: (() -> Void)?
    var onDisconnected: (() -> Void)?
    
    private init() {}
    
    // MARK: - Connection
    
    func connect(clientID: String) {
        let urlString = "ws://localhost:8080/ws?client_id=\(clientID)"
        guard let url = URL(string: urlString) else { return }
        
        var request = URLRequest(url: url)
        request.timeoutInterval = 30
        
        socket = WebSocket(request: request)
        socket?.delegate = self
        socket?.connect()
    }
    
    func disconnect() {
        socket?.disconnect()
        socket = nil
        isConnected = false
    }
    
    // MARK: - Send Messages
    
    func sendMessage(to: String, text: String) {
        let message = WSOutgoingMessage(
            type: "send_message",
            payload: SendMessagePayload(to: to, text: text),
            id: UUID().uuidString
        )
        
        if let data = try? JSONEncoder().encode(message),
           let jsonString = String(data: data, encoding: .utf8) {
            socket?.write(string: jsonString)
        }
    }
    
    func sendTypingIndicator(to: String, isTyping: Bool) {
        let message = WSOutgoingMessage(
            type: "typing",
            payload: TypingPayload(to: to, isTyping: isTyping)
        )
        
        if let data = try? JSONEncoder().encode(message),
           let jsonString = String(data: data, encoding: .utf8) {
            socket?.write(string: jsonString)
        }
    }
    
    // MARK: - WebSocketDelegate
    
    func didReceive(event: Starscream.WebSocketEvent, client: Starscream.WebSocketClient) {
        switch event {
        case .connected(_):
            print("✅ WebSocket Connected")
            isConnected = true
            reconnectAttempt = 0
            onConnected?()
            startHeartbeat()
            
        case .disconnected(let reason, let code):
            print("❌ WebSocket Disconnected: \(reason), code: \(code)")
            isConnected = false
            onDisconnected?()
            stopHeartbeat()
            attemptReconnect()
            
        case .text(let string):
            handleIncomingMessage(string)
            
        case .error(let error):
            print("⚠️ WebSocket Error: \(String(describing: error))")
            
        default:
            break
        }
    }
    
    // MARK: - Message Handling
    
    private func handleIncomingMessage(_ text: String) {
        guard let data = text.data(using: .utf8) else { return }
        
        do {
            let message = try JSONDecoder().decode(WSMessage.self, from: data)
            onMessageReceived?(message)
        } catch {
            print("Failed to decode message: \(error)")
        }
    }
    
    // MARK: - Heartbeat
    
    private var heartbeatTimer: Timer?
    
    private func startHeartbeat() {
        heartbeatTimer = Timer.scheduledTimer(withTimeInterval: 30, repeats: true) { [weak self] _ in
            self?.sendPing()
        }
    }
    
    private func stopHeartbeat() {
        heartbeatTimer?.invalidate()
        heartbeatTimer = nil
    }
    
    private func sendPing() {
        let ping = WSOutgoingMessage(
            type: "ping",
            payload: EmptyPayload(),
            timestamp: ISO8601DateFormatter().string(from: Date())
        )
        
        if let data = try? JSONEncoder().encode(ping),
           let jsonString = String(data: data, encoding: .utf8) {
            socket?.write(string: jsonString)
        }
    }
    
    // MARK: - Reconnection
    
    private func attemptReconnect() {
        guard reconnectAttempt < maxReconnectAttempt else {
            print("Max reconnection attempts reached")
            return
        }
        
        let delays = [1, 2, 4, 8, 16, 30]
        let delay = delays[min(reconnectAttempt, delays.count - 1)]
        
        reconnectAttempt += 1
        
        print("Reconnecting in \(delay) seconds... (attempt \(reconnectAttempt))")
        
        DispatchQueue.main.asyncAfter(deadline: .now() + .seconds(delay)) { [weak self] in
            guard let self = self, !self.isConnected else { return }
            // 需要 clientID，從其他地方獲取
            // self.connect(clientID: savedClientID)
        }
    }
}

// MARK: - Message Types

struct WSMessage: Decodable {
    let type: String
    let payload: PayloadData
    let timestamp: String?
    let id: String?
}

struct PayloadData: Decodable {
    // 使用 JSONValue 處理動態 payload
    let json: [String: Any]
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        json = try container.decode([String: Any].self)
    }
}

struct WSOutgoingMessage<T: Encodable>: Encodable {
    let type: String
    let payload: T
    let timestamp: String?
    let id: String?
    
    init(type: String, payload: T, timestamp: String? = nil, id: String? = nil) {
        self.type = type
        self.payload = payload
        self.timestamp = timestamp
        self.id = id
    }
}

struct SendMessagePayload: Encodable {
    let to: String
    let text: String
    let isEphemeral: Bool
    let lifetimeSeconds: Int?
    
    enum CodingKeys: String, CodingKey {
        case to, text
        case isEphemeral = "is_ephemeral"
        case lifetimeSeconds = "lifetime_seconds"
    }
    
    init(to: String, text: String, isEphemeral: Bool = false, lifetimeSeconds: Int? = nil) {
        self.to = to
        self.text = text
        self.isEphemeral = isEphemeral
        self.lifetimeSeconds = lifetimeSeconds
    }
}

struct TypingPayload: Encodable {
    let to: String
    let isTyping: Bool
    
    enum CodingKeys: String, CodingKey {
        case to
        case isTyping = "is_typing"
    }
}

struct EmptyPayload: Encodable {}
```

---

## 🤖 Android 客戶端整合 (Kotlin)

### Phase 0: HTTP REST API

#### 1. 基礎設置

**依賴** (`build.gradle.kts`):
```kotlin
dependencies {
    // Retrofit for HTTP
    implementation("com.squareup.retrofit2:retrofit:2.9.0")
    implementation("com.squareup.retrofit2:converter-gson:2.9.0")
    implementation("com.squareup.okhttp3:logging-interceptor:4.11.0")
    
    // Coroutines
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3")
    
    // ViewModel
    implementation("androidx.lifecycle:lifecycle-viewmodel-ktx:2.6.2")
    implementation("androidx.lifecycle:lifecycle-runtime-ktx:2.6.2")
}
```

#### 2. 網絡層

**ApiService.kt**:
```kotlin
import retrofit2.Response
import retrofit2.http.*

interface ApiService {
    @POST("/api/v1/clients")
    suspend fun registerClient(
        @Body request: RegisterClientRequest
    ): Response<RegisterClientResponse>
    
    @GET("/api/v1/clients")
    suspend fun getClients(): Response<ClientsResponse>
    
    @POST("/api/v1/messages")
    suspend fun sendMessage(
        @Body request: SendMessageRequest
    ): Response<SendMessageResponse>
    
    @GET("/api/v1/messages/user")
    suspend fun getUserMessages(
        @Query("id") clientId: String
    ): Response<UserMessagesResponse>
    
    @GET("/health")
    suspend fun health(): Response<HealthResponse>
}
```

**RetrofitClient.kt**:
```kotlin
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

object RetrofitClient {
    private const val BASE_URL = "http://10.0.2.2:8080"  // Android emulator localhost
    
    private val loggingInterceptor = HttpLoggingInterceptor().apply {
        level = HttpLoggingInterceptor.Level.BODY
    }
    
    private val okHttpClient = OkHttpClient.Builder()
        .addInterceptor(loggingInterceptor)
        .connectTimeout(30, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .writeTimeout(30, TimeUnit.SECONDS)
        .build()
    
    val apiService: ApiService by lazy {
        Retrofit.Builder()
            .baseUrl(BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(ApiService::class.java)
    }
}
```

#### 3. 數據模型

**Models.kt**:
```kotlin
import com.google.gson.annotations.SerializedName
import java.util.Date

data class Client(
    val id: String,
    val name: String,
    @SerializedName("device_id") val deviceId: String?,
    @SerializedName("connected_at") val connectedAt: Date,
    @SerializedName("last_seen") val lastSeen: Date
)

data class RegisterClientRequest(
    val name: String,
    @SerializedName("device_id") val deviceId: String? = null
)

data class RegisterClientResponse(
    val status: String,
    val client: Client,
    val message: String
)

data class ClientsResponse(
    val count: Int,
    val clients: List<Client>,
    val timestamp: Date
)

data class Message(
    val id: String,
    val from: String,
    val to: String,
    val text: String,
    val timestamp: Date,
    val status: String
)

data class SendMessageRequest(
    val from: String,
    val to: String,
    val text: String
)

data class SendMessageResponse(
    val status: String,
    val message: Message
)

data class UserMessagesResponse(
    @SerializedName("user_id") val userId: String,
    val count: Int,
    val messages: List<Message>
)

data class HealthResponse(
    val status: String,
    val uptime: Double,
    val clients: Int,
    val messages: Int,
    @SerializedName("total_requests") val totalRequests: Int
)
```

#### 4. Repository

**TrainBlinkRepository.kt**:
```kotlin
import android.content.Context
import android.content.SharedPreferences
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class TrainBlinkRepository(context: Context) {
    private val api = RetrofitClient.apiService
    private val prefs: SharedPreferences = context.getSharedPreferences(
        "trainblink_prefs",
        Context.MODE_PRIVATE
    )
    
    private var currentClient: Client? = null
    
    // MARK: - Client Management
    
    suspend fun registerClient(name: String, deviceId: String? = null): Result<Client> {
        return withContext(Dispatchers.IO) {
            try {
                val request = RegisterClientRequest(name, deviceId)
                val response = api.registerClient(request)
                
                if (response.isSuccessful && response.body() != null) {
                    val client = response.body()!!.client
                    currentClient = client
                    saveClientId(client.id)
                    Result.success(client)
                } else {
                    Result.failure(Exception("Registration failed: ${response.code()}"))
                }
            } catch (e: Exception) {
                Result.failure(e)
            }
        }
    }
    
    suspend fun getClients(): Result<List<Client>> {
        return withContext(Dispatchers.IO) {
            try {
                val response = api.getClients()
                if (response.isSuccessful && response.body() != null) {
                    Result.success(response.body()!!.clients)
                } else {
                    Result.failure(Exception("Failed to get clients: ${response.code()}"))
                }
            } catch (e: Exception) {
                Result.failure(e)
            }
        }
    }
    
    // MARK: - Messaging
    
    suspend fun sendMessage(to: String, text: String): Result<Message> {
        return withContext(Dispatchers.IO) {
            try {
                val clientId = currentClient?.id ?: getClientId()
                    ?: return@withContext Result.failure(Exception("Not registered"))
                
                val request = SendMessageRequest(from = clientId, to = to, text = text)
                val response = api.sendMessage(request)
                
                if (response.isSuccessful && response.body() != null) {
                    Result.success(response.body()!!.message)
                } else {
                    Result.failure(Exception("Failed to send message: ${response.code()}"))
                }
            } catch (e: Exception) {
                Result.failure(e)
            }
        }
    }
    
    suspend fun getMessages(): Result<List<Message>> {
        return withContext(Dispatchers.IO) {
            try {
                val clientId = currentClient?.id ?: getClientId()
                    ?: return@withContext Result.failure(Exception("Not registered"))
                
                val response = api.getUserMessages(clientId)
                
                if (response.isSuccessful && response.body() != null) {
                    Result.success(response.body()!!.messages)
                } else {
                    Result.failure(Exception("Failed to get messages: ${response.code()}"))
                }
            } catch (e: Exception) {
                Result.failure(e)
            }
        }
    }
    
    // MARK: - Persistence
    
    private fun saveClientId(id: String) {
        prefs.edit().putString("client_id", id).apply()
    }
    
    private fun getClientId(): String? {
        return prefs.getString("client_id", null)
    }
}
```

#### 5. ViewModel

**ChatViewModel.kt**:
```kotlin
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

class ChatViewModel(private val repository: TrainBlinkRepository) : ViewModel() {
    
    private val _currentClient = MutableStateFlow<Client?>(null)
    val currentClient: StateFlow<Client?> = _currentClient
    
    private val _messages = MutableStateFlow<List<Message>>(emptyList())
    val messages: StateFlow<List<Message>> = _messages
    
    private val _peers = MutableStateFlow<List<Client>>(emptyList())
    val peers: StateFlow<List<Client>> = _peers
    
    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading
    
    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage
    
    // MARK: - Registration
    
    fun register(name: String, deviceId: String? = null) {
        viewModelScope.launch {
            _isLoading.value = true
            
            repository.registerClient(name, deviceId).fold(
                onSuccess = { client ->
                    _currentClient.value = client
                    loadPeers()
                },
                onFailure = { error ->
                    _errorMessage.value = "Registration failed: ${error.message}"
                }
            )
            
            _isLoading.value = false
        }
    }
    
    // MARK: - Messaging
    
    fun sendMessage(text: String, to: String = "all") {
        viewModelScope.launch {
            repository.sendMessage(to, text).fold(
                onSuccess = { message ->
                    _messages.value = _messages.value + message
                },
                onFailure = { error ->
                    _errorMessage.value = "Failed to send message: ${error.message}"
                }
            )
        }
    }
    
    fun loadMessages() {
        viewModelScope.launch {
            repository.getMessages().fold(
                onSuccess = { messages ->
                    _messages.value = messages
                },
                onFailure = { error ->
                    _errorMessage.value = "Failed to load messages: ${error.message}"
                }
            )
        }
    }
    
    // MARK: - Peers
    
    fun loadPeers() {
        viewModelScope.launch {
            repository.getClients().fold(
                onSuccess = { clients ->
                    _peers.value = clients
                },
                onFailure = { error ->
                    _errorMessage.value = "Failed to load peers: ${error.message}"
                }
            )
        }
    }
}
```

#### 6. Jetpack Compose UI

**MainActivity.kt**:
```kotlin
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            TrainBlinkTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    ChatScreen()
                }
            }
        }
    }
}

@Composable
fun ChatScreen(
    viewModel: ChatViewModel = viewModel(
        factory = ChatViewModelFactory(TrainBlinkRepository(LocalContext.current))
    )
) {
    val currentClient by viewModel.currentClient.collectAsState()
    val messages by viewModel.messages.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    
    if (currentClient == null) {
        RegistrationScreen(
            onRegister = { name -> viewModel.register(name) },
            isLoading = isLoading
        )
    } else {
        ChatMessagesScreen(
            messages = messages,
            currentClientId = currentClient?.id,
            onSendMessage = { text -> viewModel.sendMessage(text) },
            onLoadMessages = { viewModel.loadMessages() }
        )
    }
}

@Composable
fun RegistrationScreen(
    onRegister: (String) -> Unit,
    isLoading: Boolean
) {
    var name by remember { mutableStateOf("") }
    
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        OutlinedTextField(
            value = name,
            onValueChange = { name = it },
            label = { Text("Enter your name") },
            modifier = Modifier.fillMaxWidth()
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Button(
            onClick = { onRegister(name) },
            enabled = name.isNotBlank() && !isLoading,
            modifier = Modifier.fillMaxWidth()
        ) {
            if (isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = MaterialTheme.colorScheme.onPrimary
                )
            } else {
                Text("Register")
            }
        }
    }
}

@Composable
fun ChatMessagesScreen(
    messages: List<Message>,
    currentClientId: String?,
    onSendMessage: (String) -> Unit,
    onLoadMessages: () -> Unit
) {
    var messageText by remember { mutableStateOf("") }
    
    LaunchedEffect(Unit) {
        onLoadMessages()
    }
    
    Column(modifier = Modifier.fillMaxSize()) {
        // Messages list
        LazyColumn(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .padding(16.dp),
            reverseLayout = false
        ) {
            items(messages) { message ->
                MessageItem(
                    message = message,
                    isFromMe = message.from == currentClientId
                )
                Spacer(modifier = Modifier.height(8.dp))
            }
        }
        
        // Input field
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            OutlinedTextField(
                value = messageText,
                onValueChange = { messageText = it },
                placeholder = { Text("Type a message...") },
                modifier = Modifier.weight(1f)
            )
            
            Spacer(modifier = Modifier.width(8.dp))
            
            Button(
                onClick = {
                    if (messageText.isNotBlank()) {
                        onSendMessage(messageText)
                        messageText = ""
                    }
                },
                enabled = messageText.isNotBlank()
            ) {
                Text("Send")
            }
        }
    }
}

@Composable
fun MessageItem(message: Message, isFromMe: Boolean) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = if (isFromMe) Arrangement.End else Arrangement.Start
    ) {
        Card(
            colors = CardDefaults.cardColors(
                containerColor = if (isFromMe) 
                    MaterialTheme.colorScheme.primary 
                else 
                    MaterialTheme.colorScheme.surfaceVariant
            )
        ) {
            Column(modifier = Modifier.padding(12.dp)) {
                Text(
                    text = message.text,
                    color = if (isFromMe) 
                        MaterialTheme.colorScheme.onPrimary 
                    else 
                        MaterialTheme.colorScheme.onSurfaceVariant
                )
                
                Spacer(modifier = Modifier.height(4.dp))
                
                Text(
                    text = message.timestamp.toString(),
                    style = MaterialTheme.typography.labelSmall,
                    color = if (isFromMe) 
                        MaterialTheme.colorScheme.onPrimary.copy(alpha = 0.7f)
                    else 
                        MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f)
                )
            }
        }
    }
}
```

---

## 通用最佳實踐

### 1. 錯誤處理

```swift
// iOS
do {
    let message = try await api.sendMessage(to: "all", text: "Hello")
} catch NetworkError.httpError(let code) {
    print("HTTP Error: \(code)")
} catch NetworkError.serverError(let message) {
    print("Server Error: \(message)")
} catch {
    print("Unknown Error: \(error)")
}
```

```kotlin
// Android
repository.sendMessage("all", "Hello").fold(
    onSuccess = { message ->
        // 成功處理
    },
    onFailure = { error ->
        when (error) {
            is IOException -> {
                // 網絡錯誤
            }
            is HttpException -> {
                // HTTP 錯誤
            }
            else -> {
                // 其他錯誤
            }
        }
    }
)
```

### 2. 線程管理

```swift
// iOS - 使用 async/await
Task {
    let messages = try await api.getMessages()
    await MainActor.run {
        self.messages = messages
    }
}
```

```kotlin
// Android - 使用 Coroutines
viewModelScope.launch {
    val result = withContext(Dispatchers.IO) {
        repository.getMessages()
    }
    // 自動回到主線程
    _messages.value = result.getOrDefault(emptyList())
}
```

### 3. 數據持久化

```swift
// iOS - UserDefaults
UserDefaults.standard.set(clientID, forKey: "client_id")
let clientID = UserDefaults.standard.string(forKey: "client_id")
```

```kotlin
// Android - SharedPreferences
val prefs = context.getSharedPreferences("trainblink_prefs", Context.MODE_PRIVATE)
prefs.edit().putString("client_id", clientId).apply()
val clientId = prefs.getString("client_id", null)
```

---

## 常見問題

### iOS

**Q: 如何處理 App 進入後台？**
```swift
NotificationCenter.default.addObserver(
    forName: UIApplication.didEnterBackgroundNotification,
    object: nil,
    queue: .main
) { _ in
    // 斷開 WebSocket
    WebSocketManager.shared.disconnect()
}
```

**Q: 如何處理網絡變化？**
```swift
import Network

let monitor = NWPathMonitor()
monitor.pathUpdateHandler = { path in
    if path.status == .satisfied {
        // 網絡可用，重連
    } else {
        // 網絡不可用
    }
}
```

### Android

**Q: 如何處理 Activity 生命週期？**
```kotlin
override fun onPause() {
    super.onPause()
    // 斷開 WebSocket
    webSocketManager.disconnect()
}

override fun onResume() {
    super.onResume()
    // 重連 WebSocket
    webSocketManager.connect(clientId)
}
```

**Q: 如何處理網絡權限？**
```xml
<!-- AndroidManifest.xml -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

---

**下一部分**: Phase 1 WebSocket 整合將在後續更新。

---

**文檔維護者**: TrainBlink Team  
**最後更新**: 2025-11-20  
**版本**: 1.0
