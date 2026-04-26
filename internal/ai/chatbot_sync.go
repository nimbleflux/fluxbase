package ai

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/middleware"
	syncframework "github.com/nimbleflux/fluxbase/internal/sync"
)

// SyncChatbotsRequest represents the request body for syncing chatbots
type SyncChatbotsRequest struct {
	Namespace string `json:"namespace"`
	Chatbots  []struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"chatbots"`
	Options struct {
		DeleteMissing bool `json:"delete_missing"`
		DryRun        bool `json:"dry_run"`
	} `json:"options"`
}

// SyncChatbots syncs chatbots from filesystem or SDK payload
// POST /api/v1/admin/ai/chatbots/sync
// If chatbots array is empty, syncs from filesystem. Otherwise syncs provided chatbots.
func (h *Handler) SyncChatbots(c fiber.Ctx) error {
	var req SyncChatbotsRequest
	_ = c.Bind().Body(&req) // Body is optional, continue with defaults

	// Default namespace to "default" if not specified
	namespace := req.Namespace
	if namespace == "" {
		namespace = "default"
	}

	// If no chatbots provided, fall back to filesystem sync
	if len(req.Chatbots) == 0 {
		return h.syncFromFilesystem(c, namespace)
	}

	// Sync from SDK payload
	return h.syncFromPayload(c, namespace, req.Chatbots, req.Options.DeleteMissing, req.Options.DryRun)
}

// syncFromFilesystem syncs chatbots from the filesystem
// All chatbots are synced to the specified namespace (default: "default")
// Any existing chatbot in that namespace not found in the filesystem will be deleted
func (h *Handler) syncFromFilesystem(c fiber.Ctx, namespace string) error {
	ctx := middleware.CtxWithTenant(c)

	// Load chatbots from filesystem
	fsChatbots, err := h.loader.LoadAll()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load chatbots from filesystem")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load chatbots from filesystem",
		})
	}

	// Override namespace for all loaded chatbots with the requested namespace
	for _, cb := range fsChatbots {
		cb.Namespace = namespace
	}

	// Get existing chatbots in this namespace only
	dbChatbots, err := h.storage.ListChatbotsByNamespace(ctx, namespace)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list chatbots from database")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list chatbots from database",
		})
	}

	// Build map of existing chatbots by name (within this namespace)
	existingMap := make(map[string]*Chatbot)
	for _, cb := range dbChatbots {
		existingMap[cb.Name] = cb
	}

	// Track sync results
	createdCount := 0
	updatedCount := 0
	deletedCount := 0
	unchangedCount := 0
	syncErrors := []string{}

	// Track created/updated/deleted names for response
	createdNames := []string{}
	updatedNames := []string{}
	deletedNames := []string{}
	unchangedNames := []string{}

	// Track which chatbots we've processed
	processedNames := make(map[string]bool)

	// Create/update chatbots from filesystem
	for _, fsChatbot := range fsChatbots {
		processedNames[fsChatbot.Name] = true

		existing, exists := existingMap[fsChatbot.Name]
		if exists {
			// Check if update is needed
			if existing.Code == fsChatbot.Code {
				// No change, skip
				unchangedCount++
				unchangedNames = append(unchangedNames, fsChatbot.Name)
				continue
			}

			// Update existing chatbot
			fsChatbot.ID = existing.ID
			fsChatbot.CreatedAt = existing.CreatedAt
			fsChatbot.CreatedBy = existing.CreatedBy
			fsChatbot.Version = existing.Version

			if err := h.storage.UpdateChatbot(ctx, fsChatbot); err != nil {
				log.Error().Err(err).Str("name", fsChatbot.Name).Msg("Failed to update chatbot")
				syncErrors = append(syncErrors, "Failed to update "+fsChatbot.Name+": "+err.Error())
				continue
			}
			updatedCount++
			updatedNames = append(updatedNames, fsChatbot.Name)
		} else {
			// Create new chatbot
			if err := h.storage.CreateChatbot(ctx, fsChatbot); err != nil {
				log.Error().Err(err).Str("name", fsChatbot.Name).Msg("Failed to create chatbot")
				syncErrors = append(syncErrors, "Failed to create "+fsChatbot.Name+": "+err.Error())
				continue
			}
			createdCount++
			createdNames = append(createdNames, fsChatbot.Name)
		}

		// Sync knowledge base links for this chatbot (if KB storage is available)
		if h.knowledgeBaseStorage != nil && len(fsChatbot.KnowledgeBases) > 0 {
			maxChunks := 5
			if fsChatbot.RAGMaxChunks > 0 {
				maxChunks = fsChatbot.RAGMaxChunks
			}
			similarityThreshold := 0.7
			if fsChatbot.RAGSimilarityThreshold > 0 {
				similarityThreshold = fsChatbot.RAGSimilarityThreshold
			}

			if err := h.knowledgeBaseStorage.SyncChatbotKnowledgeBaseLinks(ctx, fsChatbot.ID, fsChatbot.KnowledgeBases, maxChunks, similarityThreshold); err != nil {
				log.Warn().Err(err).Str("chatbot", fsChatbot.Name).Msg("Failed to sync knowledge base links")
			}
		}
	}

	// Delete chatbots in this namespace that are no longer in the filesystem
	for name, dbChatbot := range existingMap {
		if !processedNames[name] {
			if err := h.storage.DeleteChatbot(ctx, dbChatbot.ID); err != nil {
				log.Error().Err(err).Str("name", dbChatbot.Name).Msg("Failed to delete chatbot")
				syncErrors = append(syncErrors, "Failed to delete "+name+": "+err.Error())
				continue
			}
			deletedCount++
			deletedNames = append(deletedNames, name)
		}
	}

	log.Info().
		Int("created", createdCount).
		Int("updated", updatedCount).
		Int("deleted", deletedCount).
		Int("unchanged", unchangedCount).
		Int("errors", len(syncErrors)).
		Str("namespace", namespace).
		Msg("Synced chatbots from filesystem")

	return c.JSON(fiber.Map{
		"message":   "Chatbots synced from filesystem",
		"namespace": namespace,
		"summary": fiber.Map{
			"created":   createdCount,
			"updated":   updatedCount,
			"deleted":   deletedCount,
			"unchanged": unchangedCount,
			"errors":    len(syncErrors),
		},
		"details": fiber.Map{
			"created":   createdNames,
			"updated":   updatedNames,
			"deleted":   deletedNames,
			"unchanged": unchangedNames,
		},
		"errors":  syncErrors,
		"dry_run": false,
	})
}

// syncFromPayload syncs chatbots from SDK payload
func (h *Handler) syncFromPayload(c fiber.Ctx, namespace string, chatbots []struct {
	Name string `json:"name"`
	Code string `json:"code"`
}, deleteMissing bool, dryRun bool,
) error {
	ctx := middleware.CtxWithTenant(c)

	items := make([]chatbotSyncItem, len(chatbots))
	for i, spec := range chatbots {
		items[i] = chatbotSyncItem{name: spec.Name, code: spec.Code}
	}

	opts := syncframework.Options{
		Namespace:     namespace,
		DeleteMissing: deleteMissing,
		DryRun:        dryRun,
	}

	syncer := newChatbotSyncer(h, namespace)
	result, err := syncframework.Execute[chatbotSyncItem](ctx, syncer, items, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}
