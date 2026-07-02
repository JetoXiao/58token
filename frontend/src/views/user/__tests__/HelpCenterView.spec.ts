import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import HelpCenterView from '../HelpCenterView.vue'
import type { HelpCenterConfig } from '@/types'

const { helpCenterGet, apiGet, showError } = vi.hoisted(() => ({
  helpCenterGet: vi.fn(),
  apiGet: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/api', () => ({
  helpCenterAPI: {
    get: helpCenterGet,
  },
  apiClient: {
    get: apiGet,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const IconStub = { template: '<span data-icon="true" />' }
const RouterLinkStub = {
  props: ['to'],
  template: '<a :href="to"><slot /></a>',
}

function helpCenterConfig(): HelpCenterConfig {
  return {
    enabled: true,
    base_url: 'https://example.com',
    title: '帮助中心',
    description: '教程',
    key_created_prompt: {
      enabled: true,
      title: '',
      description: '',
      primary_action_label: '',
      primary_action_url: '',
      secondary_action_label: '',
      secondary_action_url: '',
      dismiss_label: '',
    },
    tutorials: [
      {
        id: 'codex',
        enabled: true,
        sort_order: 1,
        title: 'Codex',
        badge: 'Desktop',
        summary: '配置 Codex',
        content_md: '',
        steps: [
          {
            title: '准备 API Key',
            description: '创建一个 key',
            code_blocks: [
              {
                title: 'Install command',
                language: 'bash',
                content: 'cc-switch --version',
              },
            ],
            images: [
              {
                label: '创建 key 截图',
                url: 'https://example.com/help/create-key.png',
                file_name: 'create-key.png',
              },
            ],
            attachments: [
              {
                label: 'Codex config example',
                url: '/api/v1/help-center/attachments/codex-config.zip',
                file_name: 'codex-config.zip',
              },
            ],
          },
        ],
        code_blocks: [],
        links: [],
        attachments: [],
      },
    ],
    faqs: [],
  }
}

function helpCenterConfigWithUploadedImages(): HelpCenterConfig {
  const config = helpCenterConfig()
  config.tutorials[0].steps[0].images = [
    {
      label: 'first',
      url: '/api/v1/help-center/attachments/first.png',
      file_name: 'first.png',
    },
    {
      label: 'second',
      url: 'http://127.0.0.1:8080/api/v1/help-center/attachments/second.png',
      file_name: 'second.png',
    },
  ]
  return config
}

function mountHelpCenter() {
  return mount(HelpCenterView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Icon: IconStub,
        RouterLink: RouterLinkStub,
      },
    },
  })
}

describe('HelpCenterView image preview', () => {
  beforeEach(() => {
    helpCenterGet.mockReset()
    apiGet.mockReset()
    showError.mockReset()
    Object.defineProperty(URL, 'createObjectURL', {
      value: vi.fn(),
      configurable: true,
      writable: true,
    })
    Object.defineProperty(URL, 'revokeObjectURL', {
      value: vi.fn(),
      configurable: true,
      writable: true,
    })
    vi.spyOn(URL, 'createObjectURL').mockImplementation((blob) => `blob:${String((blob as Blob).size)}`)
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})
    helpCenterGet.mockResolvedValue({
      config: helpCenterConfig(),
      key_prompt_dismissed: false,
      help_center_key_prompt_dismissed: false,
    })
  })

  it('opens step images in a preview dialog and closes it', async () => {
    const wrapper = mountHelpCenter()
    await flushPromises()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)

    await wrapper.get('[data-testid="help-center-step-image"]').trigger('click')

    const dialog = wrapper.get('[role="dialog"]')
    expect(dialog.text()).toContain('创建 key 截图')
    expect(dialog.html()).not.toContain('bg-white')
    expect(dialog.get('[data-testid="help-center-preview-frame"]').classes()).toContain('help-center-preview-frame--zoom')
    const previewImage = dialog.get('[data-testid="help-center-preview-image"]')
    expect(previewImage.attributes('src')).toBe('https://example.com/help/create-key.png')
    expect(previewImage.classes()).toContain('help-center-preview-image--zoom')

    Object.defineProperty(previewImage.element, 'naturalWidth', { value: 960, configurable: true })
    await previewImage.trigger('load')

    expect(dialog.get('[data-testid="help-center-preview-frame"]').attributes('style')).toContain('--preview-natural-width: 960px')

    await dialog.get('[data-testid="help-center-preview-close"]').trigger('click')

    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('shows downloadable attachments inside each step', async () => {
    const wrapper = mountHelpCenter()
    await flushPromises()

    const image = wrapper.get('[data-testid="help-center-step-image"]')
    const attachment = wrapper.get('[data-testid="help-center-step-attachment"]')
    const codeBlock = wrapper.get('pre')

    expect(attachment.text()).toContain('Codex config example')
    expect(attachment.text()).toContain('codex-config.zip')
    expect(Boolean(image.element.compareDocumentPosition(attachment.element) & Node.DOCUMENT_POSITION_FOLLOWING)).toBe(true)
    expect(Boolean(attachment.element.compareDocumentPosition(codeBlock.element) & Node.DOCUMENT_POSITION_FOLLOWING)).toBe(true)

    await attachment.trigger('click')

    expect(apiGet).toHaveBeenCalledWith('/help-center/attachments/codex-config.zip', { responseType: 'blob' })
  })

  it('renders uploaded images with native browser urls instead of blob downloads', async () => {
    helpCenterGet.mockResolvedValue({
      config: helpCenterConfigWithUploadedImages(),
      key_prompt_dismissed: false,
      help_center_key_prompt_dismissed: false,
    })

    const wrapper = mountHelpCenter()
    await flushPromises()

    const images = wrapper.findAll('img')
    expect(images).toHaveLength(2)
    expect(images[0].attributes('src')).toBe('/api/v1/help-center/attachments/first.png')
    expect(images[0].attributes('loading')).toBe('lazy')
    expect(images[0].attributes('decoding')).toBe('async')
    expect(images[1].attributes('src')).toBe('/api/v1/help-center/attachments/second.png')
    expect(apiGet).not.toHaveBeenCalled()
  })
})
