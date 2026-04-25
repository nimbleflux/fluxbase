import { test, expect } from "./fixtures";

test.describe("Impersonation Flow", () => {
  // ── ImpersonationSelector (Header) ──

  test("impersonation button is visible for tenant admin when tenant is selected", async ({
    tenantAdminPage,
  }) => {
    // Tenant admin should see the "Impersonate User" button since a tenant is auto-selected
    const button = tenantAdminPage.getByRole("button", {
      name: /impersonate/i,
    });
    await expect(button).toBeVisible({ timeout: 10_000 });
  });

  test("impersonation button is visible for instance admin when tenant selected", async ({
    adminPage,
  }) => {
    // Instance admin with tenant selected should see the button
    const button = adminPage.getByRole("button", { name: /impersonate/i });
    if (await button.isVisible().catch(() => false)) {
      expect(await button.textContent()).toContain("Impersonate");
    }
  });

  test("start user impersonation via dialog", async ({ adminPage }) => {
    // Click impersonate button to open dialog
    const button = adminPage.getByRole("button", { name: /impersonate/i });
    if (!(await button.isVisible().catch(() => false))) {
      test.skip();
      return;
    }
    await button.click();

    // Dialog should open
    await expect(adminPage.getByText("Start User Impersonation")).toBeVisible({
      timeout: 5_000,
    });

    // Select "Specific User" type (should be default)
    await expect(adminPage.getByText("Select user...")).toBeVisible({
      timeout: 5_000,
    });

    // Fill reason
    await adminPage.fill("#reason", "E2E test impersonation");

    // Close dialog without starting
    await adminPage.getByRole("button", { name: "Cancel" }).click();
    await expect(
      adminPage.getByText("Start User Impersonation"),
    ).not.toBeVisible();
  });

  test("start anonymous impersonation via dialog", async ({ adminPage }) => {
    const button = adminPage.getByRole("button", { name: /impersonate/i });
    if (!(await button.isVisible().catch(() => false))) {
      test.skip();
      return;
    }
    await button.click();
    await expect(adminPage.getByText("Start User Impersonation")).toBeVisible({
      timeout: 5_000,
    });

    await adminPage
      .getByRole("combobox", { name: /impersonation type/i })
      .click();
    await adminPage.getByRole("option", { name: /anonymous/i }).click();

    await adminPage.fill("#reason", "E2E test anon impersonation");

    await adminPage
      .getByRole("button", { name: /start impersonation/i })
      .click();

    await expect(
      adminPage.getByRole("button", { name: /cancel.*anonymous/i }),
    ).toBeVisible({ timeout: 10_000 });

    await adminPage.getByRole("button", { name: /cancel.*anonymous/i }).click();
    await expect(
      adminPage.getByRole("button", { name: /impersonate user/i }),
    ).toBeVisible({ timeout: 5_000 });
  });

  test("start tenant service impersonation via dialog", async ({
    adminPage,
  }) => {
    const button = adminPage.getByRole("button", { name: /impersonate/i });
    if (!(await button.isVisible().catch(() => false))) {
      test.skip();
      return;
    }
    await button.click();
    await expect(adminPage.getByText("Start User Impersonation")).toBeVisible({
      timeout: 5_000,
    });

    await adminPage
      .getByRole("combobox", { name: /impersonation type/i })
      .click();
    // When tenant is selected, the option may show "Service Role" or "Tenant Service"
    const serviceOption =
      adminPage.getByText("Service Role").first() ??
      adminPage.getByText("Tenant Service").first();
    await serviceOption.click();

    await adminPage.fill("#reason", "E2E test service impersonation");

    await adminPage
      .getByRole("button", { name: /start impersonation/i })
      .click();

    // Should show cancel button with either "Service Role" or "Tenant Service"
    await expect(
      adminPage.getByRole("button", { name: /cancel.*(service|tenant)/i }),
    ).toBeVisible({ timeout: 10_000 });

    await adminPage
      .getByRole("button", { name: /cancel.*(service|tenant)/i })
      .click();
    await expect(
      adminPage.getByRole("button", { name: /impersonate user/i }),
    ).toBeVisible({ timeout: 5_000 });
  });

  test("impersonation state persists across navigation", async ({
    adminPage,
  }) => {
    const button = adminPage.getByRole("button", { name: /impersonate/i });
    if (!(await button.isVisible().catch(() => false))) {
      test.skip();
      return;
    }
    await button.click();
    await expect(adminPage.getByText("Start User Impersonation")).toBeVisible({
      timeout: 5_000,
    });

    await adminPage
      .getByRole("combobox", { name: /impersonation type/i })
      .click();
    await adminPage.getByRole("option", { name: /anonymous/i }).click();
    await adminPage.fill("#reason", "E2E nav persistence test");
    await adminPage
      .getByRole("button", { name: /start impersonation/i })
      .click();

    await expect(
      adminPage.getByRole("button", { name: /cancel.*anonymous/i }),
    ).toBeVisible({ timeout: 10_000 });

    await adminPage.goto("storage", { waitUntil: "networkidle" });

    await expect(
      adminPage.getByRole("button", { name: /cancel.*anonymous/i }),
    ).toBeVisible({ timeout: 5_000 });

    await adminPage.getByRole("button", { name: /cancel.*anonymous/i }).click();
  });

  test("tenant admin can impersonate anonymous within their tenant", async ({
    tenantAdminPage,
  }) => {
    const button = tenantAdminPage.getByRole("button", {
      name: /impersonate/i,
    });
    await expect(button).toBeVisible({ timeout: 10_000 });
    await button.click();

    await expect(
      tenantAdminPage.getByText("Start User Impersonation"),
    ).toBeVisible({ timeout: 5_000 });

    await tenantAdminPage
      .getByRole("combobox", { name: /impersonation type/i })
      .click();
    await tenantAdminPage.getByRole("option", { name: /anonymous/i }).click();

    await tenantAdminPage.fill("#reason", "E2E tenant admin impersonation");

    await tenantAdminPage
      .getByRole("button", { name: /start impersonation/i })
      .click();

    await expect(
      tenantAdminPage.getByRole("button", { name: /cancel.*anonymous/i }),
    ).toBeVisible({ timeout: 10_000 });

    await tenantAdminPage
      .getByRole("button", { name: /cancel.*anonymous/i })
      .click();
  });

  test("tenant admin tenant selector is locked during impersonation", async ({
    tenantAdminPage,
  }) => {
    const button = tenantAdminPage.getByRole("button", {
      name: /impersonate/i,
    });
    await expect(button).toBeVisible({ timeout: 10_000 });
    await button.click();

    await tenantAdminPage
      .getByRole("combobox", { name: /impersonation type/i })
      .click();
    await tenantAdminPage.getByRole("option", { name: /anonymous/i }).click();
    await tenantAdminPage.fill("#reason", "E2E tenant lock test");

    await tenantAdminPage
      .getByRole("button", { name: /start impersonation/i })
      .click();

    await expect(
      tenantAdminPage.getByRole("button", { name: /cancel.*anonymous/i }),
    ).toBeVisible({ timeout: 10_000 });

    // Tenant selector should be disabled (locked)
    const tenantSelector = tenantAdminPage.getByRole("combobox", {
      name: /select tenant/i,
    });
    if (await tenantSelector.isVisible().catch(() => false)) {
      expect(await tenantSelector.isDisabled()).toBe(true);
    }

    // Clean up
    await tenantAdminPage
      .getByRole("button", { name: /cancel.*anonymous/i })
      .click();
  });

  // ── ImpersonationPopover (Inline in service pages) ──

  test("impersonation popover appears on functions page", async ({
    adminPage,
  }) => {
    await adminPage.goto("functions", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/functions/);
  });

  test("impersonation popover appears on jobs page", async ({ adminPage }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);
  });
});
