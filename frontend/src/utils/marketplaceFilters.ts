export interface MarketplaceFilterSelections {
  provider: string
  group: string
}

export function syncMarketplaceFilterSelections(
  current: MarketplaceFilterSelections,
  availableProviders: readonly string[],
  availableGroups: readonly string[]
): MarketplaceFilterSelections {
  return {
    provider: availableProviders.includes(current.provider) ? current.provider : '',
    group: availableGroups.includes(current.group) ? current.group : ''
  }
}
