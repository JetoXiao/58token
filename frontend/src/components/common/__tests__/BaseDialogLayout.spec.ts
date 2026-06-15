import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const readSource = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf-8')

describe('BaseDialog layout styles', () => {
  it('keeps long dialogs inside the viewport with a scrollable body', () => {
    const styleSource = readSource('src/style.css')

    expect(styleSource).toContain('overflow-y-auto p-3 py-4 sm:p-6')
    expect(styleSource).toContain('max-h-[calc(100vh-2rem)] sm:max-h-[calc(100vh-3rem)]')
    expect(styleSource).toContain('flex min-h-0 flex-col overflow-hidden')
    expect(styleSource).toContain('min-h-0 overflow-y-auto')
    expect(styleSource).toContain('flex: 0 1 auto')
    expect(styleSource).toContain('max-height: calc(100dvh - 2rem)')
  })

  it('keeps the API key dialog as a clean single-column settings form', () => {
    const keysViewSource = readSource('src/views/user/KeysView.vue')

    expect(keysViewSource).toContain('width="normal"')
    expect(keysViewSource).toContain('class="space-y-5"')
    expect(keysViewSource).toContain('class="flex items-center justify-between"')
    expect(keysViewSource).not.toContain('md:grid-cols-2')
    expect(keysViewSource).not.toContain('width="form"')
  })
})
