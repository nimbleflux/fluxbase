package routes

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

type AuthRequirement string

const (
	AuthNone       AuthRequirement = "none"
	AuthOptional   AuthRequirement = "optional"
	AuthRequired   AuthRequirement = "required"
	AuthDashboard  AuthRequirement = "dashboard"
	AuthUnified    AuthRequirement = "unified"
	AuthServiceKey AuthRequirement = "service_key"
	AuthInternal   AuthRequirement = "internal"
)

// AuthMiddlewares holds the middleware handlers for each auth requirement.
// Route groups provide these to enable automatic auth middleware injection.
type AuthMiddlewares struct {
	None       fiber.Handler // For AuthNone - usually nil
	Optional   fiber.Handler // For AuthOptional
	Required   fiber.Handler // For AuthRequired
	Unified    fiber.Handler // For AuthUnified
	ServiceKey fiber.Handler // For AuthServiceKey
	Internal   fiber.Handler // For AuthInternal
	Dashboard  fiber.Handler // For AuthDashboard
}

// MiddlewareFor returns the middleware handler for the given auth requirement.
func (a *AuthMiddlewares) MiddlewareFor(auth AuthRequirement) fiber.Handler {
	if a == nil {
		return nil
	}
	switch auth {
	case AuthNone:
		return a.None
	case AuthOptional:
		return a.Optional
	case AuthRequired:
		return a.Required
	case AuthUnified:
		return a.Unified
	case AuthServiceKey:
		return a.ServiceKey
	case AuthInternal:
		return a.Internal
	case AuthDashboard:
		return a.Dashboard
	default:
		return nil
	}
}

type RateLimitConfig struct {
	Key      string
	Requests int
	Window   time.Duration
	ByIP     bool
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type RouteAuditEntry struct {
	Method       string
	Path         string
	Group        string
	Summary      string
	Auth         AuthRequirement
	Roles        []string
	Scopes       []string
	Public       bool
	Internal     bool
	TenantScoped bool
	Middlewares  []string
}

type RegistryStats struct {
	TotalRoutes          int
	PublicRoutes         int
	InternalRoutes       int
	RoleProtectedRoutes  int
	ScopeProtectedRoutes int
	ByMethod             map[string]int
}

type Middleware struct {
	Name        string
	Handler     fiber.Handler
	Description string
	DependsOn   []string
	Internal    bool
}

type Route struct {
	Method      string
	Path        string
	Handler     fiber.Handler
	Middlewares []Middleware

	Summary     string
	Description string
	Auth        AuthRequirement
	Scopes      []string
	Roles       []string
	RateLimit   *RateLimitConfig

	TenantScoped       bool
	Public             bool
	Internal           bool
	Deprecated         bool
	DeprecationMessage string
	Tags               []string
}

type RouteGroup struct {
	Name        string
	Prefix      string
	Middlewares []Middleware
	Routes      []Route
	SubGroups   []*RouteGroup
	FeatureFlag string
	Description string

	// AuthMiddlewares provides auth middleware handlers for automatic injection.
	// When set, routes with Auth field set will automatically have the corresponding
	// middleware applied, eliminating the need to specify auth middleware in Middlewares.
	AuthMiddlewares *AuthMiddlewares

	// RequireRole provides role-checking middleware for automatic injection.
	// When set, routes with non-empty Roles field will automatically have
	// this middleware applied with the specified roles.
	RequireRole func(...string) fiber.Handler

	// RequireScope provides scope-checking middleware for automatic injection.
	// When set, routes with non-empty Scopes field will automatically have
	// this middleware applied with the specified scopes.
	RequireScope func(...string) fiber.Handler

	// DefaultAuth is the default auth requirement for routes that don't specify
	// their own Auth field. This is inherited by subgroups unless overridden.
	DefaultAuth AuthRequirement

	// DefaultRoles are the default roles for routes that don't specify their own
	// Roles field. Routes with explicit Roles override this default (no merge).
	// This is inherited by subgroups unless overridden.
	DefaultRoles []string
}

func (r Route) MiddlewareNames() []string {
	names := make([]string, 0, len(r.Middlewares))
	for _, m := range r.Middlewares {
		names = append(names, m.Name)
	}
	return names
}

func (r Route) HasAuth() bool {
	return r.Auth != AuthNone && r.Auth != AuthOptional
}

func (r Route) IsPublic() bool {
	return r.Public || r.Auth == AuthNone
}

func (r Route) FullPath(groupPrefix string) string {
	if groupPrefix == "" {
		return r.Path
	}
	if r.Path == "" || r.Path == "/" {
		return groupPrefix
	}
	return groupPrefix + r.Path
}

func (r Route) Validate() error {
	if r.Method == "" {
		return &ValidationError{Field: "Method", Message: "method is required"}
	}
	if r.Handler == nil {
		return &ValidationError{Field: "Handler", Message: "handler is required"}
	}
	if r.Auth == AuthNone && len(r.Roles) > 0 {
		return &ValidationError{Field: "Roles", Message: "roles specified but auth is none"}
	}
	if r.Auth == AuthNone && len(r.Scopes) > 0 {
		return &ValidationError{Field: "Scopes", Message: "scopes specified but auth is none"}
	}
	if r.Public && r.Auth != AuthNone && r.Auth != AuthOptional {
		return &ValidationError{Field: "Public", Message: "public routes should have AuthNone or AuthOptional"}
	}
	return nil
}

func (g RouteGroup) Validate() error {
	if g.Name == "" {
		return &ValidationError{Field: "Name", Message: "group name is required"}
	}
	for i, r := range g.Routes {
		if err := r.Validate(); err != nil {
			return &ValidationError{
				Field:   fmt.Sprintf("Routes[%d]", i),
				Message: err.Error(),
			}
		}
	}
	for _, sg := range g.SubGroups {
		if err := sg.Validate(); err != nil {
			return err
		}
	}
	return nil
}
