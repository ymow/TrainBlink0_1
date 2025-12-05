package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthFlow tests the full authentication lifecycle
func TestAuthFlow(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	var accessToken string

	// Step 1: Anonymous Login
	t.Run("Anonymous Login", func(t *testing.T) {
		loginReq := map[string]interface{}{
			"device_id":   "test-device-123",
			"device_type": "test-runner",
		}

		body, _ := json.Marshal(loginReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/anonymous", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.NotEmpty(t, response["access_token"])
		assert.NotEmpty(t, response["refresh_token"])
		assert.NotEmpty(t, response["user_id"])

		// Save token for next steps
		accessToken = response["access_token"].(string)
	})

	// Step 2: Access Protected Route WITHOUT Token
	t.Run("Access Protected Without Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/trips", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Step 3: Access Protected Route WITH Token
	t.Run("Access Protected With Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/trips", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Step 4: Access Protected Route WITH Invalid Token
	t.Run("Access Protected With Invalid Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/trips", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
