import { App, Button, Form, Input, InputNumber, Space, Table, Typography } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'

type LogItem = Record<string, unknown>

type Filters = {
  from: string
  to: string
  path_prefix: string
  status_min?: number
  status_max?: number
  api_key_fp: string
  client_ip: string
  limit: number
  cursor: string
}

function buildSearchParams(f: Filters) {
  const sp = new URLSearchParams()
  if (f.from) sp.set('from', f.from)
  if (f.to) sp.set('to', f.to)
  if (f.path_prefix) sp.set('path_prefix', f.path_prefix)
  if (f.status_min != null) sp.set('status_min', String(f.status_min))
  if (f.status_max != null) sp.set('status_max', String(f.status_max))
  if (f.api_key_fp) sp.set('api_key_fp', f.api_key_fp)
  if (f.client_ip) sp.set('client_ip', f.client_ip)
  sp.set('limit', String(f.limit))
  if (f.cursor) sp.set('cursor', f.cursor)
  return sp
}

/** 代理访问日志查询与 CSV 导出（经 axios 携带 Bearer）。 */
export default function AccessLogsPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const [filters, setFilters] = useState<Filters | null>(null)

  const q = useQuery({
    queryKey: ['access-logs', filters],
    enabled: !!filters,
    queryFn: async () => {
      const sp = buildSearchParams(filters!)
      const { data } = await api.get<{
        items: LogItem[]
        next_cursor: string
        scan_truncated: boolean
      }>(`/api/admin/v1/logs/query?${sp.toString()}`)
      return data
    },
  })

  const columns = useMemo(
    () => [
      { title: t('pages.accessLogs.colTime'), dataIndex: 'ts', key: 'ts', width: 200 },
      { title: t('pages.accessLogs.colPath'), dataIndex: 'path', key: 'path', ellipsis: true },
      { title: t('pages.accessLogs.colStatus'), dataIndex: 'status', key: 'status', width: 72 },
      { title: t('pages.accessLogs.colIp'), dataIndex: 'client_ip', key: 'client_ip', width: 120 },
      {
        title: t('pages.accessLogs.colDuration'),
        dataIndex: 'duration_ms',
        key: 'duration_ms',
        width: 96,
      },
      { title: t('pages.accessLogs.colFp'), dataIndex: 'api_key_fp', key: 'api_key_fp', ellipsis: true },
    ],
    [t],
  )

  const exportCsv = async (f: Filters) => {
    try {
      const sp = buildSearchParams({ ...f, cursor: '' })
      const res = await api.get(`/api/admin/v1/logs/export.csv?${sp.toString()}`, {
        responseType: 'blob',
      })
      const url = URL.createObjectURL(res.data)
      const a = document.createElement('a')
      a.href = url
      a.download = 'access_logs.csv'
      a.click()
      URL.revokeObjectURL(url)
      message.success(t('pages.accessLogs.exportStarted'))
    } catch {
      message.error(t('pages.accessLogs.exportFail'))
    }
  }

  return (
    <div className="space-y-4">
      <Typography.Title level={4} className="!mb-0">
        {t('pages.accessLogs.title')}
      </Typography.Title>
      <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
        {t('pages.accessLogs.hint')}
      </Typography.Paragraph>
      <Form
        layout="vertical"
        initialValues={{
          from: '',
          to: '',
          path_prefix: '',
          limit: 50,
        }}
        onFinish={(vals) => {
          const f: Filters = {
            from: String(vals.from ?? ''),
            to: String(vals.to ?? ''),
            path_prefix: String(vals.path_prefix ?? ''),
            status_min: vals.status_min,
            status_max: vals.status_max,
            api_key_fp: String(vals.api_key_fp ?? ''),
            client_ip: String(vals.client_ip ?? ''),
            limit: Number(vals.limit) || 50,
            cursor: '',
          }
          setFilters(f)
        }}
      >
        <Space wrap className="w-full">
          <Form.Item name="from" label={t('pages.accessLogs.labelFrom')} className="!mb-0 min-w-[200px]">
            <Input placeholder={t('common.optional')} />
          </Form.Item>
          <Form.Item name="to" label={t('pages.accessLogs.labelTo')} className="!mb-0 min-w-[200px]">
            <Input placeholder={t('common.optional')} />
          </Form.Item>
          <Form.Item name="path_prefix" label={t('pages.accessLogs.labelPathPrefix')} className="!mb-0 min-w-[160px]">
            <Input placeholder="/v1/chat" />
          </Form.Item>
          <Form.Item name="status_min" label={t('pages.accessLogs.labelStatusGte')} className="!mb-0 w-28">
            <InputNumber min={100} max={599} className="w-full" />
          </Form.Item>
          <Form.Item name="status_max" label={t('pages.accessLogs.labelStatusLte')} className="!mb-0 w-28">
            <InputNumber min={100} max={599} className="w-full" />
          </Form.Item>
          <Form.Item name="api_key_fp" label={t('pages.accessLogs.labelKeyFp')} className="!mb-0 min-w-[140px]">
            <Input />
          </Form.Item>
          <Form.Item name="client_ip" label={t('pages.accessLogs.labelClientIp')} className="!mb-0 min-w-[120px]">
            <Input />
          </Form.Item>
          <Form.Item name="limit" label={t('pages.accessLogs.labelLimit')} className="!mb-0 w-24">
            <InputNumber min={1} max={500} className="w-full" />
          </Form.Item>
        </Space>
        <Space className="mt-3">
          <Button type="primary" htmlType="submit" loading={q.isFetching && !!filters}>
            {t('common.query')}
          </Button>
          <Button
            disabled={!q.data?.next_cursor || !filters}
            loading={q.isFetching}
            onClick={() => {
              if (!filters || !q.data?.next_cursor) return
              setFilters({ ...filters, cursor: q.data.next_cursor })
            }}
          >
            {t('common.nextPage')}
          </Button>
          <Button
            type="default"
            disabled={!filters}
            onClick={() => {
              if (!filters) {
                message.warning(t('pages.accessLogs.warnQueryFirst'))
                return
              }
              void exportCsv(filters)
            }}
          >
            {t('pages.accessLogs.exportCsv')}
          </Button>
        </Space>
      </Form>
      {q.isError ? (
        <Typography.Text type="danger">{t('pages.accessLogs.queryFail')}</Typography.Text>
      ) : null}
      <Table
        size="small"
        rowKey={(_, i) => String(i)}
        loading={q.isFetching}
        dataSource={(q.data?.items ?? []) as LogItem[]}
        columns={columns}
        pagination={false}
      />
      {q.data?.scan_truncated ? (
        <Typography.Text type="warning">{t('pages.accessLogs.scanTruncated')}</Typography.Text>
      ) : null}
    </div>
  )
}
