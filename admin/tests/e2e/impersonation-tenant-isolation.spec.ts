import { test, expect } from "./fixtures";
import {
  rawStartUserImpersonation,
  rawStopImpersonation,
  rawCreateBucket,
  rawListBuckets,
  rawCreateFunction,
  rawListFunctions,
  rawCreateSecret,
  rawListSecrets,
  rawDeleteFunction,
  rawDeleteSecret,
  rawApiRequest,
  rawLogin,
} from "./helpers/api";
import { createPlatformUser } from "./helpers/db";
import { ADMIN_EMAIL, ADMIN_PASSWORD } from "./helpers/constants";

test.describe("Impersonation Tenant Isolation", () => {
  const cleanupTasks: Array<() => Promise<void>> = [];

  test.afterAll(async () => {
    // Re-login to get a fresh admin token for cleanup
    const loginResult = await rawLogin({
      email: ADMIN_EMAIL,
      password: ADMIN_PASSWORD,
    });
    const freshToken = loginResult.body?.access_token;

    for (const cleanup of cleanupTasks) {
      await cleanup().catch(() => {});
    }

    // Stop any lingering impersonation sessions
    if (freshToken) {
      await rawStopImpersonation(freshToken).catch(() => {});
    }
  });

  test("impersonated user in tenant A cannot list tenant B's buckets", async ({
    adminToken,
    defaultTenantId,
    tenantAdminInfo,
  }) => {
    const secondTenantId = tenantAdminInfo.tenantId;

    // Create buckets in different tenants
    const bucketA = `imp-iso-A-${Date.now()}`;
    const bucketB = `imp-iso-B-${Date.now()}`;
    await rawCreateBucket(bucketA, adminToken, defaultTenantId);
    await rawCreateBucket(bucketB, adminToken, secondTenantId);

    cleanupTasks.push(
      async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/storage/buckets/${bucketA}`,
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "X-FB-Tenant": defaultTenantId,
          },
        });
      },
      async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/storage/buckets/${bucketB}`,
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "X-FB-Tenant": secondTenantId,
          },
        });
      },
    );

    // Start impersonation as the tenant admin (who belongs to second tenant)
    const impResult = await rawStartUserImpersonation(
      tenantAdminInfo.userId,
      "E2E bucket isolation test",
      adminToken,
    );
    expect(impResult.status).toBe(200);
    expect(impResult.body.access_token).toBeTruthy();
    const impToken = impResult.body.access_token;

    try {
      // List buckets with impersonation token — should only see own tenant's bucket
      const bucketsResult = await rawListBuckets(impToken);
      expect(bucketsResult.status).toBe(200);
      const bucketList = (bucketsResult.body?.buckets ||
        bucketsResult.body ||
        []) as Array<{ id: string }>;
      const bucketIds = bucketList.map((b) => b.id);

      // Should NOT see default tenant's bucket
      expect(bucketIds).not.toContain(bucketA);
      // SHOULD see own tenant's bucket
      expect(bucketIds).toContain(bucketB);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("impersonated user in tenant A cannot list tenant B's functions", async ({
    adminToken,
    defaultTenantId,
    tenantAdminInfo,
  }) => {
    // Create function in default tenant
    const funcName = `imp-iso-func-${Date.now()}`;
    const code = `export default function handler(req) { return new Response("ok"); }`;
    await rawCreateFunction(
      { name: funcName, code },
      adminToken,
      defaultTenantId,
    );
    cleanupTasks.push(async () => {
      await rawDeleteFunction(funcName, adminToken, defaultTenantId);
    });

    // Start impersonation as the tenant admin (belongs to second tenant)
    const impResult = await rawStartUserImpersonation(
      tenantAdminInfo.userId,
      "E2E function isolation test",
      adminToken,
    );
    expect(impResult.status).toBe(200);
    expect(impResult.body.access_token).toBeTruthy();
    const impToken = impResult.body.access_token;

    try {
      // List functions — should not see default tenant's function
      const functionsResult = await rawListFunctions(impToken);
      expect(functionsResult.status).toBe(200);
      const functions = (functionsResult.body?.functions ||
        functionsResult.body ||
        []) as Array<{ name: string }>;
      const funcNames = functions.map((f) => f.name);
      expect(funcNames).not.toContain(funcName);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("impersonated user in tenant A cannot list tenant B's secrets", async ({
    adminToken,
    defaultTenantId,
    tenantAdminInfo,
  }) => {
    // Create secret in default tenant
    const secretName = `imp-iso-secret-${Date.now()}`;
    const createResult = await rawCreateSecret(
      { name: secretName, value: "secret-val" },
      adminToken,
      defaultTenantId,
    );
    expect([200, 201]).toContain(createResult.status);
    const secretId = createResult.body?.id;
    if (secretId) {
      cleanupTasks.push(async () => {
        await rawDeleteSecret(secretId, adminToken, defaultTenantId);
      });
    }

    // Start impersonation as the tenant admin (belongs to second tenant)
    const impResult = await rawStartUserImpersonation(
      tenantAdminInfo.userId,
      "E2E secret isolation test",
      adminToken,
    );
    expect(impResult.status).toBe(200);
    expect(impResult.body.access_token).toBeTruthy();
    const impToken = impResult.body.access_token;

    try {
      const secretsResult = await rawListSecrets(impToken);
      expect(secretsResult.status).toBe(200);
      const secrets = (secretsResult.body || []) as Array<{ name: string }>;
      const secretNames = secrets.map((s) => s.name);
      expect(secretNames).not.toContain(secretName);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("user search in impersonation dialog is tenant-scoped", async ({
    adminToken,
    tenantAdminInfo,
  }) => {
    // This test verifies the API behavior: the user search endpoint passes
    // the current tenant ID as a filter, so only users in that tenant appear.
    // We test this by checking the API directly.

    // Create a user in the default tenant only (not in second tenant)
    const uniqueEmail = `imp-search-${Date.now()}@fluxbase.test`;
    const userId = await createPlatformUser(
      uniqueEmail,
      "password-32!!",
      "Search Test User",
      "user",
    );

    cleanupTasks.push(async () => {
      // Clean up the created user from both platform.users and auth.users
      const { execute } = await import("./helpers/db");
      await execute("DELETE FROM auth.users WHERE id = $1::uuid", [
        userId,
      ]).catch(() => {});
      await execute("DELETE FROM platform.users WHERE id = $1::uuid", [
        userId,
      ]).catch(() => {});
    });

    // List users with tenant_id filter for the second tenant
    const listResult = await rawApiRequest({
      method: "GET",
      path: `/api/v1/admin/users?tenant_id=${tenantAdminInfo.tenantId}&search=${uniqueEmail}`,
      headers: { Authorization: `Bearer ${adminToken}` },
    });

    // The user should NOT appear when filtering by second tenant
    const users = (listResult.body?.users || listResult.body || []) as Array<{
      email: string;
    }>;
    const emails = users.map((u) => u.email);
    expect(emails).not.toContain(uniqueEmail);
  });

  test("stopping impersonation restores original context", async ({
    adminToken,
    tenantAdminInfo,
  }) => {
    // Start impersonation
    const impResult = await rawStartUserImpersonation(
      tenantAdminInfo.userId,
      "E2E stop test",
      adminToken,
    );
    expect(impResult.status).toBe(200);
    expect(impResult.body.access_token).toBeTruthy();

    // Stop impersonation
    const stopResult = await rawStopImpersonation(adminToken);
    expect(stopResult.status).toBeLessThan(300);

    // After stopping, the admin token should still work normally
    const tenantsResult = await rawApiRequest({
      method: "GET",
      path: "/api/v1/admin/tenants",
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    expect(tenantsResult.status).toBe(200);
    const tenants = (tenantsResult.body || []) as Array<{ id: string }>;
    expect(tenants.length).toBeGreaterThanOrEqual(1);
  });

  test("impersonated user cannot access admin-only endpoints", async ({
    adminToken,
    tenantAdminInfo,
  }) => {
    // Start impersonation as the tenant admin
    const impResult = await rawStartUserImpersonation(
      tenantAdminInfo.userId,
      "E2E admin endpoint test",
      adminToken,
    );
    expect(impResult.status).toBe(200);
    const impToken = impResult.body.access_token;

    try {
      // Try to list all tenants — should be forbidden for impersonated tenant admin
      const tenantsResult = await rawApiRequest({
        method: "GET",
        path: "/api/v1/admin/tenants",
        headers: { Authorization: `Bearer ${impToken}` },
      });
      // Tenant admin should not have instance admin access
      expect([403, 401]).toContain(tenantsResult.status);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("impersonated user in tenant A cannot create resources in tenant B", async ({
    adminToken,
    defaultTenantId,
    tenantAdminInfo,
  }) => {
    // Start impersonation as the tenant admin (belongs to second tenant)
    const impResult = await rawStartUserImpersonation(
      tenantAdminInfo.userId,
      "E2E cross-tenant create test",
      adminToken,
    );
    expect(impResult.status).toBe(200);
    const impToken = impResult.body.access_token;

    try {
      // Try to create a bucket in the default tenant using the impersonation token
      const crossTenantBucket = `imp-cross-${Date.now()}`;
      const createResult = await rawApiRequest({
        method: "POST",
        path: `/api/v1/storage/buckets/${crossTenantBucket}`,
        headers: {
          Authorization: `Bearer ${impToken}`,
          "X-FB-Tenant": defaultTenantId,
        },
      });

      // Should be rejected or silently scoped to the user's own tenant
      if (createResult.status < 300) {
        // If creation succeeded, it must have been created in the user's own tenant, not default
        const defaultBuckets = await rawListBuckets(
          adminToken,
          defaultTenantId,
        );
        const bucketList = (defaultBuckets.body?.buckets ||
          defaultBuckets.body ||
          []) as Array<{ id: string }>;
        const bucketIds = bucketList.map((b) => b.id);
        // Cleanup if it was created somewhere
        cleanupTasks.push(async () => {
          await rawApiRequest({
            method: "DELETE",
            path: `/api/v1/storage/buckets/${crossTenantBucket}`,
            headers: { Authorization: `Bearer ${adminToken}` },
          });
        });
        // The bucket should NOT appear in the default tenant's list
        expect(bucketIds).not.toContain(crossTenantBucket);
      } else {
        // Creation was rejected — expected behavior
        expect(createResult.status).toBeGreaterThanOrEqual(400);
      }
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });
});
