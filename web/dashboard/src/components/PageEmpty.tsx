import { Empty } from 'antd'
import type { ReactNode } from 'react'

type Props = {
  description: ReactNode
  extra?: ReactNode
  className?: string
}

/** 列表/页级空态：可配合侧栏术语键做跨页引导。 */
export function PageEmpty({ description, extra, className }: Props) {
  return (
    <Empty className={className} description={description}>
      {extra ? <div className="mt-2 flex flex-wrap justify-center gap-2">{extra}</div> : null}
    </Empty>
  )
}
