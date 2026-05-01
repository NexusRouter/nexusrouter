import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import App from './App'

describe('App', () => {
  it('渲染标题', () => {
    render(<App />)
    expect(screen.getByText('NexusRouter 控制台')).toBeTruthy()
  })
})
