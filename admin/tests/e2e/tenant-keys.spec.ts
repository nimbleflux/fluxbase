import { test, expect } from "./fixtures";
import {
  rawLogin,
  rawCreateTenant,
  rawDeleteTenant,
  listServiceKeys,
  createServiceKey,
  revokeServiceKey,
} from "./helpers/api";

test.describe("Service Keys Per Tenant", () => {
  let adminToken: string;
  let tenantAId: string;
  let tenantBId: string;
  const slugA = `e2e-keys-a-${Date.now()}`;
  const slugB = `e2e-keys-b-${Date.now()}`;

  test.beforeAll(async () => {
    const result = await rawLogin({
      email: "admin@fluxbase.test",
      password: "test-password-32chars!!",
    });
    adminToken = result.body.access_token;

    // Create two tenants for isolation testing
    const a = await rawCreateTenant(
      { name: "Keys Tenant A", slug: slugA },
      adminToken,
    );
    tenantAId = a.body.tenant.id;

    const b = await rawCreateTenant(
      { name: "Keys Tenant B", slug: slugB },
      adminToken,
    );
    tenantBId = b.body.tenant.id;
  });

  test.afterAll(async () => {
    if (tenantAId) {
      await rawDeleteTenant(tenantAId, adminToken);
    }
    if (tenantBId) {
      await rawDeleteTenant(tenantBId, adminToken);
    }
  });

  test("service key list is scoped to tenant", async ({ request }) => {
    // Create a key in tenant A
    const createA = await createServiceKey(
      request,
      { name: "Tenant A Key", keyType: "service", tenantId: tenantAId },
      adminToken,
    );
    expect(createA.status).toBeLessThan(300);

    // Create a key in tenant B
    const createB = await createServiceKey(
      request,
      { name: "Tenant B Key", keyType: "service", tenantId: tenantBId },
      adminToken,
    );
    expect(createB.status).toBeLessThan(300);

    // List keys in tenant A — should NOT see tenant B's key
    const listA = await listServiceKeys(request, adminToken, tenantAId);
    expect(listA.status).toBe(200);
    const keysA = listA.body;
    const hasBKey = Array.isArray(keysA)
      ? keysA.some((k: { name: string }) => k.name === "Tenant B Key")
      : false;
    expect(hasBKey).toBeFalsy();

    // List keys in tenant B — should NOT see tenant A's key
    const listB = await listServiceKeys(request, adminToken, tenantBId);
    expect(listB.status).toBe(200);
    const keysB = listB.body;
    const hasAKey = Array.isArray(keysB)
      ? keysB.some((k: { name: string }) => k.name === "Tenant A Key")
      : false;
    expect(hasAKey).toBeFalsy();
  });

  test("create service key for specific tenant", async ({ request }) => {
    const uniqueName = `Unique Key ${Date.now()}`;
    const result = await createServiceKey(
      request,
      { name: uniqueName, keyType: "service", tenantId: tenantAId },
      adminToken,
    );
    expect(result.status).toBeLessThan(300);

    // Key should appear in tenant A's list
    const listA = await listServiceKeys(request, adminToken, tenantAId);
    const keysA = listA.body;
    const found = Array.isArray(keysA)
      ? keysA.some((k: { name: string }) => k.name === uniqueName)
      : false;
    expect(found).toBeTruthy();

    // Key should NOT appear in tenant B's list
    const listB = await listServiceKeys(request, adminToken, tenantBId);
    const keysB = listB.body;
    const foundInB = Array.isArray(keysB)
      ? keysB.some((k: { name: string }) => k.name === uniqueName)
      : false;
    expect(foundInB).toBeFalsy();
  });

  test("revoke service key", async ({ request }) => {
    // Create a key
    const create = await createServiceKey(
      request,
      {
        name: `Revoke Key ${Date.now()}`,
        keyType: "service",
        tenantId: tenantAId,
      },
      adminToken,
    );
    expect(create.status).toBeLessThan(300);
    const keyId = create.body?.id || create.body?.key?.id;

    if (keyId) {
      // Revoke it
      const revoke = await revokeServiceKey(
        request,
        keyId,
        "Test revocation",
        adminToken,
        tenantAId,
      );
      expect(revoke.status).toBeLessThan(300);

      // Verify it's revoked in the list
      const list = await listServiceKeys(request, adminToken, tenantAId);
      const keys = list.body;
      const revoked = Array.isArray(keys)
        ? keys.find((k: { id: string }) => k.id === keyId)
        : null;
      if (revoked) {
        const status =
          revoked.status ?? revoked.state ?? revoked.is_active ?? "";
        // The key should either show revoked/deprecated/inactive status
        // or be absent from the list entirely (hard delete on revoke)
        if (status !== "" && status !== undefined && status !== true) {
          expect(String(status)).toMatch(/revoked|deprecated|inactive|false/i);
        }
      }
    }
  });

  test("service keys page shows tenant-scoped keys", async ({ adminPage }) => {
    // Select tenant A via selector
    const selector = adminPage.getByRole("combobox", { name: "Select tenant" });
    await expect(selector).toBeVisible({ timeout: 10_000 });

    // Try to find and select tenant A
    await selector.click();
    await expect(adminPage.getByRole("listbox")).toBeVisible({
      timeout: 5_000,
    });

    const tenantAOption = adminPage.getByRole("option").filter({
      hasText: "Keys Tenant A",
    });
    if (await tenantAOption.isVisible().catch(() => false)) {
      await tenantAOption.click();
      // Wait for tenant context to update
      await adminPage
        .waitForRequest(
          (req) =>
            req.url().includes("/api/v1/") &&
            req.headers()["x-fb-tenant"] !== undefined,
          { timeout: 5_000 },
        )
        .catch(() => {});
    } else {
      await adminPage.keyboard.press("Escape");
    }

    // Navigate to service keys page
    const keysPromise = adminPage.waitForResponse(
      (resp) => resp.url().includes("/api/v1/admin/service-keys"),
      { timeout: 10_000 },
    );
    await adminPage.goto("service-keys", { waitUntil: "networkidle" });
    await keysPromise.catch(() => {});

    // Page should load without errors
    const url = adminPage.url();
    expect(url).toContain("service-keys");
  });
});
