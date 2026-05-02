import {
  App,
  Alert,
  Button,
  Card,
  Collapse,
  Tooltip,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Switch,
  Tag,
  Typography,
} from 'antd'
import axios from 'axios'
import type { TFunction } from 'i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'
import { systemSettingsFieldLabelI18nKey } from '../utils/systemSettingsLabels'

type SettingField = {
  key: string
  value: unknown
  mutability: string
  hint?: string
}

const LOG_LEVELS = ['info', 'error'] as const

function formatMutabilityShort(t: TFunction, m: string): string {
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

function mutabilityLong(t: TFunction, m: string, key: string): string {
  if (m === 'hot_reload') {
    if (key.startsWith('proxy_access_log_')) {
      return t('settings.mutabilityLongHotReloadLog')
    }
    return t('settings.mutabilityLongHotReload')
  }
  if (m === 'restart_required') {
    return t('settings.mutabilityLongRestart')
  }
  if (m === 'read_only') {
    return t('settings.mutabilityLongReadOnly')
  }
  return ''
}

function fieldBusinessLabel(t: TFunction, key: string): string {
  const suffix = systemSettingsFieldLabelI18nKey(key)
  if (suffix) {
    return t(`settings.fieldLabels.${suffix}`)
  }
  return t('settings.fieldLabels.unknown', { key })
}

function saveErrorMessage(err: unknown, fallback: string): string {
  if (axios.isAxiosError(err)) {
    const d = err.response?.data as { message?: string; error?: string } | undefined
    const msg = d?.message ?? d?.error
    if (msg && typeof msg === 'string' && msg.trim()) {
      return msg
    }
  }
  return fallback
}

function SettingStatusRow({ field, t }: { field: SettingField; t: TFunction }) {
  const long = mutabilityLong(t, field.mutability, field.key)
  const short = formatMutabilityShort(t, field.mutability)
  const label = fieldBusinessLabel(t, field.key)
  const tagColor =
    field.mutability === 'hot_reload' ? 'green' : field.mutability === 'restart_required' ? 'orange' : 'default'

  return (
    <div className="border-b border-slate-200 py-3 last:border-b-0 dark:border-slate-600">
      <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
        <span className="min-w-0 font-medium text-slate-800 dark:text-slate-100">{label}</span>
        <Tag color={tagColor}>{short}</Tag>
      </div>
      <div className="mt-1 text-sm text-slate-700 dark:text-slate-200">
        <span className="text-slate-500 dark:text-slate-400">{t('settings.currentValueLabel')}</span>{' '}
        <span className="break-all font-mono text-slate-800 dark:text-slate-100">{String(field.value)}</span>
      </div>
      <Typography.Paragraph type="secondary" className="!mb-0 !mt-1 text-xs leading-relaxed">
        {long}
        {field.hint ? (
          <>
            {' '}
            <span className="text-slate-500 dark:text-slate-400">{field.hint}</span>
          </>
        ) : null}
      </Typography.Paragraph>
    </div>
  )
}

/** 系统设置：运行状态 + 代理日志表单 + 高级键名（折叠）。 */
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
    onError: (err) => message.error(saveErrorMessage(err, t('settings.saveFail'))),
  })

  const settings = q.data?.settings ?? []

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div>
        <Typography.Title level={3} className="!mb-1">
          {t('settings.title')}
        </Typography.Title>
        <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
          {t('settings.pageSubtitle')}
        </Typography.Paragraph>
      </div>

      {readOnly ? (
        <Alert type="warning" showIcon message={t('settings.operatorReadOnly')} />
      ) : null}
      {q.isError ? <Alert type="error" showIcon message={t('settings.loadFail')} /> : null}

      <Card title={t('settings.sectionStatus')} loading={q.isLoading}>
        {settings.length === 0 && !q.isLoading ? (
          <Typography.Text type="secondary">{t('settings.statusEmpty')}</Typography.Text>
        ) : (
          <div className="px-0">
            {settings.map((s) => (
              <SettingStatusRow key={s.key} field={s} t={t} />
            ))}
          </div>
        )}
      </Card>

      <Card title={t('settings.sectionLog')}>
        <Typography.Paragraph type="secondary" className="!mb-3 text-sm">
          {t('settings.logFormLinkedHint')}
        </Typography.Paragraph>
        <Alert type="info" showIcon className="mb-4" message={t('settings.rotationNote')} />
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
          <Form.Item
            name="max_size_mb"
            label={t('settings.proxyLogMaxSize')}
            extra={<span className="text-xs text-slate-500">{t('settings.rotationFieldExtra')}</span>}
          >
            <InputNumber min={1} className="w-full max-w-md" />
          </Form.Item>
          <Form.Item
            name="max_backups"
            label={t('settings.proxyLogMaxBackups')}
            extra={<span className="text-xs text-slate-500">{t('settings.rotationFieldExtra')}</span>}
          >
            <InputNumber min={1} className="w-full max-w-md" />
          </Form.Item>
          <Form.Item
            name="persist"
            label={t('settings.persist')}
            valuePropName="checked"
            extra={<span className="text-xs text-slate-500">{t('settings.persistHelp')}</span>}
          >
            <Switch defaultChecked />
          </Form.Item>
          <Space>
            <Tooltip title={readOnly ? t('settings.operatorSaveTooltip') : undefined}>
              <span className="inline-block">
                <Button type="primary" htmlType="submit" loading={put.isPending} disabled={readOnly}>
                  {t('settings.save')}
                </Button>
              </span>
            </Tooltip>
          </Space>
        </Form>
      </Card>

      <Collapse
        bordered={false}
        className="bg-transparent"
        items={[
          {
            key: 'advanced',
            label: t('settings.sectionAdvanced'),
            children: (
              <div className="space-y-3">
                <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
                  {t('settings.advancedIntro')}
                </Typography.Paragraph>
                {settings.map((s) => (
                  <div
                    key={`adv-${s.key}`}
                    className="rounded border border-slate-200 bg-slate-50/80 p-3 text-sm dark:border-slate-600 dark:bg-slate-900/40"
                  >
                    <div className="font-medium text-slate-700 dark:text-slate-200">
                      {fieldBusinessLabel(t, s.key)}
                    </div>
                    <code className="mt-1 block break-all text-xs text-slate-600 dark:text-slate-300">{s.key}</code>
                    {s.hint ? (
                      <Typography.Paragraph type="secondary" className="!mb-0 !mt-1 text-xs">
                        {s.hint}
                      </Typography.Paragraph>
                    ) : null}
                  </div>
                ))}
              </div>
            ),
          },
        ]}
      />
    </div>
  )
}
