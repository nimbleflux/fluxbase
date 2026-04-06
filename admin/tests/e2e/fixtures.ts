/* eslint-disable react-hooks/rules-of-hooks, no-empty-pattern */
import { test as base, expect, type Page } from "@playwright/test";
import { rawLogin, listTenants } from "./helpers/api";
import { getUserByEmail } from "./helpers/db";

const SETUP_TOKEN =
  process.env.FLUXBASE_SECURITY_SETUP_TOKEN ||
  "test-setup-token-for-dev-environment-32chars";

const ADMIN_EMAIL = "admin@fluxbase.test";
const ADMIN_PASSWORD = "test-password-32chars!!";
const ADMIN_NAME = "Test Admin";

const BASE_URL =
  process.env.PLAYWRIGHT_BASE_URL || "http://localhost:5050/admin/";

interface TenantInfo {
  id: string;
  name: string;
  slug: string;
  is_default: boolean;
}

interface Fixtures {
  adminPage: Page;
  adminToken: string;
  tenantAdminPage: Page;
  tenantAdminToken: string;
  tenantAdminInfo: {
    userId: string;
    email: string;
    tenantId: string;
    tenantName: string;
  };
  multipleTenants: TenantInfo[];
  defaultTenantId: string;
  thirdTenantId: string;
}

const TENANT_ADMIN_EMAIL = "tenant-admin@fluxbase.test";
const TENANT_ADMIN_PASSWORD = "tenant-admin-pass-32!!";
const SECOND_TENANT_NAME = "E2E Second Tenant";
const SECOND_TENANT_SLUG = "e2e-second-tenant";
const THIRD_TENANT_SLUG = "e2e-third-tenant";

/**
 * Get an admin access token via direct API call (no browser needed).
 */
async function getAdminToken(): Promise<string> {
  const result = await rawLogin({
    email: ADMIN_EMAIL,
    password: ADMIN_PASSWORD,
  });
  if (result.status !== 200 || !result.body?.access_token) {
    throw new Error(
      `Failed to get admin token: ${result.status} ${JSON.stringify(result.body)}`,
    );
  }
  return result.body.access_token;
}

/**
 * Login a user via the browser and return the page.
 */
async function browserLogin(
  page: Page,
  email: string,
  password: string,
): Promise<Page> {
  await page.goto("login", { waitUntil: "networkidle" });
  await expect(page.locator("#email")).toBeVisible({ timeout: 15_000 });

  await page.fill("#email", email);
  await page.fill("#password", password);
  await page.click('button[type="submit"]');

  // Wait for dashboard to load
  await expect(page).toHaveURL(/\/admin\/?$/, { timeout: 10_000 });

  const token = await page.evaluate(() => {
    return localStorage.getItem("fluxbase_admin_access_token");
  });
  expect(token).toBeTruthy();

  return page;
}

export const test = base.extend<Fixtures>({
  adminPage: async ({ page }, use) => {
    await browserLogin(page, ADMIN_EMAIL, ADMIN_PASSWORD);
    await use(page);
  },

  adminToken: async ({}, use) => {
    const token = await getAdminToken();
    await use(token);
  },

  /**
   * Provides a tenant admin page (logged in as tenant admin).
   * Requires global setup to have created the second tenant + tenant admin user.
   */
  tenantAdminPage: async ({ page }, use) => {
    await browserLogin(page, TENANT_ADMIN_EMAIL, TENANT_ADMIN_PASSWORD);
    await use(page);
  },

  tenantAdminToken: async ({ tenantAdminPage }, use) => {
    const token = await tenantAdminPage.evaluate(() =>
      localStorage.getItem("fluxbase_admin_access_token"),
    );
    expect(token).toBeTruthy();
    await use(token || "");
  },

  tenantAdminInfo: async ({}, use) => {
    const user = await getUserByEmail(TENANT_ADMIN_EMAIL);
    if (!user) {
      throw new Error(
        "Tenant admin user not found. Global setup must run first.",
      );
    }

    const adminToken = await getAdminToken();
    const tenantsResult = await listTenants(adminToken);
    const secondTenant = (tenantsResult.body as TenantInfo[])?.find(
      (t) => t.slug === SECOND_TENANT_SLUG,
    );
    if (!secondTenant) {
      throw new Error("Second tenant not found. Global setup must run first.");
    }

    await use({
      userId: user.id,
      email: TENANT_ADMIN_EMAIL,
      tenantId: secondTenant.id,
      tenantName: secondTenant.name,
    });
  },

  /**
   * Provides a list of all tenants (ensures at least 2 exist).
   */
  multipleTenants: async ({ adminToken }, use) => {
    const tenantsResult = await listTenants(adminToken);
    const tenants = (tenantsResult.body as TenantInfo[]) || [];

    if (tenants.length < 2) {
      throw new Error(
        "Expected at least 2 tenants. Global setup must run first.",
      );
    }

    await use(tenants);
  },

  /**
   * Provides the default tenant ID.
   */
  defaultTenantId: async ({ adminToken }, use) => {
    const tenantsResult = await listTenants(adminToken);
    const defaultTenant = (tenantsResult.body as TenantInfo[])?.find(
      (t) => t.is_default === true,
    );
    if (!defaultTenant) {
      throw new Error("Default tenant not found.");
    }
    await use(defaultTenant.id);
  },

  /**
   * Provides the third tenant ID (for isolation tests).
   */
  thirdTenantId: async ({ adminToken }, use) => {
    const tenantsResult = await listTenants(adminToken);
    const thirdTenant = (tenantsResult.body as TenantInfo[])?.find(
      (t) => t.slug === THIRD_TENANT_SLUG,
    );
    if (!thirdTenant) {
      throw new Error("Third tenant not found. Global setup must run first.");
    }
    await use(thirdTenant.id);
  },
});

export {
  expect,
  ADMIN_EMAIL,
  ADMIN_PASSWORD,
  ADMIN_NAME,
  SETUP_TOKEN,
  BASE_URL,
  TENANT_ADMIN_EMAIL,
  TENANT_ADMIN_PASSWORD,
  SECOND_TENANT_NAME,
  SECOND_TENANT_SLUG,
  THIRD_TENANT_SLUG,
};
export type { TenantInfo };
