package routes

import (
	"github.com/gofiber/fiber/v3"
)

type DashboardAuthDeps struct {
	SetupLimiter    fiber.Handler
	LoginLimiter    fiber.Handler
	GetSetupStatus  fiber.Handler
	InitialSetup    fiber.Handler
	AdminLogin      fiber.Handler
	RefreshToken    fiber.Handler
	UnifiedAuth     fiber.Handler
	AdminLogout     fiber.Handler
	GetCurrentAdmin fiber.Handler
}

func BuildDashboardAuthRoutes(deps *DashboardAuthDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "dashboard-auth",
		Prefix: "/api/v1/admin",
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/setup/status",
				Handler: deps.GetSetupStatus,
				Summary: "Get dashboard setup status (public)",
				Auth:    AuthNone,
				Public:  true,
			},
			{
				Method:  "POST",
				Path:    "/setup",
				Handler: deps.InitialSetup,
				Middlewares: []Middleware{
					{Name: "SetupLimiter", Handler: deps.SetupLimiter},
				},
				Summary: "Initial dashboard setup (public)",
				Auth:    AuthNone,
				Public:  true,
			},
			{
				Method:  "POST",
				Path:    "/login",
				Handler: deps.AdminLogin,
				Middlewares: []Middleware{
					{Name: "LoginLimiter", Handler: deps.LoginLimiter},
				},
				Summary: "Dashboard admin login (public)",
				Auth:    AuthNone,
				Public:  true,
			},
			{
				Method:  "POST",
				Path:    "/refresh",
				Handler: deps.RefreshToken,
				Summary: "Refresh dashboard token (public)",
				Auth:    AuthNone,
				Public:  true,
			},
			{
				Method:  "POST",
				Path:    "/logout",
				Handler: deps.AdminLogout,
				Middlewares: []Middleware{
					{Name: "UnifiedAuth", Handler: deps.UnifiedAuth},
				},
				Summary: "Dashboard admin logout",
				Auth:    AuthUnified,
			},
			{
				Method:  "GET",
				Path:    "/me",
				Handler: deps.GetCurrentAdmin,
				Middlewares: []Middleware{
					{Name: "UnifiedAuth", Handler: deps.UnifiedAuth},
				},
				Summary: "Get current admin user",
				Auth:    AuthUnified,
			},
		},
	}
}
