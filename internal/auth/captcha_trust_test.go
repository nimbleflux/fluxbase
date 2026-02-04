package auth

import (
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTrustSignal_Struct(t *testing.T) {
	signal := TrustSignal{
		Name:   "verified_email",
		Score:  20,
		Reason: "Email address is verified",
	}

	assert.Equal(t, "verified_email", signal.Name)
	assert.Equal(t, 20, signal.Score)
	assert.Equal(t, "Email address is verified", signal.Reason)
}

func TestTrustResult_Struct(t *testing.T) {
	result := TrustResult{
		TotalScore: 75,
		Signals: []TrustSignal{
			{Name: "verified_email", Score: 20},
			{Name: "account_age", Score: 15},
			{Name: "known_ip", Score: 40},
		},
		CaptchaRequired: false,
		Reason:          "trusted",
	}

	assert.Equal(t, 75, result.TotalScore)
	assert.Len(t, result.Signals, 3)
	assert.False(t, result.CaptchaRequired)
	assert.Equal(t, "trusted", result.Reason)
}

func TestTrustRequest_Struct(t *testing.T) {
	userID := uuid.New()
	req := TrustRequest{
		UserID:            &userID,
		Email:             "user@example.com",
		IPAddress:         "192.168.1.1",
		DeviceFingerprint: "fp_abc123",
		UserAgent:         "Mozilla/5.0",
		TrustToken:        "tt_token123",
	}

	assert.Equal(t, &userID, req.UserID)
	assert.Equal(t, "user@example.com", req.Email)
	assert.Equal(t, "192.168.1.1", req.IPAddress)
	assert.Equal(t, "fp_abc123", req.DeviceFingerprint)
	assert.Equal(t, "Mozilla/5.0", req.UserAgent)
	assert.Equal(t, "tt_token123", req.TrustToken)
}

func TestCaptchaCheckRequest_Struct(t *testing.T) {
	req := CaptchaCheckRequest{
		Endpoint:          "/auth/signup",
		Email:             "user@example.com",
		DeviceFingerprint: "fp_xyz789",
		TrustToken:        "tt_token456",
	}

	assert.Equal(t, "/auth/signup", req.Endpoint)
	assert.Equal(t, "user@example.com", req.Email)
	assert.Equal(t, "fp_xyz789", req.DeviceFingerprint)
	assert.Equal(t, "tt_token456", req.TrustToken)
}

func TestCaptchaCheckResponse_Struct(t *testing.T) {
	t.Run("captcha required response", func(t *testing.T) {
		resp := CaptchaCheckResponse{
			CaptchaRequired: true,
			Reason:          "low_trust_score",
			TrustScore:      35,
			Provider:        "recaptcha",
			SiteKey:         "site-key-123",
			ChallengeID:     "ch_abc123def456",
			ExpiresAt:       "2024-01-15T10:30:00Z",
		}

		assert.True(t, resp.CaptchaRequired)
		assert.Equal(t, "low_trust_score", resp.Reason)
		assert.Equal(t, 35, resp.TrustScore)
		assert.Equal(t, "recaptcha", resp.Provider)
		assert.Equal(t, "site-key-123", resp.SiteKey)
		assert.NotEmpty(t, resp.ChallengeID)
	})

	t.Run("captcha not required response", func(t *testing.T) {
		resp := CaptchaCheckResponse{
			CaptchaRequired: false,
			Reason:          "trusted",
			TrustScore:      85,
			ChallengeID:     "ch_abc123def456",
			ExpiresAt:       "2024-01-15T10:30:00Z",
		}

		assert.False(t, resp.CaptchaRequired)
		assert.Equal(t, "trusted", resp.Reason)
		assert.Empty(t, resp.Provider)
		assert.Empty(t, resp.SiteKey)
	})
}

func TestCaptchaChallenge_Struct(t *testing.T) {
	now := time.Now()
	consumedAt := now.Add(5 * time.Minute)
	expiresAt := now.Add(10 * time.Minute)

	challenge := CaptchaChallenge{
		ID:                "uuid-123",
		ChallengeID:       "ch_challenge123",
		Endpoint:          "/auth/login",
		Email:             "user@example.com",
		IPAddress:         "192.168.1.100",
		DeviceFingerprint: "fp_device456",
		UserAgent:         "Chrome/120.0",
		TrustScore:        45,
		CaptchaRequired:   true,
		Reason:            "new_ip",
		CreatedAt:         now,
		ExpiresAt:         expiresAt,
		ConsumedAt:        &consumedAt,
		CaptchaVerified:   true,
	}

	assert.Equal(t, "uuid-123", challenge.ID)
	assert.Equal(t, "ch_challenge123", challenge.ChallengeID)
	assert.Equal(t, "/auth/login", challenge.Endpoint)
	assert.Equal(t, "user@example.com", challenge.Email)
	assert.Equal(t, "192.168.1.100", challenge.IPAddress)
	assert.Equal(t, 45, challenge.TrustScore)
	assert.True(t, challenge.CaptchaRequired)
	assert.True(t, challenge.CaptchaVerified)
	assert.NotNil(t, challenge.ConsumedAt)
}

func TestUserTrustSignal_Struct(t *testing.T) {
	now := time.Now()
	lastCaptcha := now.Add(-1 * time.Hour)

	signal := UserTrustSignal{
		ID:                "signal-uuid",
		UserID:            uuid.New(),
		IPAddress:         "10.0.0.1",
		DeviceFingerprint: "fp_trusted_device",
		UserAgent:         "Firefox/115.0",
		FirstSeenAt:       now.Add(-30 * 24 * time.Hour),
		LastSeenAt:        now,
		SuccessfulLogins:  25,
		FailedAttempts:    2,
		LastCaptchaAt:     &lastCaptcha,
		IsTrusted:         true,
		IsBlocked:         false,
	}

	assert.Equal(t, "signal-uuid", signal.ID)
	assert.Equal(t, "10.0.0.1", signal.IPAddress)
	assert.Equal(t, "fp_trusted_device", signal.DeviceFingerprint)
	assert.Equal(t, 25, signal.SuccessfulLogins)
	assert.Equal(t, 2, signal.FailedAttempts)
	assert.True(t, signal.IsTrusted)
	assert.False(t, signal.IsBlocked)
	assert.NotNil(t, signal.LastCaptchaAt)
}

func TestTrustErrors(t *testing.T) {
	t.Run("error constants are defined", func(t *testing.T) {
		assert.Error(t, ErrChallengeNotFound)
		assert.Error(t, ErrChallengeExpired)
		assert.Error(t, ErrChallengeConsumed)
		assert.Error(t, ErrChallengeMismatch)
		assert.Error(t, ErrTrustTokenInvalid)
		assert.Error(t, ErrTrustTokenExpired)
	})

	t.Run("error messages are descriptive", func(t *testing.T) {
		assert.Contains(t, ErrChallengeNotFound.Error(), "not found")
		assert.Contains(t, ErrChallengeExpired.Error(), "expired")
		assert.Contains(t, ErrChallengeConsumed.Error(), "consumed")
		assert.Contains(t, ErrChallengeMismatch.Error(), "mismatch")
		assert.Contains(t, ErrTrustTokenInvalid.Error(), "invalid")
		assert.Contains(t, ErrTrustTokenExpired.Error(), "expired")
	})
}

func TestGenerateChallengeID(t *testing.T) {
	t.Run("generates challenge ID with correct prefix", func(t *testing.T) {
		id := generateChallengeID()
		assert.True(t, len(id) > 3)
		assert.Equal(t, "ch_", id[:3])
	})

	t.Run("generates unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := generateChallengeID()
			assert.False(t, ids[id], "Duplicate challenge ID generated")
			ids[id] = true
		}
	})
}

func TestGenerateTrustToken(t *testing.T) {
	t.Run("generates trust token with correct prefix", func(t *testing.T) {
		token := generateTrustToken()
		assert.True(t, len(token) > 3)
		assert.Equal(t, "tt_", token[:3])
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		tokens := make(map[string]bool)
		for i := 0; i < 100; i++ {
			token := generateTrustToken()
			assert.False(t, tokens[token], "Duplicate trust token generated")
			tokens[token] = true
		}
	})
}

func TestHashTrustToken(t *testing.T) {
	t.Run("produces consistent hash for same token", func(t *testing.T) {
		token := "tt_test_token_123"
		hash1 := hashTrustToken(token)
		hash2 := hashTrustToken(token)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("produces different hashes for different tokens", func(t *testing.T) {
		hash1 := hashTrustToken("tt_token_1")
		hash2 := hashTrustToken("tt_token_2")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("hash is hex encoded", func(t *testing.T) {
		hash := hashTrustToken("test")
		// SHA256 produces 32 bytes = 64 hex characters
		assert.Len(t, hash, 64)
		// Should only contain hex characters
		for _, c := range hash {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'))
		}
	})
}

func TestCaptchaTrustService_IsEnabled(t *testing.T) {
	t.Run("returns false when config is nil", func(t *testing.T) {
		service := &CaptchaTrustService{
			config:        nil,
			captchaConfig: &config.CaptchaConfig{Enabled: true},
		}
		assert.False(t, service.IsEnabled())
	})

	t.Run("returns false when adaptive trust is disabled", func(t *testing.T) {
		service := &CaptchaTrustService{
			config:        &config.AdaptiveTrustConfig{Enabled: false},
			captchaConfig: &config.CaptchaConfig{Enabled: true},
		}
		assert.False(t, service.IsEnabled())
	})

	t.Run("returns false when captcha is disabled", func(t *testing.T) {
		service := &CaptchaTrustService{
			config:        &config.AdaptiveTrustConfig{Enabled: true},
			captchaConfig: &config.CaptchaConfig{Enabled: false},
		}
		assert.False(t, service.IsEnabled())
	})

	t.Run("returns true when both are enabled", func(t *testing.T) {
		service := &CaptchaTrustService{
			config:        &config.AdaptiveTrustConfig{Enabled: true},
			captchaConfig: &config.CaptchaConfig{Enabled: true},
		}
		assert.True(t, service.IsEnabled())
	})
}

func TestCaptchaTrustService_IsAlwaysRequired(t *testing.T) {
	service := &CaptchaTrustService{
		config: &config.AdaptiveTrustConfig{
			AlwaysRequireEndpoints: []string{"/auth/signup", "/auth/reset-password"},
		},
	}

	t.Run("returns true for always-required endpoints", func(t *testing.T) {
		assert.True(t, service.isAlwaysRequired("/auth/signup"))
		assert.True(t, service.isAlwaysRequired("/auth/reset-password"))
	})

	t.Run("returns false for other endpoints", func(t *testing.T) {
		assert.False(t, service.isAlwaysRequired("/auth/login"))
		assert.False(t, service.isAlwaysRequired("/api/users"))
	})
}

func TestCaptchaTrustService_DetermineReason(t *testing.T) {
	service := &CaptchaTrustService{}

	t.Run("returns worst signal name when negative signals exist", func(t *testing.T) {
		signals := []TrustSignal{
			{Name: "verified_email", Score: 20},
			{Name: "failed_attempts", Score: -30},
			{Name: "new_ip", Score: -10},
		}

		reason := service.determineReason(signals)
		assert.Equal(t, "failed_attempts", reason) // Most negative
	})

	t.Run("returns low_trust_score when no negative signals", func(t *testing.T) {
		signals := []TrustSignal{
			{Name: "new_ip", Score: 0},
			{Name: "no_account", Score: 5},
		}

		reason := service.determineReason(signals)
		assert.Equal(t, "low_trust_score", reason)
	})

	t.Run("handles empty signals", func(t *testing.T) {
		reason := service.determineReason([]TrustSignal{})
		assert.Equal(t, "low_trust_score", reason)
	})
}

func TestCaptchaTrustService_EvaluateUnknownUserSignals(t *testing.T) {
	service := &CaptchaTrustService{
		config: &config.AdaptiveTrustConfig{
			WeightNewIP:     -15,
			WeightNewDevice: -10,
		},
	}

	t.Run("adds signals for unknown user without device", func(t *testing.T) {
		result := &TrustResult{Signals: []TrustSignal{}}
		req := TrustRequest{
			Email:     "new@example.com",
			IPAddress: "1.2.3.4",
		}

		service.evaluateUnknownUserSignals(result, req)

		assert.Len(t, result.Signals, 1)
		assert.Equal(t, "no_account", result.Signals[0].Name)
	})

	t.Run("adds device signal when fingerprint provided", func(t *testing.T) {
		result := &TrustResult{Signals: []TrustSignal{}}
		req := TrustRequest{
			Email:             "new@example.com",
			IPAddress:         "1.2.3.4",
			DeviceFingerprint: "fp_new_device",
		}

		service.evaluateUnknownUserSignals(result, req)

		assert.Len(t, result.Signals, 2)
		signalNames := []string{result.Signals[0].Name, result.Signals[1].Name}
		assert.Contains(t, signalNames, "no_account")
		assert.Contains(t, signalNames, "new_device")
	})
}

// =============================================================================
// NewCaptchaTrustService Tests
// =============================================================================

func TestNewCaptchaTrustService(t *testing.T) {
	t.Run("creates service with nil database", func(t *testing.T) {
		captchaConfig := &config.CaptchaConfig{
			Enabled:       true,
			AdaptiveTrust: config.AdaptiveTrustConfig{Enabled: true},
		}
		svc := NewCaptchaTrustService(nil, captchaConfig, nil)

		assert.NotNil(t, svc)
		assert.Nil(t, svc.db)
		assert.NotNil(t, svc.config)
		assert.NotNil(t, svc.captchaConfig)
	})

	t.Run("creates service with all dependencies", func(t *testing.T) {
		captchaConfig := &config.CaptchaConfig{
			Enabled:       true,
			AdaptiveTrust: config.AdaptiveTrustConfig{Enabled: true},
		}
		captchaService := &CaptchaService{}

		svc := NewCaptchaTrustService(nil, captchaConfig, captchaService)

		assert.NotNil(t, svc)
		assert.NotNil(t, svc.captchaService)
	})
}

// =============================================================================
// TrustResult Additional Tests
// =============================================================================

func TestTrustResult_Defaults(t *testing.T) {
	result := TrustResult{}

	assert.Equal(t, 0, result.TotalScore)
	assert.Nil(t, result.Signals)
	assert.False(t, result.CaptchaRequired)
	assert.Empty(t, result.Reason)
}

func TestTrustResult_HighTrustScore(t *testing.T) {
	result := TrustResult{
		TotalScore:      95,
		CaptchaRequired: false,
		Reason:          "trusted",
		Signals: []TrustSignal{
			{Name: "verified_email", Score: 20},
			{Name: "account_age", Score: 25},
			{Name: "known_device", Score: 30},
			{Name: "successful_logins", Score: 20},
		},
	}

	assert.True(t, result.TotalScore >= 90)
	assert.False(t, result.CaptchaRequired)
	assert.Len(t, result.Signals, 4)
}

func TestTrustResult_LowTrustScore(t *testing.T) {
	result := TrustResult{
		TotalScore:      15,
		CaptchaRequired: true,
		Reason:          "new_device",
		Signals: []TrustSignal{
			{Name: "new_ip", Score: -10},
			{Name: "new_device", Score: -15},
			{Name: "no_history", Score: -5},
		},
	}

	assert.True(t, result.TotalScore < 50)
	assert.True(t, result.CaptchaRequired)
}

// =============================================================================
// CaptchaChallenge Additional Tests
// =============================================================================

func TestCaptchaChallenge_Defaults(t *testing.T) {
	challenge := CaptchaChallenge{}

	assert.Empty(t, challenge.ID)
	assert.Empty(t, challenge.ChallengeID)
	assert.Empty(t, challenge.Endpoint)
	assert.Empty(t, challenge.Email)
	assert.False(t, challenge.CaptchaRequired)
	assert.False(t, challenge.CaptchaVerified)
	assert.Nil(t, challenge.ConsumedAt)
}

func TestCaptchaChallenge_ExpiryCheck(t *testing.T) {
	t.Run("challenge not expired", func(t *testing.T) {
		challenge := CaptchaChallenge{
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		isExpired := time.Now().After(challenge.ExpiresAt)
		assert.False(t, isExpired)
	})

	t.Run("challenge expired", func(t *testing.T) {
		challenge := CaptchaChallenge{
			ExpiresAt: time.Now().Add(-10 * time.Minute),
		}

		isExpired := time.Now().After(challenge.ExpiresAt)
		assert.True(t, isExpired)
	})
}

// =============================================================================
// UserTrustSignal Additional Tests
// =============================================================================

func TestUserTrustSignal_Defaults(t *testing.T) {
	signal := UserTrustSignal{}

	assert.Empty(t, signal.ID)
	assert.Empty(t, signal.IPAddress)
	assert.Empty(t, signal.DeviceFingerprint)
	assert.Equal(t, 0, signal.SuccessfulLogins)
	assert.Equal(t, 0, signal.FailedAttempts)
	assert.False(t, signal.IsTrusted)
	assert.False(t, signal.IsBlocked)
	assert.Nil(t, signal.LastCaptchaAt)
}

func TestUserTrustSignal_BlockedUser(t *testing.T) {
	signal := UserTrustSignal{
		ID:             "signal-blocked",
		UserID:         uuid.New(),
		FailedAttempts: 10,
		IsTrusted:      false,
		IsBlocked:      true,
	}

	assert.True(t, signal.IsBlocked)
	assert.False(t, signal.IsTrusted)
	assert.True(t, signal.FailedAttempts >= 10)
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkGenerateChallengeID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateChallengeID()
	}
}

func BenchmarkGenerateTrustToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateTrustToken()
	}
}

func BenchmarkHashTrustToken(b *testing.B) {
	token := "tt_benchmark_token_1234567890"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hashTrustToken(token)
	}
}
