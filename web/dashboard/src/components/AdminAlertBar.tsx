import { Alert } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'

/** 管理端运行态告警条（依赖 /api/admin/v1/alerts/status）。 */
export default function AdminAlertBar() {
  const { t } = useTranslation()
  const q = useQuery({
    queryKey: ['admin-alerts-status'],
    queryFn: async () => {
      const { data } = await api.get<{
        level: string
        reasons: string[]
        enabled: boolean
      }>('/api/admin/v1/alerts/status')
      return data
    },
    refetchInterval: 30_000,
  })
  const level = q.data?.level ?? 'ok'
  if (level === 'ok' || q.isError) {
    return null
  }
  const reasons = (q.data?.reasons ?? []).map((c) => t(`alerts.${c}`, { defaultValue: c }))
  const msg = reasons.join(' · ')
  if (level === 'critical') {
    return (
      <Alert
        banner
        type="error"
        showIcon
        message={`${t('alerts.levelCritical')}: ${msg}`}
        className="mb-3"
      />
    )
  }
  return (
    <Alert
      banner
      type="warning"
      showIcon
      message={`${t('alerts.levelWarning')}: ${msg}`}
      className="mb-3"
    />
  )
}
