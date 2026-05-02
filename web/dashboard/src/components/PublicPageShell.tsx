import type { ReactNode } from 'react'
import LanguageSwitcher from './LanguageSwitcher'
import ThemeSwitcher from './ThemeSwitcher'

/** 登录 / 引导等未进管理壳页面的公共背景与右上角主题、语言。 */
export default function PublicPageShell({ children }: { children: ReactNode }) {
  return (
    <div className="relative min-h-screen overflow-x-hidden bg-gradient-to-br from-slate-100 via-indigo-50/50 to-purple-50/30 dark:from-slate-950 dark:via-indigo-950/50 dark:to-slate-950">
      <div
        className="pointer-events-none absolute inset-0 opacity-40 dark:opacity-25"
        style={{
          backgroundImage:
            'radial-gradient(ellipse 80% 50% at 50% -20%, rgba(99, 102, 241, 0.35), transparent), radial-gradient(ellipse 60% 40% at 100% 50%, rgba(168, 85, 247, 0.2), transparent)',
        }}
      />
      <div className="absolute right-2 top-2 z-10 flex flex-wrap items-center justify-end gap-2 sm:right-4 sm:top-4">
        <ThemeSwitcher />
        <LanguageSwitcher />
      </div>
      <div className="relative flex min-h-screen items-center justify-center p-4 pt-16 sm:pt-4">{children}</div>
    </div>
  )
}
