export type MarketingNavItem = 'models' | 'docs' | 'partner'

export const DEFAULT_MARKETING_NAV_ITEMS: MarketingNavItem[] = ['models', 'docs', 'partner']

const MARKETING_NAV_ITEM_SET = new Set<MarketingNavItem>(DEFAULT_MARKETING_NAV_ITEMS)

export function normalizeMarketingNavItems(value: unknown): MarketingNavItem[] {
  const parsedValue = parseMarketingNavItemsValue(value)
  if (!Array.isArray(parsedValue)) {
    return [...DEFAULT_MARKETING_NAV_ITEMS]
  }

  const enabled = new Set<MarketingNavItem>()
  for (const item of parsedValue) {
    if (typeof item !== 'string') continue
    if (MARKETING_NAV_ITEM_SET.has(item as MarketingNavItem)) {
      enabled.add(item as MarketingNavItem)
    }
  }

  return DEFAULT_MARKETING_NAV_ITEMS.filter((item) => enabled.has(item))
}

function parseMarketingNavItemsValue(value: unknown): unknown {
  if (typeof value !== 'string') {
    return value
  }

  const trimmed = value.trim()
  if (!trimmed.startsWith('[')) {
    return value
  }

  try {
    return JSON.parse(trimmed)
  } catch {
    return value
  }
}
