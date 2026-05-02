import { create } from 'zustand'

const TOKEN_KEY = 'nexus_admin_token'
const USER_KEY = 'nexus_admin_remembered_user'
const ROLE_KEY = 'nexus_admin_role'

function readToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

function readRole(): string | null {
  return localStorage.getItem(ROLE_KEY)
}

interface AuthState {
  token: string | null
  role: string | null
  rememberedUsername: string | null
  setSession: (
    token: string,
    options?: { rememberedUsername?: string | null; role?: string | null },
  ) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: readToken(),
  role: readRole(),
  rememberedUsername: localStorage.getItem(USER_KEY),
  setSession: (token, options) => {
    localStorage.setItem(TOKEN_KEY, token)
    const role = options?.role ?? readRole() ?? 'admin'
    localStorage.setItem(ROLE_KEY, role)
    if (options?.rememberedUsername !== undefined) {
      if (options.rememberedUsername) {
        localStorage.setItem(USER_KEY, options.rememberedUsername)
      } else {
        localStorage.removeItem(USER_KEY)
      }
    }
    set({
      token,
      role,
      rememberedUsername:
        options?.rememberedUsername !== undefined
          ? options.rememberedUsername
          : localStorage.getItem(USER_KEY),
    })
  },
  logout: () => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(ROLE_KEY)
    set({ token: null, role: null })
  },
}))

export function isOperatorRole(role: string | null) {
  return (role ?? '').toLowerCase() === 'operator'
}
