/**
 * Direct PostgreSQL database helpers for Playwright tests.
 * Used for verifying database state and cleaning up test data.
 */

import pg from "pg";

const DB_CONFIG = {
  host: process.env.FLUXBASE_DATABASE_HOST || "localhost",
  port: parseInt(process.env.FLUXBASE_DATABASE_PORT || "5432", 10),
  user: process.env.FLUXBASE_DATABASE_USER || "fluxbase_app",
  password: process.env.FLUXBASE_DATABASE_PASSWORD || "fluxbase_app_password",
  // Use PLAYWRIGHT_DATABASE_NAME if set (CI uses fluxbase_test),
  // otherwise default to fluxbase_playwright (local dev)
  database: process.env.PLAYWRIGHT_DATABASE_NAME || "fluxbase_playwright",
};

let pool: pg.Pool | null;

function getPool(): pg.Pool {
  if (!pool) {
    pool = new pg.Pool(DB_CONFIG);
  }
  return pool;
}

/**
 * Run a SQL query against the test database.
 */
export async function query<T extends pg.QueryResultRow = pg.QueryResultRow>(
  text: string,
  params?: unknown[],
): Promise<T[]> {
  const client = getPool();
  const result = await client.query<T>(text, params);
  return result.rows;
}

/**
 * Run a SQL command (INSERT, UPDATE, DELETE) and return affected row count.
 */
export async function execute(
  text: string,
  params?: unknown[],
): Promise<number> {
  const client = getPool();
  const result = await client.query(text, params);
  return result.rowCount;
}

/**
 * Clean up test data. Truncates key tables to restore a clean state.
 * Preserves the default tenant and admin user.
 */
export async function cleanupTestData(): Promise<void> {
  const client = getPool();
  // Delete in dependency order to respect foreign keys
  await client
    .query("DELETE FROM platform.tenant_admin_assignments")
    .catch(() => {});
  await client.query("DELETE FROM platform.tenant_memberships").catch(() => {});
  await client
    .query("DELETE FROM platform.service_keys WHERE tenant_id IS NOT NULL")
    .catch(() => {});
  await client.query("DELETE FROM platform.invitation_tokens").catch(() => {});
  await client
    .query("DELETE FROM platform.tenants WHERE is_default = false")
    .catch(() => {});
  // Clean storage objects and buckets
  await client.query("DELETE FROM storage.objects").catch(() => {});
  await client.query("DELETE FROM storage.buckets").catch(() => {});
}

/**
 * Get all tenants from the database.
 */
export async function getTenants(): Promise<
  Array<{
    id: string;
    name: string;
    slug: string;
    is_default: boolean;
    status: string;
  }>
> {
  return query<{
    id: string;
    name: string;
    slug: string;
    is_default: boolean;
    status: string;
  }>(
    "SELECT id, name, slug, is_default, status FROM platform.tenants WHERE deleted_at IS NULL ORDER BY created_at",
  );
}

/**
 * Get the default tenant from the database.
 */
export async function getDefaultTenant(): Promise<{
  id: string;
  name: string;
  slug: string;
  is_default: boolean;
  status: string;
} | null> {
  const rows = await query<{
    id: string;
    name: string;
    slug: string;
    is_default: boolean;
    status: string;
  }>(
    "SELECT id, name, slug, is_default, status FROM platform.tenants WHERE is_default = true AND deleted_at IS NULL",
  );
  return rows[0] || null;
}

/**
 * Get a platform user by email.
 */
export async function getUserByEmail(email: string): Promise<{
  id: string;
  email: string;
  role: string;
} | null> {
  const rows = await query<{
    id: string;
    email: string;
    role: string;
  }>("SELECT id, email, role FROM platform.users WHERE email = $1", [email]);
  return rows[0] || null;
}

/**
 * Create a platform user directly in the database.
 * Creates rows in both platform.users and auth.users (some API endpoints
 * like AssignAdmin check auth.users while login checks platform.users).
 * Returns the user ID.
 */
export async function createPlatformUser(
  email: string,
  password: string,
  name: string,
  role: string = "tenant_admin",
): Promise<string> {
  // Create in platform.users
  const rows = await query<{ id: string }>(
    `INSERT INTO platform.users (email, password_hash, full_name, role, email_verified)
     VALUES ($1, crypt($2, gen_salt('bf')), $3, $4, true)
     RETURNING id`,
    [email, password, name, role],
  );
  const userId = rows[0].id;

  // Also create in auth.users so API endpoints like AssignAdmin can find the user
  await execute(
    `INSERT INTO auth.users (id, email, password_hash, email_verified, role)
     VALUES ($1::uuid, $2, crypt($3, gen_salt('bf')), true, $4)
     ON CONFLICT (email) WHERE tenant_id IS NULL DO NOTHING`,
    [userId, email, password, role],
  );

  return userId;
}

/**
 * Close the database pool.
 */
export async function closePool(): Promise<void> {
  if (pool) {
    await pool.end();
    pool = null;
  }
}
