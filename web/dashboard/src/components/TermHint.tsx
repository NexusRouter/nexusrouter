import { Tooltip } from 'antd'
import { HelpCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'

/** Tooltip 术语释义，文案键指向 `pages.modelLibrary.glossary.*` */
export function TermHint({ glossaryKey }: { glossaryKey: string }) {
  const { t } = useTranslation()
  return (
    <Tooltip title={t(glossaryKey)}>
      <HelpCircle
        className="ml-0.5 inline h-3.5 w-3.5 cursor-help align-middle text-neutral-400"
        aria-label={t(glossaryKey)}
      />
    </Tooltip>
  )
}
