/* eslint-disable no-console */
import { type FullConfig } from "@playwright/test";

const API_BASE = process.env.PLAYWRIGHT_API_URL || "http://localhost:5050";

async function globalSetup(_config: FullConfig) {
  const healthURL = `${API_BASE}/health`.replace(
    ":5050/health",
    ":8082/health",
  );

  console.log(`Waiting for server at ${healthURL}...`);
  await waitForServer(healthURL, 60_000);
  console.log("Server is ready.");
}

async function waitForServer(url: string, timeout: number): Promise<void> {
  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    try {
      const response = await fetch(url);
      if (response.ok) return;
    } catch {
      // Server not ready yet
    }
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }
  throw new Error(`Server at ${url} not ready within ${timeout / 1000}s`);
}

export default globalSetup;
