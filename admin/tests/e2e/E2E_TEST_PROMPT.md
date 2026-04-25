# E2E Tests: Auth Table Tenant Isolation

## Context

We added `tenant_id` columns, RLS policies, triggers, and `tenant_service` GRANTs to 17 auth tables that previously lacked tenant scoping. The database schema changes are already committed and deployed. What's missing are E2E tests that verify tenant isolation works correctly for these tables.

## The 17 Tables

| Table                               | tenant_id Source                      | API Endpoint to Trigger                                      |
| ----------------------------------- | ------------------------------------- | ------------------------------------------------------------ |
| `auth.sessions`                     | user_id → users.tenant_id             | `POST /api/v1/auth/signin`                                   |
| `auth.oauth_links`                  | user_id → users.tenant_id             | OAuth callback flow                                          |
| `auth.oauth_tokens`                 | user_id → users.tenant_id             | OAuth callback flow                                          |
| `auth.mfa_factors`                  | user_id → users.tenant_id             | `POST /api/v1/auth/mfa/enroll`                               |
| `auth.saml_sessions`                | user_id → users.tenant_id             | SAML callback flow                                           |
| `auth.magic_links`                  | user_id (new FK column)               | `POST /api/v1/auth/magiclink`                                |
| `auth.otp_codes`                    | user_id (new FK column)               | `POST /api/v1/auth/otp/signin`                               |
| `auth.email_verification_tokens`    | user_id → users.tenant_id             | `POST /api/v1/auth/signup` (when email verification enabled) |
| `auth.password_reset_tokens`        | user_id → users.tenant_id             | `POST /api/v1/auth/password/reset`                           |
| `auth.two_factor_setups`            | user_id → users.tenant_id             | 2FA setup flow                                               |
| `auth.two_factor_recovery_attempts` | user_id → users.tenant_id             | 2FA recovery flow                                            |
| `auth.oauth_logout_states`          | user_id → users.tenant_id             | `POST /api/v1/auth/oauth/:provider/logout`                   |
| `auth.mcp_oauth_clients`            | registered_by → users.tenant_id       | MCP OAuth client registration                                |
| `auth.mcp_oauth_codes`              | user_id → users.tenant_id             | MCP OAuth flow                                               |
| `auth.mcp_oauth_tokens`             | user_id → users.tenant_id             | MCP OAuth flow                                               |
| `auth.client_key_usage`             | client_key_id → client_keys.tenant_id | Client key API call                                          |
| `auth.service_key_revocations`      | key_id → service_keys.tenant_id       | Service key revocation                                       |

## What to Test

### Test File: `admin/tests/e2e/auth-tenant-isolation.spec.ts`

Write a single comprehensive E2E test file that verifies tenant isolation for the auth tables. The tests should follow the patterns established in:

- `tenant-resource-isolation.spec.ts` — for the bidirectional A↔B isolation pattern
- `tenant-service-admin-isolation.spec.ts` — for tenant admin visibility tests

### Test Strategy

**Primary approach: Direct database verification.** Since most of these tables are populated as side effects of auth flows (not directly via CRUD APIs), the most reliable test approach is:

1. **Use the existing API endpoints** to trigger creation of records in these tables
2. **Use the `helpers/db.ts` direct PostgreSQL helpers** to verify `tenant_id` is correctly set
3. **Use `helpers/api.ts` rawApiRequest** with different `X-FB-Tenant` headers to verify isolation

### Required Tests

#### 1. Sessions Tenant Isolation

```
- Create users in two different tenants (default and second)
- Sign in as each user
- Verify auth.sessions rows have correct tenant_id (via direct DB query)
- Verify sessions are scoped: querying with tenant A context doesn't show tenant B sessions
```

Use `POST /api/v1/auth/signin` with `{ email, password }` to create sessions.

#### 2. Magic Links Tenant Isolation

```
- Create a user in tenant A and a user in tenant B
- Send magic link for tenant A user (POST /api/v1/auth/magiclink with X-FB-Tenant: A)
- Send magic link for tenant B user (POST /api/v1/auth/magiclink with X-FB-Tenant: B)
- Verify via direct DB: auth.magic_links rows have correct tenant_id
- Verify via direct DB: magic_links.user_id is populated (new FK column)
- Verify tenant A magic links are not visible when querying with tenant B context
```

#### 3. OTP Codes Tenant Isolation

```
- Send OTP for a user in tenant A (POST /api/v1/auth/otp/signin with X-FB-Tenant: A)
- Send OTP for a user in tenant B (POST /api/v1/auth/otp/signin with X-FB-Tenant: B)
- Verify via direct DB: auth.otp_codes rows have correct tenant_id
- Verify via direct DB: otp_codes.user_id is populated (new FK column)
```

#### 4. Password Reset Tokens Tenant Isolation

```
- Request password reset for a user in tenant A (POST /api/v1/auth/password/reset with X-FB-Tenant: A)
- Request password reset for a user in tenant B (POST /api/v1/auth/password/reset with X-FB-Tenant: B)
- Verify via direct DB: auth.password_reset_tokens rows have correct tenant_id
```

#### 5. Client Key Usage Tenant Isolation

```
- Create a client key in tenant A
- Create a client key in tenant B
- Make an API call using tenant A's client key
- Make an API call using tenant B's client key
- Verify via direct DB: auth.client_key_usage rows have correct tenant_id
- Verify tenant A client_key_usage rows are not visible in tenant B's context
```

#### 6. Service Key Revocations Tenant Isolation

```
- Create service keys in tenant A and tenant B
- Revoke a key in tenant A (DELETE /api/v1/admin/service-keys/:id with X-FB-Tenant: A)
- Verify via direct DB: auth.service_key_revocations rows have correct tenant_id
```

#### 7. Cross-Tenant Signin Blocked

```
- Create a user in tenant A
- Attempt to sign in with that user's credentials while X-FB-Tenant is set to tenant B
- Verify the signin fails (401) because RLS prevents finding the user in tenant B's context
```

This is the most important test — it proves the entire chain works: tenant_id in auth.users → RLS policy → signin fails for wrong tenant.

### Test Infrastructure

#### Test Users

The provisioning script (`_provisioning.spec.ts`) already creates:

- `admin@fluxbase.test` — instance admin
- `tenant-admin@fluxbase.test` — tenant admin for `e2e-second-tenant`

You'll likely need to create additional test users directly via the API or DB for different tenants. Follow the pattern:

```typescript
// Create a user in a specific tenant
const signUpResult = await rawApiRequest({
  method: "POST",
  path: "/api/v1/auth/signup",
  data: {
    email: "test-user-a@test.com",
    password: "test-password-32chars!!",
    name: "Test User A",
  },
  headers: { "X-FB-Tenant": tenantId },
});
```

#### DB Verification Pattern

```typescript
import { query } from "./helpers/db";

// Verify tenant_id is set correctly
const rows = await query(
  "SELECT tenant_id, user_id FROM auth.magic_links WHERE email = $1",
  [email],
);
expect(rows[0].tenant_id).toBe(tenantAId);
expect(rows[0].user_id).toBeTruthy();
```

#### Cross-Tenant Query Verification

To verify RLS prevents cross-tenant access, query using `tenant_service` role:

```typescript
// Set tenant context and verify rows are scoped
const rows = await query(
  "SET app.current_tenant_id = $1; SELECT count(*) FROM auth.sessions WHERE tenant_id = $2",
  [tenantAId, tenantBId],
);
// Should be 0 — no tenant B sessions visible when tenant A is set
```

Or more practically, use the API with different tenant headers:

```typescript
// Sign in with tenant A's user, but send X-FB-Tenant: B
const result = await rawApiRequest({
  method: "POST",
  path: "/api/v1/auth/signin",
  data: { email: userAEmail, password: userAPassword },
  headers: { "X-FB-Tenant": tenantBId },
});
expect(result.status).toBe(401); // User not found in tenant B
```

### Conventions

1. **Use `test.describe`** to group related tests
2. **Track created resources** in `createdResources` array for `afterAll` cleanup
3. **Use `rawApiRequest`** from `helpers/api.ts` for HTTP calls (not browser)
4. **Use `query`** from `helpers/db.ts` for direct DB verification
5. **Use timestamps** in unique identifiers: `test-${Date.now()}`
6. **Import from `fixtures.ts`**: `adminToken`, `defaultTenantId`, `thirdTenantId`
7. **All tests should pass independently** — no ordering dependencies

### File Locations

- New test file: `admin/tests/e2e/auth-tenant-isolation.spec.ts`
- API helpers: `admin/tests/e2e/helpers/api.ts` (add new helpers if needed)
- DB helpers: `admin/tests/e2e/helpers/db.ts` (already has `query` and `execute`)
- Constants: `admin/tests/e2e/helpers/constants.ts`
- Fixtures: `admin/tests/e2e/fixtures.ts`

### Running the Tests

```bash
# Start test servers
make test-e2e-ui-server  # Go :8082 + Vite :5050

# Run the new tests (in another terminal)
npx playwright test admin/tests/e2e/auth-tenant-isolation.spec.ts --reporter=list

# Or run all E2E tests
make test-e2e-ui
```

### Important Notes

- The database is `fluxbase_playwright` (not the dev database)
- Tests run against Go backend on port 8082, proxied through Vite on port 5050
- The `_provisioning.spec.ts` test must run first to create tenants and test users
- Some auth flows (OAuth, SAML, MCP OAuth) are difficult to test via E2E without mock providers. Focus on the tables that can be tested via direct API calls (sessions, magic_links, otp_codes, password_reset_tokens, client_key_usage, service_key_revocations). For OAuth/SAML/MCP tables, use direct DB INSERT verification instead of going through the full flow.
- The `tenant_id` values in the database are UUIDs (strings), not integers

### Acceptance Criteria

- [ ] New test file `auth-tenant-isolation.spec.ts` exists and follows existing patterns
- [ ] Sessions created via signin have correct tenant_id
- [ ] Magic links have correct tenant_id AND user_id populated
- [ ] OTP codes have correct tenant_id AND user_id populated
- [ ] Password reset tokens have correct tenant_id
- [ ] Client key usage rows have correct tenant_id
- [ ] Service key revocations have correct tenant_id
- [ ] Cross-tenant signin is blocked (401 for wrong tenant context)
- [ ] All existing tests still pass
- [ ] `make test-e2e-ui` passes with the new tests included
