import { describe, expect, it, vi } from 'vitest'
import { useAccountUpstreamPricing } from '../useAccountUpstreamPricing'
import type { Account } from '@/types'

function account(id: number, name: string): Account {
  return {
    id,
    name,
    platform: 'openai',
    type: 'apikey',
    status: 'active',
    credentials: { base_url: 'https://same-upstream.example.com/v1', upstream_group_key: `group-${id}` }
  } as Account
}

describe('useAccountUpstreamPricing', () => {
  it('loads accounts independently even when they use the same upstream host', async () => {
    const fetchPricing = vi.fn(async (accountId: number) => ({
      source: 'newapi',
      endpoint: 'https://same-upstream.example.com/api/pricing',
      checked_at: '2026-08-07T00:00:00Z',
      group_ratios: { [`group-${accountId}`]: accountId / 10 },
      group_names: {},
      models: []
    }))
    const loader = useAccountUpstreamPricing({ fetchPricing, concurrency: 2, now: () => 1000 })
    const accountA = account(1, 'Group A')
    const accountB = account(2, 'Group B')

    await loader.load([accountA, accountB])

    expect(fetchPricing).toHaveBeenCalledTimes(2)
    expect(fetchPricing).toHaveBeenCalledWith(1)
    expect(fetchPricing).toHaveBeenCalledWith(2)
    expect(loader.stateFor(accountA)?.snapshot?.group_ratios).toEqual({ 'group-1': 0.1 })
    expect(loader.stateFor(accountB)?.snapshot?.group_ratios).toEqual({ 'group-2': 0.2 })
  })

  it('does not refetch a fresh account during table auto refresh', async () => {
    let now = 1000
    const fetchPricing = vi.fn(async () => ({
      source: 'newapi', endpoint: 'https://example.com/api/pricing', checked_at: '',
      group_ratios: {}, group_names: {}, models: []
    }))
    const loader = useAccountUpstreamPricing({ fetchPricing, staleMs: 300_000, now: () => now })
    const row = account(3, 'Group C')

    await loader.load([row])
    now += 5_000
    await loader.load([{ ...row }])

    expect(fetchPricing).toHaveBeenCalledTimes(1)
  })
})
