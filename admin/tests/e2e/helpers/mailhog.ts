/**
 * MailHog integration helpers for E2E tests.
 * MailHog runs at http://localhost:8025 in the devcontainer.
 */

const MAILHOG_BASE_URL = process.env.MAILHOG_URL || "http://localhost:8025";

interface MailHogEmail {
  Content: {
    Body: string;
    Headers: Record<string, string[]>;
  };
  Raw: {
    To: string[];
    From: string;
  };
}

interface MailHogResponse {
  items: MailHogEmail[];
  count: number;
  total: number;
}

/**
 * Delete all emails in MailHog.
 * Useful for test isolation.
 */
export async function deleteAllEmails(): Promise<void> {
  const response = await fetch(`${MAILHOG_BASE_URL}/api/v1/messages`, {
    method: "DELETE",
  });
  if (!response.ok) {
    throw new Error(`Failed to delete MailHog messages: ${response.status}`);
  }
}

/**
 * Get all emails from MailHog.
 */
export async function getAllEmails(): Promise<MailHogResponse> {
  const response = await fetch(`${MAILHOG_BASE_URL}/api/v2/messages`);
  if (!response.ok) {
    throw new Error(`Failed to get MailHog messages: ${response.status}`);
  }
  return response.json();
}

/**
 * Find an email sent to a specific address.
 */
export async function findEmailTo(
  toAddress: string,
): Promise<MailHogEmail | null> {
  const data = await getAllEmails();
  if (!data.items || data.count === 0) {
    return null;
  }
  return (
    data.items.find((item) =>
      item.Raw.To.some((addr) => addr.includes(toAddress)),
    ) || null
  );
}

/**
 * Find an email with a specific subject.
 */
export async function findEmailWithSubject(
  subject: string,
): Promise<MailHogEmail | null> {
  const data = await getAllEmails();
  if (!data.items || data.count === 0) {
    return null;
  }
  return (
    data.items.find((item) =>
      item.Content.Headers.Subject?.some((s) => s.includes(subject)),
    ) || null
  );
}

/**
 * Extract a token from email body using regex.
 */
export function extractTokenFromEmail(
  body: string,
  pattern: RegExp,
): string | null {
  const match = body.match(pattern);
  return match ? match[1] : null;
}

/**
 * Wait for an email matching a predicate. Polls every 500ms up to timeout.
 */
export async function waitForEmail(
  predicate: (email: MailHogEmail) => boolean,
  options: { timeout?: number; interval?: number } = {},
): Promise<MailHogEmail | null> {
  const timeout = options.timeout ?? 15_000;
  const interval = options.interval ?? 500;

  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    const data = await getAllEmails().catch(() => ({
      items: [],
      count: 0,
      total: 0,
    }));
    if (data.items) {
      const email = data.items.find(predicate);
      if (email) return email;
    }
    await new Promise((resolve) => setTimeout(resolve, interval));
  }
  return null;
}
