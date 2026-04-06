package database

import (
	"context"

	"github.com/rs/zerolog/log"
)

// AuthContextKey is the context key for storing authorization information
type AuthContextKey struct{}

// TenantContextKey is the context key for storing tenant information
type TenantContextKey struct{}

// AuthContext represents authentication and authorization context for schema introspection
type AuthContext struct {
	UserID    string
	UserRole  string // "authenticated", "anon", "service_role", "admin", etc.
	IsAdmin   bool
	ClientKey string // For service role access via client keys
}

// ContextWithAuth returns a context with auth information for audit logging
func ContextWithAuth(ctx context.Context, userID, userRole string, isAdmin bool) context.Context {
	return context.WithValue(ctx, AuthContextKey{}, &AuthContext{
		UserID:   userID,
		UserRole: userRole,
		IsAdmin:  isAdmin,
	})
}

// AuthFromContext extracts auth context from context for audit logging
// Returns nil if no auth context is set
func AuthFromContext(ctx context.Context) *AuthContext {
	if ctx == nil {
		return nil
	}
	auth, ok := ctx.Value(AuthContextKey{}).(*AuthContext)
	if !ok {
		return nil
	}
	return auth
}

// LogSchemaIntrospection logs schema introspection for audit purposes
func LogSchemaIntrospection(ctx context.Context, operation string, details map[string]interface{}) {
	auth := AuthFromContext(ctx)
	if auth != nil {
		log.Info().
			Str("operation", operation).
			Str("user_id", auth.UserID).
			Str("user_role", auth.UserRole).
			Bool("is_admin", auth.IsAdmin).
			Interface("details", details).
			Msg("Schema introspection (authenticated)")
	} else {
		log.Debug().
			Str("operation", operation).
			Interface("details", details).
			Msg("Schema introspection (no auth context)")
	}
}

// ContextWithTenant returns a context with tenant ID set
func ContextWithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantContextKey{}, tenantID)
}

// TenantFromContext extracts tenant ID from context.
// Returns empty string if no tenant context is set
func TenantFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	tenantID, ok := ctx.Value(TenantContextKey{}).(string)
	if !ok {
		return ""
	}
	return tenantID
}

// ContextWithTenantID returns a context with tenant ID for multi-tenancy support
// This is an alias for ContextWithTenant for API consistency
func ContextWithTenantID(ctx context.Context, tenantID string) context.Context {
	return ContextWithTenant(ctx, tenantID)
}

// TenantIDFromContext extracts tenant ID from context
// Returns empty string if no tenant context is set
// This is an alias for TenantFromContext for API consistency
func TenantIDFromContext(ctx context.Context) string {
	return TenantFromContext(ctx)
}
