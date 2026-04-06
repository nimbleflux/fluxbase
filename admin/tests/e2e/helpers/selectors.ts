/**
 * Common selectors for shadcn/ui components used in the Fluxbase admin UI.
 * These use data-testid attributes and accessible roles for stability.
 */

import { expect, type Page } from "@playwright/test";

// Form inputs — matches by Label text, then by input id, then by placeholder, then by name attribute.
export const formInputByLabel = (label: string) =>
  `label:has-text("${label}") >> .. >> input`;

// Matches by role="alert-dialog" aria attribute
export const alertDialog = '[role="alert-dialog"]';

// Buttons by text content
export const buttonByText = (text: string) => `button:has-text("${text}")`;

// Data table rows
export const tableRow = (row: number) => `tbody tr >> nth-child(${row})`;

// Dialog close button (the X button in top-right corner)
export const dialogCloseButton = '[class*="close"]';

// Toast notifications
export const toast = (title: string) =>
  `[data-sonner-toast][data-title="${title}"]`;

// Navigation sidebar links
export const sidebarLink = (text: string) => `nav a:has-text("${text}")`;

// Tab triggers
export const tabTrigger = (name: string) =>
  `[data-state="active"][data-value="${name}"]`;

/**
 * Wait for a shadcn/ui dialog to open by polling for visibility.
 */
export async function waitForDialog(page: Page) {
  await page.waitForSelector('[role="dialog"]', { state: "visible" });
}

/**
 * Wait for a toast notification to appear.
 */
export async function waitForToast(page: Page, title?: string) {
  if (title) {
    await page.waitForSelector(`[data-sonner-toast][data-title="${title}"]`);
  } else {
    await page.waitForSelector("[data-sonner-toast]");
  }
}

/**
 * Open the tenant selector dropdown and wait for it to be visible.
 */
export async function openTenantSelector(page: Page) {
  const selector = page.getByRole("combobox", { name: "Select tenant" });
  await selector.click();
  await expect(page.getByRole("listbox")).toBeVisible({ timeout: 5_000 });
}

/**
 * Select a specific tenant by name from the tenant selector.
 */
export async function selectTenant(page: Page, tenantName: string) {
  await openTenantSelector(page);
  const option = page.getByRole("option").filter({ hasText: tenantName });
  await option.click();
}

/**
 * Select the Nth tenant option from the tenant selector (0-indexed).
 */
export async function selectTenantByIndex(page: Page, index: number) {
  await openTenantSelector(page);
  const options = page.getByRole("option");
  const count = await options.count();
  if (index >= count) {
    throw new Error(
      `Tenant option index ${index} out of range (found ${count} options)`,
    );
  }
  await options.nth(index).click();
}

/**
 * Wait for an API call matching a URL pattern.
 */
export async function waitForApiCall(
  page: Page,
  urlPattern: string | RegExp,
  options?: { timeout?: number },
) {
  await page.waitForRequest(
    (req) =>
      typeof urlPattern === "string"
        ? req.url().includes(urlPattern)
        : urlPattern.test(req.url()),
    { timeout: options?.timeout ?? 10_000 },
  );
}
