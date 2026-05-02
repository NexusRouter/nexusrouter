import { Dropdown, Button } from 'antd'
import type { MenuProps } from 'antd'
import { Languages } from 'lucide-react'
import { useTranslation } from 'react-i18next'

/** 顶栏语言图标 + 下拉菜单（与主题切换同一视觉语言）。 */
export default function LanguageSwitcher() {
  const { i18n, t } = useTranslation()
  const v = i18n.language.startsWith('en') ? 'en' : 'zh'

  const menu: MenuProps = {
    selectedKeys: [v],
    items: [
      { key: 'zh', label: t('layout.langZh') },
      { key: 'en', label: t('layout.langEn') },
    ],
    onClick: ({ key }) => {
      void i18n.changeLanguage(key as string)
    },
  }

  return (
    <Dropdown menu={menu} trigger={['click']} placement="bottomRight">
      <Button
        type="text"
        aria-label={t('layout.languageSwitcherAria')}
        className="!flex !h-8 !min-w-8 !items-center !justify-center !rounded-full !border-0 !bg-white/10 !p-1.5 !text-white hover:!bg-white/20"
        icon={<Languages className="h-[18px] w-[18px]" aria-hidden />}
      />
    </Dropdown>
  )
}
