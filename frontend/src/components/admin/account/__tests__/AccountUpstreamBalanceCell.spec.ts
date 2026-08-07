import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AccountUpstreamBalanceCell from '../AccountUpstreamBalanceCell.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

describe('AccountUpstreamBalanceCell', () => {
  it('shows a live upstream currency balance', () => {
    const wrapper = mount(AccountUpstreamBalanceCell, {
      props: {
        account: {
          id: 10,
          name: 'Sub2API upstream',
          platform: 'openai',
          type: 'apikey',
          status: 'active',
          credentials: { base_url: 'https://upstream.example.com/v1', upstream_group_key: '12' }
        } as any,
        state: {
          loading: false,
          error: '',
          updatedAt: Date.now(),
          snapshot: {
            source: 'sub2api',
            endpoint: 'https://upstream.example.com/api/v1/groups/available',
            checked_at: '',
            group_ratios: { 12: 0.2 },
            group_names: { 12: 'Codex Pro' },
            models: [],
            balance: {
              amount: 12.3456,
              unit: 'currency',
              currency: 'USD',
              source: 'sub2api',
              endpoint: 'https://upstream.example.com/api/v1/auth/me',
              checked_at: '2026-08-07T05:00:00Z'
            }
          }
        }
      },
      global: { stubs: { Icon: true } }
    })

    expect(wrapper.text()).toContain('$12.3456')
    expect(wrapper.text()).toContain('Sub2API')
  })

  it('shows unavailable without failing the pricing row', () => {
    const wrapper = mount(AccountUpstreamBalanceCell, {
      props: {
        account: {
          id: 11,
          name: 'Public pricing only',
          platform: 'openai',
          type: 'apikey',
          status: 'active',
          credentials: { base_url: 'https://upstream.example.com/v1', upstream_group_key: 'default' }
        } as any,
        state: {
          loading: false,
          error: '',
          updatedAt: Date.now(),
          snapshot: {
            source: 'newapi',
            endpoint: 'https://upstream.example.com/api/pricing',
            checked_at: '',
            group_ratios: { default: 0.15 },
            group_names: { default: 'Default' },
            models: []
          }
        }
      },
      global: { stubs: { Icon: true } }
    })

    expect(wrapper.text()).toContain('admin.accounts.upstreamPricing.balanceUnavailable')
    expect(wrapper.text()).not.toContain('admin.accounts.upstreamPricing.failedShort')
  })
})
