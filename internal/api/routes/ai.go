package routes

import (
	"github.com/gofiber/fiber/v3"
)

type AIDeps struct {
	RequireAIEnabled       fiber.Handler
	OptionalAuth           fiber.Handler
	RequireAuth            fiber.Handler
	HandleWebSocket        fiber.Handler
	ListPublicChatbots     fiber.Handler
	LookupChatbotByName    fiber.Handler
	GetPublicChatbot       fiber.Handler
	ListUserConversations  fiber.Handler
	GetUserConversation    fiber.Handler
	DeleteUserConversation fiber.Handler
	UpdateUserConversation fiber.Handler
}

func BuildAIRoutes(deps *AIDeps) *RouteGroup {
	requireAI := []Middleware{{Name: "RequireAIEnabled", Handler: deps.RequireAIEnabled}}

	return &RouteGroup{
		Name: "ai",
		Routes: []Route{
			{
				Method:      "GET",
				Path:        "/ai/ws",
				Handler:     deps.HandleWebSocket,
				Summary:     "WebSocket endpoint for AI chat",
				Auth:        AuthOptional,
				Middlewares: append(requireAI, Middleware{Name: "OptionalAuth", Handler: deps.OptionalAuth}),
			},
			{
				Method:      "GET",
				Path:        "/api/v1/ai/chatbots",
				Handler:     deps.ListPublicChatbots,
				Summary:     "List public chatbots",
				Auth:        AuthOptional,
				Middlewares: append(requireAI, Middleware{Name: "OptionalAuth", Handler: deps.OptionalAuth}),
			},
			{
				Method:      "GET",
				Path:        "/api/v1/ai/chatbots/by-name/:name",
				Handler:     deps.LookupChatbotByName,
				Summary:     "Lookup chatbot by name (smart namespace resolution)",
				Auth:        AuthOptional,
				Middlewares: append(requireAI, Middleware{Name: "OptionalAuth", Handler: deps.OptionalAuth}),
			},
			{
				Method:      "GET",
				Path:        "/api/v1/ai/chatbots/:id",
				Handler:     deps.GetPublicChatbot,
				Summary:     "Get public chatbot by ID",
				Auth:        AuthOptional,
				Middlewares: append(requireAI, Middleware{Name: "OptionalAuth", Handler: deps.OptionalAuth}),
			},
			{
				Method:      "GET",
				Path:        "/api/v1/ai/conversations",
				Handler:     deps.ListUserConversations,
				Summary:     "List user's conversations",
				Auth:        AuthRequired,
				Middlewares: append(requireAI, Middleware{Name: "RequireAuth", Handler: deps.RequireAuth}),
			},
			{
				Method:      "GET",
				Path:        "/api/v1/ai/conversations/:id",
				Handler:     deps.GetUserConversation,
				Summary:     "Get user conversation",
				Auth:        AuthRequired,
				Middlewares: append(requireAI, Middleware{Name: "RequireAuth", Handler: deps.RequireAuth}),
			},
			{
				Method:      "DELETE",
				Path:        "/api/v1/ai/conversations/:id",
				Handler:     deps.DeleteUserConversation,
				Summary:     "Delete user conversation",
				Auth:        AuthRequired,
				Middlewares: append(requireAI, Middleware{Name: "RequireAuth", Handler: deps.RequireAuth}),
			},
			{
				Method:      "PATCH",
				Path:        "/api/v1/ai/conversations/:id",
				Handler:     deps.UpdateUserConversation,
				Summary:     "Update user conversation",
				Auth:        AuthRequired,
				Middlewares: append(requireAI, Middleware{Name: "RequireAuth", Handler: deps.RequireAuth}),
			},
		},
	}
}
