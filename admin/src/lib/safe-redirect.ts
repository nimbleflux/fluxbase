/**
 * Validates a redirect target taken from a URL parameter (e.g. ?return_to=).
 *
 * Returns a safe relative path (starting with a single "/") or null when the
 * value must be rejected. Same-origin absolute URLs are normalized to their
 * relative path so server-issued redirects (e.g. the MCP OAuth authorize URL
 * passed as return_to) keep working; anything cross-origin, protocol-relative
 * ("//evil.com"), a backslash trick ("/\evil.com") or a non-HTTP scheme
 * returns null so the caller can fall back to its default landing route.
 */
export function safeRelativePath(
  value: string | null | undefined
): string | null {
  if (!value || typeof window === 'undefined') return null

  let url: URL
  try {
    url = new URL(value, window.location.origin)
  } catch {
    return null
  }

  // Rejects cross-origin URLs, protocol-relative hosts and backslash tricks
  // (the URL parser treats "\" as "/" so "/\evil.com" resolves to another origin)
  if (url.origin !== window.location.origin) return null
  // javascript:, data: and other non-HTTP schemes resolve with a null origin
  // but must still be rejected
  if (url.protocol !== 'http:' && url.protocol !== 'https:') return null

  return url.pathname + url.search + url.hash
}
