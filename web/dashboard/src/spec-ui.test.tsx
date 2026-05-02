/**
 * 对应 dashboard-frontend spec：Tailwind 实用类与 antd 集成在根页面可见。
 */
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import App from './App'

describe('spec: UI 基线', () => {
  it('登录页容器包含 Tailwind 布局类', () => {
    const { container } = render(
      <MemoryRouter initialEntries={['/login']}>
        <App />
      </MemoryRouter>,
    )
    const root = container.querySelector('.flex.min-h-screen')
    expect(root).toBeTruthy()
  })

  it('渲染登录主按钮', () => {
    render(
      <MemoryRouter initialEntries={['/login']}>
        <App />
      </MemoryRouter>,
    )
    expect(screen.getByRole('button', { name: /登\s*录/ })).toBeTruthy()
  })
})
