import {
  App,
  Button,
  Modal,
  Space,
  Switch,
  Table,
  Typography,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import dayjs from 'dayjs'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { isOperatorRole, useAuthStore } from '../stores/authStore'

type Row = {
  id: string
  masked_secret: string
  disabled: boolean
  expires_at?: string
  created_at?: string
}

/** API Key 列表、创建、禁用与批量操作。 */
export default function ApiKeysPage() {
  const { message, modal } = App.useApp()
  const { t } = useTranslation()
  const readOnly = isOperatorRole(useAuthStore((s) => s.role))
  const qc = useQueryClient()
  const [selected, setSelected] = useState<string[]>([])
  const [newSecretModal, setNewSecretModal] = useState<string | null>(null)

  const list = useQuery({
    queryKey: ['admin-keys'],
    queryFn: async () => {
      const { data } = await api.get<{ items: Row[] }>('/api/admin/v1/keys')
      return data.items
    },
  })

  const createMut = useMutation({
    mutationFn: async () => {
      const { data } = await api.post<{
        id: string
        secret: string
        warning?: string
      }>('/api/admin/v1/keys', {})
      return data
    },
    onSuccess: (data) => {
      message.success(t('pages.apiKeys.createdOk'))
      setNewSecretModal(data.secret)
      qc.invalidateQueries({ queryKey: ['admin-keys'] })
    },
    onError: () => message.error(t('pages.apiKeys.createFail')),
  })

  const patchMut = useMutation({
    mutationFn: async (p: { id: string; disabled: boolean }) => {
      await api.patch(`/api/admin/v1/keys/${encodeURIComponent(p.id)}`, {
        disabled: p.disabled,
      })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-keys'] })
    },
  })

  const delMut = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/api/admin/v1/keys/${encodeURIComponent(id)}`)
    },
    onSuccess: () => {
      message.success(t('common.deleted'))
      qc.invalidateQueries({ queryKey: ['admin-keys'] })
    },
  })

  const batchDisable = useMutation({
    mutationFn: async (ids: string[]) => {
      await api.post('/api/admin/v1/keys/batch-disable', { ids })
    },
    onSuccess: () => {
      message.success(t('pages.apiKeys.batchDisabledOk'))
      setSelected([])
      qc.invalidateQueries({ queryKey: ['admin-keys'] })
    },
  })

  const batchDelete = useMutation({
    mutationFn: async (ids: string[]) => {
      await api.post('/api/admin/v1/keys/batch-delete', { ids })
    },
    onSuccess: () => {
      message.success(t('pages.apiKeys.batchDeletedOk'))
      setSelected([])
      qc.invalidateQueries({ queryKey: ['admin-keys'] })
    },
  })

  const columns: ColumnsType<Row> = useMemo(
    () => [
      { title: t('pages.apiKeys.colId'), dataIndex: 'id', key: 'id' },
      { title: t('pages.apiKeys.colMasked'), dataIndex: 'masked_secret', key: 'ms' },
      {
        title: t('pages.apiKeys.colStatus'),
        dataIndex: 'disabled',
        key: 'dis',
        render: (d: boolean, r) => (
          <Switch
            checked={!d}
            disabled={readOnly}
            onChange={(on) =>
              patchMut.mutate({ id: r.id, disabled: !on })
            }
          />
        ),
      },
      {
        title: t('pages.apiKeys.colExpires'),
        dataIndex: 'expires_at',
        key: 'ex',
        render: (v: string | undefined) =>
          v ? dayjs(v).format('YYYY-MM-DD HH:mm') : t('common.emDash'),
      },
      {
        title: t('pages.apiKeys.colCreated'),
        dataIndex: 'created_at',
        key: 'cr',
        render: (v: string | undefined) =>
          v ? dayjs(v).format('YYYY-MM-DD HH:mm') : t('common.emDash'),
      },
      {
        title: t('pages.apiKeys.colActions'),
        key: 'op',
        render: (_, r) =>
          readOnly ? null : (
            <Button
              type="link"
              danger
              size="small"
              onClick={() => {
                modal.confirm({
                  title: t('pages.apiKeys.deleteTitle'),
                  content: t('pages.apiKeys.deleteContent', { id: r.id }),
                  onOk: () => delMut.mutateAsync(r.id),
                })
              }}
            >
              {t('pages.apiKeys.deleteBtn')}
            </Button>
          ),
      },
    ],
    [t, readOnly, patchMut, delMut, modal],
  )

  return (
    <div className="space-y-4">
      <Typography.Title level={4} className="!mb-0">
        {t('pages.apiKeys.title')}
      </Typography.Title>
      <Typography.Paragraph type="secondary">
        {t('pages.apiKeys.introBefore')}{' '}
        <code className="rounded bg-slate-100 px-1">NEXUSROUTER_GATEWAY_KEYS_FILE</code>
        {t('pages.apiKeys.introAfter')}
      </Typography.Paragraph>
      {!readOnly ? (
        <Space wrap className="mb-2">
          <Button type="primary" onClick={() => createMut.mutate()} loading={createMut.isPending}>
            {t('pages.apiKeys.create')}
          </Button>
          <Button
            disabled={selected.length === 0}
            onClick={() => {
              modal.confirm({
                title: t('pages.apiKeys.batchDisable'),
                onOk: () =>
                  batchDisable.mutate(selected.map(String) as string[]),
              })
            }}
          >
            {t('pages.apiKeys.batchDisable')}
          </Button>
          <Button
            danger
            disabled={selected.length === 0}
            onClick={() => {
              modal.confirm({
                title: t('pages.apiKeys.batchDelete'),
                content: t('pages.apiKeys.batchDeleteContent'),
                onOk: () =>
                  batchDelete.mutate(selected.map(String) as string[]),
              })
            }}
          >
            {t('pages.apiKeys.batchDelete')}
          </Button>
        </Space>
      ) : null}
      <Table<Row>
        rowKey={(r) => r.id}
        loading={list.isLoading}
        dataSource={list.data ?? []}
        columns={columns}
        rowSelection={
          readOnly
            ? undefined
            : {
                selectedRowKeys: selected,
                onChange: (keys) => setSelected(keys.map(String)),
              }
        }
      />
      <Modal
        title={t('pages.apiKeys.modalSecretTitle')}
        open={newSecretModal !== null}
        onOk={() => setNewSecretModal(null)}
        onCancel={() => setNewSecretModal(null)}
        okText={t('pages.apiKeys.modalSecretOk')}
      >
        <Typography.Paragraph copyable className="font-mono break-all">
          {newSecretModal}
        </Typography.Paragraph>
      </Modal>
    </div>
  )
}
