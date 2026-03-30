package middleware

import (
	"time"
)

// KeyStrategy defines how rate limit keys are generated.
type KeyStrategy string

const (
	// KeyStrategyIP generates keys based on client IP address.
	KeyStrategyIP KeyStrategy = "ip"
	// KeyStrategyUser generates keys based on authenticated user ID, falling back to IP.
	KeyStrategyUser KeyStrategy = "user"
	// KeyStrategyClientKey generates keys based on client key ID, falling back to IP.
	KeyStrategyClientKey KeyStrategy = "client_key"
	// KeyStrategyServiceKey generates keys based on service key ID, falling back to IP.
	KeyStrategyServiceKey KeyStrategy = "service_key"
	// KeyStrategyToken generates keys based on token prefix, falling back to IP.
	KeyStrategyToken KeyStrategy = "token"
	// KeyStrategyEmail generates keys based on email from request body, falling back to IP.
	KeyStrategyEmail KeyStrategy = "email"
	// KeyStrategyTiered uses different limits for anonymous, authenticated, and client key users.
	KeyStrategyTiered KeyStrategy = "tiered"
)

// RateLimitDefinition declares a rate limiter's configuration.
// It provides the default values and metadata for a rate limiter.
type RateLimitDefinition struct {
	// Name is the identifier for the rate limiter (used in metrics and logs).
	Name string
	// DefaultMax is the default maximum number of requests allowed.
	DefaultMax int
	// DefaultWindow is the default time window for the rate limit.
	DefaultWindow time.Duration
	// KeyStrategy defines how the rate limit key is generated.
	KeyStrategy KeyStrategy
	// KeyPrefix is prepended to the generated key for namespacing.
	KeyPrefix string
	// Message is the error message returned when rate limit is exceeded.
	// If empty, a default message is generated.
	Message string
	// ConfigMaxField is the config field name for max requests override (e.g., "AuthLoginRateLimit").
	ConfigMaxField string
	// ConfigWindowField is the config field name for window override (e.g., "AuthLoginRateWindow").
	ConfigWindowField string
}

// Registry contains all rate limiter definitions.
// This is the single source of truth for rate limit configurations.
var Registry = map[string]RateLimitDefinition{
	// Authentication rate limiters
	"auth_login": {
		Name:              "auth_login",
		DefaultMax:        10,
		DefaultWindow:     15 * time.Minute,
		KeyStrategy:       KeyStrategyIP,
		KeyPrefix:         "login",
		ConfigMaxField:    "AuthLoginRateLimit",
		ConfigWindowField: "AuthLoginRateWindow",
	},
	"auth_signup": {
		Name:              "auth_signup",
		DefaultMax:        10,
		DefaultWindow:     15 * time.Minute,
		KeyStrategy:       KeyStrategyIP,
		KeyPrefix:         "signup",
		ConfigMaxField:    "AuthSignupRateLimit",
		ConfigWindowField: "AuthSignupRateWindow",
	},
	"auth_password_reset": {
		Name:              "auth_password_reset",
		DefaultMax:        5,
		DefaultWindow:     15 * time.Minute,
		KeyStrategy:       KeyStrategyIP,
		KeyPrefix:         "password_reset",
		ConfigMaxField:    "AuthPasswordResetRateLimit",
		ConfigWindowField: "AuthPasswordResetRateWindow",
	},
	"auth_2fa": {
		Name:              "auth_2fa",
		DefaultMax:        5,
		DefaultWindow:     5 * time.Minute,
		KeyStrategy:       KeyStrategyIP,
		KeyPrefix:         "2fa",
		ConfigMaxField:    "Auth2FARateLimit",
		ConfigWindowField: "Auth2FARateWindow",
	},
	"auth_refresh": {
		Name:              "auth_refresh",
		DefaultMax:        10,
		DefaultWindow:     1 * time.Minute,
		KeyStrategy:       KeyStrategyToken,
		KeyPrefix:         "refresh",
		ConfigMaxField:    "AuthRefreshRateLimit",
		ConfigWindowField: "AuthRefreshRateWindow",
	},
	"auth_magic_link": {
		Name:              "auth_magic_link",
		DefaultMax:        5,
		DefaultWindow:     15 * time.Minute,
		KeyStrategy:       KeyStrategyIP,
		KeyPrefix:         "magiclink",
		ConfigMaxField:    "AuthMagicLinkRateLimit",
		ConfigWindowField: "AuthMagicLinkRateWindow",
	},

	// Admin rate limiters
	"admin_setup": {
		Name:          "admin_setup",
		DefaultMax:    5,
		DefaultWindow: 15 * time.Minute,
		KeyStrategy:   KeyStrategyIP,
		KeyPrefix:     "admin_setup",
	},
	"admin_login": {
		Name:          "admin_login",
		DefaultMax:    4,
		DefaultWindow: 1 * time.Minute,
		KeyStrategy:   KeyStrategyIP,
		KeyPrefix:     "admin_login",
	},

	// API rate limiters
	"global_api": {
		Name:          "global",
		DefaultMax:    100,
		DefaultWindow: 1 * time.Minute,
		KeyStrategy:   KeyStrategyUser,
		KeyPrefix:     "global",
	},
	"authenticated_user": {
		Name:          "authenticated_user",
		DefaultMax:    500,
		DefaultWindow: 1 * time.Minute,
		KeyStrategy:   KeyStrategyUser,
		KeyPrefix:     "user",
	},
	"client_key": {
		Name:          "client_key",
		DefaultMax:    1000,
		DefaultWindow: 1 * time.Minute,
		KeyStrategy:   KeyStrategyClientKey,
		KeyPrefix:     "clientkey",
	},

	// Migration API rate limiter
	"migration_api": {
		Name:          "migration_api",
		DefaultMax:    10,
		DefaultWindow: 1 * time.Hour,
		KeyStrategy:   KeyStrategyServiceKey,
		KeyPrefix:     "migration",
	},

	// Storage rate limiter
	"storage_upload": {
		Name:          "storage_upload",
		DefaultMax:    60,
		DefaultWindow: 1 * time.Minute,
		KeyStrategy:   KeyStrategyUser,
		KeyPrefix:     "storage_upload",
	},

	// Webhook rate limiters
	"github_webhook": {
		Name:          "github_webhook",
		DefaultMax:    30,
		DefaultWindow: 1 * time.Minute,
		KeyStrategy:   KeyStrategyIP,
		KeyPrefix:     "github_webhook",
	},
}

// GetDefinition returns the rate limit definition for the given name.
// Returns false if the definition doesn't exist.
func GetDefinition(name string) (RateLimitDefinition, bool) {
	def, ok := Registry[name]
	return def, ok
}

// ListDefinitions returns all registered rate limit definitions.
func ListDefinitions() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	return names
}
