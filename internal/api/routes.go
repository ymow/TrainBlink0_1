package api

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/ymow/messenger_protocol_research/internal/api/handlers"
	"github.com/ymow/messenger_protocol_research/internal/api/middleware"
	"github.com/ymow/messenger_protocol_research/internal/cleanup"
	"github.com/ymow/messenger_protocol_research/internal/discovery"
	"github.com/ymow/messenger_protocol_research/internal/matrix"
	"github.com/ymow/messenger_protocol_research/internal/message"
	"github.com/ymow/messenger_protocol_research/internal/trip"
	"github.com/ymow/messenger_protocol_research/internal/websocket"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(router *gin.Engine, db *gorm.DB, redisClient *redis.Client) {
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

	// Initialize services
	tripService := trip.NewService(db, redisClient)
	discoveryService := discovery.NewService(db)
	messageService := message.NewService(db, redisClient)

	// Initialize WebSocket Hub (Phase 1)
	hub := websocket.NewHub(redisClient, messageService)
	wsMessageHandler := websocket.NewMessageHandler(hub)

	// Start Hub in background
	go hub.Run()
	log.Println("✅ WebSocket Hub started")

	// Initialize and start message status worker (Phase 1)
	statusWorker := message.NewStatusWorker(messageService)
	ctx := context.Background()
	go statusWorker.Start(ctx)
	log.Println("✅ Message status worker started")

	// Initialize Matrix services (for Week 2)
	// Note: In production, use actual Matrix homeserver URL from config
	// For development, use matrix.org as a public Matrix server
	matrixClient := matrix.NewClient("https://matrix.org")
	ephemeralRoomMgr := matrix.NewEphemeralRoomManager(matrixClient, db)

	// Initialize cleanup service
	cleanupService := cleanup.NewService(db, redisClient, tripService, discoveryService, ephemeralRoomMgr)

	// Initialize handlers
	tripHandler := NewTripHandler(tripService)
	discoveryHandler := NewDiscoveryHandler(discoveryService)
	matrixHandler := NewMatrixHandler(ephemeralRoomMgr, tripService)
	cleanupHandler := NewCleanupHandler(cleanupService)
	connectionsHandler := NewConnectionsHandler()
	messageHandler := NewMessageHandler(db, messageService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Hello endpoint
		v1.GET("/hello", handlers.HelloHandler)
		v1.GET("/welcome", handlers.WelcomeHandler)

		// Trip endpoints (Week 1)
		trips := v1.Group("/trips")
		{
			trips.POST("/start", tripHandler.StartTrip)
			trips.POST("/end", tripHandler.EndTrip)
			trips.GET("/active", tripHandler.GetActiveTrip)
			trips.GET("/stats", tripHandler.GetActiveTripStats) // Admin endpoint
			trips.GET("/:id", tripHandler.GetTripByID)
			trips.GET("", tripHandler.GetUserTrips)
			trips.DELETE("/:id", tripHandler.CancelTrip)
			trips.PATCH("/:id/discovery", tripHandler.UpdateDiscoveryEnabled)
		}

		// Discovery endpoints (Week 1 - Anonymized Analytics)
		discoveries := v1.Group("/discoveries")
		{
			discoveries.POST("/log", discoveryHandler.LogDiscovery)
			discoveries.GET("/stats", discoveryHandler.GetDiscoveryStats)
			discoveries.GET("/popular-routes", discoveryHandler.GetPopularRoutes)
			discoveries.GET("/trends", discoveryHandler.GetDiscoveryTrends)
			discoveries.GET("/range", discoveryHandler.GetDiscoveriesByDateRange)
			discoveries.GET("/route/:route", discoveryHandler.GetDiscoveriesByRoute)
			discoveries.GET("/route/:route/stats", discoveryHandler.GetRouteStats)
		}

		// Matrix endpoints (Week 2 - Ephemeral DMs)
		matrixGroup := v1.Group("/matrix")
		{
			dm := matrixGroup.Group("/dm")
			{
				dm.POST("/create", matrixHandler.CreateEphemeralDM)
				dm.GET("/active", matrixHandler.GetMyActiveEphemeralDMs)
				dm.GET("/stats", matrixHandler.GetEphemeralRoomStats)
				dm.GET("/:id", matrixHandler.GetEphemeralDMByID)
				dm.PATCH("/:id/extend", matrixHandler.ExtendRoomLifetime)
				dm.POST("/:roomId/message", matrixHandler.IncrementMessageCount)
			}
		}

		// Cleanup endpoints (Week 2 - Admin/System endpoints)
		cleanupGroup := v1.Group("/cleanup")
		{
			cleanupGroup.GET("/stats", cleanupHandler.GetCleanupStats)
			cleanupGroup.POST("/run", cleanupHandler.RunManualCleanup) // Admin only
			cleanupGroup.GET("/expiring-rooms", cleanupHandler.GetExpiringRoomWarnings)
		}

		// Chat endpoints (Phase 0)
		v1.GET("/connections", connectionsHandler.GetConnections)

		// Message endpoints (Phase 1)
		messages := v1.Group("/messages")
		{
			messages.POST("", messageHandler.PostMessage)                         // Create message
			messages.GET("", messageHandler.GetMessages)                          // Get messages with filters
			messages.GET("/conversation/:peer_id", messageHandler.GetConversation) // Get conversation
			messages.PATCH("/:id/read", messageHandler.MarkAsRead)                 // Mark as read
		}

		// TODO: Add more endpoints in future phases
		// v1.POST("/auth/anonymous", handlers.AnonymousAuthHandler)
	}

	// WebSocket endpoint (Phase 1 - Proper Hub Integration)
	router.GET("/ws", func(c *gin.Context) {
		// Get user info from query params
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		deviceID := c.Query("device_id")
		sessionID := c.Query("session_id")
		stationID := c.Query("station_id")

		// Upgrade connection
		conn, err := websocket.UpgradeConnection(c.Writer, c.Request)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}

		// Create client
		client := websocket.NewClient(conn, hub, userID, deviceID, sessionID, stationID, wsMessageHandler)

		// Register client with hub
		hub.RegisterClient(client)

		// Start client read and write pumps
		go client.WritePump()
		go client.ReadPump()

		log.Printf("✅ WebSocket client connected: user=%s, station=%s", userID, stationID)
	})
}
