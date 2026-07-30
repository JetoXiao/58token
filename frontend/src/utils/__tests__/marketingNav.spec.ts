import { describe, expect, it } from 'vitest'

import {
  DEFAULT_MARKETING_NAV_ITEMS,
  normalizeMarketingNavItems,
} from '@/utils/marketingNav'

describe('marketingNav utils', () => {
  it('defaults to all marketing nav items when the setting is missing', () => {
    expect(normalizeMarketingNavItems(undefined)).toEqual(
      DEFAULT_MARKETING_NAV_ITEMS,
    )
  })

  it('preserves an explicit empty selection', () => {
    expect(normalizeMarketingNavItems([])).toEqual([])
    expect(normalizeMarketingNavItems('[]')).toEqual([])
  })

  it('normalizes supported items in display order', () => {
    expect(normalizeMarketingNavItems(['partner', 'resources', 'unknown', 'models'])).toEqual(
      ['models', 'resources', 'partner'],
    )
    expect(normalizeMarketingNavItems('["partner","resources","docs"]')).toEqual([
      'docs',
      'resources',
      'partner',
    ])
  })
})
