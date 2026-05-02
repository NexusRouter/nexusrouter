import { App } from 'antd'
import { useEffect } from 'react'
import { setAppMessageApi } from '../message/appMessageBridge'

/** 将 App.useApp().message 注册到桥接，供 axios 等非 React 代码使用。 */
export function MessageBridge() {
  const { message } = App.useApp()
  useEffect(() => {
    setAppMessageApi(message)
    return () => setAppMessageApi(null)
  }, [message])
  return null
}
