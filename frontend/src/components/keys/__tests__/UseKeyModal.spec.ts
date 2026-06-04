import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

import UseKeyModal from '../UseKeyModal.vue'

describe('UseKeyModal', () => {
  it('renders complete Claude Code env and settings config for Anthropic keys', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-claude',
        baseUrl: 'https://useaifor.me',
        platform: 'anthropic'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const codeBlocks = wrapper.findAll('pre code')

    expect(codeBlocks[0].text()).toContain('export ANTHROPIC_MODEL="claude-opus-4-8"')
    expect(codeBlocks[0].text()).toContain('export ANTHROPIC_DEFAULT_OPUS_MODEL="claude-opus-4-8"')
    expect(codeBlocks[0].text()).toContain('export API_TIMEOUT_MS="3000000"')
    expect(codeBlocks[0].text()).toContain('export CLAUDE_CODE_ATTRIBUTION_HEADER="0"')

    const settings = JSON.parse(codeBlocks[1].text())
    expect(settings).toMatchObject({
      enabledPlugins: {
        'commit-commands@claude-plugins-official': true,
        'context7@claude-plugins-official': true,
        'frontend-design@claude-plugins-official': true,
        'playwright@claude-plugins-official': true,
        'pyright-lsp@claude-plugins-official': true,
        'superpowers@claude-plugins-official': true
      },
      env: {
        ANTHROPIC_AUTH_TOKEN: 'sk-claude',
        ANTHROPIC_BASE_URL: 'https://useaifor.me',
        ANTHROPIC_DEFAULT_HAIKU_MODEL: 'claude-opus-4-8',
        ANTHROPIC_DEFAULT_OPUS_MODEL: 'claude-opus-4-8',
        ANTHROPIC_DEFAULT_SONNET_MODEL: 'claude-opus-4-8',
        ANTHROPIC_MODEL: 'claude-opus-4-8',
        API_TIMEOUT_MS: '3000000',
        CLAUDE_CODE_ATTRIBUTION_HEADER: '0',
        CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: '1'
      },
      includeCoAuthoredBy: false
    })
  })

  it('renders Codex CLI config with latest default model and Responses auth', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-codex',
        baseUrl: 'https://useaifor.me',
        platform: 'openai'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const codeBlocks = wrapper.findAll('pre code')
    const configToml = codeBlocks[0].text()
    const authJson = JSON.parse(codeBlocks[1].text())

    expect(configToml).toContain('model_provider = "OpenAI"')
    expect(configToml).toContain('model = "gpt-5.5"')
    expect(configToml).toContain('review_model = "gpt-5.4"')
    expect(configToml).toContain('model_reasoning_effort = "medium"')
    expect(configToml).toContain('base_url = "https://useaifor.me"')
    expect(configToml).toContain('wire_api = "responses"')
    expect(configToml).toContain('requires_openai_auth = true')
    expect(configToml).not.toContain('responses_websockets_v2')
    expect(authJson).toEqual({
      OPENAI_API_KEY: 'sk-codex'
    })
  })

  it('renders Codex CLI WebSocket config without removed feature flags', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-codex',
        baseUrl: 'https://useaifor.me',
        platform: 'openai'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const websocketTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexCliWs')
    )

    expect(websocketTab).toBeDefined()
    await websocketTab!.trigger('click')
    await nextTick()

    const configToml = wrapper.findAll('pre code')[0].text()

    expect(configToml).toContain('model = "gpt-5.5"')
    expect(configToml).toContain('review_model = "gpt-5.4"')
    expect(configToml).toContain('model_reasoning_effort = "medium"')
    expect(configToml).toContain('wire_api = "responses"')
    expect(configToml).toContain('supports_websockets = true')
    expect(configToml).not.toContain('[features]')
    expect(configToml).not.toContain('responses_websockets_v2')
  })

  it('renders GPT-5.4 mini entry in OpenCode config', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.text()).toContain('"name": "GPT-5.4 Mini"')
    expect(codeBlock.text()).not.toContain('"name": "GPT-5.4 Nano"')

    const config = JSON.parse(codeBlock.text())
    expect(config.model).toBe('openai/gpt-5.5')
    expect(config.small_model).toBe('openai/gpt-5.4-mini')
    expect(config.provider.openai.options).toMatchObject({
      baseURL: 'https://example.com/v1',
      apiKey: 'sk-test',
      timeout: 3000000,
      headerTimeout: 3000000,
      chunkTimeout: 3000000
    })
  })

  it('renders gpt-image-2 configs for image-enabled OpenAI groups across Codex and OpenCode tabs', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-image',
        baseUrl: 'https://example.com',
        platform: 'openai',
        allowImageGeneration: true
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const codexBlocks = wrapper.findAll('pre code')
    const codexConfigToml = codexBlocks[0].text()
    const codexRequest = JSON.parse(codexBlocks[2].text())

    expect(codexConfigToml).toContain('model = "gpt-image-2"')
    expect(codexConfigToml).not.toContain('model = "gpt-5.5"')
    expect(codexConfigToml).not.toContain('review_model = "gpt-5.4"')
    expect(codexRequest).toMatchObject({
      model: 'gpt-image-2',
      size: '1K',
      response_format: 'b64_json'
    })

    const websocketTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexCliWs')
    )
    expect(websocketTab).toBeDefined()
    await websocketTab!.trigger('click')
    await nextTick()

    const websocketConfigToml = wrapper.findAll('pre code')[0].text()
    expect(websocketConfigToml).toContain('model = "gpt-image-2"')
    expect(websocketConfigToml).toContain('supports_websockets = true')
    expect(websocketConfigToml).not.toContain('model = "gpt-5.5"')

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )
    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const opencodeConfig = JSON.parse(wrapper.find('pre code').text())
    expect(opencodeConfig.model).toBe('openai/gpt-image-2')
    expect(opencodeConfig.small_model).toBe('openai/gpt-image-2')
    expect(opencodeConfig.provider.openai.models['gpt-image-2']).toMatchObject({
      name: 'GPT Image 2',
      modalities: {
        input: ['text', 'image'],
        output: ['image']
      }
    })
    expect(opencodeConfig.provider.openai.models).not.toHaveProperty('gpt-5.5')
  })

  it('renders OpenCode Anthropic config with explicit Claude default model', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-claude',
        baseUrl: 'https://useaifor.me',
        platform: 'anthropic'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const config = JSON.parse(wrapper.find('pre code').text())

    expect(config.model).toBe('anthropic/claude-opus-4-8')
    expect(config.small_model).toBe('anthropic/claude-opus-4-8')
    expect(config.provider.anthropic.npm).toBe('@ai-sdk/anthropic')
    expect(config.provider.anthropic.options).toMatchObject({
      baseURL: 'https://useaifor.me/v1',
      apiKey: 'sk-claude',
      timeout: 3000000,
      headerTimeout: 3000000,
      chunkTimeout: 3000000
    })
    expect(config.provider.anthropic.models['claude-opus-4-8']).toMatchObject({
      name: 'Claude Opus 4.8',
      limit: {
        context: 200000,
        output: 128000
      }
    })
  })
})
