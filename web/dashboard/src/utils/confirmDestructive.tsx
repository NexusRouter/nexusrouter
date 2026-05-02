/**
 * 危险操作确认强度（与 OpenSpec admin-console-global-ux 对齐）：
 * - L1：可逆或影响面小 — 仍须 Modal 确认，可不强调资源全名（如表单内移除未保存行）。
 * - L2：删除远端已持久化资源 — MUST 展示资源标识 + 不可恢复说明 + okType danger（本封装默认）。
 */
import type { ModalFuncProps } from 'antd'
import type { TFunction } from 'i18next'
import type { ReactNode } from 'react'

export type ConfirmDestructiveOptions = {
  title: string
  /** 展示在正文中的资源名称或标识 */
  resourceName: string
  /** 附加说明（可选） */
  extraContent?: ReactNode
  okText?: string
  onOk: () => Promise<unknown> | unknown
}

export function confirmDestructive(
  modal: { confirm: (props: ModalFuncProps) => unknown },
  t: TFunction,
  opts: ConfirmDestructiveOptions,
) {
  return modal.confirm({
    title: opts.title,
    content: (
      <div className="space-y-2">
        <p>{t('common.destructive.resourceLabel', { name: opts.resourceName })}</p>
        <p className="text-slate-600 dark:text-slate-400">{t('common.destructive.irreversible')}</p>
        {opts.extraContent}
      </div>
    ),
    okText: opts.okText ?? t('common.destructive.confirmOk'),
    okType: 'danger',
    okButtonProps: { danger: true },
    onOk: () => Promise.resolve(opts.onOk()),
  })
}
