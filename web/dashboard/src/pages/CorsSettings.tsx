import { App, Button, Form, Input, InputNumber, Select, Space, Switch, Typography } from 'antd'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
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

/** CORS 段编辑与 bulk origins。 */
export default function CorsSettingsPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()
  const [form] = Form.useForm<FormVals>()

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

  return (
    <div className="max-w-3xl space-y-4">
      <Typography.Title level={4} className="!mb-0">
        {t('pages.cors.title')}
      </Typography.Title>
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
        <Form.Item name="allow_origins" label={t('pages.cors.allowOrigins')}>
          <Select mode="tags" placeholder={t('pages.cors.originsPlaceholder')} tokenSeparators={[',']} />
        </Form.Item>
        <Form.Item name="allow_origins_bulk" label={t('pages.cors.bulkOrigins')}>
          <Input.TextArea rows={3} placeholder={t('pages.cors.bulkPlaceholder')} />
        </Form.Item>
        <Form.Item name="allow_methods" label={t('pages.cors.allowMethods')}>
          <Select mode="tags" tokenSeparators={[',']} />
        </Form.Item>
        <Form.Item name="allow_headers" label={t('pages.cors.allowHeaders')}>
          <Select mode="tags" tokenSeparators={[',']} />
        </Form.Item>
        <Form.Item name="max_age_seconds" label={t('pages.cors.maxAge')}>
          <InputNumber min={0} className="w-full" />
        </Form.Item>
        <Form.Item name="persist" label={t('settings.persist')} valuePropName="checked">
          <Switch />
        </Form.Item>
        {!readOnly ? (
          <Space>
            <Button type="primary" htmlType="submit" loading={put.isPending}>
              {t('common.save')}
            </Button>
          </Space>
        ) : null}
      </Form>
    </div>
  )
}
