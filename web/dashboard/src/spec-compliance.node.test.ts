/**
 * @vitest-environment node
 * 对应 dashboard-frontend spec：源码目录约定（components/pages/... 存在）。
 */
import { existsSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const srcDir = join(dirname(fileURLToPath(import.meta.url)))

describe('spec: 源码目录约定', () => {
  it.each(['components', 'pages', 'stores', 'services', 'utils'])(
    '存在 src/%s',
    (dir) => {
      expect(existsSync(join(srcDir, dir))).toBe(true)
    },
  )
})
