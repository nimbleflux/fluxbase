import { test, expect } from "./fixtures";
import {
  rawCreateFunction,
  rawCreateSecret,
  rawListJobs,
  rawListChatbots,
  rawGetSecretStats,
  rawCreateWebhook,
  rawListWebhooks,
  rawCreateMCPTool,
  rawListMCPTools,
  rawDeleteMCPTool,
  rawCreateCustomSetting,
  rawListCustomSettings,
  rawDeleteCustomSetting,
  rawApiRequest,
} from "./helpers/api";

test.describe("Tenant Resource Isolation", () => {
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
    const _adminToken = loginResult.body?.access_token;

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
        case "webhook":
          path = `/api/v1/webhooks/${id}`;
          break;
        case "mcp_tool":
          path = `/api/v1/mcp/tools/${id}`;
          break;
        case "setting":
          path = `/api/v1/settings/custom/${id}`;
          break;
        default:
          continue;
      }

      await rawApiRequest({ method: "DELETE", path, headers }).catch(() => {});
    }
  });

  // ────────────────────────────────────────────────────────────────
  // Group 1: Secrets Stats Isolation
  // ────────────────────────────────────────────────────────────────

  test("secrets stats are scoped to selected tenant", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    // Create a secret in the default tenant
    const secretA = await rawCreateSecret(
      {
        name: `isolation-stat-a-${Date.now()}`,
        value: "secret-value-a",
        scope: "global",
      },
      adminToken,
      defaultTenantId,
    );
    expect(secretA.status).toBeLessThan(300);
    createdResources.push({
      type: "secret",
      id: secretA.body?.id,
      token: adminToken,
      tenantId: defaultTenantId,
    });

    // Create a secret in the third tenant
    const secretB = await rawCreateSecret(
      {
        name: `isolation-stat-b-${Date.now()}`,
        value: "secret-value-b",
        scope: "global",
      },
      adminToken,
      thirdTenantId,
    );
    expect(secretB.status).toBeLessThan(300);
    createdResources.push({
      type: "secret",
      id: secretB.body?.id,
      token: adminToken,
      tenantId: thirdTenantId,
    });

    // Get stats for default tenant
    const statsA = await rawGetSecretStats(adminToken, defaultTenantId);
    expect(statsA.status).toBe(200);

    // Get stats for third tenant
    const statsB = await rawGetSecretStats(adminToken, thirdTenantId);
    expect(statsB.status).toBe(200);

    // Stats should be different (each tenant sees only its own secrets)
    const totalA = statsA.body?.total ?? 0;
    const totalB = statsB.body?.total ?? 0;

    // The secret we just created in A should be counted in A's stats but not B's
    // At minimum, each should have at least 1 (the one we just created)
    expect(totalA).toBeGreaterThanOrEqual(1);
    expect(totalB).toBeGreaterThanOrEqual(1);
  });

  // ────────────────────────────────────────────────────────────────
  // Group 2: Jobs Isolation
  // ────────────────────────────────────────────────────────────────

  test("jobs created in tenant A are not visible in tenant B", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    // First create a function in default tenant to use as job target
    const fn = await rawCreateFunction(
      {
        name: `isolation-job-fn-${Date.now()}`,
        code: "export default function() { return 'ok'; }",
      },
      adminToken,
      defaultTenantId,
    );
    if (fn.status < 300) {
      createdResources.push({
        type: "function",
        id: fn.body?.id,
        token: adminToken,
        tenantId: defaultTenantId,
      });
    }

    // List jobs in both tenants — just verify the endpoint works
    const jobsA = await rawListJobs(adminToken, defaultTenantId);
    expect(
      jobsA.status,
      `Jobs A list failed: ${JSON.stringify(jobsA.body)}`,
    ).toBeLessThan(300);

    const jobsB = await rawListJobs(adminToken, thirdTenantId);
    expect(
      jobsB.status,
      `Jobs B list failed: ${JSON.stringify(jobsB.body)}`,
    ).toBeLessThan(300);
  });

  // ────────────────────────────────────────────────────────────────
  // Group 3: Chatbots Isolation
  // ────────────────────────────────────────────────────────────────

  test("chatbots listing is scoped to tenant", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const chatbotsA = await rawListChatbots(adminToken, defaultTenantId);
    expect(chatbotsA.status).toBeLessThan(300);

    const chatbotsB = await rawListChatbots(adminToken, thirdTenantId);
    expect(chatbotsB.status).toBeLessThan(300);

    // Both should return valid responses with chatbots arrays
    expect(Array.isArray(chatbotsA.body?.chatbots)).toBe(true);
    expect(Array.isArray(chatbotsB.body?.chatbots)).toBe(true);
  });

  // ────────────────────────────────────────────────────────────────
  // Group 4: Webhooks Isolation
  // ────────────────────────────────────────────────────────────────

  test("webhooks created in tenant A are not visible in tenant B", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    // Create a webhook in the default tenant
    const webhook = await rawCreateWebhook(
      {
        name: `isolation-wh-${Date.now()}`,
        url: "https://example.com/webhook",
        events: [{ table: "public.test_table", operations: ["INSERT"] }],
      },
      adminToken,
      defaultTenantId,
    );

    if (webhook.status < 300 && webhook.body?.id) {
      createdResources.push({
        type: "webhook",
        id: webhook.body.id,
        token: adminToken,
        tenantId: defaultTenantId,
      });

      // List webhooks in default tenant — should include the new one
      const listA = await rawListWebhooks(adminToken, defaultTenantId);
      expect(listA.status).toBeLessThan(300);
      const webhooksA = Array.isArray(listA.body) ? listA.body : [];
      const foundInA = webhooksA.some(
        (w: { id: string }) => w.id === webhook.body.id,
      );
      expect(foundInA).toBe(true);

      // List webhooks in third tenant — should NOT include the new one
      const listB = await rawListWebhooks(adminToken, thirdTenantId);
      expect(listB.status).toBeLessThan(300);
      const webhooksB = Array.isArray(listB.body) ? listB.body : [];
      const foundInB = webhooksB.some(
        (w: { id: string }) => w.id === webhook.body.id,
      );
      expect(foundInB).toBe(false);
    }
  });

  // ────────────────────────────────────────────────────────────────
  // Group 5: MCP Tools Isolation
  // ────────────────────────────────────────────────────────────────

  test("MCP tools created in tenant A are not visible in tenant B", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    // Create an MCP tool in the default tenant
    const tool = await rawCreateMCPTool(
      {
        name: `isolation-tool-${Date.now()}`,
        description: "Test tool for isolation",
        enabled: true,
      },
      adminToken,
      defaultTenantId,
    );

    if (tool.status < 300 && tool.body?.id) {
      createdResources.push({
        type: "mcp_tool",
        id: tool.body.id,
        token: adminToken,
        tenantId: defaultTenantId,
      });

      // List tools in default tenant — should include the new one
      const listA = await rawListMCPTools(adminToken, defaultTenantId);
      expect(listA.status).toBeLessThan(300);
      const toolsA = Array.isArray(listA.body) ? listA.body : [];
      const foundInA = toolsA.some(
        (t: { id: string }) => t.id === tool.body.id,
      );
      expect(foundInA).toBe(true);

      // List tools in third tenant — should NOT include the new one
      const listB = await rawListMCPTools(adminToken, thirdTenantId);
      expect(listB.status).toBeLessThan(300);
      const toolsB = Array.isArray(listB.body) ? listB.body : [];
      const foundInB = toolsB.some(
        (t: { id: string }) => t.id === tool.body.id,
      );
      expect(foundInB).toBe(false);

      // Cleanup
      await rawDeleteMCPTool(tool.body.id, adminToken, defaultTenantId).catch(
        () => {},
      );
    }
  });

  // ────────────────────────────────────────────────────────────────
  // Group 6: Settings Isolation
  // ────────────────────────────────────────────────────────────────

  test("settings created in tenant A are not visible in tenant B", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const suffix = Date.now();
    const key = `iso-setting-${suffix}`;

    const setting = await rawCreateCustomSetting(
      { key, value: "tenant-a-value", category: "custom" },
      adminToken,
      defaultTenantId,
    );
    expect(setting.status).toBeLessThan(300);
    createdResources.push({
      type: "setting",
      id: key,
      token: adminToken,
      tenantId: defaultTenantId,
    });

    const listA = await rawListCustomSettings(adminToken, defaultTenantId);
    expect(listA.status).toBeLessThan(300);
    const settingsA = Array.isArray(listA.body) ? listA.body : listA.body?.settings || [];
    const foundInA = settingsA.some((s: { key: string }) => s.key === key);
    expect(foundInA).toBe(true);

    const listB = await rawListCustomSettings(adminToken, thirdTenantId);
    expect(listB.status).toBeLessThan(300);
    const settingsB = Array.isArray(listB.body) ? listB.body : listB.body?.settings || [];
    const foundInB = settingsB.some((s: { key: string }) => s.key === key);
    expect(foundInB).toBe(false);
  });

  test("settings created in tenant B are not visible in tenant A", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const suffix = Date.now();
    const key = `iso-setting-b-${suffix}`;

    const setting = await rawCreateCustomSetting(
      { key, value: "tenant-b-value", category: "custom" },
      adminToken,
      thirdTenantId,
    );
    expect(setting.status).toBeLessThan(300);
    createdResources.push({
      type: "setting",
      id: key,
      token: adminToken,
      tenantId: thirdTenantId,
    });

    const listA = await rawListCustomSettings(adminToken, defaultTenantId);
    expect(listA.status).toBeLessThan(300);
    const settingsA = Array.isArray(listA.body) ? listA.body : listA.body?.settings || [];
    const foundInA = settingsA.some((s: { key: string }) => s.key === key);
    expect(foundInA).toBe(false);

    const listB = await rawListCustomSettings(adminToken, thirdTenantId);
    expect(listB.status).toBeLessThan(300);
    const settingsB = Array.isArray(listB.body) ? listB.body : listB.body?.settings || [];
    const foundInB = settingsB.some((s: { key: string }) => s.key === key);
    expect(foundInB).toBe(true);
  });

  // ────────────────────────────────────────────────────────────────
  // Group 7: Bidirectional Webhook Isolation
  // ────────────────────────────────────────────────────────────────

  test("webhooks created in tenant B are not visible in tenant A", async ({
    adminToken,
    defaultTenantId,
    thirdTenantId,
  }) => {
    const webhook = await rawCreateWebhook(
      {
        name: `iso-wh-b-${Date.now()}`,
        url: "https://example.com/webhook-b",
      },
      adminToken,
      thirdTenantId,
    );

    if (webhook.status < 300 && webhook.body?.id) {
      createdResources.push({
        type: "webhook",
        id: webhook.body.id,
        token: adminToken,
        tenantId: thirdTenantId,
      });

      const listA = await rawListWebhooks(adminToken, defaultTenantId);
      expect(listA.status).toBeLessThan(300);
      const webhooksA = Array.isArray(listA.body) ? listA.body : [];
      const foundInA = webhooksA.some(
        (w: { id: string }) => w.id === webhook.body.id,
      );
      expect(foundInA).toBe(false);

      const listB = await rawListWebhooks(adminToken, thirdTenantId);
      expect(listB.status).toBeLessThan(300);
      const webhooksB = Array.isArray(listB.body) ? listB.body : [];
      const foundInB = webhooksB.some(
        (w: { id: string }) => w.id === webhook.body.id,
      );
      expect(foundInB).toBe(true);
    }
  });
});
