package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// PingResponse represents the ping response
type PingResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Server    string    `json:"server"`
}

// PingHandler handles ping requests
func PingHandler(c *gin.Context) {
	response := PingResponse{
		Message:   "pong",
		Timestamp: time.Now(),
		Server:    "TrainBlink Server v0.1.0",
	}

	c.JSON(http.StatusOK, response)
}

// HealthHandler handles health check requests
func HealthHandler(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime).Seconds(),
	}

	c.JSON(http.StatusOK, health)
}

var startTime = time.Now()
