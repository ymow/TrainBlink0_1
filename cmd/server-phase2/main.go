package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ymow/messenger_protocol_research/internal/admin"
	"github.com/ymow/messenger_protocol_research/internal/api"
	"github.com/ymow/messenger_protocol_research/internal/auth"
	"github.com/ymow/messenger_protocol_research/internal/cache"
	"github.com/ymow/messenger_protocol_research/internal/config"
	"github.com/ymow/messenger_protocol_research/internal/database"
	"github.com/ymow/messenger_protocol_research/internal/geofence"
	"github.com/ymow/messenger_protocol_research/internal/matrix"
	"github.com/ymow/messenger_protocol_research/internal/middleware"
	"github.com/ymow/messenger_protocol_research/internal/model"
	"github.com/ymow/messenger_protocol_research/internal/websocket"
)

const version = "0.6.0-phase2"

func main() {
	fmt.Printf("🚀 TrainBlink Server v%s (Phase 2: Full Stack)\n", version)
	fmt.Println("=========================================================")

	// ============================================================
	// Phase 2: Load Configuration
	// ============================================================
	fmt.Println("\n📋 Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	fmt.Printf("   Environment: %s\n", cfg.Server.Env)

	// ============================================================
	// Phase 2: Initialize PostgreSQL Database
	// ============================================================
	fmt.Println("\n🐘 Connecting to PostgreSQL...")
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	fmt.Println("   ✅ PostgreSQL connected")

	// Run auto migrations
	fmt.Println("   Running auto-migrations...")
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}
	fmt.Println("   ✅ Migrations complete")

	// Close database on exit
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
			fmt.Println("\n🐘 PostgreSQL connection closed")
		}
	}()

	// ============================================================
	// Phase 2: Initialize Redis Cache
	// ============================================================
	fmt.Println("\n🔴 Connecting to Redis...")
	redisService, err := cache.NewRedisService(cfg.Redis)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer func() {
		if err := redisService.Close(); err != nil {
			log.Printf("⚠️  Error closing Redis: %v", err)
		} else {
			fmt.Println("🔴 Redis connection closed")
		}
	}()

	// ============================================================
	// Phase 2: Initialize JWT Service
	// ============================================================
	fmt.Println("\n🔐 Initializing JWT service...")
	jwtService := auth.NewJWTService(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)
	fmt.Println("   ✅ JWT service ready")

	// ============================================================
	// Phase 2: Initialize Admin Repository & Service
	// ============================================================
	fmt.Println("\n👤 Initializing Admin service...")
	adminRepo := admin.NewRepository(db)
	adminService := admin.NewService(adminRepo, jwtService, redisService)
	fmt.Println("   ✅ Admin service ready")

	// ============================================================
	// Phase 1: Initialize Matrix Bridge Service
	// ============================================================
	fmt.Println("\n📡 Initializing Matrix Bridge Service...")
	matrixBridge := matrix.NewBridgeService(
		"https://matrix.trainblink.org",
		"trainblink.org",
	)

	// ============================================================
	// Phase 1: Initialize Geofence Service
	// ============================================================
	fmt.Println("\n📍 Initializing Geofence Service...")
	geofenceService := geofence.NewService(matrixBridge)

	// Load sample stations
	fmt.Println("🚉 Loading sample stations...")
	sampleStations := loadSampleStations()
	geofenceService.LoadStations(sampleStations)
	fmt.Printf("   Loaded %d stations\n", len(sampleStations))

	// ============================================================
	// Phase 1: Initialize WebSocket Hub
	// ============================================================
	fmt.Println("\n🔌 Initializing WebSocket Hub...")
	wsHub := websocket.NewHub()
	go wsHub.Run() // Start hub in background goroutine
	fmt.Println("   ✅ WebSocket Hub running")

	// ============================================================
	// Initialize HTTP Handlers
	// ============================================================
	geofenceHandler := api.NewGeofenceHandler(geofenceService)
	geofenceHandler.SetWebSocketHub(wsHub)
	wsHandler := api.NewWebSocketHandler(wsHub, geofenceService)
	adminHandler := admin.NewHandler(adminService)

	// ============================================================
	// Setup Router with Middleware
	// ============================================================
	mux := http.NewServeMux()

	// Health check endpoints (no auth required)
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/health", healthHandler(cfg.Server.Env, version))

	// Phase 1: Geofence API endpoints (public)
	mux.HandleFunc("/api/v1/geofence/enter", geofenceHandler.EnterStation)
	mux.HandleFunc("/api/v1/geofence/exit", geofenceHandler.ExitStation)
	mux.HandleFunc("/api/v1/geofence/stats", geofenceHandler.GetStats)

	// Phase 1: WebSocket endpoint (public)
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Phase 2: Admin authentication endpoints (public)
	mux.HandleFunc("/api/v1/admin/login", adminHandler.Login)
	mux.HandleFunc("/api/v1/admin/refresh", adminHandler.RefreshToken)

	// Phase 2: Admin authenticated endpoints
	adminAuthMiddleware := middleware.AdminAuthMiddleware(jwtService, redisService)

	// Logout (requires auth)
	mux.Handle("/api/v1/admin/logout", adminAuthMiddleware(http.HandlerFunc(adminHandler.Logout)))

	// Admin management endpoints (requires admin role)
	mux.Handle("/api/v1/admin/create",
		adminAuthMiddleware(
			middleware.RequirePermission("admin.create")(
				http.HandlerFunc(adminHandler.CreateAdmin))))

	mux.Handle("/api/v1/admin/list",
		adminAuthMiddleware(
			middleware.RequirePermission("admin.view")(
				http.HandlerFunc(adminHandler.ListAdmins))))

	mux.Handle("/api/v1/admin/roles",
		adminAuthMiddleware(
			middleware.RequirePermission("admin.view")(
				http.HandlerFunc(adminHandler.ListRoles))))

	// ============================================================
	// Apply Global Middleware Stack
	// ============================================================
	handler := middleware.LoggingMiddleware(mux)
	handler = middleware.SecurityHeadersMiddleware(handler)
	handler = middleware.CORSMiddleware(cfg.CORS)(handler)

	// Optional: Apply global rate limiting (can be overridden per route)
	// handler = middleware.IPRateLimitMiddleware(redisService, 100)(handler)

	// ============================================================
	// Start HTTP Server
	// ============================================================
	fmt.Printf("\n✅ Server ready!\n")
	fmt.Printf("   Address: http://localhost%s\n", cfg.Server.Port)
	fmt.Println("\n📚 Available Endpoints:")
	fmt.Println("\n  Phase 1: Geofencing & WebSocket")
	fmt.Println("   GET  /ping                          - Ping test")
	fmt.Println("   GET  /health                        - Health check + service status")
	fmt.Println("   POST /api/v1/geofence/enter         - Enter station")
	fmt.Println("   POST /api/v1/geofence/exit          - Exit station")
	fmt.Println("   GET  /api/v1/geofence/stats         - Get statistics")
	fmt.Println("   GET  /ws?session_id={id}            - WebSocket connection")
	fmt.Println("\n  Phase 2: Admin Authentication")
	fmt.Println("   POST /api/v1/admin/login            - Admin login")
	fmt.Println("   POST /api/v1/admin/refresh          - Refresh access token")
	fmt.Println("   POST /api/v1/admin/logout           - Admin logout [Auth Required]")
	fmt.Println("   POST /api/v1/admin/create           - Create new admin [Auth + Permission]")
	fmt.Println("   GET  /api/v1/admin/list             - List all admins [Auth + Permission]")
	fmt.Println("   GET  /api/v1/admin/roles            - List all roles [Auth + Permission]")
	fmt.Println()
	fmt.Println("🔧 Phase 1 Features:")
	fmt.Println("   ✅ P2P Coordination")
	fmt.Println("   ✅ Matrix Bridge Integration")
	fmt.Println("   ✅ WebSocket Real-time Messaging")
	fmt.Println("   ✅ Rate Limiting & Validation")
	fmt.Println()
	fmt.Println("🔧 Phase 2 Features:")
	fmt.Println("   ✅ PostgreSQL Database (GORM)")
	fmt.Println("   ✅ Redis Cache & Sessions")
	fmt.Println("   ✅ JWT Authentication")
	fmt.Println("   ✅ Security Middleware")
	fmt.Println("   ✅ Admin API (Login, Logout, Refresh)")
	fmt.Println("   ✅ RBAC with Permissions")
	fmt.Println("   ⏳ User Management (coming next)")
	fmt.Println()

	addr := cfg.Server.Port
	log.Printf("🚀 Starting server on %s...\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}

// pingHandler handles GET /ping
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message":"pong","timestamp":"%s","server":"TrainBlink v%s"}`,
		time.Now().Format(time.RFC3339), version)
}

// healthHandler handles GET /health
func healthHandler(env, ver string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"status":"healthy",
			"version":"%s",
			"environment":"%s",
			"timestamp":"%s",
			"services":{
				"database":"connected",
				"redis":"connected",
				"websocket":"running",
				"matrix":"initialized"
			},
			"features":["geofence","p2p","matrix","websocket","jwt","postgres","redis"]
		}`, ver, env, time.Now().Format(time.RFC3339))
	}
}

// loadSampleStations creates sample station data
func loadSampleStations() []*model.Station {
	return []*model.Station{
		{
			ID:        "station_tokyo_001",
			PlaceID:   "ChIJ51cu8IcbXWARiRtXIothAS4",
			Name:      "東京駅",
			NameEn:    "Tokyo Station",
			Type:      model.StationTypeTRA,
			Latitude:  35.6812,
			Longitude: 139.7671,
			Radius:    500,
			Address:   "東京都千代田区丸の内１丁目",
			City:      "Tokyo",
			Country:   "JP",
			Lines:     []string{"JR Yamanote", "JR Chuo", "Shinkansen"},
		},
		{
			ID:        "station_taipei_001",
			PlaceID:   "ChIJ_SDBVGSrQjQR_DJJqrU6DuI",
			Name:      "台北車站",
			NameEn:    "Taipei Main Station",
			Type:      model.StationTypeTRA,
			Latitude:  25.0478,
			Longitude: 121.5170,
			Radius:    500,
			Address:   "台北市中正區北平西路3號",
			City:      "Taipei",
			Country:   "TW",
			Lines:     []string{"TRA", "THSR", "MRT Blue", "MRT Red"},
		},
		{
			ID:        "station_shibuya_001",
			PlaceID:   "ChIJXSModoGLGGARYl_7vUhqJeA",
			Name:      "渋谷駅",
			NameEn:    "Shibuya Station",
			Type:      model.StationTypeMRT,
			Latitude:  35.6580,
			Longitude: 139.7016,
			Radius:    500,
			Address:   "東京都渋谷区道玄坂１丁目",
			City:      "Tokyo",
			Country:   "JP",
			Lines:     []string{"JR Yamanote", "Tokyo Metro Ginza", "Tokyo Metro Hanzomon"},
		},
		{
			ID:        "station_taichung_001",
			PlaceID:   "ChIJDVXrDmUWaTQRn5V7MkqJUuY",
			Name:      "台中車站",
			NameEn:    "Taichung Station",
			Type:      model.StationTypeTRA,
			Latitude:  24.1370,
			Longitude: 120.6852,
			Radius:    500,
			Address:   "台中市中區建國路172號",
			City:      "Taichung",
			Country:   "TW",
			Lines:     []string{"TRA"},
		},
		{
			ID:        "station_kaohsiung_001",
			PlaceID:   "ChIJqafBOl8QbjQRoSgvXxcmJxA",
			Name:      "高雄車站",
			NameEn:    "Kaohsiung Station",
			Type:      model.StationTypeTRA,
			Latitude:  22.6391,
			Longitude: 120.3023,
			Radius:    500,
			Address:   "高雄市三民區建國二路318號",
			City:      "Kaohsiung",
			Country:   "TW",
			Lines:     []string{"TRA", "MRT Red"},
		},
	}
}
