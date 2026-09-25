import { create } from 'zustand'
import { getCookie, setCookie, removeCookie } from '@/lib/cookies'
import { setAuthToken as setFluxbaseAuthToken } from '@/lib/fluxbase-client'

const AUTH_COOKIE_NAME = 'fluxbase_admin_token'
// The refresh token is no longer persisted in a JavaScript-readable cookie:
// the server sets it as an HttpOnly cookie (`fluxbase_admin_refresh`) and the
// client keeps a session-scoped copy (per browser tab) for API calls. The
// legacy cookie name is kept only to migrate/clean up older sessions.
const REFRESH_SESSION_KEY = 'fluxbase_admin_refresh_token'
const LEGACY_REFRESH_COOKIE_NAME = 'fluxbase_admin_refresh_token'

interface AuthUser {
  accountNo: string
  email: string
  role: string[]
  exp: number
}

interface AuthState {
  auth: {
    user: AuthUser | null
    setUser: (user: AuthUser | null) => void
    accessToken: string
    refreshToken: string
    setAccessToken: (accessToken: string) => void
    setTokens: (accessToken: string, refreshToken: string) => void
    resetAccessToken: () => void
    reset: () => void
  }
}

function parseCookieToken(name: string): string {
  try {
    const raw = getCookie(name)
    return raw ? JSON.parse(raw) : ''
  } catch {
    removeCookie(name)
    return ''
  }
}

function getSessionRefreshToken(): string {
  if (typeof window === 'undefined') return ''
  try {
    const raw = sessionStorage.getItem(REFRESH_SESSION_KEY)
    return raw ? JSON.parse(raw) : ''
  } catch {
    return ''
  }
}

function setSessionRefreshToken(token: string): void {
  if (typeof window === 'undefined') return
  try {
    sessionStorage.setItem(REFRESH_SESSION_KEY, JSON.stringify(token))
  } catch {
    // sessionStorage may be unavailable (privacy mode); the HttpOnly cookie
    // set by the server keeps the refresh flow working without it.
  }
}

function clearSessionRefreshToken(): void {
  if (typeof window === 'undefined') return
  try {
    sessionStorage.removeItem(REFRESH_SESSION_KEY)
  } catch {
    // ignore
  }
}

// Migrates sessions from the legacy JS refresh cookie into sessionStorage.
function migrateLegacyRefreshCookie(): string {
  const legacy = parseCookieToken(LEGACY_REFRESH_COOKIE_NAME)
  if (legacy) {
    removeCookie(LEGACY_REFRESH_COOKIE_NAME)
    setSessionRefreshToken(legacy)
  }
  return legacy
}

export const useAuthStore = create<AuthState>()((set) => {
  const initToken = parseCookieToken(AUTH_COOKIE_NAME)
  const initRefreshToken = getSessionRefreshToken() || migrateLegacyRefreshCookie()

  if (initToken) {
    setFluxbaseAuthToken(initToken)
  }

  return {
    auth: {
      user: null,
      setUser: (user) =>
        set((state) => ({ ...state, auth: { ...state.auth, user } })),
      accessToken: initToken,
      refreshToken: initRefreshToken,
      setAccessToken: (accessToken) =>
        set((state) => {
          setCookie(AUTH_COOKIE_NAME, JSON.stringify(accessToken))
          setFluxbaseAuthToken(accessToken)
          return { ...state, auth: { ...state.auth, accessToken } }
        }),
      setTokens: (accessToken, refreshToken) =>
        set((state) => {
          setCookie(AUTH_COOKIE_NAME, JSON.stringify(accessToken))
          setSessionRefreshToken(refreshToken)
          setFluxbaseAuthToken(accessToken)
          return {
            ...state,
            auth: { ...state.auth, accessToken, refreshToken },
          }
        }),
      resetAccessToken: () =>
        set((state) => {
          removeCookie(AUTH_COOKIE_NAME)
          setFluxbaseAuthToken(null)
          return { ...state, auth: { ...state.auth, accessToken: '' } }
        }),
      reset: () =>
        set((state) => {
          removeCookie(AUTH_COOKIE_NAME)
          // Clean up both the session copy and the legacy JS cookie.
          clearSessionRefreshToken()
          removeCookie(LEGACY_REFRESH_COOKIE_NAME)
          setFluxbaseAuthToken(null)
          return {
            ...state,
            auth: {
              ...state.auth,
              user: null,
              accessToken: '',
              refreshToken: '',
            },
          }
        }),
    },
  }
})
