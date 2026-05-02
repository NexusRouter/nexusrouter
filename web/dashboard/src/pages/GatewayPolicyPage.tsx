import { Affix, Alert, Anchor, Card, Typography } from 'antd'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useLocation } from 'react-router'
import CorsSettingsPage from './CorsSettings'
import IpAccessPage from './IpAccess'
import RateLimitRulesPage from './RateLimitRules'

const SECTION_IDS = {
  ip: 'section-ip-access',
  rate: 'section-rate-limits',
  cors: 'section-cors',
} as const

/** 网关策略：场景化单页（IP 访问 / 限流 / 跨域），带锚点导航。 */
export default function GatewayPolicyPage() {
  const { t } = useTranslation()
  const { hash } = useLocation()

  useEffect(() => {
    if (!hash || hash.length < 2) return
    const id = hash.slice(1)
    requestAnimationFrame(() => {
      document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    })
  }, [hash])

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div>
        <Typography.Title level={3} className="!mb-1">
          {t('pages.gatewayPolicy.pageTitle')}
        </Typography.Title>
        <Typography.Paragraph type="secondary" className="!mb-0 text-sm">
          {t('pages.gatewayPolicy.pageSubtitle')}
        </Typography.Paragraph>
      </div>

      <Alert type="info" showIcon className="text-sm" message={t('pages.gatewayPolicy.persistYamlAlert')} />

      <div className="flex flex-col gap-6 lg:flex-row lg:items-start">
        <div className="lg:w-44 lg:shrink-0">
          <Affix offsetTop={88}>
            <Anchor
              affix={false}
              items={[
                {
                  key: SECTION_IDS.ip,
                  href: `#${SECTION_IDS.ip}`,
                  title: t('pages.gatewayPolicy.navIp'),
                },
                {
                  key: SECTION_IDS.rate,
                  href: `#${SECTION_IDS.rate}`,
                  title: t('pages.gatewayPolicy.navRate'),
                },
                {
                  key: SECTION_IDS.cors,
                  href: `#${SECTION_IDS.cors}`,
                  title: t('pages.gatewayPolicy.navCors'),
                },
              ]}
            />
          </Affix>
        </div>

        <div className="min-w-0 flex-1 space-y-6">
          <Card id={SECTION_IDS.ip} title={t('pages.gatewayPolicy.cardIpTitle')}>
            <IpAccessPage embedded />
          </Card>
          <Card id={SECTION_IDS.rate} title={t('pages.gatewayPolicy.cardRateTitle')}>
            <RateLimitRulesPage embedded />
          </Card>
          <Card id={SECTION_IDS.cors} title={t('pages.gatewayPolicy.cardCorsTitle')}>
            <CorsSettingsPage embedded />
          </Card>
        </div>
      </div>
    </div>
  )
}
