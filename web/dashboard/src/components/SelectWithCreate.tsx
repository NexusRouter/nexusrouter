import { Button, Select, type SelectProps } from 'antd'
import type { BaseOptionType } from 'antd/es/select'

type Num = number | undefined

export type SelectWithCreateProps = Omit<SelectProps<Num>, 'options'> & {
  options: BaseOptionType[] | { value: number; label: string }[]
  readOnly?: boolean
  createLabel: string
  onRequestCreate: () => void
}

/** 下拉底部「新建…」，点击由父级打开子表单；不改变 Select 的 options 类型契约 */
export function SelectWithCreate({
  readOnly,
  createLabel,
  onRequestCreate,
  dropdownRender,
  disabled,
  ...rest
}: SelectWithCreateProps) {
  const mergedDropdownRender: SelectProps<Num>['dropdownRender'] = (menu) => {
    const inner = typeof dropdownRender === 'function' ? dropdownRender(menu) : menu
    if (readOnly) return <>{inner}</>
    return (
      <div>
        {inner}
        <div className="border-t border-neutral-200 p-1 dark:border-neutral-700">
          <Button
            type="link"
            block
            className="!justify-start"
            disabled={disabled}
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => onRequestCreate()}
          >
            {createLabel}
          </Button>
        </div>
      </div>
    )
  }

  return <Select {...rest} disabled={disabled} dropdownRender={mergedDropdownRender} />
}
