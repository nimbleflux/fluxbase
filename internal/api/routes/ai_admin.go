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
		Name: "ai_admin",
		Routes: []Route{
			// Chatbots
			{Method: "GET", Path: "/ai/chatbots", Handler: deps.ListChatbots, Summary: "List chatbots", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/ai/chatbots/:id", Handler: deps.GetChatbot, Summary: "Get chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/ai/chatbots/:id/toggle", Handler: deps.ToggleChatbot, Summary: "Toggle chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "PUT", Path: "/ai/chatbots/:id", Handler: deps.UpdateChatbot, Summary: "Update chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "DELETE", Path: "/ai/chatbots/:id", Handler: deps.DeleteChatbot, Summary: "Delete chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/ai/chatbots/sync", Handler: deps.SyncChatbots, Summary: "Sync chatbots", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/ai/metrics", Handler: deps.GetAIMetrics, Summary: "Get AI metrics", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// AI Providers - instance admin only
			{Method: "GET", Path: "/ai/providers", Handler: deps.ListAIProviders, Summary: "List AI providers", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// Conversations & Audit - instance admin only
			{Method: "GET", Path: "/ai/conversations", Handler: deps.ListAIConversations, Summary: "List AI conversations", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/ai/conversations/:id/messages", Handler: deps.GetAIConversationMessages, Summary: "Get AI conversation messages", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/ai/audit", Handler: deps.GetAIAuditLog, Summary: "Get AI audit log", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// AI Tables
			{Method: "GET", Path: "/ai/tables", Handler: deps.ListExportableTables, Summary: "List exportable AI tables", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/ai/tables/:schema/:table", Handler: deps.GetExportableTableDetails, Summary: "Get exportable table details", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/ai/tables/:schema/:table/export", Handler: deps.ExportTableToKnowledgeBase, Summary: "Export table to knowledge base", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// Chatbot Knowledge Base linking
			{Method: "GET", Path: "/ai/chatbots/:id/knowledge-bases", Handler: deps.ListChatbotKnowledgeBases, Summary: "List chatbot knowledge bases", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/ai/chatbots/:id/knowledge-bases", Handler: deps.LinkKnowledgeBase, Summary: "Link knowledge base to chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "PUT", Path: "/ai/chatbots/:id/knowledge-bases/:kb_id", Handler: deps.UpdateChatbotKnowledgeBase, Summary: "Update chatbot knowledge base link", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "DELETE", Path: "/ai/chatbots/:id/knowledge-bases/:kb_id", Handler: deps.UnlinkKnowledgeBase, Summary: "Unlink knowledge base from chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		},
	}
}
