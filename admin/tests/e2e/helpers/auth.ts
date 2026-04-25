/**
 * Auth token helpers for Playwright E2E tests.
 *
 * The admin UI stores tokens in cookies via Zustand (not localStorage):
 *   - fluxbase_admin_token → JSON.stringify(accessToken)
 *   - fluxbase_admin_refresh_token → JSON.stringify(refreshToken)
 *
 * The fluxbase_admin_user object is still stored in localStorage.
 */

const ACCESS_TOKEN_COOKIE = "fluxbase_admin_token";
const REFRESH_TOKEN_COOKIE = "fluxbase_admin_refresh_token";

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

export function getRefreshTokenFromCookies(): string | null {
  const raw = getCookieValue(document.cookie, REFRESH_TOKEN_COOKIE);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return raw;
  }
}

export function clearAuthCookies(): void {
  document.cookie = `${ACCESS_TOKEN_COOKIE}=; path=/; max-age=0; SameSite=Lax`;
  document.cookie = `${REFRESH_TOKEN_COOKIE}=; path=/; max-age=0; SameSite=Lax`;
}

export function setAccessTokenCookie(token: string): void {
  document.cookie = `${ACCESS_TOKEN_COOKIE}=${encodeURIComponent(JSON.stringify(token))}; path=/; max-age=${60 * 60 * 24 * 7}; SameSite=Lax`;
}
