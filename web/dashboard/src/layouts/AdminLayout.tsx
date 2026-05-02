import type { ReactNode } from 'react'
import {
  BarChart3,
  KeyRound,
  LogOut,
  Menu,
  ScrollText,
  Server,
  Settings,
  SlidersHorizontal,
  Sparkles,
  Boxes,
} from 'lucide-react'
import { Button, Drawer, Layout, Tooltip, theme } from 'antd'
import { useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router'
import { useTranslation } from 'react-i18next'
import AdminAlertBar from '../components/AdminAlertBar'
import LanguageSwitcher from '../components/LanguageSwitcher'
import ThemeSwitcher from '../components/ThemeSwitcher'
import { useAuthStore } from '../stores/authStore'

const { Content } = Layout

type NavItem = { key: string; path: string; labelKey: string; icon: ReactNode }

/** 管理端壳层：渐变顶栏、桌面侧栏与移动抽屉导航；主题由全局 ConfigProvider 控制。 */
export default function AdminLayout() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const loc = useLocation()
  const logout = useAuthStore((s) => s.logout)
  const { token } = theme.useToken()
  const [mobileOpen, setMobileOpen] = useState(false)

  const items: NavItem[] = [
    { key: 'dashboard', path: '/dashboard', labelKey: 'layout.menuDashboard', icon: <BarChart3 className="h-4 w-4" /> },
    { key: 'upstreams', path: '/upstreams', labelKey: 'layout.menuUpstreams', icon: <Server className="h-4 w-4" /> },
    { key: 'model-library', path: '/model-library', labelKey: 'layout.menuModelLibrary', icon: <Boxes className="h-4 w-4" /> },
    { key: 'api-keys', path: '/api-keys', labelKey: 'layout.menuApiKeys', icon: <KeyRound className="h-4 w-4" /> },
    { key: 'logs', path: '/logs', labelKey: 'layout.menuLogs', icon: <ScrollText className="h-4 w-4" /> },
    {
      key: 'gateway-policy',
      path: '/gateway/policy',
      labelKey: 'layout.menuGatewayPolicy',
      icon: <SlidersHorizontal className="h-4 w-4" />,
    },
    { key: 'settings', path: '/settings', labelKey: 'layout.menuSettings', icon: <Settings className="h-4 w-4" /> },
  ]

  const selected = (() => {
    const p = loc.pathname
    if (p.startsWith('/model-library')) {
      return 'model-library'
    }
    if (p.startsWith('/gateway')) {
      return 'gateway-policy'
    }
    for (const it of items) {
      if (it.key !== 'dashboard' && p.startsWith(it.path)) {
        return it.key
      }
    }
    return 'dashboard'
  })()

  const navigateTo = (path: string) => {
    navigate(path)
    setMobileOpen(false)
  }

  const NavList = ({ onPick }: { onPick?: () => void }) => (
    <nav className="flex flex-col gap-1.5">
      {items.map((it) => {
        const active = selected === it.key
        return (
          <button
            key={it.key}
            type="button"
            onClick={() => {
              navigateTo(it.path)
              onPick?.()
            }}
            className={`group flex items-center gap-3 rounded-xl px-3 py-2.5 text-left text-sm font-semibold transition-all duration-300 active:scale-95 ${
              active
                ? 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white shadow-lg shadow-indigo-500/25'
                : 'text-slate-700 hover:scale-[1.02] hover:bg-white/80 hover:text-indigo-600 hover:shadow-md hover:shadow-indigo-100/50 dark:text-slate-200 dark:hover:bg-slate-800/80 dark:hover:text-indigo-300'
            } `}
          >
            <span
              className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-lg transition-all ${
                active
                  ? 'bg-white/20 text-white'
                  : 'bg-slate-100 text-slate-600 group-hover:bg-indigo-100 group-hover:text-indigo-600 dark:bg-slate-800 dark:text-slate-400 dark:group-hover:bg-indigo-900/50 dark:group-hover:text-indigo-300'
              }`}
            >
              {it.icon}
            </span>
            <span className="relative z-10 tracking-wide">{t(it.labelKey)}</span>
          </button>
        )
      })}
    </nav>
  )

  const SidebarInner = () => (
    <div className="flex h-full flex-col gap-y-5 overflow-y-auto border-r border-indigo-200/40 bg-gradient-to-br from-white via-indigo-50/30 to-white px-4 pb-6 shadow-xl backdrop-blur-sm dark:border-slate-700/60 dark:from-slate-900 dark:via-indigo-950/40 dark:to-slate-900">
      <div className="shrink-0 border-b border-indigo-200/40 pb-4 pt-5 dark:border-slate-700/60">
        <button
          type="button"
          onClick={() => navigateTo('/dashboard')}
          className="group relative flex w-full items-center gap-3 overflow-hidden rounded-xl border-2 border-indigo-100/60 bg-gradient-to-br from-indigo-50/50 via-purple-50/30 to-white p-3 text-left transition-all duration-300 hover:scale-[1.02] hover:border-indigo-300 hover:from-indigo-100/80 hover:via-purple-100/60 hover:shadow-lg hover:shadow-indigo-200/30 active:scale-95 dark:border-indigo-800/50 dark:from-slate-800/80 dark:via-indigo-950/50 dark:to-slate-900 dark:hover:border-indigo-500/50"
        >
          <span className="relative flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-gradient-to-br from-amber-400 via-orange-400 to-orange-500 shadow-md ring-2 ring-white/50 dark:ring-slate-700/50">
            <Sparkles className="h-6 w-6 text-white drop-shadow" aria-hidden />
          </span>
          <span className="min-w-0 flex-1">
            <span className="block truncate bg-gradient-to-r from-indigo-600 via-purple-600 to-indigo-600 bg-clip-text text-base font-extrabold text-transparent dark:from-indigo-400 dark:via-purple-400 dark:to-indigo-400">
              {t('layout.brand')}
            </span>
            <span className="mt-0.5 block truncate text-[10px] font-medium text-slate-500 transition-colors group-hover:text-indigo-600 dark:text-slate-400 dark:group-hover:text-indigo-300">
              NexusRouter
            </span>
          </span>
        </button>
      </div>
      <NavList />
    </div>
  )

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-indigo-50/25 to-slate-50 dark:from-slate-950 dark:via-indigo-950/30 dark:to-slate-950">
      <header className="fixed left-0 right-0 top-0 z-[100] border-b border-white/10 bg-white/80 backdrop-blur-md dark:border-slate-800/80 dark:bg-slate-950/85 lg:left-60">
        <div className="mx-auto max-w-7xl px-2 py-1.5 sm:px-4 lg:px-6 xl:px-8">
          <div className="relative overflow-hidden rounded-lg border border-indigo-400/20 bg-gradient-to-r from-indigo-600 via-purple-600 to-indigo-600 shadow-lg shadow-indigo-500/20">
            <div
              className="pointer-events-none absolute inset-0 opacity-30"
              style={{
                backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cpath d='M36 18c3.314 0 6 2.686 6 6s-2.686 6-6 6-6-2.686-6-6 2.686-6 6-6z' stroke='%23fff' stroke-opacity='.05'/%3E%3C/g%3E%3C/svg%3E")`,
              }}
            />
            <div className="pointer-events-none absolute -left-16 -top-16 h-32 w-32 animate-pulse-slow rounded-full bg-purple-400/20 blur-3xl" />
            <div
              className="pointer-events-none absolute -bottom-16 -right-16 h-32 w-32 animate-pulse-slow rounded-full bg-indigo-400/20 blur-3xl"
              style={{ animationDelay: '1s' }}
            />
            <div className="relative flex items-center gap-2 p-2 sm:gap-3 sm:p-2.5">
              <Button
                type="text"
                className="!text-white hover:!bg-white/10 lg:!hidden"
                icon={<Menu className="h-5 w-5" />}
                onClick={() => setMobileOpen(true)}
                aria-label="Open menu"
              />
              <div className="relative shrink-0">
                <span className="relative flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-amber-400 via-orange-400 to-orange-500 shadow-md sm:h-9 sm:w-9">
                  <Sparkles className="h-4 w-4 text-white drop-shadow sm:h-5 sm:w-5" aria-hidden />
                </span>
              </div>
              <p className="hidden min-w-0 flex-1 font-semibold leading-relaxed text-white drop-shadow-sm md:block md:overflow-hidden md:text-ellipsis md:whitespace-nowrap md:text-sm">
                {t('layout.topBannerHint')}
              </p>
              <p className="max-w-[40%] flex-1 truncate text-xs font-semibold leading-snug text-white md:hidden">
                {t('layout.brand')}
              </p>
              <div className="ml-auto flex shrink-0 items-center gap-1 sm:gap-2">
                <ThemeSwitcher />
                <LanguageSwitcher />
                <Tooltip title={t('layout.logout')} placement="bottom">
                  <Button
                    type="text"
                    aria-label={t('layout.logout')}
                    className="!flex !h-8 !min-w-8 !items-center !justify-center !rounded-full !border-0 !bg-white/10 !p-1.5 !text-white hover:!bg-white/20"
                    icon={<LogOut className="h-[18px] w-[18px]" aria-hidden />}
                    onClick={() => {
                      logout()
                      navigate('/login', { replace: true })
                    }}
                  />
                </Tooltip>
              </div>
            </div>
            <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/30 to-transparent" />
          </div>
        </div>
      </header>

      <div className="h-[72px] sm:h-[76px]" />

      <aside className="hidden lg:fixed lg:bottom-0 lg:left-0 lg:top-0 lg:z-40 lg:flex lg:w-60 lg:flex-col">
        <SidebarInner />
      </aside>

      <Drawer
        title={t('layout.brand')}
        placement="left"
        width={280}
        onClose={() => setMobileOpen(false)}
        open={mobileOpen}
        styles={{ body: { padding: 0 } }}
        className="lg:hidden"
      >
        <div className="bg-gradient-to-br from-white via-indigo-50/30 to-white px-3 pb-6 pt-2 dark:from-slate-900 dark:via-indigo-950/40 dark:to-slate-900">
          <NavList onPick={() => setMobileOpen(false)} />
        </div>
      </Drawer>

      <Layout className="bg-transparent lg:pl-60">
        <Content className="px-3 py-4 sm:px-5 lg:px-8 lg:py-6" style={{ minHeight: 'calc(100vh - 72px)' }}>
          <div
            className="rounded-2xl border border-slate-200/80 p-4 shadow-sm backdrop-blur-sm sm:p-6 dark:border-slate-700/80"
            style={{ background: token.colorBgContainer }}
          >
            <AdminAlertBar />
            <Outlet />
          </div>
        </Content>
      </Layout>
    </div>
  )
}
