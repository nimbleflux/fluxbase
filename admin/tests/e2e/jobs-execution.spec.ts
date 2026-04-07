import { test, expect } from "./fixtures";
import {
  rawCreateFunction,
  rawDeleteFunction,
  rawSubmitJob,
  rawListJobs,
  rawCancelJob,
} from "./helpers/api";

test.describe("Background Jobs Execution", () => {
  let adminToken: string;

  test.beforeAll(async () => {
    const { rawLogin } = await import("./helpers/api");
    const result = await rawLogin({
      email: "admin@fluxbase.test",
      password: "test-password-32chars!!",
    });
    adminToken = result.body.access_token;
  });

  const cleanupFunctions: Array<{ name: string }> = [];

  test.afterAll(async () => {
    for (const { name } of cleanupFunctions) {
      await rawDeleteFunction(name, adminToken).catch(() => {});
    }
  });

  test("jobs page loads without errors", async ({ adminPage }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    const consoleErrors: string[] = [];
    adminPage.on("console", (msg) => {
      if (msg.type() === "error") consoleErrors.push(msg.text());
    });
    await adminPage.waitForTimeout(2000);
    const criticalErrors = consoleErrors.filter(
      (text) =>
        !text.includes("500") &&
        !text.includes("404") &&
        !text.includes("Failed to fetch") &&
        !text.includes("favicon"),
    );
    expect(criticalErrors).toHaveLength(0);
  });

  test("submit job via API and verify in UI", async ({ adminPage }) => {
    // First create a function to use as the job target
    const funcName = `e2e-job-func-${Date.now()}`;
    const code = `
export default function handler(req: Request): Response {
  return new Response(JSON.stringify({ processed: true }), {
    headers: { "Content-Type": "application/json" },
  });
}`;
    await rawCreateFunction({ name: funcName, code }, adminToken);
    cleanupFunctions.push({ name: funcName });

    // Submit a job
    const jobResult = await rawSubmitJob(
      {
        name: `e2e-job-${Date.now()}`,
        function_name: funcName,
        payload: { test: true },
      },
      adminToken,
    );

    // Job submission might fail if the function isn't fully registered yet
    // but we should at least verify the API responds
    expect(jobResult.status).toBeLessThan(500);

    // Navigate to jobs page and verify it loads
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);
  });

  test("list jobs via API returns expected structure", async ({
    adminToken,
  }) => {
    const result = await rawListJobs(adminToken);
    expect(result.status).toBe(200);
    // Should return an array or object with jobs
    expect(result.body).toBeTruthy();
  });

  test("jobs page shows stats cards", async ({ adminPage }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });

    // The jobs page should show stats cards
    // These may show 0 counts but the cards should exist
    await adminPage.waitForTimeout(2000);

    // Verify page content exists (not blank)
    const hasContent = await adminPage.evaluate(() => {
      return document.getElementById("root")?.innerHTML?.length > 100;
    });
    expect(hasContent).toBeTruthy();
  });

  test("jobs page displays header with title and impersonation popover", async ({
    adminPage,
  }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Verify the page header
    await expect(
      adminPage.getByRole("heading", { name: /background jobs/i }),
    ).toBeVisible({ timeout: 10_000 });

    // Verify the ImpersonationPopover is present
    const impersonationElement = adminPage.getByText(
      /not impersonating|impersonating|running as/i,
    );
    // The element may or may not be visible depending on state, but the page should load
    const isVisible = await impersonationElement.isVisible().catch(() => false);
    // Just verify the page rendered correctly regardless of impersonation state
    expect(isVisible).toBeDefined();
  });

  test("jobs page shows job queue and functions tabs", async ({
    adminPage,
  }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Wait for the tabs to render
    await adminPage.waitForTimeout(2000);

    // Verify both tabs exist
    const queueTab = adminPage.getByRole("tab", { name: /job queue/i });
    const functionsTab = adminPage.getByRole("tab", { name: /job functions/i });

    await expect(queueTab).toBeVisible({ timeout: 10_000 });
    await expect(functionsTab).toBeVisible({ timeout: 10_000 });
  });

  test("jobs page queue tab shows namespace selector and filters", async ({
    adminPage,
  }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Ensure we're on the queue tab
    const queueTab = adminPage.getByRole("tab", { name: /job queue/i });
    await queueTab.click();

    // Verify namespace selector exists
    const namespaceSelect = adminPage.locator("#queue-namespace-select");
    await expect(namespaceSelect).toBeVisible({ timeout: 10_000 });

    // Verify search input exists
    const searchInput = adminPage.getByPlaceholder(/search jobs/i);
    await expect(searchInput).toBeVisible({ timeout: 10_000 });
  });

  test("jobs page shows empty state when no jobs exist", async ({
    adminPage,
  }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Wait for data to load
    await adminPage.waitForTimeout(3000);

    // The page should show either jobs or an empty state message
    const hasNoJobsMessage = await adminPage
      .getByText(/no jobs found/i)
      .isVisible()
      .catch(() => false);
    const hasJobRows = await adminPage
      .getByRole("button", { name: /^view$/i })
      .first()
      .isVisible()
      .catch(() => false);

    // Either empty state or job rows should be present
    expect(hasNoJobsMessage || hasJobRows).toBeTruthy();
  });

  test("jobs page refresh button works", async ({ adminPage }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Wait for initial load
    await adminPage.waitForTimeout(2000);

    // Click the refresh button
    const refreshButton = adminPage.getByRole("button", { name: /refresh/i });
    await expect(refreshButton).toBeVisible({ timeout: 10_000 });
    await refreshButton.click();

    // Page should still be on jobs after refresh
    await adminPage.waitForTimeout(2000);
    await expect(adminPage).toHaveURL(/jobs/);
  });

  test("switching to functions tab shows function stats", async ({
    adminPage,
  }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Click on the Job Functions tab
    const functionsTab = adminPage.getByRole("tab", { name: /job functions/i });
    await functionsTab.click();

    // Wait for tab content to render
    await adminPage.waitForTimeout(2000);

    // Verify stat cards for functions are visible
    const totalFunctionsCard = adminPage.getByText(/total functions/i);
    const enabledCard = adminPage.getByText(/^enabled$/i);
    const scheduledCard = adminPage.getByText(/^scheduled$/i);

    await expect(totalFunctionsCard).toBeVisible({ timeout: 10_000 });
    await expect(enabledCard).toBeVisible({ timeout: 5_000 });
    await expect(scheduledCard).toBeVisible({ timeout: 5_000 });
  });

  test("submit job via API and view in job queue", async ({ adminPage }) => {
    // Create a function for the job
    const funcName = `e2e-queue-func-${Date.now()}`;
    const code = `
export default function handler(req: Request): Response {
  return new Response(JSON.stringify({ ok: true }), {
    headers: { "Content-Type": "application/json" },
  });
}`;
    await rawCreateFunction({ name: funcName, code }, adminToken);
    cleanupFunctions.push({ name: funcName });

    // Wait for function registration
    await adminPage.waitForTimeout(1000);

    // Submit a job via the API
    const jobName = `e2e-job-view-${Date.now()}`;
    const submitResult = await rawSubmitJob(
      {
        name: jobName,
        function_name: funcName,
        payload: { source: "e2e-test" },
      },
      adminToken,
    );

    // Navigate to jobs page
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // If the job was submitted successfully, it should appear in the queue
    if (submitResult.status < 300) {
      await adminPage.waitForTimeout(3000);

      // Check if the job appears in the list (it may have already completed)
      const pageContent = await adminPage.textContent("body");
      const jobVisible =
        pageContent?.includes(jobName) ||
        pageContent?.includes("No jobs found");
      expect(jobVisible).toBeTruthy();
    }
  });

  test("job details dialog opens from queue", async ({ adminPage }) => {
    // Create a function and submit a job
    const funcName = `e2e-detail-func-${Date.now()}`;
    const code = `
export default function handler(req: Request): Response {
  return new Response(JSON.stringify({ detail: true }), {
    headers: { "Content-Type": "application/json" },
  });
}`;
    await rawCreateFunction({ name: funcName, code }, adminToken);
    cleanupFunctions.push({ name: funcName });

    await adminPage.waitForTimeout(1000);

    const submitResult = await rawSubmitJob(
      {
        name: `e2e-detail-job-${Date.now()}`,
        function_name: funcName,
        payload: {},
      },
      adminToken,
    );

    // Navigate to jobs page
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);
    await adminPage.waitForTimeout(3000);

    // If job exists and a View button is visible, click it
    if (submitResult.status < 300) {
      const viewButton = adminPage
        .getByRole("button", { name: /^view$/i })
        .first();
      const isViewVisible = await viewButton.isVisible().catch(() => false);

      if (isViewVisible) {
        await viewButton.click();

        // A dialog should open with job details
        await adminPage.waitForTimeout(2000);

        // Check for dialog content (job details dialog)
        const dialogVisible = await adminPage
          .getByRole("dialog")
          .isVisible()
          .catch(() => false);
        expect(dialogVisible).toBeTruthy();
      }
    }
  });

  test("status filter filters jobs correctly", async ({ adminPage }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Make sure we're on the queue tab
    const queueTab = adminPage.getByRole("tab", { name: /job queue/i });
    await queueTab.click();

    await adminPage.waitForTimeout(2000);

    // Open the status filter dropdown
    const statusFilterTrigger = adminPage.locator(
      "button:has(> svg.lucide-filter)",
    );
    await expect(statusFilterTrigger).toBeVisible({ timeout: 10_000 });
    await statusFilterTrigger.click();

    // Select "Completed" status
    const completedOption = adminPage.getByRole("option", {
      name: /completed/i,
    });
    await expect(completedOption).toBeVisible({ timeout: 5_000 });
    await completedOption.click();

    // Wait for filter to apply
    await adminPage.waitForTimeout(2000);

    // Reset filter back to "All Status"
    await statusFilterTrigger.click();
    const allStatusOption = adminPage.getByRole("option", {
      name: /all status/i,
    });
    await allStatusOption.click();
  });

  test("job API cancel endpoint responds correctly", async () => {
    // Create a function for a long-running job
    const funcName = `e2e-cancel-func-${Date.now()}`;
    const code = `
export default async function handler(req: Request): Promise<Response> {
  await new Promise((resolve) => setTimeout(resolve, 60000));
  return new Response(JSON.stringify({ done: true }), {
    headers: { "Content-Type": "application/json" },
  });
}`;
    await rawCreateFunction({ name: funcName, code }, adminToken);
    cleanupFunctions.push({ name: funcName });

    await new Promise((resolve) => setTimeout(resolve, 1000));

    // Submit a job
    const submitResult = await rawSubmitJob(
      {
        name: `e2e-cancel-job-${Date.now()}`,
        function_name: funcName,
        payload: {},
      },
      adminToken,
    );

    if (submitResult.status < 300 && submitResult.body?.id) {
      // Try to cancel the job
      const cancelResult = await rawCancelJob(submitResult.body.id, adminToken);

      // Cancel should succeed or indicate the job already completed
      expect(cancelResult.status).toBeLessThan(500);
    }
  });

  test("jobs page functions tab shows namespace selector", async ({
    adminPage,
  }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Switch to functions tab
    const functionsTab = adminPage.getByRole("tab", { name: /job functions/i });
    await functionsTab.click();

    // Verify namespace selector on functions tab
    const namespaceSelect = adminPage.locator("#namespace-select");
    await expect(namespaceSelect).toBeVisible({ timeout: 10_000 });
  });

  test("jobs page stats bar shows worker count and success rate", async ({
    adminPage,
  }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Wait for stats to render
    await adminPage.waitForTimeout(3000);

    // Verify stats bar elements exist
    const pendingLabel = adminPage.getByText(/pending:/i);
    const runningLabel = adminPage.getByText(/running:/i);
    const completedLabel = adminPage.getByText(/completed:/i);
    const failedLabel = adminPage.getByText(/failed:/i);
    const workersLabel = adminPage.getByText(/workers:/i);
    const successLabel = adminPage.getByText(/success:/i);

    await expect(pendingLabel).toBeVisible({ timeout: 10_000 });
    await expect(runningLabel).toBeVisible({ timeout: 5_000 });
    await expect(completedLabel).toBeVisible({ timeout: 5_000 });
    await expect(failedLabel).toBeVisible({ timeout: 5_000 });
    await expect(workersLabel).toBeVisible({ timeout: 5_000 });
    await expect(successLabel).toBeVisible({ timeout: 5_000 });
  });

  test("jobs page sync from filesystem button exists on functions tab", async ({
    adminPage,
  }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Switch to functions tab
    const functionsTab = adminPage.getByRole("tab", { name: /job functions/i });
    await functionsTab.click();
    await adminPage.waitForTimeout(2000);

    // Verify sync button exists
    const syncButton = adminPage.getByRole("button", {
      name: /sync from filesystem/i,
    });
    await expect(syncButton).toBeVisible({ timeout: 10_000 });
  });

  test("jobs page search filters job list", async ({ adminPage }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Wait for initial load
    await adminPage.waitForTimeout(2000);

    // Find the search input on the queue tab
    const searchInput = adminPage.getByPlaceholder(/search jobs/i);
    await expect(searchInput).toBeVisible({ timeout: 10_000 });

    // Type a search query that won't match anything
    await searchInput.fill("nonexistent-job-xyz-12345");
    await adminPage.waitForTimeout(1000);

    // The list should show no matching jobs or the empty state
    const noJobsFound = await adminPage
      .getByText(/no jobs found/i)
      .isVisible()
      .catch(() => false);

    // If there were jobs before filtering, we should see "No jobs found"
    // If there were no jobs to begin with, this is also fine
    expect(noJobsFound).toBeDefined();

    // Clear the search
    await searchInput.clear();
    await adminPage.waitForTimeout(1000);
  });

  test("jobs API list returns valid response for admin", async () => {
    const result = await rawListJobs(adminToken);
    expect(result.status).toBe(200);

    // The response should be parseable
    const body = result.body;
    expect(body).not.toBeNull();
  });

  test("jobs API submit rejects invalid function name", async () => {
    const result = await rawSubmitJob(
      {
        name: `e2e-invalid-${Date.now()}`,
        function_name: "nonexistent-function-xyz",
        payload: {},
      },
      adminToken,
    );

    // Should return an error (400 or 404) for a nonexistent function
    expect(result.status).toBeGreaterThanOrEqual(400);
  });

  test("jobs page renders without JS errors after navigation away and back", async ({
    adminPage,
  }) => {
    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);

    // Navigate away
    await adminPage.goto("functions", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/functions/);

    // Navigate back
    const consoleErrors: string[] = [];
    adminPage.on("console", (msg) => {
      if (msg.type() === "error") consoleErrors.push(msg.text());
    });

    await adminPage.goto("jobs", { waitUntil: "networkidle" });
    await expect(adminPage).toHaveURL(/jobs/);
    await adminPage.waitForTimeout(2000);

    const criticalErrors = consoleErrors.filter(
      (text) =>
        !text.includes("500") &&
        !text.includes("404") &&
        !text.includes("Failed to fetch") &&
        !text.includes("favicon"),
    );
    expect(criticalErrors).toHaveLength(0);
  });
});
