import { type APIRequestContext } from "@playwright/test";

const BASE_URL =
  process.env.PLAYWRIGHT_API_URL ||
  (process.env.CI
    ? `http://localhost:${process.env.PLAYWRIGHT_BACKEND_PORT || "8082"}`
    : "http://localhost:5050");

/**
 * Make an API request to the Fluxbase server.
 */
export async function apiRequest(
  request: APIRequestContext,
  options: {
    method: string;
    path: string;
    data?: unknown;
    headers?: Record<string, string>;
  },
) {
  const response = await request.fetch(`${BASE_URL}${options.path}`, {
    method: options.method,
    data: options.data ? JSON.stringify(options.data) : undefined,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  return {
    status: response.status(),
    body: await response.json().catch(() => null),
    headers: response.headers(),
  };
}

/**
 * Make a raw API request (without Playwright request context).
 * Uses native fetch — suitable for global setup and fixtures.
 */
export async function rawApiRequest(options: {
  method: string;
  path: string;
  data?: unknown;
  headers?: Record<string, string>;
}) {
  const response = await fetch(`${BASE_URL}${options.path}`, {
    method: options.method,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    body: options.data ? JSON.stringify(options.data) : undefined,
  });

  return {
    status: response.status,
    body: await response.json().catch(() => null),
    headers: Object.fromEntries(response.headers.entries()),
  };
}

/**
 * Perform initial setup via the API (creates admin user).
 */
export async function performSetup(
  request: APIRequestContext,
  options: {
    setupToken: string;
    name: string;
    email: string;
    password: string;
  },
) {
  return apiRequest(request, {
    method: "POST",
    path: "/api/v1/admin/setup",
    data: {
      setup_token: options.setupToken,
      name: options.name,
      email: options.email,
      password: options.password,
    },
  });
}

/**
 * Login and get access/refresh tokens.
 */
export async function login(
  request: APIRequestContext,
  options: {
    email: string;
    password: string;
  },
) {
  return apiRequest(request, {
    method: "POST",
    path: "/dashboard/auth/login",
    data: {
      email: options.email,
      password: options.password,
    },
  });
}

/**
 * Raw login using native fetch (for use outside Playwright test context).
 */
export async function rawLogin(options: { email: string; password: string }) {
  return rawApiRequest({
    method: "POST",
    path: "/dashboard/auth/login",
    data: { email: options.email, password: options.password },
  });
}

/**
 * Check if setup has been completed.
 */
export async function checkSetupStatus(request: APIRequestContext) {
  return apiRequest(request, {
    method: "GET",
    path: "/api/v1/admin/setup/status",
  });
}

/**
 * List all tenants via the API.
 */
export async function listTenants(accessToken: string) {
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/admin/tenants",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Create a tenant via the API.
 */
export async function createTenant(
  request: APIRequestContext,
  options: {
    name: string;
    slug: string;
    autoGenerateKeys?: boolean;
    adminEmail?: string;
  },
  accessToken: string,
) {
  return apiRequest(request, {
    method: "POST",
    path: "/api/v1/admin/tenants",
    data: {
      name: options.name,
      slug: options.slug,
      auto_generate_keys: options.autoGenerateKeys ?? true,
      admin_email: options.adminEmail,
    },
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Create a tenant using native fetch.
 */
export async function rawCreateTenant(
  options: {
    name: string;
    slug: string;
    autoGenerateKeys?: boolean;
    adminEmail?: string;
  },
  accessToken: string,
) {
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/admin/tenants",
    data: {
      name: options.name,
      slug: options.slug,
      auto_generate_keys: options.autoGenerateKeys ?? true,
      admin_email: options.adminEmail,
    },
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Delete a tenant via the API.
 */
export async function deleteTenant(
  request: APIRequestContext,
  tenantId: string,
  accessToken: string,
) {
  return apiRequest(request, {
    method: "DELETE",
    path: `/api/v1/admin/tenants/${tenantId}`,
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Delete a tenant using native fetch.
 */
export async function rawDeleteTenant(tenantId: string, accessToken: string) {
  return rawApiRequest({
    method: "DELETE",
    path: `/api/v1/admin/tenants/${tenantId}`,
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Update a tenant via the API.
 */
export async function updateTenant(
  request: APIRequestContext,
  tenantId: string,
  data: { name?: string; metadata?: Record<string, unknown> },
  accessToken: string,
) {
  return apiRequest(request, {
    method: "PATCH",
    path: `/api/v1/admin/tenants/${tenantId}`,
    data,
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Update a tenant using native fetch.
 */
export async function rawUpdateTenant(
  tenantId: string,
  data: { name?: string; metadata?: Record<string, unknown> },
  accessToken: string,
) {
  return rawApiRequest({
    method: "PATCH",
    path: `/api/v1/admin/tenants/${tenantId}`,
    data,
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * List tenant admins/members.
 */
export async function listTenantAdmins(
  request: APIRequestContext,
  tenantId: string,
  accessToken: string,
) {
  return apiRequest(request, {
    method: "GET",
    path: `/api/v1/admin/tenants/${tenantId}/admins`,
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * List tenant admins using native fetch.
 */
export async function rawListTenantAdmins(
  tenantId: string,
  accessToken: string,
) {
  return rawApiRequest({
    method: "GET",
    path: `/api/v1/admin/tenants/${tenantId}/admins`,
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Assign an admin to a tenant.
 */
export async function assignTenantAdmin(
  request: APIRequestContext,
  tenantId: string,
  userId: string,
  accessToken: string,
) {
  return apiRequest(request, {
    method: "POST",
    path: `/api/v1/admin/tenants/${tenantId}/admins`,
    data: { user_id: userId },
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * List service keys for the current tenant.
 */
export async function listServiceKeys(
  request: APIRequestContext,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return apiRequest(request, {
    method: "GET",
    path: "/api/v1/admin/service-keys",
    headers,
  });
}

/**
 * List service keys using native fetch.
 */
export async function rawListServiceKeys(
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/admin/service-keys",
    headers,
  });
}

/**
 * Create a service key.
 */
export async function createServiceKey(
  request: APIRequestContext,
  options: {
    name: string;
    keyType: "anon" | "service";
    tenantId?: string;
  },
  accessToken: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (options.tenantId) {
    headers["X-FB-Tenant"] = options.tenantId;
  }
  return apiRequest(request, {
    method: "POST",
    path: "/api/v1/admin/service-keys",
    data: {
      name: options.name,
      key_type: options.keyType,
    },
    headers,
  });
}

/**
 * Create a service key using native fetch.
 */
export async function rawCreateServiceKey(
  options: {
    name: string;
    keyType: "anon" | "service";
    tenantId?: string;
  },
  accessToken: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (options.tenantId) {
    headers["X-FB-Tenant"] = options.tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/admin/service-keys",
    data: {
      name: options.name,
      key_type: options.keyType,
    },
    headers,
  });
}

/**
 * Revoke a service key.
 */
export async function revokeServiceKey(
  request: APIRequestContext,
  keyId: string,
  reason: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return apiRequest(request, {
    method: "POST",
    path: `/api/v1/admin/service-keys/${keyId}/revoke`,
    data: { reason },
    headers,
  });
}

/**
 * Revoke a service key using native fetch.
 */
export async function rawRevokeServiceKey(
  keyId: string,
  reason: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: `/api/v1/admin/service-keys/${keyId}/revoke`,
    data: { reason },
    headers,
  });
}

/**
 * Create a storage bucket using native fetch.
 *
 * The backend endpoint is POST /api/v1/storage/buckets/:bucket (name in URL path),
 * with optional JSON body for {public, allowed_mime_types, max_file_size}.
 */
export async function rawCreateBucket(
  bucketId: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: `/api/v1/storage/buckets/${bucketId}`,
    headers,
  });
}

/**
 * List storage buckets.
 */
export async function listBuckets(
  request: APIRequestContext,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return apiRequest(request, {
    method: "GET",
    path: "/api/v1/storage/buckets",
    headers,
  });
}

/**
 * List storage buckets using native fetch.
 */
export async function rawListBuckets(accessToken: string, tenantId?: string) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/storage/buckets",
    headers,
  });
}

/**
 * Execute SQL query.
 */
export async function executeSQL(
  request: APIRequestContext,
  sql: string,
  accessToken: string,
) {
  return apiRequest(request, {
    method: "POST",
    path: "/api/v1/admin/sql/execute",
    data: { sql },
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * Execute SQL query using native fetch.
 */
export async function rawExecuteSQL(
  sql: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/admin/sql/execute",
    data: { sql },
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// Functions (Edge Functions)
// ────────────────────────────────────────────────────────────────

export async function rawCreateFunction(
  options: {
    name: string;
    code: string;
    verifyJWT?: boolean;
  },
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/functions",
    data: {
      name: options.name,
      code: options.code,
      verify_jwt: options.verifyJWT ?? false,
    },
    headers,
  });
}

export async function rawListFunctions(accessToken: string, tenantId?: string) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/functions",
    headers,
  });
}

export async function rawDeleteFunction(
  name: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "DELETE",
    path: `/api/v1/functions/${name}`,
    headers,
  });
}

export async function rawInvokeFunction(
  name: string,
  accessToken: string,
  tenantId?: string,
  body?: unknown,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: `/api/v1/functions/${name}/invoke`,
    data: body,
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// Jobs (Background Jobs)
// ────────────────────────────────────────────────────────────────

export async function rawSubmitJob(
  options: {
    name: string;
    function_name: string;
    payload?: unknown;
    schedule?: string;
  },
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/jobs/submit",
    data: options,
    headers,
  });
}

export async function rawListJobs(accessToken: string, tenantId?: string) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/jobs",
    headers,
  });
}

export async function rawCancelJob(
  jobId: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: `/api/v1/jobs/${jobId}/cancel`,
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// Chatbots (AI Chatbots - Admin)
// ────────────────────────────────────────────────────────────────

export async function rawListChatbots(accessToken: string, tenantId?: string) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/admin/ai/chatbots",
    headers,
  });
}

export async function rawDeleteChatbot(
  chatbotId: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "DELETE",
    path: `/api/v1/admin/ai/chatbots/${chatbotId}`,
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// Knowledge Bases
// ────────────────────────────────────────────────────────────────

export async function rawCreateKnowledgeBase(
  options: {
    name: string;
    description?: string;
  },
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/ai/knowledge-bases",
    data: options,
    headers,
  });
}

export async function rawListKnowledgeBases(
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/ai/knowledge-bases",
    headers,
  });
}

export async function rawDeleteKnowledgeBase(
  kbId: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "DELETE",
    path: `/api/v1/ai/knowledge-bases/${kbId}`,
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// Secrets
// ────────────────────────────────────────────────────────────────

export async function rawCreateSecret(
  options: {
    name: string;
    value: string;
    description?: string;
  },
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/secrets/",
    data: options,
    headers,
  });
}

export async function rawListSecrets(accessToken: string, tenantId?: string) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/secrets/",
    headers,
  });
}

export async function rawDeleteSecret(
  secretId: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "DELETE",
    path: `/api/v1/secrets/${secretId}`,
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// Impersonation
// ────────────────────────────────────────────────────────────────

export async function rawStartUserImpersonation(
  targetUserId: string,
  reason: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/auth/impersonate",
    data: { target_user_id: targetUserId, reason },
    headers,
  });
}

export async function rawStartAnonImpersonation(
  reason: string,
  accessToken: string,
) {
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/auth/impersonate/anon",
    data: { reason },
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

export async function rawStartServiceImpersonation(
  reason: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/auth/impersonate/service",
    data: { reason },
    headers,
  });
}

export async function rawStopImpersonation(accessToken: string) {
  return rawApiRequest({
    method: "DELETE",
    path: "/api/v1/auth/impersonate",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

// ---------------------------------------------------------------------------
// OAuth Providers
// ---------------------------------------------------------------------------

export async function rawCreateOAuthProvider(
  options: {
    provider_name: string;
    display_name: string;
    client_id: string;
    client_secret: string;
    redirect_url: string;
    scopes: string[];
    is_custom: boolean;
    authorization_url: string;
    token_url: string;
    user_info_url: string;
    allow_dashboard_login?: boolean;
  },
  accessToken: string,
) {
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/admin/oauth/providers",
    data: {
      provider_name: options.provider_name,
      display_name: options.display_name,
      enabled: true,
      client_id: options.client_id,
      client_secret: options.client_secret,
      redirect_url: options.redirect_url,
      scopes: options.scopes,
      is_custom: options.is_custom,
      authorization_url: options.authorization_url,
      token_url: options.token_url,
      user_info_url: options.user_info_url,
      allow_dashboard_login: options.allow_dashboard_login ?? false,
    },
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

export async function rawDeleteOAuthProvider(
  providerId: string,
  accessToken: string,
) {
  return rawApiRequest({
    method: "DELETE",
    path: `/api/v1/admin/oauth/providers/${providerId}`,
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

export async function rawListOAuthProviders(accessToken: string) {
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/admin/oauth/providers",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

// ────────────────────────────────────────────────────────────────
// Secrets Stats
// ────────────────────────────────────────────────────────────────

export async function rawGetSecretStats(
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/secrets/stats",
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// Webhooks
// ────────────────────────────────────────────────────────────────

export async function rawCreateWebhook(
  options: {
    name: string;
    url: string;
    events?: Array<{ table: string; operations: string[] }>;
  },
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  const data: Record<string, unknown> = {
    name: options.name,
    url: options.url,
    events: options.events || [
      { table: "public.test_table", operations: ["INSERT"] },
    ],
  };
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/webhooks",
    data,
    headers,
  });
}

export async function rawListWebhooks(accessToken: string, tenantId?: string) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/webhooks",
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// MCP Custom Tools
// ────────────────────────────────────────────────────────────────

export async function rawCreateMCPTool(
  options: {
    name: string;
    description: string;
    enabled?: boolean;
  },
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/mcp/tools",
    data: options,
    headers,
  });
}

export async function rawListMCPTools(accessToken: string, tenantId?: string) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/mcp/tools",
    headers,
  });
}

export async function rawDeleteMCPTool(
  toolId: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "DELETE",
    path: `/api/v1/mcp/tools/${toolId}`,
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// Webhook Deliveries
// ────────────────────────────────────────────────────────────────

export async function rawListWebhookDeliveries(
  webhookId: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: `/api/v1/webhooks/${webhookId}/deliveries`,
    headers,
  });
}

// ────────────────────────────────────────────────────────────────
// Settings (Custom)
// ────────────────────────────────────────────────────────────────

export async function rawCreateCustomSetting(
  options: {
    key: string;
    value: string;
    category?: string;
    description?: string;
  },
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "POST",
    path: "/api/v1/settings/custom",
    data: options,
    headers,
  });
}

export async function rawListCustomSettings(
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "GET",
    path: "/api/v1/settings/custom",
    headers,
  });
}

export async function rawDeleteCustomSetting(
  key: string,
  accessToken: string,
  tenantId?: string,
) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${accessToken}`,
  };
  if (tenantId) {
    headers["X-FB-Tenant"] = tenantId;
  }
  return rawApiRequest({
    method: "DELETE",
    path: `/api/v1/settings/custom/${key}`,
    headers,
  });
}
