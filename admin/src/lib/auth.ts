import { useAuthStore } from "@/stores/auth-store";

const USER_KEY = "fluxbase_admin_user";

export interface AdminUser {
  id: string;
  email: string;
  role: string;
  email_verified: boolean;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface DashboardUser {
  id: string;
  email: string;
  email_verified: boolean;
  full_name: string | null;
  avatar_url: string | null;
  totp_enabled: boolean;
  is_active: boolean;
  is_locked: boolean;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
  role?: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  const token = useAuthStore.getState().auth.accessToken;
  return token || null;
}

export function getRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  const token = useAuthStore.getState().auth.refreshToken;
  return token || null;
}

export function getStoredUser(): AdminUser | DashboardUser | null {
  if (typeof window === "undefined") return null;
  const userJson =
    localStorage.getItem(USER_KEY) || localStorage.getItem("user");
  if (!userJson) return null;
  try {
    return JSON.parse(userJson);
  } catch {
    return null;
  }
}

export function setTokens(tokens: TokenPair, user: AdminUser): void {
  if (typeof window === "undefined") return;
  const store = useAuthStore.getState();
  store.auth.setTokens(tokens.access_token, tokens.refresh_token);
  store.auth.setUser({
    accountNo: user.id,
    email: user.email,
    role: [user.role],
    exp: Date.now() + tokens.expires_in * 1000,
  });
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function clearTokens(): void {
  if (typeof window === "undefined") return;
  useAuthStore.getState().auth.reset();
  localStorage.removeItem(USER_KEY);
  localStorage.removeItem("user");
  localStorage.removeItem("fluxbase_admin_access_token");
  localStorage.removeItem("fluxbase_admin_refresh_token");
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
}

export function isAuthenticated(): boolean {
  return !!getAccessToken();
}

export function logout(): void {
  clearTokens();
  window.location.href = "/admin/login";
}
