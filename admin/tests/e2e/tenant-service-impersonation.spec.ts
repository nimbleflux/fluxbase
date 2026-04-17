import { test, expect } from "./fixtures";
import {
  rawStartServiceImpersonation,
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
} from "./helpers/api";

test.describe("Tenant Service Impersonation API Access", () => {
  const cleanupTasks: Array<() => Promise<void>> = [];

  test.afterAll(async () => {
    for (const cleanup of cleanupTasks) {
      await cleanup().catch(() => {});
    }
  });

  async function startTenantServiceImpersonation(
    adminToken: string,
    tenantId: string,
  ) {
    const impResult = await rawStartServiceImpersonation(
      "E2E tenant service impersonation test",
      adminToken,
      tenantId,
    );
    expect(impResult.status).toBe(200);
    expect(impResult.body.access_token).toBeTruthy();
    expect(impResult.body.target_user.role).toBe("tenant_service");
    return impResult.body.access_token as string;
  }

  test("tenant_service impersonation can list buckets", async ({
    adminToken,
    defaultTenantId,
  }) => {
    const bucketName = `ts-imp-bucket-${Date.now()}`;
    await rawCreateBucket(bucketName, adminToken, defaultTenantId);
    cleanupTasks.push(async () => {
      await rawApiRequest({
        method: "DELETE",
        path: `/api/v1/storage/buckets/${bucketName}`,
        headers: {
          Authorization: `Bearer ${adminToken}`,
          "X-FB-Tenant": defaultTenantId,
        },
      });
    });

    const impToken = await startTenantServiceImpersonation(
      adminToken,
      defaultTenantId,
    );

    try {
      const result = await rawListBuckets(impToken, defaultTenantId);
      expect(result.status).toBe(200);

      const rawBuckets = result.body?.buckets || result.body || [];
      const buckets = (Array.isArray(rawBuckets) ? rawBuckets : []) as Array<{
        id: string;
      }>;
      const bucketIds = buckets.map((b) => b.id);
      expect(bucketIds).toContain(bucketName);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("tenant_service impersonation can list functions", async ({
    adminToken,
    defaultTenantId,
  }) => {
    const funcName = `ts-imp-func-${Date.now()}`;
    const code = `export default function handler(req) { return new Response("ok"); }`;
    await rawCreateFunction(
      { name: funcName, code },
      adminToken,
      defaultTenantId,
    );
    cleanupTasks.push(async () => {
      await rawDeleteFunction(funcName, adminToken, defaultTenantId);
    });

    const impToken = await startTenantServiceImpersonation(
      adminToken,
      defaultTenantId,
    );

    try {
      const result = await rawListFunctions(impToken, defaultTenantId);
      expect(result.status).toBe(200);

      const rawFunctions = result.body?.functions || result.body || [];
      const functions = (Array.isArray(rawFunctions) ? rawFunctions : []) as Array<{
        name: string;
      }>;
      const funcNames = functions.map((f) => f.name);
      expect(funcNames).toContain(funcName);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("tenant_service impersonation can list secrets", async ({
    adminToken,
    defaultTenantId,
  }) => {
    const secretName = `ts-imp-secret-${Date.now()}`;
    const createResult = await rawCreateSecret(
      { name: secretName, value: "test-value" },
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

    const impToken = await startTenantServiceImpersonation(
      adminToken,
      defaultTenantId,
    );

    try {
      const result = await rawListSecrets(impToken, defaultTenantId);
      expect(result.status).toBe(200);

      const rawSecrets = result.body?.secrets || result.body || [];
      const secrets = (Array.isArray(rawSecrets) ? rawSecrets : []) as Array<{
        name: string;
      }>;
      const secretNames = secrets.map((s) => s.name);
      expect(secretNames).toContain(secretName);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("tenant_service impersonation can list admin jobs", async ({
    adminToken,
    defaultTenantId,
  }) => {
    const impToken = await startTenantServiceImpersonation(
      adminToken,
      defaultTenantId,
    );

    try {
      const result = await rawApiRequest({
        method: "GET",
        path: "/api/v1/admin/jobs",
        headers: {
          Authorization: `Bearer ${impToken}`,
          "X-FB-Tenant": defaultTenantId,
        },
      });
      expect(result.status).toBe(200);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("tenant_service impersonation can list knowledge bases", async ({
    adminToken,
    defaultTenantId,
  }) => {
    const impToken = await startTenantServiceImpersonation(
      adminToken,
      defaultTenantId,
    );

    try {
      const result = await rawApiRequest({
        method: "GET",
        path: "/api/v1/ai/knowledge-bases",
        headers: {
          Authorization: `Bearer ${impToken}`,
          "X-FB-Tenant": defaultTenantId,
        },
      });
      expect(result.status).toBe(200);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("tenant_service impersonation cannot access instance-admin endpoints", async ({
    adminToken,
    defaultTenantId,
  }) => {
    const impToken = await startTenantServiceImpersonation(
      adminToken,
      defaultTenantId,
    );

    try {
      const result = await rawApiRequest({
        method: "GET",
        path: "/api/v1/admin/tenants",
        headers: {
          Authorization: `Bearer ${impToken}`,
          "X-FB-Tenant": defaultTenantId,
        },
      });
      expect([401, 403]).toContain(result.status);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });

  test("tenant_service impersonation is isolated to own tenant", async ({
    adminToken,
    defaultTenantId,
    tenantAdminInfo,
  }) => {
    const bucketInDefault = `ts-iso-default-${Date.now()}`;
    const bucketInSecond = `ts-iso-second-${Date.now()}`;
    await rawCreateBucket(bucketInDefault, adminToken, defaultTenantId);
    await rawCreateBucket(
      bucketInSecond,
      adminToken,
      tenantAdminInfo.tenantId,
    );
    cleanupTasks.push(
      async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/storage/buckets/${bucketInDefault}`,
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "X-FB-Tenant": defaultTenantId,
          },
        });
      },
      async () => {
        await rawApiRequest({
          method: "DELETE",
          path: `/api/v1/storage/buckets/${bucketInSecond}`,
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "X-FB-Tenant": tenantAdminInfo.tenantId,
          },
        });
      },
    );

    const impToken = await startTenantServiceImpersonation(
      adminToken,
      defaultTenantId,
    );

    try {
      const result = await rawListBuckets(impToken, defaultTenantId);
      expect(result.status).toBe(200);

      const rawBuckets = result.body?.buckets || result.body || [];
      const buckets = (Array.isArray(rawBuckets) ? rawBuckets : []) as Array<{
        id: string;
      }>;
      const bucketIds = buckets.map((b) => b.id);

      expect(bucketIds).toContain(bucketInDefault);
      expect(bucketIds).not.toContain(bucketInSecond);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });
});
