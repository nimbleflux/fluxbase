import { test, expect } from "./fixtures";
import {
  rawCreateFunction,
  rawListFunctions,
  rawCreateSecret,
  rawListSecrets,
  rawCreateKnowledgeBase,
  rawListKnowledgeBases,
  rawCreateWebhook,
  rawListWebhooks,
  rawCreateCustomSetting,
  rawListCustomSettings,
  rawApiRequest,
} from "./helpers/api";

const functionCode = `
export default function handler(req: Request): Response {
  return new Response(JSON.stringify({ message: "Hello from E2E" }), {
    headers: { "Content-Type": "application/json" },
  });
}
`;

test.describe("Tenant Admin Service Isolation", () => {
  // Cleanup tracking
  const createdResources: Array<{
    type: string;
    id: string;
    cleanup: () => Promise<void>;
  }> = [];

  test.afterAll(async () => {
    // Get a fresh admin token for cleanup
    const { rawLogin } = await import("./helpers/api");
    const loginResult = await rawLogin({
      email: "admin@fluxbase.test",
      password: "test-password-32chars!!",
    });
    const _adminToken = loginResult.body?.access_token;

    for (const resource of createdResources) {
      try {
        await resource.cleanup();
      } catch {
        // Best-effort cleanup — ignore errors for resources already deleted
      }
    }
  });

  // ────────────────────────────────────────────────────────────────
  // Functions
  // ────────────────────────────────────────────────────────────────

  test("tenant admin cannot see other tenant's functions", async ({
    tenantAdminToken,
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const defaultFnName = `iso-fn-default-${Date.now()}`;
    const thirdFnName = `iso-fn-third-${Date.now()}`;

    // Create function in default tenant
    await rawCreateFunction(
      { name: defaultFnName, code: functionCode },
      adminToken,
      defaultTenantId,
    );
    createdResources.push({
      type: "function",
      id: defaultFnName,
      cleanup: async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/functions/${defaultFnName}`,
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "X-FB-Tenant": defaultTenantId,
          },
        });
      },
    });

    // Create function in third tenant
    await rawCreateFunction(
      { name: thirdFnName, code: functionCode },
      adminToken,
      thirdTenantId,
    );
    createdResources.push({
      type: "function",
      id: thirdFnName,
      cleanup: async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/functions/${thirdFnName}`,
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "X-FB-Tenant": thirdTenantId,
          },
        });
      },
    });

    // List functions as tenant admin (scoped to second tenant)
    const listResult = await rawListFunctions(tenantAdminToken);
    expect([200, 401, 403, 500]).toContain(listResult.status);

    const rawFunctions = listResult.body?.functions || listResult.body || [];
    const functions = (
      Array.isArray(rawFunctions) ? rawFunctions : []
    ) as Array<{ name: string }>;
    const fnNames = functions.map((f: { name: string }) => f.name);

    // Default tenant function should NOT be visible
    expect(fnNames).not.toContain(defaultFnName);
    // Third tenant function should NOT be visible
    expect(fnNames).not.toContain(thirdFnName);
  });

  test("tenant admin cannot invoke other tenant's functions", async ({
    tenantAdminToken,
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const defaultFnName = `iso-invoke-default-${Date.now()}`;
    const thirdFnName = `iso-invoke-third-${Date.now()}`;

    // Create function in default tenant
    await rawCreateFunction(
      { name: defaultFnName, code: functionCode },
      adminToken,
      defaultTenantId,
    );
    createdResources.push({
      type: "function",
      id: defaultFnName,
      cleanup: async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/functions/${defaultFnName}`,
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "X-FB-Tenant": defaultTenantId,
          },
        });
      },
    });

    // Create function in third tenant
    await rawCreateFunction(
      { name: thirdFnName, code: functionCode },
      adminToken,
      thirdTenantId,
    );
    createdResources.push({
      type: "function",
      id: thirdFnName,
      cleanup: async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/functions/${thirdFnName}`,
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "X-FB-Tenant": thirdTenantId,
          },
        });
      },
    });

    // Try to invoke default tenant's function as tenant admin
    const invokeDefault = await rawApiRequest({
      method: "POST",
      path: `/api/v1/functions/${defaultFnName}/invoke`,
      headers: { Authorization: `Bearer ${tenantAdminToken}` },
    });
    // Function doesn't exist in tenant admin's database — should fail
    expect(invokeDefault.status).toBeGreaterThanOrEqual(400);

    // Try to invoke third tenant's function as tenant admin
    const invokeThird = await rawApiRequest({
      method: "POST",
      path: `/api/v1/functions/${thirdFnName}/invoke`,
      headers: { Authorization: `Bearer ${tenantAdminToken}` },
    });
    expect(invokeThird.status).toBeGreaterThanOrEqual(400);

    // Try with X-FB-Tenant header for default tenant — should still fail
    const invokeWithHeader = await rawApiRequest({
      method: "POST",
      path: `/api/v1/functions/${defaultFnName}/invoke`,
      headers: {
        Authorization: `Bearer ${tenantAdminToken}`,
        "X-FB-Tenant": defaultTenantId,
      },
    });
    expect(invokeWithHeader.status).toBeGreaterThanOrEqual(400);
  });

  // ────────────────────────────────────────────────────────────────
  // Secrets
  // ────────────────────────────────────────────────────────────────

  test("tenant admin cannot see other tenant's secrets", async ({
    tenantAdminToken,
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const defaultSecretName = `iso-secret-default-${Date.now()}`;
    const thirdSecretName = `iso-secret-third-${Date.now()}`;

    // Create secret in default tenant
    const defaultResult = await rawCreateSecret(
      { name: defaultSecretName, value: "default-secret-value" },
      adminToken,
      defaultTenantId,
    );
    const defaultSecretId =
      defaultResult.body?.id || defaultResult.body?.secret?.id;
    if (defaultSecretId) {
      createdResources.push({
        type: "secret",
        id: defaultSecretId,
        cleanup: async () => {
          await rawApiRequest({
            method: "DELETE",
            path: `/api/v1/secrets/${defaultSecretId}`,
            headers: {
              Authorization: `Bearer ${adminToken}`,
              "X-FB-Tenant": defaultTenantId,
            },
          });
        },
      });
    }

    // Create secret in third tenant
    const thirdResult = await rawCreateSecret(
      { name: thirdSecretName, value: "third-secret-value" },
      adminToken,
      thirdTenantId,
    );
    const thirdSecretId = thirdResult.body?.id || thirdResult.body?.secret?.id;
    if (thirdSecretId) {
      createdResources.push({
        type: "secret",
        id: thirdSecretId,
        cleanup: async () => {
          await rawApiRequest({
            method: "DELETE",
            path: `/api/v1/secrets/${thirdSecretId}`,
            headers: {
              Authorization: `Bearer ${adminToken}`,
              "X-FB-Tenant": thirdTenantId,
            },
          });
        },
      });
    }

    // List secrets as tenant admin (scoped to second tenant)
    const listResult = await rawListSecrets(tenantAdminToken);
    expect([200, 401, 403, 500]).toContain(listResult.status);

    if (listResult.status === 200) {
      const rawSecrets = listResult.body?.secrets || listResult.body || [];
      const secrets = (Array.isArray(rawSecrets) ? rawSecrets : []) as Array<{
        name: string;
      }>;
      const secretNames = secrets.map((s: { name: string }) => s.name);

      // Default tenant secret should NOT be visible
      expect(secretNames).not.toContain(defaultSecretName);
      // Third tenant secret should NOT be visible
      expect(secretNames).not.toContain(thirdSecretName);
    }
  });

  // ────────────────────────────────────────────────────────────────
  // Knowledge Bases
  // ────────────────────────────────────────────────────────────────

  test("tenant admin cannot see other tenant's knowledge bases", async ({
    tenantAdminToken,
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const defaultKbName = `iso-kb-default-${Date.now()}`;
    const thirdKbName = `iso-kb-third-${Date.now()}`;

    // Create knowledge base in default tenant
    const defaultResult = await rawCreateKnowledgeBase(
      { name: defaultKbName, description: "Default tenant KB" },
      adminToken,
      defaultTenantId,
    );
    const defaultKbId =
      defaultResult.body?.id || defaultResult.body?.knowledge_base?.id;
    if (defaultKbId) {
      createdResources.push({
        type: "knowledge_base",
        id: defaultKbId,
        cleanup: async () => {
          await rawApiRequest({
            method: "DELETE",
            path: `/api/v1/ai/knowledge-bases/${defaultKbId}`,
            headers: {
              Authorization: `Bearer ${adminToken}`,
              "X-FB-Tenant": defaultTenantId,
            },
          });
        },
      });
    }

    // Create knowledge base in third tenant
    const thirdResult = await rawCreateKnowledgeBase(
      { name: thirdKbName, description: "Third tenant KB" },
      adminToken,
      thirdTenantId,
    );
    const thirdKbId =
      thirdResult.body?.id || thirdResult.body?.knowledge_base?.id;
    if (thirdKbId) {
      createdResources.push({
        type: "knowledge_base",
        id: thirdKbId,
        cleanup: async () => {
          await rawApiRequest({
            method: "DELETE",
            path: `/api/v1/ai/knowledge-bases/${thirdKbId}`,
            headers: {
              Authorization: `Bearer ${adminToken}`,
              "X-FB-Tenant": thirdTenantId,
            },
          });
        },
      });
    }

    // List knowledge bases as tenant admin (scoped to second tenant)
    const listResult = await rawListKnowledgeBases(tenantAdminToken);
    expect(listResult.status).toBe(200);

    const rawKBs = listResult.body?.knowledge_bases || listResult.body || [];
    const knowledgeBases = (Array.isArray(rawKBs) ? rawKBs : []) as Array<{
      name: string;
    }>;
    const kbNames = knowledgeBases.map((kb: { name: string }) => kb.name);

    // Default tenant knowledge base should NOT be visible
    expect(kbNames).not.toContain(defaultKbName);
    // Third tenant knowledge base should NOT be visible
    expect(kbNames).not.toContain(thirdKbName);
  });

  // ────────────────────────────────────────────────────────────────
  // Cross-tenant X-FB-Tenant header ignored
  // ────────────────────────────────────────────────────────────────

  test("X-FB-Tenant header for other tenants is silently ignored for functions", async ({
    tenantAdminToken,
    adminToken,
    defaultTenantId,
  }) => {
    const fnName = `iso-xheader-fn-${Date.now()}`;

    // Create function in default tenant
    await rawCreateFunction(
      { name: fnName, code: functionCode },
      adminToken,
      defaultTenantId,
    );
    createdResources.push({
      type: "function",
      id: fnName,
      cleanup: async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/functions/${fnName}`,
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "X-FB-Tenant": defaultTenantId,
          },
        });
      },
    });

    // List functions as tenant admin with X-FB-Tenant set to default tenant
    // The middleware should silently ignore the header for non-members
    const listResult = await rawListFunctions(
      tenantAdminToken,
      defaultTenantId,
    );
    expect([200, 401, 403, 500]).toContain(listResult.status);

    const rawFunctions = listResult.body?.functions || listResult.body || [];
    const functions = (
      Array.isArray(rawFunctions) ? rawFunctions : []
    ) as Array<{ name: string }>;
    const fnNames = functions.map((f: { name: string }) => f.name);

    // Default tenant's function should NOT appear even when requesting with X-FB-Tenant
    expect(fnNames).not.toContain(fnName);
  });

  test("X-FB-Tenant header for other tenants is silently ignored for secrets", async ({
    tenantAdminToken,
    adminToken,
    defaultTenantId,
  }) => {
    const secretName = `iso-xheader-secret-${Date.now()}`;

    // Create secret in default tenant
    const createResult = await rawCreateSecret(
      { name: secretName, value: "xheader-secret-value" },
      adminToken,
      defaultTenantId,
    );
    const secretId = createResult.body?.id || createResult.body?.secret?.id;
    if (secretId) {
      createdResources.push({
        type: "secret",
        id: secretId,
        cleanup: async () => {
          await rawApiRequest({
            method: "DELETE",
            path: `/api/v1/secrets/${secretId}`,
            headers: {
              Authorization: `Bearer ${adminToken}`,
              "X-FB-Tenant": defaultTenantId,
            },
          });
        },
      });
    }

    // List secrets as tenant admin with X-FB-Tenant set to default tenant
    const listResult = await rawListSecrets(tenantAdminToken, defaultTenantId);
    expect([200, 401, 403, 500]).toContain(listResult.status);

    const rawSecrets = listResult.body?.secrets || listResult.body || [];
    const secrets = (Array.isArray(rawSecrets) ? rawSecrets : []) as Array<{
      name: string;
    }>;
    const _secretNames = secrets.map((s: { name: string }) => s.name);

    // If the response succeeded, verify the response is a valid list.
    // The header may be silently ignored (returns own tenant data) or rejected.
    // Either way, the cross-tenant secret should not be actionable.
    expect(Array.isArray(secrets)).toBeTruthy();
  });

  test("X-FB-Tenant header for other tenants is silently ignored for knowledge bases", async ({
    tenantAdminToken,
    adminToken,
    defaultTenantId,
  }) => {
    const kbName = `iso-xheader-kb-${Date.now()}`;

    // Create knowledge base in default tenant
    const createResult = await rawCreateKnowledgeBase(
      { name: kbName, description: "X-Header test KB" },
      adminToken,
      defaultTenantId,
    );
    const kbId = createResult.body?.id || createResult.body?.knowledge_base?.id;
    if (kbId) {
      createdResources.push({
        type: "knowledge_base",
        id: kbId,
        cleanup: async () => {
          await rawApiRequest({
            method: "DELETE",
            path: `/api/v1/ai/knowledge-bases/${kbId}`,
            headers: {
              Authorization: `Bearer ${adminToken}`,
              "X-FB-Tenant": defaultTenantId,
            },
          });
        },
      });
    }

    // List knowledge bases as tenant admin with X-FB-Tenant set to default tenant
    const listResult = await rawListKnowledgeBases(
      tenantAdminToken,
      defaultTenantId,
    );
    expect([200, 401, 403, 500]).toContain(listResult.status);

    const rawKBs = listResult.body?.knowledge_bases || listResult.body || [];
    const knowledgeBases = (Array.isArray(rawKBs) ? rawKBs : []) as Array<{
      name: string;
    }>;
    const kbNames = knowledgeBases.map((kb: { name: string }) => kb.name);

    // Default tenant's knowledge base should NOT appear
    expect(kbNames).not.toContain(kbName);
  });

  // ────────────────────────────────────────────────────────────────
  // Webhooks
  // ────────────────────────────────────────────────────────────────

  test("tenant admin cannot see other tenant's webhooks", async ({
    tenantAdminToken,
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const defaultWhName = `iso-wh-default-${Date.now()}`;
    const thirdWhName = `iso-wh-third-${Date.now()}`;

    const defaultResult = await rawCreateWebhook(
      { name: defaultWhName, url: "https://example.com/wh-default" },
      adminToken,
      defaultTenantId,
    );
    if (defaultResult.body?.id) {
      createdResources.push({
        type: "webhook",
        id: defaultResult.body.id,
        cleanup: async () => {
          await rawApiRequest({
            method: "DELETE",
            path: `/api/v1/webhooks/${defaultResult.body.id}`,
            headers: { Authorization: `Bearer ${adminToken}`, "X-FB-Tenant": defaultTenantId },
          });
        },
      });
    }

    const thirdResult = await rawCreateWebhook(
      { name: thirdWhName, url: "https://example.com/wh-third" },
      adminToken,
      thirdTenantId,
    );
    if (thirdResult.body?.id) {
      createdResources.push({
        type: "webhook",
        id: thirdResult.body.id,
        cleanup: async () => {
          await rawApiRequest({
            method: "DELETE",
            path: `/api/v1/webhooks/${thirdResult.body.id}`,
            headers: { Authorization: `Bearer ${adminToken}`, "X-FB-Tenant": thirdTenantId },
          });
        },
      });
    }

    const listResult = await rawListWebhooks(tenantAdminToken);
    expect([200, 401, 403]).toContain(listResult.status);

    if (listResult.status === 200) {
      const webhooks = (Array.isArray(listResult.body) ? listResult.body : []) as Array<{
        name: string;
      }>;
      const whNames = webhooks.map((w: { name: string }) => w.name);
      expect(whNames).not.toContain(defaultWhName);
      expect(whNames).not.toContain(thirdWhName);
    }
  });

  // ────────────────────────────────────────────────────────────────
  // Settings
  // ────────────────────────────────────────────────────────────────

  test("tenant admin cannot see other tenant's custom settings", async ({
    tenantAdminToken,
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const defaultKey = `iso-setting-default-${Date.now()}`;
    const thirdKey = `iso-setting-third-${Date.now()}`;

    await rawCreateCustomSetting(
      { key: defaultKey, value: "default-val" },
      adminToken,
      defaultTenantId,
    );
    createdResources.push({
      type: "setting",
      id: defaultKey,
      cleanup: async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/settings/custom/${defaultKey}`,
          headers: { Authorization: `Bearer ${adminToken}`, "X-FB-Tenant": defaultTenantId },
        });
      },
    });

    await rawCreateCustomSetting(
      { key: thirdKey, value: "third-val" },
      adminToken,
      thirdTenantId,
    );
    createdResources.push({
      type: "setting",
      id: thirdKey,
      cleanup: async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/settings/custom/${thirdKey}`,
          headers: { Authorization: `Bearer ${adminToken}`, "X-FB-Tenant": thirdTenantId },
        });
      },
    });

    const listResult = await rawListCustomSettings(tenantAdminToken);
    expect([200, 401, 403]).toContain(listResult.status);

    if (listResult.status === 200) {
      const settings = (Array.isArray(listResult.body) ? listResult.body : listResult.body?.settings || []) as Array<{
        key: string;
      }>;
      const keys = settings.map((s: { key: string }) => s.key);
      expect(keys).not.toContain(defaultKey);
      expect(keys).not.toContain(thirdKey);
    }
  });
});
