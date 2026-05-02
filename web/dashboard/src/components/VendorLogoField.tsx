import type { UploadProps } from 'antd'
import { App, Avatar, Button, Form, Input, Space, Upload } from 'antd'
import { Inbox, Upload as UploadIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { api } from '../services/api'
import { resolveVendorLogoUrl } from '../utils/vendorLogoUrl'

type Props = {
  disabled?: boolean
}

/**
 * 厂商 Logo：支持本地上传（写入 /uploads/...）或手动填写外链/站内路径。
 */
export function VendorLogoField({ disabled }: Props) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const form = Form.useFormInstance()
  const logo = Form.useWatch('logo', form)

  const customRequest: NonNullable<UploadProps['customRequest']> = async (options) => {
    const { file, onError, onSuccess } = options
    const fd = new FormData()
    fd.append('file', file as Blob)
    try {
      const { data } = await api.post<{ path: string }>('/api/admin/v1/model-library/vendor-logo', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      form.setFieldValue('logo', data.path)
      onSuccess?.(data, new XMLHttpRequest())
      message.success(t('pages.modelLibrary.logoUploadOk'))
    } catch (e) {
      onError?.(e as Error)
      message.error(t('pages.modelLibrary.logoUploadFail'))
    }
  }

  return (
    <Space direction="vertical" className="w-full" size="middle">
      <div className="flex flex-wrap items-center gap-3">
        <Avatar src={resolveVendorLogoUrl(logo)} size={48} className="shrink-0 bg-neutral-100">
          {!logo ? <Inbox className="h-5 w-5 text-neutral-400" /> : null}
        </Avatar>
        <Upload disabled={disabled} accept=".svg,.png,.jpg,.jpeg,.gif,.webp" maxCount={1} showUploadList={false} customRequest={customRequest}>
          <Button type="default" size="small" icon={<UploadIcon className="h-4 w-4" />} disabled={disabled}>
            {t('pages.modelLibrary.logoUpload')}
          </Button>
        </Upload>
      </div>
      <p className="text-neutral-500 text-xs m-0">{t('pages.modelLibrary.logoHint')}</p>
      <Form.Item name="logo" label={t('pages.modelLibrary.logoUrlOrPath')} className="mb-0">
        <Input disabled={disabled} placeholder={t('pages.modelLibrary.logoUrlPlaceholder')} allowClear />
      </Form.Item>
    </Space>
  )
}
