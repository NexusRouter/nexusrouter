import { api } from './api'

export type BootstrapPhase = 'ready' | 'initializing' | 'completed'

export type BootstrapStatus = {
  initialized: boolean
  phase: BootstrapPhase
}

/** 查询全局首次初始化状态（未初始化时无需登录）。 */
export async function fetchBootstrapStatus(): Promise<BootstrapStatus> {
  const { data } = await api.get<BootstrapStatus>('/api/bootstrap/v1/status')
  return data
}

/** 提交首次初始化（匿名，至多成功一次）。 */
export async function completeBootstrap(payload: {
  admin_username: string
  admin_password: string
  site_display_name?: string
}): Promise<void> {
  await api.post('/api/bootstrap/v1/complete', payload)
}

/** 超级管理员将系统恢复为未初始化（需 Bearer）。 */
export async function resetBootstrap(): Promise<void> {
  await api.post('/api/bootstrap/v1/reset')
}
