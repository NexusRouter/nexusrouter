import { Select } from 'antd'
import { useTranslation } from 'react-i18next'

/** 中英语言切换（写入 localStorage，键 nexus_i18n_lang）。 */
export default function LanguageSwitcher() {
  const { i18n, t } = useTranslation()
  const v = i18n.language.startsWith('en') ? 'en' : 'zh'
  return (
    <Select
      size="small"
      value={v}
      style={{ width: 120 }}
      options={[
        { value: 'zh', label: t('layout.langZh') },
        { value: 'en', label: t('layout.langEn') },
      ]}
      onChange={(lng) => {
        void i18n.changeLanguage(lng)
      }}
    />
  )
}
