const NON_CAPABILITY_TAGS = new Set([
  'openai',
  'openai api',
  'openai-compatible',
  'openai-compatible api',
  'anthropic',
  'anthropic api',
  'gemini',
  'gemini api',
  'antigravity'
])

export function parseMarketplaceCapabilities(tags: string[]): string[] {
  const tagItems = tags
    .map((tag) => tag.trim())
    .filter((tag) => tag && !NON_CAPABILITY_TAGS.has(tag.toLowerCase()))
  return Array.from(new Set(tagItems))
}
