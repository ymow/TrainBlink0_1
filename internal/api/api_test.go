package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ymow/messenger_protocol_research/internal/model"
	"github.com/ymow/messenger_protocol_research/pkg/logger"
)

// setupTestRouter creates a test router with in-memory database
func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
	// Initialize logger for tests
	_ = logger.Initialize("error") // Use error level to reduce test noise

	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables manually with SQLite-compatible SQL
	db.Exec(`CREATE TABLE IF NOT EXISTS trips (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		route TEXT NOT NULL,
		train_number TEXT,
		departure_time DATETIME NOT NULL,
		estimated_arrival DATETIME NOT NULL,
		actual_end_time DATETIME,
		discovery_enabled INTEGER DEFAULT 1,
		ble_anonymous_id TEXT NOT NULL UNIQUE,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS discoveries (
		id TEXT PRIMARY KEY,
		trip_route TEXT NOT NULL,
		discovered_user_anonymous_id TEXT NOT NULL,
		distance_estimate TEXT,
		rssi INTEGER,
		discoverer_age_range TEXT,
		discovered_age_range TEXT,
		discoverer_gender TEXT,
		discovered_gender TEXT,
		discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS matrix_ephemeral_rooms (
		id TEXT PRIMARY KEY,
		room_id TEXT NOT NULL UNIQUE,
		trip1_id TEXT NOT NULL,
		trip2_id TEXT NOT NULL,
		anonymous_id_1 TEXT NOT NULL,
		anonymous_id_2 TEXT NOT NULL,
		mls_group_id TEXT,
		expires_at DATETIME NOT NULL,
		auto_delete_queued INTEGER DEFAULT 0,
		deleted_at DATETIME,
		message_count INTEGER DEFAULT 0,
		last_message_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes (using nil for Redis since we're not testing Redis features)
	SetupRoutes(router, db, nil)

	// Cleanup function
	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return router, db, cleanup
}

// TestHealthEndpoints tests basic health check endpoints
func TestHealthEndpoints(t *testing.T) {
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	tests := []struct {
		name         string
		endpoint     string
		expectedCode int
	}{
		{"Ping", "/ping", http.StatusOK},
		{"Health", "/health", http.StatusOK},
		{"Hello", "/api/v1/hello", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.endpoint, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// TestTripEndpoints tests trip management endpoints
func TestTripEndpoints(t *testing.T) {
	router, db, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create a test user
	testUserID := uuid.New()

	t.Run("Start Trip", func(t *testing.T) {
		tripReq := map[string]interface{}{
			"route":             "Tokyo → Osaka",
			"departure_time":    time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			"estimated_arrival": time.Now().Add(4 * time.Hour).Format(time.RFC3339),
		}

		body, _ := json.Marshal(tripReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/trips/start", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", testUserID.String())
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		assert.NotNil(t, response["data"])
	})

	t.Run("Get Active Trip", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/trips/active", nil)
		req.Header.Set("X-User-ID", testUserID.String())
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])
	})

	t.Run("Get User Trips", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/trips?limit=10&offset=0", nil)
		req.Header.Set("X-User-ID", testUserID.String())
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get Trip Stats", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/trips/stats", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("End Trip", func(t *testing.T) {
		// First get the active trip
		var trip model.Trip
		db.Where("user_id = ? AND status = ?", testUserID, "active").First(&trip)

		endReq := map[string]interface{}{
			"trip_id": trip.ID.String(),
		}

		body, _ := json.Marshal(endReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/trips/end", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", testUserID.String())
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestDiscoveryEndpoints tests discovery tracking endpoints
func TestDiscoveryEndpoints(t *testing.T) {
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("Log Discovery", func(t *testing.T) {
		discoveryReq := map[string]interface{}{
			"trip_route":                   "Tokyo → Osaka",
			"discovered_user_anonymous_id": "TB_test123",
			"distance_estimate":            "Close (2-10m)",
			"rssi":                         -65,
		}

		body, _ := json.Marshal(discoveryReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/discoveries/log", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Get Discovery Stats", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/discoveries/stats", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])
	})

	t.Run("Get Popular Routes", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/discoveries/popular-routes?limit=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get Discovery Trends", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/discoveries/trends?days=7", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get Discoveries by Route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/discoveries/route/Tokyo%20→%20Osaka", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestMatrixEndpoints tests Matrix ephemeral DM endpoints
func TestMatrixEndpoints(t *testing.T) {
	router, db, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create two test users with active trips
	user1ID := uuid.New()
	user2ID := uuid.New()

	trip1 := &model.Trip{
		UserID:           user1ID,
		Route:            "Tokyo → Osaka",
		DepartureTime:    time.Now(),
		EstimatedArrival: time.Now().Add(3 * time.Hour),
		BLEAnonymousID:   "TB_user1",
		Status:           "active",
		DiscoveryEnabled: true,
	}
	db.Create(trip1)

	trip2 := &model.Trip{
		UserID:           user2ID,
		Route:            "Tokyo → Osaka",
		DepartureTime:    time.Now(),
		EstimatedArrival: time.Now().Add(3 * time.Hour),
		BLEAnonymousID:   "TB_user2",
		Status:           "active",
		DiscoveryEnabled: true,
	}
	db.Create(trip2)

	t.Run("Get Ephemeral Room Stats", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/matrix/dm/stats", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get My Active DMs - No Active Trip", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/matrix/dm/active", nil)
		req.Header.Set("X-User-ID", uuid.New().String()) // User with no trip
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(0), data["total"])
	})

	t.Run("Get My Active DMs - With Active Trip", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/matrix/dm/active", nil)
		req.Header.Set("X-User-ID", user1ID.String())
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestCleanupEndpoints tests cleanup admin endpoints
func TestCleanupEndpoints(t *testing.T) {
	router, db, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create some expired trips for testing
	expiredTrip := &model.Trip{
		UserID:           uuid.New(),
		Route:            "Old Route",
		DepartureTime:    time.Now().Add(-5 * time.Hour),
		EstimatedArrival: time.Now().Add(-3 * time.Hour),
		BLEAnonymousID:   "TB_expired",
		Status:           "active",
	}
	db.Create(expiredTrip)

	t.Run("Get Cleanup Stats", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/cleanup/stats", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])
	})

	t.Run("Get Expiring Rooms", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/cleanup/expiring-rooms?hours=1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Run Manual Cleanup", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/cleanup/run", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])
	})
}

// TestErrorCases tests error handling
func TestErrorCases(t *testing.T) {
	router, _, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("Invalid Trip ID Format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/trips/invalid-uuid", nil)
		req.Header.Set("X-User-ID", uuid.New().String())
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Start Trip Without User ID", func(t *testing.T) {
		tripReq := map[string]interface{}{
			"route":             "Tokyo → Osaka",
			"departure_time":    time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			"estimated_arrival": time.Now().Add(4 * time.Hour).Format(time.RFC3339),
		}

		body, _ := json.Marshal(tripReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/trips/start", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		// No X-User-ID header
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/discoveries/log", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestPagination tests pagination parameters
func TestPagination(t *testing.T) {
	router, db, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create test user and trips
	userID := uuid.New()
	for i := 0; i < 5; i++ {
		trip := &model.Trip{
			UserID:           userID,
			Route:            "Test Route",
			DepartureTime:    time.Now(),
			EstimatedArrival: time.Now().Add(3 * time.Hour),
			BLEAnonymousID:   uuid.New().String()[:8],
			Status:           "active",
		}
		db.Create(trip)
	}

	t.Run("Pagination with Limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/trips?limit=2&offset=0", nil)
		req.Header.Set("X-User-ID", userID.String())
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(2), data["limit"])
	})
}
