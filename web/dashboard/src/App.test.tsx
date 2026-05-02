import { render, screen, waitFor } from '@testing-library/react'
import { I18nextProvider } from 'react-i18next'
import { MemoryRouter } from 'react-router'
import { describe, expect, it, vi } from 'vitest'
import i18n from './i18n/config'
import App from './App'

vi.mock('./services/bootstrap', () => ({
  fetchBootstrapStatus: vi.fn(() =>
    Promise.resolve({ initialized: true, phase: 'completed' as const }),
  ),
}))

describe('App', () => {
  it('登录页渲染标题', async () => {
    render(
      <I18nextProvider i18n={i18n}>
        <MemoryRouter initialEntries={['/login']}>
          <App />
        </MemoryRouter>
      </I18nextProvider>,
    )
    await waitFor(() => {
      expect(screen.getByText('管理员登录')).toBeTruthy()
    })
  })
})
