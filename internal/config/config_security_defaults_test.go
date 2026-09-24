package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestSecurityRateLimitDefaults verifies that every auth-endpoint rate limiter
// has a viper default matching docs/guides/rate-limiting.md. Without these,
// limiter constructors receive Max=0 / Expiration=0 when the YAML omits the
// keys, producing broken (or missing) rate limiting.
func TestSecurityRateLimitDefaults(t *testing.T) {
	// Re-run setDefaults to ensure registration (idempotent).
	setDefaults()

	cases := []struct {
		key        string
		wantLimit  int
		wantWindow time.Duration
	}{
		{"security.auth_login_rate_limit", 10, time.Minute},
		{"security.auth_signup_rate_limit", 10, 15 * time.Minute},
		{"security.auth_password_reset_rate_limit", 5, 15 * time.Minute},
		{"security.auth_2fa_rate_limit", 5, 5 * time.Minute},
		{"security.auth_refresh_rate_limit", 10, time.Minute},
		{"security.auth_magic_link_rate_limit", 5, 15 * time.Minute},
		{"security.admin_setup_rate_limit", 5, 15 * time.Minute},
		{"security.admin_login_rate_limit", 10, time.Minute},
		{"security.dashboard_login_rate_limit", 60, time.Minute},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			assert.True(t, viper.IsSet(tc.key), "missing viper default for %s", tc.key)
			assert.Equal(t, tc.wantLimit, viper.GetInt(tc.key))

			windowKey := tc.key[:len(tc.key)-len("_limit")] + "_window"
			assert.True(t, viper.IsSet(windowKey), "missing viper default for %s", windowKey)
			assert.Equal(t, tc.wantWindow, viper.GetDuration(windowKey))
		})
	}
}

// TestSecurityRateLimitFieldsUnmarshal verifies the SecurityConfig mapstructure
// tags line up with the default keys so configured values actually reach the
// limiter constructors (the app unmarshals the whole Config from viper).
func TestSecurityRateLimitFieldsUnmarshal(t *testing.T) {
	setDefaults()

	viper.Set("security.auth_signup_rate_limit", 42)
	viper.Set("security.auth_signup_rate_window", "7m")
	t.Cleanup(func() {
		viper.Set("security.auth_signup_rate_limit", 10)
		viper.Set("security.auth_signup_rate_window", "15m")
	})

	var cfg Config
	err := viper.Unmarshal(&cfg)
	assert.NoError(t, err)
	assert.Equal(t, 42, cfg.Security.AuthSignupRateLimit)
	assert.Equal(t, 7*time.Minute, cfg.Security.AuthSignupRateWindow)
}
