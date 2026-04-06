import { test, expect } from "./fixtures";

test.describe("Middleware Verification", () => {
  test("authenticated API calls include Authorization header", async ({
    page,
  }) => {
    // Set up request interception BEFORE navigating
    const apiRequests: { url: string; headers: Record<string, string> }[] = [];
    page.context().on("request", (req) => {
      if (
        (req.url().includes("/api/v1/") || req.url().includes("/dashboard/")) &&
        req.method() !== "OPTIONS"
      ) {
        apiRequests.push({
          url: req.url(),
          headers: req.headers(),
        });
      }
    });

    // Login via the login page
    await page.goto("login", { waitUntil: "networkidle" });
    await expect(page.locator("#email")).toBeVisible({ timeout: 15_000 });

    await page.fill("#email", "admin@fluxbase.test");
    await page.fill("#password", "test-password-32chars!!");
    await page.click('button[type="submit"]');

    // Wait for dashboard
    await expect(page).toHaveURL(/\/admin\/?$/, { timeout: 10_000 });
    await page.waitForTimeout(2000);

    // The login API call and dashboard API calls should have been captured
    const authCalls = apiRequests.filter(
      (r) =>
        r.url.includes("/dashboard/auth/login") || r.url.includes("/api/v1/"),
    );

    // At least the login call should have been made
    expect(authCalls.length).toBeGreaterThan(0);

    // Check that at least one request has the Authorization header
    // (dashboard calls after login should have it)
    const callsWithAuth = authCalls.filter(
      (r) =>
        r.headers["authorization"] &&
        r.headers["authorization"].includes("Bearer"),
    );
    expect(callsWithAuth.length).toBeGreaterThan(0);
  });

  test("dashboard loads without critical JS errors", async ({ adminPage }) => {
    const errors: string[] = [];
    adminPage.on("console", (msg) => {
      if (msg.type() === "error") {
        errors.push(msg.text());
      }
    });

    await adminPage.goto("./", { waitUntil: "networkidle" });
    await adminPage.waitForTimeout(3000);

    // Filter out errors from empty database (500s from API calls)
    const criticalErrors = errors.filter(
      (text) =>
        !text.includes("500") &&
        !text.includes("404") &&
        !text.includes("401") &&
        !text.includes("403") &&
        !text.includes("Failed to fetch") &&
        !text.includes("NetworkError") &&
        !text.includes("net::ERR") &&
        !text.includes("favicon") &&
        !text.includes("Failed to load resource"),
    );
    expect(criticalErrors).toHaveLength(0);
  });
});
