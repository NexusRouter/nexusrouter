import { Tabs } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router'
import { useTranslation } from 'react-i18next'

const TAB_PATH: Record<string, string> = {
  'rate-limits': '/gateway/rate-limits',
  cors: '/gateway/cors',
  'ip-access': '/gateway/ip-access',
}

/** 限流 / CORS / IP 聚合：顶栏 Tabs + 子路由出口。 */
export default function GatewayPolicyLayout() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { pathname } = useLocation()

  const activeKey = (() => {
    if (pathname.includes('/gateway/cors')) return 'cors'
    if (pathname.includes('/gateway/ip-access')) return 'ip-access'
    return 'rate-limits'
  })()

  return (
    <div className="space-y-4">
      <Tabs
        activeKey={activeKey}
        onChange={(key) => {
          const path = TAB_PATH[key]
          if (path) navigate(path)
        }}
        items={[
          { key: 'rate-limits', label: t('layout.gatewayTabRateLimits') },
          { key: 'cors', label: t('layout.gatewayTabCors') },
          { key: 'ip-access', label: t('layout.gatewayTabIpAccess') },
        ]}
      />
      <Outlet />
    </div>
  )
}
