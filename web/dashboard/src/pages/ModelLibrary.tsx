import { App, Button, Drawer, Form, Input, InputNumber, Modal, Space, Switch, Table, Tag, Typography } from 'antd'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Boxes } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'

type CatalogEntry = {
  id: string
  display_name: string
  owned_by: string
  metadata: string
  created_at?: string
  updated_at?: string
}

type Binding = {
  id: number
  catalog_entry_id: string
  upstream_id: string
  enabled: boolean
  priority: number
  actual_model?: string | null
}

/** 模型库：目录项与上游绑定；同步上游 /v1/models 仅拉取列表供参考。 */
export default function ModelLibraryPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()

  const listQ = useQuery({
    queryKey: ['model-library-entries'],
    queryFn: async () => {
      const { data } = await api.get<{ items: CatalogEntry[]; total: number }>(
        '/api/admin/v1/model-library/entries',
        { params: { limit: 500, offset: 0 } },
      )
      return data
    },
  })

  const snap = useQuery({
    queryKey: ['gateway-snapshot'],
    queryFn: async () => {
      const { data } = await api.get<{ upstreams: { id: string; base_url: string }[] }>(
        '/api/admin/v1/gateway/snapshot',
      )
      return data
    },
  })

  const [entryModal, setEntryModal] = useState(false)
  const [editing, setEditing] = useState<CatalogEntry | null>(null)
  const [bindDrawer, setBindDrawer] = useState<CatalogEntry | null>(null)
  const [syncOpen, setSyncOpen] = useState(false)
  const [entryForm] = Form.useForm<{ id: string; display_name: string; owned_by: string; metadata: string }>()
  const [bindForm] = Form.useForm<{ upstream_id: string; enabled: boolean; priority: number; actual_model: string }>()
  const [syncForm] = Form.useForm<{ upstream_id: string; bearer: string }>()

  const createEntry = useMutation({
    mutationFn: async (body: Partial<CatalogEntry> & { id: string }) => {
      await api.post('/api/admin/v1/model-library/entries', body)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      qc.invalidateQueries({ queryKey: ['model-library-entries'] })
      setEntryModal(false)
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const updateEntry = useMutation({
    mutationFn: async (p: { id: string; body: Partial<CatalogEntry> }) => {
      await api.put(`/api/admin/v1/model-library/entries/${encodeURIComponent(p.id)}`, p.body)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      qc.invalidateQueries({ queryKey: ['model-library-entries'] })
      setEntryModal(false)
      setEditing(null)
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const deleteEntry = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/api/admin/v1/model-library/entries/${encodeURIComponent(id)}`)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      qc.invalidateQueries({ queryKey: ['model-library-entries'] })
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const bindingsQ = useQuery({
    queryKey: ['model-library-bindings', bindDrawer?.id],
    enabled: !!bindDrawer?.id,
    queryFn: async () => {
      const { data } = await api.get<{ items: Binding[] }>(
        `/api/admin/v1/model-library/entries/${encodeURIComponent(bindDrawer!.id)}/bindings`,
      )
      return data.items
    },
  })

  const createBinding = useMutation({
    mutationFn: async (p: { catalogId: string; body: Record<string, unknown> }) => {
      await api.post(`/api/admin/v1/model-library/entries/${encodeURIComponent(p.catalogId)}/bindings`, p.body)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      qc.invalidateQueries({ queryKey: ['model-library-bindings', bindDrawer?.id] })
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const patchBinding = useMutation({
    mutationFn: async (p: { bid: number; body: Record<string, unknown> }) => {
      await api.patch(`/api/admin/v1/model-library/bindings/${p.bid}`, p.body)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      qc.invalidateQueries({ queryKey: ['model-library-bindings', bindDrawer?.id] })
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const deleteBinding = useMutation({
    mutationFn: async (bid: number) => {
      await api.delete(`/api/admin/v1/model-library/bindings/${bid}`)
    },
    onSuccess: () => {
      message.success(t('common.saved'))
      qc.invalidateQueries({ queryKey: ['model-library-bindings', bindDrawer?.id] })
    },
    onError: () => message.error(t('common.saveFailed')),
  })

  const syncMut = useMutation({
    mutationFn: async (body: { upstream_id: string; bearer?: string }) => {
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

  const columns = useMemo(
    () => [
      { title: t('pages.modelLibrary.colId'), dataIndex: 'id', key: 'id', width: 180 },
      { title: t('pages.modelLibrary.colDisplay'), dataIndex: 'display_name', key: 'display_name' },
      { title: t('pages.modelLibrary.colOwnedBy'), dataIndex: 'owned_by', key: 'owned_by', width: 120 },
      {
        title: t('common.actions'),
        key: 'actions',
        width: 220,
        render: (_: unknown, row: CatalogEntry) => (
          <Space size="small">
            <Button
              type="link"
              size="small"
              onClick={() => {
                setBindDrawer(row)
              }}
            >
              {t('pages.modelLibrary.bindings')}
            </Button>
            {!readOnly && (
              <>
                <Button
                  type="link"
                  size="small"
                  onClick={() => {
                    setEditing(row)
                    entryForm.setFieldsValue({
                      id: row.id,
                      display_name: row.display_name,
                      owned_by: row.owned_by,
                      metadata: row.metadata,
                    })
                    setEntryModal(true)
                  }}
                >
                  {t('common.edit')}
                </Button>
                <Button
                  type="link"
                  size="small"
                  danger
                  onClick={() => void deleteEntry.mutateAsync(row.id)}
                >
                  {t('common.delete')}
                </Button>
              </>
            )}
          </Space>
        ),
      },
    ],
    [t, readOnly, entryForm, deleteEntry],
  )

  return (
    <div className="space-y-4">
      <Typography.Title level={4} className="!mb-0 flex items-center gap-2">
        <Boxes className="h-6 w-6 text-indigo-500" aria-hidden />
        {t('pages.modelLibrary.title')}
      </Typography.Title>
      <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
        {t('pages.modelLibrary.hint')}
      </Typography.Paragraph>
      <Space wrap>
        {!readOnly && (
          <Button
            type="primary"
            onClick={() => {
              setEditing(null)
              entryForm.resetFields()
              setEntryModal(true)
            }}
          >
            {t('pages.modelLibrary.addEntry')}
          </Button>
        )}
        {!readOnly && (
          <Button onClick={() => setSyncOpen(true)}>{t('pages.modelLibrary.syncFromUpstream')}</Button>
        )}
      </Space>
      <Table<CatalogEntry>
        size="small"
        rowKey="id"
        loading={listQ.isFetching}
        dataSource={listQ.data?.items ?? []}
        columns={columns}
        pagination={false}
      />

      <Modal
        title={editing ? t('pages.modelLibrary.editEntry') : t('pages.modelLibrary.addEntry')}
        open={entryModal}
        onCancel={() => setEntryModal(false)}
        okButtonProps={{ loading: createEntry.isPending || updateEntry.isPending }}
        onOk={() => {
          void entryForm.validateFields().then((v) => {
            if (editing) {
              updateEntry.mutate({
                id: editing.id,
                body: {
                  display_name: v.display_name,
                  owned_by: v.owned_by,
                  metadata: v.metadata ?? '',
                },
              })
            } else {
              createEntry.mutate({
                id: v.id,
                display_name: v.display_name,
                owned_by: v.owned_by,
                metadata: v.metadata ?? '',
              })
            }
          })
        }}
      >
        <Form form={entryForm} layout="vertical">
          {!editing && (
            <Form.Item name="id" label={t('pages.modelLibrary.colId')} rules={[{ required: true }]}>
              <Input />
            </Form.Item>
          )}
          {editing && (
            <Typography.Text type="secondary" className="mb-2 block">
              ID: {editing.id}
            </Typography.Text>
          )}
          <Form.Item name="display_name" label={t('pages.modelLibrary.colDisplay')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="owned_by" label={t('pages.modelLibrary.colOwnedBy')}>
            <Input />
          </Form.Item>
          <Form.Item name="metadata" label={t('pages.modelLibrary.metadata')}>
            <Input.TextArea rows={3} placeholder="{}" />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={t('pages.modelLibrary.bindingsTitle', { id: bindDrawer?.id ?? '' })}
        open={!!bindDrawer}
        onClose={() => setBindDrawer(null)}
        width={560}
      >
        {bindDrawer && (
          <>
            {!readOnly && (
              <Form
                form={bindForm}
                layout="vertical"
                className="mb-4"
                initialValues={{ enabled: true, priority: 0 }}
                onFinish={(v) => {
                  createBinding.mutate({
                    catalogId: bindDrawer.id,
                    body: {
                      upstream_id: v.upstream_id,
                      enabled: v.enabled ?? true,
                      priority: v.priority ?? 0,
                      actual_model: v.actual_model || null,
                    },
                  })
                  bindForm.resetFields()
                }}
              >
                <Form.Item name="upstream_id" label={t('pages.modelLibrary.upstreamId')} rules={[{ required: true }]}>
                  <Input list="upstream-ids" />
                </Form.Item>
                <datalist id="upstream-ids">
                  {(snap.data?.upstreams ?? []).map((u) => (
                    <option key={u.id} value={u.id} />
                  ))}
                </datalist>
                <Form.Item name="enabled" label={t('pages.modelLibrary.enabled')} valuePropName="checked">
                  <Switch />
                </Form.Item>
                <Form.Item name="priority" label={t('pages.modelLibrary.priority')}>
                  <InputNumber className="w-full" />
                </Form.Item>
                <Form.Item name="actual_model" label={t('pages.modelLibrary.actualModel')}>
                  <Input placeholder={t('pages.modelLibrary.actualModelHint')} />
                </Form.Item>
                <Button type="primary" htmlType="submit" loading={createBinding.isPending}>
                  {t('pages.modelLibrary.addBinding')}
                </Button>
              </Form>
            )}
            <Table<Binding>
              size="small"
              rowKey="id"
              loading={bindingsQ.isFetching}
              dataSource={bindingsQ.data ?? []}
              columns={[
                { title: t('pages.modelLibrary.upstreamId'), dataIndex: 'upstream_id', key: 'upstream_id' },
                {
                  title: t('pages.modelLibrary.enabled'),
                  dataIndex: 'enabled',
                  render: (en: boolean, row) =>
                    readOnly ? (
                      <Tag>{en ? t('pages.modelLibrary.stateOn') : t('pages.modelLibrary.stateOff')}</Tag>
                    ) : (
                      <Switch
                        checked={en}
                        onChange={(checked) =>
                          patchBinding.mutate({
                            bid: row.id,
                            body: { enabled: checked, priority: row.priority, actual_model: row.actual_model },
                          })
                        }
                      />
                    ),
                },
                {
                  title: t('common.actions'),
                  key: 'a',
                  render: (_: unknown, row) =>
                    readOnly ? null : (
                      <Button type="link" danger size="small" onClick={() => void deleteBinding.mutateAsync(row.id)}>
                        {t('common.delete')}
                      </Button>
                    ),
                },
              ]}
            />
          </>
        )}
      </Drawer>

      <Modal
        title={t('pages.modelLibrary.syncFromUpstream')}
        open={syncOpen}
        onCancel={() => setSyncOpen(false)}
        onOk={() =>
          void syncForm.validateFields().then((v) =>
            syncMut.mutate({ upstream_id: v.upstream_id, bearer: v.bearer || undefined }),
          )
        }
        confirmLoading={syncMut.isPending}
      >
        <Form form={syncForm} layout="vertical">
          <Form.Item name="upstream_id" label={t('pages.modelLibrary.upstreamId')} rules={[{ required: true }]}>
            <Input list="sync-upstream-ids" />
          </Form.Item>
          <datalist id="sync-upstream-ids">
            {(snap.data?.upstreams ?? []).map((u) => (
              <option key={u.id} value={u.id} />
            ))}
          </datalist>
          <Form.Item name="bearer" label={t('pages.modelLibrary.syncBearer')}>
            <Input.Password placeholder={t('pages.modelLibrary.syncBearerHint')} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
