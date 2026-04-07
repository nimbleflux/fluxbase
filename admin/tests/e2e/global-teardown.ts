/* eslint-disable no-console */
import { type FullConfig } from "@playwright/test";
import { closePool } from "./helpers/db";
import { deleteAllEmails } from "./helpers/mailhog";

async function globalTeardown(_config: FullConfig) {
  // Clean up MailHog messages
  await deleteAllEmails().catch(() => {
    // MailHog may not be available
  });

  // Close the database pool
  await closePool();

  console.log("Global teardown complete.");
}

export default globalTeardown;
