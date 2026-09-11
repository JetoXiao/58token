import { describe, expect, it } from 'vitest'

import { syncMarketplaceFilterSelections } from '../marketplaceFilters'

describe('syncMarketplaceFilterSelections', () => {
  it('keeps the marketplace unfiltered when no selection exists', () => {
    expect(syncMarketplaceFilterSelections(
      { provider: '', group: '' },
      ['OpenAI', 'Anthropic'],
      ['1', '测试专属分组']
    )).toEqual({ provider: '', group: '' })
  })

  it('clears selections that are no longer available without selecting a fallback', () => {
    expect(syncMarketplaceFilterSelections(
      { provider: 'OpenAI', group: '1' },
      ['Anthropic'],
      ['测试专属分组']
    )).toEqual({ provider: '', group: '' })
  })

  it('preserves selections that are still available', () => {
    expect(syncMarketplaceFilterSelections(
      { provider: 'Anthropic', group: '测试专属分组' },
      ['OpenAI', 'Anthropic'],
      ['1', '测试专属分组']
    )).toEqual({ provider: 'Anthropic', group: '测试专属分组' })
  })
})
