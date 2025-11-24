package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	version = "0.1.0"
	banner  = `
╔══════════════════════════════════════════════════════╗
║                                                      ║
║     🚄 TrainBlink Server - Phase 0: Hello World     ║
║                                                      ║
║     Version: %s                                ║
║     Stage:   HTTP Foundation (Standard Library)     ║
║                                                      ║
╚══════════════════════════════════════════════════════╝
`
)

var startTime = time.Now()

// Response types
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
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    float64   `json:"uptime"`
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
		"docs":    "/api/v1/hello",
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
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).Seconds(),
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
			"geofencing",
			"p2p-messaging",
		},
		Metadata: map[string]string{
			"protocol": "TrainBlink Protocol v1.0",
			"stage":    "Phase 0 - Hello World",
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
		"tip":       "Connect to a train station to start chatting!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		log.Println("Try these endpoints:")
		log.Println("  • GET  http://localhost:8080/ping")
		log.Println("  • GET  http://localhost:8080/health")
		log.Println("  • GET  http://localhost:8080/api/v1/hello")
		log.Println("  • GET  http://localhost:8080/api/v1/welcome?name=Alice")

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
