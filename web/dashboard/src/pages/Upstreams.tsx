import {
  App,
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'

type Upstream = { id: string; base_url: string; weight: number }
type Routing = {
  strategy: string
  default_upstream_id: string
  active_upstream_id: string
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
      { title: 'ID', dataIndex: 'id', key: 'id' },
      { title: 'Base URL', dataIndex: 'base_url', key: 'base_url' },
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
        <code className="rounded bg-slate-100 px-1">gateway.yaml</code>{' '}
        {t('pages.upstreams.introAfter')}{' '}
        <code className="rounded bg-slate-100 px-1">NEXUSROUTER_GATEWAY_CONFIG_FILE</code>
        {t('pages.upstreams.introEnd')}
      </Typography.Paragraph>
      {data && (
        <div className="flex flex-wrap gap-2">
          <Tag>
            {t('pages.upstreams.strategy')}: {data.routing.strategy || 'round_robin'}
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
                  <Space key={f.key} align="baseline" className="flex w-full">
                    <Form.Item
                      {...f}
                      label="ID"
                      name={[f.name, 'id']}
                      rules={[{ required: true }]}
                    >
                      <Input />
                    </Form.Item>
                    <Form.Item
                      {...f}
                      label="base_url"
                      name={[f.name, 'base_url']}
                      rules={[{ required: true }]}
                    >
                      <Input style={{ width: 280 }} />
                    </Form.Item>
                    <Form.Item
                      {...f}
                      label="weight"
                      name={[f.name, 'weight']}
                      initialValue={1}
                    >
                      <InputNumber min={0} />
                    </Form.Item>
                    <Button type="link" danger onClick={() => remove(f.name)}>
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
          <Form.Item label="strategy" name={['routing', 'strategy']}>
            <Input placeholder="round_robin | weighted_random" />
          </Form.Item>
          <Form.Item
            label="default_upstream_id"
            name={['routing', 'default_upstream_id']}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="active_upstream_id（可空）"
            name={['routing', 'active_upstream_id']}
          >
            <Input />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={persistPut.isPending}>
            {t('pages.upstreams.submitWrite')}
          </Button>
        </Form>
      </Modal>
    </div>
  )
}
