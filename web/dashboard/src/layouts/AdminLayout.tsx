import {
  BarChart3,
  Braces,
  Globe,
  KeyRound,
  LogOut,
  ScrollText,
  Server,
  Settings,
  Shield,
  Gauge,
} from 'lucide-react'
import { Layout, Menu, theme } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router'
import { useTranslation } from 'react-i18next'
import AdminAlertBar from '../components/AdminAlertBar'
import LanguageSwitcher from '../components/LanguageSwitcher'
import { useAuthStore } from '../stores/authStore'

const { Header, Sider, Content } = Layout

/** 管理端壳层：侧栏导航、语言切换、运行告警与标题。 */
export default function AdminLayout() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const loc = useLocation()
  const logout = useAuthStore((s) => s.logout)
  const { token } = theme.useToken()

  const selected = (() => {
    if (loc.pathname.startsWith('/upstreams')) {
      return ['upstreams']
    }
    if (loc.pathname.startsWith('/api-keys')) {
      return ['api-keys']
    }
    if (loc.pathname.startsWith('/logs')) {
      return ['logs']
    }
    if (loc.pathname.startsWith('/rate-limits')) {
      return ['rate-limits']
    }
    if (loc.pathname.startsWith('/cors')) {
      return ['cors']
    }
    if (loc.pathname.startsWith('/ip-access')) {
      return ['ip-access']
    }
    if (loc.pathname.startsWith('/settings')) {
      return ['settings']
    }
    if (loc.pathname.startsWith('/debug')) {
      return ['debug']
    }
    return ['dashboard']
  })()

  return (
    <Layout className="min-h-screen">
      <Sider breakpoint="lg" collapsedWidth={64}>
        <div className="flex h-14 items-center justify-center border-b border-white/10 px-2 text-sm font-medium text-white">
          {t('layout.brand')}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selected}
          items={[
            {
              key: 'dashboard',
              icon: <BarChart3 className="h-4 w-4" />,
              label: t('layout.menuDashboard'),
              onClick: () => navigate('/dashboard'),
            },
            {
              key: 'upstreams',
              icon: <Server className="h-4 w-4" />,
              label: t('layout.menuUpstreams'),
              onClick: () => navigate('/upstreams'),
            },
            {
              key: 'api-keys',
              icon: <KeyRound className="h-4 w-4" />,
              label: t('layout.menuApiKeys'),
              onClick: () => navigate('/api-keys'),
            },
            {
              key: 'logs',
              icon: <ScrollText className="h-4 w-4" />,
              label: t('layout.menuLogs'),
              onClick: () => navigate('/logs'),
            },
            {
              key: 'rate-limits',
              icon: <Gauge className="h-4 w-4" />,
              label: t('layout.menuRateLimits'),
              onClick: () => navigate('/rate-limits'),
            },
            {
              key: 'cors',
              icon: <Globe className="h-4 w-4" />,
              label: t('layout.menuCors'),
              onClick: () => navigate('/cors'),
            },
            {
              key: 'ip-access',
              icon: <Shield className="h-4 w-4" />,
              label: t('layout.menuIpAccess'),
              onClick: () => navigate('/ip-access'),
            },
            {
              key: 'settings',
              icon: <Settings className="h-4 w-4" />,
              label: t('layout.menuSettings'),
              onClick: () => navigate('/settings'),
            },
            {
              key: 'debug',
              icon: <Braces className="h-4 w-4" />,
              label: t('layout.menuDebug'),
              onClick: () => navigate('/debug'),
            },
          ]}
        />
      </Sider>
      <Layout>
        <Header
          className="flex items-center justify-between border-b px-6"
          style={{ background: token.colorBgContainer }}
        >
          <h1 className="m-0 text-lg font-semibold text-slate-800">{t('layout.title')}</h1>
          <div className="flex items-center gap-3">
            <LanguageSwitcher />
            <button
              type="button"
              className="inline-flex items-center gap-2 rounded-md border border-slate-200 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
              onClick={() => {
                logout()
                navigate('/login', { replace: true })
              }}
            >
              <LogOut className="h-4 w-4" />
              {t('layout.logout')}
            </button>
          </div>
        </Header>
        <Content className="m-4 rounded-lg bg-white p-6 shadow-sm">
          <AdminAlertBar />
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
