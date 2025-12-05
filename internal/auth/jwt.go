package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// JWTService handles JWT token generation and verification
type JWTService struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Claims represents JWT claims (supports both admin and user tokens)
type Claims struct {
	// Admin fields
	AdminID     string   `json:"admin_id,omitempty"`
	Email       string   `json:"email,omitempty"`
	Role        string   `json:"role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`

	// User fields
	UserID   string `json:"user_id,omitempty"`
	DeviceID string `json:"device_id,omitempty"`

	// Common fields
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// GenerateTokenPair generates access and refresh tokens
func (s *JWTService) GenerateTokenPair(
	adminID uuid.UUID,
	email, role string,
	permissions []string,
) (*TokenPair, error) {
	// Generate Access Token
	accessToken, err := s.generateToken(
		adminID.String(),
		email,
		role,
		permissions,
		"access",
		s.accessTokenTTL,
	)
	if err != nil {
		return nil, err
	}

	// Generate Refresh Token
	refreshToken, err := s.generateToken(
		adminID.String(),
		email,
		role,
		permissions,
		"refresh",
		s.refreshTokenTTL,
	)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}

// GenerateUserTokenPair generates access and refresh tokens for regular users
func (s *JWTService) GenerateUserTokenPair(
	userID uuid.UUID,
	deviceID string,
) (*TokenPair, error) {
	// Generate Access Token
	accessToken, err := s.generateUserToken(
		userID.String(),
		deviceID,
		"access",
		s.accessTokenTTL,
	)
	if err != nil {
		return nil, err
	}

	// Generate Refresh Token
	refreshToken, err := s.generateUserToken(
		userID.String(),
		deviceID,
		"refresh",
		s.refreshTokenTTL,
	)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}

// generateToken creates a JWT token
func (s *JWTService) generateToken(
	adminID, email, role string,
	permissions []string,
	tokenType string,
	ttl time.Duration,
) (string, error) {
	now := time.Now()
	claims := Claims{
		AdminID:     adminID,
		Email:       email,
		Role:        role,
		Permissions: permissions,
		TokenType:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "trainblink-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// generateUserToken creates a JWT token for regular users
func (s *JWTService) generateUserToken(
	userID, deviceID string,
	tokenType string,
	ttl time.Duration,
) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		DeviceID:  deviceID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "trainblink-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// VerifyToken verifies and parses a JWT token
func (s *JWTService) VerifyToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return s.secretKey, nil
		},
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	// Convert to AdminClaims
	adminClaims := (*AdminClaims)(claims)
	return adminClaims, nil
}

// ValidateToken validates and parses a JWT token (for both admin and user tokens)
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return s.secretKey, nil
		},
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from refresh token
func (s *JWTService) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := s.VerifyToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// Verify it's a refresh token
	if claims.TokenType != "refresh" {
		return "", ErrInvalidToken
	}

	// Generate new access token
	return s.generateToken(
		claims.AdminID,
		claims.Email,
		claims.Role,
		claims.Permissions,
		"access",
		s.accessTokenTTL,
	)
}

// AdminClaims is an alias for Claims with additional helper methods
type AdminClaims Claims

// HasPermission checks if admin has a specific permission
// Supports wildcard permissions like "user.*" or "*"
func (c *AdminClaims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		// Super admin with "*" permission
		if p == "*" {
			return true
		}

		// Exact match
		if p == permission {
			return true
		}

		// Wildcard match (e.g., "user.*" matches "user.view")
		if strings.HasSuffix(p, ".*") {
			prefix := strings.TrimSuffix(p, ".*")
			if strings.HasPrefix(permission, prefix+".") {
				return true
			}
		}
	}
	return false
}

// HashToken generates SHA256 hash of token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
