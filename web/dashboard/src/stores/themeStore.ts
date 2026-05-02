import { create } from 'zustand'

export type ThemeMode = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'nexus_theme_mode'

function readStored(): ThemeMode {
  try {
    const v = localStorage.getItem(STORAGE_KEY)
    if (v === 'light' || v === 'dark' || v === 'system') {
      return v
    }
  } catch {
    /* ignore */
  }
  return 'system'
}

interface ThemeState {
  mode: ThemeMode
  setMode: (m: ThemeMode) => void
}

export const useThemeStore = create<ThemeState>((set) => ({
  mode: typeof localStorage !== 'undefined' ? readStored() : 'system',
  setMode: (m) => {
    try {
      localStorage.setItem(STORAGE_KEY, m)
    } catch {
      /* ignore */
    }
    set({ mode: m })
  },
}))
