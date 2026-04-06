import { test, expect } from "./fixtures";
import {
  rawLogin,
  rawCreateTenant,
  rawDeleteTenant,
} from "./helpers/api";
import { createPlatformUser, getUserByEmail } from "./helpers/db";

test.describe("Tenant Members", () => {
  let adminToken: string;
  let testTenantId: string;
  const testMemberEmail = `member-${Date.now()}@fluxbase.test`;
  const testMemberPassword = "member-password-32!!";
  const testTenantSlug = `e2e-members-${Date.now()}`;

  test.beforeAll(async () => {
    const result = await rawLogin({
      email: "admin@fluxbase.test",
      password: "test-password-32chars!!",
    });
    adminToken = result.body.access_token;

    // Create a tenant for member tests
    const createResult = await rawCreateTenant(
      {
        name: "Members Test Tenant",
        slug: testTenantSlug,
        autoGenerateKeys: true,
      },
      adminToken,
    );
    testTenantId = createResult.body.tenant.id;
  });

  test.afterAll(async () => {
    // Cleanup
    if (testTenantId) {
      await rawDeleteTenant(testTenantId, adminToken);
    }
  });

  test("view tenant members tab", async ({ adminPage }) => {
    // Navigate to tenant detail
    await adminPage.goto(`tenants/${testTenantId}`, {
      waitUntil: "networkidle",
    });

    // Should be on tenant detail page
    await expect(adminPage).toHaveURL(new RegExp(`/tenants/${testTenantId}`));
  });

  test("add member to tenant via API", async ({ request }) => {
    // Create a user to add as member
    const userId = await createPlatformUser(
      testMemberEmail,
      testMemberPassword,
      "Test Member",
      "tenant_member",
    );
    expect(userId).toBeTruthy();

    // Assign to tenant
    const assignResp = await request.fetch(
      `${process.env.PLAYWRIGHT_API_URL || "http://localhost:5050"}/api/v1/admin/tenants/${testTenantId}/admins`,
      {
        method: "POST",
        data: JSON.stringify({ user_id: userId }),
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${adminToken}`,
        },
      },
    );

    expect(assignResp.status()).toBeLessThan(300);
  });

  test("list tenant members", async ({ request }) => {
    const listResp = await request.fetch(
      `${process.env.PLAYWRIGHT_API_URL || "http://localhost:5050"}/api/v1/admin/tenants/${testTenantId}/admins`,
      {
        headers: { Authorization: `Bearer ${adminToken}` },
      },
    );

    expect(listResp.status()).toBe(200);
    const members = await listResp.json();
    expect(Array.isArray(members)).toBeTruthy();
    expect(members.length).toBeGreaterThanOrEqual(1);
  });

  test("remove member from tenant", async ({ request }) => {
    // Get the user ID
    const user = await getUserByEmail(testMemberEmail);
    if (!user) {
      test.skip();
      return;
    }

    // Remove the member
    const removeResp = await request.fetch(
      `${process.env.PLAYWRIGHT_API_URL || "http://localhost:5050"}/api/v1/admin/tenants/${testTenantId}/admins/${user.id}`,
      {
        method: "DELETE",
        headers: { Authorization: `Bearer ${adminToken}` },
      },
    );

    expect(removeResp.status()).toBeLessThan(300);

    // Verify member is removed
    const listResp = await request.fetch(
      `${process.env.PLAYWRIGHT_API_URL || "http://localhost:5050"}/api/v1/admin/tenants/${testTenantId}/admins`,
      {
        headers: { Authorization: `Bearer ${adminToken}` },
      },
    );
    const members = await listResp.json();
    const found = members.find(
      (m: { user_id: string }) => m.user_id === user.id,
    );
    expect(found).toBeFalsy();
  });
});
