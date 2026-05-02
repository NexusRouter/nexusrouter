import {
  App,
  Avatar,
  Button,
  Drawer,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Table,
  Tabs,
  Tag,
  Typography,
} from 'antd'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Boxes } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'

/** 与 one-api / Simple Icons 风格一致的常见厂商图标（可被子段 logo 覆盖） */
const VENDOR_LOGO_FALLBACK: Record<string, string> = {
  openai: 'https://cdn.simpleicons.org/openai/412991',
  anthropic: 'https://cdn.simpleicons.org/anthropic/D4A27F',
  zhipu: 'https://cdn.simpleicons.org/baidu/2932E1',
  google: 'https://cdn.simpleicons.org/google/4285F4',
  azure: 'https://cdn.simpleicons.org/microsoftazure/0078D4',
  deepseek: 'https://cdn.simpleicons.org/deepseek/4D6BFE',
  moonshot: 'https://cdn.simpleicons.org/kimi/000000',
  oneapi: 'https://cdn.simpleicons.org/openapiinitiative/6BA539',
}

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

type Instance = {
  id: number
  base_model_id: number
  vendor_id: number
  upstream_id: number
  instance_name: string
  provider_model_code: string
  weight: number
  priority: number
  is_official: number
  status: number
}

function vendorAvatar(v: Pick<Vendor, 'logo' | 'vendor_code'>) {
  const src = (v.logo && String(v.logo).trim()) || VENDOR_LOGO_FALLBACK[v.vendor_code?.toLowerCase?.() ?? ''] || undefined
  return <Avatar src={src} size={28} style={{ flexShrink: 0 }} />
}

export default function ModelLibraryPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()

  const vendorsQ = useQuery({
    queryKey: ['ml-vendors'],
    queryFn: async () => {
      const { data } = await api.get<{ items: Vendor[] }>('/api/admin/v1/model-library/vendors')
      return data.items
    },
  })
  const basesQ = useQuery({
    queryKey: ['ml-bases'],
    queryFn: async () => {
      const { data } = await api.get<{ items: Base[] }>('/api/admin/v1/model-library/bases')
      return data.items
    },
  })
  const upsQ = useQuery({
    queryKey: ['ml-upstreams'],
    queryFn: async () => {
      const { data } = await api.get<{ items: Upstream[] }>('/api/admin/v1/model-library/upstreams')
      return data.items
    },
  })
  const instQ = useQuery({
    queryKey: ['ml-instances'],
    queryFn: async () => {
      const { data } = await api.get<{ items: Instance[] }>('/api/admin/v1/model-library/instances')
      return data.items
    },
  })

  const vendorById = useMemo(() => {
    const m = new Map<number, Vendor>()
    for (const v of vendorsQ.data ?? []) m.set(v.id, v)
    return m
  }, [vendorsQ.data])

  const [tab, setTab] = useState('vendors')
  const [modal, setModal] = useState<{ kind: string; row?: unknown } | null>(null)
  const [syncOpen, setSyncOpen] = useState(false)

  const invalidateAll = () => {
    qc.invalidateQueries({ queryKey: ['ml-vendors'] })
    qc.invalidateQueries({ queryKey: ['ml-bases'] })
    qc.invalidateQueries({ queryKey: ['ml-upstreams'] })
    qc.invalidateQueries({ queryKey: ['ml-instances'] })
  }

  const delVendor = useMutation({
    mutationFn: (id: number) => api.delete(`/api/admin/v1/model-library/vendors/${id}`),
    onSuccess: () => {
      message.success(t('common.saved'))
      invalidateAll()
    },
    onError: () => message.error(t('common.saveFailed')),
  })
  const delBase = useMutation({
    mutationFn: (id: number) => api.delete(`/api/admin/v1/model-library/bases/${id}`),
    onSuccess: () => {
      message.success(t('common.saved'))
      invalidateAll()
    },
    onError: () => message.error(t('common.saveFailed')),
  })
  const delUp = useMutation({
    mutationFn: (id: number) => api.delete(`/api/admin/v1/model-library/upstreams/${id}`),
    onSuccess: () => {
      message.success(t('common.saved'))
      invalidateAll()
    },
    onError: () => message.error(t('common.saveFailed')),
  })
  const delInst = useMutation({
    mutationFn: (id: number) => api.delete(`/api/admin/v1/model-library/instances/${id}`),
    onSuccess: () => {
      message.success(t('common.saved'))
      invalidateAll()
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const syncMut = useMutation({
    mutationFn: async (body: { model_upstream_id: number; bearer?: string }) => {
      const { data } = await api.post<{ model_ids: string[]; count: number }>(
        '/api/admin/v1/model-library/sync',
        body,
      )
      return data
    },
    onSuccess: (data) => {
      message.success(t('pages.modelLibrary.syncOk', { count: data.count }))
      setSyncOpen(false)
    },
    onError: () => message.error(t('pages.modelLibrary.syncFail')),
  })

  const vendorCols = useMemo(
    () => [
      {
        title: t('pages.modelLibrary.vendorLogo'),
        key: 'logo',
        width: 56,
        render: (_: unknown, r: Vendor) => (
          <Space>
            {vendorAvatar(r)}
            <Typography.Text code>{r.vendor_code}</Typography.Text>
          </Space>
        ),
      },
      { title: t('pages.modelLibrary.vendorName'), dataIndex: 'vendor_name', key: 'vendor_name' },
      {
        title: t('pages.modelLibrary.vendorType'),
        dataIndex: 'vendor_type',
        key: 'vendor_type',
        width: 100,
        render: (v: number) =>
          v === 1 ? <Tag color="blue">{t('pages.modelLibrary.typeOfficial')}</Tag> : <Tag>{t('pages.modelLibrary.typeThirdParty')}</Tag>,
      },
      {
        title: t('pages.modelLibrary.status'),
        dataIndex: 'status',
        width: 80,
        render: (s: number) => (s === 1 ? t('pages.modelLibrary.stateOn') : t('pages.modelLibrary.stateOff')),
      },
      {
        title: t('common.actions'),
        key: 'a',
        width: 160,
        render: (_: unknown, r: Vendor) => (
          <Space>
            <Button type="link" size="small" disabled={readOnly} onClick={() => setModal({ kind: 'vendor-edit', row: r })}>
              {t('common.edit')}
            </Button>
            <Button type="link" size="small" danger disabled={readOnly} onClick={() => delVendor.mutate(r.id)}>
              {t('common.delete')}
            </Button>
          </Space>
        ),
      },
    ],
    [t, readOnly, delVendor],
  )

  const baseCols = useMemo(
    () => [
      { title: t('pages.modelLibrary.modelCode'), dataIndex: 'model_code', key: 'model_code', width: 200 },
      { title: t('pages.modelLibrary.modelNameCol'), dataIndex: 'model_name', key: 'model_name' },
      {
        title: t('pages.modelLibrary.modelType'),
        dataIndex: 'model_type',
        width: 100,
        render: (v: number) =>
          (
            {
              1: t('pages.modelLibrary.mtchat'),
              2: t('pages.modelLibrary.mtembedding'),
              3: t('pages.modelLibrary.mtimage'),
              4: t('pages.modelLibrary.mtaudio'),
            } as Record<number, string>
          )[v] ?? v,
      },
      {
        title: t('common.actions'),
        key: 'a',
        width: 160,
        render: (_: unknown, r: Base) => (
          <Space>
            <Button type="link" size="small" disabled={readOnly} onClick={() => setModal({ kind: 'base-edit', row: r })}>
              {t('common.edit')}
            </Button>
            <Button type="link" size="small" danger disabled={readOnly} onClick={() => delBase.mutate(r.id)}>
              {t('common.delete')}
            </Button>
          </Space>
        ),
      },
    ],
    [t, readOnly, delBase],
  )

  const upCols = useMemo(
    () => [
      { title: 'ID', dataIndex: 'id', width: 72 },
      {
        title: t('pages.modelLibrary.vendor'),
        dataIndex: 'vendor_id',
        render: (id: number) => {
          const v = vendorById.get(id)
          return v ? (
            <Space>
              {vendorAvatar(v)}
              {v.vendor_name}
            </Space>
          ) : (
            id
          )
        },
      },
      { title: t('pages.modelLibrary.upstreamName'), dataIndex: 'upstream_name' },
      { title: t('pages.modelLibrary.baseUrl'), dataIndex: 'base_url', ellipsis: true },
      {
        title: t('pages.modelLibrary.apiKeySet'),
        dataIndex: 'api_key_set',
        width: 100,
        render: (x: boolean) => (x ? t('pages.modelLibrary.yes') : t('pages.modelLibrary.no')),
      },
      {
        title: t('common.actions'),
        key: 'a',
        width: 160,
        render: (_: unknown, r: Upstream) => (
          <Space>
            <Button type="link" size="small" disabled={readOnly} onClick={() => setModal({ kind: 'upstream-edit', row: r })}>
              {t('common.edit')}
            </Button>
            <Button type="link" size="small" danger disabled={readOnly} onClick={() => delUp.mutate(r.id)}>
              {t('common.delete')}
            </Button>
          </Space>
        ),
      },
    ],
    [t, readOnly, delUp, vendorById],
  )

  const instCols = useMemo(
    () => [
      { title: 'ID', dataIndex: 'id', width: 72 },
      { title: t('pages.modelLibrary.instanceName'), dataIndex: 'instance_name' },
      { title: t('pages.modelLibrary.providerModel'), dataIndex: 'provider_model_code', width: 180 },
      { title: t('pages.modelLibrary.baseModelId'), dataIndex: 'base_model_id', width: 100 },
      { title: t('pages.modelLibrary.priority'), dataIndex: 'priority', width: 88 },
      { title: t('pages.modelLibrary.weight'), dataIndex: 'weight', width: 72 },
      {
        title: t('pages.modelLibrary.official'),
        dataIndex: 'is_official',
        width: 80,
        render: (v: number) => (v === 1 ? t('pages.modelLibrary.yes') : t('pages.modelLibrary.no')),
      },
      {
        title: t('common.actions'),
        key: 'a',
        width: 160,
        render: (_: unknown, r: Instance) => (
          <Space>
            <Button type="link" size="small" disabled={readOnly} onClick={() => setModal({ kind: 'instance-edit', row: r })}>
              {t('common.edit')}
            </Button>
            <Button type="link" size="small" danger disabled={readOnly} onClick={() => delInst.mutate(r.id)}>
              {t('common.delete')}
            </Button>
          </Space>
        ),
      },
    ],
    [t, readOnly, delInst],
  )

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <Typography.Title level={4} className="!mb-1 flex items-center gap-2">
            <Boxes className="h-5 w-5 opacity-80" />
            {t('pages.modelLibrary.title')}
          </Typography.Title>
          <Typography.Paragraph type="secondary" className="!mb-0 max-w-3xl text-sm">
            {t('pages.modelLibrary.hintAgg')}
          </Typography.Paragraph>
        </div>
        <Space wrap>
          <Button onClick={() => setSyncOpen(true)} disabled={readOnly}>
            {t('pages.modelLibrary.syncFromUpstream')}
          </Button>
        </Space>
      </div>

      <Tabs
        activeKey={tab}
        onChange={setTab}
        items={[
          {
            key: 'vendors',
            label: t('pages.modelLibrary.tabVendors'),
            children: (
              <div className="space-y-3">
                <Button type="primary" disabled={readOnly} onClick={() => setModal({ kind: 'vendor-new' })}>
                  {t('pages.modelLibrary.addVendor')}
                </Button>
                <Table
                  rowKey="id"
                  loading={vendorsQ.isLoading}
                  dataSource={vendorsQ.data}
                  columns={vendorCols}
                  pagination={false}
                  scroll={{ x: true }}
                />
              </div>
            ),
          },
          {
            key: 'bases',
            label: t('pages.modelLibrary.tabBases'),
            children: (
              <div className="space-y-3">
                <Button type="primary" disabled={readOnly} onClick={() => setModal({ kind: 'base-new' })}>
                  {t('pages.modelLibrary.addBase')}
                </Button>
                <Table rowKey="id" loading={basesQ.isLoading} dataSource={basesQ.data} columns={baseCols} pagination={false} scroll={{ x: true }} />
              </div>
            ),
          },
          {
            key: 'upstreams',
            label: t('pages.modelLibrary.tabUpstreams'),
            children: (
              <div className="space-y-3">
                <Button type="primary" disabled={readOnly} onClick={() => setModal({ kind: 'upstream-new' })}>
                  {t('pages.modelLibrary.addUpstream')}
                </Button>
                <Table rowKey="id" loading={upsQ.isLoading} dataSource={upsQ.data} columns={upCols} pagination={false} scroll={{ x: true }} />
              </div>
            ),
          },
          {
            key: 'instances',
            label: t('pages.modelLibrary.tabInstances'),
            children: (
              <div className="space-y-3">
                <Button type="primary" disabled={readOnly} onClick={() => setModal({ kind: 'instance-new' })}>
                  {t('pages.modelLibrary.addInstance')}
                </Button>
                <Table rowKey="id" loading={instQ.isLoading} dataSource={instQ.data} columns={instCols} pagination={false} scroll={{ x: true }} />
              </div>
            ),
          },
        ]}
      />

      <EditModal
        open={modal !== null}
        modal={modal}
        onClose={() => setModal(null)}
        onSaved={() => {
          invalidateAll()
          setModal(null)
        }}
        vendors={vendorsQ.data ?? []}
        bases={basesQ.data ?? []}
        upstreams={upsQ.data ?? []}
        readOnly={readOnly}
      />

      <Modal
        title={t('pages.modelLibrary.syncFromUpstream')}
        open={syncOpen}
        onCancel={() => setSyncOpen(false)}
        footer={null}
        destroyOnClose
      >
        <SyncForm
          upstreams={upsQ.data ?? []}
          readOnly={readOnly}
          onSubmit={(v) => syncMut.mutate(v)}
          loading={syncMut.isPending}
        />
      </Modal>
    </div>
  )
}

function SyncForm({
  upstreams,
  readOnly,
  onSubmit,
  loading,
}: {
  upstreams: Upstream[]
  readOnly: boolean
  onSubmit: (v: { model_upstream_id: number; bearer?: string }) => void
  loading: boolean
}) {
  const { t } = useTranslation()
  const [form] = Form.useForm<{ model_upstream_id: number; bearer?: string }>()
  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={(v) => onSubmit({ model_upstream_id: v.model_upstream_id, bearer: v.bearer })}
    >
      <Form.Item
        name="model_upstream_id"
        label={t('pages.modelLibrary.syncUpstreamId')}
        rules={[{ required: true }]}
      >
        <Select
          showSearch
          optionFilterProp="label"
          disabled={readOnly}
          options={upstreams.map((u) => ({
            value: u.id,
            label: `${u.id} — ${u.upstream_name} (${u.base_url})`,
          }))}
        />
      </Form.Item>
      <Form.Item name="bearer" label={t('pages.modelLibrary.syncBearer')}>
        <Input.Password autoComplete="off" disabled={readOnly} placeholder={t('pages.modelLibrary.syncBearerHintAgg')} />
      </Form.Item>
      <Button type="primary" htmlType="submit" loading={loading} disabled={readOnly}>
        {t('pages.modelLibrary.runSync')}
      </Button>
    </Form>
  )
}

function EditModal({
  open,
  modal,
  onClose,
  onSaved,
  vendors,
  bases,
  upstreams,
  readOnly,
}: {
  open: boolean
  modal: { kind: string; row?: unknown } | null
  onClose: () => void
  onSaved: () => void
  vendors: Vendor[]
  bases: Base[]
  upstreams: Upstream[]
  readOnly: boolean
}) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const [form] = Form.useForm()

  useEffect(() => {
    if (!open || !modal) return
    form.resetFields()
    if (modal.row) {
      form.setFieldsValue(modal.row as Record<string, unknown>)
    }
  }, [open, modal, form])

  const save = useMutation({
    mutationFn: async () => {
      const v = await form.validateFields()
      const k = modal?.kind ?? ''
      if (k === 'vendor-new') {
        await api.post('/api/admin/v1/model-library/vendors', v)
      } else if (k === 'vendor-edit') {
        const row = modal?.row as Vendor
        await api.put(`/api/admin/v1/model-library/vendors/${row.id}`, v)
      } else if (k === 'base-new') {
        await api.post('/api/admin/v1/model-library/bases', v)
      } else if (k === 'base-edit') {
        const row = modal?.row as Base
        await api.put(`/api/admin/v1/model-library/bases/${row.id}`, v)
      } else if (k === 'upstream-new') {
        await api.post('/api/admin/v1/model-library/upstreams', v)
      } else if (k === 'upstream-edit') {
        const row = modal?.row as Upstream
        await api.put(`/api/admin/v1/model-library/upstreams/${row.id}`, v)
      } else if (k === 'instance-new') {
        await api.post('/api/admin/v1/model-library/instances', v)
      } else if (k === 'instance-edit') {
        const row = modal?.row as Instance
        await api.put(`/api/admin/v1/model-library/instances/${row.id}`, v)
      }
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      onSaved()
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const title = useMemo(() => {
    const k = modal?.kind ?? ''
    if (k.startsWith('vendor')) return t('pages.modelLibrary.vendorModal')
    if (k.startsWith('base')) return t('pages.modelLibrary.baseModal')
    if (k.startsWith('upstream')) return t('pages.modelLibrary.upstreamModal')
    if (k.startsWith('instance')) return t('pages.modelLibrary.instanceModal')
    return ''
  }, [modal?.kind, t])

  return (
    <Drawer
      title={title}
      open={open}
      onClose={onClose}
      width={480}
      destroyOnClose
      extra={
        <Button type="primary" disabled={readOnly} loading={save.isPending} onClick={() => save.mutate()}>
          {t('common.save')}
        </Button>
      }
    >
      <Form form={form} layout="vertical" disabled={readOnly}>
        {(modal?.kind === 'vendor-new' || modal?.kind === 'vendor-edit') && (
          <>
            <Form.Item name="vendor_name" label={t('pages.modelLibrary.vendorName')} rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="vendor_type" label={t('pages.modelLibrary.vendorType')} rules={[{ required: true }]}>
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
          </>
        )}
        {(modal?.kind === 'base-new' || modal?.kind === 'base-edit') && (
          <>
            <Form.Item name="model_name" label={t('pages.modelLibrary.modelNameCol')} rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="model_code" label={t('pages.modelLibrary.modelCode')} rules={[{ required: true }]}>
              <Input disabled={modal?.kind === 'base-edit'} />
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
          </>
        )}
        {(modal?.kind === 'upstream-new' || modal?.kind === 'upstream-edit') && (
          <>
            <Form.Item name="vendor_id" label={t('pages.modelLibrary.vendor')} rules={[{ required: true }]}>
              <Select
                options={vendors.map((v) => ({
                  value: v.id,
                  label: `${v.vendor_code} — ${v.vendor_name}`,
                }))}
              />
            </Form.Item>
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
          </>
        )}
        {(modal?.kind === 'instance-new' || modal?.kind === 'instance-edit') && (
          <>
            <Form.Item name="base_model_id" label={t('pages.modelLibrary.baseModelId')} rules={[{ required: true }]}>
              <Select
                options={bases.map((b) => ({
                  value: b.id,
                  label: `${b.model_code} — ${b.model_name}`,
                }))}
              />
            </Form.Item>
            <Form.Item name="vendor_id" label={t('pages.modelLibrary.vendor')} rules={[{ required: true }]}>
              <Select
                options={vendors.map((v) => ({
                  value: v.id,
                  label: `${v.vendor_code} — ${v.vendor_name}`,
                }))}
              />
            </Form.Item>
            <Form.Item name="upstream_id" label={t('pages.modelLibrary.upstreamRowId')} rules={[{ required: true }]}>
              <Select
                options={upstreams.map((u) => ({
                  value: u.id,
                  label: `${u.id} — ${u.upstream_name}`,
                }))}
              />
            </Form.Item>
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
          </>
        )}
      </Form>
    </Drawer>
  )
}
