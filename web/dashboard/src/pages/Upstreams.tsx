import {
  App,
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd'
import type { TFunction } from 'i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router'
import { PageEmpty } from '../components/PageEmpty'
import { PageError } from '../components/PageError'
import { api } from '../services/api'
import { confirmDestructive } from '../utils/confirmDestructive'
import { isOperatorRole, useAuthStore } from '../stores/authStore'

type Upstream = { id: string; base_url: string; weight: number }
type Routing = {
  strategy: string
  default_upstream_id: string
  active_upstream_id: string
}

const ROUTING_STRATEGIES = ['round_robin', 'weighted_random'] as const

function formatRoutingStrategy(t: TFunction, strategy: string | undefined): string {
  const s = strategy?.trim() || 'round_robin'
  if (s === 'round_robin') return t('pages.upstreams.strategyRoundRobin')
  if (s === 'weighted_random') return t('pages.upstreams.strategyWeightedRandom')
  return s
}

function upstreamRoutingOptions(ups: Upstream[], currentId: string | undefined) {
  const base = ups
    .filter((u) => u.id)
    .map((u) => ({
      value: u.id,
      label: `${u.id} (${u.base_url})`,
    }))
  const cur = String(currentId ?? '').trim()
  const extra =
    cur && !base.some((o) => o.value === cur) ? [{ value: cur, label: cur }] : []
  return [...base, ...extra]
}

/** 上游列表与固定当前上游；支持写盘持久化。 */
export default function UpstreamsPage() {
  const { message, modal } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()
  const snap = useQuery({
    queryKey: ['gateway-snapshot'],
    queryFn: async () => {
      const { data } = await api.get<{
        upstreams: Upstream[]
        routing: Routing
        config_file_set: boolean
      }>('/api/admin/v1/gateway/snapshot')
      return data
    },
  })

  const persistPut = useMutation({
    mutationFn: async (body: {
      upstreams: Upstream[]
      routing: Routing
      persist: boolean
    }) => {
      await api.put('/api/admin/v1/gateway/config', body)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      qc.invalidateQueries({ queryKey: ['gateway-snapshot'] })
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const activePut = useMutation({
    mutationFn: async (body: { active_upstream_id: string; persist: boolean }) => {
      await api.put('/api/admin/v1/gateway/active-upstream', body)
    },
    onSuccess: () => {
      message.success(t('pages.upstreams.upstreamUpdated'))
      qc.invalidateQueries({ queryKey: ['gateway-snapshot'] })
    },
    onError: () => message.error(t('common.updateFailed')),
  })

  const [open, setOpen] = useState(false)
  const [form] = Form.useForm<{ upstreams: Upstream[]; routing: Routing }>()

  const data = snap.data
  const columns = useMemo(
    () => [
      { title: t('pages.upstreams.colId'), dataIndex: 'id', key: 'id' },
      { title: t('pages.upstreams.colBaseUrl'), dataIndex: 'base_url', key: 'base_url' },
      { title: t('pages.upstreams.colWeight'), dataIndex: 'weight', key: 'weight' },
    ],
    [t],
  )

  return (
    <div className="space-y-4">
      <Typography.Title level={4} className="!mb-0">
        {t('pages.upstreams.title')}
      </Typography.Title>
      <Typography.Paragraph type="secondary">
        {t('pages.upstreams.introBefore')}{' '}
        <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">gateway.yaml</code>{' '}
        {t('pages.upstreams.introAfter')}{' '}
        <code className="rounded bg-slate-100 px-1 dark:bg-slate-800">NEXUSROUTER_GATEWAY_CONFIG_FILE</code>
        {t('pages.upstreams.introEnd')}
      </Typography.Paragraph>
      {snap.isError ? (
        <PageError
          title={t('common.pageError.title')}
          retryLabel={t('common.pageError.retry')}
          onRetry={() => snap.refetch()}
        />
      ) : null}
      {data && (
        <div className="flex flex-wrap gap-2">
          <Tag>
            {t('pages.upstreams.strategy')}: {formatRoutingStrategy(t, data.routing.strategy)}
          </Tag>
          <Tag>
            {t('pages.upstreams.default')}: {data.routing.default_upstream_id || t('common.emDash')}
          </Tag>
          <Tag color="blue">
            {t('pages.upstreams.pinCurrent')}:{' '}
            {data.routing.active_upstream_id || t('pages.upstreams.pinUnset')}
          </Tag>
          <Tag>
            {t('pages.upstreams.configYaml')}:{' '}
            {data.config_file_set ? t('pages.upstreams.configSet') : t('pages.upstreams.configUnset')}
          </Tag>
        </div>
      )}
      <Table
        loading={snap.isLoading}
        rowKey={(r) => r.id}
        dataSource={data?.upstreams ?? []}
        columns={columns}
        pagination={false}
        locale={{
          emptyText: (
            <PageEmpty
              description={t('common.pageEmpty.upstreamsHint')}
              extra={
                <Link className="text-indigo-600 dark:text-indigo-400" to="/model-library">
                  {t('consoleTerms.guideModelLibrary')}
                </Link>
              }
            />
          ),
        }}
      />
      {!readOnly ? (
        <Space wrap>
          <Button
            type="primary"
            onClick={() => {
              if (!data) {
                return
              }
              form.setFieldsValue({
                upstreams: data.upstreams,
                routing: data.routing,
              })
              setOpen(true)
            }}
          >
            {t('pages.upstreams.editSave')}
          </Button>
          <Button
            onClick={() => {
              modal.confirm({
                title: t('pages.upstreams.unpinTitle'),
                content: t('pages.upstreams.unpinContent'),
                onOk: async () => {
                  await activePut.mutateAsync({
                    active_upstream_id: '',
                    persist: true,
                  })
                },
              })
            }}
          >
            {t('pages.upstreams.unpinBtn')}
          </Button>
        </Space>
      ) : null}

      <Modal
        title={t('pages.upstreams.modalTitle')}
        open={open}
        onCancel={() => setOpen(false)}
        width={720}
        footer={null}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={async (v) => {
            await persistPut.mutateAsync({
              upstreams: v.upstreams,
              routing: v.routing,
              persist: true,
            })
            setOpen(false)
          }}
        >
          <Form.List name="upstreams">
            {(fields, { add, remove }) => (
              <div className="space-y-2">
                {fields.map((f) => (
                  <Space key={f.key} align="baseline" className="flex w-full flex-wrap">
                    <Form.Item
                      {...f}
                      label={t('pages.upstreams.fieldId')}
                      name={[f.name, 'id']}
                      rules={[{ required: true }]}
                    >
                      <Input className="min-w-[120px]" />
                    </Form.Item>
                    <Form.Item
                      {...f}
                      label={t('pages.upstreams.fieldBaseUrl')}
                      name={[f.name, 'base_url']}
                      rules={[{ required: true }]}
                    >
                      <Input className="min-w-[220px]" style={{ minWidth: 220 }} />
                    </Form.Item>
                    <Form.Item
                      {...f}
                      label={t('pages.upstreams.fieldWeight')}
                      name={[f.name, 'weight']}
                      initialValue={1}
                    >
                      <InputNumber min={0} />
                    </Form.Item>
                    <Button
                      type="link"
                      danger
                      onClick={() => {
                        const ups = form.getFieldValue('upstreams') as Upstream[] | undefined
                        const row = ups?.[f.name]
                        const id = String(row?.id ?? '').trim() || `#${f.name + 1}`
                        confirmDestructive(modal, t, {
                          title: t('pages.upstreams.removeRowConfirmTitle'),
                          resourceName: id,
                          onOk: async () => {
                            remove(f.name)
                          },
                        })
                      }}
                    >
                      {t('pages.upstreams.removeRow')}
                    </Button>
                  </Space>
                ))}
                <Button type="dashed" onClick={() => add({ id: '', base_url: '', weight: 1 })}>
                  {t('pages.upstreams.addUpstream')}
                </Button>
              </div>
            )}
          </Form.List>

          <Form.Item noStyle shouldUpdate>
            {() => {
              const ups = (form.getFieldValue('upstreams') as Upstream[] | undefined) ?? []
              const strat = form.getFieldValue(['routing', 'strategy']) as string | undefined
              const strategyOptions = [
                { value: 'round_robin', label: t('pages.upstreams.strategyRoundRobin') },
                { value: 'weighted_random', label: t('pages.upstreams.strategyWeightedRandom') },
                ...(strat &&
                !(ROUTING_STRATEGIES as readonly string[]).includes(strat)
                  ? [{ value: strat, label: strat }]
                  : []),
              ]
              const defCur = form.getFieldValue(['routing', 'default_upstream_id']) as string | undefined
              const actCur = form.getFieldValue(['routing', 'active_upstream_id']) as string | undefined
              return (
                <>
                  <Form.Item name={['routing', 'strategy']} label={t('pages.upstreams.fieldStrategy')}>
                    <Select options={strategyOptions} className="max-w-md" />
                  </Form.Item>
                  <Form.Item
                    name={['routing', 'default_upstream_id']}
                    label={t('pages.upstreams.fieldDefaultUpstream')}
                  >
                    <Select
                      allowClear
                      showSearch
                      optionFilterProp="label"
                      placeholder={t('pages.upstreams.placeholderPickUpstream')}
                      options={upstreamRoutingOptions(ups, defCur)}
                      className="max-w-xl"
                    />
                  </Form.Item>
                  <Form.Item
                    name={['routing', 'active_upstream_id']}
                    label={t('pages.upstreams.fieldActiveUpstream')}
                  >
                    <Select
                      allowClear
                      showSearch
                      optionFilterProp="label"
                      placeholder={t('common.optional')}
                      options={upstreamRoutingOptions(ups, actCur)}
                      className="max-w-xl"
                    />
                  </Form.Item>
                </>
              )
            }}
          </Form.Item>

          <Button type="primary" htmlType="submit" loading={persistPut.isPending}>
            {t('pages.upstreams.submitWrite')}
          </Button>
        </Form>
      </Modal>
    </div>
  )
}
