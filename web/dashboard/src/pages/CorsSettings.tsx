import { App, Button, Collapse, Form, Input, InputNumber, Select, Switch, Typography } from 'antd'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'

type Cors = {
  enabled: boolean
  allow_origins: string[]
  allow_methods: string[]
  allow_headers: string[]
  max_age_seconds: number
}

type FormVals = Cors & { allow_origins_bulk?: string; persist: boolean }

type Props = { embedded?: boolean }

function splitBulk(s: string) {
  return s
    .split(/\n|,|;/g)
    .map((x) => x.trim())
    .filter(Boolean)
}

function dedupe(a: string[]) {
  const m = new Set<string>()
  const o: string[] = []
  for (const x of a) {
    if (m.has(x)) continue
    m.add(x)
    o.push(x)
  }
  return o
}

const SUGGEST_HTTP_METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS']
const SUGGEST_HEADERS = [
  'Authorization',
  'Content-Type',
  'X-Request-ID',
  'Accept',
  'Accept-Language',
]

function isReasonableOrigin(s: string): boolean {
  const t = s.trim()
  if (t === '*') return true
  return /^https?:\/\/.+/i.test(t) || /^null$/i.test(t)
}

/** CORS 段编辑与 bulk origins。 */
export default function CorsSettingsPage({ embedded = false }: Props) {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()
  const [form] = Form.useForm<FormVals>()

  const methodOptions = useMemo(
    () =>
      SUGGEST_HTTP_METHODS.map((v) => ({
        value: v,
        label:
          v === 'PATCH'
            ? `${v} (${t('pages.cors.methodHintPatch')})`
            : v === 'DELETE'
              ? `${v} (${t('pages.cors.methodHintDelete')})`
              : v,
      })),
    [t],
  )

  const headerOptions = useMemo(
    () =>
      SUGGEST_HEADERS.map((v) => ({
        value: v,
        label: v,
      })),
    [],
  )

  const q = useQuery({
    queryKey: ['gateway-cors'],
    queryFn: async () => {
      const { data } = await api.get<Cors>('/api/admin/v1/gateway/cors')
      return data
    },
  })

  useEffect(() => {
    if (q.data) {
      form.setFieldsValue({
        ...q.data,
        allow_origins_bulk: '',
        persist: true,
      })
    }
  }, [q.data, form])

  const put = useMutation({
    mutationFn: async (body: Cors & { allow_origins_bulk?: string; persist: boolean }) => {
      await api.put('/api/admin/v1/gateway/cors', body)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      void qc.invalidateQueries({ queryKey: ['gateway-cors'] })
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const originRules = [
    {
      validator: async (_: unknown, value: string[] | undefined) => {
        if (!value?.length) return
        for (const x of value) {
          if (!isReasonableOrigin(x)) {
            throw new Error(t('pages.gatewayPolicy.invalidOrigin', { value: x }))
          }
        }
      },
    },
  ]

  return (
    <div className="space-y-4">
      {!embedded ? (
        <Typography.Title level={4} className="!mb-0">
          {t('pages.cors.title')}
        </Typography.Title>
      ) : null}

      <Form
        form={form}
        layout="vertical"
        disabled={readOnly}
        initialValues={{
          enabled: false,
          allow_origins: [],
          allow_methods: ['GET', 'POST', 'OPTIONS'],
          allow_headers: ['Authorization', 'Content-Type'],
          max_age_seconds: 86400,
          allow_origins_bulk: '',
          persist: true,
        }}
        onFinish={(v) => {
          const extra = splitBulk(String(v.allow_origins_bulk ?? ''))
          const allow_origins = dedupe([...(v.allow_origins ?? []), ...extra])
          put.mutate({
            enabled: v.enabled,
            allow_origins,
            allow_methods: v.allow_methods ?? [],
            allow_headers: v.allow_headers ?? [],
            max_age_seconds: v.max_age_seconds ?? 0,
            allow_origins_bulk: '',
            persist: v.persist,
          })
        }}
      >
        <Form.Item name="enabled" label={t('pages.cors.enable')} valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="allow_origins" label={t('pages.cors.allowOrigins')} rules={originRules}>
          <Select mode="tags" placeholder={t('pages.cors.originsPlaceholder')} tokenSeparators={[',']} />
        </Form.Item>
        <Form.Item name="allow_origins_bulk" label={t('pages.cors.bulkOrigins')}>
          <Input.TextArea rows={3} placeholder={t('pages.cors.bulkPlaceholder')} />
        </Form.Item>
        <Form.Item
          name="max_age_seconds"
          label={t('pages.cors.maxAge')}
          rules={[
            {
              validator: async (_: unknown, v: number | null) => {
                if (v == null || v < 0) throw new Error(t('pages.gatewayPolicy.maxAgeMin'))
              },
            },
          ]}
        >
          <InputNumber min={0} className="w-full" />
        </Form.Item>

        <Collapse
          bordered={false}
          className="bg-transparent"
          items={[
            {
              key: 'advanced',
              label: t('pages.cors.advancedSection'),
              children: (
                <>
                  <Form.Item
                    name="allow_methods"
                    label={t('pages.cors.allowMethods')}
                    extra={<span className="text-xs text-slate-500">{t('pages.cors.suggestMethodsHint')}</span>}
                  >
                    <Select mode="tags" tokenSeparators={[',']} options={methodOptions} />
                  </Form.Item>
                  <Form.Item
                    name="allow_headers"
                    label={t('pages.cors.allowHeaders')}
                    extra={<span className="text-xs text-slate-500">{t('pages.cors.suggestHeadersHint')}</span>}
                  >
                    <Select mode="tags" tokenSeparators={[',']} options={headerOptions} />
                  </Form.Item>
                </>
              ),
            },
          ]}
        />

        <div className="flex flex-col gap-3 border-t border-slate-200/80 pt-4 dark:border-slate-700/80 sm:flex-row sm:items-end sm:justify-between">
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
