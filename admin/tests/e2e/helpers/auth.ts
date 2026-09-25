/**
 * Auth token helpers for Playwright E2E tests.
 *
 * The admin UI stores the access token in a cookie via Zustand:
 *   - fluxbase_admin_token → JSON.stringify(accessToken)
 *
 * The refresh token lives in sessionStorage (`fluxbase_admin_refresh_token`)
 * plus an HttpOnly cookie set by the server; it is never written to a
 * JavaScript-readable cookie anymore.
 *
 * The fluxbase_admin_user object is still stored in localStorage.
 */

const ACCESS_TOKEN_COOKIE = "fluxbase_admin_token";
const REFRESH_SESSION_KEY = "fluxbase_admin_refresh_token";

function getCookieValue(cookieStr: string, name: string): string | null {
  const prefix = `${name}=`;
  const parts = cookieStr.split("; ");
  for (const part of parts) {
    if (part.startsWith(prefix)) {
      return part.substring(prefix.length);
    }
  }
  return null;
}

export function getAccessTokenFromCookies(): string | null {
  const raw = getCookieValue(document.cookie, ACCESS_TOKEN_COOKIE);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return raw;
  }
}

export function getRefreshTokenFromSession(): string | null {
  const raw = sessionStorage.getItem(REFRESH_SESSION_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return raw;
  }
}

export function clearAuthState(): void {
  document.cookie = `${ACCESS_TOKEN_COOKIE}=; path=/; max-age=0; SameSite=Lax`;
  sessionStorage.removeItem(REFRESH_SESSION_KEY);
}

export function setAccessTokenCookie(token: string): void {
  document.cookie = `${ACCESS_TOKEN_COOKIE}=${encodeURIComponent(JSON.stringify(token))}; path=/; max-age=${60 * 60 * 24 * 7}; SameSite=Lax`;
}
