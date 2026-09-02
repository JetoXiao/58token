import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const readSource = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf-8')

const docsSource = readSource('src/views/public/IntegrationDocsView.vue')
const englishSource = readSource('src/i18n/locales/en.ts')
const chineseSource = readSource('src/i18n/locales/zh.ts')

describe('IntegrationDocsView docs snippets', () => {
  it('documents Codex CLI with Responses config and auth.json', () => {
    expect(docsSource).toContain('baseUrl: gatewayOrigin.value')
    expect(docsSource).toContain('model_provider = "OpenAI"')
    expect(docsSource).toContain('model = "gpt-5.5"')
    expect(docsSource).toContain('review_model = "gpt-5.4"')
    expect(docsSource).toContain('model_reasoning_effort = "medium"')
    expect(docsSource).toContain('wire_api = "responses"')
    expect(docsSource).toContain('requires_openai_auth = true')
    expect(docsSource).toContain('"OPENAI_API_KEY": "sk-your-api-key"')

    expect(docsSource).not.toContain('wire_api = "chat"')
    expect(docsSource).not.toContain('env_key = "USE_AIFORME_API_KEY"')
    expect(docsSource).not.toContain('responses_websockets_v2')
  })

  it('keeps public docs aligned with current Claude model examples', () => {
    expect(docsSource).toContain('claude-opus-4-8')
    expect(docsSource).not.toContain('claude-opus-4-7')
    expect(englishSource).not.toContain('claude-opus-4.7')
    expect(chineseSource).not.toContain('claude-opus-4.7')
  })

  it('describes Codex CLI base URL as the gateway origin in both locales', () => {
    expect(englishSource).toContain('Codex CLI Responses providers use https://58token.vip')
    expect(chineseSource).toContain('Codex CLI 的 Responses provider 填 https://58token.vip')

    expect(englishSource).not.toContain('Codex CLI OpenAI-compatible providers use https://58token.vip/v1')
    expect(chineseSource).not.toContain('Codex CLI 的 OpenAI 兼容 provider 填 https://58token.vip/v1')
  })

  it('keeps the long docs table of contents scroll-aware', () => {
    expect(docsSource).toContain('scrollProgressPercent')
    expect(docsSource).toContain('updateTocPanelPosition')
    expect(docsSource).toContain("window.addEventListener('scroll', handleDocsScroll")
    expect(docsSource).toContain("window.removeEventListener('scroll', handleDocsScroll")
  })
})
