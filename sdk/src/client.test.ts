/**
 * Client Tests
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { FluxbaseClient, createClient } from "./client";

// Mock the dependencies
// Note: Vitest 4+ requires constructor functions (not arrow functions) for mocks
vi.mock("./fetch", () => ({
  FluxbaseFetch: vi.fn().mockImplementation(function () {
    return {
      setAnonKey: vi.fn(),
      setAuthToken: vi.fn(),
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    };
  }),
}));

vi.mock("./auth", () => ({
  FluxbaseAuth: vi.fn().mockImplementation(function () {
    return {
      getAccessToken: vi.fn().mockReturnValue("test-token"),
      refreshSession: vi.fn().mockResolvedValue({
        data: { session: { access_token: "new-token" } },
        error: null,
      }),
    };
  }),
}));

vi.mock("./realtime", () => ({
  FluxbaseRealtime: vi.fn().mockImplementation(function () {
    return {
      setAuth: vi.fn(),
      setTokenRefreshCallback: vi.fn(),
      channel: vi.fn().mockReturnValue({}),
      removeChannel: vi.fn().mockResolvedValue("ok"),
    };
  }),
}));

vi.mock("./storage", () => ({
  FluxbaseStorage: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./functions", () => ({
  FluxbaseFunctions: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./jobs", () => ({
  FluxbaseJobs: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./admin", () => ({
  FluxbaseAdmin: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./management", () => ({
  FluxbaseManagement: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./settings", () => ({
  SettingsClient: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./ai", () => ({
  FluxbaseAI: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./vector", () => ({
  FluxbaseVector: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./graphql", () => ({
  FluxbaseGraphQL: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./branching", () => ({
  FluxbaseBranching: vi.fn().mockImplementation(function () {
    return {};
  }),
}));

vi.mock("./rpc", () => ({
  FluxbaseRPC: vi.fn().mockImplementation(function () {
    return {
      invoke: vi
        .fn()
        .mockResolvedValue({ data: { result: "test" }, error: null }),
      list: vi.fn(),
      getStatus: vi.fn(),
      getLogs: vi.fn(),
      waitForCompletion: vi.fn(),
    };
  }),
}));

vi.mock("./query-builder", () => ({
  QueryBuilder: vi.fn().mockImplementation(function () {
    return {
      select: vi.fn().mockReturnThis(),
      execute: vi.fn(),
    };
  }),
}));

vi.mock("./schema-query-builder", () => ({
  SchemaQueryBuilder: vi.fn().mockImplementation(function () {
    return {
      from: vi.fn().mockReturnThis(),
    };
  }),
}));

describe("FluxbaseClient", () => {
  const testUrl = "http://localhost:8080";
  const testKey = "test-anon-key";

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("constructor", () => {
    it("should create client with URL and key", () => {
      const client = new FluxbaseClient(testUrl, testKey);

      expect(client).toBeDefined();
      expect(client.auth).toBeDefined();
      expect(client.realtime).toBeDefined();
      expect(client.storage).toBeDefined();
      expect(client.functions).toBeDefined();
      expect(client.jobs).toBeDefined();
      expect(client.admin).toBeDefined();
      expect(client.management).toBeDefined();
      expect(client.settings).toBeDefined();
      expect(client.ai).toBeDefined();
      expect(client.vector).toBeDefined();
      expect(client.graphql).toBeDefined();
      expect(client.branching).toBeDefined();
      expect(client.rpc).toBeDefined();
    });

    it("should create client with options", () => {
      const client = new FluxbaseClient(testUrl, testKey, {
        timeout: 60000,
        debug: true,
        headers: { "X-Custom": "header" },
      });

      expect(client).toBeDefined();
    });

    it("should set auth token from options", () => {
      const client = new FluxbaseClient(testUrl, testKey, {
        auth: {
          token: "custom-token",
          autoRefresh: false,
          persist: false,
        },
      });

      expect(client).toBeDefined();
    });
  });

  describe("from", () => {
    it("should return a QueryBuilder for a table", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      const builder = client.from("users");

      expect(builder).toBeDefined();
    });
  });

  describe("schema", () => {
    it("should return a SchemaQueryBuilder for a schema", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      const schemaBuilder = client.schema("auth");

      expect(schemaBuilder).toBeDefined();
    });
  });

  describe("getAuthToken", () => {
    it("should return the current access token", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      const token = client.getAuthToken();

      expect(token).toBe("test-token");
    });
  });

  describe("setAuthToken", () => {
    it("should set auth token on fetch and realtime", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      client.setAuthToken("new-token");

      expect(client.realtime.setAuth).toHaveBeenCalledWith("new-token");
    });

    it("should clear auth token when null", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      client.setAuthToken(null);

      expect(client.realtime.setAuth).toHaveBeenCalledWith(null);
    });
  });

  describe("channel", () => {
    it("should create a realtime channel", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      const channel = client.channel("room-1");

      expect(client.realtime.channel).toHaveBeenCalledWith("room-1", undefined);
    });

    it("should create a channel with config", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      const config = { broadcast: { self: true } };
      client.channel("room-1", config);

      expect(client.realtime.channel).toHaveBeenCalledWith("room-1", config);
    });
  });

  describe("removeChannel", () => {
    it("should remove a channel", async () => {
      const client = new FluxbaseClient(testUrl, testKey);
      const mockChannel = {} as any;
      await client.removeChannel(mockChannel);

      expect(client.realtime.removeChannel).toHaveBeenCalledWith(mockChannel);
    });
  });

  describe("http", () => {
    it("should return the internal fetch client", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      const http = client.http;

      expect(http).toBeDefined();
    });
  });

  describe("rpc callable", () => {
    it("should support Supabase-style direct RPC calls", async () => {
      const client = new FluxbaseClient(testUrl, testKey);
      const result = await client.rpc("get_user_orders", { user_id: "123" });

      expect(result.data).toBe("test");
      expect(result.error).toBeNull();
    });

    it("should have access to invoke method", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      expect(client.rpc.invoke).toBeDefined();
    });

    it("should have access to list method", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      expect(client.rpc.list).toBeDefined();
    });

    it("should have access to getStatus method", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      expect(client.rpc.getStatus).toBeDefined();
    });

    it("should have access to getLogs method", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      expect(client.rpc.getLogs).toBeDefined();
    });

    it("should have access to waitForCompletion method", () => {
      const client = new FluxbaseClient(testUrl, testKey);
      expect(client.rpc.waitForCompletion).toBeDefined();
    });
  });
});

describe("createClient", () => {
  const originalEnv = process.env;

  beforeEach(() => {
    vi.clearAllMocks();
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it("should create client with URL and key arguments", () => {
    const client = createClient("http://localhost:8080", "test-key");
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should create client with options", () => {
    const client = createClient("http://localhost:8080", "test-key", {
      timeout: 60000,
    });
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should read URL from FLUXBASE_URL env var", () => {
    process.env.FLUXBASE_URL = "http://env-url:8080";
    process.env.FLUXBASE_ANON_KEY = "env-key";

    const client = createClient();
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should read URL from NEXT_PUBLIC_FLUXBASE_URL env var", () => {
    process.env.NEXT_PUBLIC_FLUXBASE_URL = "http://next-url:8080";
    process.env.FLUXBASE_ANON_KEY = "env-key";

    const client = createClient();
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should read URL from VITE_FLUXBASE_URL env var", () => {
    process.env.VITE_FLUXBASE_URL = "http://vite-url:8080";
    process.env.FLUXBASE_ANON_KEY = "env-key";

    const client = createClient();
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should read key from FLUXBASE_ANON_KEY env var", () => {
    process.env.FLUXBASE_URL = "http://localhost:8080";
    process.env.FLUXBASE_ANON_KEY = "anon-key";

    const client = createClient();
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should read key from FLUXBASE_SERVICE_TOKEN env var", () => {
    process.env.FLUXBASE_URL = "http://localhost:8080";
    process.env.FLUXBASE_SERVICE_TOKEN = "service-token";

    const client = createClient();
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should read key from FLUXBASE_JOB_TOKEN env var", () => {
    process.env.FLUXBASE_URL = "http://localhost:8080";
    process.env.FLUXBASE_JOB_TOKEN = "job-token";

    const client = createClient();
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should read key from NEXT_PUBLIC_FLUXBASE_ANON_KEY env var", () => {
    process.env.FLUXBASE_URL = "http://localhost:8080";
    process.env.NEXT_PUBLIC_FLUXBASE_ANON_KEY = "next-anon-key";

    const client = createClient();
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should read key from VITE_FLUXBASE_ANON_KEY env var", () => {
    process.env.FLUXBASE_URL = "http://localhost:8080";
    process.env.VITE_FLUXBASE_ANON_KEY = "vite-anon-key";

    const client = createClient();
    expect(client).toBeInstanceOf(FluxbaseClient);
  });

  it("should throw error when URL is missing", () => {
    delete process.env.FLUXBASE_URL;
    delete process.env.NEXT_PUBLIC_FLUXBASE_URL;
    delete process.env.VITE_FLUXBASE_URL;

    expect(() => createClient(undefined, "test-key")).toThrow(
      "Fluxbase URL is required",
    );
  });

  it("should throw error when key is missing", () => {
    delete process.env.FLUXBASE_ANON_KEY;
    delete process.env.FLUXBASE_SERVICE_TOKEN;
    delete process.env.FLUXBASE_JOB_TOKEN;
    delete process.env.NEXT_PUBLIC_FLUXBASE_ANON_KEY;
    delete process.env.VITE_FLUXBASE_ANON_KEY;

    expect(() => createClient("http://localhost:8080")).toThrow(
      "Fluxbase key is required",
    );
  });

  it("should support generic database type", () => {
    interface MyDatabase {
      public: {
        users: { id: string; name: string };
      };
    }

    const client = createClient<MyDatabase>(
      "http://localhost:8080",
      "test-key",
    );
    expect(client).toBeInstanceOf(FluxbaseClient);
  });
});
