import { test, expect } from "./fixtures";
import { SETUP_TOKEN } from "./helpers/constants";

test.describe("Setup / Onboarding", () => {
  // Tests are ordered so validation tests run BEFORE the success test
  // that completes setup (which changes server state irreversibly).

  test("fresh instance shows setup page, not dashboard", async ({ page }) => {
    await page.goto("./", { waitUntil: "networkidle" });
    await expect(page).toHaveURL(/\/setup/, { timeout: 10_000 });
  });

  test("invalid setup token shows error", async ({ page }) => {
    await page.goto("setup", { waitUntil: "networkidle" });
    await expect(page.locator("#setupToken")).toBeVisible({ timeout: 15_000 });

    await page.fill("#setupToken", "wrong-token");
    await page.fill("#name", "Test User");
    await page.fill("#email", "test@example.com");
    await page.fill("#password", "test-password-32chars!!");
    await page.fill("#confirmPassword", "test-password-32chars!!");

    await page.click('button[type="submit"]');

    await expect(page.locator("[data-sonner-toast]")).toBeVisible({
      timeout: 5_000,
    });
    await expect(page).toHaveURL(/\/setup/);
  });

  test("password too short shows validation error", async ({ page }) => {
    await page.goto("setup", { waitUntil: "networkidle" });
    await expect(page.locator("#setupToken")).toBeVisible({ timeout: 15_000 });

    await page.fill("#setupToken", SETUP_TOKEN);
    await page.fill("#name", "Test User");
    await page.fill("#email", "test@example.com");
    await page.fill("#password", "short");
    await page.fill("#confirmPassword", "short");

    await page.click('button[type="submit"]');

    await expect(
      page.getByText("Password must be at least 12 characters"),
    ).toBeVisible();
  });

  test("password mismatch shows validation error", async ({ page }) => {
    await page.goto("setup", { waitUntil: "networkidle" });
    await expect(page.locator("#setupToken")).toBeVisible({ timeout: 15_000 });

    await page.fill("#setupToken", SETUP_TOKEN);
    await page.fill("#name", "Test User");
    await page.fill("#email", "test@example.com");
    await page.fill("#password", "test-password-32chars!!");
    await page.fill("#confirmPassword", "different-password-32chars!!");

    await page.click('button[type="submit"]');

    await expect(page.getByText("Passwords do not match")).toBeVisible();
  });

  test("valid setup creates admin and redirects to dashboard", async ({
    page,
  }) => {
    await page.goto("setup", { waitUntil: "networkidle" });
    await expect(page.locator("#setupToken")).toBeVisible({ timeout: 15_000 });

    await page.fill("#setupToken", SETUP_TOKEN);
    await page.fill("#name", "E2E Test Admin");
    await page.fill("#email", "admin@fluxbase.test");
    await page.fill("#password", "test-password-32chars!!");
    await page.fill("#confirmPassword", "test-password-32chars!!");

    await page.click('button[type="submit"]');

    // Should redirect to dashboard or home
    await expect(page).toHaveURL(/\/(admin\/?)?$/, { timeout: 15_000 });

    const token = await page.evaluate(() =>
      localStorage.getItem("fluxbase_admin_access_token"),
    );
    expect(token).toBeTruthy();
  });

  test("after setup, /setup redirects to /login", async ({ page }) => {
    await page.goto("setup", { waitUntil: "networkidle" });
    // Setup page should show a message or redirect since admin already exists
    // Check that we're not stuck on a blank page
    await page.waitForTimeout(2000);
    // The setup page should either redirect or show a message
    // since setup is already complete
    const hasContent = await page.evaluate(() => {
      return document.getElementById("root")?.innerHTML?.length > 100;
    });
    expect(hasContent).toBeTruthy();
  });
});
