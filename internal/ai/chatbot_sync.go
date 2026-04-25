package ai

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/middleware"
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

	// Get existing chatbots in this namespace
	dbChatbots, err := h.storage.ListChatbotsByNamespace(ctx, namespace)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list existing chatbots",
		})
	}

	// Build map of existing chatbots by name
	existingMap := make(map[string]*Chatbot)
	for _, cb := range dbChatbots {
		existingMap[cb.Name] = cb
	}

	// Build set of payload chatbot names
	payloadNames := make(map[string]bool)
	for _, spec := range chatbots {
		payloadNames[spec.Name] = true
	}

	// Track results
	created := []string{}
	updated := []string{}
	deleted := []string{}
	errorList := []fiber.Map{}

	// If dry run, calculate what would be done
	if dryRun {
		for _, spec := range chatbots {
			if _, exists := existingMap[spec.Name]; exists {
				updated = append(updated, spec.Name)
			} else {
				created = append(created, spec.Name)
			}
		}

		if deleteMissing {
			for name := range existingMap {
				if !payloadNames[name] {
					deleted = append(deleted, name)
				}
			}
		}

		return c.JSON(fiber.Map{
			"summary": fiber.Map{
				"created": len(created),
				"updated": len(updated),
				"deleted": len(deleted),
				"errors":  0,
			},
			"details": fiber.Map{
				"created": created,
				"updated": updated,
				"deleted": deleted,
			},
			"dry_run": true,
		})
	}

	// Process chatbots
	for _, spec := range chatbots {
		existing, exists := existingMap[spec.Name]

		// Parse and compile the chatbot code
		chatbot, err := h.loader.ParseChatbotFromCode(spec.Code, namespace)
		if err != nil {
			log.Error().Err(err).Str("name", spec.Name).Msg("Failed to parse chatbot")
			errorList = append(errorList, fiber.Map{
				"name":  spec.Name,
				"error": "Failed to parse chatbot: " + err.Error(),
			})
			continue
		}

		// Set the name and code
		chatbot.Name = spec.Name
		chatbot.Code = spec.Code
		chatbot.Source = "sdk"

		if exists {
			// Update existing chatbot
			chatbot.ID = existing.ID
			chatbot.CreatedAt = existing.CreatedAt
			chatbot.CreatedBy = existing.CreatedBy
			chatbot.Version = existing.Version

			if err := h.storage.UpdateChatbot(ctx, chatbot); err != nil {
				log.Error().Err(err).Str("name", spec.Name).Msg("Failed to update chatbot")
				errorList = append(errorList, fiber.Map{
					"name":  spec.Name,
					"error": "Failed to update: " + err.Error(),
				})
				continue
			}
			updated = append(updated, spec.Name)
		} else {
			// Create new chatbot
			if err := h.storage.CreateChatbot(ctx, chatbot); err != nil {
				log.Error().Err(err).Str("name", spec.Name).Msg("Failed to create chatbot")
				errorList = append(errorList, fiber.Map{
					"name":  spec.Name,
					"error": "Failed to create: " + err.Error(),
				})
				continue
			}
			created = append(created, spec.Name)
		}

		// Sync knowledge base links for this chatbot (if KB storage is available)
		if h.knowledgeBaseStorage != nil && len(chatbot.KnowledgeBases) > 0 {
			maxChunks := 5
			if chatbot.RAGMaxChunks > 0 {
				maxChunks = chatbot.RAGMaxChunks
			}
			similarityThreshold := 0.7
			if chatbot.RAGSimilarityThreshold > 0 {
				similarityThreshold = chatbot.RAGSimilarityThreshold
			}

			if err := h.knowledgeBaseStorage.SyncChatbotKnowledgeBaseLinks(ctx, chatbot.ID, chatbot.KnowledgeBases, maxChunks, similarityThreshold); err != nil {
				log.Warn().Err(err).Str("chatbot", chatbot.Name).Msg("Failed to sync knowledge base links")
			}
		}
	}

	// Delete missing chatbots if requested
	if deleteMissing {
		for name, chatbot := range existingMap {
			if !payloadNames[name] && chatbot.Source == "sdk" {
				if err := h.storage.DeleteChatbot(ctx, chatbot.ID); err != nil {
					log.Error().Err(err).Str("name", name).Msg("Failed to delete chatbot")
					errorList = append(errorList, fiber.Map{
						"name":  name,
						"error": "Failed to delete: " + err.Error(),
					})
					continue
				}
				deleted = append(deleted, name)
			}
		}
	}

	log.Info().
		Int("created", len(created)).
		Int("updated", len(updated)).
		Int("deleted", len(deleted)).
		Int("errors", len(errorList)).
		Str("namespace", namespace).
		Msg("Synced chatbots from SDK payload")

	return c.JSON(fiber.Map{
		"summary": fiber.Map{
			"created": len(created),
			"updated": len(updated),
			"deleted": len(deleted),
			"errors":  len(errorList),
		},
		"details": fiber.Map{
			"created": created,
			"updated": updated,
			"deleted": deleted,
		},
		"errors":  errorList,
		"dry_run": false,
	})
}
