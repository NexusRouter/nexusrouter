import { App as AntdApp, ConfigProvider } from 'antd'
import enUS from 'antd/locale/en_US'
import zhCN from 'antd/locale/zh_CN'
import { Navigate, Outlet, Route, Routes, useLocation, useNavigate } from 'react-router'
import { useTranslation } from 'react-i18next'
import AdminLayout from './layouts/AdminLayout'
import DashboardPage from './pages/Dashboard'
import LoginPage from './pages/Login'
import ApiKeysPage from './pages/ApiKeys'
import SwaggerDebugPage from './pages/SwaggerDebug'
import UpstreamsPage from './pages/Upstreams'
import AccessLogsPage from './pages/AccessLogs'
import RateLimitRulesPage from './pages/RateLimitRules'
import CorsSettingsPage from './pages/CorsSettings'
import IpAccessPage from './pages/IpAccess'
import SystemSettingsPage from './pages/SystemSettings'
import { useAuthStore } from './stores/authStore'
import { useEffect } from 'react'
import { MessageBridge } from './components/MessageBridge'

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

function LocalizedApp() {
  const { i18n } = useTranslation()
  const antLocale = i18n.language.startsWith('en') ? enUS : zhCN
  return (
    <ConfigProvider locale={antLocale}>
      <AntdApp>
        <MessageBridge />
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<RequireAuth />}>
            <Route element={<AdminLayout />}>
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/upstreams" element={<UpstreamsPage />} />
              <Route path="/api-keys" element={<ApiKeysPage />} />
              <Route path="/logs" element={<AccessLogsPage />} />
              <Route path="/rate-limits" element={<RateLimitRulesPage />} />
              <Route path="/cors" element={<CorsSettingsPage />} />
              <Route path="/ip-access" element={<IpAccessPage />} />
              <Route path="/settings" element={<SystemSettingsPage />} />
              <Route path="/debug" element={<SwaggerDebugPage />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </AntdApp>
    </ConfigProvider>
  )
}

/** 根路由：登录与受保护的管理页。 */
export default function App() {
  return <LocalizedApp />
}
