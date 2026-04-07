/* eslint-disable no-console */
/**
 * Provisioning for E2E tests.
 *
 * Runs after setup.spec.ts (which creates the admin user) and before all
 * other E2E tests. Creates tenants, tenant admin user, and cleans up
 * MailHog. If this test fails, the "e2e" project is skipped entirely.
 */
import { test, expect } from "@playwright/test";
import { rawLogin, rawCreateTenant, listTenants } from "./helpers/api";
import { deleteAllEmails } from "./helpers/mailhog";
import {
  createPlatformUser,
  getUserByEmail,
  query as dbQuery,
} from "./helpers/db";
import {
  ADMIN_EMAIL,
  ADMIN_PASSWORD,
  TENANT_ADMIN_EMAIL,
  TENANT_ADMIN_PASSWORD,
  TENANT_ADMIN_NAME,
  SECOND_TENANT_NAME,
  SECOND_TENANT_SLUG,
  THIRD_TENANT_NAME,
  THIRD_TENANT_SLUG,
} from "./helpers/constants";

test.describe("Provisioning", () => {
  test("provision test data for E2E tests", async () => {
    // --- Step 1: Verify setup is complete by logging in ---
    const loginResult = await rawLogin({
      email: ADMIN_EMAIL,
      password: ADMIN_PASSWORD,
    });
    expect(loginResult.status).toBe(200);
    expect(loginResult.body?.access_token).toBeTruthy();
    const adminToken = loginResult.body.access_token;
    console.log("Provisioning: admin login verified.");

    // --- Step 2: Create second tenant ---
    const tenantsResult = await listTenants(adminToken);
    const tenants = tenantsResult.body;
    const existingSecond = tenants?.find(
      (t: { slug: string }) => t.slug === SECOND_TENANT_SLUG,
    );

    let secondTenantId: string;
    if (existingSecond) {
      secondTenantId = existingSecond.id;
      console.log(
        `Provisioning: second tenant already exists (${secondTenantId}).`,
      );
    } else {
      const createResult = await rawCreateTenant(
        {
          name: SECOND_TENANT_NAME,
          slug: SECOND_TENANT_SLUG,
          autoGenerateKeys: true,
        },
        adminToken,
      );
      expect(createResult.status).toBeOneOf([200, 201]);
      secondTenantId = createResult.body.tenant.id;
      console.log(`Provisioning: second tenant created (${secondTenantId}).`);
    }

    // --- Step 3: Create third tenant (for isolation tests) ---
    const existingThird = tenants?.find(
      (t: { slug: string }) => t.slug === THIRD_TENANT_SLUG,
    );

    if (!existingThird) {
      const createResult = await rawCreateTenant(
        {
          name: THIRD_TENANT_NAME,
          slug: THIRD_TENANT_SLUG,
          autoGenerateKeys: false,
        },
        adminToken,
      );
      if (createResult.status === 200 || createResult.status === 201) {
        console.log(
          `Provisioning: third tenant created (${createResult.body.tenant?.id}).`,
        );
      } else {
        console.log(
          `Provisioning: warning — third tenant creation failed (${createResult.status}).`,
        );
      }
    } else {
      console.log(
        `Provisioning: third tenant already exists (${existingThird.id}).`,
      );
    }

    // --- Step 4: Create tenant admin user ---
    let tenantAdminUserId: string;
    const existingUser = await getUserByEmail(TENANT_ADMIN_EMAIL);
    if (existingUser) {
      tenantAdminUserId = existingUser.id;
      console.log(
        `Provisioning: tenant admin user already exists (${tenantAdminUserId}).`,
      );
    } else {
      tenantAdminUserId = await createPlatformUser(
        TENANT_ADMIN_EMAIL,
        TENANT_ADMIN_PASSWORD,
        TENANT_ADMIN_NAME,
        "tenant_admin",
      );
      console.log(
        `Provisioning: tenant admin user created (${tenantAdminUserId}).`,
      );
    }

    // --- Step 5: Assign tenant admin to the second tenant ---
    const assignmentRows = await dbQuery<{ id: string }>(
      `SELECT id FROM platform.tenant_admin_assignments
       WHERE user_id = $1::uuid AND tenant_id = $2::uuid`,
      [tenantAdminUserId, secondTenantId],
    );

    if (assignmentRows.length === 0) {
      const adminRows = await dbQuery<{ id: string }>(
        "SELECT id FROM platform.users WHERE email = $1",
        [ADMIN_EMAIL],
      );
      const assignedBy = adminRows[0]?.id || null;

      await dbQuery(
        `INSERT INTO platform.tenant_admin_assignments (user_id, tenant_id, assigned_by)
         VALUES ($1::uuid, $2::uuid, $3::uuid)
         ON CONFLICT (user_id, tenant_id) DO NOTHING`,
        [tenantAdminUserId, secondTenantId, assignedBy],
      );
      console.log("Provisioning: tenant admin assigned to second tenant.");
    } else {
      console.log(
        "Provisioning: tenant admin already assigned to second tenant.",
      );
    }

    // --- Step 6: Clean up MailHog ---
    await deleteAllEmails().catch(() => {
      console.log("Note: MailHog not available for cleanup.");
    });

    console.log("Provisioning complete.");
  });
});
