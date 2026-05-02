import { App, Button, Form, Input, InputNumber, Space, Switch, Table, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'

type Rule = {
  /** 服务端生成；表单用隐藏域回传，不在界面展示 */
  id?: string
  priority: number
  match_path_prefix: string
  dimension: string
  rps: number
  burst: number
  enabled: boolean
}

type FieldRow = { key: number; name: number }

/** 限流规则表编辑与整体保存。 */
export default function RateLimitRulesPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()
  const [form] = Form.useForm<{ rules: Rule[] }>()

  const snap = useQuery({
    queryKey: ['rate-limit-rules'],
    queryFn: async () => {
      const { data } = await api.get<{ rules: Rule[]; rate_limit: { rps_per_ip: number; rps_per_key: number } }>(
        '/api/admin/v1/gateway/rate-limit-rules',
      )
      return data
    },
  })

  useEffect(() => {
    if (snap.data?.rules) {
      form.setFieldsValue({ rules: snap.data.rules })
    }
  }, [snap.data, form])

  const put = useMutation({
    mutationFn: async (body: { rules: Rule[]; persist: boolean }) => {
      await api.put('/api/admin/v1/gateway/rate-limit-rules', body)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      void qc.invalidateQueries({ queryKey: ['rate-limit-rules'] })
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  return (
    <div className="space-y-4">
      <Typography.Title level={4} className="!mb-0">
        {t('pages.rateLimits.title')}
      </Typography.Title>
      <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
        {t('pages.rateLimits.hint')}
      </Typography.Paragraph>
      {snap.data?.rate_limit ? (
        <Typography.Text type="secondary">
          {t('pages.rateLimits.globalLine', {
            ip: snap.data.rate_limit.rps_per_ip,
            key: snap.data.rate_limit.rps_per_key,
          })}
        </Typography.Text>
      ) : null}
      <Form
        form={form}
        layout="vertical"
        initialValues={{ rules: [] }}
        disabled={readOnly}
        onFinish={(v) => put.mutate({ rules: v.rules ?? [], persist: true })}
      >
        <Form.List name="rules">
          {(fields, { add, remove }) => {
            const cols: ColumnsType<FieldRow> = [
              {
                title: t('pages.rateLimits.colPriority'),
                render: (_, r) => (
                  <>
                    <Form.Item name={[r.name, 'id']} hidden>
                      <Input type="hidden" />
                    </Form.Item>
                    <Form.Item name={[r.name, 'priority']} noStyle>
                      <InputNumber className="w-full" />
                    </Form.Item>
                  </>
                ),
              },
              {
                title: t('pages.rateLimits.colPathPrefix'),
                render: (_, r) => (
                  <Form.Item name={[r.name, 'match_path_prefix']} noStyle>
                    <Input placeholder={t('common.allPaths')} />
                  </Form.Item>
                ),
              },
              {
                title: t('pages.rateLimits.colDimension'),
                render: (_, r) => (
                  <Form.Item name={[r.name, 'dimension']} noStyle rules={[{ required: true }]}>
                    <Input placeholder={t('pages.rateLimits.dimPlaceholder')} />
                  </Form.Item>
                ),
              },
              {
                title: t('pages.rateLimits.colRps'),
                render: (_, r) => (
                  <Form.Item name={[r.name, 'rps']} noStyle rules={[{ required: true }]}>
                    <InputNumber min={0.01} step={0.1} className="w-full" />
                  </Form.Item>
                ),
              },
              {
                title: t('pages.rateLimits.colBurst'),
                render: (_, r) => (
                  <Form.Item name={[r.name, 'burst']} noStyle>
                    <InputNumber min={0} className="w-full" />
                  </Form.Item>
                ),
              },
              {
                title: t('pages.rateLimits.colEnabled'),
                render: (_, r) => (
                  <Form.Item name={[r.name, 'enabled']} valuePropName="checked" noStyle>
                    <Switch />
                  </Form.Item>
                ),
              },
              ...(readOnly
                ? []
                : [
                    {
                      title: '',
                      key: 'del',
                      width: 64,
                      render: (_: unknown, r: FieldRow) => (
                        <Button type="link" danger onClick={() => remove(r.name)}>
                          {t('pages.rateLimits.removeRow')}
                        </Button>
                      ),
                    },
                  ]),
            ]
            return (
              <div className="space-y-2">
                <Table<FieldRow>
                  size="small"
                  pagination={false}
                  loading={snap.isLoading}
                  dataSource={fields}
                  rowKey="key"
                  columns={cols}
                />
                {!readOnly ? (
                  <Button
                    type="dashed"
                    onClick={() =>
                      add({
                        id: '',
                        priority: 0,
                        match_path_prefix: '',
                        dimension: 'ip',
                        rps: 1,
                        burst: 0,
                        enabled: true,
                      })
                    }
                  >
                    {t('pages.rateLimits.addRule')}
                  </Button>
                ) : null}
              </div>
            )
          }}
        </Form.List>
        {!readOnly ? (
          <Space className="mt-4">
            <Button type="primary" htmlType="submit" loading={put.isPending}>
              {t('pages.rateLimits.saveDisk')}
            </Button>
          </Space>
        ) : null}
      </Form>
    </div>
  )
}
