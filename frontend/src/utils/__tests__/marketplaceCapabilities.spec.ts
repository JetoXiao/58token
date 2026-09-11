import { describe, expect, it } from 'vitest'

import { parseMarketplaceCapabilities } from '../marketplaceCapabilities'

describe('parseMarketplaceCapabilities', () => {
  it('keeps model tags but filters protocol labels', () => {
    expect(parseMarketplaceCapabilities([
      'Reasoning',
      'OpenAI API',
      'openai',
      'Tools',
      'Reasoning'
    ])).toEqual(['Reasoning', 'Tools'])
  })
})
