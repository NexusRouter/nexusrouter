import {
  App,
  Avatar,
  Button,
  Collapse,
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
import { SelectWithCreate } from '../components/SelectWithCreate'
import { TermHint } from '../components/TermHint'
import { VendorLogoField } from '../components/VendorLogoField'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'
import { confirmDestructive } from '../utils/confirmDestructive'
import { resolveVendorLogoUrl } from '../utils/vendorLogoUrl'
import { ModelLibraryWizard } from './ModelLibraryWizard'

/** 本地 public/vendor-logos/<vendor_code>.svg（与网关预置厂商一致；DB logo 可覆盖） */
const VENDOR_LOGO_FALLBACK: Record<string, string> = {
  openai: '/vendor-logos/openai.svg',
  anthropic: '/vendor-logos/anthropic.svg',
  google_gemini: '/vendor-logos/google_gemini.svg',
  azure_openai: '/vendor-logos/azure_openai.svg',
  baidu: '/vendor-logos/baidu.svg',
  zhipu: '/vendor-logos/zhipu.svg',
  aliyun_dashscope: '/vendor-logos/aliyun_dashscope.svg',
  moonshot: '/vendor-logos/moonshot.svg',
  baichuan: '/vendor-logos/baichuan.svg',
  minimax: '/vendor-logos/minimax.svg',
  mistral: '/vendor-logos/mistral.svg',
  groq: '/vendor-logos/groq.svg',
  deepseek: '/vendor-logos/deepseek.svg',
  cohere: '/vendor-logos/cohere.svg',
  xai: '/vendor-logos/xai.svg',
  together: '/vendor-logos/together.svg',
  cloudflare: '/vendor-logos/cloudflare.svg',
  doubao: '/vendor-logos/doubao.svg',
  novita: '/vendor-logos/novita.svg',
  replicate: '/vendor-logos/replicate.svg',
  hunyuan: '/vendor-logos/hunyuan.svg',
  google: '/vendor-logos/google_gemini.svg',
  azure: '/vendor-logos/azure_openai.svg',
  oneapi: '/vendor-logos/openai.svg',
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
  const code = v.vendor_code?.toLowerCase?.() ?? ''
  const src =
    resolveVendorLogoUrl(v.logo ? String(v.logo) : '') ||
    resolveVendorLogoUrl(VENDOR_LOGO_FALLBACK[code])
  return <Avatar src={src} size={28} style={{ flexShrink: 0 }} />
}

function truncateUrl(s: string, max = 48) {
  if (s.length <= max) return s
  return `${s.slice(0, max)}…`
}

export default function ModelLibraryPage() {
  const { message, modal: antdModal } = App.useApp()
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

  const baseById = useMemo(() => {
    const m = new Map<number, Base>()
    for (const b of basesQ.data ?? []) m.set(b.id, b)
    return m
  }, [basesQ.data])

  const upById = useMemo(() => {
    const m = new Map<number, Upstream>()
    for (const u of upsQ.data ?? []) m.set(u.id, u)
    return m
  }, [upsQ.data])

  const [advTab, setAdvTab] = useState('vendors')
  const [wizardOpen, setWizardOpen] = useState(false)
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
        render: (_: unknown, r: Vendor) =>
          readOnly ? (
            '—'
          ) : (
            <Space>
              <Button type="link" size="small" onClick={() => setModal({ kind: 'vendor-edit', row: r })}>
                {t('common.edit')}
              </Button>
              <Button
                type="link"
                size="small"
                danger
                onClick={() =>
                  confirmDestructive(antdModal, t, {
                    title: t('pages.modelLibrary.deleteVendorTitle'),
                    resourceName: `${r.vendor_name} (${r.vendor_code})`,
                    onOk: () => delVendor.mutateAsync(r.id),
                  })
                }
              >
                {t('common.delete')}
              </Button>
            </Space>
          ),
      },
    ],
    [t, readOnly, delVendor, antdModal],
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
        render: (_: unknown, r: Base) =>
          readOnly ? (
            '—'
          ) : (
            <Space>
              <Button type="link" size="small" onClick={() => setModal({ kind: 'base-edit', row: r })}>
                {t('common.edit')}
              </Button>
              <Button
                type="link"
                size="small"
                danger
                onClick={() =>
                  confirmDestructive(antdModal, t, {
                    title: t('pages.modelLibrary.deleteBaseTitle'),
                    resourceName: `${r.model_name} (${r.model_code})`,
                    onOk: () => delBase.mutateAsync(r.id),
                  })
                }
              >
                {t('common.delete')}
              </Button>
            </Space>
          ),
      },
    ],
    [t, readOnly, delBase, antdModal],
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
        render: (_: unknown, r: Upstream) =>
          readOnly ? (
            '—'
          ) : (
            <Space>
              <Button type="link" size="small" onClick={() => setModal({ kind: 'upstream-edit', row: r })}>
                {t('common.edit')}
              </Button>
              <Button
                type="link"
                size="small"
                danger
                onClick={() =>
                  confirmDestructive(antdModal, t, {
                    title: t('pages.modelLibrary.deleteUpstreamTitle'),
                    resourceName: `${r.upstream_name} (#${r.id})`,
                    onOk: () => delUp.mutateAsync(r.id),
                  })
                }
              >
                {t('common.delete')}
              </Button>
            </Space>
          ),
      },
    ],
    [t, readOnly, delUp, vendorById, antdModal],
  )

  const instColsMain = useMemo(
    () => [
      { title: t('pages.modelLibrary.instanceName'), dataIndex: 'instance_name', key: 'instance_name', width: 160 },
      {
        title: (
          <span>
            {t('pages.modelLibrary.modelCode')}
            <TermHint glossaryKey="pages.modelLibrary.glossary.modelCode" />
          </span>
        ),
        key: 'model_code',
        width: 180,
        render: (_: unknown, r: Instance) => {
          const b = baseById.get(r.base_model_id)
          return b?.model_code ?? r.base_model_id
        },
      },
      {
        title: (
          <span>
            {t('pages.modelLibrary.colLogicalModel')}
            <TermHint glossaryKey="pages.modelLibrary.glossary.base" />
          </span>
        ),
        key: 'logical',
        render: (_: unknown, r: Instance) => {
          const b = baseById.get(r.base_model_id)
          return b ? `${b.model_name}` : r.base_model_id
        },
      },
      {
        title: (
          <span>
            {t('pages.modelLibrary.colVendorCol')}
            <TermHint glossaryKey="pages.modelLibrary.glossary.vendor" />
          </span>
        ),
        key: 'vendor',
        render: (_: unknown, r: Instance) => {
          const v = vendorById.get(r.vendor_id)
          return v ? (
            <Space>
              {vendorAvatar(v)}
              <span>{v.vendor_name}</span>
            </Space>
          ) : (
            r.vendor_id
          )
        },
      },
      {
        title: (
          <span>
            {t('pages.modelLibrary.colUpstreamSummary')}
            <TermHint glossaryKey="pages.modelLibrary.glossary.upstream" />
          </span>
        ),
        key: 'up',
        ellipsis: true,
        render: (_: unknown, r: Instance) => {
          const u = upById.get(r.upstream_id)
          return u ? `${u.upstream_name} — ${truncateUrl(u.base_url)}` : r.upstream_id
        },
      },
      {
        title: (
          <span>
            {t('pages.modelLibrary.providerModel')}
            <TermHint glossaryKey="pages.modelLibrary.glossary.providerModel" />
          </span>
        ),
        dataIndex: 'provider_model_code',
        key: 'provider_model_code',
        width: 160,
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
        fixed: 'right' as const,
        render: (_: unknown, r: Instance) =>
          readOnly ? (
            '—'
          ) : (
            <Space>
              <Button type="link" size="small" onClick={() => setModal({ kind: 'instance-edit', row: r })}>
                {t('common.edit')}
              </Button>
              <Button
                type="link"
                size="small"
                danger
                onClick={() =>
                  confirmDestructive(antdModal, t, {
                    title: t('pages.modelLibrary.deleteInstanceTitle'),
                    resourceName: `${r.instance_name} (#${r.id})`,
                    onOk: () => delInst.mutateAsync(r.id),
                  })
                }
              >
                {t('common.delete')}
              </Button>
            </Space>
          ),
      },
    ],
    [t, readOnly, delInst, baseById, vendorById, upById, antdModal],
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
        render: (_: unknown, r: Instance) =>
          readOnly ? (
            '—'
          ) : (
            <Space>
              <Button type="link" size="small" onClick={() => setModal({ kind: 'instance-edit', row: r })}>
                {t('common.edit')}
              </Button>
              <Button
                type="link"
                size="small"
                danger
                onClick={() =>
                  confirmDestructive(antdModal, t, {
                    title: t('pages.modelLibrary.deleteInstanceTitle'),
                    resourceName: `${r.instance_name} (#${r.id})`,
                    onOk: () => delInst.mutateAsync(r.id),
                  })
                }
              >
                {t('common.delete')}
              </Button>
            </Space>
          ),
      },
    ],
    [t, readOnly, delInst, antdModal],
  )

  const emptyMain = useMemo(
    () =>
      readOnly ? (
        <Typography.Text type="secondary">{t('pages.modelLibrary.emptyInstancesTitle')}</Typography.Text>
      ) : (
        <div className="py-8 text-center">
          <Typography.Title level={5} className="!mb-2">
            {t('pages.modelLibrary.emptyInstancesTitle')}
          </Typography.Title>
          <Typography.Paragraph type="secondary" className="!mb-4 max-w-md mx-auto">
            {t('pages.modelLibrary.emptyInstancesHint')}
          </Typography.Paragraph>
          <Button type="primary" onClick={() => setWizardOpen(true)}>
            {t('pages.modelLibrary.primaryConfigure')}
          </Button>
        </div>
      ),
    [t, readOnly],
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
            {t('pages.modelLibrary.hintUser')}
          </Typography.Paragraph>
        </div>
        <Space wrap>
          {!readOnly ? (
            <Button type="primary" onClick={() => setWizardOpen(true)}>
              {t('pages.modelLibrary.primaryConfigure')}
            </Button>
          ) : null}
          {!readOnly ? (
            <Button onClick={() => setSyncOpen(true)}>{t('pages.modelLibrary.syncFromUpstream')}</Button>
          ) : null}
        </Space>
      </div>

      <Collapse
        size="small"
        items={[
          {
            key: 'tech',
            label: t('pages.modelLibrary.technicalDetails'),
            children: (
              <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
                {t('pages.modelLibrary.hintAgg')}
              </Typography.Paragraph>
            ),
          },
        ]}
      />

      <div>
        <Typography.Title level={5} className="!mb-3">
          {t('pages.modelLibrary.mainListTitle')}
        </Typography.Title>
        {!readOnly ? (
          <div className="mb-3 flex flex-wrap gap-2">
            <Button type="primary" onClick={() => setModal({ kind: 'instance-new' })}>
              {t('pages.modelLibrary.addInstance')}
            </Button>
          </div>
        ) : null}
        <Table
          rowKey="id"
          loading={instQ.isLoading}
          dataSource={instQ.data}
          columns={instColsMain}
          pagination={false}
          scroll={{ x: true }}
          locale={{ emptyText: emptyMain }}
        />
      </div>

      <Collapse
        items={[
          {
            key: 'adv',
            label: t('pages.modelLibrary.advancedMaintenance'),
            children: (
              <Tabs
                activeKey={advTab}
                onChange={setAdvTab}
                items={[
                  {
                    key: 'vendors',
                    label: t('pages.modelLibrary.tabVendors'),
                    children: (
                      <div className="space-y-3">
                        {!readOnly ? (
                          <Button type="primary" onClick={() => setModal({ kind: 'vendor-new' })}>
                            {t('pages.modelLibrary.addVendor')}
                          </Button>
                        ) : null}
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
                        {!readOnly ? (
                          <Button type="primary" onClick={() => setModal({ kind: 'base-new' })}>
                            {t('pages.modelLibrary.addBase')}
                          </Button>
                        ) : null}
                        <Table
                          rowKey="id"
                          loading={basesQ.isLoading}
                          dataSource={basesQ.data}
                          columns={baseCols}
                          pagination={false}
                          scroll={{ x: true }}
                        />
                      </div>
                    ),
                  },
                  {
                    key: 'upstreams',
                    label: t('pages.modelLibrary.tabUpstreams'),
                    children: (
                      <div className="space-y-3">
                        {!readOnly ? (
                          <Button type="primary" onClick={() => setModal({ kind: 'upstream-new' })}>
                            {t('pages.modelLibrary.addUpstream')}
                          </Button>
                        ) : null}
                        <Table
                          rowKey="id"
                          loading={upsQ.isLoading}
                          dataSource={upsQ.data}
                          columns={upCols}
                          pagination={false}
                          scroll={{ x: true }}
                        />
                      </div>
                    ),
                  },
                  {
                    key: 'instances',
                    label: t('pages.modelLibrary.tabInstances'),
                    children: (
                      <div className="space-y-3">
                        {!readOnly ? (
                          <Button type="primary" onClick={() => setModal({ kind: 'instance-new' })}>
                            {t('pages.modelLibrary.addInstance')}
                          </Button>
                        ) : null}
                        <Table
                          rowKey="id"
                          loading={instQ.isLoading}
                          dataSource={instQ.data}
                          columns={instCols}
                          pagination={false}
                          scroll={{ x: true }}
                        />
                      </div>
                    ),
                  },
                ]}
              />
            ),
          },
        ]}
      />

      <ModelLibraryWizard open={wizardOpen} onClose={() => setWizardOpen(false)} />

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

type QuickKind = 'vendor' | 'base' | 'upstream'

function NestedQuickCreateModal({
  open,
  kind,
  onClose,
  onCreated,
  vendors,
  defaultVendorId,
  readOnly,
}: {
  open: boolean
  kind: QuickKind | null
  onClose: () => void
  onCreated: (row: { id: number }) => void
  vendors: Vendor[]
  defaultVendorId?: number
  readOnly: boolean
}) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const qc = useQueryClient()
  const [form] = Form.useForm()

  useEffect(() => {
    if (!open || !kind) return
    form.resetFields()
    if (kind === 'vendor') form.setFieldsValue({ vendor_type: 1, status: 1 })
    if (kind === 'base') form.setFieldsValue({ model_type: 1, sort: 0, status: 1 })
    if (kind === 'upstream')
      form.setFieldsValue({ timeout: 30, max_concurrent: 100, status: 1, vendor_id: defaultVendorId })
  }, [open, kind, defaultVendorId, form])

  const save = useMutation({
    mutationFn: async () => {
      const v = await form.validateFields()
      if (kind === 'vendor') {
        const { data } = await api.post<Vendor>('/api/admin/v1/model-library/vendors', v)
        return data
      }
      if (kind === 'base') {
        const { data } = await api.post<Base>('/api/admin/v1/model-library/bases', v)
        return data
      }
      if (kind === 'upstream') {
        if (v.vendor_id == null) {
          throw new Error('vendor')
        }
        const { data } = await api.post<Upstream>('/api/admin/v1/model-library/upstreams', v)
        return data
      }
      throw new Error('kind')
    },
    onSuccess: (data) => {
      message.success(t('common.saved'))
      qc.invalidateQueries({ queryKey: ['ml-vendors'] })
      qc.invalidateQueries({ queryKey: ['ml-bases'] })
      qc.invalidateQueries({ queryKey: ['ml-upstreams'] })
      qc.invalidateQueries({ queryKey: ['ml-instances'] })
      onCreated(data)
      onClose()
    },
    onError: (err: unknown) => {
      if (err instanceof Error && err.message === 'vendor') {
        message.warning(t('pages.modelLibrary.needVendorBeforeUpstream'))
        return
      }
      message.error(t('common.saveFailed'))
    },
  })

  if (!kind) return null

  const title =
    kind === 'vendor'
      ? t('pages.modelLibrary.quickCreateTitleVendor')
      : kind === 'base'
        ? t('pages.modelLibrary.quickCreateTitleBase')
        : t('pages.modelLibrary.quickCreateTitleUpstream')

  return (
    <Modal
      title={title}
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnClose
      width={480}
    >
      <Form form={form} layout="vertical" disabled={readOnly}>
        {kind === 'vendor' && (
          <>
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
            <VendorLogoField disabled={readOnly} />
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
        {kind === 'base' && (
          <>
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
          </>
        )}
        {kind === 'upstream' && (
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
      </Form>
      <div className="mt-4 flex justify-end">
        <Button type="primary" disabled={readOnly} loading={save.isPending} onClick={() => save.mutate()}>
          {t('common.save')}
        </Button>
      </div>
    </Modal>
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
  const [quickKind, setQuickKind] = useState<QuickKind | null>(null)
  const vendorIdWatched = Form.useWatch('vendor_id', form)

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

  const baseOptions = bases.map((b) => ({
    value: b.id,
    label: `${b.model_code} — ${b.model_name}`,
  }))
  const vendorOptions = vendors.map((v) => ({
    value: v.id,
    label: `${v.vendor_code} — ${v.vendor_name}`,
  }))
  const upstreamOptions = upstreams.map((u) => ({
    value: u.id,
    label: `${u.id} — ${u.upstream_name}`,
  }))

  const instanceMode = modal?.kind === 'instance-new' || modal?.kind === 'instance-edit'

  return (
    <>
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
              <VendorLogoField disabled={readOnly} />
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
          {instanceMode && (
            <>
              <Form.Item
                name="base_model_id"
                label={
                  <span>
                    {t('pages.modelLibrary.baseModelId')}
                    <TermHint glossaryKey="pages.modelLibrary.glossary.base" />
                  </span>
                }
                rules={[{ required: true }]}
              >
                <SelectWithCreate
                  showSearch
                  optionFilterProp="label"
                  readOnly={readOnly}
                  placeholder={t('common.select')}
                  options={baseOptions}
                  createLabel={t('pages.modelLibrary.createBaseInline')}
                  onRequestCreate={() => setQuickKind('base')}
                />
              </Form.Item>
              <Form.Item
                name="vendor_id"
                label={
                  <span>
                    {t('pages.modelLibrary.vendor')}
                    <TermHint glossaryKey="pages.modelLibrary.glossary.vendor" />
                  </span>
                }
                rules={[{ required: true }]}
              >
                <SelectWithCreate
                  showSearch
                  optionFilterProp="label"
                  readOnly={readOnly}
                  placeholder={t('common.select')}
                  options={vendorOptions}
                  createLabel={t('pages.modelLibrary.createVendorInline')}
                  onRequestCreate={() => setQuickKind('vendor')}
                />
              </Form.Item>
              <Form.Item
                name="upstream_id"
                label={
                  <span>
                    {t('pages.modelLibrary.upstreamRowId')}
                    <TermHint glossaryKey="pages.modelLibrary.glossary.upstream" />
                  </span>
                }
                rules={[{ required: true }]}
              >
                <SelectWithCreate
                  showSearch
                  optionFilterProp="label"
                  readOnly={readOnly}
                  placeholder={t('common.select')}
                  options={upstreamOptions}
                  createLabel={t('pages.modelLibrary.createUpstreamInline')}
                  onRequestCreate={() => {
                    if (vendorIdWatched == null) {
                      message.warning(t('pages.modelLibrary.needVendorBeforeUpstream'))
                      return
                    }
                    setQuickKind('upstream')
                  }}
                />
              </Form.Item>
              <Form.Item name="instance_name" label={t('pages.modelLibrary.instanceName')} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item
                name="provider_model_code"
                label={
                  <span>
                    {t('pages.modelLibrary.providerModel')}
                    <TermHint glossaryKey="pages.modelLibrary.glossary.providerModel" />
                  </span>
                }
                rules={[{ required: true }]}
              >
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

      <NestedQuickCreateModal
        open={quickKind !== null}
        kind={quickKind}
        onClose={() => setQuickKind(null)}
        onCreated={(row) => {
          if (quickKind === 'vendor') form.setFieldsValue({ vendor_id: row.id })
          if (quickKind === 'base') form.setFieldsValue({ base_model_id: row.id })
          if (quickKind === 'upstream') form.setFieldsValue({ upstream_id: row.id })
        }}
        defaultVendorId={vendorIdWatched ?? undefined}
        vendors={vendors}
        readOnly={readOnly}
      />
    </>
  )
}
