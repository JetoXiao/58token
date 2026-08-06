import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AccountUpstreamPricingCell from '../AccountUpstreamPricingCell.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => `${key}:${JSON.stringify(params || {})}`
    })
  }
})

describe('AccountUpstreamPricingCell', () => {
  it('shows the upstream group selected for the account', () => {
    const wrapper = mount(AccountUpstreamPricingCell, {
      props: {
        account: {
          id: 10,
          name: 'Multi-group upstream',
          platform: 'openai',
          type: 'apikey',
          status: 'active',
          credentials: { base_url: 'https://upstream.example.com/v1', upstream_group_key: 'codex_pro' }
        } as any,
        state: {
          loading: false,
          error: '',
          updatedAt: Date.now(),
          snapshot: {
            source: 'newapi',
            endpoint: 'https://upstream.example.com/api/pricing',
            checked_at: '',
            group_ratios: { codex_pro: 0.45 },
            group_names: { codex_pro: 'Codex Pro' },
            models: [],
          }
        }
      },
      global: { stubs: { Icon: true } }
    })

    expect(wrapper.text()).toContain('Codex Pro 0.45x')
    expect(wrapper.text()).not.toContain('Default')
  })
})
