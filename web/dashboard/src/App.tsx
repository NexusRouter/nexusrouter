import { App as AntdApp, ConfigProvider, Spin, theme as antdTheme } from 'antd'
import enUS from 'antd/locale/en_US'
import zhCN from 'antd/locale/zh_CN'
import { Navigate, Outlet, Route, Routes, useLocation, useNavigate } from 'react-router'
import { useTranslation } from 'react-i18next'
import AdminLayout from './layouts/AdminLayout'
import DashboardPage from './pages/Dashboard'
import LoginPage from './pages/Login'
import SetupPage from './pages/Setup'
import ApiKeysPage from './pages/ApiKeys'
import UpstreamsPage from './pages/Upstreams'
import AccessLogsPage from './pages/AccessLogs'
import GatewayPolicyPage from './pages/GatewayPolicyPage'
import SystemSettingsPage from './pages/SystemSettings'
import ModelLibraryPage from './pages/ModelLibrary'
import { useAuthStore } from './stores/authStore'
import { useEffect, useLayoutEffect, useState } from 'react'
import axios from 'axios'
import { MessageBridge } from './components/MessageBridge'
import { fetchBootstrapStatus } from './services/bootstrap'
import { useResolvedDark } from './hooks/useResolvedDark'

/** 未登录时跳转登录页，并带上 redirect。 */
function RequireAuth() {
  const token = useAuthStore((s) => s.token)
  const loc = useLocation()
  const navigate = useNavigate()

  useEffect(() => {
    if (!token) {
      navigate('/login', {
        replace: true,
        state: { from: loc.pathname + loc.search },
      })
    }
  }, [token, loc.pathname, loc.search, navigate])

  if (!token) {
    return null
  }
  return <Outlet />
}

/** 根据网关全局初始化状态重定向：未完成则仅允许 /setup；完成后禁止访问 /setup。 */
function BootstrapShell() {
  const loc = useLocation()
  /** 当前路径下是否已拿到 definitive 的 bootstrap 结论；路由一变先置 false，避免沿用旧路径的 initialized 误判或短暂渲染 /setup。 */
  const [ready, setReady] = useState(false)
  const [initialized, setInitialized] = useState(false)

  useLayoutEffect(() => {
    setReady(false)
  }, [loc.pathname])

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const st = await fetchBootstrapStatus()
        if (!cancelled) {
          setInitialized(st.initialized)
        }
      } catch (err) {
        // 仅当网关确实无该路由（404）时视为旧版本、跳过向导；其它错误（代理失败、5xx 等）走向导，避免静默当作「已初始化」。
        if (!cancelled) {
          const st = axios.isAxiosError(err) ? err.response?.status : undefined
          setInitialized(st === 404)
        }
      } finally {
        if (!cancelled) {
          setReady(true)
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [loc.pathname])

  if (!ready) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50 dark:bg-slate-950">
        <Spin size="large" />
      </div>
    )
  }

  if (!initialized && loc.pathname !== '/setup') {
    return <Navigate to="/setup" replace />
  }
  if (initialized && loc.pathname === '/setup') {
    return <Navigate to="/login" replace />
  }
  return <Outlet />
}

function LocalizedApp() {
  const { i18n } = useTranslation()
  const antLocale = i18n.language.startsWith('en') ? enUS : zhCN
  const dark = useResolvedDark()
  return (
    <ConfigProvider
      locale={antLocale}
      theme={{
        algorithm: dark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        token: {
          colorPrimary: '#6366f1',
          borderRadiusLG: 12,
        },
      }}
    >
      <AntdApp>
        <MessageBridge />
        <Routes>
          <Route element={<BootstrapShell />}>
            <Route path="/setup" element={<SetupPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route element={<RequireAuth />}>
              <Route element={<AdminLayout />}>
                <Route path="/" element={<Navigate to="/dashboard" replace />} />
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/upstreams" element={<UpstreamsPage />} />
                <Route path="/model-library" element={<ModelLibraryPage />} />
                <Route path="/api-keys" element={<ApiKeysPage />} />
                <Route path="/logs" element={<AccessLogsPage />} />
                <Route path="/gateway/policy" element={<GatewayPolicyPage />} />
                <Route
                  path="/gateway/rate-limits"
                  element={<Navigate to="/gateway/policy#section-rate-limits" replace />}
                />
                <Route path="/gateway/cors" element={<Navigate to="/gateway/policy#section-cors" replace />} />
                <Route
                  path="/gateway/ip-access"
                  element={<Navigate to="/gateway/policy#section-ip-access" replace />}
                />
                <Route path="/gateway" element={<Navigate to="/gateway/policy" replace />} />
                <Route path="/settings" element={<SystemSettingsPage />} />
              </Route>
            </Route>
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Route>
        </Routes>
      </AntdApp>
    </ConfigProvider>
  )
}

/** 根路由：登录与受保护的管理页。 */
export default function App() {
  return <LocalizedApp />
}
