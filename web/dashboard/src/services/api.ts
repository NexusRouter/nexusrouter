import axios from 'axios'
import i18n from '../i18n/config'
import { getAppMessage } from '../message/appMessageBridge'

/** 空字符串表示走 Vite 代理到同源网关（开发）或同域部署（生产）。 */
const baseURL = import.meta.env.VITE_API_BASE_URL ?? ''

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

api.interceptors.response.use(
  (res) => res,
  (err) => {
    const st = err?.response?.status
    if (st === 403) {
      void getAppMessage().error(i18n.t('errors.forbidden'))
    }
    return Promise.reject(err)
  },
)
