import { App, Button, Form, Input, InputNumber, Select, Space, Switch, Typography } from 'antd'
import type { TFunction } from 'i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'

type SettingField = {
  key: string
  value: unknown
  mutability: string
  hint?: string
}

const LOG_LEVELS = ['info', 'error'] as const

function formatMutability(t: TFunction, m: string): string {
  switch (m) {
    case 'hot_reload':
      return t('settings.mutabilityHotReload')
    case 'restart_required':
      return t('settings.mutabilityRestart')
    case 'read_only':
      return t('settings.mutabilityReadOnly')
    default:
      return m
  }
}

/** 系统设置：只读聚合 + 管理员可写 proxy_access_log。 */
export default function SystemSettingsPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const qc = useQueryClient()
  const role = useAuthStore((s) => s.role)
  const readOnly = isOperatorRole(role)

  const q = useQuery({
    queryKey: ['system-settings'],
    queryFn: async () => {
      const { data } = await api.get<{ settings: SettingField[] }>(
        '/api/admin/v1/system/settings',
      )
      return data
    },
  })

  const [form] = Form.useForm<{
    enabled: boolean
    path: string
    level: string
    max_size_mb: number
    max_backups: number
    persist: boolean
  }>()

  useEffect(() => {
    const list = q.data?.settings
    if (!list?.length) return
    const map = Object.fromEntries(list.map((s) => [s.key, s.value]))
    form.setFieldsValue({
      enabled: Boolean(map.proxy_access_log_enabled),
      path: String(map.proxy_access_log_path ?? ''),
      level: String(map.proxy_access_log_level ?? 'info'),
    })
  }, [q.data, form])

  const put = useMutation({
    mutationFn: async (body: {
      proxy_access_log: {
        enabled: boolean
        path: string
        level: string
        max_size_mb: number
        max_backups: number
      }
      persist: boolean
    }) => {
      await api.put('/api/admin/v1/system/settings', body)
    },
    onSuccess: () => {
      message.success(t('settings.saveOk'))
      void qc.invalidateQueries({ queryKey: ['system-settings'] })
    },
    onError: () => message.error(t('settings.saveFail')),
  })

  return (
    <div className="max-w-3xl space-y-4">
      <Typography.Title level={4}>{t('settings.title')}</Typography.Title>
      <Typography.Paragraph type="secondary">{t('settings.intro')}</Typography.Paragraph>
      {readOnly ? (
        <Typography.Text type="warning">{t('settings.operatorReadOnly')}</Typography.Text>
      ) : null}
      {q.isError ? <Typography.Text type="danger">{t('settings.loadFail')}</Typography.Text> : null}
      <div className="space-y-2 rounded border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-900/40">
        {(q.data?.settings ?? []).map((s) => (
          <div key={s.key} className="flex flex-wrap items-baseline gap-x-2 gap-y-1 text-sm">
            <span className="font-medium text-slate-700 dark:text-slate-200">{t('settings.configKey')}:</span>
            <code className="rounded bg-white px-1 dark:bg-slate-800">{s.key}</code>
            <span className="font-medium text-slate-700 dark:text-slate-200">{t('settings.configValue')}:</span>
            <span className="text-slate-600 dark:text-slate-300">{String(s.value)}</span>
            <Typography.Text type="secondary">
              ({formatMutability(t, s.mutability)})
            </Typography.Text>
            {s.hint ? <span className="w-full text-slate-500">{s.hint}</span> : null}
          </div>
        ))}
      </div>
      <Typography.Title level={5}>{t('settings.save')}</Typography.Title>
      <Form
        form={form}
        layout="vertical"
        disabled={readOnly}
        initialValues={{
          enabled: false,
          path: '',
          level: 'info',
          max_size_mb: 100,
          max_backups: 3,
          persist: true,
        }}
        onFinish={(v) =>
          put.mutate({
            proxy_access_log: {
              enabled: v.enabled,
              path: v.path,
              level: v.level,
              max_size_mb: v.max_size_mb,
              max_backups: v.max_backups,
            },
            persist: v.persist,
          })
        }
      >
        <Form.Item name="enabled" label={t('settings.proxyLogEnabled')} valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="path" label={t('settings.proxyLogPath')}>
          <Input />
        </Form.Item>
        <Form.Item noStyle shouldUpdate>
          {() => {
            const lv = form.getFieldValue('level') as string | undefined
            const base = LOG_LEVELS.map((value) => ({
              value,
              label: value === 'info' ? t('settings.logLevelInfo') : t('settings.logLevelError'),
            }))
            const extra =
              lv && !(LOG_LEVELS as readonly string[]).includes(lv) ? [{ value: lv, label: lv }] : []
            return (
              <Form.Item name="level" label={t('settings.proxyLogLevel')}>
                <Select className="w-full max-w-md" options={[...base, ...extra]} />
              </Form.Item>
            )
          }}
        </Form.Item>
        <Form.Item name="max_size_mb" label={t('settings.proxyLogMaxSize')}>
          <InputNumber min={1} className="w-full max-w-md" />
        </Form.Item>
        <Form.Item name="max_backups" label={t('settings.proxyLogMaxBackups')}>
          <InputNumber min={1} className="w-full max-w-md" />
        </Form.Item>
        <Form.Item name="persist" label={t('settings.persist')} valuePropName="checked">
          <Switch defaultChecked />
        </Form.Item>
        <Space>
          <Button type="primary" htmlType="submit" loading={put.isPending} disabled={readOnly}>
            {t('settings.save')}
          </Button>
        </Space>
      </Form>
    </div>
  )
}
