package routes

import (
	"fmt"
	"strings"
)

// ValidateAuthConsistency ensures auth requirements are consistent with roles/scopes.
func ValidateAuthConsistency(group *RouteGroup, route Route, fullpath string) error {
	if route.Auth == AuthNone || route.Auth == AuthOptional {
		if len(route.Roles) > 0 {
			return fmt.Errorf("route %s: roles specified with %s auth (roles require AuthRequired or higher)", fullpath, route.Auth)
		}
	}

	if route.Auth == AuthNone {
		if len(route.Scopes) > 0 {
			return fmt.Errorf("route %s: scopes specified with AuthNone (scopes require authentication)", fullpath)
		}
	}

	if route.Public && route.Auth != AuthNone && route.Auth != AuthOptional {
		return fmt.Errorf("route %s: marked as public but has auth requirement %s", fullpath, route.Auth)
	}

	if route.Internal && route.Auth == AuthNone && !route.Public {
		return fmt.Errorf("route %s: internal route without auth should be marked public for clarity", fullpath)
	}

	return nil
}

// ValidateMiddlewareDependencies ensures required middleware is present.
func ValidateMiddlewareDependencies(group *RouteGroup, route Route, fullpath string) error {
	allMiddleware := make([]Middleware, 0, len(group.Middlewares)+len(route.Middlewares))
	allMiddleware = append(allMiddleware, group.Middlewares...)
	allMiddleware = append(allMiddleware, route.Middlewares...)

	middlewareNames := make(map[string]bool)
	for _, mw := range allMiddleware {
		middlewareNames[mw.Name] = true
	}

	for _, mw := range allMiddleware {
		for _, dep := range mw.DependsOn {
			if !middlewareNames[dep] {
				return fmt.Errorf("route %s: middleware %q depends on %q which is not present", fullpath, mw.Name, dep)
			}
		}
	}

	// Check for auth - either via explicit middleware or auto-injected via AuthMiddlewares or RequireRole
	if len(route.Roles) > 0 || len(route.Scopes) > 0 {
		hasAuth := false

		// Check explicit middleware
		for _, mw := range allMiddleware {
			if strings.Contains(strings.ToLower(mw.Name), "auth") {
				hasAuth = true
				break
			}
		}

		// Determine effective auth: route.Auth > group.DefaultAuth
		effectiveAuth := route.Auth
		if effectiveAuth == "" {
			effectiveAuth = group.DefaultAuth
		}

		// Check auto-injected auth middleware from AuthMiddlewares
		// Note: AuthMiddlewares may be inherited from parent at apply time,
		// so we also check if DefaultAuth is set (indicating auth will be provided)
		if !hasAuth && group.AuthMiddlewares != nil && effectiveAuth != AuthNone && effectiveAuth != "" {
			if group.AuthMiddlewares.MiddlewareFor(effectiveAuth) != nil {
				hasAuth = true
			}
		}

		// Check auto-injected role middleware from RequireRole (role middleware includes auth check)
		// Note: RequireRole may be inherited from parent at apply time
		if !hasAuth && group.RequireRole != nil && len(route.Roles) > 0 {
			hasAuth = true
		}

		// If the group has DefaultAuth or DefaultRoles set, auth will be provided via inheritance
		// even if AuthMiddlewares/RequireRole are not set on this group directly
		if !hasAuth && (group.DefaultAuth != "" && group.DefaultAuth != AuthNone) {
			hasAuth = true
		}

		if !hasAuth && effectiveAuth != AuthNone && effectiveAuth != "" {
			return fmt.Errorf("route %s: has roles/scopes but no auth middleware detected (neither explicit nor via AuthMiddlewares or RequireRole)", fullpath)
		}
	}

	return nil
}

// ValidatePublicRoutes ensures public routes are intentional and documented.
func ValidatePublicRoutes(group *RouteGroup, route Route, fullpath string) error {
	if !route.Public {
		return nil
	}

	if route.Summary == "" {
		return fmt.Errorf("route %s: public routes must have a summary documenting why they're public", fullpath)
	}

	if route.Auth == AuthRequired || route.Auth == AuthDashboard || route.Auth == AuthServiceKey {
		return fmt.Errorf("route %s: marked public but has strict auth requirement %s", fullpath, route.Auth)
	}

	return nil
}

// ValidateFeatureFlags ensures feature-flagged routes are properly documented.
func ValidateFeatureFlags(group *RouteGroup, route Route, fullpath string) error {
	if group.FeatureFlag != "" && route.Summary == "" {
		return fmt.Errorf("route %s: feature-flagged route (flag: %s) must have a summary", fullpath, group.FeatureFlag)
	}
	return nil
}

// ValidateRateLimiting ensures rate-limited routes have proper configuration.
func ValidateRateLimiting(group *RouteGroup, route Route, fullpath string) error {
	if route.RateLimit == nil {
		return nil
	}

	if route.RateLimit.Key == "" {
		return fmt.Errorf("route %s: rate limit configured but no key specified", fullpath)
	}

	if route.RateLimit.Requests > 0 && route.RateLimit.Window == 0 {
		return fmt.Errorf("route %s: rate limit has requests but no window", fullpath)
	}

	return nil
}
