package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HelloResponse represents the hello response
type HelloResponse struct {
	Message   string            `json:"message"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Features  []string          `json:"features"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// HelloHandler handles hello requests
func HelloHandler(c *gin.Context) {
	response := HelloResponse{
		Message:   "Hello from TrainBlink Server!",
		Version:   "0.1.0",
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

	c.JSON(http.StatusOK, response)
}

// WelcomeHandler handles welcome requests with user info
func WelcomeHandler(c *gin.Context) {
	// Get optional query parameters
	name := c.DefaultQuery("name", "Anonymous")
	deviceID := c.DefaultQuery("device_id", "unknown")

	response := gin.H{
		"message":   "Welcome to TrainBlink!",
		"user":      name,
		"device_id": deviceID,
		"timestamp": time.Now(),
		"tip":       "Connect to a train station to start chatting!",
	}

	c.JSON(http.StatusOK, response)
}
