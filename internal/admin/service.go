package admin

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ymow/messenger_protocol_research/internal/auth"
	"github.com/ymow/messenger_protocol_research/internal/cache"
)

const (
	// MaxFailedLoginAttempts before account lockout
	MaxFailedLoginAttempts = 5
	// AccountLockDuration after max failed attempts
	AccountLockDuration = 15 * time.Minute
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked due to too many failed login attempts")
	ErrAccountInactive    = errors.New("account is inactive")
)

// Service handles admin business logic
type Service struct {
	repo         *Repository
	jwtService   *auth.JWTService
	redisService *cache.RedisService
}

// NewService creates a new admin service
func NewService(repo *Repository, jwtService *auth.JWTService, redisService *cache.RedisService) *Service {
	return &Service{
		repo:         repo,
		jwtService:   jwtService,
		redisService: redisService,
	}
}

// Login authenticates admin and returns token pair
func (s *Service) Login(ctx context.Context, req *LoginRequest, clientIP string) (*LoginResponse, error) {
	// Get admin by email
	admin, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrAdminNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if account is active
	if !admin.IsActive {
		return nil, ErrAccountInactive
	}

	// Check if account is locked
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now()) {
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		// Increment failed login attempts
		_ = s.repo.IncrementFailedLoginAttempts(ctx, admin.ID)

		// Check if we need to lock the account
		if admin.FailedLoginAttempts+1 >= MaxFailedLoginAttempts {
			lockUntil := time.Now().Add(AccountLockDuration)
			_ = s.repo.LockAccount(ctx, admin.ID, lockUntil)
		}

		return nil, ErrInvalidCredentials
	}

	// Generate token pair
	tokenPair, err := s.jwtService.GenerateTokenPair(
		admin.ID,
		admin.Email,
		admin.Role,
		admin.Permissions,
	)
	if err != nil {
		return nil, err
	}

	// Store refresh token in Redis
	refreshTokenHash := auth.HashToken(tokenPair.RefreshToken)
	if err := s.redisService.StoreRefreshToken(
		ctx,
		refreshTokenHash,
		admin.ID.String(),
		7*24*time.Hour, // 7 days
	); err != nil {
		return nil, err
	}

	// Update last login
	_ = s.repo.UpdateLastLogin(ctx, admin.ID, clientIP)

	return &LoginResponse{
		Admin:        admin.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RefreshToken generates new access token from refresh token
func (s *Service) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*LoginResponse, error) {
	// Verify refresh token
	claims, err := s.jwtService.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Check if it's a refresh token
	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Verify token exists in Redis
	refreshTokenHash := auth.HashToken(req.RefreshToken)
	adminID, err := s.redisService.GetRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		return nil, errors.New("refresh token not found or expired")
	}

	// Parse admin ID
	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		return nil, err
	}

	// Get admin details
	admin, err := s.repo.GetByID(ctx, adminUUID)
	if err != nil {
		return nil, err
	}

	// Check if account is still active
	if !admin.IsActive {
		return nil, ErrAccountInactive
	}

	// Generate new token pair
	tokenPair, err := s.jwtService.GenerateTokenPair(
		admin.ID,
		admin.Email,
		admin.Role,
		admin.Permissions,
	)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token
	_ = s.redisService.DeleteRefreshToken(ctx, refreshTokenHash)

	// Store new refresh token in Redis
	newRefreshTokenHash := auth.HashToken(tokenPair.RefreshToken)
	if err := s.redisService.StoreRefreshToken(
		ctx,
		newRefreshTokenHash,
		admin.ID.String(),
		7*24*time.Hour,
	); err != nil {
		return nil, err
	}

	return &LoginResponse{
		Admin:        admin.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// Logout revokes refresh token and optionally access token
func (s *Service) Logout(ctx context.Context, refreshToken, accessToken string) error {
	// Revoke refresh token
	if refreshToken != "" {
		refreshTokenHash := auth.HashToken(refreshToken)
		_ = s.redisService.DeleteRefreshToken(ctx, refreshTokenHash)
	}

	// Revoke access token (add to revocation list)
	if accessToken != "" {
		accessTokenHash := auth.HashToken(accessToken)
		// Store until token would have expired (1 hour default)
		_ = s.redisService.RevokeAccessToken(ctx, accessTokenHash, 1*time.Hour)
	}

	return nil
}

// CreateAdmin creates a new admin user
func (s *Service) CreateAdmin(ctx context.Context, req *CreateAdminRequest, createdBy uuid.UUID) (*AdminResponse, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Get role permissions
	role, err := s.repo.GetRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}

	// Extract permissions from role
	permissions := []string{}
	if perms, ok := role.Permissions["permissions"].([]interface{}); ok {
		for _, p := range perms {
			if perm, ok := p.(string); ok {
				permissions = append(permissions, perm)
			}
		}
	}

	// Create admin
	admin := &Admin{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		Permissions:  permissions,
		IsActive:     true,
		CreatedBy:    &createdBy,
	}

	if err := s.repo.Create(ctx, admin); err != nil {
		return nil, err
	}

	return admin.ToResponse(), nil
}

// GetAdminByID retrieves admin by ID
func (s *Service) GetAdminByID(ctx context.Context, id uuid.UUID) (*AdminResponse, error) {
	admin, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return admin.ToResponse(), nil
}

// ListAdmins retrieves all admins with pagination
func (s *Service) ListAdmins(ctx context.Context, page, pageSize int) ([]*AdminResponse, int64, error) {
	offset := (page - 1) * pageSize
	admins, total, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*AdminResponse, len(admins))
	for i, admin := range admins {
		responses[i] = admin.ToResponse()
	}

	return responses, total, nil
}

// UpdateAdminStatus updates admin active status
func (s *Service) UpdateAdminStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	admin, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	admin.IsActive = isActive
	return s.repo.Update(ctx, admin)
}

// DeleteAdmin soft deletes an admin
func (s *Service) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ListRoles retrieves all available roles
func (s *Service) ListRoles(ctx context.Context) ([]*Role, error) {
	return s.repo.ListRoles(ctx)
}
