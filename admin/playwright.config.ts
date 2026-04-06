import { defineConfig } from "@playwright/test";

export default defineConfig({
  fullyParallel: false,
  retries: process.env.CI ? 2 : 0,
  reporter: [["list"], ["html"]],
  timeout: 30_000,
  expect: { timeout: 10_000 },
  globalSetup: "./tests/e2e/global-setup.ts",
  globalTeardown: "./tests/e2e/global-teardown.ts",
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || "http://localhost:5050/admin/",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  webServer: {
    command: "../scripts/start-e2e-ui.sh --clean",
    port: 5050,
    reuseExistingServer: process.env.PLAYWRIGHT_REUSE_SERVER === "true",
    timeout: 120_000,
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
    },
    {
      name: "e2e",
      testDir: "./tests/e2e",
      testIgnore: [/setup\.spec\.ts$/, /_provisioning\.spec\.ts$/],
      dependencies: ["provisioning"],
    },
  ],
});
