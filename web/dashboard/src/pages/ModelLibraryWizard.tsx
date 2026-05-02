import { App, Button, Form, Input, InputNumber, Modal, Radio, Select, Steps } from 'antd'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'

type Vendor = {
  id: number
  vendor_name: string
  vendor_type: number
  vendor_code: string
  logo?: string | null
  status: number
}

type Base = {
  id: number
  model_name: string
  model_code: string
  model_type: number
  capability?: unknown
  sort: number
  status: number
}

type Upstream = {
  id: number
  vendor_id: number
  upstream_name: string
  base_url: string
  api_key_set?: boolean
  timeout: number
  max_concurrent: number
  status: number
}

function invalidateMl(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['ml-vendors'] })
  qc.invalidateQueries({ queryKey: ['ml-bases'] })
  qc.invalidateQueries({ queryKey: ['ml-upstreams'] })
  qc.invalidateQueries({ queryKey: ['ml-instances'] })
}

export function ModelLibraryWizard({
  open,
  onClose,
}: {
  open: boolean
  onClose: () => void
}) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const qc = useQueryClient()

  const [step, setStep] = useState(0)
  const [vendorMode, setVendorMode] = useState<'new' | 'existing'>('new')
  const [upstreamMode, setUpstreamMode] = useState<'new' | 'existing'>('new')
  const [baseMode, setBaseMode] = useState<'new' | 'existing'>('new')

  const [vendorId, setVendorId] = useState<number | null>(null)
  const [upstreamId, setUpstreamId] = useState<number | null>(null)
  const [baseId, setBaseId] = useState<number | null>(null)

  const [vendorForm] = Form.useForm()
  const [upForm] = Form.useForm()
  const [baseForm] = Form.useForm()
  const [instForm] = Form.useForm()

  const vendorsQ = useQuery({
    queryKey: ['ml-vendors'],
    queryFn: async () => {
      const { data } = await api.get<{ items: Vendor[] }>('/api/admin/v1/model-library/vendors')
      return data.items
    },
    enabled: open,
  })
  const basesQ = useQuery({
    queryKey: ['ml-bases'],
    queryFn: async () => {
      const { data } = await api.get<{ items: Base[] }>('/api/admin/v1/model-library/bases')
      return data.items
    },
    enabled: open,
  })
  const upsQ = useQuery({
    queryKey: ['ml-upstreams'],
    queryFn: async () => {
      const { data } = await api.get<{ items: Upstream[] }>('/api/admin/v1/model-library/upstreams')
      return data.items
    },
    enabled: open,
  })

  useEffect(() => {
    if (!open) return
    setStep(0)
    setVendorMode('new')
    setUpstreamMode('new')
    setBaseMode('new')
    setVendorId(null)
    setUpstreamId(null)
    setBaseId(null)
    vendorForm.resetFields()
    upForm.resetFields()
    baseForm.resetFields()
    instForm.resetFields()
    vendorForm.setFieldsValue({ vendor_type: 1, status: 1 })
    upForm.setFieldsValue({ timeout: 30, max_concurrent: 100, status: 1 })
    baseForm.setFieldsValue({ model_type: 1, sort: 0, status: 1 })
    instForm.setFieldsValue({ weight: 10, priority: 1, is_official: 0, status: 1 })
  }, [open, vendorForm, upForm, baseForm, instForm])

  const createVendor = useMutation({
    mutationFn: async (body: Record<string, unknown>) => {
      const { data } = await api.post<Vendor>('/api/admin/v1/model-library/vendors', body)
      return data
    },
    onSuccess: () => invalidateMl(qc),
  })
  const createUp = useMutation({
    mutationFn: async (body: Record<string, unknown>) => {
      const { data } = await api.post<Upstream>('/api/admin/v1/model-library/upstreams', body)
      return data
    },
    onSuccess: () => invalidateMl(qc),
  })
  const createBase = useMutation({
    mutationFn: async (body: Record<string, unknown>) => {
      const { data } = await api.post<Base>('/api/admin/v1/model-library/bases', body)
      return data
    },
    onSuccess: () => invalidateMl(qc),
  })
  const createInst = useMutation({
    mutationFn: async (body: Record<string, unknown>) => {
      const { data } = await api.post<{ id: number }>('/api/admin/v1/model-library/instances', body)
      return data
    },
    onSuccess: () => invalidateMl(qc),
  })

  const vendorOptions = (vendorsQ.data ?? []).map((v) => ({
    value: v.id,
    label: `${v.vendor_code} — ${v.vendor_name}`,
  }))

  const upstreamOptionsFiltered = (upsQ.data ?? [])
    .filter((u) => (vendorId ? u.vendor_id === vendorId : true))
    .map((u) => ({
      value: u.id,
      label: `${u.id} — ${u.upstream_name} (${u.base_url})`,
    }))

  const baseOptions = (basesQ.data ?? []).map((b) => ({
    value: b.id,
    label: `${b.model_code} — ${b.model_name}`,
  }))

  const goNext = async () => {
    try {
      if (step === 0) {
        if (vendorMode === 'existing') {
          const id = vendorForm.getFieldValue('existing_vendor_id') as number | undefined
          if (id == null) {
            message.warning(t('pages.modelLibrary.wizardPickRequired'))
            return
          }
          setVendorId(id)
          setStep(1)
          return
        }
        const v = await vendorForm.validateFields()
        const row = await createVendor.mutateAsync(v)
        setVendorId(row.id)
        setStep(1)
        return
      }
      if (step === 1) {
        if (!vendorId) {
          message.error(t('pages.modelLibrary.wizardMissingVendor'))
          return
        }
        if (upstreamMode === 'existing') {
          const id = upForm.getFieldValue('existing_upstream_id') as number | undefined
          if (id == null) {
            message.warning(t('pages.modelLibrary.wizardPickRequired'))
            return
          }
          setUpstreamId(id)
          setStep(2)
          return
        }
        const v = await upForm.validateFields()
        const body = { ...v, vendor_id: vendorId }
        const row = await createUp.mutateAsync(body)
        setUpstreamId(row.id)
        setStep(2)
        return
      }
      if (step === 2) {
        if (baseMode === 'existing') {
          const id = baseForm.getFieldValue('existing_base_id') as number | undefined
          if (id == null) {
            message.warning(t('pages.modelLibrary.wizardPickRequired'))
            return
          }
          setBaseId(id)
          setStep(3)
          return
        }
        const v = await baseForm.validateFields()
        const row = await createBase.mutateAsync(v)
        setBaseId(row.id)
        setStep(3)
        return
      }
    } catch {
      message.error(t('common.saveFailed'))
    }
  }

  const goPrev = () => {
    if (step > 0) setStep((s) => s - 1)
  }

  const finish = async () => {
    try {
      const v = await instForm.validateFields()
      if (vendorId == null || upstreamId == null || baseId == null) {
        message.error(t('pages.modelLibrary.wizardMissingDeps'))
        return
      }
      await createInst.mutateAsync({
        ...v,
        base_model_id: baseId,
        vendor_id: vendorId,
        upstream_id: upstreamId,
      })
      message.success(t('common.saved'))
      invalidateMl(qc)
      onClose()
    } catch {
      message.error(t('common.saveFailed'))
    }
  }

  const loading =
    createVendor.isPending || createUp.isPending || createBase.isPending || createInst.isPending

  const stepItems = [
    { title: t('pages.modelLibrary.wizardStepVendor') },
    { title: t('pages.modelLibrary.wizardStepUpstream') },
    { title: t('pages.modelLibrary.wizardStepBase') },
    { title: t('pages.modelLibrary.wizardStepInstance') },
  ]

  return (
    <Modal
      title={t('pages.modelLibrary.wizardTitle')}
      open={open}
      onCancel={onClose}
      width={640}
      footer={null}
      destroyOnClose
    >
      <Steps current={step} items={stepItems} className="mb-6" />

      {step === 0 && (
        <div className="space-y-4">
          <Radio.Group
            value={vendorMode}
            onChange={(e) => setVendorMode(e.target.value)}
            optionType="button"
            buttonStyle="solid"
          >
            <Radio.Button value="new">{t('pages.modelLibrary.wizardModeNew')}</Radio.Button>
            <Radio.Button value="existing">{t('pages.modelLibrary.wizardModeExisting')}</Radio.Button>
          </Radio.Group>
          {vendorMode === 'existing' ? (
            <Form form={vendorForm} layout="vertical">
              <Form.Item name="existing_vendor_id" label={t('pages.modelLibrary.vendor')} rules={[{ required: true }]}>
                <Select showSearch optionFilterProp="label" options={vendorOptions} placeholder={t('pages.modelLibrary.wizardPickVendor')} />
              </Form.Item>
            </Form>
          ) : (
            <Form form={vendorForm} layout="vertical">
              <Form.Item name="vendor_name" label={t('pages.modelLibrary.vendorName')} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="vendor_type" label={t('pages.modelLibrary.vendorType')} rules={[{ required: true }]} initialValue={1}>
                <Select
                  options={[
                    { value: 1, label: t('pages.modelLibrary.typeOfficial') },
                    { value: 2, label: t('pages.modelLibrary.typeThirdParty') },
                  ]}
                />
              </Form.Item>
              <Form.Item name="vendor_code" label={t('pages.modelLibrary.vendorCode')} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="logo" label={t('pages.modelLibrary.logoUrl')}>
                <Input placeholder="https://..." />
              </Form.Item>
              <Form.Item name="status" label={t('pages.modelLibrary.status')} initialValue={1}>
                <Select
                  options={[
                    { value: 1, label: t('pages.modelLibrary.stateOn') },
                    { value: 0, label: t('pages.modelLibrary.stateOff') },
                  ]}
                />
              </Form.Item>
            </Form>
          )}
        </div>
      )}

      {step === 1 && (
        <div className="space-y-4">
          <TypographyMuted>{t('pages.modelLibrary.wizardUpstreamIntro')}</TypographyMuted>
          <Radio.Group
            value={upstreamMode}
            onChange={(e) => setUpstreamMode(e.target.value)}
            optionType="button"
            buttonStyle="solid"
          >
            <Radio.Button value="new">{t('pages.modelLibrary.wizardModeNew')}</Radio.Button>
            <Radio.Button value="existing">{t('pages.modelLibrary.wizardModeExisting')}</Radio.Button>
          </Radio.Group>
          {upstreamMode === 'existing' ? (
            <Form form={upForm} layout="vertical">
              <Form.Item name="existing_upstream_id" label={t('pages.modelLibrary.upstreamModal')} rules={[{ required: true }]}>
                <Select
                  showSearch
                  optionFilterProp="label"
                  options={upstreamOptionsFiltered}
                  placeholder={t('pages.modelLibrary.wizardPickUpstream')}
                />
              </Form.Item>
            </Form>
          ) : (
            <Form form={upForm} layout="vertical">
              <Form.Item name="upstream_name" label={t('pages.modelLibrary.upstreamName')} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="base_url" label={t('pages.modelLibrary.baseUrl')} rules={[{ required: true }]}>
                <Input placeholder="https://api.openai.com" />
              </Form.Item>
              <Form.Item name="api_key" label={t('pages.modelLibrary.apiKey')}>
                <Input.Password placeholder={t('pages.modelLibrary.apiKeyHint')} autoComplete="off" />
              </Form.Item>
              <Form.Item name="timeout" label={t('pages.modelLibrary.timeoutSec')} initialValue={30}>
                <InputNumber min={1} className="w-full" />
              </Form.Item>
              <Form.Item name="max_concurrent" label={t('pages.modelLibrary.maxConcurrent')} initialValue={100}>
                <InputNumber min={1} className="w-full" />
              </Form.Item>
              <Form.Item name="status" label={t('pages.modelLibrary.status')} initialValue={1}>
                <Select
                  options={[
                    { value: 1, label: t('pages.modelLibrary.stateOn') },
                    { value: 0, label: t('pages.modelLibrary.stateOff') },
                  ]}
                />
              </Form.Item>
            </Form>
          )}
        </div>
      )}

      {step === 2 && (
        <div className="space-y-4">
          <Radio.Group
            value={baseMode}
            onChange={(e) => setBaseMode(e.target.value)}
            optionType="button"
            buttonStyle="solid"
          >
            <Radio.Button value="new">{t('pages.modelLibrary.wizardModeNew')}</Radio.Button>
            <Radio.Button value="existing">{t('pages.modelLibrary.wizardModeExisting')}</Radio.Button>
          </Radio.Group>
          {baseMode === 'existing' ? (
            <Form form={baseForm} layout="vertical">
              <Form.Item name="existing_base_id" label={t('pages.modelLibrary.baseModal')} rules={[{ required: true }]}>
                <Select showSearch optionFilterProp="label" options={baseOptions} placeholder={t('pages.modelLibrary.wizardPickBase')} />
              </Form.Item>
            </Form>
          ) : (
            <Form form={baseForm} layout="vertical">
              <Form.Item name="model_name" label={t('pages.modelLibrary.modelNameCol')} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="model_code" label={t('pages.modelLibrary.modelCode')} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="model_type" label={t('pages.modelLibrary.modelType')} rules={[{ required: true }]} initialValue={1}>
                <Select
                  options={[
                    { value: 1, label: t('pages.modelLibrary.mtchat') },
                    { value: 2, label: t('pages.modelLibrary.mtembedding') },
                    { value: 3, label: t('pages.modelLibrary.mtimage') },
                    { value: 4, label: t('pages.modelLibrary.mtaudio') },
                  ]}
                />
              </Form.Item>
              <Form.Item name="capability" label={t('pages.modelLibrary.capabilityJson')}>
                <Input.TextArea rows={3} placeholder="{}" />
              </Form.Item>
              <Form.Item name="sort" label={t('pages.modelLibrary.sort')} initialValue={0}>
                <InputNumber className="w-full" />
              </Form.Item>
              <Form.Item name="status" label={t('pages.modelLibrary.status')} initialValue={1}>
                <Select
                  options={[
                    { value: 1, label: t('pages.modelLibrary.stateOn') },
                    { value: 0, label: t('pages.modelLibrary.stateOff') },
                  ]}
                />
              </Form.Item>
            </Form>
          )}
        </div>
      )}

      {step === 3 && (
        <Form form={instForm} layout="vertical">
          <Form.Item name="instance_name" label={t('pages.modelLibrary.instanceName')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="provider_model_code" label={t('pages.modelLibrary.providerModel')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="weight" label={t('pages.modelLibrary.weight')} initialValue={10}>
            <InputNumber min={1} className="w-full" />
          </Form.Item>
          <Form.Item name="priority" label={t('pages.modelLibrary.priority')} initialValue={1}>
            <Select
              options={[
                { value: 1, label: t('pages.modelLibrary.prioHigh') },
                { value: 2, label: t('pages.modelLibrary.prioMid') },
                { value: 3, label: t('pages.modelLibrary.prioLow') },
              ]}
            />
          </Form.Item>
          <Form.Item name="is_official" label={t('pages.modelLibrary.official')} initialValue={0}>
            <Select
              options={[
                { value: 1, label: t('pages.modelLibrary.yes') },
                { value: 0, label: t('pages.modelLibrary.no') },
              ]}
            />
          </Form.Item>
          <Form.Item name="status" label={t('pages.modelLibrary.status')} initialValue={1}>
            <Select
              options={[
                { value: 1, label: t('pages.modelLibrary.stateOn') },
                { value: 0, label: t('pages.modelLibrary.stateOff') },
              ]}
            />
          </Form.Item>
        </Form>
      )}

      <div className="mt-6 flex justify-end gap-2">
        <Button onClick={step === 0 ? onClose : goPrev}>{step === 0 ? t('common.cancel') : t('pages.modelLibrary.wizardPrev')}</Button>
        {step < 3 ? (
          <Button type="primary" loading={loading} onClick={() => void goNext()}>
            {t('pages.modelLibrary.wizardNext')}
          </Button>
        ) : (
          <Button type="primary" loading={loading} onClick={() => void finish()}>
            {t('pages.modelLibrary.wizardFinish')}
          </Button>
        )}
      </div>
    </Modal>
  )
}

function TypographyMuted({ children }: { children: ReactNode }) {
  return <p className="text-sm text-neutral-500 dark:text-neutral-400">{children}</p>
}
