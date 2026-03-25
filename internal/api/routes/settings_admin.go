package routes

import (
	"github.com/gofiber/fiber/v3"
)

// SettingsAdminDeps contains dependencies for settings admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all settings including instance-level
//   - tenant_admin: Access to custom settings within their tenant
type SettingsAdminDeps struct {
	// System settings - instance admin only
	ListSystemSettings  fiber.Handler
	GetSystemSetting    fiber.Handler
	UpdateSystemSetting fiber.Handler
	DeleteSystemSetting fiber.Handler

	// Custom settings - tenant accessible
	CreateCustomSetting fiber.Handler
	ListCustomSettings  fiber.Handler
	CreateSecretSetting fiber.Handler
	ListSecretSettings  fiber.Handler
	GetSecretSetting    fiber.Handler
	UpdateSecretSetting fiber.Handler
	DeleteSecretSetting fiber.Handler
	GetUserSecretValue  fiber.Handler
	GetCustomSetting    fiber.Handler
	UpdateCustomSetting fiber.Handler
	DeleteCustomSetting fiber.Handler

	// App settings
	GetAppSettings    fiber.Handler
	UpdateAppSettings fiber.Handler

	// Email settings - instance admin only
	ListEmailSettings   fiber.Handler
	GetEmailSetting     fiber.Handler
	UpdateEmailSetting  fiber.Handler
	TestEmailSettings   fiber.Handler
	ListEmailTemplates  fiber.Handler
	GetEmailTemplate    fiber.Handler
	UpdateEmailTemplate fiber.Handler
	TestEmailTemplate   fiber.Handler
	ResetEmailTemplate  fiber.Handler

	// Captcha settings - instance admin only
	GetCaptchaSettings    fiber.Handler
	UpdateCaptchaSettings fiber.Handler

	// Instance settings - instance admin only
	GetInstanceSettings       fiber.Handler
	UpdateInstanceSettings    fiber.Handler
	GetOverridableSettings    fiber.Handler
	UpdateOverridableSettings fiber.Handler
}

// BuildSettingsAdminRoutes creates the settings admin route group.
func BuildSettingsAdminRoutes(deps *SettingsAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name: "settings_admin",
		Routes: []Route{
			// System Settings - instance admin only
			{Method: "GET", Path: "/system/settings", Handler: deps.ListSystemSettings, Summary: "List system settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/system/settings/*", Handler: deps.GetSystemSetting, Summary: "Get system setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PUT", Path: "/system/settings/*", Handler: deps.UpdateSystemSetting, Summary: "Update system setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "DELETE", Path: "/system/settings/*", Handler: deps.DeleteSystemSetting, Summary: "Delete system setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// Custom Settings - tenant accessible
			{Method: "POST", Path: "/settings/custom", Handler: deps.CreateCustomSetting, Summary: "Create custom setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
			{Method: "GET", Path: "/settings/custom", Handler: deps.ListCustomSettings, Summary: "List custom settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
			{Method: "POST", Path: "/settings/custom/secret", Handler: deps.CreateSecretSetting, Summary: "Create secret setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
			{Method: "GET", Path: "/settings/custom/secrets", Handler: deps.ListSecretSettings, Summary: "List secret settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
			{Method: "GET", Path: "/settings/custom/secret/*", Handler: deps.GetSecretSetting, Summary: "Get secret setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
			{Method: "PUT", Path: "/settings/custom/secret/*", Handler: deps.UpdateSecretSetting, Summary: "Update secret setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
			{Method: "DELETE", Path: "/settings/custom/secret/*", Handler: deps.DeleteSecretSetting, Summary: "Delete secret setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
			{Method: "GET", Path: "/settings/user/:user_id/secret/:key/decrypt", Handler: deps.GetUserSecretValue, Summary: "Decrypt user secret (service_role only)", Auth: AuthRequired, Roles: []string{"service_role"}},
			{Method: "GET", Path: "/settings/custom/*", Handler: deps.GetCustomSetting, Summary: "Get custom setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
			{Method: "PUT", Path: "/settings/custom/*", Handler: deps.UpdateCustomSetting, Summary: "Update custom setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
			{Method: "DELETE", Path: "/settings/custom/*", Handler: deps.DeleteCustomSetting, Summary: "Delete custom setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},

			// App Settings
			{Method: "GET", Path: "/app/settings", Handler: deps.GetAppSettings, Summary: "Get app settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "PUT", Path: "/app/settings", Handler: deps.UpdateAppSettings, Summary: "Update app settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// Email Settings - instance admin only
			{Method: "GET", Path: "/email/settings", Handler: deps.ListEmailSettings, Summary: "List email settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/email/settings/:provider", Handler: deps.GetEmailSetting, Summary: "Get email settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PUT", Path: "/email/settings/:provider", Handler: deps.UpdateEmailSetting, Summary: "Update email settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/email/settings/:provider/test", Handler: deps.TestEmailSettings, Summary: "Test email settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/email/templates", Handler: deps.ListEmailTemplates, Summary: "List email templates", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/email/templates/:name", Handler: deps.GetEmailTemplate, Summary: "Get email template", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PUT", Path: "/email/templates/:name", Handler: deps.UpdateEmailTemplate, Summary: "Update email template", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/email/templates/:name/test", Handler: deps.TestEmailTemplate, Summary: "Test email template", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/email/templates/:name/reset", Handler: deps.ResetEmailTemplate, Summary: "Reset email template", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// Captcha Settings - instance admin only
			{Method: "GET", Path: "/settings/captcha", Handler: deps.GetCaptchaSettings, Summary: "Get captcha settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PUT", Path: "/settings/captcha", Handler: deps.UpdateCaptchaSettings, Summary: "Update captcha settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// Instance Settings - instance admin only
			{Method: "GET", Path: "/instance/settings", Handler: deps.GetInstanceSettings, Summary: "Get instance settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PATCH", Path: "/instance/settings", Handler: deps.UpdateInstanceSettings, Summary: "Update instance settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/instance/settings/overridable", Handler: deps.GetOverridableSettings, Summary: "Get overridable settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PUT", Path: "/instance/settings/overridable", Handler: deps.UpdateOverridableSettings, Summary: "Update overridable settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		},
	}
}
