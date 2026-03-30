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
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/*",
				Handler: deps.GetSetting,
				Summary: "Get a setting",
				Auth:    AuthOptional,
			},
			{
				Method:  "GET",
				Path:    "/",
				Handler: deps.GetSettings,
				Summary: "List all settings",
				Auth:    AuthOptional,
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Optional: deps.OptionalAuth,
		},
	}
}

func BuildUserSettingsRoutes(deps *UserSettingsDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "user-settings",
		Prefix: "/api/v1/settings/user",
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
				Summary: "Get system setting (public info)",
				Auth:    AuthRequired,
			},
			{
				Method:  "GET",
				Path:    "/*",
				Handler: deps.GetSetting,
				Summary: "Get a user setting",
				Auth:    AuthRequired,
			},
			{
				Method:  "PUT",
				Path:    "/*",
				Handler: deps.SetSetting,
				Summary: "Set a user setting",
				Auth:    AuthRequired,
			},
			{
				Method:  "DELETE",
				Path:    "/*",
				Handler: deps.DeleteSetting,
				Summary: "Delete a user setting",
				Auth:    AuthRequired,
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
	}
}

func BuildUserSecretsRoutes(deps *UserSettingsDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "user-secrets",
		Prefix: "/api/v1/settings/user/secrets",
		Routes: []Route{
			{
				Method:  "POST",
				Path:    "/",
				Handler: deps.CreateSecret,
				Summary: "Create a user secret",
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
				Summary: "Get a user secret",
				Auth:    AuthRequired,
			},
			{
				Method:  "PUT",
				Path:    "/*",
				Handler: deps.UpdateSecret,
				Summary: "Update a user secret",
				Auth:    AuthRequired,
			},
			{
				Method:  "DELETE",
				Path:    "/*",
				Handler: deps.DeleteSecret,
				Summary: "Delete a user secret",
				Auth:    AuthRequired,
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
	}
}
