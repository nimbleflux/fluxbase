package routes

import (
	"github.com/gofiber/fiber/v3"
)

type WebhookDeps struct {
	RequireAuth    fiber.Handler
	RequireScope   func(...string) fiber.Handler
	ListWebhooks   fiber.Handler
	GetWebhook     fiber.Handler
	ListDeliveries fiber.Handler
	CreateWebhook  fiber.Handler
	UpdateWebhook  fiber.Handler
	DeleteWebhook  fiber.Handler
	TestWebhook    fiber.Handler
}

func BuildWebhookRoutes(deps *WebhookDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "webhooks",
		Prefix: "/api/v1/webhooks",
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/",
				Handler: deps.ListWebhooks,
				Summary: "List all webhooks",
				Auth:    AuthRequired,
				Scopes:  []string{"webhooks:read"},
			},
			{
				Method:  "GET",
				Path:    "/:id",
				Handler: deps.GetWebhook,
				Summary: "Get a webhook by ID",
				Auth:    AuthRequired,
				Scopes:  []string{"webhooks:read"},
			},
			{
				Method:  "GET",
				Path:    "/:id/deliveries",
				Handler: deps.ListDeliveries,
				Summary: "List webhook deliveries",
				Auth:    AuthRequired,
				Scopes:  []string{"webhooks:read"},
			},
			{
				Method:  "POST",
				Path:    "/",
				Handler: deps.CreateWebhook,
				Summary: "Create a new webhook",
				Auth:    AuthRequired,
				Scopes:  []string{"webhooks:write"},
			},
			{
				Method:  "PATCH",
				Path:    "/:id",
				Handler: deps.UpdateWebhook,
				Summary: "Update a webhook",
				Auth:    AuthRequired,
				Scopes:  []string{"webhooks:write"},
			},
			{
				Method:  "DELETE",
				Path:    "/:id",
				Handler: deps.DeleteWebhook,
				Summary: "Delete a webhook",
				Auth:    AuthRequired,
				Scopes:  []string{"webhooks:write"},
			},
			{
				Method:  "POST",
				Path:    "/:id/test",
				Handler: deps.TestWebhook,
				Summary: "Test a webhook",
				Auth:    AuthRequired,
				Scopes:  []string{"webhooks:write"},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
		RequireScope: deps.RequireScope,
	}
}
