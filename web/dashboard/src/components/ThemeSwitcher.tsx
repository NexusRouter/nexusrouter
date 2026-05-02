import { Dropdown, Button } from 'antd'
import type { MenuProps } from 'antd'
import { Monitor, Moon, Sun } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { ThemeMode } from '../stores/themeStore'
import { useThemeStore } from '../stores/themeStore'

function iconForMode(mode: ThemeMode) {
  const cls = 'h-[18px] w-[18px]'
  if (mode === 'light') {
    return <Sun className={cls} aria-hidden />
  }
  if (mode === 'dark') {
    return <Moon className={cls} aria-hidden />
  }
  return <Monitor className={cls} aria-hidden />
}

/** 顶栏图标 + 下拉菜单切换浅色 / 深色 / 跟随系统（与 New API 式站点一致，避免大块 Segmented）。 */
export default function ThemeSwitcher() {
  const { t } = useTranslation()
  const mode = useThemeStore((s) => s.mode)
  const setMode = useThemeStore((s) => s.setMode)

  const menu: MenuProps = {
    selectedKeys: [mode],
    items: [
      {
        key: 'light',
        icon: <Sun className="h-4 w-4" aria-hidden />,
        label: t('layout.themeLight'),
      },
      {
        key: 'dark',
        icon: <Moon className="h-4 w-4" aria-hidden />,
        label: t('layout.themeDark'),
      },
      {
        key: 'system',
        icon: <Monitor className="h-4 w-4" aria-hidden />,
        label: t('layout.themeSystem'),
      },
    ],
    onClick: ({ key }) => setMode(key as ThemeMode),
  }

  return (
    <Dropdown menu={menu} trigger={['click']} placement="bottomRight">
      <Button
        type="text"
        aria-label={t('layout.themeSwitcherAria')}
        className="!flex !h-8 !min-w-8 !items-center !justify-center !rounded-full !border-0 !bg-white/10 !p-1.5 !text-white hover:!bg-white/20"
        icon={iconForMode(mode)}
      />
    </Dropdown>
  )
}
