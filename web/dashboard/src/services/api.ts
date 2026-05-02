import axios from 'axios'
import i18n from '../i18n/config'
import { getAppMessage } from '../message/appMessageBridge'
import { useAuthStore } from '../stores/authStore'

/** 默认直连本机网关；需走 Vite 代理时可在 `.env` 设 `VITE_API_BASE_URL=`（空串）。生产同域部署可设为空或相对路径。 */
const baseURL = import.meta.env.VITE_API_BASE_URL ?? 'http://127.0.0.1:8080'

export const api = axios.create({
  baseURL,
  timeout: 60_000,
})

api.interceptors.request.use((config) => {
  const raw = localStorage.getItem('nexus_admin_token')
  if (raw) {
    config.headers.Authorization = `Bearer ${raw}`
  }
  return config
})

function loginPath(): string {
  const base = import.meta.env.BASE_URL || '/'
  return `${base.endsWith('/') ? base : `${base}/`}login`
}

function isAdminLoginRequest(config: { method?: string; url?: string } | undefined): boolean {
  if (!config?.url) {
    return false
  }
  const m = (config.method ?? 'get').toLowerCase()
  return m === 'post' && config.url.includes('/api/admin/v1/auth/login')
}

api.interceptors.response.use(
  (res) => res,
  (err) => {
    const st = err?.response?.status
    const cfg = err?.config as { method?: string; url?: string } | undefined
    if (st === 401 && !isAdminLoginRequest(cfg)) {
      useAuthStore.getState().logout()
      window.location.assign(loginPath())
      return Promise.reject(err)
    }
    if (st === 403) {
      void getAppMessage().error(i18n.t('errors.forbidden'))
    }
    return Promise.reject(err)
  },
)
