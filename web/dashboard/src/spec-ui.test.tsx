/**
 * 对应 dashboard-frontend spec：Tailwind 实用类与 antd 集成在根页面可见。
 */
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import App from './App'

describe('spec: UI 基线', () => {
  it('根容器包含 Tailwind 布局类', () => {
    const { container } = render(<App />)
    const root = container.querySelector('.flex.min-h-screen')
    expect(root).toBeTruthy()
  })

  it('渲染 Ant Design 主按钮文案', () => {
    render(<App />)
    expect(screen.getByRole('button', { name: /Ant Design 按钮/ })).toBeTruthy()
  })
})
