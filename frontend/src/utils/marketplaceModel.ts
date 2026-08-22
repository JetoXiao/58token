const MARKETPLACE_MODEL_NAME_ALIASES: Record<string, string> = {
  'claude-opus-4.8': 'claude-opus-4-8',
  'claude-opus-4.7': 'claude-opus-4-7',
  'claude-opus-4.6': 'claude-opus-4-6',
  'claude-sonnet-4.6': 'claude-sonnet-4-6',
  'claude-haiku-4.5': 'claude-haiku-4-5'
}

// Marketplace names are copied verbatim into API clients. Normalize legacy
// dotted Claude version segments to the API-compatible hyphenated model ID.
export function normalizeMarketplaceModelName(modelName: string): string {
  const normalized = String(modelName || '').trim().toLowerCase()
  return MARKETPLACE_MODEL_NAME_ALIASES[normalized] || normalized.replace(/(claude-[^-]+-\d+)\.(\d+)(?=$|-)/, '$1-$2')
}
