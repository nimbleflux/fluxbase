package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// Well-known UUIDs for synthetic users (anon/service role impersonation)
// These are valid UUIDs that don't conflict with real user IDs
const (
	// AnonUserID is the UUID used for anonymous user impersonation
	// Using a nil UUID variant to indicate a synthetic/anonymous user
	AnonUserID = "00000000-0000-0000-0000-000000000000"
	// ServiceUserID is the UUID used for service role impersonation
	ServiceUserID = "00000000-0000-0000-0000-000000000001"
)

var (
	// ErrNotAdmin is returned when a non-dashboard-admin tries to impersonate
	ErrNotAdmin = errors.New("only dashboard admins can impersonate users")
	// ErrSelfImpersonation is returned when trying to impersonate yourself
	ErrSelfImpersonation = errors.New("cannot impersonate yourself")
	// ErrNoActiveImpersonation is returned when trying to stop non-existent impersonation
	ErrNoActiveImpersonation = errors.New("no active impersonation session found")
)

// ImpersonationType represents the type of impersonation
type ImpersonationType string

const (
	ImpersonationTypeUser    ImpersonationType = "user"
	ImpersonationTypeAnon    ImpersonationType = "anon"
	ImpersonationTypeService ImpersonationType = "service"
)

// ImpersonationSession represents an admin impersonation session
type ImpersonationSession struct {
	ID                string            `json:"id" db:"id"`
	AdminUserID       string            `json:"admin_user_id" db:"admin_user_id"`
	TargetUserID      *string           `json:"target_user_id,omitempty" db:"target_user_id"`
	ImpersonationType ImpersonationType `json:"impersonation_type" db:"impersonation_type"`
	TargetRole        *string           `json:"target_role,omitempty" db:"target_role"`
	Reason            string            `json:"reason,omitempty" db:"reason"`
	StartedAt         time.Time         `json:"started_at" db:"started_at"`
	EndedAt           *time.Time        `json:"ended_at,omitempty" db:"ended_at"`
	IPAddress         string            `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent         string            `json:"user_agent,omitempty" db:"user_agent"`
	IsActive          bool              `json:"is_active" db:"is_active"`
	AccessTokenJTI    string            `json:"access_token_jti,omitempty" db:"access_token_jti"`
	RefreshTokenJTI   string            `json:"refresh_token_jti,omitempty" db:"refresh_token_jti"`
}

// ImpersonationRepository handles database operations for impersonation sessions
type ImpersonationRepository struct {
	db *database.Connection
}

// NewImpersonationRepository creates a new impersonation repository
func NewImpersonationRepository(db *database.Connection) *ImpersonationRepository {
	return &ImpersonationRepository{db: db}
}

// Create creates a new impersonation session
func (r *ImpersonationRepository) Create(ctx context.Context, session *ImpersonationSession) (*ImpersonationSession, error) {
	query := `
		INSERT INTO auth.impersonation_sessions
		(id, admin_user_id, target_user_id, impersonation_type, target_role, reason, started_at, ip_address, user_agent, is_active, access_token_jti, refresh_token_jti)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, admin_user_id, target_user_id, impersonation_type, target_role, reason, started_at, ended_at, ip_address, user_agent, is_active, access_token_jti, refresh_token_jti
	`

	result := &ImpersonationSession{}
	err := database.WrapWithServiceRole(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query,
			session.ID,
			session.AdminUserID,
			session.TargetUserID,
			session.ImpersonationType,
			session.TargetRole,
			session.Reason,
			session.StartedAt,
			session.IPAddress,
			session.UserAgent,
			session.IsActive,
			session.AccessTokenJTI,
			session.RefreshTokenJTI,
		).Scan(
			&result.ID,
			&result.AdminUserID,
			&result.TargetUserID,
			&result.ImpersonationType,
			&result.TargetRole,
			&result.Reason,
			&result.StartedAt,
			&result.EndedAt,
			&result.IPAddress,
			&result.UserAgent,
			&result.IsActive,
			&result.AccessTokenJTI,
			&result.RefreshTokenJTI,
		)
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// EndSession marks an impersonation session as ended
func (r *ImpersonationRepository) EndSession(ctx context.Context, sessionID string) error {
	query := `
		UPDATE auth.impersonation_sessions
		SET ended_at = NOW(), is_active = false
		WHERE id = $1 AND is_active = true
	`

	return database.WrapWithServiceRole(ctx, r.db, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, query, sessionID)
		if err != nil {
			return err
		}

		if result.RowsAffected() == 0 {
			return ErrNoActiveImpersonation
		}

		return nil
	})
}

// GetActiveByAdmin gets the active impersonation session for an admin
func (r *ImpersonationRepository) GetActiveByAdmin(ctx context.Context, adminUserID string) (*ImpersonationSession, error) {
	query := `
		SELECT id, admin_user_id, target_user_id, impersonation_type, target_role, reason, started_at, ended_at, ip_address, user_agent, is_active, access_token_jti, refresh_token_jti
		FROM auth.impersonation_sessions
		WHERE admin_user_id = $1 AND is_active = true
		ORDER BY started_at DESC
		LIMIT 1
	`

	session := &ImpersonationSession{}
	err := database.WrapWithServiceRole(ctx, r.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, query, adminUserID).Scan(
			&session.ID,
			&session.AdminUserID,
			&session.TargetUserID,
			&session.ImpersonationType,
			&session.TargetRole,
			&session.Reason,
			&session.StartedAt,
			&session.EndedAt,
			&session.IPAddress,
			&session.UserAgent,
			&session.IsActive,
			&session.AccessTokenJTI,
			&session.RefreshTokenJTI,
		)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoActiveImpersonation
		}
		return nil, err
	}

	return session, nil
}

// ListByAdmin lists all impersonation sessions for an admin (audit trail)
func (r *ImpersonationRepository) ListByAdmin(ctx context.Context, adminUserID string, limit, offset int) ([]*ImpersonationSession, error) {
	query := `
		SELECT id, admin_user_id, target_user_id, impersonation_type, target_role, reason, started_at, ended_at, ip_address, user_agent, is_active, access_token_jti, refresh_token_jti
		FROM auth.impersonation_sessions
		WHERE admin_user_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`

	var sessions []*ImpersonationSession
	err := database.WrapWithServiceRole(ctx, r.db, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, adminUserID, limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			session := &ImpersonationSession{}
			err := rows.Scan(
				&session.ID,
				&session.AdminUserID,
				&session.TargetUserID,
				&session.ImpersonationType,
				&session.TargetRole,
				&session.Reason,
				&session.StartedAt,
				&session.EndedAt,
				&session.IPAddress,
				&session.UserAgent,
				&session.IsActive,
				&session.AccessTokenJTI,
				&session.RefreshTokenJTI,
			)
			if err != nil {
				return err
			}
			sessions = append(sessions, session)
		}

		return rows.Err()
	})

	return sessions, err
}

// ImpersonationService provides business logic for admin impersonation
type ImpersonationService struct {
	repo                  *ImpersonationRepository
	userRepo              *UserRepository
	jwtManager            *JWTManager
	db                    *database.Connection
	tokenBlacklistService *TokenBlacklistService
}

// NewImpersonationService creates a new impersonation service
func NewImpersonationService(
	repo *ImpersonationRepository,
	userRepo *UserRepository,
	jwtManager *JWTManager,
	db *database.Connection,
) *ImpersonationService {
	return &ImpersonationService{
		repo:       repo,
		userRepo:   userRepo,
		jwtManager: jwtManager,
		db:         db,
	}
}

// SetTokenBlacklistService sets the token blacklist service dependency
// This is called during service initialization to wire up dependencies
func (s *ImpersonationService) SetTokenBlacklistService(service *TokenBlacklistService) {
	s.tokenBlacklistService = service
}

// verifyAdminUser checks if the user is a platform admin
// Returns nil if the user is a valid platform admin, error otherwise
// Checks both platform.users and auth.users tables
func (s *ImpersonationService) verifyAdminUser(ctx context.Context, adminUserID string) error {
	// First, check if user exists in platform.users (they are always admins)
	var count int
	err := database.WrapWithServiceRole(ctx, s.db, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM platform.users
			WHERE id = $1 AND deleted_at IS NULL AND is_active = true
		`, adminUserID).Scan(&count)
	})

	if err != nil {
		log.Debug().Err(err).Str("admin_user_id", adminUserID).Msg("Failed to check platform.users, falling back to auth.users")
	} else if count > 0 {
		// User exists in platform.users and is active
		log.Debug().Str("admin_user_id", adminUserID).Msg("Admin verified via platform.users")
		return nil
	}

	// Fall back to checking auth.users for users with instance_admin role
	adminUser, err := s.userRepo.GetByID(ctx, adminUserID)
	if err != nil {
		log.Debug().Err(err).Str("admin_user_id", adminUserID).Msg("Admin user not found in auth.users either")
		return fmt.Errorf("admin user not found: %w", err)
	}

	if adminUser.Role != "instance_admin" {
		log.Debug().Str("admin_user_id", adminUserID).Str("role", adminUser.Role).Msg("User is not a instance_admin")
		return ErrNotAdmin
	}

	log.Debug().Str("admin_user_id", adminUserID).Msg("Admin verified via auth.users with instance_admin role")
	return nil
}

// StartImpersonationRequest represents a request to start impersonating a user
type StartImpersonationRequest struct {
	TargetUserID string `json:"target_user_id"`
	Reason       string `json:"reason"`
	IPAddress    string `json:"-"` // Set from request context
	UserAgent    string `json:"-"` // Set from request context
}

// StartImpersonationResponse represents the response when starting impersonation
type StartImpersonationResponse struct {
	Session      *ImpersonationSession `json:"session"`
	TargetUser   *User                 `json:"target_user"`
	AccessToken  string                `json:"access_token"`
	RefreshToken string                `json:"refresh_token"`
	ExpiresIn    int64                 `json:"expires_in"`
}

// StartImpersonation starts an impersonation session for a specific user
func (s *ImpersonationService) StartImpersonation(
	ctx context.Context,
	adminUserID string,
	req StartImpersonationRequest,
) (*StartImpersonationResponse, error) {
	// Verify admin user exists and is admin (checks both platform.users and auth.users)
	if err := s.verifyAdminUser(ctx, adminUserID); err != nil {
		return nil, err
	}

	// Verify target user exists
	targetUser, err := s.userRepo.GetByID(ctx, req.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("target user not found: %w", err)
	}

	// Prevent self-impersonation
	if adminUserID == req.TargetUserID {
		return nil, ErrSelfImpersonation
	}

	// Check if admin already has an active impersonation session
	existingSession, err := s.repo.GetActiveByAdmin(ctx, adminUserID)
	if err == nil && existingSession != nil {
		// End the existing session first
		if err := s.repo.EndSession(ctx, existingSession.ID); err != nil {
			return nil, fmt.Errorf("failed to end existing session: %w", err)
		}
	}

	// Create new impersonation session
	targetUserID := targetUser.ID
	session := &ImpersonationSession{
		ID:                uuid.New().String(),
		AdminUserID:       adminUserID,
		TargetUserID:      &targetUserID,
		ImpersonationType: ImpersonationTypeUser,
		TargetRole:        &targetUser.Role,
		Reason:            req.Reason,
		StartedAt:         time.Now(),
		IPAddress:         req.IPAddress,
		UserAgent:         req.UserAgent,
		IsActive:          true,
	}

	createdSession, err := s.repo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create impersonation session: %w", err)
	}

	// Generate JWT tokens for the target user with their metadata
	// SECURITY: Include impersonated_by claim to mark these as impersonation tokens
	// Note: The JWT contains the target user's info, but we track admin in the session and token
	accessToken, accessClaims, err := s.jwtManager.GenerateAccessToken(targetUser.ID, targetUser.Email, targetUser.Role, targetUser.UserMetadata, targetUser.AppMetadata, WithImpersonatedBy(adminUserID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshClaims, err := s.jwtManager.GenerateRefreshToken(targetUser.ID, targetUser.Email, targetUser.Role, "", targetUser.UserMetadata, targetUser.AppMetadata, WithImpersonatedBy(adminUserID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with JTIs for later revocation
	createdSession.AccessTokenJTI = accessClaims.ID
	createdSession.RefreshTokenJTI = refreshClaims.ID

	return &StartImpersonationResponse{
		Session:      createdSession,
		TargetUser:   targetUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.accessTokenTTL.Seconds()),
	}, nil
}

// StartAnonImpersonation starts an impersonation session as an anonymous user
func (s *ImpersonationService) StartAnonImpersonation(
	ctx context.Context,
	adminUserID string,
	reason string,
	ipAddress string,
	userAgent string,
) (*StartImpersonationResponse, error) {
	// Verify admin user exists and is admin (checks both platform.users and auth.users)
	if err := s.verifyAdminUser(ctx, adminUserID); err != nil {
		return nil, err
	}

	// Check if admin already has an active impersonation session
	existingSession, err := s.repo.GetActiveByAdmin(ctx, adminUserID)
	if err == nil && existingSession != nil {
		// End the existing session first
		if err := s.repo.EndSession(ctx, existingSession.ID); err != nil {
			return nil, fmt.Errorf("failed to end existing session: %w", err)
		}
	}

	// Create new impersonation session
	anonRole := "anon"
	session := &ImpersonationSession{
		ID:                uuid.New().String(),
		AdminUserID:       adminUserID,
		TargetUserID:      nil, // No target user for anon
		ImpersonationType: ImpersonationTypeAnon,
		TargetRole:        &anonRole,
		Reason:            reason,
		StartedAt:         time.Now(),
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		IsActive:          true,
	}

	createdSession, err := s.repo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create impersonation session: %w", err)
	}

	// Generate JWT tokens for anonymous user (no metadata for anonymous users)
	// SECURITY: Include impersonated_by claim to mark these as impersonation tokens
	// Use well-known nil UUID for anonymous users
	accessToken, accessClaims, err := s.jwtManager.GenerateAccessToken(AnonUserID, "anonymous@fluxbase.local", "anon", nil, nil, WithImpersonatedBy(adminUserID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshClaims, err := s.jwtManager.GenerateRefreshToken(AnonUserID, "anonymous@fluxbase.local", "anon", "", nil, nil, WithImpersonatedBy(adminUserID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with JTIs for later revocation
	createdSession.AccessTokenJTI = accessClaims.ID
	createdSession.RefreshTokenJTI = refreshClaims.ID

	// Create a synthetic user object for response
	targetUser := &User{
		ID:    AnonUserID,
		Email: "anonymous@fluxbase.local",
		Role:  "anon",
	}

	return &StartImpersonationResponse{
		Session:      createdSession,
		TargetUser:   targetUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.accessTokenTTL.Seconds()),
	}, nil
}

// StartServiceImpersonation starts an impersonation session with service role
func (s *ImpersonationService) StartServiceImpersonation(
	ctx context.Context,
	adminUserID string,
	reason string,
	ipAddress string,
	userAgent string,
) (*StartImpersonationResponse, error) {
	// Verify admin user exists and is admin (checks both platform.users and auth.users)
	if err := s.verifyAdminUser(ctx, adminUserID); err != nil {
		return nil, err
	}

	// Check if admin already has an active impersonation session
	existingSession, err := s.repo.GetActiveByAdmin(ctx, adminUserID)
	if err == nil && existingSession != nil {
		// End the existing session first
		if err := s.repo.EndSession(ctx, existingSession.ID); err != nil {
			return nil, fmt.Errorf("failed to end existing session: %w", err)
		}
	}

	// Create new impersonation session
	serviceRole := "service"
	session := &ImpersonationSession{
		ID:                uuid.New().String(),
		AdminUserID:       adminUserID,
		TargetUserID:      nil, // No target user for service role
		ImpersonationType: ImpersonationTypeService,
		TargetRole:        &serviceRole,
		Reason:            reason,
		StartedAt:         time.Now(),
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		IsActive:          true,
	}

	createdSession, err := s.repo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create impersonation session: %w", err)
	}

	// Generate JWT tokens for service role (no metadata for service role)
	// SECURITY: Include impersonated_by claim to mark these as impersonation tokens
	// Use well-known UUID for service role users
	accessToken, accessClaims, err := s.jwtManager.GenerateAccessToken(ServiceUserID, "service@fluxbase.local", "service_role", nil, nil, WithImpersonatedBy(adminUserID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshClaims, err := s.jwtManager.GenerateRefreshToken(ServiceUserID, "service@fluxbase.local", "service_role", "", nil, nil, WithImpersonatedBy(adminUserID))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with JTIs for later revocation
	createdSession.AccessTokenJTI = accessClaims.ID
	createdSession.RefreshTokenJTI = refreshClaims.ID

	// Create a synthetic user object for response
	targetUser := &User{
		ID:    ServiceUserID,
		Email: "service@fluxbase.local",
		Role:  "service_role",
	}

	return &StartImpersonationResponse{
		Session:      createdSession,
		TargetUser:   targetUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.accessTokenTTL.Seconds()),
	}, nil
}

// StopImpersonation stops the active impersonation session for an admin
func (s *ImpersonationService) StopImpersonation(ctx context.Context, adminUserID string) error {
	// Get active session
	session, err := s.repo.GetActiveByAdmin(ctx, adminUserID)
	if err != nil {
		return err
	}

	// Revoke the access token if JTI is available
	if session.AccessTokenJTI != "" && s.tokenBlacklistService != nil {
		// Calculate expiry time - access tokens typically live for minutes to hours
		// We keep the blacklist entry until the token would naturally expire
		expiresAt := time.Now().Add(1 * time.Hour) // Conservative 1 hour for access tokens

		// Convert adminUserID to pointer for revokedBy parameter
		if err := s.tokenBlacklistService.repo.Add(ctx, session.AccessTokenJTI, &adminUserID, "impersonation_stopped", expiresAt); err != nil {
			// Log error but don't fail - session ending is more important
			log.Error().Err(err).Str("jti", session.AccessTokenJTI).Msg("Failed to blacklist access token during impersonation stop")
		} else {
			log.Info().Str("jti", session.AccessTokenJTI).Str("admin_user_id", adminUserID).Msg("Blacklisted impersonation access token")
		}
	}

	// Revoke the refresh token if JTI is available
	if session.RefreshTokenJTI != "" && s.tokenBlacklistService != nil {
		// Refresh tokens live longer, so blacklist for 24 hours
		expiresAt := time.Now().Add(24 * time.Hour)

		if err := s.tokenBlacklistService.repo.Add(ctx, session.RefreshTokenJTI, &adminUserID, "impersonation_stopped", expiresAt); err != nil {
			// Log error but don't fail - session ending is more important
			log.Error().Err(err).Str("jti", session.RefreshTokenJTI).Msg("Failed to blacklist refresh token during impersonation stop")
		} else {
			log.Info().Str("jti", session.RefreshTokenJTI).Str("admin_user_id", adminUserID).Msg("Blacklisted impersonation refresh token")
		}
	}

	// End the session
	return s.repo.EndSession(ctx, session.ID)
}

// GetActiveSession gets the active impersonation session for an admin
func (s *ImpersonationService) GetActiveSession(ctx context.Context, adminUserID string) (*ImpersonationSession, error) {
	return s.repo.GetActiveByAdmin(ctx, adminUserID)
}

// ListSessions lists impersonation sessions for audit purposes
func (s *ImpersonationService) ListSessions(ctx context.Context, adminUserID string, limit, offset int) ([]*ImpersonationSession, error) {
	return s.repo.ListByAdmin(ctx, adminUserID, limit, offset)
}
