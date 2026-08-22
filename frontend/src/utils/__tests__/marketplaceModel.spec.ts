import { describe, expect, it } from 'vitest'
import { normalizeMarketplaceModelName } from '@/utils/marketplaceModel'

describe('normalizeMarketplaceModelName', () => {
  it.each([
    ['claude-opus-4.8', 'claude-opus-4-8'],
    ['claude-opus-4.7', 'claude-opus-4-7'],
    ['claude-opus-4.6', 'claude-opus-4-6'],
    ['claude-sonnet-4.6', 'claude-sonnet-4-6'],
    ['claude-haiku-4.5', 'claude-haiku-4-5']
  ])('normalizes %s to a directly callable model ID', (input, expected) => {
    expect(normalizeMarketplaceModelName(input)).toBe(expected)
  })

  it('preserves already callable model IDs', () => {
    expect(normalizeMarketplaceModelName('claude-opus-4-8')).toBe('claude-opus-4-8')
    expect(normalizeMarketplaceModelName('gpt-5.6-sol')).toBe('gpt-5.6-sol')
  })
})
