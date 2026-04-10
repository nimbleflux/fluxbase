import { describe, it, expect, beforeEach, vi } from "vitest";
import { FluxbaseAdminAI } from "./admin-ai";
import { FluxbaseFetch } from "./fetch";
import type {
  AIChatbot,
  AIChatbotSummary,
  AIProvider,
  SyncChatbotsResult,
  ChatbotKnowledgeBaseLink,
} from "./types";

// Mock FluxbaseFetch
vi.mock("./fetch");

describe("FluxbaseAdminAI", () => {
  let ai: FluxbaseAdminAI;
  let mockFetch: any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    };
    ai = new FluxbaseAdminAI(mockFetch as unknown as FluxbaseFetch);
  });

  describe("Chatbot Management", () => {
    describe("listChatbots()", () => {
      it("should list all chatbots", async () => {
        const response = {
          chatbots: [
            {
              id: "bot-1",
              name: "assistant",
              namespace: "default",
              enabled: true,
            },
            {
              id: "bot-2",
              name: "sql-helper",
              namespace: "default",
              enabled: false,
            },
          ] as AIChatbotSummary[],
          count: 2,
        };

        vi.mocked(mockFetch.get).mockResolvedValue(response);

        const { data, error } = await ai.listChatbots();

        expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/admin/ai/chatbots");
        expect(error).toBeNull();
        expect(data).toHaveLength(2);
      });

      it("should list chatbots by namespace", async () => {
        vi.mocked(mockFetch.get).mockResolvedValue({ chatbots: [], count: 0 });

        await ai.listChatbots("custom");

        expect(mockFetch.get).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots?namespace=custom",
        );
      });

      it("should handle empty response", async () => {
        vi.mocked(mockFetch.get).mockResolvedValue({
          chatbots: null,
          count: 0,
        });

        const { data, error } = await ai.listChatbots();

        expect(error).toBeNull();
        expect(data).toEqual([]);
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

        const { data, error } = await ai.listChatbots();

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("getChatbot()", () => {
      it("should get a specific chatbot", async () => {
        const response: AIChatbot = {
          id: "bot-1",
          name: "assistant",
          namespace: "default",
          enabled: true,
          system_prompt: "You are a helpful assistant",
          created_at: "2024-01-26T10:00:00Z",
        };

        vi.mocked(mockFetch.get).mockResolvedValue(response);

        const { data, error } = await ai.getChatbot("bot-1");

        expect(mockFetch.get).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots/bot-1",
        );
        expect(error).toBeNull();
        expect(data!.name).toBe("assistant");
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.get).mockRejectedValue(new Error("Not found"));

        const { data, error } = await ai.getChatbot("unknown");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("toggleChatbot()", () => {
      it("should enable a chatbot", async () => {
        const response: AIChatbot = {
          id: "bot-1",
          name: "assistant",
          namespace: "default",
          enabled: true,
          created_at: "2024-01-26T10:00:00Z",
        };

        vi.mocked(mockFetch.put).mockResolvedValue(response);

        const { data, error } = await ai.toggleChatbot("bot-1", true);

        expect(mockFetch.put).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots/bot-1/toggle",
          { enabled: true },
        );
        expect(error).toBeNull();
        expect(data!.enabled).toBe(true);
      });

      it("should disable a chatbot", async () => {
        const response: AIChatbot = {
          id: "bot-1",
          name: "assistant",
          namespace: "default",
          enabled: false,
          created_at: "2024-01-26T10:00:00Z",
        };

        vi.mocked(mockFetch.put).mockResolvedValue(response);

        const { data, error } = await ai.toggleChatbot("bot-1", false);

        expect(data!.enabled).toBe(false);
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.put).mockRejectedValue(new Error("Update failed"));

        const { data, error } = await ai.toggleChatbot("bot-1", true);

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("deleteChatbot()", () => {
      it("should delete a chatbot", async () => {
        vi.mocked(mockFetch.delete).mockResolvedValue({});

        const { data, error } = await ai.deleteChatbot("bot-1");

        expect(mockFetch.delete).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots/bot-1",
        );
        expect(error).toBeNull();
        expect(data).toBeNull();
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.delete).mockRejectedValue(
          new Error("Delete failed"),
        );

        const { data, error } = await ai.deleteChatbot("bot-1");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("sync()", () => {
      it("should sync chatbots without options", async () => {
        const response: SyncChatbotsResult = {
          message: "Sync completed",
          namespace: "default",
          summary: {
            created: 1,
            updated: 0,
            deleted: 0,
            unchanged: 0,
            errors: 0,
          },
          details: {
            created: ["bot-1"],
            updated: [],
            deleted: [],
            unchanged: [],
            errors: [],
          },
          dry_run: false,
        };

        vi.mocked(mockFetch.post).mockResolvedValue(response);

        const { data, error } = await ai.sync();

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots/sync",
          {
            namespace: "default",
            chatbots: undefined,
            options: { delete_missing: false, dry_run: false },
          },
        );
        expect(error).toBeNull();
      });

      it("should sync with provided chatbots", async () => {
        const response: SyncChatbotsResult = {
          message: "Sync completed",
          namespace: "custom",
          summary: {
            created: 1,
            updated: 0,
            deleted: 0,
            unchanged: 0,
            errors: 0,
          },
          details: {
            created: ["my-bot"],
            updated: [],
            deleted: [],
            unchanged: [],
            errors: [],
          },
          dry_run: false,
        };

        vi.mocked(mockFetch.post).mockResolvedValue(response);

        await ai.sync({
          namespace: "custom",
          chatbots: [{ name: "my-bot", code: "system: You are helpful" }],
          options: { delete_missing: true, dry_run: false },
        });

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots/sync",
          {
            namespace: "custom",
            chatbots: [{ name: "my-bot", code: "system: You are helpful" }],
            options: { delete_missing: true, dry_run: false },
          },
        );
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.post).mockRejectedValue(new Error("Sync failed"));

        const { data, error } = await ai.sync();

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });
  });

  describe("Provider Management", () => {
    describe("listProviders()", () => {
      it("should list all providers", async () => {
        const response = {
          providers: [
            {
              id: "prov-1",
              name: "openai",
              display_name: "OpenAI",
              enabled: true,
            },
            {
              id: "prov-2",
              name: "anthropic",
              display_name: "Anthropic",
              enabled: true,
            },
          ] as AIProvider[],
          count: 2,
        };

        vi.mocked(mockFetch.get).mockResolvedValue(response);

        const { data, error } = await ai.listProviders();

        expect(mockFetch.get).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers",
        );
        expect(error).toBeNull();
        expect(data).toHaveLength(2);
      });

      it("should handle empty response", async () => {
        vi.mocked(mockFetch.get).mockResolvedValue({
          providers: null,
          count: 0,
        });

        const { data, error } = await ai.listProviders();

        expect(error).toBeNull();
        expect(data).toEqual([]);
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

        const { data, error } = await ai.listProviders();

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("getProvider()", () => {
      it("should get a specific provider", async () => {
        const response: AIProvider = {
          id: "prov-1",
          name: "openai",
          display_name: "OpenAI",
          provider_type: "openai",
          enabled: true,
          is_default: true,
          created_at: "2024-01-26T10:00:00Z",
        };

        vi.mocked(mockFetch.get).mockResolvedValue(response);

        const { data, error } = await ai.getProvider("prov-1");

        expect(mockFetch.get).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers/prov-1",
        );
        expect(error).toBeNull();
        expect(data!.name).toBe("openai");
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.get).mockRejectedValue(new Error("Not found"));

        const { data, error } = await ai.getProvider("unknown");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("createProvider()", () => {
      it("should create a provider", async () => {
        const response: AIProvider = {
          id: "prov-1",
          name: "openai-main",
          display_name: "OpenAI (Main)",
          provider_type: "openai",
          enabled: true,
          is_default: true,
          created_at: "2024-01-26T10:00:00Z",
        };

        vi.mocked(mockFetch.post).mockResolvedValue(response);

        const { data, error } = await ai.createProvider({
          name: "openai-main",
          display_name: "OpenAI (Main)",
          provider_type: "openai",
          is_default: true,
          config: { api_key: "sk-xxx", model: "gpt-4-turbo" },
        });

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers",
          {
            name: "openai-main",
            display_name: "OpenAI (Main)",
            provider_type: "openai",
            is_default: true,
            config: { api_key: "sk-xxx", model: "gpt-4-turbo" },
          },
        );
        expect(error).toBeNull();
      });

      it("should normalize config values to strings", async () => {
        vi.mocked(mockFetch.post).mockResolvedValue({});

        await ai.createProvider({
          name: "test",
          provider_type: "openai",
          config: { max_tokens: 100, temperature: 0.7 } as any,
        });

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers",
          {
            name: "test",
            provider_type: "openai",
            config: { max_tokens: "100", temperature: "0.7" },
          },
        );
      });

      it("should skip undefined and null config values", async () => {
        vi.mocked(mockFetch.post).mockResolvedValue({});

        await ai.createProvider({
          name: "test",
          provider_type: "openai",
          config: {
            api_key: "key",
            optional: undefined,
            nullable: null,
          } as any,
        });

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers",
          {
            name: "test",
            provider_type: "openai",
            config: { api_key: "key" },
          },
        );
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.post).mockRejectedValue(new Error("Create failed"));

        const { data, error } = await ai.createProvider({
          name: "test",
          provider_type: "openai",
        });

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("updateProvider()", () => {
      it("should update a provider", async () => {
        const response: AIProvider = {
          id: "prov-1",
          name: "openai",
          display_name: "Updated Name",
          provider_type: "openai",
          enabled: true,
          created_at: "2024-01-26T10:00:00Z",
        };

        vi.mocked(mockFetch.put).mockResolvedValue(response);

        const { data, error } = await ai.updateProvider("prov-1", {
          display_name: "Updated Name",
          enabled: true,
        });

        expect(mockFetch.put).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers/prov-1",
          {
            display_name: "Updated Name",
            enabled: true,
          },
        );
        expect(error).toBeNull();
      });

      it("should normalize config values on update", async () => {
        vi.mocked(mockFetch.put).mockResolvedValue({});

        await ai.updateProvider("prov-1", {
          config: { max_tokens: 200 } as any,
        });

        expect(mockFetch.put).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers/prov-1",
          {
            config: { max_tokens: "200" },
          },
        );
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.put).mockRejectedValue(new Error("Update failed"));

        const { data, error } = await ai.updateProvider("prov-1", {});

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("setDefaultProvider()", () => {
      it("should set default provider", async () => {
        const response: AIProvider = {
          id: "prov-1",
          name: "openai",
          display_name: "OpenAI",
          provider_type: "openai",
          is_default: true,
          enabled: true,
          created_at: "2024-01-26T10:00:00Z",
        };

        vi.mocked(mockFetch.put).mockResolvedValue(response);

        const { data, error } = await ai.setDefaultProvider("prov-1");

        expect(mockFetch.put).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers/prov-1/default",
          {},
        );
        expect(error).toBeNull();
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.put).mockRejectedValue(new Error("Failed"));

        const { data, error } = await ai.setDefaultProvider("prov-1");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("deleteProvider()", () => {
      it("should delete a provider", async () => {
        vi.mocked(mockFetch.delete).mockResolvedValue({});

        const { data, error } = await ai.deleteProvider("prov-1");

        expect(mockFetch.delete).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers/prov-1",
        );
        expect(error).toBeNull();
        expect(data).toBeNull();
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.delete).mockRejectedValue(
          new Error("Delete failed"),
        );

        const { data, error } = await ai.deleteProvider("prov-1");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("setEmbeddingProvider()", () => {
      it("should set embedding provider", async () => {
        const response = { id: "prov-1", use_for_embeddings: true };
        vi.mocked(mockFetch.put).mockResolvedValue(response);

        const { data, error } = await ai.setEmbeddingProvider("prov-1");

        expect(mockFetch.put).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers/prov-1/embedding",
          {},
        );
        expect(error).toBeNull();
        expect(data!.use_for_embeddings).toBe(true);
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.put).mockRejectedValue(new Error("Failed"));

        const { data, error } = await ai.setEmbeddingProvider("prov-1");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("clearEmbeddingProvider()", () => {
      it("should clear embedding provider", async () => {
        const response = { use_for_embeddings: false };
        vi.mocked(mockFetch.delete).mockResolvedValue(response);

        const { data, error } = await ai.clearEmbeddingProvider("prov-1");

        expect(mockFetch.delete).toHaveBeenCalledWith(
          "/api/v1/admin/ai/providers/prov-1/embedding",
        );
        expect(error).toBeNull();
        expect(data!.use_for_embeddings).toBe(false);
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.delete).mockRejectedValue(new Error("Failed"));

        const { data, error } = await ai.clearEmbeddingProvider("prov-1");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });
  });

  describe("Chatbot Knowledge Base Linking", () => {
    describe("listChatbotKnowledgeBases()", () => {
      it("should list linked knowledge bases", async () => {
        const response = {
          knowledge_bases: [
            { chatbot_id: "bot-1", knowledge_base_id: "kb-1", priority: 1 },
            { chatbot_id: "bot-1", knowledge_base_id: "kb-2", priority: 2 },
          ] as ChatbotKnowledgeBaseLink[],
          count: 2,
        };

        vi.mocked(mockFetch.get).mockResolvedValue(response);

        const { data, error } = await ai.listChatbotKnowledgeBases("bot-1");

        expect(mockFetch.get).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots/bot-1/knowledge-bases",
        );
        expect(error).toBeNull();
        expect(data).toHaveLength(2);
      });

      it("should handle empty response", async () => {
        vi.mocked(mockFetch.get).mockResolvedValue({
          knowledge_bases: null,
          count: 0,
        });

        const { data, error } = await ai.listChatbotKnowledgeBases("bot-1");

        expect(error).toBeNull();
        expect(data).toEqual([]);
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

        const { data, error } = await ai.listChatbotKnowledgeBases("bot-1");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("linkKnowledgeBase()", () => {
      it("should link a knowledge base", async () => {
        const response: ChatbotKnowledgeBaseLink = {
          chatbot_id: "bot-1",
          knowledge_base_id: "kb-1",
          priority: 1,
          max_chunks: 5,
          similarity_threshold: 0.7,
          enabled: true,
        };

        vi.mocked(mockFetch.post).mockResolvedValue(response);

        const { data, error } = await ai.linkKnowledgeBase("bot-1", {
          knowledge_base_id: "kb-1",
          priority: 1,
          max_chunks: 5,
          similarity_threshold: 0.7,
        });

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots/bot-1/knowledge-bases",
          {
            knowledge_base_id: "kb-1",
            priority: 1,
            max_chunks: 5,
            similarity_threshold: 0.7,
          },
        );
        expect(error).toBeNull();
        expect(data!.enabled).toBe(true);
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.post).mockRejectedValue(new Error("Link failed"));

        const { data, error } = await ai.linkKnowledgeBase("bot-1", {
          knowledge_base_id: "kb-1",
        });

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("updateChatbotKnowledgeBase()", () => {
      it("should update link configuration", async () => {
        const response: ChatbotKnowledgeBaseLink = {
          chatbot_id: "bot-1",
          knowledge_base_id: "kb-1",
          max_chunks: 10,
          enabled: true,
        };

        vi.mocked(mockFetch.put).mockResolvedValue(response);

        const { data, error } = await ai.updateChatbotKnowledgeBase(
          "bot-1",
          "kb-1",
          { max_chunks: 10, enabled: true },
        );

        expect(mockFetch.put).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots/bot-1/knowledge-bases/kb-1",
          { max_chunks: 10, enabled: true },
        );
        expect(error).toBeNull();
        expect(data!.max_chunks).toBe(10);
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.put).mockRejectedValue(new Error("Update failed"));

        const { data, error } = await ai.updateChatbotKnowledgeBase(
          "bot-1",
          "kb-1",
          {},
        );

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("unlinkKnowledgeBase()", () => {
      it("should unlink a knowledge base", async () => {
        vi.mocked(mockFetch.delete).mockResolvedValue({});

        const { data, error } = await ai.unlinkKnowledgeBase("bot-1", "kb-1");

        expect(mockFetch.delete).toHaveBeenCalledWith(
          "/api/v1/admin/ai/chatbots/bot-1/knowledge-bases/kb-1",
        );
        expect(error).toBeNull();
        expect(data).toBeNull();
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.delete).mockRejectedValue(
          new Error("Unlink failed"),
        );

        const { data, error } = await ai.unlinkKnowledgeBase("bot-1", "kb-1");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });
  });
});
