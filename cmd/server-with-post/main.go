package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	version = "0.1.0"
	banner  = `
╔══════════════════════════════════════════════════════╗
║                                                      ║
║     🚄 TrainBlink Server - Phase 0: Day 2           ║
║                                                      ║
║     Version: %s                                ║
║     Stage:   HTTP + POST Endpoints                  ║
║                                                      ║
╚══════════════════════════════════════════════════════╝
`
)

var (
	startTime     = time.Now()
	requestCount  = 0
	messageStore  = &MessageStore{messages: make([]Message, 0)}
	clientStore   = &ClientStore{clients: make(map[string]*Client)}
)

// Models
type Client struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DeviceID    string    `json:"device_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeen    time.Time `json:"last_seen"`
}

type Message struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // sent, delivered, read
}

type ClientStore struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func (cs *ClientStore) Add(client *Client) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.clients[client.ID] = client
	log.Printf("✅ Client added: %s (%s)", client.ID, client.Name)
}

func (cs *ClientStore) Get(id string) (*Client, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	client, ok := cs.clients[id]
	return client, ok
}

func (cs *ClientStore) Remove(id string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if _, ok := cs.clients[id]; ok {
		delete(cs.clients, id)
		log.Printf("❌ Client removed: %s", id)
	}
}

func (cs *ClientStore) List() []Client {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	clients := make([]Client, 0, len(cs.clients))
	for _, client := range cs.clients {
		clients = append(clients, *client)
	}
	return clients
}

func (cs *ClientStore) Count() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.clients)
}

type MessageStore struct {
	messages []Message
	mu       sync.RWMutex
}

func (ms *MessageStore) Add(message Message) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.messages = append(ms.messages, message)
	log.Printf("📨 Message stored: from=%s to=%s", message.From, message.To)
}

func (ms *MessageStore) GetForUser(userID string) []Message {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	messages := make([]Message, 0)
	for _, msg := range ms.messages {
		if msg.To == userID || msg.From == userID || msg.To == "all" {
			messages = append(messages, msg)
		}
	}
	return messages
}

func (ms *MessageStore) GetAll() []Message {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.messages
}

func (ms *MessageStore) Count() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.messages)
}

// Response types
type PingResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Server    string    `json:"server"`
}

type HealthResponse struct {
	Status         string    `json:"status"`
	Timestamp      time.Time `json:"timestamp"`
	Uptime         float64   `json:"uptime"`
	Clients        int       `json:"clients"`
	Messages       int       `json:"messages"`
	TotalRequests  int       `json:"total_requests"`
}

type HelloResponse struct {
	Message   string            `json:"message"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Features  []string          `json:"features"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Request types
type RegisterClientRequest struct {
	Name     string `json:"name"`
	DeviceID string `json:"device_id"`
}

type SendMessageRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestCount++
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s - %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

// Handlers
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "TrainBlink Server API",
		"version": version,
		"endpoints": map[string]string{
			"GET /ping":                "Health check",
			"GET /health":              "Detailed health status",
			"GET /api/v1/hello":        "Hello World",
			"GET /api/v1/welcome":      "Welcome message",
			"POST /api/v1/clients":     "Register client",
			"GET /api/v1/clients":      "List clients",
			"GET /api/v1/clients/{id}": "Get client",
			"POST /api/v1/messages":    "Send message",
			"GET /api/v1/messages":     "Get all messages",
			"GET /api/v1/messages/{id}": "Get messages for user",
		},
	})
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	response := PingResponse{
		Message:   "pong",
		Timestamp: time.Now(),
		Server:    fmt.Sprintf("TrainBlink Server v%s", version),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:        "healthy",
		Timestamp:     time.Now(),
		Uptime:        time.Since(startTime).Seconds(),
		Clients:       clientStore.Count(),
		Messages:      messageStore.Count(),
		TotalRequests: requestCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	response := HelloResponse{
		Message:   "Hello from TrainBlink Server!",
		Version:   version,
		Timestamp: time.Now(),
		Features: []string{
			"http",
			"client-management",
			"message-storage",
			"message-relay",
		},
		Metadata: map[string]string{
			"protocol": "TrainBlink Protocol v1.0",
			"stage":    "Phase 0 - Day 2 - POST Endpoints",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Anonymous"
	}

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		deviceID = "unknown"
	}

	response := map[string]interface{}{
		"message":   "Welcome to TrainBlink!",
		"user":      name,
		"device_id": deviceID,
		"timestamp": time.Now(),
		"tip":       "Use POST /api/v1/clients to register",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func registerClientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	client := &Client{
		ID:          fmt.Sprintf("client-%d", time.Now().UnixNano()),
		Name:        req.Name,
		DeviceID:    req.DeviceID,
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}

	clientStore.Add(client)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "registered",
		"client":  client,
		"message": fmt.Sprintf("Welcome %s! Your client ID is %s", client.Name, client.ID),
	})
}

func listClientsHandler(w http.ResponseWriter, r *http.Request) {
	clients := clientStore.List()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":     len(clients),
		"clients":   clients,
		"timestamp": time.Now(),
	})
}

func getClientHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/v1/clients/{id}
	// For simplicity, using query parameter
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	client, ok := clientStore.Get(id)
	if !ok {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate from client exists
	if _, ok := clientStore.Get(req.From); !ok && req.From != "system" {
		http.Error(w, fmt.Sprintf("Sender client %s not found", req.From), http.StatusBadRequest)
		return
	}

	// Validate to client exists (unless broadcasting)
	if req.To != "all" {
		if _, ok := clientStore.Get(req.To); !ok {
			http.Error(w, fmt.Sprintf("Recipient client %s not found", req.To), http.StatusNotFound)
			return
		}
	}

	message := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		From:      req.From,
		To:        req.To,
		Text:      req.Text,
		Timestamp: time.Now(),
		Status:    "sent",
	}

	messageStore.Add(message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "sent",
		"message": message,
	})
}

func listMessagesHandler(w http.ResponseWriter, r *http.Request) {
	messages := messageStore.GetAll()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":     len(messages),
		"messages":  messages,
		"timestamp": time.Now(),
	})
}

func getUserMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	messages := messageStore.GetForUser(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":   userID,
		"count":     len(messages),
		"messages":  messages,
		"timestamp": time.Now(),
	})
}

func main() {
	fmt.Printf(banner, version)

	mux := http.NewServeMux()

	// Root
	mux.HandleFunc("/", rootHandler)

	// Health
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/health", healthHandler)

	// API v1
	mux.HandleFunc("/api/v1/hello", helloHandler)
	mux.HandleFunc("/api/v1/welcome", welcomeHandler)

	// Clients
	mux.HandleFunc("/api/v1/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			registerClientHandler(w, r)
		} else {
			listClientsHandler(w, r)
		}
	})
	mux.HandleFunc("/api/v1/clients/get", getClientHandler)

	// Messages
	mux.HandleFunc("/api/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sendMessageHandler(w, r)
		} else {
			listMessagesHandler(w, r)
		}
	})
	mux.HandleFunc("/api/v1/messages/user", getUserMessagesHandler)

	// Apply middleware
	handler := corsMiddleware(loggingMiddleware(mux))

	server := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("🚀 Server is running on http://localhost%s", server.Addr)
		log.Println("")
		log.Println("HTTP Endpoints:")
		log.Println("  • GET  http://localhost:8080/ping")
		log.Println("  • GET  http://localhost:8080/health")
		log.Println("  • GET  http://localhost:8080/api/v1/hello")
		log.Println("  • POST http://localhost:8080/api/v1/clients - Register client")
		log.Println("  • GET  http://localhost:8080/api/v1/clients - List clients")
		log.Println("  • POST http://localhost:8080/api/v1/messages - Send message")
		log.Println("  • GET  http://localhost:8080/api/v1/messages - List messages")
		log.Println("")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏹️  Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited successfully")
}
