package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Trust-related errors
var (
	ErrChallengeNotFound = errors.New("challenge not found")
	ErrChallengeExpired  = errors.New("challenge expired")
	ErrChallengeConsumed = errors.New("challenge already consumed")
	ErrChallengeMismatch = errors.New("challenge context mismatch")
	ErrTrustTokenInvalid = errors.New("trust token invalid")
	ErrTrustTokenExpired = errors.New("trust token expired")
)

// TrustSignal represents a single factor in trust calculation
type TrustSignal struct {
	Name   string `json:"name"`
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

// TrustResult contains the complete trust evaluation
type TrustResult struct {
	TotalScore      int           `json:"trust_score"`
	Signals         []TrustSignal `json:"signals,omitempty"`
	CaptchaRequired bool          `json:"captcha_required"`
	Reason          string        `json:"reason"`
}

// TrustRequest contains information for trust evaluation
type TrustRequest struct {
	UserID            *uuid.UUID
	Email             string
	IPAddress         string
	DeviceFingerprint string
	UserAgent         string
	TrustToken        string // Previously issued trust token
}

// CaptchaCheckRequest is the API request for checking if CAPTCHA is required
type CaptchaCheckRequest struct {
	Endpoint          string `json:"endpoint"`
	Email             string `json:"email,omitempty"`
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
	TrustToken        string `json:"trust_token,omitempty"`
}

// CaptchaCheckResponse is the API response for CAPTCHA check
type CaptchaCheckResponse struct {
	CaptchaRequired bool   `json:"captcha_required"`
	Reason          string `json:"reason,omitempty"`
	TrustScore      int    `json:"trust_score,omitempty"`

	// Widget configuration (only if captcha_required=true)
	Provider string `json:"provider,omitempty"`
	SiteKey  string `json:"site_key,omitempty"`

	// Challenge tracking
	ChallengeID string `json:"challenge_id"`
	ExpiresAt   string `json:"expires_at"`
}

// CaptchaChallenge represents a stored challenge
type CaptchaChallenge struct {
	ID                string
	ChallengeID       string
	Endpoint          string
	Email             string
	IPAddress         string
	DeviceFingerprint string
	UserAgent         string
	TrustScore        int
	CaptchaRequired   bool
	Reason            string
	CreatedAt         time.Time
	ExpiresAt         time.Time
	ConsumedAt        *time.Time
	CaptchaVerified   bool
}

// UserTrustSignal represents a stored trust signal for a user
type UserTrustSignal struct {
	ID                string
	UserID            uuid.UUID
	IPAddress         string
	DeviceFingerprint string
	UserAgent         string
	FirstSeenAt       time.Time
	LastSeenAt        time.Time
	SuccessfulLogins  int
	FailedAttempts    int
	LastCaptchaAt     *time.Time
	IsTrusted         bool
	IsBlocked         bool
}

// CaptchaTrustService handles adaptive CAPTCHA trust evaluation
type CaptchaTrustService struct {
	db             *pgxpool.Pool
	config         *config.AdaptiveTrustConfig
	captchaConfig  *config.CaptchaConfig
	captchaService *CaptchaService
}

// NewCaptchaTrustService creates a new trust service
func NewCaptchaTrustService(db *pgxpool.Pool, captchaConfig *config.CaptchaConfig, captchaService *CaptchaService) *CaptchaTrustService {
	return &CaptchaTrustService{
		db:             db,
		config:         &captchaConfig.AdaptiveTrust,
		captchaConfig:  captchaConfig,
		captchaService: captchaService,
	}
}

// IsEnabled returns whether adaptive trust is enabled
func (s *CaptchaTrustService) IsEnabled() bool {
	return s.config != nil && s.config.Enabled && s.captchaConfig.Enabled
}

// CheckCaptchaRequired evaluates trust and creates a challenge
func (s *CaptchaTrustService) CheckCaptchaRequired(ctx context.Context, req CaptchaCheckRequest, ipAddress, userAgent string) (*CaptchaCheckResponse, error) {
	// If adaptive trust is disabled, fall back to static behavior
	if !s.IsEnabled() {
		return s.staticCheck(ctx, req, ipAddress)
	}

	// Check if this endpoint always requires CAPTCHA
	if s.isAlwaysRequired(req.Endpoint) {
		return s.createChallenge(ctx, req, ipAddress, userAgent, 0, true, "sensitive_action")
	}

	// Build trust request
	trustReq := TrustRequest{
		Email:             req.Email,
		IPAddress:         ipAddress,
		DeviceFingerprint: req.DeviceFingerprint,
		UserAgent:         userAgent,
		TrustToken:        req.TrustToken,
	}

	// Calculate trust score
	result, err := s.CalculateTrust(ctx, trustReq)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to calculate trust, requiring CAPTCHA")
		return s.createChallenge(ctx, req, ipAddress, userAgent, 0, true, "trust_calculation_error")
	}

	return s.createChallenge(ctx, req, ipAddress, userAgent, result.TotalScore, result.CaptchaRequired, result.Reason)
}

// staticCheck performs static CAPTCHA check when adaptive trust is disabled
func (s *CaptchaTrustService) staticCheck(ctx context.Context, req CaptchaCheckRequest, ipAddress string) (*CaptchaCheckResponse, error) {
	required := s.captchaService.IsEnabledForEndpoint(req.Endpoint)

	response := &CaptchaCheckResponse{
		CaptchaRequired: required,
		ChallengeID:     generateChallengeID(),
		ExpiresAt:       time.Now().Add(5 * time.Minute).Format(time.RFC3339),
	}

	if required {
		response.Reason = "captcha_enabled_for_endpoint"
		response.Provider = s.captchaConfig.Provider
		response.SiteKey = s.captchaConfig.SiteKey
	} else {
		response.Reason = "captcha_disabled"
	}

	// Store the challenge for validation
	if err := s.storeChallenge(ctx, response.ChallengeID, req.Endpoint, req.Email, ipAddress, req.DeviceFingerprint, "", 100, required, response.Reason); err != nil {
		log.Warn().Err(err).Msg("Failed to store challenge, continuing without")
	}

	return response, nil
}

// CalculateTrust evaluates trust signals and returns a trust result
func (s *CaptchaTrustService) CalculateTrust(ctx context.Context, req TrustRequest) (*TrustResult, error) {
	result := &TrustResult{
		Signals: make([]TrustSignal, 0),
	}

	// Check for valid trust token first (highest priority)
	if req.TrustToken != "" {
		valid, err := s.ValidateTrustToken(ctx, req.TrustToken, req.IPAddress, req.DeviceFingerprint)
		if err == nil && valid {
			result.Signals = append(result.Signals, TrustSignal{
				Name:   "recent_captcha",
				Score:  s.config.WeightRecentCaptcha,
				Reason: "Valid trust token from recent CAPTCHA",
			})
			result.TotalScore += s.config.WeightRecentCaptcha
		}
	}

	// Try to find user by email
	var user *User
	var userTrustSignal *UserTrustSignal
	if req.Email != "" {
		user, _ = s.getUserByEmail(ctx, req.Email)
		if user != nil {
			if userUUID, err := uuid.Parse(user.ID); err == nil {
				req.UserID = &userUUID
				userTrustSignal, _ = s.getTrustSignal(ctx, userUUID, req.IPAddress, req.DeviceFingerprint)
			}
		}
	}

	// Evaluate signals based on user existence
	if user != nil {
		// Known user signals
		s.evaluateKnownUserSignals(ctx, result, user, userTrustSignal, req)
	} else {
		// Unknown user (signup or unknown email)
		s.evaluateUnknownUserSignals(result, req)
	}

	// Calculate final result
	result.TotalScore = 0
	for _, signal := range result.Signals {
		result.TotalScore += signal.Score
	}

	// Determine if CAPTCHA is required
	result.CaptchaRequired = result.TotalScore < s.config.CaptchaThreshold
	if result.CaptchaRequired {
		result.Reason = s.determineReason(result.Signals)
	} else {
		result.Reason = "trusted"
	}

	return result, nil
}

// evaluateKnownUserSignals adds trust signals for known users
func (s *CaptchaTrustService) evaluateKnownUserSignals(ctx context.Context, result *TrustResult, user *User, trustSignal *UserTrustSignal, req TrustRequest) {
	// Check verified email
	if user.EmailVerified {
		result.Signals = append(result.Signals, TrustSignal{
			Name:   "verified_email",
			Score:  s.config.WeightVerifiedEmail,
			Reason: "Email address is verified",
		})
	}

	// Check account age
	if time.Since(user.CreatedAt) > 7*24*time.Hour {
		result.Signals = append(result.Signals, TrustSignal{
			Name:   "account_age",
			Score:  s.config.WeightAccountAge,
			Reason: "Account older than 7 days",
		})
	}

	// Check MFA
	userUUID, _ := uuid.Parse(user.ID)
	if s.userHasMFA(ctx, userUUID) {
		result.Signals = append(result.Signals, TrustSignal{
			Name:   "mfa_enabled",
			Score:  s.config.WeightMFAEnabled,
			Reason: "MFA is enabled",
		})
	}

	// Check trust signal history
	if trustSignal != nil {
		// Known IP
		result.Signals = append(result.Signals, TrustSignal{
			Name:   "known_ip",
			Score:  s.config.WeightKnownIP,
			Reason: fmt.Sprintf("IP seen %d times", trustSignal.SuccessfulLogins),
		})

		// Known device (if fingerprint matches)
		if req.DeviceFingerprint != "" && trustSignal.DeviceFingerprint == req.DeviceFingerprint {
			result.Signals = append(result.Signals, TrustSignal{
				Name:   "known_device",
				Score:  s.config.WeightKnownDevice,
				Reason: "Device fingerprint recognized",
			})
		}

		// Successful logins
		if trustSignal.SuccessfulLogins >= 3 {
			result.Signals = append(result.Signals, TrustSignal{
				Name:   "successful_logins",
				Score:  s.config.WeightSuccessfulLogins,
				Reason: fmt.Sprintf("%d successful logins", trustSignal.SuccessfulLogins),
			})
		}

		// Recent failed attempts (negative)
		if trustSignal.FailedAttempts > 0 {
			result.Signals = append(result.Signals, TrustSignal{
				Name:   "failed_attempts",
				Score:  s.config.WeightFailedAttempts * trustSignal.FailedAttempts,
				Reason: fmt.Sprintf("%d recent failed attempts", trustSignal.FailedAttempts),
			})
		}

		// Recent CAPTCHA solve
		if trustSignal.LastCaptchaAt != nil && time.Since(*trustSignal.LastCaptchaAt) < s.config.TrustTokenTTL {
			result.Signals = append(result.Signals, TrustSignal{
				Name:   "recent_captcha",
				Score:  s.config.WeightRecentCaptcha,
				Reason: "CAPTCHA solved recently",
			})
		}

		// Explicitly blocked
		if trustSignal.IsBlocked {
			result.Signals = append(result.Signals, TrustSignal{
				Name:   "blocked",
				Score:  -1000,
				Reason: "IP/device explicitly blocked",
			})
		}

		// Explicitly trusted
		if trustSignal.IsTrusted {
			result.Signals = append(result.Signals, TrustSignal{
				Name:   "admin_trusted",
				Score:  100,
				Reason: "Explicitly trusted by admin",
			})
		}
	} else {
		// New IP for known user
		result.Signals = append(result.Signals, TrustSignal{
			Name:   "new_ip",
			Score:  s.config.WeightNewIP,
			Reason: "New IP address for this user",
		})

		// New device
		if req.DeviceFingerprint != "" {
			result.Signals = append(result.Signals, TrustSignal{
				Name:   "new_device",
				Score:  s.config.WeightNewDevice,
				Reason: "New device for this user",
			})
		}
	}
}

// evaluateUnknownUserSignals adds trust signals for unknown users (signup)
func (s *CaptchaTrustService) evaluateUnknownUserSignals(result *TrustResult, req TrustRequest) {
	// New user, no trust history
	result.Signals = append(result.Signals, TrustSignal{
		Name:   "no_account",
		Score:  s.config.WeightNewIP,
		Reason: "No existing account found",
	})

	if req.DeviceFingerprint != "" {
		result.Signals = append(result.Signals, TrustSignal{
			Name:   "new_device",
			Score:  s.config.WeightNewDevice,
			Reason: "Unknown device",
		})
	}
}

// determineReason picks the primary reason CAPTCHA is required
func (s *CaptchaTrustService) determineReason(signals []TrustSignal) string {
	// Find the most negative signal
	var worstSignal *TrustSignal
	for i := range signals {
		if signals[i].Score < 0 {
			if worstSignal == nil || signals[i].Score < worstSignal.Score {
				worstSignal = &signals[i]
			}
		}
	}

	if worstSignal != nil {
		return worstSignal.Name
	}
	return "low_trust_score"
}

// isAlwaysRequired checks if endpoint is in the always-require list
func (s *CaptchaTrustService) isAlwaysRequired(endpoint string) bool {
	for _, e := range s.config.AlwaysRequireEndpoints {
		if e == endpoint {
			return true
		}
	}
	return false
}

// createChallenge creates and stores a challenge, returning the response
func (s *CaptchaTrustService) createChallenge(ctx context.Context, req CaptchaCheckRequest, ipAddress, userAgent string, trustScore int, captchaRequired bool, reason string) (*CaptchaCheckResponse, error) {
	challengeID := generateChallengeID()
	expiresAt := time.Now().Add(s.config.ChallengeExpiry)

	response := &CaptchaCheckResponse{
		CaptchaRequired: captchaRequired,
		Reason:          reason,
		TrustScore:      trustScore,
		ChallengeID:     challengeID,
		ExpiresAt:       expiresAt.Format(time.RFC3339),
	}

	if captchaRequired {
		response.Provider = s.captchaConfig.Provider
		response.SiteKey = s.captchaConfig.SiteKey
	}

	// Store challenge
	if err := s.storeChallenge(ctx, challengeID, req.Endpoint, req.Email, ipAddress, req.DeviceFingerprint, userAgent, trustScore, captchaRequired, reason); err != nil {
		log.Warn().Err(err).Msg("Failed to store challenge")
		// Continue anyway - challenge validation will just be skipped
	}

	return response, nil
}

// storeChallenge saves a challenge to the database
func (s *CaptchaTrustService) storeChallenge(ctx context.Context, challengeID, endpoint, email, ipAddress, deviceFingerprint, userAgent string, trustScore int, captchaRequired bool, reason string) error {
	expiresAt := time.Now().Add(s.config.ChallengeExpiry)
	if s.config.ChallengeExpiry == 0 {
		expiresAt = time.Now().Add(5 * time.Minute)
	}

	query := `
		INSERT INTO auth.captcha_challenges
		(challenge_id, endpoint, email, ip_address, device_fingerprint, user_agent, trust_score, captcha_required, reason, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := s.db.Exec(ctx, query, challengeID, endpoint, email, ipAddress, deviceFingerprint, userAgent, trustScore, captchaRequired, reason, expiresAt)
	return err
}

// ValidateChallenge checks if a challenge is valid and optionally consumes it
func (s *CaptchaTrustService) ValidateChallenge(ctx context.Context, challengeID, endpoint, ipAddress string, captchaVerified bool) error {
	if challengeID == "" {
		// No challenge ID provided - skip validation (backwards compatibility)
		return nil
	}

	query := `
		SELECT id, endpoint, ip_address, captcha_required, expires_at, consumed_at
		FROM auth.captcha_challenges
		WHERE challenge_id = $1
	`

	var challenge struct {
		ID              uuid.UUID
		Endpoint        string
		IPAddress       net.IP
		CaptchaRequired bool
		ExpiresAt       time.Time
		ConsumedAt      *time.Time
	}

	err := s.db.QueryRow(ctx, query, challengeID).Scan(
		&challenge.ID, &challenge.Endpoint, &challenge.IPAddress,
		&challenge.CaptchaRequired, &challenge.ExpiresAt, &challenge.ConsumedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChallengeNotFound
		}
		return fmt.Errorf("failed to get challenge: %w", err)
	}

	// Check if expired
	if time.Now().After(challenge.ExpiresAt) {
		return ErrChallengeExpired
	}

	// Check if already consumed
	if challenge.ConsumedAt != nil {
		return ErrChallengeConsumed
	}

	// Check endpoint matches
	if challenge.Endpoint != endpoint {
		return ErrChallengeMismatch
	}

	// Check IP matches (if configured)
	if s.config.TrustTokenBoundIP && challenge.IPAddress.String() != ipAddress {
		return ErrChallengeMismatch
	}

	// Check if CAPTCHA was required but not verified
	if challenge.CaptchaRequired && !captchaVerified {
		return ErrCaptchaRequired
	}

	// Mark challenge as consumed
	updateQuery := `
		UPDATE auth.captcha_challenges
		SET consumed_at = NOW(), captcha_verified = $2
		WHERE challenge_id = $1
	`
	_, err = s.db.Exec(ctx, updateQuery, challengeID, captchaVerified)
	if err != nil {
		log.Warn().Err(err).Str("challenge_id", challengeID).Msg("Failed to mark challenge as consumed")
	}

	return nil
}

// GetChallenge retrieves a challenge by ID
func (s *CaptchaTrustService) GetChallenge(ctx context.Context, challengeID string) (*CaptchaChallenge, error) {
	query := `
		SELECT id, challenge_id, endpoint, email, ip_address, device_fingerprint, user_agent,
		       trust_score, captcha_required, reason, created_at, expires_at, consumed_at, captcha_verified
		FROM auth.captcha_challenges
		WHERE challenge_id = $1
	`

	var c CaptchaChallenge
	var id uuid.UUID
	var ip net.IP
	err := s.db.QueryRow(ctx, query, challengeID).Scan(
		&id, &c.ChallengeID, &c.Endpoint, &c.Email, &ip, &c.DeviceFingerprint, &c.UserAgent,
		&c.TrustScore, &c.CaptchaRequired, &c.Reason, &c.CreatedAt, &c.ExpiresAt, &c.ConsumedAt, &c.CaptchaVerified,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChallengeNotFound
		}
		return nil, err
	}

	c.ID = id.String()
	c.IPAddress = ip.String()
	return &c, nil
}

// IssueTrustToken creates a trust token after successful CAPTCHA verification
func (s *CaptchaTrustService) IssueTrustToken(ctx context.Context, ipAddress, deviceFingerprint, userAgent string) (string, error) {
	token := generateTrustToken()
	tokenHash := hashTrustToken(token)
	expiresAt := time.Now().Add(s.config.TrustTokenTTL)
	if s.config.TrustTokenTTL == 0 {
		expiresAt = time.Now().Add(15 * time.Minute)
	}

	query := `
		INSERT INTO auth.captcha_trust_tokens
		(token_hash, ip_address, device_fingerprint, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.Exec(ctx, query, tokenHash, ipAddress, deviceFingerprint, userAgent, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to store trust token: %w", err)
	}

	return token, nil
}

// ValidateTrustToken checks if a trust token is valid
func (s *CaptchaTrustService) ValidateTrustToken(ctx context.Context, token, ipAddress, deviceFingerprint string) (bool, error) {
	tokenHash := hashTrustToken(token)

	query := `
		SELECT ip_address, device_fingerprint, expires_at
		FROM auth.captcha_trust_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	var storedIP net.IP
	var storedFingerprint *string
	var expiresAt time.Time

	err := s.db.QueryRow(ctx, query, tokenHash).Scan(&storedIP, &storedFingerprint, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrTrustTokenInvalid
		}
		return false, err
	}

	// Check IP binding
	if s.config.TrustTokenBoundIP && storedIP.String() != ipAddress {
		return false, ErrTrustTokenInvalid
	}

	// Update usage stats
	updateQuery := `
		UPDATE auth.captcha_trust_tokens
		SET used_count = used_count + 1, last_used_at = NOW()
		WHERE token_hash = $1
	`
	_, _ = s.db.Exec(ctx, updateQuery, tokenHash)

	return true, nil
}

// RecordSuccessfulLogin updates trust signals after a successful login
func (s *CaptchaTrustService) RecordSuccessfulLogin(ctx context.Context, userID uuid.UUID, ipAddress, deviceFingerprint, userAgent string) error {
	query := `
		INSERT INTO auth.user_trust_signals
		(user_id, ip_address, device_fingerprint, user_agent, successful_logins, last_seen_at)
		VALUES ($1, $2, $3, $4, 1, NOW())
		ON CONFLICT (user_id, ip_address, COALESCE(device_fingerprint, ''))
		DO UPDATE SET
			successful_logins = auth.user_trust_signals.successful_logins + 1,
			failed_attempts = 0,  -- Reset failed attempts on success
			last_seen_at = NOW(),
			user_agent = EXCLUDED.user_agent
	`
	_, err := s.db.Exec(ctx, query, userID, ipAddress, deviceFingerprint, userAgent)
	return err
}

// RecordFailedAttempt updates trust signals after a failed login attempt
func (s *CaptchaTrustService) RecordFailedAttempt(ctx context.Context, userID *uuid.UUID, ipAddress, deviceFingerprint, userAgent string) error {
	if userID == nil {
		// Can't record without a user ID
		return nil
	}

	query := `
		INSERT INTO auth.user_trust_signals
		(user_id, ip_address, device_fingerprint, user_agent, failed_attempts, last_seen_at)
		VALUES ($1, $2, $3, $4, 1, NOW())
		ON CONFLICT (user_id, ip_address, COALESCE(device_fingerprint, ''))
		DO UPDATE SET
			failed_attempts = auth.user_trust_signals.failed_attempts + 1,
			last_seen_at = NOW()
	`
	_, err := s.db.Exec(ctx, query, userID, ipAddress, deviceFingerprint, userAgent)
	return err
}

// RecordCaptchaSolved updates trust signals after a successful CAPTCHA
func (s *CaptchaTrustService) RecordCaptchaSolved(ctx context.Context, userID *uuid.UUID, ipAddress, deviceFingerprint, userAgent string) error {
	if userID == nil {
		return nil
	}

	query := `
		INSERT INTO auth.user_trust_signals
		(user_id, ip_address, device_fingerprint, user_agent, last_captcha_at, last_seen_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (user_id, ip_address, COALESCE(device_fingerprint, ''))
		DO UPDATE SET
			last_captcha_at = NOW(),
			last_seen_at = NOW()
	`
	_, err := s.db.Exec(ctx, query, userID, ipAddress, deviceFingerprint, userAgent)
	return err
}

// Helper functions

func (s *CaptchaTrustService) getUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, email_verified, created_at
		FROM auth.users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user User
	err := s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.EmailVerified, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *CaptchaTrustService) getTrustSignal(ctx context.Context, userID uuid.UUID, ipAddress, deviceFingerprint string) (*UserTrustSignal, error) {
	query := `
		SELECT id, user_id, ip_address, device_fingerprint, user_agent,
		       first_seen_at, last_seen_at, successful_logins, failed_attempts,
		       last_captcha_at, is_trusted, is_blocked
		FROM auth.user_trust_signals
		WHERE user_id = $1 AND ip_address = $2
	`

	var signal UserTrustSignal
	var id uuid.UUID
	var uid uuid.UUID
	var ip net.IP
	err := s.db.QueryRow(ctx, query, userID, ipAddress).Scan(
		&id, &uid, &ip, &signal.DeviceFingerprint, &signal.UserAgent,
		&signal.FirstSeenAt, &signal.LastSeenAt, &signal.SuccessfulLogins, &signal.FailedAttempts,
		&signal.LastCaptchaAt, &signal.IsTrusted, &signal.IsBlocked,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	signal.ID = id.String()
	signal.UserID = uid
	signal.IPAddress = ip.String()
	return &signal, nil
}

func (s *CaptchaTrustService) userHasMFA(ctx context.Context, userID uuid.UUID) bool {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM auth.mfa_factors
			WHERE user_id = $1 AND status = 'verified'
		)
	`
	var hasMFA bool
	_ = s.db.QueryRow(ctx, query, userID).Scan(&hasMFA)
	return hasMFA
}

func generateChallengeID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "ch_" + hex.EncodeToString(b)
}

func generateTrustToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "tt_" + hex.EncodeToString(b)
}

func hashTrustToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
