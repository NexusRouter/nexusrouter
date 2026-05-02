import type { MessageInstance, MessageType } from 'antd/es/message/interface'

let ref: MessageInstance | null = null

export function setAppMessageApi(api: MessageInstance | null) {
  ref = api
}

function noopMessageType(): MessageType {
  const resolved = Promise.resolve(false)
  const callable = () => {}
  return Object.assign(callable, {
    then: resolved.then.bind(resolved),
  }) as MessageType
}

const noopOpen: MessageInstance['error'] = () => noopMessageType()

const noopMessage: MessageInstance = {
  info: noopOpen,
  success: noopOpen,
  error: noopOpen,
  warning: noopOpen,
  loading: noopOpen,
  open: () => noopMessageType(),
  destroy: () => {},
}

/** 供 axios 等非 React 代码调用；App 未挂载前为 no-op。 */
export function getAppMessage(): MessageInstance {
  return ref ?? noopMessage
}
