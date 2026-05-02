import { App, Button, Card, Checkbox, Form, Input, Typography } from 'antd'
import { useLocation, useNavigate } from 'react-router'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { useAuthStore } from '../stores/authStore'

/** 管理员 / 操作员登录：令牌存 localStorage。 */
export default function LoginPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const loc = useLocation()
  const token = useAuthStore((s) => s.token)
  const setSession = useAuthStore((s) => s.setSession)
  const remembered = useAuthStore((s) => s.rememberedUsername)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (token) {
      navigate('/dashboard', { replace: true })
    }
  }, [token, navigate])

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-100 p-4">
      <Card className="w-full max-w-md shadow-md" title={t('login.title')}>
        <Typography.Paragraph type="secondary" className="!mb-4">
          {t('login.hint')}
        </Typography.Paragraph>
        <Form
          layout="vertical"
          initialValues={{
            username: remembered ?? '',
            remember: !!remembered,
          }}
          onFinish={async (v: {
            username: string
            password: string
            remember: boolean
          }) => {
            setLoading(true)
            try {
              const { data } = await api.post<{
                access_token: string
                role?: string
              }>('/api/admin/v1/auth/login', {
                username: v.username,
                password: v.password,
                remember: v.remember,
              })
              setSession(data.access_token, {
                rememberedUsername: v.remember ? v.username : null,
                role: data.role ?? 'admin',
              })
              message.success(t('login.success'))
              const target =
                (loc.state as { from?: string } | null)?.from ?? '/dashboard'
              navigate(target, { replace: true })
            } catch {
              message.error(t('login.fail'))
            } finally {
              setLoading(false)
            }
          }}
        >
          <Form.Item
            name="username"
            label={t('login.username')}
            rules={[{ required: true, message: t('login.needUsername') }]}
          >
            <Input autoComplete="username" />
          </Form.Item>
          <Form.Item
            name="password"
            label={t('login.password')}
            rules={[{ required: true, message: t('login.needPassword') }]}
          >
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          <Form.Item name="remember" valuePropName="checked">
            <Checkbox>{t('login.remember')}</Checkbox>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              {t('login.submit')}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
