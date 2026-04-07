import { test, expect } from "./fixtures";
import {
  rawCreateFunction,
  rawDeleteFunction,
  rawInvokeFunction,
} from "./helpers/api";

test.describe("Edge Functions Execution", () => {
  let adminToken: string;

  test.beforeAll(async () => {
    const { rawLogin } = await import("./helpers/api");
    const result = await rawLogin({
      email: "admin@fluxbase.test",
      password: "test-password-32chars!!",
    });
    adminToken = result.body.access_token;
  });

  const createdFunctions: Array<{ name: string; tenantId?: string }> = [];

  test.afterAll(async () => {
    for (const { name, tenantId } of createdFunctions) {
      await rawDeleteFunction(name, adminToken, tenantId).catch(() => {});
    }
  });

  test("functions page loads without errors", async ({ adminPage }) => {
    await adminPage.goto("functions", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/functions/);

    // Check for JS errors
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

  test("create function via API and verify in UI", async ({ adminPage }) => {
    const funcName = `e2e-func-${Date.now()}`;
    const code = `
export default function handler(req: Request): Response {
  return new Response(JSON.stringify({ message: "Hello from E2E" }), {
    headers: { "Content-Type": "application/json" },
  });
}`;

    // Create via API
    const createResult = await rawCreateFunction(
      { name: funcName, code },
      adminToken,
    );
    expect(createResult.status).toBeLessThan(300);
    createdFunctions.push({ name: funcName });

    // Navigate to functions page and verify
    await adminPage.goto("functions", { waitUntil: "networkidle" });
    await expect(adminPage.getByText(funcName)).toBeVisible({
      timeout: 10_000,
    });
  });

  test("invoke function via API and verify response", async () => {
    const funcName = `e2e-invoke-${Date.now()}`;
    const code = `
export default function handler(req: Request): Response {
  return new Response(JSON.stringify({ result: "success", timestamp: Date.now() }), {
    headers: { "Content-Type": "application/json" },
  });
}`;

    await rawCreateFunction({ name: funcName, code }, adminToken);
    createdFunctions.push({ name: funcName });

    // Invoke via API
    const invokeResult = await rawInvokeFunction(funcName, adminToken);
    expect(invokeResult.status).toBeLessThan(300);
    expect(invokeResult.body).toBeTruthy();
  });

  test("delete function via API and verify removal from UI", async ({
    adminPage,
  }) => {
    const funcName = `e2e-delete-${Date.now()}`;
    const code = `export default function handler(req: Request): Response { return new Response("ok"); }`;

    await rawCreateFunction({ name: funcName, code }, adminToken);
    createdFunctions.push({ name: funcName });

    // Verify it's visible
    await adminPage.goto("functions", { waitUntil: "networkidle" });
    await expect(adminPage.getByText(funcName)).toBeVisible({
      timeout: 10_000,
    });

    // Delete via API
    await rawDeleteFunction(funcName, adminToken);

    // Verify it's gone from UI
    await adminPage.goto("functions", { waitUntil: "networkidle" });
    await expect(adminPage.getByText(funcName)).not.toBeVisible({
      timeout: 5_000,
    });
  });

  test("functions are tenant-scoped in UI", async ({
    adminPage,
    adminToken,
  }) => {
    const funcNameA = `e2e-tenant-A-${Date.now()}`;
    const funcNameB = `e2e-tenant-B-${Date.now()}`;
    const code = `export default function handler(req: Request): Response { return new Response("ok"); }`;

    // Create in default tenant
    const { listTenants } = await import("./helpers/api");
    const tenantsResult = await listTenants(adminToken);
    const tenants = tenantsResult.body;
    const defaultTenant = tenants.find(
      (t: { is_default: boolean }) => t.is_default,
    );
    const otherTenant = tenants.find(
      (t: { is_default: boolean }) => !t.is_default,
    );

    if (!otherTenant) {
      test.skip();
      return;
    }

    await rawCreateFunction(
      { name: funcNameA, code },
      adminToken,
      defaultTenant.id,
    );
    await rawCreateFunction(
      { name: funcNameB, code },
      adminToken,
      otherTenant.id,
    );
    createdFunctions.push({ name: funcNameA, tenantId: defaultTenant.id });
    createdFunctions.push({ name: funcNameB, tenantId: otherTenant.id });

    // View functions in default tenant context
    await adminPage.goto("functions", { waitUntil: "networkidle" });
    await expect(adminPage.getByText(funcNameA)).toBeVisible({
      timeout: 10_000,
    });

    // Switch to other tenant
    const selector = adminPage.getByRole("combobox", {
      name: "Select tenant",
    });
    if ((await selector.isVisible().catch(() => false)) === true) {
      await selector.click();
      await expect(adminPage.getByRole("listbox")).toBeVisible({
        timeout: 5_000,
      });
      const otherOption = adminPage
        .getByRole("option")
        .filter({ hasText: otherTenant.name });
      if ((await otherOption.isVisible().catch(() => false)) === true) {
        await otherOption.click();
        await adminPage.waitForTimeout(1000);
        await adminPage.goto("functions", { waitUntil: "networkidle" });
        // Should see funcNameB but NOT funcNameA
        await expect(adminPage.getByText(funcNameB)).toBeVisible({
          timeout: 10_000,
        });
        await expect(adminPage.getByText(funcNameA)).not.toBeVisible({
          timeout: 5_000,
        });
      }
    }
  });
});
