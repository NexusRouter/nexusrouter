import { App, Button, Form, Input, InputNumber, Modal, Select, Switch, Table, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'
import { isValidPathPrefix } from '../utils/gatewayValidators'

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

const RATE_LIMIT_DIMENSIONS = ['ip', 'api_key_fp'] as const

type Props = { embedded?: boolean }

/** 限流规则表编辑与整体保存。 */
export default function RateLimitRulesPage({ embedded = false }: Props) {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()
  const [form] = Form.useForm<{ rules: Rule[]; persist: boolean }>()

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
      form.setFieldsValue({ rules: snap.data.rules, persist: true })
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

  const findRuleConflicts = (rules: Rule[]) => {
    const key = (r: Rule) =>
      `${r.dimension}|${(r.match_path_prefix ?? '').trim()}|${r.priority}`
    const seen = new Map<string, number>()
    const dup: string[] = []
    for (const r of rules) {
      if (r.enabled === false) continue
      const k = key(r)
      const n = (seen.get(k) ?? 0) + 1
      seen.set(k, n)
      if (n === 2) dup.push(k)
    }
    return dup
  }

  const submitRules = (rules: Rule[], persist: boolean) => {
    put.mutate({ rules, persist })
  }

  const onFinish = (v: { rules: Rule[]; persist: boolean }) => {
    const rules = v.rules ?? []
    const dups = findRuleConflicts(rules)
    if (dups.length > 0) {
      Modal.confirm({
        title: t('pages.rateLimits.conflictTitle'),
        content: t('pages.rateLimits.conflictBody'),
        okText: t('common.save'),
        onOk: () => submitRules(rules, !!v.persist),
      })
      return
    }
    submitRules(rules, !!v.persist)
  }

  const pathRules = (name: number) => [
    {
      validator: async () => {
        const p = form.getFieldValue(['rules', name, 'match_path_prefix']) as string | undefined
        if (p != null && p !== '' && !isValidPathPrefix(p)) {
          throw new Error(t('pages.gatewayPolicy.invalidPathPrefix'))
        }
      },
    },
  ]

  const burstRules = (name: number) => [
    {
      validator: async () => {
        const b = form.getFieldValue(['rules', name, 'burst']) as number | undefined | null
        if (b == null || b === undefined) return
        if (b < 0) throw new Error(t('pages.gatewayPolicy.burstNegative'))
        if (b !== 0 && b < 1) throw new Error(t('pages.gatewayPolicy.burstMin'))
      },
    },
  ]

  return (
    <div className="space-y-4">
      {!embedded ? (
        <>
          <Typography.Title level={4} className="!mb-0">
            {t('pages.rateLimits.title')}
          </Typography.Title>
          <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
            {t('pages.rateLimits.hint')}
          </Typography.Paragraph>
        </>
      ) : (
        <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
          {t('pages.rateLimits.hint')}
        </Typography.Paragraph>
      )}
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
        initialValues={{ rules: [], persist: true }}
        disabled={readOnly}
        onFinish={onFinish}
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
                    <Form.Item
                      name={[r.name, 'priority']}
                      noStyle
                      rules={[{ required: true, message: t('common.required') }]}
                    >
                      <InputNumber className="w-full" step={1} />
                    </Form.Item>
                  </>
                ),
              },
              {
                title: t('pages.rateLimits.colPathPrefix'),
                render: (_, r) => (
                  <Form.Item name={[r.name, 'match_path_prefix']} noStyle rules={pathRules(r.name)}>
                    <Input placeholder={t('common.allPaths')} />
                  </Form.Item>
                ),
              },
              {
                title: t('pages.rateLimits.colDimension'),
                width: 220,
                render: (_, r) => (
                  <Form.Item noStyle shouldUpdate>
                    {() => {
                      const raw = form.getFieldValue(['rules', r.name, 'dimension']) as string | undefined
                      const labels = new Map<string, string>([
                        ['ip', t('pages.rateLimits.dimOptionIp')],
                        ['api_key_fp', t('pages.rateLimits.dimOptionApiKey')],
                      ])
                      const options = [
                        ...RATE_LIMIT_DIMENSIONS.map((value) => ({
                          value,
                          label: labels.get(value)!,
                        })),
                        ...(raw && !labels.has(raw) ? [{ value: raw, label: raw }] : []),
                      ]
                      return (
                        <Form.Item name={[r.name, 'dimension']} noStyle rules={[{ required: true }]}>
                          <Select className="w-full min-w-[168px]" popupMatchSelectWidth={false} options={options} />
                        </Form.Item>
                      )
                    }}
                  </Form.Item>
                ),
              },
              {
                title: t('pages.rateLimits.colRps'),
                render: (_, r) => (
                  <Form.Item
                    name={[r.name, 'rps']}
                    noStyle
                    rules={[
                      { required: true, message: t('common.required') },
                      {
                        validator: async (_: unknown, v: number | null) => {
                          if (v == null || v < 0.01) throw new Error(t('pages.gatewayPolicy.rpsMin'))
                        },
                      },
                    ]}
                  >
                    <InputNumber min={0.01} step={0.1} className="w-full" />
                  </Form.Item>
                ),
              },
              {
                title: t('pages.rateLimits.colBurst'),
                render: (_, r) => (
                  <Form.Item name={[r.name, 'burst']} noStyle rules={burstRules(r.name)}>
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

        <div className="mt-4 flex flex-col gap-3 border-t border-slate-200/80 pt-4 dark:border-slate-700/80 sm:flex-row sm:items-end sm:justify-between">
          <div className="flex min-w-0 flex-1 flex-col gap-1 sm:flex-row sm:items-center sm:gap-3">
            <span className="shrink-0 text-sm text-slate-600 dark:text-slate-300">
              {t('pages.gatewayPolicy.persistYamlLabel')}
            </span>
            <Form.Item name="persist" valuePropName="checked" noStyle>
              <Switch />
            </Form.Item>
            <Typography.Text type="secondary" className="text-xs">
              {t('pages.gatewayPolicy.persistYamlHintShort')}
            </Typography.Text>
          </div>
          {!readOnly ? (
            <Button type="primary" htmlType="submit" loading={put.isPending}>
              {t('common.save')}
            </Button>
          ) : null}
        </div>
      </Form>
    </div>
  )
}
