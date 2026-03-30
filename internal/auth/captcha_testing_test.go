package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

func TestMockCaptchaProvider_Verify(t *testing.T) {
	t.Run("returns success by default", func(t *testing.T) {
		provider := &MockCaptchaProvider{}
		ctx := context.Background()

		result, err := provider.Verify(ctx, "any-token", "192.168.1.1")
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("uses custom verify function", func(t *testing.T) {
		provider := &MockCaptchaProvider{
			VerifyFunc: func(ctx context.Context, token string, remoteIP string) (*CaptchaResult, error) {
				if token == "valid-token" {
					return &CaptchaResult{Success: true}, nil
				}
				return &CaptchaResult{Success: false, ErrorCode: "invalid"}, nil
			},
		}

		ctx := context.Background()

		// Valid token
		result, err := provider.Verify(ctx, "valid-token", "192.168.1.1")
		require.NoError(t, err)
		assert.True(t, result.Success)

		// Invalid token
		result, err = provider.Verify(ctx, "invalid-token", "192.168.1.1")
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Equal(t, "invalid", result.ErrorCode)
	})
}

func TestMockCaptchaProvider_Name(t *testing.T) {
	t.Run("returns mock by default", func(t *testing.T) {
		provider := &MockCaptchaProvider{}
		assert.Equal(t, "mock", provider.Name())
	})

	t.Run("returns custom name", func(t *testing.T) {
		provider := &MockCaptchaProvider{NameValue: "custom-mock"}
		assert.Equal(t, "custom-mock", provider.Name())
	})
}

func TestNewMockCaptchaService(t *testing.T) {
	t.Run("creates disabled service with nil config", func(t *testing.T) {
		service := NewMockCaptchaService(nil, nil)
		assert.NotNil(t, service)
		assert.Nil(t, service.config)
	})

	t.Run("creates disabled service with disabled config", func(t *testing.T) {
		cfg := &config.CaptchaConfig{
			Enabled: false,
		}
		service := NewMockCaptchaService(cfg, nil)
		assert.NotNil(t, service)
	})

	t.Run("creates enabled service with provider", func(t *testing.T) {
		cfg := &config.CaptchaConfig{
			Enabled:   true,
			Provider:  "mock",
			SiteKey:   "site-key",
			SecretKey: "secret-key",
			Endpoints: []string{"/auth/signup", "/auth/login"},
		}
		provider := &MockCaptchaProvider{}

		service := NewMockCaptchaService(cfg, provider)
		assert.NotNil(t, service)
		assert.NotNil(t, service.provider)
		assert.True(t, service.enabledEndpoints["/auth/signup"])
		assert.True(t, service.enabledEndpoints["/auth/login"])
	})
}

func TestNewTestCaptchaService(t *testing.T) {
	t.Run("creates service with specified endpoints and valid token", func(t *testing.T) {
		endpoints := []string{"/auth/signup", "/auth/login"}
		validToken := "test-valid-token"

		service := NewTestCaptchaService(endpoints, validToken)
		assert.NotNil(t, service)
		assert.True(t, service.config.Enabled)
	})

	t.Run("verifies valid token successfully", func(t *testing.T) {
		service := NewTestCaptchaService([]string{"/auth/signup"}, "my-token")
		ctx := context.Background()

		result, err := service.provider.Verify(ctx, "my-token", "192.168.1.1")
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		service := NewTestCaptchaService([]string{"/auth/signup"}, "my-token")
		ctx := context.Background()

		result, err := service.provider.Verify(ctx, "wrong-token", "192.168.1.1")
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Equal(t, "invalid-token", result.ErrorCode)
	})
}

func TestNewDisabledCaptchaService(t *testing.T) {
	t.Run("creates disabled service", func(t *testing.T) {
		service := NewDisabledCaptchaService()
		assert.NotNil(t, service)
		assert.NotNil(t, service.config)
		assert.False(t, service.config.Enabled)
	})
}

func TestNewTestAuthServiceWithSettings(t *testing.T) {
	t.Run("creates service with signup enabled", func(t *testing.T) {
		service := NewTestAuthServiceWithSettings(true, true)
		assert.NotNil(t, service)
		assert.NotNil(t, service.config)
		assert.True(t, service.config.SignupEnabled)
	})

	t.Run("creates service with signup disabled", func(t *testing.T) {
		service := NewTestAuthServiceWithSettings(false, true)
		assert.NotNil(t, service)
		assert.False(t, service.config.SignupEnabled)
	})

	t.Run("settings cache is pre-populated", func(t *testing.T) {
		service := NewTestAuthServiceWithSettings(true, false)
		ctx := context.Background()

		// Check signup_enabled is cached
		signupEnabled := service.settingsCache.GetBool(ctx, "app.auth.signup_enabled", false)
		assert.True(t, signupEnabled)

		// Check password_login is disabled
		passwordDisabled := service.settingsCache.GetBool(ctx, "app.auth.disable_app_password_login", false)
		assert.True(t, passwordDisabled) // We passed false for passwordLoginEnabled
	})

	t.Run("password hasher has minimal requirements", func(t *testing.T) {
		service := NewTestAuthServiceWithSettings(true, true)
		assert.NotNil(t, service.passwordHasher)

		// Should accept any password due to minimal requirements
		err := service.passwordHasher.ValidatePassword("a")
		assert.NoError(t, err)
	})
}
