import { useEffect, useState } from 'react'
import type { ThemeMode } from '../stores/themeStore'
import { useThemeStore } from '../stores/themeStore'

function compute(mode: ThemeMode): boolean {
  if (mode === 'dark') {
    return true
  }
  if (mode === 'light') {
    return false
  }
  if (typeof window === 'undefined') {
    return false
  }
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

/** 解析 light/dark/system，并同步 `document.documentElement` 的 `dark` class（供 Tailwind dark: 使用）。 */
export function useResolvedDark(): boolean {
  const mode = useThemeStore((s) => s.mode)
  const [dark, setDark] = useState(() => compute(useThemeStore.getState().mode))

  useEffect(() => {
    setDark(compute(mode))
  }, [mode])

  useEffect(() => {
    if (mode !== 'system') {
      return
    }
    const mql = window.matchMedia('(prefers-color-scheme: dark)')
    const onChange = () => setDark(mql.matches)
    mql.addEventListener('change', onChange)
    return () => mql.removeEventListener('change', onChange)
  }, [mode])

  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark)
  }, [dark])

  return dark
}
