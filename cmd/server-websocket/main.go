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

	"github.com/gorilla/websocket"
)

const (
	version = "0.1.0"
	banner  = `
╔══════════════════════════════════════════════════════╗
║                                                      ║
║     🚄 TrainBlink Server - Phase 0: Day 2           ║
║                                                      ║
║     Version: %s                                ║
║     Stage:   HTTP + WebSocket                       ║
║                                                      ║
╚══════════════════════════════════════════════════════╝
`
)

var (
	startTime = time.Now()
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for testing
		},
	}

	// Connection manager
	connections = &ConnectionManager{
		clients: make(map[*Client]bool),
	}
)

// Client represents a WebSocket client
type Client struct {
	ID          string
	Conn        *websocket.Conn
	Send        chan []byte
	ConnectedAt time.Time
}

// ConnectionManager manages all WebSocket connections
type ConnectionManager struct {
	clients map[*Client]bool
	mu      sync.RWMutex
}

// Add adds a client to the manager
func (cm *ConnectionManager) Add(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[client] = true
	log.Printf("✅ Client connected: %s (Total: %d)", client.ID, len(cm.clients))
}

// Remove removes a client from the manager
func (cm *ConnectionManager) Remove(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.clients[client]; ok {
		delete(cm.clients, client)
		close(client.Send)
		log.Printf("❌ Client disconnected: %s (Total: %d)", client.ID, len(cm.clients))
	}
}

// Count returns the number of connected clients
func (cm *ConnectionManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.clients)
}

// Broadcast sends a message to all connected clients
func (cm *ConnectionManager) Broadcast(message []byte, exclude *Client) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for client := range cm.clients {
		if client != exclude {
			select {
			case client.Send <- message:
			default:
				// Client send buffer is full, skip
			}
		}
	}
}

// Message types
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type PostMessage struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}

// HTTP Response types
type PingResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Server    string    `json:"server"`
}

type HelloResponse struct {
	Message   string            `json:"message"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Features  []string          `json:"features"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type HealthResponse struct {
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
	Uptime        float64   `json:"uptime"`
	Connections   int       `json:"connections"`
	TotalRequests int       `json:"total_requests"`
}

type ConnectionsResponse struct {
	Count     int       `json:"count"`
	ClientIDs []string  `json:"client_ids"`
	Timestamp time.Time `json:"timestamp"`
}

var requestCount = 0

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

// HTTP Handlers
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "TrainBlink Server API",
		"version": version,
		"endpoints": map[string]string{
			"GET /ping":               "Health check",
			"GET /health":             "Detailed health status",
			"GET /api/v1/hello":       "Hello World",
			"GET /api/v1/welcome":     "Welcome message",
			"POST /api/v1/message":    "Send message to client",
			"GET /api/v1/connections": "Get active connections",
			"WS /ws":                  "WebSocket connection",
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
		Connections:   connections.Count(),
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
			"websocket",
			"echo",
			"broadcast",
			"message-relay",
		},
		Metadata: map[string]string{
			"protocol": "TrainBlink Protocol v1.0",
			"stage":    "Phase 0 - Day 2 - WebSocket",
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
		"tip":       "Connect via WebSocket at ws://localhost:8080/ws",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func connectionsHandler(w http.ResponseWriter, r *http.Request) {
	connections.mu.RLock()
	defer connections.mu.RUnlock()

	clientIDs := make([]string, 0, len(connections.clients))
	for client := range connections.clients {
		clientIDs = append(clientIDs, client.ID)
	}

	response := ConnectionsResponse{
		Count:     len(clientIDs),
		ClientIDs: clientIDs,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg PostMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create WebSocket message
	wsMsg := WebSocketMessage{
		Type:      "message",
		From:      msg.From,
		To:        msg.To,
		Message:   msg.Message,
		Timestamp: time.Now(),
	}

	msgBytes, _ := json.Marshal(wsMsg)

	// Send to specific client or broadcast
	if msg.To == "" || msg.To == "all" {
		connections.Broadcast(msgBytes, nil)
		log.Printf("📢 Broadcast message from %s: %s", msg.From, msg.Message)
	} else {
		// Find specific client and send
		sent := false
		connections.mu.RLock()
		for client := range connections.clients {
			if client.ID == msg.To {
				select {
				case client.Send <- msgBytes:
					sent = true
					log.Printf("📨 Sent message from %s to %s: %s", msg.From, msg.To, msg.Message)
				default:
					log.Printf("⚠️ Failed to send to %s (buffer full)", msg.To)
				}
				break
			}
		}
		connections.mu.RUnlock()

		if !sent {
			http.Error(w, fmt.Sprintf("Client %s not found", msg.To), http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "sent",
		"from":      msg.From,
		"to":        msg.To,
		"message":   msg.Message,
		"timestamp": time.Now(),
	})
}

// WebSocket Handlers
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	client := &Client{
		ID:          clientID,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		ConnectedAt: time.Now(),
	}

	// Register client
	connections.Add(client)

	// Start goroutines
	go client.writePump()
	go client.readPump()

	// Send welcome message
	welcomeMsg := WebSocketMessage{
		Type:      "welcome",
		Message:   fmt.Sprintf("Welcome! Your ID is: %s", clientID),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"client_id":      clientID,
			"server_version": version,
		},
	}
	welcomeBytes, _ := json.Marshal(welcomeMsg)
	client.Send <- welcomeBytes
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		connections.Remove(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			break
		}

		log.Printf("📩 Received from %s: %s", c.ID, string(message))

		// Parse message
		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("⚠️ Invalid JSON from %s", c.ID)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "echo":
			// Echo back to sender
			echoMsg := WebSocketMessage{
				Type:      "echo",
				Message:   fmt.Sprintf("Echo: %s", msg.Message),
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"original": msg.Message,
				},
			}
			echoBytes, _ := json.Marshal(echoMsg)
			c.Send <- echoBytes

		case "broadcast":
			// Broadcast to all other clients
			broadcastMsg := WebSocketMessage{
				Type:      "broadcast",
				From:      c.ID,
				Message:   msg.Message,
				Timestamp: time.Now(),
			}
			broadcastBytes, _ := json.Marshal(broadcastMsg)
			connections.Broadcast(broadcastBytes, c)
			log.Printf("📢 Broadcast from %s: %s", c.ID, msg.Message)

		case "message":
			// Send to specific client
			if msg.To == "" {
				// No recipient, echo back
				c.Send <- message
			} else {
				// Find recipient
				sent := false
				connections.mu.RLock()
				for client := range connections.clients {
					if client.ID == msg.To {
						relayMsg := WebSocketMessage{
							Type:      "message",
							From:      c.ID,
							Message:   msg.Message,
							Timestamp: time.Now(),
						}
						relayBytes, _ := json.Marshal(relayMsg)
						select {
						case client.Send <- relayBytes:
							sent = true
							log.Printf("📨 Relayed message from %s to %s", c.ID, msg.To)
						default:
							log.Printf("⚠️ Failed to send to %s (buffer full)", msg.To)
						}
						break
					}
				}
				connections.mu.RUnlock()

				// Send confirmation
				confirmMsg := WebSocketMessage{
					Type:      "sent",
					Message:   fmt.Sprintf("Message sent to %s", msg.To),
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"to":     msg.To,
						"status": map[string]bool{"delivered": sent},
					},
				}
				confirmBytes, _ := json.Marshal(confirmMsg)
				c.Send <- confirmBytes
			}

		default:
			// Unknown type, echo back
			c.Send <- message
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func main() {
	// Print banner
	fmt.Printf(banner, version)

	// Setup routes
	mux := http.NewServeMux()

	// Root endpoint
	mux.HandleFunc("/", rootHandler)

	// Health checks
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/health", healthHandler)

	// API v1
	mux.HandleFunc("/api/v1/hello", helloHandler)
	mux.HandleFunc("/api/v1/welcome", welcomeHandler)
	mux.HandleFunc("/api/v1/connections", connectionsHandler)
	mux.HandleFunc("/api/v1/message", sendMessageHandler)

	// WebSocket
	mux.HandleFunc("/ws", wsHandler)

	// Apply middleware
	handler := corsMiddleware(loggingMiddleware(mux))

	// Create server
	server := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server is running on http://localhost%s", server.Addr)
		log.Println("")
		log.Println("HTTP Endpoints:")
		log.Println("  • GET  http://localhost:8080/ping")
		log.Println("  • GET  http://localhost:8080/health")
		log.Println("  • GET  http://localhost:8080/api/v1/hello")
		log.Println("  • GET  http://localhost:8080/api/v1/connections")
		log.Println("  • POST http://localhost:8080/api/v1/message")
		log.Println("")
		log.Println("WebSocket:")
		log.Println("  • WS   ws://localhost:8080/ws")
		log.Println("")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏹️  Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited successfully")
}
