import { create } from 'zustand'

export type ImpersonationType = 'user' | 'anon' | 'service'

export interface ImpersonationSession {
  id: string
  admin_user_id: string
  target_user_id?: string
  impersonation_type: ImpersonationType
  target_role?: string
  reason: string
  started_at: string
  ip_address?: string
  user_agent?: string
  is_active: boolean
  tenant_id?: string
}

export interface ImpersonatedUser {
  id: string
  email: string
  role?: string
}

interface ImpersonationState {
  isImpersonating: boolean
  impersonationType: ImpersonationType | null
  impersonationToken: string | null
  impersonationRefreshToken: string | null
  impersonatedUser: ImpersonatedUser | null
  session: ImpersonationSession | null
  impersonationTenantId: string | null

  // Actions
  startImpersonation: (
    token: string,
    refreshToken: string,
    user: ImpersonatedUser,
    session: ImpersonationSession,
    type: ImpersonationType
  ) => void
  stopImpersonation: () => void
  updateSession: (session: ImpersonationSession) => void
}

const STORAGE_KEYS = {
  TOKEN: 'fluxbase_impersonation_token',
  REFRESH_TOKEN: 'fluxbase_impersonation_refresh_token',
  USER: 'fluxbase_impersonated_user',
  SESSION: 'fluxbase_impersonation_session',
  TYPE: 'fluxbase_impersonation_type',
  TENANT_ID: 'fluxbase_impersonation_tenant_id',
}

const loadFromStorage = () => {
  try {
    const token = localStorage.getItem(STORAGE_KEYS.TOKEN)
    const refreshToken = localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)
    const userStr = localStorage.getItem(STORAGE_KEYS.USER)
    const sessionStr = localStorage.getItem(STORAGE_KEYS.SESSION)
    const typeStr = localStorage.getItem(STORAGE_KEYS.TYPE)
    const tenantId = localStorage.getItem(STORAGE_KEYS.TENANT_ID)

    if (token && userStr && sessionStr && typeStr) {
      return {
        isImpersonating: true,
        impersonationType: typeStr as ImpersonationType,
        impersonationToken: token,
        impersonationRefreshToken: refreshToken,
        impersonatedUser: JSON.parse(userStr),
        session: JSON.parse(sessionStr),
        impersonationTenantId: tenantId,
      }
    }
  } catch {
    // Failed to load impersonation state from storage - will use default state
  }

  return {
    isImpersonating: false,
    impersonationType: null,
    impersonationToken: null,
    impersonationRefreshToken: null,
    impersonatedUser: null,
    session: null,
    impersonationTenantId: null,
  }
}

export const useImpersonationStore = create<ImpersonationState>((set) => ({
  ...loadFromStorage(),

  startImpersonation: (token, refreshToken, user, session, type) => {
    localStorage.setItem(STORAGE_KEYS.TOKEN, token)
    if (refreshToken) {
      localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refreshToken)
    }
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user))
    localStorage.setItem(STORAGE_KEYS.SESSION, JSON.stringify(session))
    localStorage.setItem(STORAGE_KEYS.TYPE, type)

    const tenantId = session?.tenant_id || null
    if (tenantId) {
      localStorage.setItem(STORAGE_KEYS.TENANT_ID, tenantId)
    }

    set({
      isImpersonating: true,
      impersonationType: type,
      impersonationToken: token,
      impersonationRefreshToken: refreshToken,
      impersonatedUser: user,
      session,
      impersonationTenantId: tenantId,
    })
  },

  stopImpersonation: () => {
    localStorage.removeItem(STORAGE_KEYS.TOKEN)
    localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN)
    localStorage.removeItem(STORAGE_KEYS.USER)
    localStorage.removeItem(STORAGE_KEYS.SESSION)
    localStorage.removeItem(STORAGE_KEYS.TYPE)
    localStorage.removeItem(STORAGE_KEYS.TENANT_ID)

    set({
      isImpersonating: false,
      impersonationType: null,
      impersonationToken: null,
      impersonationRefreshToken: null,
      impersonatedUser: null,
      session: null,
      impersonationTenantId: null,
    })
  },

  updateSession: (session) => {
    localStorage.setItem(STORAGE_KEYS.SESSION, JSON.stringify(session))
    set({ session })
  },
}))
