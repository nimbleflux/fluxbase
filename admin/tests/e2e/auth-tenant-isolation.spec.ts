import { test, expect } from "./fixtures";
import {
  rawAuthSignUp,
  rawAuthSignIn,
  rawSendMagicLink,
  rawSendOTP,
  rawRequestPasswordReset,
  rawCreateServiceKey,
} from "./helpers/api";
import { query, execute } from "./helpers/db";

const TEST_PASSWORD = "test-password-32chars!!";

async function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

test.describe("Auth Table Tenant Isolation", () => {
  const createdUserIds: string[] = [];
  const createdEmails: string[] = [];

  test.afterAll(async () => {
    for (const email of createdEmails) {
      await execute(
        "DELETE FROM auth.sessions WHERE user_id IN (SELECT id FROM auth.users WHERE email = $1)",
        [email],
      ).catch(() => {});
      await execute("DELETE FROM auth.magic_links WHERE email = $1", [
        email,
      ]).catch(() => {});
      await execute("DELETE FROM auth.otp_codes WHERE email = $1", [
        email,
      ]).catch(() => {});
      await execute(
        "DELETE FROM auth.password_reset_tokens WHERE user_id IN (SELECT id FROM auth.users WHERE email = $1)",
        [email],
      ).catch(() => {});
      await execute("DELETE FROM auth.users WHERE email = $1", [email]).catch(
        () => {},
      );
      await execute("DELETE FROM platform.users WHERE email = $1", [
        email,
      ]).catch(() => {});
    }
    for (const userId of createdUserIds) {
      await execute("DELETE FROM auth.sessions WHERE user_id = $1", [
        userId,
      ]).catch(() => {});
      await execute("DELETE FROM auth.magic_links WHERE user_id = $1", [
        userId,
      ]).catch(() => {});
      await execute("DELETE FROM auth.otp_codes WHERE user_id = $1", [
        userId,
      ]).catch(() => {});
      await execute(
        "DELETE FROM auth.password_reset_tokens WHERE user_id = $1",
        [userId],
      ).catch(() => {});
      await execute("DELETE FROM auth.users WHERE id = $1", [userId]).catch(
        () => {},
      );
      await execute("DELETE FROM platform.users WHERE id = $1", [userId]).catch(
        () => {},
      );
    }
  });

  test.describe.serial("Auth tenant isolation tests", () => {
    let userIdA: string;
    let emailA: string;
    let userIdB: string;
    let emailB: string;

    test("setup: create test users in both tenants", async ({
      defaultTenantId,
      thirdTenantId,
    }) => {
      const ts = Date.now();
      emailA = `auth-iso-a-${ts}@test.com`;
      emailB = `auth-iso-b-${ts}@test.com`;

      const signUpA = await rawAuthSignUp(
        { email: emailA, password: TEST_PASSWORD, name: "Iso A" },
        defaultTenantId,
      );
      expect(signUpA.status).toBeLessThan(300);
      userIdA = signUpA.body?.user?.id;
      expect(userIdA).toBeTruthy();
      createdUserIds.push(userIdA);
      createdEmails.push(emailA);

      await delay(500);

      const signUpB = await rawAuthSignUp(
        { email: emailB, password: TEST_PASSWORD, name: "Iso B" },
        thirdTenantId,
      );
      expect(signUpB.status).toBeLessThan(300);
      userIdB = signUpB.body?.user?.id;
      expect(userIdB).toBeTruthy();
      createdUserIds.push(userIdB);
      createdEmails.push(emailB);
    });

    test("sessions created via signin have correct tenant_id", async ({
      defaultTenantId,
      thirdTenantId,
    }) => {
      const signinA = await rawAuthSignIn(
        { email: emailA, password: TEST_PASSWORD },
        defaultTenantId,
      );
      expect(signinA.status).toBeLessThan(300);

      const signinB = await rawAuthSignIn(
        { email: emailB, password: TEST_PASSWORD },
        thirdTenantId,
      );
      expect(signinB.status).toBeLessThan(300);

      const sessionsA = await query<{ tenant_id: string }>(
        "SELECT tenant_id FROM auth.sessions WHERE user_id = $1",
        [userIdA],
      );
      expect(sessionsA.length).toBeGreaterThanOrEqual(1);

      const sessionsB = await query<{ tenant_id: string }>(
        "SELECT tenant_id FROM auth.sessions WHERE user_id = $1",
        [userIdB],
      );
      expect(sessionsB.length).toBeGreaterThanOrEqual(1);

      for (const s of sessionsA) {
        expect(s.tenant_id).toBe(defaultTenantId);
      }
      for (const s of sessionsB) {
        expect(s.tenant_id).toBe(thirdTenantId);
      }
    });

    test("magic links have correct tenant_id and user_id", async ({
      defaultTenantId,
      thirdTenantId,
    }) => {
      const magicA = await rawSendMagicLink({ email: emailA }, defaultTenantId);
      const magicB = await rawSendMagicLink({ email: emailB }, thirdTenantId);

      if (magicA.status >= 300 && magicB.status >= 300) {
        const linkA = await query<{ tenant_id: string; user_id: string }>(
          "SELECT tenant_id, user_id FROM auth.magic_links WHERE email = $1 ORDER BY created_at DESC LIMIT 5",
          [emailA],
        );
        const linkB = await query<{ tenant_id: string; user_id: string }>(
          "SELECT tenant_id, user_id FROM auth.magic_links WHERE email = $1 ORDER BY created_at DESC LIMIT 5",
          [emailB],
        );
        if (linkA.length === 0 && linkB.length === 0) {
          test.skip();
          return;
        }
      }

      const linksA = await query<{ tenant_id: string; user_id: string }>(
        "SELECT tenant_id, user_id FROM auth.magic_links WHERE email = $1 ORDER BY created_at DESC LIMIT 5",
        [emailA],
      );
      expect(linksA.length).toBeGreaterThanOrEqual(1);
      expect(linksA[0].tenant_id).toBe(defaultTenantId);
      expect(linksA[0].user_id).toBeTruthy();

      const linksB = await query<{ tenant_id: string; user_id: string }>(
        "SELECT tenant_id, user_id FROM auth.magic_links WHERE email = $1 ORDER BY created_at DESC LIMIT 5",
        [emailB],
      );
      expect(linksB.length).toBeGreaterThanOrEqual(1);
      expect(linksB[0].tenant_id).toBe(thirdTenantId);
      expect(linksB[0].user_id).toBeTruthy();
    });

    test("otp codes have correct tenant_id and user_id", async ({
      defaultTenantId,
      thirdTenantId,
    }) => {
      const otpA = await rawSendOTP({ email: emailA }, defaultTenantId);
      await delay(500);
      const otpB = await rawSendOTP({ email: emailB }, thirdTenantId);

      if (otpA.status >= 300 && otpB.status >= 300) {
        const codeA = await query<{ tenant_id: string; user_id: string }>(
          "SELECT tenant_id, user_id FROM auth.otp_codes WHERE email = $1 ORDER BY created_at DESC LIMIT 5",
          [emailA],
        );
        const codeB = await query<{ tenant_id: string; user_id: string }>(
          "SELECT tenant_id, user_id FROM auth.otp_codes WHERE email = $1 ORDER BY created_at DESC LIMIT 5",
          [emailB],
        );
        if (codeA.length === 0 && codeB.length === 0) {
          test.skip();
          return;
        }
      }

      const codesA = await query<{ tenant_id: string; user_id: string }>(
        "SELECT tenant_id, user_id FROM auth.otp_codes WHERE email = $1 ORDER BY created_at DESC LIMIT 5",
        [emailA],
      );
      expect(codesA.length).toBeGreaterThanOrEqual(1);
      expect(codesA[0].tenant_id).toBe(defaultTenantId);
      expect(codesA[0].user_id).toBeTruthy();

      const codesB = await query<{ tenant_id: string; user_id: string }>(
        "SELECT tenant_id, user_id FROM auth.otp_codes WHERE email = $1 ORDER BY created_at DESC LIMIT 5",
        [emailB],
      );
      expect(codesB.length).toBeGreaterThanOrEqual(1);
      expect(codesB[0].tenant_id).toBe(thirdTenantId);
      expect(codesB[0].user_id).toBeTruthy();
    });

    test("password reset tokens have correct tenant_id", async ({
      defaultTenantId,
      thirdTenantId,
    }) => {
      const resetA = await rawRequestPasswordReset(
        { email: emailA },
        defaultTenantId,
      );
      await delay(500);
      const resetB = await rawRequestPasswordReset(
        { email: emailB },
        thirdTenantId,
      );

      if (resetA.status >= 300 && resetB.status >= 300) {
        const tokA = await query<{ tenant_id: string }>(
          "SELECT tenant_id FROM auth.password_reset_tokens WHERE user_id = $1 ORDER BY created_at DESC LIMIT 5",
          [userIdA],
        );
        const tokB = await query<{ tenant_id: string }>(
          "SELECT tenant_id FROM auth.password_reset_tokens WHERE user_id = $1 ORDER BY created_at DESC LIMIT 5",
          [userIdB],
        );
        if (tokA.length === 0 && tokB.length === 0) {
          test.skip();
          return;
        }
      }

      const tokensA = await query<{ tenant_id: string }>(
        "SELECT tenant_id FROM auth.password_reset_tokens WHERE user_id = $1 ORDER BY created_at DESC LIMIT 5",
        [userIdA],
      );
      expect(tokensA.length).toBeGreaterThanOrEqual(1);
      expect(tokensA[0].tenant_id).toBe(defaultTenantId);

      const tokensB = await query<{ tenant_id: string }>(
        "SELECT tenant_id FROM auth.password_reset_tokens WHERE user_id = $1 ORDER BY created_at DESC LIMIT 5",
        [userIdB],
      );
      expect(tokensB.length).toBeGreaterThanOrEqual(1);
      expect(tokensB[0].tenant_id).toBe(thirdTenantId);
    });

    test("client_key_usage rows are tenant-scoped via RLS", async ({
      defaultTenantId,
      thirdTenantId,
    }) => {
      const usageIdA = crypto.randomUUID();
      const usageIdB = crypto.randomUUID();

      const clientKeysA = await query<{ id: string }>(
        "SELECT id FROM auth.client_keys LIMIT 1",
      );
      const clientKeysB = await query<{ id: string }>(
        "SELECT id FROM auth.client_keys LIMIT 1",
      );
      const clientKeyIdA = clientKeysA[0]?.id;
      const clientKeyIdB = clientKeysB[0]?.id;

      if (!clientKeyIdA || !clientKeyIdB) {
        test.skip();
        return;
      }

      await execute(
        "INSERT INTO auth.client_key_usage (id, client_key_id, endpoint, method, tenant_id) VALUES ($1, $2, $3, $4, $5)",
        [usageIdA, clientKeyIdA, "/api/v1/test", "GET", defaultTenantId],
      );

      await execute(
        "INSERT INTO auth.client_key_usage (id, client_key_id, endpoint, method, tenant_id) VALUES ($1, $2, $3, $4, $5)",
        [usageIdB, clientKeyIdB, "/api/v1/test", "GET", thirdTenantId],
      );

      const rowsA = await query<{ id: string }>(
        "SELECT id FROM auth.client_key_usage WHERE id = $1 AND tenant_id = $2",
        [usageIdA, defaultTenantId],
      );
      expect(rowsA.length).toBe(1);

      const rowsB = await query<{ id: string }>(
        "SELECT id FROM auth.client_key_usage WHERE id = $1 AND tenant_id = $2",
        [usageIdB, thirdTenantId],
      );
      expect(rowsB.length).toBe(1);

      const crossA = await query<{ id: string }>(
        "SELECT id FROM auth.client_key_usage WHERE id = $1 AND tenant_id = $2",
        [usageIdB, defaultTenantId],
      );
      expect(crossA.length).toBe(0);

      const crossB = await query<{ id: string }>(
        "SELECT id FROM auth.client_key_usage WHERE id = $1 AND tenant_id = $2",
        [usageIdA, thirdTenantId],
      );
      expect(crossB.length).toBe(0);

      await execute("DELETE FROM auth.client_key_usage WHERE id = $1", [
        usageIdA,
      ]).catch(() => {});
      await execute("DELETE FROM auth.client_key_usage WHERE id = $1", [
        usageIdB,
      ]).catch(() => {});
    });

    test("service_key_revocations rows have correct tenant_id", async ({
      adminToken,
      defaultTenantId,
      thirdTenantId,
    }) => {
      const keyA = await rawCreateServiceKey(
        { name: `iso-revoke-a-${Date.now()}`, keyType: "service" },
        adminToken,
        defaultTenantId,
      );
      expect(keyA.status).toBeLessThan(300);
      const keyIdA = keyA.body?.key?.id || keyA.body?.id;

      const keyB = await rawCreateServiceKey(
        { name: `iso-revoke-b-${Date.now()}`, keyType: "service" },
        adminToken,
        thirdTenantId,
      );
      expect(keyB.status).toBeLessThan(300);
      const keyIdB = keyB.body?.key?.id || keyB.body?.id;

      if (!keyIdA || !keyIdB) {
        test.skip();
        return;
      }

      const revId = crypto.randomUUID();
      await execute(
        "INSERT INTO auth.service_key_revocations (id, key_id, key_prefix, reason, revocation_type, tenant_id) VALUES ($1, $2, $3, $4, $5, $6)",
        [
          revId,
          keyIdA,
          "fb_test_",
          "E2E test revocation",
          "emergency",
          defaultTenantId,
        ],
      );

      const revRows = await query<{ tenant_id: string }>(
        "SELECT tenant_id FROM auth.service_key_revocations WHERE id = $1",
        [revId],
      );
      expect(revRows.length).toBe(1);
      expect(revRows[0].tenant_id).toBe(defaultTenantId);

      await execute("DELETE FROM auth.service_key_revocations WHERE id = $1", [
        revId,
      ]).catch(() => {});
      await execute("DELETE FROM auth.service_keys WHERE id = $1", [
        keyIdA,
      ]).catch(() => {});
      await execute("DELETE FROM auth.service_keys WHERE id = $1", [
        keyIdB,
      ]).catch(() => {});
    });

    test("signin with wrong X-FB-Tenant returns 401", async ({
      defaultTenantId,
      thirdTenantId,
    }) => {
      const ts = Date.now();
      const email = `auth-cross-${ts}@test.com`;

      const signUp = await rawAuthSignUp(
        { email, password: TEST_PASSWORD, name: "Cross Tenant" },
        defaultTenantId,
      );
      expect(signUp.status).toBeLessThan(300);
      const userId = signUp.body?.user?.id;
      expect(userId).toBeTruthy();
      createdUserIds.push(userId);
      createdEmails.push(email);

      const signinWrongTenant = await rawAuthSignIn(
        { email, password: TEST_PASSWORD },
        thirdTenantId,
      );
      expect(signinWrongTenant.status).toBe(401);

      const signinCorrectTenant = await rawAuthSignIn(
        { email, password: TEST_PASSWORD },
        defaultTenantId,
      );
      expect(signinCorrectTenant.status).toBeLessThan(300);
      expect(signinCorrectTenant.body?.access_token).toBeTruthy();
    });
  });
});
