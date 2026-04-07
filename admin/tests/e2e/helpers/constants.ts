/**
 * Shared test constants for Playwright E2E tests.
 *
 * Single source of truth for credentials, tenant slugs, and other
 * constants used across fixtures, provisioning, and test files.
 */

const SETUP_TOKEN =
  process.env.FLUXBASE_SECURITY_SETUP_TOKEN ||
  "test-setup-token-for-dev-environment-32chars";

const ADMIN_EMAIL = "admin@fluxbase.test";
const ADMIN_PASSWORD = "test-password-32chars!!";
const ADMIN_NAME = "E2E Test Admin";

const TENANT_ADMIN_EMAIL = "tenant-admin@fluxbase.test";
const TENANT_ADMIN_PASSWORD = "tenant-admin-pass-32!!";
const TENANT_ADMIN_NAME = "Tenant Admin";

const SECOND_TENANT_NAME = "E2E Second Tenant";
const SECOND_TENANT_SLUG = "e2e-second-tenant";
const THIRD_TENANT_NAME = "E2E Third Tenant";
const THIRD_TENANT_SLUG = "e2e-third-tenant";

export {
  SETUP_TOKEN,
  ADMIN_EMAIL,
  ADMIN_PASSWORD,
  ADMIN_NAME,
  TENANT_ADMIN_EMAIL,
  TENANT_ADMIN_PASSWORD,
  TENANT_ADMIN_NAME,
  SECOND_TENANT_NAME,
  SECOND_TENANT_SLUG,
  THIRD_TENANT_NAME,
  THIRD_TENANT_SLUG,
};
