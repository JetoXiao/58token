import { describe, expect, it } from 'vitest'
import { formatMarketplaceMultiplier } from '@/utils/marketplaceDisplay'

describe('formatMarketplaceMultiplier', () => {
  it('formats the multiplier as a compact labeled capsule value', () => {
    expect(formatMarketplaceMultiplier('倍率', 1.25)).toBe('倍率 1.25x')
  })

  it('uses a placeholder for invalid multiplier values', () => {
    expect(formatMarketplaceMultiplier('倍率', 0)).toBe('倍率 -')
  })
})
