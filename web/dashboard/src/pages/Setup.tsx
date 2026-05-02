import { App, Button, Card, Form, Input, Typography } from 'antd'
import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import PublicPageShell from '../components/PublicPageShell'
import { completeBootstrap } from '../services/bootstrap'

/** 首次部署向导：创建超级管理员并写入站点显示名。 */
export default function SetupPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)

  return (
    <PublicPageShell>
      <Card className="w-full max-w-md shadow-lg dark:border-slate-700" title={t('setup.title')}>
        <Typography.Paragraph type="secondary" className="!mb-4">
          {t('setup.intro')}
        </Typography.Paragraph>
        <Form
          layout="vertical"
          onFinish={async (v: {
            admin_username: string
            admin_password: string
            site_display_name?: string
          }) => {
            setLoading(true)
            try {
              await completeBootstrap({
                admin_username: v.admin_username,
                admin_password: v.admin_password,
                site_display_name: v.site_display_name?.trim() || undefined,
              })
              message.success(t('setup.success'))
              navigate('/login', { replace: true })
            } catch (e: unknown) {
              const code = (e as { response?: { data?: { code?: string } } })?.response?.data?.code
              if (code === 'BOOTSTRAP_ALREADY_COMPLETED') {
                message.warning(t('setup.alreadyDone'))
                navigate('/login', { replace: true })
              } else if (code === 'BOOTSTRAP_IN_PROGRESS') {
                message.warning(t('setup.inProgress'))
              } else if (code === 'BOOTSTRAP_JWT_MISSING') {
                message.error(t('setup.jwtMissing'))
              } else if (code === 'ADMIN_DISABLED') {
                message.error(t('setup.adminDisabled'))
              } else {
                message.error(t('setup.fail'))
              }
            } finally {
              setLoading(false)
            }
          }}
        >
          <Form.Item
            name="admin_username"
            label={t('setup.adminUsername')}
            rules={[{ required: true, message: t('setup.needAdminUsername') }]}
          >
            <Input autoComplete="username" />
          </Form.Item>
          <Form.Item
            name="admin_password"
            label={t('setup.adminPassword')}
            rules={[
              { required: true, message: t('setup.needAdminPassword') },
              { min: 8, message: t('setup.passwordMin') },
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Form.Item name="site_display_name" label={t('setup.siteName')}>
            <Input maxLength={255} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              {t('setup.submit')}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </PublicPageShell>
  )
}
