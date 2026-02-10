package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/auth"
	"github.com/fluxbase-eu/fluxbase/internal/observability"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/storage/memory/v2"
	"github.com/rs/zerolog/log"
)

var (
	rateLimiterMetrics          *observability.Metrics
	rateLimiterWarningDisplayed bool
	rateLimiterWarningMu        sync.Once
)

// SetRateLimiterMetrics sets the metrics instance for rate limiter
func SetRateLimiterMetrics(m *observability.Metrics) {
	rateLimiterMetrics = m
}

// logRateLimiterWarning logs a warning about in-memory rate limiting in multi-instance environments.
// The warning is only logged once per process to avoid log spam.
func logRateLimiterWarning() {
	rateLimiterWarningMu.Do(func() {
		// Check for indicators of multi-instance deployment
		isKubernetes := os.Getenv("KUBERNETES_SERVICE_HOST") != ""
		isPodName := os.Getenv("POD_NAME") != "" || os.Getenv("HOSTNAME") != ""
		isDockerCompose := os.Getenv("COMPOSE_PROJECT_NAME") != ""
		hasRedisURL := os.Getenv("FLUXBASE_REDIS_URL") != "" || os.Getenv("REDIS_URL") != ""
		hasDragonflyURL := os.Getenv("FLUXBASE_DRAGONFLY_URL") != "" || os.Getenv("DRAGONFLY_URL") != ""

		// If Redis/Dragonfly is configured, rate limiting can be distributed
		if hasRedisURL || hasDragonflyURL {
			return // Distributed rate limiting is likely configured
		}

		// Log warning if we detect multi-instance environment indicators
		if isKubernetes || isPodName || isDockerCompose {
			log.Warn().
				Bool("kubernetes_detected", isKubernetes).
				Bool("container_detected", isPodName).
				Bool("compose_detected", isDockerCompose).
				Msg("SECURITY WARNING: Using in-memory rate limiting in a multi-instance environment. " +
					"Rate limits are per-instance only and can be bypassed by targeting different instances. " +
					"For production, configure Redis/Dragonfly (FLUXBASE_REDIS_URL or FLUXBASE_DRAGONFLY_URL) " +
					"for distributed rate limiting, or use a reverse proxy with centralized rate limiting.")
			rateLimiterWarningDisplayed = true
		}
	})
}

// IsRateLimiterWarningDisplayed returns true if the rate limiter warning was displayed
func IsRateLimiterWarningDisplayed() bool {
	return rateLimiterWarningDisplayed
}

// extractRoleFromToken attempts to extract the role claim from a JWT token
// without performing full signature validation. This is used for rate limiting
// exemption checks only. The token will be fully validated later by auth middleware.
func extractRoleFromToken(token string) string {
	// JWT format: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		log.Debug().Int("parts", len(parts)).Msg("Rate limiter: token is not a valid JWT (wrong number of parts)")
		return ""
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// Try standard base64 encoding
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			log.Debug().Err(err).Msg("Rate limiter: failed to decode JWT payload")
			return ""
		}
	}

	// Parse JSON to extract role
	var claims struct {
		Role string `json:"role"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		log.Debug().Err(err).Msg("Rate limiter: failed to parse JWT payload JSON")
		return ""
	}

	log.Debug().Str("role", claims.Role).Msg("Rate limiter: extracted role from JWT")
	return claims.Role
}

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	Name       string                 // Name of the rate limiter (for metrics)
	Max        int                    // Maximum number of requests
	Expiration time.Duration          // Time window for the rate limit
	KeyFunc    func(fiber.Ctx) string // Function to generate the key for rate limiting
	Message    string                 // Custom error message
}

// NewRateLimiter creates a new rate limiter middleware with custom configuration.
//
// IMPORTANT: This middleware uses Fiber's native in-memory storage for rate limiting.
// Rate limiting is per-instance only. In multi-instance deployments, each instance
// maintains its own rate limit counters independently.
//
// SECURITY WARNING: In-memory rate limiting is per-instance only. In multi-instance deployments,
// attackers can bypass rate limits by targeting different instances. For production environments
// with horizontal scaling, consider using a reverse proxy (nginx, Traefik) with centralized
// rate limiting, or implement custom middleware with Redis-backed storage.
// See docs/deployment/production-checklist.md for details.
func NewRateLimiter(config RateLimiterConfig) fiber.Handler {
	// Log warning about in-memory rate limiting in multi-instance environments
	logRateLimiterWarning()

	// Always use Fiber's native memory storage for compatibility with Fiber's limiter.
	// The limiter middleware uses MessagePack encoding internally, which is incompatible
	// with our custom IncrementAdapter's binary encoding.
	storage := memory.New(memory.Config{
		GCInterval: 10 * time.Minute,
	})

	// Default key function uses IP address
	if config.KeyFunc == nil {
		config.KeyFunc = func(c fiber.Ctx) string {
			return c.IP()
		}
	}

	// Default error message
	if config.Message == "" {
		config.Message = fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %s allowed.",
			config.Max, config.Expiration.String())
	}

	// Capture name for closure
	limiterName := config.Name
	if limiterName == "" {
		limiterName = "default"
	}

	return limiter.New(limiter.Config{
		Max:          config.Max,
		Expiration:   config.Expiration,
		KeyGenerator: config.KeyFunc,
		LimitReached: func(c fiber.Ctx) error {
			// Record rate limit hit metric
			if rateLimiterMetrics != nil {
				rateLimiterMetrics.RecordRateLimitHit(limiterName, c.IP())
			}

			retryAfter := int(config.Expiration.Seconds())
			c.Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"code":        "RATE_LIMIT_EXCEEDED",
				"error":       "Rate limit exceeded",
				"message":     config.Message,
				"retry_after": retryAfter,
			})
		},
		Storage: storage,
	})
}

// AuthLoginLimiter limits login attempts per IP
func AuthLoginLimiter() fiber.Handler {
	return AuthLoginLimiterWithConfig(10, 15*time.Minute)
}

// AuthLoginLimiterWithConfig creates an auth login rate limiter with custom limits
func AuthLoginLimiterWithConfig(max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Name:       "auth_login",
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			return "login:" + c.IP()
		},
		Message: fmt.Sprintf("Too many login attempts. Please try again in %d minutes.", int(expiration.Minutes())),
	})
}

// AuthSignupLimiter limits signup attempts per IP
func AuthSignupLimiter() fiber.Handler {
	return AuthSignupLimiterWithConfig(10, 15*time.Minute)
}

// AuthSignupLimiterWithConfig creates an auth signup rate limiter with custom limits
func AuthSignupLimiterWithConfig(max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Name:       "auth_signup",
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			return "signup:" + c.IP()
		},
		Message: fmt.Sprintf("Too many signup attempts. Please try again in %d minutes.", int(expiration.Minutes())),
	})
}

// AuthPasswordResetLimiter limits password reset requests per IP
func AuthPasswordResetLimiter() fiber.Handler {
	return AuthPasswordResetLimiterWithConfig(5, 15*time.Minute)
}

// AuthPasswordResetLimiterWithConfig creates an auth password reset rate limiter with custom limits
func AuthPasswordResetLimiterWithConfig(max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Name:       "auth_password_reset",
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			return "password_reset:" + c.IP()
		},
		Message: fmt.Sprintf("Too many password reset requests. Please try again in %d minutes.", int(expiration.Minutes())),
	})
}

// Auth2FALimiter limits 2FA verification attempts per IP
// Strict rate limiting to prevent brute-force attacks on 6-digit TOTP codes
func Auth2FALimiter() fiber.Handler {
	return Auth2FALimiterWithConfig(5, 5*time.Minute)
}

// Auth2FALimiterWithConfig creates an auth 2FA rate limiter with custom limits
func Auth2FALimiterWithConfig(max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Name:       "auth_2fa",
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			return "2fa:" + c.IP()
		},
		Message: fmt.Sprintf("Too many 2FA verification attempts. Please try again in %d minutes.", int(expiration.Minutes())),
	})
}

// AuthRefreshLimiter limits token refresh attempts per token
func AuthRefreshLimiter() fiber.Handler {
	return AuthRefreshLimiterWithConfig(10, 1*time.Minute)
}

// AuthRefreshLimiterWithConfig creates an auth token refresh rate limiter with custom limits
func AuthRefreshLimiterWithConfig(max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Name:       "auth_refresh",
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			// Try to get token from request body
			var req struct {
				RefreshToken string `json:"refresh_token"`
			}
			if err := c.Bind().Body(&req); err == nil && req.RefreshToken != "" {
				return "refresh:" + req.RefreshToken[:20] // Use first 20 chars as key
			}
			// Fallback to IP if no token found
			return "refresh:" + c.IP()
		},
		Message: fmt.Sprintf("Too many token refresh attempts. Please wait %d minute(s).", int(expiration.Minutes())),
	})
}

// AuthMagicLinkLimiter limits magic link requests per IP
func AuthMagicLinkLimiter() fiber.Handler {
	return AuthMagicLinkLimiterWithConfig(5, 15*time.Minute)
}

// AuthMagicLinkLimiterWithConfig creates an auth magic link rate limiter with custom limits
func AuthMagicLinkLimiterWithConfig(max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Name:       "auth_magic_link",
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			return "magiclink:" + c.IP()
		},
		Message: fmt.Sprintf("Too many magic link requests. Please try again in %d minutes.", int(expiration.Minutes())),
	})
}

// AuthEmailBasedLimiter limits requests per email address (for sensitive operations)
func AuthEmailBasedLimiter(prefix string, max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			var req struct {
				Email string `json:"email"`
			}
			if err := c.Bind().Body(&req); err == nil && req.Email != "" {
				return prefix + ":" + req.Email
			}
			// Fallback to IP if no email found
			return prefix + ":" + c.IP()
		},
		Message: "Too many requests. Please try again later.",
	})
}

// GlobalAPILimiter is a general rate limiter for all API endpoints
// Uses per-IP rate limiting by default, can use per-user rate limiting if enabled
func GlobalAPILimiter() fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Name:       "global",
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyFunc: func(c fiber.Ctx) string {
			// Try to get user ID from locals (set by auth middleware)
			userID := c.Locals("user_id")
			if userID != nil {
				if uid, ok := userID.(string); ok && uid != "" && uid != "anonymous" {
					return "global_user:" + uid
				}
			}
			// Fallback to IP for anonymous users or when user ID not available
			return "global_ip:" + c.IP()
		},
		Message: "API rate limit exceeded. Maximum 100 requests per minute allowed.",
	})
}

// DynamicGlobalAPILimiter creates a rate limiter that respects the dynamic setting
// It checks the settings cache on each request, allowing real-time toggling of rate limiting
// without server restart
// Admin users (admin, dashboard_admin) are exempt from rate limiting
// service_role users can be rate-limited if service_role_rate_limit > 0
func DynamicGlobalAPILimiter(settingsCache *auth.SettingsCache) fiber.Handler {
	// Create the actual rate limiter once
	rateLimiter := GlobalAPILimiter()

	return func(c fiber.Ctx) error {
		// Only apply rate limiting to API endpoints
		// Skip for static files, health checks, admin UI, favicon, etc.
		if !strings.HasPrefix(c.Path(), "/api/") {
			return c.Next()
		}

		// First check if role is already set by auth middleware
		role := c.Locals("user_role")
		if role == "admin" || role == "dashboard_admin" {
			log.Debug().
				Str("role", role.(string)).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Rate limiter: bypassing for admin user (role already set)")
			return c.Next()
		}

		// If role is not set, try to extract it from JWT token
		// This handles the case where global rate limiting runs before auth middleware
		tokenSource := "none"
		token := c.Cookies("fluxbase_access_token")
		if token == "" {
			// Try Authorization header
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				if strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
					tokenSource = "header"
				} else {
					token = authHeader
					tokenSource = "header"
				}
			}
		} else {
			tokenSource = "cookie"
		}

		// If we have a token, try to extract the role claim without full validation
		// This is a lightweight check just for rate limiting exemption
		if token != "" {
			// Parse JWT token to extract role claim
			// We use a simplified parsing that doesn't validate signatures
			// since the auth middleware will do full validation later
			extractedRole := extractRoleFromToken(token)
			log.Debug().
				Str("path", c.Path()).
				Str("method", c.Method()).
				Str("token_source", tokenSource).
				Str("extracted_role", extractedRole).
				Msg("Rate limiter: checked token for role")

			if extractedRole == "admin" || extractedRole == "dashboard_admin" {
				log.Debug().
					Str("role", extractedRole).
					Str("path", c.Path()).
					Str("method", c.Method()).
					Msg("Rate limiter: bypassing for admin user (extracted from token)")
				return c.Next()
			}
			// For service_role, check if rate limiting is configured
			if extractedRole == "service_role" {
				// Check if service_role rate limiting is enabled
				ctx := c.RequestCtx()
				serviceRoleRateLimit := settingsCache.GetInt(ctx, "app.security.service_role_rate_limit", 0)
				if serviceRoleRateLimit <= 0 {
					// No rate limiting for service_role (default)
					log.Debug().Msg("Rate limiter: bypassing for service_role (no rate limit configured)")
					return c.Next()
				}
				// Apply service_role rate limiting
				rateWindow := settingsCache.GetDuration(ctx, "app.security.service_role_rate_window", 1*time.Minute)
				serviceRoleLimiter := NewRateLimiter(RateLimiterConfig{
					Name:       "service_role",
					Max:        serviceRoleRateLimit,
					Expiration: rateWindow,
					KeyFunc: func(c fiber.Ctx) string {
						return "service_role:" + c.IP()
					},
					Message: fmt.Sprintf("Service role rate limit exceeded. Maximum %d requests per %s allowed.", serviceRoleRateLimit, rateWindow.String()),
				})
				log.Debug().
					Int("limit", serviceRoleRateLimit).
					Str("window", rateWindow.String()).
					Msg("Rate limiter: applying service_role rate limit")
				return serviceRoleLimiter(c)
			}
		}

		// Check if rate limiting is enabled via settings cache
		ctx := c.RequestCtx()
		isEnabled := settingsCache.GetBool(ctx, "app.security.enable_global_rate_limit", false)

		if !isEnabled {
			log.Debug().Msg("Rate limiter: disabled via settings, skipping")
			return c.Next() // Skip rate limiting
		}

		log.Debug().
			Str("path", c.Path()).
			Str("method", c.Method()).
			Str("ip", c.IP()).
			Msg("Rate limiter: applying global rate limit")
		return rateLimiter(c)
	}
}

// AuthenticatedUserLimiter limits requests per authenticated user (higher limits than IP-based)
// Should be applied AFTER authentication middleware
func AuthenticatedUserLimiter() fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Max:        500, // Higher limit for authenticated users
		Expiration: 1 * time.Minute,
		KeyFunc: func(c fiber.Ctx) string {
			// Try to get user ID from locals (set by auth middleware)
			userID := c.Locals("user_id")
			if userID != nil {
				if uid, ok := userID.(string); ok && uid != "" {
					return "user:" + uid
				}
			}
			// Fallback to IP if no user ID (shouldn't happen if auth middleware ran)
			return "user:" + c.IP()
		},
		Message: "Rate limit exceeded for your account. Maximum 500 requests per minute allowed.",
	})
}

// ClientKeyLimiter limits requests per client key with configurable limits
// Should be applied AFTER client key authentication middleware
func ClientKeyLimiter(maxRequests int, duration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Max:        maxRequests,
		Expiration: duration,
		KeyFunc: func(c fiber.Ctx) string {
			// Try to get client key ID from locals (set by client key auth middleware)
			keyID := c.Locals("client_key_id")
			if keyID != nil {
				if kid, ok := keyID.(string); ok && kid != "" {
					return "clientkey:" + kid
				}
			}
			// Fallback to IP if no client key ID
			return "clientkey:" + c.IP()
		},
		Message: fmt.Sprintf("Client key rate limit exceeded. Maximum %d requests per %s allowed.", maxRequests, duration.String()),
	})
}

// DefaultClientKeyLimiter returns a client key limiter with default limits (1000 req/min)
func DefaultClientKeyLimiter() fiber.Handler {
	return ClientKeyLimiter(1000, 1*time.Minute)
}

// PerUserOrIPLimiter implements tiered rate limiting:
// - Authenticated users: higher limit
// - Client keys: configurable limit
// - Anonymous (IP): lower limit
func PerUserOrIPLimiter(anonMax, userMax, clientKeyMax int, duration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Max:        anonMax, // Base max (will be adjusted by key function)
		Expiration: duration,
		KeyFunc: func(c fiber.Ctx) string {
			// Priority 1: Check for client key
			clientKeyID := c.Locals("client_key_id")
			if clientKeyID != nil {
				if kid, ok := clientKeyID.(string); ok && kid != "" {
					// Use client key specific limit
					return fmt.Sprintf("clientkey:%s:%d", kid, clientKeyMax)
				}
			}

			// Priority 2: Check for authenticated user
			userID := c.Locals("user_id")
			if userID != nil {
				if uid, ok := userID.(string); ok && uid != "" {
					// Use user specific limit
					return fmt.Sprintf("user:%s:%d", uid, userMax)
				}
			}

			// Priority 3: Fallback to IP (anonymous)
			return fmt.Sprintf("ip:%s:%d", c.IP(), anonMax)
		},
		Message: "Rate limit exceeded. Please try again later.",
	})
}

// AdminSetupLimiter limits admin setup attempts per IP
// Very strict since this is a one-time operation
func AdminSetupLimiter() fiber.Handler {
	return AdminSetupLimiterWithConfig(5, 15*time.Minute)
}

// AdminSetupLimiterWithConfig creates an admin setup rate limiter with custom limits
func AdminSetupLimiterWithConfig(max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			return "admin_setup:" + c.IP()
		},
		Message: fmt.Sprintf("Too many admin setup attempts. Please try again in %d minutes.", int(expiration.Minutes())),
	})
}

// AdminLoginLimiter limits admin login attempts per IP
// Max is set to 4 to trigger rate limiting before account lockout (which happens at 5 failed attempts)
func AdminLoginLimiter() fiber.Handler {
	return AdminLoginLimiterWithConfig(4, 1*time.Minute)
}

// AdminLoginLimiterWithConfig creates an admin login rate limiter with custom limits
func AdminLoginLimiterWithConfig(max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			return "admin_login:" + c.IP()
		},
		Message: fmt.Sprintf("Too many admin login attempts. Please try again in %d minutes.", int(expiration.Minutes())),
	})
}

// GitHubWebhookLimiter limits GitHub webhook requests per IP and repository
// Prevents abuse of the webhook endpoint for branch creation/deletion
func GitHubWebhookLimiter() fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Max:        30,              // 30 requests
		Expiration: 1 * time.Minute, // per minute per IP
		KeyFunc: func(c fiber.Ctx) string {
			return "github_webhook:" + c.IP()
		},
		Message: "GitHub webhook rate limit exceeded. Maximum 30 requests per minute allowed.",
	})
}

// MigrationAPILimiter limits migrations API requests per service key
// Very strict rate limiting due to powerful DDL operations
// Should be applied AFTER service key authentication middleware
// NOTE: service_role JWT tokens bypass rate limiting entirely (trusted keys)
// Service keys (sk_*) use per-key configurable rate limits from the database
// Deprecated: Use MigrationAPILimiterWithConfig for H-2 security fix
func MigrationAPILimiter() fiber.Handler {
	return MigrationAPILimiterWithConfig(0, 0)
}

// MigrationAPILimiterWithConfig creates a migrations API rate limiter with custom limits
// H-2: Enforces rate limiting for service_role tokens when configured
// serviceRoleRateLimit: Max requests for service_role tokens (0 = unlimited, for backward compatibility)
// serviceRoleRateWindow: Time window for service_role rate limiting
func MigrationAPILimiterWithConfig(serviceRoleRateLimit int, serviceRoleRateWindow time.Duration) fiber.Handler {
	// Default rate limiter for service keys without custom limits
	defaultRateLimiter := NewRateLimiter(RateLimiterConfig{
		Max:        10,            // 10 requests
		Expiration: 1 * time.Hour, // per hour
		KeyFunc: func(c fiber.Ctx) string {
			keyID := c.Locals("service_key_id")
			if keyID != nil {
				if kid, ok := keyID.(string); ok && kid != "" {
					return "migration_key:" + kid
				}
			}
			return "migration_ip:" + c.IP()
		},
		Message: "Migrations API rate limit exceeded. Maximum 10 requests per hour allowed.",
	})

	// H-2: Service role rate limiter (if configured)
	var serviceRoleLimiter fiber.Handler
	if serviceRoleRateLimit > 0 && serviceRoleRateWindow > 0 {
		serviceRoleLimiter = NewRateLimiter(RateLimiterConfig{
			Max:        serviceRoleRateLimit,
			Expiration: serviceRoleRateWindow,
			KeyFunc: func(c fiber.Ctx) string {
				// Rate limit by JWT ID (jti) for service_role tokens
				if jti := c.Locals("jti"); jti != nil {
					if jtiStr, ok := jti.(string); ok && jtiStr != "" {
						return "service_role:" + jtiStr
					}
				}
				// Fallback to service key ID
				if keyID := c.Locals("service_key_id"); keyID != nil {
					if kid, ok := keyID.(string); ok && kid != "" {
						return "service_role_key:" + kid
					}
				}
				return "service_role_ip:" + c.IP()
			},
			Message: fmt.Sprintf("Service role rate limit exceeded. Maximum %d requests per %v allowed.", serviceRoleRateLimit, serviceRoleRateWindow),
		})
	}

	// Cache for per-key rate limiters (keyed by key ID + limit config)
	perKeyLimiters := make(map[string]fiber.Handler)
	var limiterMu sync.RWMutex

	return func(c fiber.Ctx) error {
		role := c.Locals("user_role")
		if role == "service_role" {
			// H-2: Apply rate limiting to service_role tokens if configured
			if serviceRoleLimiter != nil {
				return serviceRoleLimiter(c)
			}
			// Backward compatibility: bypass if no rate limit configured
			return c.Next()
		}

		// Check for per-key rate limits from service key context
		rateLimitPerHour := c.Locals("service_key_rate_limit_per_hour")

		// If no custom rate limit is set (nil), use the default
		if rateLimitPerHour == nil {
			return defaultRateLimiter(c)
		}

		// Get the rate limit value
		limitPtr, ok := rateLimitPerHour.(*int)
		if !ok || limitPtr == nil {
			return defaultRateLimiter(c)
		}
		limit := *limitPtr

		// Get key ID for cache lookup
		keyID := c.Locals("service_key_id")
		keyIDStr := ""
		if keyID != nil {
			if kid, ok := keyID.(string); ok {
				keyIDStr = kid
			}
		}

		// Create cache key based on key ID and limit
		cacheKey := fmt.Sprintf("%s:%d", keyIDStr, limit)

		// Try to get cached limiter
		limiterMu.RLock()
		limiter, exists := perKeyLimiters[cacheKey]
		limiterMu.RUnlock()

		if !exists {
			// Create new limiter for this key's rate limit
			limiter = NewRateLimiter(RateLimiterConfig{
				Max:        limit,
				Expiration: 1 * time.Hour,
				KeyFunc: func(c fiber.Ctx) string {
					keyID := c.Locals("service_key_id")
					if keyID != nil {
						if kid, ok := keyID.(string); ok && kid != "" {
							return "migration_key:" + kid
						}
					}
					return "migration_ip:" + c.IP()
				},
				Message: fmt.Sprintf("Migrations API rate limit exceeded. Maximum %d requests per hour allowed.", limit),
			})

			// Cache the limiter
			limiterMu.Lock()
			perKeyLimiters[cacheKey] = limiter
			limiterMu.Unlock()
		}

		return limiter(c)
	}
}

// StorageUploadLimiter limits file upload requests per user/IP
// Prevents abuse of storage upload endpoints including streaming uploads
func StorageUploadLimiter() fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Name:       "storage_upload",
		Max:        60, // 60 uploads
		Expiration: 1 * time.Minute,
		KeyFunc: func(c fiber.Ctx) string {
			// Try to get user ID from locals (set by auth middleware)
			userID := c.Locals("user_id")
			if userID != nil {
				if uid, ok := userID.(string); ok && uid != "" && uid != "anonymous" {
					return "storage_upload_user:" + uid
				}
			}
			// Fallback to IP for anonymous users
			return "storage_upload_ip:" + c.IP()
		},
		Message: "Storage upload rate limit exceeded. Maximum 60 uploads per minute allowed.",
	})
}

// StorageUploadLimiterWithConfig creates a storage upload rate limiter with custom limits
func StorageUploadLimiterWithConfig(max int, expiration time.Duration) fiber.Handler {
	return NewRateLimiter(RateLimiterConfig{
		Name:       "storage_upload",
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c fiber.Ctx) string {
			// Try to get user ID from locals (set by auth middleware)
			userID := c.Locals("user_id")
			if userID != nil {
				if uid, ok := userID.(string); ok && uid != "" && uid != "anonymous" {
					return "storage_upload_user:" + uid
				}
			}
			// Fallback to IP for anonymous users
			return "storage_upload_ip:" + c.IP()
		},
		Message: fmt.Sprintf("Storage upload rate limit exceeded. Maximum %d requests per %s allowed.", max, expiration.String()),
	})
}

// fiber:context-methods migrated
