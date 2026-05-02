import { App, Button, Form, Input, Select, Spin, Switch, Typography } from 'antd'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'

type IPAccess = {
  mode: string
  cidrs: string[]
}

function splitLines(s: string) {
  return s
    .split(/\n|,|;/g)
    .map((x) => x.trim())
    .filter(Boolean)
}

/** IP 名单：模式与 CIDR 列表，支持批量 PATCH。 */
export default function IpAccessPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()
  const [form] = Form.useForm<IPAccess & { bulk_add: string; persist: boolean }>()
  const [patchForm] = Form.useForm<{ bulk_add: string; bulk_remove: string; persist: boolean }>()

  const modeOptions = useMemo(
    () => [
      { value: 'off', label: t('pages.ipAccess.modeOff') },
      { value: 'allowlist', label: t('pages.ipAccess.modeAllow') },
      { value: 'denylist', label: t('pages.ipAccess.modeDeny') },
    ],
    [t],
  )

  const q = useQuery({
    queryKey: ['ip-access'],
    queryFn: async () => {
      const { data } = await api.get<IPAccess>('/api/admin/v1/security/ip-access')
      return data
    },
  })

  useEffect(() => {
    if (q.data) {
      form.setFieldsValue({ ...q.data, bulk_add: '', persist: true })
    }
  }, [q.data, form])

  const put = useMutation({
    mutationFn: async (body: IPAccess & { persist: boolean }) => {
      await api.put('/api/admin/v1/security/ip-access', body)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      void qc.invalidateQueries({ queryKey: ['ip-access'] })
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const patch = useMutation({
    mutationFn: async (body: { add: string[]; remove: string[]; persist: boolean }) => {
      await api.patch('/api/admin/v1/security/ip-access', body)
    },
    onSuccess: () => {
      message.success(t('common.updated'))
      void qc.invalidateQueries({ queryKey: ['ip-access'] })
    },
    onError: () => message.error(t('common.updateFailed')),
  })

  return (
    <Spin spinning={q.isLoading}>
      <div className="max-w-3xl space-y-6">
        <Typography.Title level={4} className="!mb-0">
          {t('pages.ipAccess.title')}
        </Typography.Title>
        <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
          {t('pages.ipAccess.hint')}
        </Typography.Paragraph>

        <Form
          form={form}
          layout="vertical"
          disabled={readOnly}
          onFinish={(v) => {
            const extra = splitLines(String(v.bulk_add ?? ''))
            const cidrs = Array.from(new Set([...(v.cidrs ?? []), ...extra]))
            put.mutate({ mode: v.mode, cidrs, persist: !!v.persist })
          }}
        >
          <Form.Item name="mode" label={t('pages.ipAccess.mode')} rules={[{ required: true }]}>
            <Select options={modeOptions} />
          </Form.Item>
          <Form.Item name="cidrs" label={t('pages.ipAccess.cidrs')}>
            <Select mode="tags" placeholder={t('pages.ipAccess.cidrPlaceholder')} tokenSeparators={[',']} />
          </Form.Item>
          <Form.Item name="bulk_add" label={t('pages.ipAccess.bulkAdd')}>
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="persist" label={t('pages.ipAccess.persistYaml')} valuePropName="checked">
            <Switch />
          </Form.Item>
          {!readOnly ? (
            <Button type="primary" htmlType="submit" loading={put.isPending}>
              {t('pages.ipAccess.fullSave')}
            </Button>
          ) : null}
        </Form>

        <Typography.Title level={5}>{t('pages.ipAccess.patchTitle')}</Typography.Title>
        <Form
          form={patchForm}
          layout="vertical"
          disabled={readOnly}
          initialValues={{ persist: true }}
          onFinish={(v) => {
            patch.mutate({
              add: splitLines(String(v.bulk_add ?? '')),
              remove: splitLines(String(v.bulk_remove ?? '')),
              persist: !!v.persist,
            })
          }}
        >
          <Form.Item name="bulk_add" label={t('pages.ipAccess.patchAdd')}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="bulk_remove" label={t('pages.ipAccess.patchRemove')}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="persist" label={t('pages.ipAccess.persistShort')} valuePropName="checked">
            <Switch />
          </Form.Item>
          {!readOnly ? (
            <Button type="default" htmlType="submit" loading={patch.isPending}>
              {t('pages.ipAccess.applyPatch')}
            </Button>
          ) : null}
        </Form>
      </div>
    </Spin>
  )
}
