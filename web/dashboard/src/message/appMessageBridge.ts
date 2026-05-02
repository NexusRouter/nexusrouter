import type { MessageInstance } from 'antd/es/message/interface'

let ref: MessageInstance | null = null

export function setAppMessageApi(api: MessageInstance | null) {
  ref = api
}

/** 供 axios 等非 React 代码调用；App 未挂载前为 no-op。 */
export function getAppMessage(): MessageInstance {
  return ref ?? noopMessage
}

const noopType = () => Promise.resolve(false) as ReturnType<MessageInstance['error']>

const noopMessage: MessageInstance = {
  info: noopType,
  success: noopType,
  error: noopType,
  warning: noopType,
  loading: noopType,
  open: noopType,
  destroy: () => {},
}
