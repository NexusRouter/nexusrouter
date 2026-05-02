/** 浏览器端 IP / CIDR 轻量校验（最终以服务端为准）。 */

const IPV4_OCTET = '(25[0-5]|2[0-4]\\d|1\\d{2}|[1-9]?\\d)'

/** IPv4 或 IPv4 CIDR */
const RE_IPV4_CIDR = new RegExp(
  `^${IPV4_OCTET}\\.${IPV4_OCTET}\\.${IPV4_OCTET}\\.${IPV4_OCTET}(/([0-9]|[12][0-9]|3[0-2]))?$`,
)

/** 宽松 IPv6：含冒号，避免与 IPv4 混淆；完整校验由服务端完成 */
const RE_IPV6_LOOSE = /^[0-9a-fA-F:.]+(\/([0-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8]))?$/

export function isValidIpOrCidrToken(s: string): boolean {
  const t = s.trim()
  if (!t) return false
  if (t.includes(':')) return RE_IPV6_LOOSE.test(t)
  return RE_IPV4_CIDR.test(t)
}

/** 拆分多行/逗号输入后逐条校验 */
export function validateCidrListInput(raw: string): string | true {
  const parts = raw
    .split(/\n|,|;/g)
    .map((x) => x.trim())
    .filter(Boolean)
  for (const p of parts) {
    if (!isValidIpOrCidrToken(p)) return p
  }
  return true
}

const RE_PATH_PREFIX = /^\/[^\s]*$/

export function isValidPathPrefix(s: string): boolean {
  if (s === '') return true
  return RE_PATH_PREFIX.test(s)
}
