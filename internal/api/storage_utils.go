package api

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// getTenantIDArg returns the tenant_id as an interface{} for SQL parameters.
// Returns nil if no tenant context is set (which maps to SQL NULL).
func getTenantIDArg(c fiber.Ctx) interface{} {
	if id, ok := c.Locals("tenant_id").(string); ok && id != "" {
		return id
	}
	return nil
}

// detectContentType detects content type from file extension
// SECURITY NOTE: This function only checks file extension, which can be spoofed.
// For enhanced security, consider using detectContentTypeFromBytes() which validates
// magic bytes. However, the primary security control should be:
// 1. Never execute uploaded files
// 2. Serve files with Content-Disposition: attachment
// 3. Use strict CSP headers on storage endpoints
// 4. Implement bucket-level MIME type whitelists
func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".html": "text/html",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// parseMetadata parses metadata from form fields starting with "metadata_"
func parseMetadata(c fiber.Ctx) map[string]string {
	metadata := make(map[string]string)

	for key, value := range c.Request().PostArgs().All() {
		keyStr := string(key)
		if strings.HasPrefix(keyStr, "metadata_") {
			metaKey := strings.TrimPrefix(keyStr, "metadata_")
			metadata[metaKey] = string(value)
		}
	}

	return metadata
}

// getUserID gets the user ID from Fiber context
func getUserID(c fiber.Ctx) string {
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return "anonymous"
}

// setRLSContext sets PostgreSQL session variables for RLS enforcement in a transaction
func (h *StorageHandler) setRLSContext(ctx context.Context, tx pgx.Tx, c fiber.Ctx) error {
	// Get user ID and role from context
	userID := c.Locals("user_id")
	role := c.Locals("user_role")

	// Determine the role
	var roleStr string
	if role != nil {
		if r, ok := role.(string); ok {
			roleStr = r
		}
	}

	// Default role based on authentication state
	if roleStr == "" {
		if userID != nil {
			roleStr = "authenticated"
		} else {
			roleStr = "anon"
		}
	}

	// Convert userID to string
	var userIDStr string
	if userID != nil {
		userIDStr = fmt.Sprintf("%v", userID)
	}

	// Set request.jwt.claims with user ID and role (Supabase/Fluxbase format)
	// This is read by auth.current_user_id() and auth.current_user_role() functions
	var jwtClaims string
	if userIDStr != "" {
		jwtClaims = fmt.Sprintf(`{"sub":"%s","role":"%s"}`, userIDStr, roleStr)
	} else {
		jwtClaims = fmt.Sprintf(`{"role":"%s"}`, roleStr)
	}

	if _, err := tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", jwtClaims); err != nil {
		return fmt.Errorf("failed to set request.jwt.claims: %w", err)
	}

	// Set tenant context for multi-tenancy
	tenantID := c.Locals("tenant_id")
	if tid, ok := tenantID.(string); ok && tid != "" {
		if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tid); err != nil {
			return fmt.Errorf("failed to set tenant context: %w", err)
		}
		log.Debug().Str("tenant_id", tid).Msg("Set tenant context for storage operation")
	}

	// Switch to a non-BYPASSRLS role to enforce RLS policies.
	// The pool connects as fluxbase_app (BYPASSRLS), so without this,
	// all RLS policies are bypassed regardless of FORCE ROW LEVEL SECURITY.
	// Exception: instance_admin and service_role keep BYPASSRLS for full admin access.
	if roleStr == "instance_admin" || roleStr == "service_role" {
		log.Debug().Str("user_id", userIDStr).Str("role", roleStr).Msg("Keeping BYPASSRLS for admin role")
		return nil
	}

	dbRole := "authenticated"
	if roleStr == "anon" {
		dbRole = "anon"
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf("SET LOCAL ROLE %s", quoteIdentifier(dbRole))); err != nil {
		return fmt.Errorf("failed to SET LOCAL ROLE %s: %w", dbRole, err)
	}

	log.Debug().Str("user_id", userIDStr).Str("role", roleStr).Str("db_role", dbRole).Msg("Set RLS context for storage operation")
	return nil
}

// sanitizeFilename sanitizes uploaded filenames to prevent path traversal and control characters
// H-20: Removes null bytes, control characters, and prevents path traversal attacks
func sanitizeFilename(filename string) string {
	if filename == "" {
		return ""
	}

	// Remove null bytes and control characters (except tab)
	var sanitized strings.Builder
	for _, r := range filename {
		if r == '\t' || !unicode.IsControl(r) {
			sanitized.WriteRune(r)
		}
	}
	filename = sanitized.String()

	// Prevent path traversal by removing .. sequences
	filename = strings.ReplaceAll(filename, "..", "")
	// Remove absolute path attempts
	filename = strings.TrimPrefix(filename, "/")
	filename = strings.TrimPrefix(filename, "\\")
	// Remove drive letters (Windows)
	if len(filename) >= 2 && filename[1] == ':' {
		filename = filename[2:]
	}

	// Clean the path but preserve the structure
	filename = filepath.Clean(filename)
	// Remove leading slashes that Clean() might add back
	filename = strings.TrimPrefix(filename, "/")
	filename = strings.TrimPrefix(filename, "\\")

	return filename
}
