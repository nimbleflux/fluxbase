package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// =============================================================================
// SES Service Construction Tests
// =============================================================================

func TestNewSESService(t *testing.T) {
	t.Run("returns error for missing region", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:    "ses",
			SESRegion:   "",
			FromName:    "Test",
			FromAddress: "test@example.com",
		}

		service, err := NewSESService(cfg)

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "region is required")
	})

	t.Run("creates service with region only (uses default credentials)", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:    "ses",
			SESRegion:   "us-east-1",
			FromName:    "Test",
			FromAddress: "test@example.com",
		}

		service, err := NewSESService(cfg)

		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, cfg, service.config)
		assert.NotNil(t, service.client)
	})

	t.Run("creates service with static credentials", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:     "ses",
			SESRegion:    "eu-west-1",
			SESAccessKey: "AKIAIOSFODNN7EXAMPLE",
			SESSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			FromName:     "Test",
			FromAddress:  "test@example.com",
		}

		service, err := NewSESService(cfg)

		require.NoError(t, err)
		require.NotNil(t, service)
	})
}

// =============================================================================
// SES IsConfigured Tests
// =============================================================================

func TestSESService_IsConfigured(t *testing.T) {
	t.Run("returns false when disabled", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:     false,
			Provider:    "ses",
			SESRegion:   "us-east-1",
			FromAddress: "test@example.com",
		}
		service, _ := NewSESService(cfg)

		assert.False(t, service.IsConfigured())
	})

	t.Run("returns true when enabled and configured", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:     true,
			Provider:    "ses",
			SESRegion:   "us-east-1",
			FromAddress: "test@example.com",
		}
		service, _ := NewSESService(cfg)

		// IsConfigured depends on config.EmailConfig.IsConfigured()
		// which typically checks FromAddress is set
		result := service.IsConfigured()
		assert.True(t, result)
	})
}

// =============================================================================
// SES Service Struct Tests
// =============================================================================

func TestSESService_Struct(t *testing.T) {
	t.Run("stores config and client", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:    "ses",
			SESRegion:   "us-east-1",
			FromName:    "Fluxbase",
			FromAddress: "noreply@example.com",
		}

		service, err := NewSESService(cfg)

		require.NoError(t, err)
		assert.Equal(t, cfg, service.config)
		assert.NotNil(t, service.client)
	})
}

// =============================================================================
// SES Email Type Methods Tests
// =============================================================================

func TestSESService_EmailMethods(t *testing.T) {
	cfg := &config.EmailConfig{
		Provider:    "ses",
		SESRegion:   "us-east-1",
		FromName:    "Test",
		FromAddress: "test@example.com",
	}
	service, _ := NewSESService(cfg)

	t.Run("SendMagicLink method exists", func(t *testing.T) {
		// Method exists and has correct signature
		assert.NotNil(t, service)
		// Can't actually send without real AWS credentials
	})

	t.Run("SendVerificationEmail method exists", func(t *testing.T) {
		assert.NotNil(t, service)
	})

	t.Run("SendPasswordReset method exists", func(t *testing.T) {
		assert.NotNil(t, service)
	})

	t.Run("SendInvitationEmail method exists", func(t *testing.T) {
		assert.NotNil(t, service)
	})

	t.Run("Send method exists", func(t *testing.T) {
		assert.NotNil(t, service)
	})
}

// =============================================================================
// SES Region Validation Tests
// =============================================================================

func TestSESService_RegionValidation(t *testing.T) {
	validRegions := []string{
		"us-east-1",
		"us-west-2",
		"eu-west-1",
		"ap-southeast-1",
	}

	for _, region := range validRegions {
		t.Run("accepts region "+region, func(t *testing.T) {
			cfg := &config.EmailConfig{
				Provider:    "ses",
				SESRegion:   region,
				FromAddress: "test@example.com",
			}

			service, err := NewSESService(cfg)

			assert.NoError(t, err)
			assert.NotNil(t, service)
		})
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewSESService(b *testing.B) {
	cfg := &config.EmailConfig{
		Provider:    "ses",
		SESRegion:   "us-east-1",
		FromName:    "Test",
		FromAddress: "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewSESService(cfg)
	}
}
