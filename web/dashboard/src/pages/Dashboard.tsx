import { Card, Col, Row, Statistic, Table, Typography } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { PageError } from '../components/PageError'
import { api } from '../services/api'

/** 仪表盘：聚合指标与健康状态。 */
export default function DashboardPage() {
  const { t } = useTranslation()
  const q = useQuery({
    queryKey: ['admin-metrics'],
    queryFn: async () => {
      const { data } = await api.get<Record<string, unknown>>(
        '/api/admin/v1/metrics/summary',
      )
      return data
    },
    refetchInterval: 15_000,
  })

  const m = q.data
  const errToday = (m?.errors_today_by_code as Record<string, number>) ?? {}

  if (q.isError) {
    return (
      <PageError
        title={t('pages.dashboard.loadError')}
        retryLabel={t('common.pageError.retry')}
        onRetry={() => q.refetch()}
      />
    )
  }

  return (
    <div className="space-y-4">
      <Typography.Title level={4} className="!mb-0">
        {t('pages.dashboard.title')}
      </Typography.Title>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={q.isLoading}>
            <Statistic
              title={t('pages.dashboard.onlineStatus')}
              value={m?.online ? t('pages.dashboard.online') : t('pages.dashboard.unknown')}
              valueStyle={{ color: m?.online ? '#16a34a' : '#94a3b8' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={q.isLoading}>
            <Statistic
              title={t('pages.dashboard.currentRps')}
              value={Number(m?.current_rps_estimate ?? 0).toFixed(2)}
              suffix={t('pages.dashboard.reqPerS')}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={q.isLoading}>
            <Statistic
              title={t('pages.dashboard.successRate')}
              value={((Number(m?.success_rate ?? 0) || 0) * 100).toFixed(2)}
              suffix={t('pages.dashboard.percent')}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={q.isLoading}>
            <Statistic
              title={t('pages.dashboard.avgLatency')}
              value={Number(m?.avg_latency_ms ?? 0).toFixed(1)}
              suffix={t('pages.dashboard.ms')}
            />
          </Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card title={t('pages.dashboard.todayYesterday')} loading={q.isLoading}>
            <Row gutter={16}>
              <Col span={12}>
                <Statistic title={t('pages.dashboard.todayReq')} value={Number(m?.requests_today ?? 0)} />
              </Col>
              <Col span={12}>
                <Statistic
                  title={t('pages.dashboard.yesterdayReq')}
                  value={Number(m?.requests_yesterday ?? 0)}
                />
              </Col>
            </Row>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title={t('pages.dashboard.errorsToday')} loading={q.isLoading}>
            <Table
              size="small"
              pagination={false}
              dataSource={Object.entries(errToday).map(([code, c]) => ({
                key: code,
                code,
                count: c,
              }))}
              columns={[
                { title: t('pages.dashboard.colCode'), dataIndex: 'code' },
                { title: t('pages.dashboard.colCount'), dataIndex: 'count' },
              ]}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
