package routes

import (
	"github.com/gofiber/fiber/v3"
)

// AIAdminDeps contains dependencies for AI admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all AI management operations
//   - tenant_admin: Can manage chatbots and view AI tables for their tenant
type AIAdminDeps struct {
	ListChatbots               fiber.Handler
	GetChatbot                 fiber.Handler
	ToggleChatbot              fiber.Handler
	UpdateChatbot              fiber.Handler
	DeleteChatbot              fiber.Handler
	SyncChatbots               fiber.Handler
	GetAIMetrics               fiber.Handler
	ListAIProviders            fiber.Handler
	ListAIConversations        fiber.Handler
	GetAIConversationMessages  fiber.Handler
	GetAIAuditLog              fiber.Handler
	ListExportableTables       fiber.Handler
	GetExportableTableDetails  fiber.Handler
	ExportTableToKnowledgeBase fiber.Handler
	ListChatbotKnowledgeBases  fiber.Handler
	LinkKnowledgeBase          fiber.Handler
	UpdateChatbotKnowledgeBase fiber.Handler
	UnlinkKnowledgeBase        fiber.Handler
}

// BuildAIAdminRoutes creates the AI admin route group.
func BuildAIAdminRoutes(deps *AIAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "ai_admin",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin", "tenant_admin"},
		Routes: []Route{
			// Chatbots (uses default roles)
			{Method: "GET", Path: "/ai/chatbots", Handler: deps.ListChatbots, Summary: "List chatbots"},
			{Method: "GET", Path: "/ai/chatbots/:id", Handler: deps.GetChatbot, Summary: "Get chatbot"},
			{Method: "POST", Path: "/ai/chatbots/:id/toggle", Handler: deps.ToggleChatbot, Summary: "Toggle chatbot"},
			{Method: "PUT", Path: "/ai/chatbots/:id", Handler: deps.UpdateChatbot, Summary: "Update chatbot"},
			{Method: "DELETE", Path: "/ai/chatbots/:id", Handler: deps.DeleteChatbot, Summary: "Delete chatbot"},
			// Sync/Metrics are instance_admin only (override roles)
			{Method: "POST", Path: "/ai/chatbots/sync", Handler: deps.SyncChatbots, Summary: "Sync chatbots", Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/ai/metrics", Handler: deps.GetAIMetrics, Summary: "Get AI metrics", Roles: []string{"admin", "instance_admin"}},

			// AI Providers - instance admin only (override roles)
			{Method: "GET", Path: "/ai/providers", Handler: deps.ListAIProviders, Summary: "List AI providers", Roles: []string{"admin", "instance_admin"}},

			// Conversations & Audit - instance admin only (override roles)
			{Method: "GET", Path: "/ai/conversations", Handler: deps.ListAIConversations, Summary: "List AI conversations", Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/ai/conversations/:id/messages", Handler: deps.GetAIConversationMessages, Summary: "Get AI conversation messages", Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/ai/audit", Handler: deps.GetAIAuditLog, Summary: "Get AI audit log", Roles: []string{"admin", "instance_admin"}},

			// AI Tables (uses default roles)
			{Method: "GET", Path: "/ai/tables", Handler: deps.ListExportableTables, Summary: "List exportable AI tables"},
			{Method: "GET", Path: "/ai/tables/:schema/:table", Handler: deps.GetExportableTableDetails, Summary: "Get exportable table details"},
			{Method: "POST", Path: "/ai/tables/:schema/:table/export", Handler: deps.ExportTableToKnowledgeBase, Summary: "Export table to knowledge base"},

			// Chatbot Knowledge Base linking (uses default roles)
			{Method: "GET", Path: "/ai/chatbots/:id/knowledge-bases", Handler: deps.ListChatbotKnowledgeBases, Summary: "List chatbot knowledge bases"},
			{Method: "POST", Path: "/ai/chatbots/:id/knowledge-bases", Handler: deps.LinkKnowledgeBase, Summary: "Link knowledge base to chatbot"},
			{Method: "PUT", Path: "/ai/chatbots/:id/knowledge-bases/:kb_id", Handler: deps.UpdateChatbotKnowledgeBase, Summary: "Update chatbot knowledge base link"},
			{Method: "DELETE", Path: "/ai/chatbots/:id/knowledge-bases/:kb_id", Handler: deps.UnlinkKnowledgeBase, Summary: "Unlink knowledge base from chatbot"},
		},
	}
}
