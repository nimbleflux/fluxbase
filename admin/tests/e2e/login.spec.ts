import { test, expect } from "./fixtures";
import { ADMIN_EMAIL, ADMIN_PASSWORD } from "./helpers/constants";

test.describe("Login / Logout", () => {
  test("login with valid credentials redirects to dashboard", async ({
    page,
  }) => {
    await page.goto("login", { waitUntil: "networkidle" });
    await expect(page.locator("#email")).toBeVisible({ timeout: 15_000 });

    await page.fill("#email", ADMIN_EMAIL);
    await page.fill("#password", ADMIN_PASSWORD);
    await page.click('button[type="submit"]');

    // Should redirect to dashboard
    await expect(page).toHaveURL(/\/admin\/$/, { timeout: 10_000 });

    // Tokens should be stored (in cookies via Zustand auth store)
    const token = await page.evaluate(() => {
      const prefix = "fluxbase_admin_token=";
      const parts = document.cookie.split("; ");
      for (const part of parts) {
        if (part.startsWith(prefix)) {
          try {
            return JSON.parse(part.substring(prefix.length));
          } catch {
            return part.substring(prefix.length);
          }
        }
      }
      return null;
    });
    expect(token).toBeTruthy();
  });

  test("login with invalid credentials shows error", async ({ page }) => {
    await page.goto("login", { waitUntil: "networkidle" });
    await expect(page.locator("#email")).toBeVisible({ timeout: 15_000 });

    await page.fill("#email", ADMIN_EMAIL);
    await page.fill("#password", "wrong-password-32chars!!");
    await page.click('button[type="submit"]');

    // Should show error toast (sonner)
    await expect(page.locator("[data-sonner-toast]")).toBeVisible({
      timeout: 5_000,
    });

    // Should stay on login page
    await expect(page).toHaveURL(/\/login/);
  });

  test("login page has email and password form", async ({ page }) => {
    await page.goto("login", { waitUntil: "networkidle" });
    await expect(page.locator("#email")).toBeVisible({ timeout: 15_000 });

    // The login page should have email/password form
    await expect(page.locator("#email")).toBeVisible();
    await expect(page.locator("#password")).toBeVisible();

    // No JS errors on the login page
    const consoleErrors: string[] = [];
    page.on("console", (msg) => {
      if (msg.type() === "error") {
        consoleErrors.push(msg.text());
      }
    });
    await page.waitForTimeout(1000);
    expect(consoleErrors).toHaveLength(0);
  });

  test("login page is accessible when not authenticated", async ({ page }) => {
    await page.goto("login", { waitUntil: "networkidle" });

    // Should be on login page
    await expect(page).toHaveURL(/\/login/);

    // Form elements should be visible and enabled
    await expect(page.locator("#email")).toBeVisible({ timeout: 15_000 });
    await expect(page.locator("#email")).toBeEnabled();
    await expect(page.locator("#password")).toBeVisible();
    await expect(page.locator("#password")).toBeEnabled();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeEnabled();
  });
});
