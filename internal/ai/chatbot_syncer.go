package ai

import (
	"context"

	syncframework "github.com/nimbleflux/fluxbase/internal/sync"
)

type chatbotSyncItem struct {
	name string
	code string
}

func (i chatbotSyncItem) GetName() string {
	return i.name
}

type chatbotSyncer struct {
	handler   *Handler
	namespace string
	parsed    map[string]*Chatbot
	existing  map[string]*Chatbot
}

func newChatbotSyncer(h *Handler, namespace string) *chatbotSyncer {
	return &chatbotSyncer{
		handler:   h,
		namespace: namespace,
		parsed:    make(map[string]*Chatbot),
		existing:  make(map[string]*Chatbot),
	}
}

func (s *chatbotSyncer) ListExisting(ctx context.Context, opts syncframework.Options) (map[string]string, error) {
	chatbots, err := s.handler.storage.ListChatbotsByNamespace(ctx, opts.Namespace)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(chatbots))
	for _, cb := range chatbots {
		s.existing[cb.Name] = cb
		result[cb.Name] = cb.ID
	}
	return result, nil
}

func (s *chatbotSyncer) IsChanged(ctx context.Context, existingID string, item chatbotSyncItem, opts syncframework.Options) (bool, error) {
	return true, nil
}

func (s *chatbotSyncer) Preprocess(ctx context.Context, item chatbotSyncItem) error {
	chatbot, err := s.handler.loader.ParseChatbotFromCode(item.code, s.namespace)
	if err != nil {
		return err
	}
	chatbot.Name = item.name
	chatbot.Code = item.code
	chatbot.Source = "sdk"
	s.parsed[item.name] = chatbot
	return nil
}

func (s *chatbotSyncer) Create(ctx context.Context, item chatbotSyncItem, opts syncframework.Options) error {
	chatbot := s.parsed[item.name]
	if err := s.handler.storage.CreateChatbot(ctx, chatbot); err != nil {
		return err
	}
	s.syncKnowledgeBaseLinks(ctx, chatbot)
	return nil
}

func (s *chatbotSyncer) Update(ctx context.Context, item chatbotSyncItem, existingID string, opts syncframework.Options) error {
	chatbot := s.parsed[item.name]
	existing := s.existing[item.name]
	chatbot.ID = existing.ID
	chatbot.CreatedAt = existing.CreatedAt
	chatbot.CreatedBy = existing.CreatedBy
	chatbot.Version = existing.Version
	if err := s.handler.storage.UpdateChatbot(ctx, chatbot); err != nil {
		return err
	}
	s.syncKnowledgeBaseLinks(ctx, chatbot)
	return nil
}

func (s *chatbotSyncer) Delete(ctx context.Context, name string, existingID string, opts syncframework.Options) (bool, error) {
	existing, ok := s.existing[name]
	if !ok || existing.Source != "sdk" {
		return false, nil
	}
	if err := s.handler.storage.DeleteChatbot(ctx, existing.ID); err != nil {
		return false, err
	}
	return true, nil
}

func (s *chatbotSyncer) PostSync(ctx context.Context, result *syncframework.Result, opts syncframework.Options) error {
	return nil
}

func (s *chatbotSyncer) syncKnowledgeBaseLinks(ctx context.Context, chatbot *Chatbot) {
	if s.handler.knowledgeBaseStorage == nil || len(chatbot.KnowledgeBases) == 0 {
		return
	}
	maxChunks := 5
	if chatbot.RAGMaxChunks > 0 {
		maxChunks = chatbot.RAGMaxChunks
	}
	similarityThreshold := 0.7
	if chatbot.RAGSimilarityThreshold > 0 {
		similarityThreshold = chatbot.RAGSimilarityThreshold
	}
	if err := s.handler.knowledgeBaseStorage.SyncChatbotKnowledgeBaseLinks(ctx, chatbot.ID, chatbot.KnowledgeBases, maxChunks, similarityThreshold); err != nil {
		// Log but don't fail the sync
	}
}
