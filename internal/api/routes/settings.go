package routes

import (
	"github.com/gofiber/fiber/v3"
)

type SettingsDeps struct {
	OptionalAuth fiber.Handler
	RequireAuth  fiber.Handler
	GetSetting   fiber.Handler
	GetSettings  fiber.Handler
}

type UserSettingsDeps struct {
	RequireAuth       fiber.Handler
	ListSettings      fiber.Handler
	GetUserOwnSetting fiber.Handler
	GetSystemSetting  fiber.Handler
	GetSetting        fiber.Handler
	SetSetting        fiber.Handler
	DeleteSetting     fiber.Handler
	CreateSecret      fiber.Handler
	ListSecrets       fiber.Handler
	GetSecret         fiber.Handler
	UpdateSecret      fiber.Handler
	DeleteSecret      fiber.Handler
}

func BuildSettingsRoutes(deps *SettingsDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "settings",
		Prefix: "/api/v1/settings",
		Middlewares: []Middleware{
			{Name: "OptionalAuth", Handler: deps.OptionalAuth},
		},
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/:key",
				Handler: deps.GetSetting,
				Summary: "Get setting by key (respects RLS)",
				Auth:    AuthOptional,
			},
			{
				Method:  "POST",
				Path:    "/batch",
				Handler: deps.GetSettings,
				Summary: "Get multiple settings",
				Auth:    AuthOptional,
			},
		},
	}
}

func BuildUserSettingsRoutes(deps *UserSettingsDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "user-settings",
		Prefix: "/api/v1/settings/user",
		Middlewares: []Middleware{
			{Name: "RequireAuth", Handler: deps.RequireAuth},
		},
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/list",
				Handler: deps.ListSettings,
				Summary: "List user's own settings",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/own/:key",
				Handler: deps.GetUserOwnSetting,
				Summary: "Get user's own setting only",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/system/:key",
				Handler: deps.GetSystemSetting,
				Summary: "Get system setting",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/:key",
				Handler: deps.GetSetting,
				Summary: "Get setting with user->system fallback",
				Auth:    AuthRequired,
			},
			{
				Method:  "PUT",
				Path:    "/:key",
				Handler: deps.SetSetting,
				Summary: "Create/update user setting",
				Auth:    AuthRequired,
			},
			{
				Method:  "DELETE",
				Path:    "/:key",
				Handler: deps.DeleteSetting,
				Summary: "Delete user setting",
				Auth:    AuthRequired,
			},
		},
	}
}

func BuildUserSecretsRoutes(deps *UserSettingsDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "user-secrets",
		Prefix: "/api/v1/settings/secret",
		Middlewares: []Middleware{
			{Name: "RequireAuth", Handler: deps.RequireAuth},
		},
		Routes: []Route{
			{
				Method:  "POST",
				Path:    "/",
				Handler: deps.CreateSecret,
				Summary: "Create user secret (encrypted)",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/",
				Handler: deps.ListSecrets,
				Summary: "List user secrets",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/*",
				Handler: deps.GetSecret,
				Summary: "Get user secret",
				Auth:    AuthRequired,
			},
			{
				Method:  "PUT",
				Path:    "/*",
				Handler: deps.UpdateSecret,
				Summary: "Update user secret",
				Auth:    AuthRequired,
			},
			{
				Method:  "DELETE",
				Path:    "/*",
				Handler: deps.DeleteSecret,
				Summary: "Delete user secret",
				Auth:    AuthRequired,
			},
		},
	}
}
