import { test, expect } from "./fixtures";
import {
  rawCreateBucket,
  rawStartServiceImpersonation,
  rawStopImpersonation,
  rawApiRequest,
} from "./helpers/api";

test.describe("Storage Page During Tenant Service Impersonation", () => {
  const cleanupTasks: Array<() => Promise<void>> = [];

  test.afterAll(async () => {
    for (const cleanup of cleanupTasks) {
      await cleanup().catch(() => {});
    }
  });

  test("storage page renders with buckets when impersonating tenant service", async ({
    adminPage,
    adminToken,
    defaultTenantId,
  }) => {
    const bucketName = `ui-ts-imp-${Date.now()}`;
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

    const impResult = await rawStartServiceImpersonation(
      "E2E storage page test",
      adminToken,
      defaultTenantId,
    );
    expect(impResult.status).toBe(200);
    const impToken = impResult.body.access_token as string;

    try {
      await adminPage.evaluate(
        ({ token, type }) => {
          localStorage.setItem("fluxbase_impersonation_token", token);
          localStorage.setItem("fluxbase_impersonation_type", type);
          localStorage.setItem(
            "fluxbase_impersonation_session",
            JSON.stringify({
              id: "test-session",
              admin_user_id: "admin",
              impersonation_type: type,
              reason: "E2E storage page test",
              started_at: new Date().toISOString(),
              is_active: true,
            }),
          );
        },
        { token: impToken, type: "service" },
      );

      await adminPage.goto("storage", { waitUntil: "networkidle" });
      await expect(adminPage).toHaveURL(/storage/);

      const consoleErrors: string[] = [];
      adminPage.on("console", (msg) => {
        if (msg.type() === "error") {
          consoleErrors.push(msg.text());
        }
      });

      const cancelBtn = adminPage.getByRole("button", {
        name: /cancel.*(service|tenant)/i,
      });
      await expect(cancelBtn).toBeVisible({ timeout: 10_000 });

      const bucketLocator = adminPage.getByText(bucketName);
      await expect(bucketLocator).toBeVisible({ timeout: 10_000 });

      expect(consoleErrors).toEqual([]);
    } finally {
      await rawStopImpersonation(adminToken).catch(() => {});
    }
  });
});
