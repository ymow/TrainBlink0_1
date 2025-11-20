package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/middleware"
)

// Handler handles admin HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new admin handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login handles POST /api/v1/admin/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get client IP
	clientIP := getClientIP(r)

	// Login
	resp, err := h.service.Login(r.Context(), &req, clientIP)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			respondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
		case ErrAccountLocked:
			respondError(w, http.StatusForbidden, "ACCOUNT_LOCKED", err.Error())
		case ErrAccountInactive:
			respondError(w, http.StatusForbidden, "ACCOUNT_INACTIVE", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "LOGIN_FAILED", "Failed to login")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   resp,
	})
}

// RefreshToken handles POST /api/v1/admin/refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	resp, err := h.service.RefreshToken(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "REFRESH_FAILED", "Failed to refresh token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   resp,
	})
}

// Logout handles POST /api/v1/admin/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Get access token from header
	accessToken := extractToken(r)

	// Logout
	_ = h.service.Logout(r.Context(), req.RefreshToken, accessToken)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Logged out successfully",
	})
}

// CreateAdmin handles POST /api/v1/admin/create
func (h *Handler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	// Get current admin from context
	claims, ok := middleware.GetAdminClaims(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req CreateAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Parse creator admin ID
	creatorID, err := uuid.Parse(claims.AdminID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid admin ID")
		return
	}

	// Create admin
	admin, err := h.service.CreateAdmin(r.Context(), &req, creatorID)
	if err != nil {
		if err == ErrEmailAlreadyExists {
			respondError(w, http.StatusConflict, "EMAIL_EXISTS", err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create admin")
		}
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"data":   admin,
	})
}

// GetAdmin handles GET /api/v1/admin/{id}
func (h *Handler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Admin ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid admin ID")
		return
	}

	admin, err := h.service.GetAdminByID(r.Context(), id)
	if err != nil {
		if err == ErrAdminNotFound {
			respondError(w, http.StatusNotFound, "ADMIN_NOT_FOUND", "Admin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get admin")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   admin,
	})
}

// ListAdmins handles GET /api/v1/admin/list
func (h *Handler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	admins, total, err := h.service.ListAdmins(r.Context(), page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list admins")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"admins":    admins,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// UpdateAdminStatus handles PUT /api/v1/admin/{id}/status
func (h *Handler) UpdateAdminStatus(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/"), "/")
	if len(pathParts) < 2 {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid path")
		return
	}

	id, err := uuid.Parse(pathParts[0])
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid admin ID")
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.UpdateAdminStatus(r.Context(), id, req.IsActive); err != nil {
		if err == ErrAdminNotFound {
			respondError(w, http.StatusNotFound, "ADMIN_NOT_FOUND", "Admin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update admin status")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Admin status updated successfully",
	})
}

// DeleteAdmin handles DELETE /api/v1/admin/{id}
func (h *Handler) DeleteAdmin(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Admin ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid admin ID")
		return
	}

	if err := h.service.DeleteAdmin(r.Context(), id); err != nil {
		if err == ErrAdminNotFound {
			respondError(w, http.StatusNotFound, "ADMIN_NOT_FOUND", "Admin not found")
		} else {
			respondError(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete admin")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Admin deleted successfully",
	})
}

// ListRoles handles GET /api/v1/admin/roles
func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListRoles(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list roles")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   roles,
	})
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "error",
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
