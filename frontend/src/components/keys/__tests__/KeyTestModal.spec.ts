import { describe, expect, it, vi, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import KeyTestModal from '../KeyTestModal.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (params?.model) return `${key}:${params.model}`
      return key
    }
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn()
  })
}))

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: {
    show: { type: Boolean, default: false }
  },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

const SelectStub = defineComponent({
  name: 'Select',
  props: {
    modelValue: { type: [String, Number, Boolean, null], default: '' },
    options: { type: Array, default: () => [] }
  },
  emits: ['update:modelValue'],
  template: `
    <select :value="modelValue" @change="$emit('update:modelValue', $event.target.value)">
      <option v-for="option in options" :key="option.value" :value="option.value">
        {{ option.label }}
      </option>
    </select>
  `
})

const buildApiKey = () => ({
  id: 1,
  name: 'Claude key',
  key: 'sk-test',
  status: 'active',
  group: {
    id: 1,
    name: 'Claude Max',
    platform: 'anthropic'
  }
})

const buildCodexApiKey = () => ({
  id: 2,
  name: 'Codex key',
  key: 'sk-codex',
  status: 'active',
  group: {
    id: 2,
    name: 'Codex Pro',
    platform: 'openai'
  }
})

const buildCodexImageApiKey = () => ({
  id: 3,
  name: 'Codex image key',
  key: 'sk-image',
  status: 'active',
  group: {
    id: 3,
    name: 'Codex Image2',
    platform: 'openai',
    allow_image_generation: true
  }
})

describe('KeyTestModal', () => {
  const originalFetch = global.fetch

  afterEach(() => {
    global.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('loads supported models through the API key model endpoint', async () => {
    global.fetch = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          object: 'list',
          data: [
            {
              id: 'claude-opus-4-8',
              display_name: 'Claude Opus 4.8'
            }
          ]
        })
      } as any)
      .mockResolvedValueOnce({
        ok: true,
        body: {
          getReader: () => ({
            read: vi.fn().mockResolvedValue({ done: true, value: undefined })
          })
        }
      } as any)

    const wrapper = mount(KeyTestModal, {
      props: {
        show: false,
        apiKey: buildApiKey() as any
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          Icon: true
        }
      }
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    const option = wrapper.find('option')
    expect(option.text()).toBe('Claude Opus 4.8')
    expect(option.attributes('value')).toBe('claude-opus-4-8')

    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(2)
    expect((global.fetch as any).mock.calls[0][0]).toBe('/v1/models')
    expect((global.fetch as any).mock.calls[0][1].headers.Authorization).toBe('Bearer sk-test')
    const [url, options] = (global.fetch as any).mock.calls[1]
    expect(url).toBe('/v1/messages')
    expect(JSON.parse(options.body)).toMatchObject({
      model: 'claude-opus-4-8'
    })
  })

  it('tests Codex with account-supported models through Chat Completions API', async () => {
    const codexModelsResponse = {
      ok: true,
      json: async () => ({
        object: 'list',
        data: [
          { id: 'gpt-5.5', display_name: 'GPT-5.5' },
          { id: 'gpt-5.4', display_name: 'GPT-5.4' },
          { id: 'gpt-5.4-mini', display_name: 'GPT-5.4 Mini' },
          { id: 'gpt-5.3-codex', display_name: 'GPT-5.3 Codex' },
          { id: 'gpt-5.3-codex-spark', display_name: 'GPT-5.3 Codex Spark' },
          { id: 'gpt-5.2', display_name: 'GPT-5.2' },
          { id: 'gpt-image-2', display_name: 'GPT Image 2' }
        ]
      })
    } as any

    global.fetch = vi.fn()
      .mockResolvedValueOnce(codexModelsResponse)
      .mockResolvedValueOnce(codexModelsResponse)
      .mockResolvedValueOnce({
        ok: true,
        body: {
          getReader: () => ({
            read: vi.fn().mockResolvedValue({ done: true, value: undefined })
          })
        }
      } as any)

    const wrapper = mount(KeyTestModal, {
      props: {
        show: false,
        apiKey: buildCodexApiKey() as any
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          Icon: true
        }
      }
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('gpt-5.5')
    expect(wrapper.findAll('option').map((option) => option.attributes('value'))).toEqual([
      'gpt-5.5',
      'gpt-5.4',
      'gpt-5.4-mini',
      'gpt-5.3-codex',
      'gpt-5.3-codex-spark',
      'gpt-5.2',
      'gpt-image-2'
    ])

    await wrapper.find('select').setValue('gpt-5.4')
    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('gpt-5.4')

    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    await flushPromises()

    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('gpt-5.5')

    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(3)
    expect((global.fetch as any).mock.calls[0][0]).toBe('/v1/models')
    expect((global.fetch as any).mock.calls[0][1].headers.Authorization).toBe('Bearer sk-codex')
    const [url, options] = (global.fetch as any).mock.calls[2]
    expect(url).toBe('/v1/chat/completions')
    expect(JSON.parse(options.body)).toMatchObject({
      model: 'gpt-5.5',
      messages: [{ role: 'user', content: 'hi' }],
      max_tokens: 32,
      stream: true
    })
  })

  it('supports image generation test mode for image-capable OpenAI models', async () => {
    global.fetch = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          object: 'list',
          data: [
            { id: 'gpt-5.5', display_name: 'GPT-5.5' },
            { id: 'gpt-image-2', display_name: 'GPT Image 2' }
          ]
        })
      } as any)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          data: [{ b64_json: 'aGVsbG8=', revised_prompt: 'draw a cat' }]
        })
      } as any)

    const wrapper = mount(KeyTestModal, {
      props: {
        show: false,
        apiKey: buildCodexImageApiKey() as any
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          Icon: true
        }
      }
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    const modelSelect = wrapper.find('select').element as HTMLSelectElement
    expect(modelSelect.value).toBe('gpt-5.5')

    await wrapper.find('select').setValue('gpt-image-2')
    await flushPromises()

    expect(wrapper.text()).toContain('keys.testModal.imagePromptLabel')
    expect(wrapper.text()).toContain('keys.testModal.imageSizeLabel')

    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(2)
    const [url, options] = (global.fetch as any).mock.calls[1]
    expect(url).toBe('/v1/images/generations')
    expect(JSON.parse(options.body)).toMatchObject({
      model: 'gpt-image-2',
      size: '1K',
      response_format: 'b64_json'
    })
  })
})
