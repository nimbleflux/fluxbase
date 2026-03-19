package middleware

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type migrationsRateLimiter struct {
	limit    int
	window   time.Duration
	storage  fiber.Storage
	requests map[string]*migrationsRateLimitEntry
	mu       sync.Mutex
}

type migrationsRateLimitEntry struct {
	count     int
	expiresAt time.Time
}

func newRateLimiter(limit int, window time.Duration, storage fiber.Storage) *migrationsRateLimiter {
	return &migrationsRateLimiter{
		limit:    limit,
		window:   window,
		storage:  storage,
		requests: make(map[string]*migrationsRateLimitEntry),
	}
}

func (rl *migrationsRateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if entry, exists := rl.requests[key]; exists {
		if now.After(entry.expiresAt) {
			entry.count = 1
			entry.expiresAt = now.Add(rl.window)
			return true
		}
		if entry.count >= rl.limit {
			return false
		}
		entry.count++
		return true
	}

	rl.requests[key] = &migrationsRateLimitEntry{
		count:     1,
		expiresAt: now.Add(rl.window),
	}
	return true
}

func RequireMigrationsFullSecurity(
	cfg *config.MigrationsConfig,
	serverCfg *config.ServerConfig,
	db *pgxpool.Pool,
	authService *auth.Service,
	rateLimit int,
	rateWindow time.Duration,
	storage fiber.Storage,
) fiber.Handler {
	var allowedNets []*net.IPNet
	for _, ipRange := range cfg.AllowedIPRanges {
		_, network, err := net.ParseCIDR(ipRange)
		if err != nil {
			log.Error().Err(err).Str("range", ipRange).Msg("Invalid IP range in migrations config")
			continue
		}
		allowedNets = append(allowedNets, network)
	}

	limiter := newRateLimiter(rateLimit, rateWindow, storage)

	return func(c fiber.Ctx) error {
		if !cfg.Enabled {
			log.Warn().Str("path", c.Path()).Str("ip", c.IP()).Msg("Migrations API access denied - feature disabled")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not Found"})
		}

		if len(allowedNets) > 0 {
			clientIP := GetTrustedClientIP(c, serverCfg)
			allowed := false
			for _, network := range allowedNets {
				if network.Contains(clientIP) {
					allowed = true
					break
				}
			}
			if !allowed {
				log.Warn().Str("ip", clientIP.String()).Str("path", c.Path()).Msg("Migrations API access denied - IP not in allowlist")
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied - IP not allowlisted for migrations"})
			}
		}

		if !migrationsValidateAuthAndScope(c, db, authService) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Service key or service_role JWT authentication required for migrations API"})
		}

		rateLimitKey := migrationsGetRateLimitKey(c)
		if !limiter.allow(rateLimitKey) {
			log.Warn().Str("key", rateLimitKey).Int("limit", rateLimit).Msg("Migrations API rate limit exceeded")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Rate limit exceeded"})
		}

		start := time.Now()
		serviceKeyID := c.Locals("service_key_id")
		serviceKeyName := c.Locals("service_key_name")
		log.Info().Str("method", c.Method()).Str("path", c.Path()).Str("ip", c.IP()).
			Interface("service_key_id", serviceKeyID).Interface("service_key_name", serviceKeyName).
			Msg("Migrations API request started")

		err := c.Next()

		log.Info().Str("method", c.Method()).Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).Dur("duration", time.Since(start)).
			Str("ip", c.IP()).Interface("service_key_id", serviceKeyID).
			Msg("Migrations API request completed")

		return err
	}
}

func migrationsValidateAuthAndScope(c fiber.Ctx, db *pgxpool.Pool, authService *auth.Service) bool {
	authHeader := c.Get("Authorization")
	clientkey := c.Get("clientkey")

	var jwtToken string
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if strings.HasPrefix(token, "eyJ") {
			jwtToken = token
		}
	}
	if jwtToken == "" && strings.HasPrefix(clientkey, "eyJ") {
		jwtToken = clientkey
	}

	if jwtToken != "" {
		claims, err := authService.ValidateToken(jwtToken)
		if err == nil && claims.Role == "service_role" {
			c.Locals("auth_type", "jwt")
			c.Locals("user_role", claims.Role)
			c.Locals("user_id", claims.UserID)
			return true
		}
	}

	serviceKey := c.Get("X-Service-Key")
	if serviceKey == "" {
		if strings.HasPrefix(authHeader, "ServiceKey ") {
			serviceKey = strings.TrimPrefix(authHeader, "ServiceKey ")
		} else if serviceKey == "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if strings.HasPrefix(token, "sk_") {
				serviceKey = token
			}
		}
	}
	if serviceKey == "" && strings.HasPrefix(clientkey, "sk_") {
		serviceKey = clientkey
	}

	if serviceKey != "" && migrationsValidateServiceKeyWithScope(c, db, serviceKey, "migrations:execute") {
		return true
	}

	log.Warn().Str("path", c.Path()).Str("ip", c.IP()).Msg("Migrations API auth failed")
	return false
}

func migrationsValidateServiceKeyWithScope(c fiber.Ctx, db *pgxpool.Pool, serviceKey, requiredScope string) bool {
	if len(serviceKey) < 16 || !strings.HasPrefix(serviceKey, "sk_") {
		return false
	}
	keyPrefix := serviceKey[:16]

	var keyHash, keyID, keyName string
	var scopes []string
	var enabled bool
	var expiresAt *time.Time

	err := db.QueryRow(c.RequestCtx(),
		`SELECT id, name, key_hash, scopes, enabled, expires_at FROM auth.service_keys WHERE key_prefix = $1`,
		keyPrefix,
	).Scan(&keyID, &keyName, &keyHash, &scopes, &enabled, &expiresAt)
	if err != nil {
		return false
	}

	if !enabled || (expiresAt != nil && expiresAt.Before(time.Now())) {
		return false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(serviceKey)); err != nil {
		return false
	}

	hasScope := false
	for _, scope := range scopes {
		if scope == requiredScope || scope == "*" {
			hasScope = true
			break
		}
	}
	if !hasScope {
		log.Warn().Str("required", requiredScope).Interface("scopes", scopes).Msg("Service key missing required scope")
		return false
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = db.Exec(ctx, `UPDATE auth.service_keys SET last_used_at = NOW() WHERE id = $1`, keyID)
	}()

	c.Locals("service_key_id", keyID)
	c.Locals("service_key_name", keyName)
	c.Locals("service_key_scopes", scopes)
	c.Locals("auth_type", "service_key")

	return true
}

func migrationsGetRateLimitKey(c fiber.Ctx) string {
	if keyID := c.Locals("service_key_id"); keyID != nil {
		if id, ok := keyID.(string); ok {
			return "migration:" + id
		}
	}
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			return "migration:" + id
		}
	}
	return "migration:ip:" + c.IP()
}
