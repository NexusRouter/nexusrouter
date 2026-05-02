import { Button, Result } from 'antd'
import type { ReactNode } from 'react'

type Props = {
  title: string
  onRetry?: () => void
  retryLabel: string
  extra?: ReactNode
}

/** 页级加载失败：统一重试入口。 */
export function PageError({ title, onRetry, retryLabel, extra }: Props) {
  return (
    <Result
      status="error"
      title={title}
      extra={
        <div className="flex flex-wrap justify-center gap-2">
          {onRetry ? (
            <Button type="primary" onClick={onRetry}>
              {retryLabel}
            </Button>
          ) : null}
          {extra}
        </div>
      }
    />
  )
}
