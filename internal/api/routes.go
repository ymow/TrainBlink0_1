package api

import (
	"github.com/gin-gonic/gin"

	"github.com/ymow/messenger_protocol_research/internal/api/handlers"
	"github.com/ymow/messenger_protocol_research/internal/api/middleware"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(router *gin.Engine) {
	// Apply global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "TrainBlink Server API",
			"version": "0.1.0",
			"docs":    "/api/v1/hello",
		})
	})

	// Health checks
	router.GET("/ping", handlers.PingHandler)
	router.GET("/health", handlers.HealthHandler)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Hello endpoint
		v1.GET("/hello", handlers.HelloHandler)
		v1.GET("/welcome", handlers.WelcomeHandler)

		// TODO: Add more endpoints in future phases
		// v1.POST("/auth/anonymous", handlers.AnonymousAuthHandler)
		// v1.GET("/stations", handlers.GetStationsHandler)
		// v1.POST("/stations/:id/join", handlers.JoinStationHandler)
	}

	// WebSocket endpoint (Phase 0 - Day 2)
	// router.GET("/ws", handlers.WebSocketHandler)
}
