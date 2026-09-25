//go:build integration
// +build integration

package ai_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/testutil/e2e"
)

// TestChatbotKnowledgeBaseLinkAccessLevel_Integration verifies the storage
// layer can move an existing chatbot→KB link between access levels (e.g.
// "full" → "filtered" for per-user scoping) without unlink/relink, and that
// partial updates which omit access_level leave the existing level intact.
func TestChatbotKnowledgeBaseLinkAccessLevel_Integration(t *testing.T) {
	tc := e2e.NewIntegrationTestContextWithNamespace(t, "ai")
	ctx := context.Background()
	storage := ai.NewKnowledgeBaseStorage(tc.TestContext.DB)

	chatbotID := uuid.New().String()
	kbID := uuid.New().String()

	// Seed the parent rows the link's foreign keys reference. The link is
	// deleted by ON DELETE CASCADE when the parents go.
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO ai.chatbots (id, name, code)
		VALUES ($1, $2, 'return {}')
	`, chatbotID, "itest-kb-link-"+chatbotID)
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO ai.knowledge_bases (id, name)
		VALUES ($1, $2)
	`, kbID, "itest-kb-link-"+kbID)

	defer func() {
		tc.ExecuteSQLAsSuperuser(`DELETE FROM ai.chatbots WHERE id = $1`, chatbotID)
		tc.ExecuteSQLAsSuperuser(`DELETE FROM ai.knowledge_bases WHERE id = $1`, kbID)
	}()

	t.Run("creates the link with the requested access level", func(t *testing.T) {
		link, err := storage.LinkChatbotKnowledgeBaseSimple(ctx, chatbotID, kbID, "full", 1, 5, 0.7)
		require.NoError(t, err)
		require.NotNil(t, link)
		assert.Equal(t, "full", link.AccessLevel)
	})

	t.Run("moves an existing full link to filtered without unlink/relink", func(t *testing.T) {
		filtered := "filtered"
		updated, err := storage.UpdateChatbotKnowledgeBaseLink(ctx, chatbotID, kbID, ai.UpdateChatbotKnowledgeBaseOptions{
			AccessLevel: &filtered,
		})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "filtered", updated.AccessLevel)

		// The change must be persisted, not just applied to the returned copy.
		links, err := storage.GetChatbotKnowledgeBases(ctx, chatbotID)
		require.NoError(t, err)
		require.Len(t, links, 1)
		assert.Equal(t, "filtered", links[0].AccessLevel)
	})

	t.Run("partial update without access_level leaves the level untouched", func(t *testing.T) {
		priority := 2
		updated, err := storage.UpdateChatbotKnowledgeBaseLink(ctx, chatbotID, kbID, ai.UpdateChatbotKnowledgeBaseOptions{
			Priority: &priority,
		})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "filtered", updated.AccessLevel)
		assert.Equal(t, 2, updated.Priority)
	})
}
