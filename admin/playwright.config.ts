import { defineConfig } from "@playwright/test";

const isCI = !!process.env.CI;
const backendPort = process.env.PLAYWRIGHT_BACKEND_PORT || "8082";
const vitePort = "5050";

export default defineConfig({
  fullyParallel: false,
  retries: isCI ? 2 : 0,
  reporter: [["list"], ["html"]],
  timeout: 60_000,
  expect: { timeout: 10_000 },
  globalSetup: "./tests/e2e/global-setup.ts",
  globalTeardown: "./tests/e2e/global-teardown.ts",
  use: {
    baseURL:
      process.env.PLAYWRIGHT_BASE_URL ||
      (isCI
        ? `http://localhost:${backendPort}/admin/`
        : `http://localhost:${vitePort}/admin/`),
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  webServer: isCI
    ? {
        // In CI, the Go backend serves the admin UI directly (embedded build).
        // It's started as a separate CI step before Playwright runs.
        command: "true",
        port: Number(backendPort),
        reuseExistingServer: true,
        timeout: 5_000,
      }
    : {
        // Local dev: clean DB and start both servers.
        // The --clean-foreground flag resets the DB and starts in foreground
        // so Playwright can manage the process lifecycle.
        command: "../scripts/start-e2e-ui.sh --clean-foreground",
        port: Number(vitePort),
        reuseExistingServer: process.env.PLAYWRIGHT_REUSE_SERVER === "true",
        timeout: 180_000,
      },
  projects: [
    {
      name: "setup",
      testDir: "./tests/e2e",
      testMatch: /setup\.spec\.ts$/,
      type: "setup",
    },
    {
      name: "provisioning",
      testDir: "./tests/e2e",
      testMatch: /_provisioning\.spec\.ts$/,
      type: "setup",
      dependencies: ["setup"],
    },
    {
      name: "e2e",
      testDir: "./tests/e2e",
      testIgnore: [/setup\.spec\.ts$/, /_provisioning\.spec\.ts$/],
      dependencies: ["provisioning"],
    },
  ],
});
