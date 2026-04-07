import { test, expect } from "./fixtures";
import {
  rawCreateFunction,
  rawListFunctions,
  rawCreateSecret,
  rawListSecrets,
  rawCreateKnowledgeBase,
  rawListKnowledgeBases,
  rawApiRequest,
} from "./helpers/api";

test.describe("Tenant Service Isolation", () => {
  // Cleanup tracking
  const createdResources: Array<{
    type: string;
    id: string;
    token: string;
    tenantId?: string;
  }> = [];

  test.afterAll(async () => {
    const { rawLogin } = await import("./helpers/api");
    const loginResult = await rawLogin({
      email: "admin@fluxbase.test",
      password: "test-password-32chars!!",
    });
    const adminToken = loginResult.body?.access_token;

    for (const { type, id, token, tenantId } of createdResources) {
      const headers: Record<string, string> = {
        Authorization: `Bearer ${token}`,
      };
      if (tenantId) headers["X-FB-Tenant"] = tenantId;

      let path: string;
      switch (type) {
        case "function":
          path = `/api/v1/functions/${id}`;
          break;
        case "secret":
          path = `/api/v1/secrets/${id}`;
          break;
        case "knowledge_base":
          path = `/api/v1/ai/knowledge-bases/${id}`;
          break;
        default:
          continue;
      }

      await rawApiRequest({ method: "DELETE", path, headers }).catch(() => {});
    }
  });

  // ────────────────────────────────────────────────────────────────
  // Group 1: Functions Isolation
  // ────────────────────────────────────────────────────────────────

  test("functions created in tenant A are not visible in tenant B", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const suffix = Date.now();
    const funcName = `iso-func-A-${suffix}`;

    const code = `
export default function handler(req: Request): Response {
  return new Response(JSON.stringify({ message: "Hello from E2E" }), {
    headers: { "Content-Type": "application/json" },
  });
}
`;

    // Create function in default tenant (tenant A)
    const createResult = await rawCreateFunction(
      { name: funcName, code, verifyJWT: false },
      adminToken,
      defaultTenantId,
    );
    expect([200, 201]).toContain(createResult.status);
    if (createResult.body?.name) {
      createdResources.push({
        type: "function",
        id: funcName,
        token: adminToken,
        tenantId: defaultTenantId,
      });
    }

    // List functions from third tenant context (tenant B)
    const listResult = await rawListFunctions(adminToken, thirdTenantId);
    expect(listResult.status).toBe(200);
    const functions = (listResult.body || []) as Array<{ name: string }>;
    const funcNames = functions.map((f: { name: string }) => f.name);

    // Function from tenant A should NOT appear in tenant B's list
    expect(funcNames).not.toContain(funcName);
  });

  test("functions cannot be invoked from other tenant", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const suffix = Date.now();
    const funcName = `iso-func-invoke-${suffix}`;

    const code = `
export default function handler(req: Request): Response {
  return new Response(JSON.stringify({ message: "Hello from E2E" }), {
    headers: { "Content-Type": "application/json" },
  });
}
`;

    // Create function in default tenant (tenant A)
    const createResult = await rawCreateFunction(
      { name: funcName, code, verifyJWT: false },
      adminToken,
      defaultTenantId,
    );
    expect([200, 201]).toContain(createResult.status);
    if (createResult.body?.name) {
      createdResources.push({
        type: "function",
        id: funcName,
        token: adminToken,
        tenantId: defaultTenantId,
      });
    }

    // Try to invoke the function from third tenant context (tenant B)
    const invokeResult = await rawApiRequest({
      method: "POST",
      path: `/api/v1/functions/${funcName}/invoke`,
      headers: {
        Authorization: `Bearer ${adminToken}`,
        "X-FB-Tenant": thirdTenantId,
      },
    });

    // Function should not be found in tenant B's context
    expect(invokeResult.status).toBeGreaterThanOrEqual(400);
  });

  test("functions cannot be deleted from other tenant", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const suffix = Date.now();
    const funcName = `iso-func-del-${suffix}`;

    const code = `
export default function handler(req: Request): Response {
  return new Response(JSON.stringify({ message: "Hello from E2E" }), {
    headers: { "Content-Type": "application/json" },
  });
}
`;

    // Create function in default tenant (tenant A)
    const createResult = await rawCreateFunction(
      { name: funcName, code, verifyJWT: false },
      adminToken,
      defaultTenantId,
    );
    expect([200, 201]).toContain(createResult.status);
    if (createResult.body?.name) {
      createdResources.push({
        type: "function",
        id: funcName,
        token: adminToken,
        tenantId: defaultTenantId,
      });
    }

    // Try to delete from third tenant context (tenant B)
    const deleteResult = await rawDeleteFunction(
      funcName,
      adminToken,
      thirdTenantId,
    );

    // Delete should fail — function doesn't exist in tenant B
    expect(deleteResult.status).toBeGreaterThanOrEqual(400);

    // Verify function still exists in tenant A
    const listResult = await rawListFunctions(adminToken, defaultTenantId);
    expect(listResult.status).toBe(200);
    const functions = (listResult.body || []) as Array<{ name: string }>;
    const funcNames = functions.map((f: { name: string }) => f.name);
    expect(funcNames).toContain(funcName);
  });

  // ────────────────────────────────────────────────────────────────
  // Group 2: Secrets Isolation
  // ────────────────────────────────────────────────────────────────

  test("secrets created in tenant A are not visible in tenant B", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const suffix = Date.now();
    const secretName = `iso-secret-A-${suffix}`;

    // Create secret in default tenant (tenant A)
    const createResult = await rawCreateSecret(
      { name: secretName, value: "secret-value-123" },
      adminToken,
      defaultTenantId,
    );
    expect([200, 201]).toContain(createResult.status);
    const secretId = createResult.body?.id;
    if (secretId) {
      createdResources.push({
        type: "secret",
        id: secretId,
        token: adminToken,
        tenantId: defaultTenantId,
      });
    }

    // List secrets from third tenant context (tenant B)
    const listResult = await rawListSecrets(adminToken, thirdTenantId);
    expect(listResult.status).toBe(200);
    const secrets = (listResult.body || []) as Array<{ name: string }>;
    const secretNames = secrets.map((s: { name: string }) => s.name);

    // Secret from tenant A should NOT appear in tenant B's list
    expect(secretNames).not.toContain(secretName);
  });

  test("secrets cannot be accessed from other tenant", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const suffix = Date.now();
    const secretName = `iso-secret-access-${suffix}`;

    // Create secret in default tenant (tenant A)
    const createResult = await rawCreateSecret(
      { name: secretName, value: "secret-value-456" },
      adminToken,
      defaultTenantId,
    );
    expect([200, 201]).toContain(createResult.status);
    const secretId = createResult.body?.id;
    if (secretId) {
      createdResources.push({
        type: "secret",
        id: secretId,
        token: adminToken,
        tenantId: defaultTenantId,
      });
    }

    // Try to access the secret from third tenant context (tenant B)
    if (secretId) {
      const accessResult = await rawApiRequest({
        method: "GET",
        path: `/api/v1/secrets/${secretId}`,
        headers: {
          Authorization: `Bearer ${adminToken}`,
          "X-FB-Tenant": thirdTenantId,
        },
      });

      // Secret should not be found in tenant B's context
      expect(accessResult.status).toBeGreaterThanOrEqual(400);
    }
  });

  // ────────────────────────────────────────────────────────────────
  // Group 3: Knowledge Bases Isolation
  // ────────────────────────────────────────────────────────────────

  test("knowledge bases created in tenant A are not visible in tenant B", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const suffix = Date.now();
    const kbName = `iso-kb-A-${suffix}`;

    // Create knowledge base in default tenant (tenant A)
    const createResult = await rawCreateKnowledgeBase(
      { name: kbName, description: "Isolation test KB" },
      adminToken,
      defaultTenantId,
    );
    expect([200, 201]).toContain(createResult.status);
    const kbId = createResult.body?.id;
    if (kbId) {
      createdResources.push({
        type: "knowledge_base",
        id: kbId,
        token: adminToken,
        tenantId: defaultTenantId,
      });
    }

    // List knowledge bases from third tenant context (tenant B)
    const listResult = await rawListKnowledgeBases(adminToken, thirdTenantId);
    expect(listResult.status).toBe(200);
    const knowledgeBases = (listResult.body || []) as Array<{
      name: string;
    }>;
    const kbNames = knowledgeBases.map((kb: { name: string }) => kb.name);

    // Knowledge base from tenant A should NOT appear in tenant B's list
    expect(kbNames).not.toContain(kbName);
  });

  test("knowledge bases cannot be accessed from other tenant", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const suffix = Date.now();
    const kbName = `iso-kb-access-${suffix}`;

    // Create knowledge base in default tenant (tenant A)
    const createResult = await rawCreateKnowledgeBase(
      { name: kbName, description: "Isolation test KB" },
      adminToken,
      defaultTenantId,
    );
    expect([200, 201]).toContain(createResult.status);
    const kbId = createResult.body?.id;
    if (kbId) {
      createdResources.push({
        type: "knowledge_base",
        id: kbId,
        token: adminToken,
        tenantId: defaultTenantId,
      });
    }

    // Try to access the knowledge base from third tenant context (tenant B)
    if (kbId) {
      const accessResult = await rawApiRequest({
        method: "GET",
        path: `/api/v1/ai/knowledge-bases/${kbId}`,
        headers: {
          Authorization: `Bearer ${adminToken}`,
          "X-FB-Tenant": thirdTenantId,
        },
      });

      // Knowledge base should not be found in tenant B's context
      expect(accessResult.status).toBeGreaterThanOrEqual(400);
    }
  });
});
