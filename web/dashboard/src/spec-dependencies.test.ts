/**
 * 对应 dashboard-frontend spec：核心依赖矩阵可解析（关键包可导入）。
 */
import { describe, expect, it } from 'vitest'

describe('spec: 核心依赖可导入', () => {
  it('antd', async () => {
    const antd = await import('antd')
    // antd 6 导出为 React 组件对象，非裸函数
    expect(antd.Button).toBeDefined()
    expect(typeof antd.Button === 'function' || typeof antd.Button === 'object').toBe(
      true,
    )
  })

  it('@tanstack/react-query', async () => {
    const rq = await import('@tanstack/react-query')
    expect(rq.QueryClient).toBeTypeOf('function')
  })

  it('zustand', async () => {
    const z = await import('zustand')
    expect(z.create).toBeTypeOf('function')
  })

  it('axios', async () => {
    const ax = await import('axios')
    expect(ax.default.request).toBeTypeOf('function')
  })

  it('react-router', async () => {
    const rr = await import('react-router')
    expect(rr.BrowserRouter ?? rr.createBrowserRouter).toBeDefined()
  })

  it('react-hook-form', async () => {
    const rhf = await import('react-hook-form')
    expect(rhf.useForm).toBeTypeOf('function')
  })

  it('zod', async () => {
    const { z } = await import('zod')
    expect(z.string().parse('a')).toBe('a')
  })

  it('dayjs', async () => {
    const dayjs = (await import('dayjs')).default
    expect(dayjs('2020-01-01').year()).toBe(2020)
  })
})
