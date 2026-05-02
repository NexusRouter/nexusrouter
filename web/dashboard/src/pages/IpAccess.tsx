import { App, Button, Form, Input, Modal, Select, Spin, Switch, Typography } from 'antd'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useMemo, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'
import { isValidIpOrCidrToken, validateCidrListInput } from '../utils/gatewayValidators'

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

type Props = { embedded?: boolean }

/** IP 名单：模式与 CIDR 列表，支持批量 PATCH。 */
export default function IpAccessPage({ embedded = false }: Props) {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()
  const [form] = Form.useForm<IPAccess & { bulk_add: string; persist: boolean }>()
  const [patchForm] = Form.useForm<{ bulk_add: string; bulk_remove: string; persist: boolean }>()
  const loadedMode = useRef<string | undefined>(undefined)

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
      loadedMode.current = q.data.mode
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

  const runFullSave = (v: IPAccess & { bulk_add: string; persist: boolean }) => {
    const extra = splitLines(String(v.bulk_add ?? ''))
    const cidrs = Array.from(new Set([...(v.cidrs ?? []), ...extra]))
    put.mutate({ mode: v.mode, cidrs, persist: !!v.persist })
  }

  const onFullFinish = (v: IPAccess & { bulk_add: string; persist: boolean }) => {
    const prev = loadedMode.current
    const nextMode = v.mode
    if (prev === 'off' && (nextMode === 'allowlist' || nextMode === 'denylist')) {
      Modal.confirm({
        title: t('pages.ipAccess.modeChangeConfirmTitle'),
        content: t('pages.ipAccess.modeChangeConfirmBody'),
        okText: t('common.save'),
        onOk: () => runFullSave(v),
      })
      return
    }
    runFullSave(v)
  }

  const cidrTagRules = [
    {
      validator: async (_: unknown, value: string[] | undefined) => {
        if (!value?.length) return
        for (const x of value) {
          if (!isValidIpOrCidrToken(x)) {
            throw new Error(t('pages.gatewayPolicy.invalidCidr', { value: x }))
          }
        }
      },
    },
  ]

  const bulkValidator = {
    validator: async (_: unknown, value: string | undefined) => {
      const r = validateCidrListInput(String(value ?? ''))
      if (r !== true) throw new Error(t('pages.gatewayPolicy.invalidCidr', { value: r }))
    },
  }

  const inner = (
    <>
      {!embedded ? (
        <>
          <Typography.Title level={4} className="!mb-0">
            {t('pages.ipAccess.title')}
          </Typography.Title>
          <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
            {t('pages.ipAccess.hint')}
          </Typography.Paragraph>
        </>
      ) : (
        <Typography.Paragraph type="secondary" className="!mb-4 text-sm">
          {t('pages.ipAccess.hint')}
        </Typography.Paragraph>
      )}

      <Form
        form={form}
        layout="vertical"
        disabled={readOnly}
        onFinish={onFullFinish}
      >
        <Form.Item name="mode" label={t('pages.ipAccess.mode')} rules={[{ required: true }]}>
          <Select options={modeOptions} disabled={readOnly} />
        </Form.Item>
        <Form.Item name="cidrs" label={t('pages.ipAccess.cidrs')} rules={cidrTagRules}>
          <Select mode="tags" placeholder={t('pages.ipAccess.cidrPlaceholder')} tokenSeparators={[',']} />
        </Form.Item>
        <Form.Item name="bulk_add" label={t('pages.ipAccess.bulkAdd')} rules={[bulkValidator]}>
          <Input.TextArea rows={3} />
        </Form.Item>

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

      <Typography.Title level={5} className="mt-8">
        {t('pages.ipAccess.patchTitle')}
      </Typography.Title>
      <Typography.Paragraph type="secondary" className="text-sm">
        {t('pages.ipAccess.patchIntro')}
      </Typography.Paragraph>

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
        <Form.Item name="bulk_add" label={t('pages.ipAccess.patchAdd')} rules={[bulkValidator]}>
          <Input.TextArea rows={2} />
        </Form.Item>
        <Form.Item name="bulk_remove" label={t('pages.ipAccess.patchRemove')} rules={[bulkValidator]}>
          <Input.TextArea rows={2} />
        </Form.Item>

        <div className="flex flex-col gap-3 border-t border-slate-200/80 pt-4 dark:border-slate-700/80 sm:flex-row sm:items-center sm:justify-between">
          <Form.Item name="persist" className="!mb-0" valuePropName="checked">
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-sm text-slate-600 dark:text-slate-300">{t('pages.gatewayPolicy.persistYamlLabel')}</span>
              <Switch disabled={readOnly} />
            </div>
          </Form.Item>
          {!readOnly ? (
            <Button type="default" htmlType="submit" loading={patch.isPending}>
              {t('pages.ipAccess.applyPatch')}
            </Button>
          ) : null}
        </div>
      </Form>
    </>
  )

  return (
    <Spin spinning={q.isLoading}>
      <div className={embedded ? '' : 'max-w-3xl space-y-6'}>
        <div className="space-y-6">{inner}</div>
      </div>
    </Spin>
  )
}
