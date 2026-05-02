import { Alert, Typography } from 'antd'
import { useTranslation } from 'react-i18next'

/** 内嵌 Swagger UI（经 Vite 代理到网关同源路径）。 */
export default function SwaggerDebugPage() {
  const { t } = useTranslation()
  return (
    <div className="space-y-4">
      <Typography.Title level={4} className="!mb-0">
        {t('pages.swagger.title')}
      </Typography.Title>
      <Alert type="info" showIcon message={t('pages.swagger.alertMsg')} />
      <Typography.Paragraph type="secondary">
        {t('pages.swagger.hintBefore')}{' '}
        <code className="rounded bg-slate-100 px-1">NEXUSROUTER_ENABLE_SWAGGER_UI</code>{' '}
        {t('pages.swagger.hintMid')}{' '}
        <code className="rounded bg-slate-100 px-1">/swagger</code>
        {t('pages.swagger.hintEnd')}
      </Typography.Paragraph>
      <div className="h-[calc(100vh-220px)] min-h-[480px] overflow-hidden rounded border border-slate-200">
        <iframe
          title={t('pages.swagger.iframeTitle')}
          src="/swagger/index.html"
          className="h-full w-full border-0"
        />
      </div>
    </div>
  )
}
