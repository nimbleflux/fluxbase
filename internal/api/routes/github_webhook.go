package routes

import (
	"github.com/gofiber/fiber/v3"
)

type GitHubWebhookDeps struct {
	GitHubWebhookLimiter fiber.Handler
	HandleWebhook        fiber.Handler
}

func BuildGitHubWebhookRoutes(deps *GitHubWebhookDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "github-webhook",
		Prefix: "/api/v1/webhooks/github",
		Routes: []Route{
			{
				Method:  "POST",
				Path:    "/",
				Handler: deps.HandleWebhook,
				Summary: "GitHub webhook endpoint for branch automation",
				Auth:    AuthNone,
				Public:  true,
				Middlewares: []Middleware{
					{Name: "GitHubWebhookLimiter", Handler: deps.GitHubWebhookLimiter},
				},
			},
		},
	}
}
