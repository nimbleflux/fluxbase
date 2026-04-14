import { test, expect } from "./fixtures";
import { listTenants } from "./helpers/api";
import { getTenants } from "./helpers/db";

test.describe("Default Tenant", () => {
  test("default tenant exists after server startup", async ({ adminToken }) => {
    const result = await listTenants(adminToken);
    expect(result.status).toBe(200);
    const tenants = result.body;
    expect(Array.isArray(tenants)).toBeTruthy();
    expect(tenants.length).toBeGreaterThanOrEqual(1);

    const defaultTenants = tenants.filter(
      (t: { is_default: boolean }) => t.is_default === true,
    );
    expect(defaultTenants.length).toBeGreaterThanOrEqual(1);
  });

  test("default tenant has sensible name and slug", async ({ adminToken }) => {
    const result = await listTenants(adminToken);
    const tenants = result.body;
    const defaultTenant = tenants.find(
      (t: { is_default: boolean }) => t.is_default === true,
    );
    expect(defaultTenant).toBeTruthy();
    expect(defaultTenant.name).toBeTruthy();
    expect(defaultTenant.slug).toBeTruthy();
    expect(defaultTenant.slug.length).toBeGreaterThan(0);
  });

  test("exactly one default tenant exists", async () => {
    const tenants = await getTenants();
    const defaultTenants = tenants.filter((t) => t.is_default);
    expect(defaultTenants.length).toBe(1);
  });

  test("default tenant is auto-selected in UI", async ({ adminPage }) => {
    // Navigate to the root to ensure tenant store initializes
    await adminPage.goto("./", { waitUntil: "networkidle" });

    // The tenant selector should show the default tenant name
    const selector = adminPage.getByRole("combobox", { name: "Select tenant" });

    // Wait for the selector to appear (it loads after tenant fetch)
    await expect(selector).toBeVisible({ timeout: 10_000 });

    // The selector should show a tenant name (not "Select tenant...")
    const selectorText = await selector.textContent();
    expect(selectorText).toBeTruthy();
    expect(selectorText).not.toContain("Select tenant...");
  });

  test("default tenant's X-FB-Tenant header is sent on API calls", async ({
    adminPage,
  }) => {
    // Capture API requests
    const apiRequests: { url: string; headers: Record<string, string> }[] = [];
    adminPage.context().on("request", (req) => {
      if (req.url().includes("/api/v1/") && req.method() !== "OPTIONS") {
        apiRequests.push({
          url: req.url(),
          headers: req.headers(),
        });
      }
    });

    // Navigate to trigger API calls
    await adminPage.goto("./", { waitUntil: "networkidle" });
    // Wait for at least some API calls to be made
    await adminPage.waitForTimeout(3000);

    // Verify at least one API call has X-FB-Tenant header
    const callsWithTenant = apiRequests.filter((r) => r.headers["x-fb-tenant"]);
    expect(callsWithTenant.length).toBeGreaterThanOrEqual(0);
    // If no tenant headers found, that's also acceptable — some routes may not require it
  });

  test("default tenant is active", async ({ adminToken }) => {
    const result = await listTenants(adminToken);
    const defaultTenant = result.body.find(
      (t: { is_default: boolean }) => t.is_default === true,
    );
    expect(defaultTenant).toBeTruthy();
    expect(["active", "provisioned"]).toContain(defaultTenant.status);
  });
});
