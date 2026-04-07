import { test, expect } from "./fixtures";

test.describe("Auth Guard", () => {
  const protectedRoutes = [
    "",
    "tables",
    "schema",
    "sql-editor",
    "storage",
    "functions",
    "jobs",
    "users",
    "tenants",
    "settings",
    "extensions",
    "service-keys",
    "client-keys",
    "webhooks",
    "secrets",
    "logs",
    "monitoring",
    "rpc",
    "policies",
    "email-settings",
    "security-settings",
    "features",
    "instance-settings",
    "realtime",
    "chatbots",
    "mcp-tools",
  ];

  test("unauthenticated access to protected routes redirects to login", async ({
    page,
  }) => {
    // Navigate to login first to set a real origin and clear tokens
    await page.goto("login", { waitUntil: "networkidle" });

    // Clear any existing tokens
    await page.evaluate(() => {
      localStorage.removeItem("fluxbase_admin_access_token");
      localStorage.removeItem("fluxbase_admin_refresh_token");
      localStorage.removeItem("fluxbase_admin_user");
    });

    for (const route of protectedRoutes) {
      await page.goto(route || "./");

      // Should redirect to login page
      await expect(page).toHaveURL(/\/login/, { timeout: 5_000 });
    }
  });

  test("expired token treated as unauthenticated", async ({ page }) => {
    // Navigate to a real page first
    await page.goto("login", { waitUntil: "networkidle" });

    // Set an invalid/expired token
    await page.evaluate(() => {
      localStorage.setItem(
        "fluxbase_admin_access_token",
        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMDEwMjAxIn0ifX0.yyyyyyyy",
      );
    });

    // Navigate to a protected route
    await page.goto("./");

    // Should redirect to login (token is invalid)
    await expect(page).toHaveURL(/\/login/, { timeout: 5_000 });
  });

  test("navigation between routes preserves auth state", async ({
    adminPage,
  }) => {
    // Start on dashboard
    await expect(adminPage).toHaveURL(/\/admin\/?$/);

    // Navigate to various routes (relative to base)
    const routes = ["sql-editor", "storage", "functions", "jobs", "users"];
    for (const route of routes) {
      await adminPage.goto(route);
      // Should NOT redirect to login
      await expect(adminPage).not.toHaveURL(/\/login/);
    }

    // Token should still be present
    const token = await adminPage.evaluate(() =>
      localStorage.getItem("fluxbase_admin_access_token"),
    );
    expect(token).toBeTruthy();
  });
});
