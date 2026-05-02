import { api } from '../services/api'

/** 将 model_vendor.logo 转为 <img> / Avatar 可用的绝对 URL */
export function resolveVendorLogoUrl(logo: string | null | undefined): string | undefined {
  const s = logo?.trim()
  if (!s) return undefined
  if (/^https?:\/\//i.test(s)) return s
  if (s.startsWith('/uploads/')) {
    const base = api.defaults.baseURL?.replace(/\/$/, '') ?? ''
    return `${base}${s}`
  }
  return s
}
