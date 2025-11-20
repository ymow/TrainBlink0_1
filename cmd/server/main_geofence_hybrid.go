package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ymow/messenger_protocol_research/internal/api"
	"github.com/ymow/messenger_protocol_research/internal/geofence"
	"github.com/ymow/messenger_protocol_research/internal/matrix"
	"github.com/ymow/messenger_protocol_research/internal/model"
	"github.com/ymow/messenger_protocol_research/internal/websocket"
)

const (
	version = "0.5.0-websocket"
	port    = "8080"
)

func main() {
	fmt.Printf("🚀 TrainBlink Geofence + Matrix Hybrid Server v%s\n", version)
	fmt.Println("=================================================")

	// 1. Initialize Matrix Bridge Service
	fmt.Println("📡 Initializing Matrix Bridge Service...")
	matrixBridge := matrix.NewBridgeService(
		"https://matrix.trainblink.org",
		"trainblink.org",
	)

	// 2. Initialize Geofence Service
	fmt.Println("📍 Initializing Geofence Service...")
	geofenceService := geofence.NewService(matrixBridge)

	// 3. Load sample stations
	fmt.Println("🚉 Loading sample stations...")
	sampleStations := loadSampleStations()
	geofenceService.LoadStations(sampleStations)
	fmt.Printf("   Loaded %d stations\n", len(sampleStations))

	// 4. Initialize WebSocket Hub
	fmt.Println("🔌 Initializing WebSocket Hub...")
	wsHub := websocket.NewHub()
	go wsHub.Run() // Start hub in background goroutine
	fmt.Println("   WebSocket Hub running")

	// 5. Initialize HTTP handlers
	geofenceHandler := api.NewGeofenceHandler(geofenceService)
	wsHandler := api.NewWebSocketHandler(wsHub, geofenceService)

	// 6. Setup routes
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/health", healthHandler)

	// Geofence API endpoints
	mux.HandleFunc("/api/v1/geofence/enter", enableCORS(geofenceHandler.EnterStation))
	mux.HandleFunc("/api/v1/geofence/exit", enableCORS(geofenceHandler.ExitStation))
	mux.HandleFunc("/api/v1/geofence/stats", enableCORS(geofenceHandler.GetStats))

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// 7. Start server
	addr := ":" + port
	fmt.Printf("\n✅ Server ready!\n")
	fmt.Printf("   Address: http://localhost:%s\n", port)
	fmt.Println("\n📚 Available Endpoints:")
	fmt.Println("   GET  /ping                      - Ping test")
	fmt.Println("   GET  /health                    - Health check")
	fmt.Println("   POST /api/v1/geofence/enter     - Enter station (Hybrid: P2P + Matrix)")
	fmt.Println("   POST /api/v1/geofence/exit      - Exit station")
	fmt.Println("   GET  /api/v1/geofence/stats     - Get statistics")
	fmt.Println("   GET  /ws?session_id={id}        - WebSocket connection (real-time messaging)")
	fmt.Println("\n🔧 Features Enabled:")
	fmt.Println("   ✅ P2P Coordination")
	fmt.Println("   ✅ Matrix Bridge Integration")
	fmt.Println("   ✅ Dual-channel support")
	fmt.Println("   ✅ Anonymous Matrix users")
	fmt.Println("   ✅ Station room management")
	fmt.Println("   ✅ WebSocket real-time messaging")
	fmt.Println("   ✅ Station-based chat rooms")
	fmt.Println("   ✅ Typing indicators & presence")
	fmt.Println("\n🎯 Test with:")
	fmt.Println("   curl -X POST http://localhost:8080/api/v1/geofence/enter \\")
	fmt.Println("        -H 'Content-Type: application/json' \\")
	fmt.Println("        -H 'X-User-ID: user-001' \\")
	fmt.Println("        -H 'X-Device-ID: iOS-123' \\")
	fmt.Println("        -d '{")
	fmt.Println("          \"station_id\": \"station_tokyo_001\",")
	fmt.Println("          \"coordinates\": {\"latitude\": 35.6812, \"longitude\": 139.7671, \"accuracy\": 10},")
	fmt.Println("          \"capabilities\": {\"p2p_enabled\": true, \"matrix_enabled\": true}")
	fmt.Println("        }'")
	fmt.Println()

	log.Fatal(http.ListenAndServe(addr, mux))
}

// pingHandler handles GET /ping
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message":"pong","timestamp":"%s","server":"TrainBlink Geofence+Matrix v%s"}`,
		time.Now().Format(time.RFC3339), version)
}

// healthHandler handles GET /health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"healthy","version":"%s","timestamp":"%s","features":["geofence","p2p","matrix"]}`,
		version, time.Now().Format(time.RFC3339))
}

// enableCORS adds CORS headers
func enableCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Device-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

// loadSampleStations creates sample station data
func loadSampleStations() []*model.Station {
	return []*model.Station{
		{
			ID:       "station_tokyo_001",
			PlaceID:  "ChIJ51cu8IcbXWARiRtXIothAS4",
			Name:     "東京駅",
			NameEn:   "Tokyo Station",
			Type:     model.StationTypeTRA,
			Latitude: 35.6812,
			Longitude: 139.7671,
			Radius:   500,
			Address:  "東京都千代田区丸の内１丁目",
			City:     "Tokyo",
			Country:  "JP",
			Lines:    []string{"JR Yamanote", "JR Chuo", "Shinkansen"},
		},
		{
			ID:       "station_taipei_001",
			PlaceID:  "ChIJ_SDBVGSrQjQR_DJJqrU6DuI",
			Name:     "台北車站",
			NameEn:   "Taipei Main Station",
			Type:     model.StationTypeTRA,
			Latitude: 25.0478,
			Longitude: 121.5170,
			Radius:   500,
			Address:  "台北市中正區北平西路3號",
			City:     "Taipei",
			Country:  "TW",
			Lines:    []string{"TRA", "THSR", "MRT Blue", "MRT Red"},
		},
		{
			ID:       "station_shibuya_001",
			PlaceID:  "ChIJXSModoGLGGARYl_7vUhqJeA",
			Name:     "渋谷駅",
			NameEn:   "Shibuya Station",
			Type:     model.StationTypeMRT,
			Latitude: 35.6580,
			Longitude: 139.7016,
			Radius:   500,
			Address:  "東京都渋谷区道玄坂１丁目",
			City:     "Tokyo",
			Country:  "JP",
			Lines:    []string{"JR Yamanote", "Tokyo Metro Ginza", "Tokyo Metro Hanzomon"},
		},
		{
			ID:       "station_taichung_001",
			PlaceID:  "ChIJDVXrDmUWaTQRn5V7MkqJUuY",
			Name:     "台中車站",
			NameEn:   "Taichung Station",
			Type:     model.StationTypeTRA,
			Latitude: 24.1370,
			Longitude: 120.6852,
			Radius:   500,
			Address:  "台中市中區建國路172號",
			City:     "Taichung",
			Country:  "TW",
			Lines:    []string{"TRA"},
		},
		{
			ID:       "station_kaohsiung_001",
			PlaceID:  "ChIJqafBOl8QbjQRoSgvXxcmJxA",
			Name:     "高雄車站",
			NameEn:   "Kaohsiung Station",
			Type:     model.StationTypeTRA,
			Latitude: 22.6391,
			Longitude: 120.3023,
			Radius:   500,
			Address:  "高雄市三民區建國二路318號",
			City:     "Kaohsiung",
			Country:  "TW",
			Lines:    []string{"TRA", "MRT Red"},
		},
	}
}
