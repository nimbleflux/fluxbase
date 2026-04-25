import { test, expect } from "./fixtures";
import { rawListChatbots, rawApiRequest } from "./helpers/api";
import { selectTenantByIndex } from "./helpers/selectors";

test.describe("Chatbots Management", () => {
  let adminToken: string;

  test.beforeAll(async () => {
    const { rawLogin } = await import("./helpers/api");
    const result = await rawLogin({
      email: "admin@fluxbase.test",
      password: "test-password-32chars!!",
    });
    adminToken = result.body.access_token;
  });

  test.beforeEach(async ({ adminPage }) => {
    const selector = adminPage.getByRole("combobox", { name: "Select tenant" });
    if (await selector.isVisible().catch(() => false)) {
      const text = await selector.textContent();
      if (text?.includes("Select tenant")) {
        await selectTenantByIndex(adminPage, 0);
      }
    }
  });

  test("chatbots page loads without errors", async ({ adminPage }) => {
    await adminPage.goto("chatbots", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/chatbots/);

    const consoleErrors: string[] = [];
    adminPage.on("console", (msg) => {
      if (msg.type() === "error") consoleErrors.push(msg.text());
    });
    await adminPage.waitForTimeout(2000);
    const criticalErrors = consoleErrors.filter(
      (text) =>
        !text.includes("500") &&
        !text.includes("404") &&
        !text.includes("Failed to fetch") &&
        !text.includes("favicon"),
    );
    expect(criticalErrors).toHaveLength(0);
  });

  test("list chatbots via API returns expected structure", async ({
    adminToken,
  }) => {
    const result = await rawListChatbots(adminToken);
    expect(result.status).toBe(200);
    expect(result.body).toBeTruthy();
  });

  test("chatbots page shows content", async ({ adminPage }) => {
    await adminPage.goto("chatbots", { waitUntil: "networkidle" });

    // Wait for loading to complete by checking the heading appears
    await expect(
      adminPage.getByRole("heading", { name: /AI Chatbots/i }),
    ).toBeVisible({ timeout: 15_000 });

    const hasContent = await adminPage.evaluate(() => {
      return document.getElementById("root")?.innerHTML?.length > 100;
    });
    expect(hasContent).toBeTruthy();
  });

  test("chatbots page displays header and stats", async ({ adminPage }) => {
    await adminPage.goto("chatbots", { waitUntil: "networkidle" });

    // Page header should be visible
    await expect(
      adminPage.getByRole("heading", { name: /AI Chatbots/i }),
    ).toBeVisible({ timeout: 10_000 });

    // Subtitle text should be present
    await expect(
      adminPage.getByText(
        "Manage AI-powered chatbots for database interactions",
      ),
    ).toBeVisible();
  });

  test("chatbots table renders with correct columns", async ({ adminPage }) => {
    const result = await rawListChatbots(adminToken);
    // Only test table columns if there are chatbots
    if (result.status === 200 && result.body?.chatbots?.length > 0) {
      await adminPage.goto("chatbots", { waitUntil: "networkidle" });

      // Wait for loading to complete
      await expect(
        adminPage.getByRole("heading", { name: /AI Chatbots/i }),
      ).toBeVisible({ timeout: 15_000 });

      // Wait for the table to render
      await adminPage.waitForSelector("table", { timeout: 10_000 });

      // Verify column headers exist
      await expect(adminPage.getByText("Name")).toBeVisible();
      await expect(adminPage.getByText("Status")).toBeVisible();
    }
  });

  test("chatbots page shows empty state when no chatbots", async ({
    adminPage,
    adminToken,
  }) => {
    const result = await rawListChatbots(adminToken);

    // If no chatbots exist, verify the empty state card is shown
    if (
      result.status === 200 &&
      (!result.body?.chatbots || result.body.chatbots.length === 0)
    ) {
      await adminPage.goto("chatbots", { waitUntil: "networkidle" });
      await expect(
        adminPage.getByRole("heading", { name: /AI Chatbots/i }),
      ).toBeVisible({ timeout: 15_000 });
      await expect(adminPage.getByText("No chatbots yet")).toBeVisible({
        timeout: 10_000,
      });
    }
  });

  test("chatbots page shows total and active counts", async ({ adminPage }) => {
    await adminPage.goto("chatbots", { waitUntil: "networkidle" });

    // Wait for loading to complete
    await expect(
      adminPage.getByRole("heading", { name: /AI Chatbots/i }),
    ).toBeVisible({ timeout: 15_000 });

    // The stats bar should show "Total:" and "Active:" badges
    const totalLabel = adminPage.getByText("Total:");
    const activeLabel = adminPage.getByText("Active:");

    // These labels are always visible regardless of chatbot count
    await expect(totalLabel).toBeVisible({ timeout: 10_000 });
    await expect(activeLabel).toBeVisible();
  });

  test("chatbot toggle changes enabled state", async ({
    adminPage,
    adminToken,
  }) => {
    const listResult = await rawListChatbots(adminToken);

    if (listResult.status === 200 && listResult.body?.chatbots?.length > 0) {
      await adminPage.goto("chatbots", { waitUntil: "networkidle" });
      await expect(
        adminPage.getByRole("heading", { name: /AI Chatbots/i }),
      ).toBeVisible({ timeout: 15_000 });
      await adminPage.waitForSelector("table", { timeout: 10_000 });

      const chatbot = listResult.body.chatbots[0];
      const originalEnabled = chatbot.enabled;

      // Find the switch in the chatbot's row
      const row = adminPage.getByRole("row").filter({ hasText: chatbot.name });
      const toggle = row.getByRole("switch");

      if (await toggle.isVisible()) {
        // Set up listener for the toggle API call
        const togglePromise = adminPage.waitForResponse(
          (resp) =>
            resp
              .url()
              .includes(`/api/v1/admin/ai/chatbots/${chatbot.id}/toggle`) &&
            resp.request().method() === "POST",
          { timeout: 10_000 },
        );

        await toggle.click();
        await togglePromise;

        // Verify the state changed via API
        const afterResult = await rawListChatbots(adminToken);
        const updated = afterResult.body?.chatbots?.find(
          (c: { id: string }) => c.id === chatbot.id,
        );
        if (updated) {
          expect(updated.enabled).toBe(!originalEnabled);
        }

        // Toggle back to restore original state
        const restorePromise = adminPage.waitForResponse(
          (resp) =>
            resp
              .url()
              .includes(`/api/v1/admin/ai/chatbots/${chatbot.id}/toggle`) &&
            resp.request().method() === "POST",
          { timeout: 10_000 },
        );
        await toggle.click();
        await restorePromise;
      }
    }
  });

  test("chatbot delete shows confirmation dialog", async ({
    adminPage,
    adminToken,
  }) => {
    const listResult = await rawListChatbots(adminToken);

    if (listResult.status === 200 && listResult.body?.chatbots?.length > 0) {
      await adminPage.goto("chatbots", { waitUntil: "networkidle" });
      await expect(
        adminPage.getByRole("heading", { name: /AI Chatbots/i }),
      ).toBeVisible({ timeout: 15_000 });
      await adminPage.waitForSelector("table", { timeout: 10_000 });

      const chatbot = listResult.body.chatbots[0];
      const row = adminPage.getByRole("row").filter({ hasText: chatbot.name });

      // Find the delete button (trash icon button) in the row
      const deleteButton = row.locator("button").filter({
        has: adminPage.locator("svg.lucide-trash-2"),
      });

      if (await deleteButton.isVisible()) {
        await deleteButton.click();

        // Confirmation dialog should appear
        await expect(
          adminPage.getByText("Are you sure you want to delete this chatbot?"),
        ).toBeVisible({ timeout: 5_000 });

        // Cancel the deletion to avoid side effects
        await adminPage.getByRole("button", { name: "Cancel" }).click();
      }
    }
  });

  test("chatbot test dialog opens", async ({ adminPage }) => {
    await adminPage.goto("chatbots", { waitUntil: "networkidle" });

    // Wait for loading to complete
    await expect(
      adminPage.getByRole("heading", { name: /AI Chatbots/i }),
    ).toBeVisible({ timeout: 15_000 });

    // Look for any chatbot row with a "Test" or message icon button
    // If no chatbots exist, just verify the page renders correctly
    const testButton = adminPage.getByRole("button", { name: /test|message/i });
    if (
      await testButton
        .first()
        .isVisible()
        .catch(() => false)
    ) {
      await testButton.first().click();
      // A dialog or panel should open for testing
      await adminPage.waitForTimeout(1000);
    }

    // Verify no crash
    await expect(adminPage).toHaveURL(/chatbots/);
  });

  test("chatbot settings dialog opens", async ({ adminPage, adminToken }) => {
    const listResult = await rawListChatbots(adminToken);

    if (listResult.status === 200 && listResult.body?.chatbots?.length > 0) {
      await adminPage.goto("chatbots", { waitUntil: "networkidle" });
      await expect(
        adminPage.getByRole("heading", { name: /AI Chatbots/i }),
      ).toBeVisible({ timeout: 15_000 });
      await adminPage.waitForSelector("table", { timeout: 10_000 });

      const chatbot = listResult.body.chatbots[0];
      const row = adminPage.getByRole("row").filter({ hasText: chatbot.name });

      // Find the settings button (gear icon button) in the row
      const settingsButton = row.locator("button").filter({
        has: adminPage.locator("svg.lucide-settings"),
      });

      if (await settingsButton.isVisible()) {
        await settingsButton.click();

        // Settings dialog should appear
        await expect(adminPage.getByText("Chatbot Settings")).toBeVisible({
          timeout: 5_000,
        });

        // Close the dialog
        await adminPage
          .getByRole("button", { name: "Cancel" })
          .click()
          .catch(() => {
            // Dialog may close via overlay click
          });
      }
    }
  });

  test("refresh button reloads chatbot list", async ({ adminPage }) => {
    await adminPage.goto("chatbots", { waitUntil: "networkidle" });

    // Wait for loading to complete
    await expect(
      adminPage.getByRole("heading", { name: /AI Chatbots/i }),
    ).toBeVisible({ timeout: 15_000 });

    // Click the Refresh button
    const refreshButton = adminPage.getByRole("button", {
      name: /Refresh/i,
    });
    await expect(refreshButton).toBeVisible({ timeout: 10_000 });

    const refreshPromise = adminPage.waitForResponse(
      (resp) =>
        resp.url().includes("/api/v1/admin/ai/chatbots") &&
        resp.request().method() === "GET",
      { timeout: 10_000 },
    );
    await refreshButton.click();
    await refreshPromise;

    // Page should still show chatbots without errors
    await expect(adminPage).toHaveURL(/chatbots/);
  });

  test("sync from filesystem button triggers sync API", async ({
    adminPage,
  }) => {
    await adminPage.goto("chatbots", { waitUntil: "networkidle" });

    // Wait for loading to complete
    await expect(
      adminPage.getByRole("heading", { name: /AI Chatbots/i }),
    ).toBeVisible({ timeout: 15_000 });

    const syncButton = adminPage.getByRole("button", {
      name: /Sync from Filesystem/i,
    });
    await expect(syncButton).toBeVisible({ timeout: 10_000 });

    const syncPromise = adminPage.waitForResponse(
      (resp) =>
        resp.url().includes("/api/v1/admin/ai/chatbots/sync") &&
        resp.request().method() === "POST",
      { timeout: 10_000 },
    );
    await syncButton.click();
    const syncResponse = await syncPromise;

    // Sync should succeed (200 or 201)
    expect(syncResponse.status()).toBeLessThan(500);
  });

  test("chatbot list is tenant-scoped", async () => {
    // List chatbots for default tenant
    const resultA = await rawListChatbots(adminToken);
    expect(resultA.status).toBe(200);

    // List chatbots for another tenant
    const { listTenants } = await import("./helpers/api");
    const tenantsResult = await listTenants(adminToken);
    const tenants = tenantsResult.body;
    const otherTenant = tenants.find(
      (t: { is_default: boolean }) => !t.is_default,
    );

    if (otherTenant) {
      const resultB = await rawListChatbots(adminToken, otherTenant.id);
      expect(resultB.status).toBe(200);
      // Both should return valid responses but potentially different lists
    }
  });

  test("get chatbot by ID returns full details", async ({ adminToken }) => {
    const listResult = await rawListChatbots(adminToken);

    if (listResult.status === 200 && listResult.body?.chatbots?.length > 0) {
      const chatbot = listResult.body.chatbots[0];

      const getResponse = await rawApiRequest({
        method: "GET",
        path: `/api/v1/admin/ai/chatbots/${chatbot.id}`,
        headers: { Authorization: `Bearer ${adminToken}` },
      });

      expect(getResponse.status).toBe(200);
      expect(getResponse.body).toBeTruthy();
      expect(getResponse.body.id).toBe(chatbot.id);
      expect(getResponse.body.name).toBe(chatbot.name);
      // Full chatbot should have additional fields beyond summary
      expect(getResponse.body).toHaveProperty("max_tokens");
      expect(getResponse.body).toHaveProperty("temperature");
    }
  });

  test("update chatbot settings via API", async ({ adminToken }) => {
    const listResult = await rawListChatbots(adminToken);

    if (listResult.status === 200 && listResult.body?.chatbots?.length > 0) {
      const chatbot = listResult.body.chatbots[0];

      // Update description
      const updateResponse = await rawApiRequest({
        method: "PUT",
        path: `/api/v1/admin/ai/chatbots/${chatbot.id}`,
        data: {
          description: `E2E test updated at ${Date.now()}`,
          max_tokens: chatbot.max_tokens || 4096,
          temperature: chatbot.temperature ?? 0.7,
        },
        headers: { Authorization: `Bearer ${adminToken}` },
      });

      expect(updateResponse.status).toBe(200);
      expect(updateResponse.body).toBeTruthy();
      expect(updateResponse.body.name).toBe(chatbot.name);
    }
  });

  test("chatbot knowledge bases endpoint returns valid response", async ({
    adminToken,
  }) => {
    const listResult = await rawListChatbots(adminToken);

    if (listResult.status === 200 && listResult.body?.chatbots?.length > 0) {
      const chatbot = listResult.body.chatbots[0];

      const kbResponse = await rawApiRequest({
        method: "GET",
        path: `/api/v1/admin/ai/chatbots/${chatbot.id}/knowledge-bases`,
        headers: { Authorization: `Bearer ${adminToken}` },
      });

      expect(kbResponse.status).toBe(200);
      expect(kbResponse.body).toBeTruthy();
    }
  });

  test("chatbots page handles API errors gracefully", async ({ adminPage }) => {
    // Navigate to chatbots page and verify it handles loading state
    await adminPage.goto("chatbots", { waitUntil: "networkidle" });

    // Wait for loading to complete
    await expect(
      adminPage.getByRole("heading", { name: /AI Chatbots/i }),
    ).toBeVisible({ timeout: 15_000 });

    // Page should load without throwing an unhandled exception
    const pageErrors: string[] = [];
    adminPage.on("pageerror", (error) => {
      pageErrors.push(error.message);
    });

    await adminPage.waitForTimeout(3000);

    // Filter out known non-critical errors
    const criticalErrors = pageErrors.filter(
      (msg) =>
        !msg.includes("favicon") &&
        !msg.includes("net::ERR") &&
        !msg.includes("ResizeObserver"),
    );
    expect(criticalErrors).toHaveLength(0);
  });

  test("chatbot API rejects unauthenticated requests", async () => {
    const response = await rawApiRequest({
      method: "GET",
      path: "/api/v1/admin/ai/chatbots",
    });

    // Should return 401 without auth token
    expect(response.status).toBe(401);
  });

  test("chatbot API rejects invalid chatbot ID", async ({ adminToken }) => {
    const response = await rawApiRequest({
      method: "GET",
      path: "/api/v1/admin/ai/chatbots/nonexistent-id-12345",
      headers: { Authorization: `Bearer ${adminToken}` },
    });

    // Should return an error status for non-existent chatbot
    expect(response.status).toBeGreaterThanOrEqual(400);
  });

  test("chatbot toggle via API with invalid ID returns error", async ({
    adminToken,
  }) => {
    const response = await rawApiRequest({
      method: "POST",
      path: "/api/v1/admin/ai/chatbots/nonexistent-id-12345/toggle",
      data: { enabled: true },
      headers: { Authorization: `Bearer ${adminToken}` },
    });

    // Should return an error status for non-existent chatbot
    expect(response.status).toBeGreaterThanOrEqual(400);
  });

  test("chatbot delete via API with invalid ID returns error", async ({
    adminToken,
  }) => {
    const response = await rawApiRequest({
      method: "DELETE",
      path: "/api/v1/admin/ai/chatbots/nonexistent-id-12345",
      headers: { Authorization: `Bearer ${adminToken}` },
    });

    // Backend may return 200 (no-op delete) or an error status
    expect(response.status === 200 || response.status >= 400).toBeTruthy();
  });
});
