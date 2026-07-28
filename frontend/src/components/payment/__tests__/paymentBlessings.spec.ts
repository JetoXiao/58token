import { describe, expect, it } from 'vitest'
import {
  PAYMENT_BLESSINGS,
  normalizePaymentBlessingLanguage,
  randomPaymentBlessing,
} from '../paymentBlessings'

describe('paymentBlessings', () => {
  it('provides dozens of localized blessings', () => {
    expect(PAYMENT_BLESSINGS.zh).toHaveLength(40)
    expect(PAYMENT_BLESSINGS.en).toHaveLength(40)
    expect(new Set(PAYMENT_BLESSINGS.zh).size).toBe(40)
    expect(new Set(PAYMENT_BLESSINGS.en).size).toBe(40)
  })

  it('normalizes Chinese locales and falls back to English', () => {
    expect(normalizePaymentBlessingLanguage('zh-CN')).toBe('zh')
    expect(normalizePaymentBlessingLanguage('zh-TW')).toBe('zh')
    expect(normalizePaymentBlessingLanguage('en-US')).toBe('en')
    expect(normalizePaymentBlessingLanguage(undefined)).toBe('en')
  })

  it('selects deterministically across the full list boundaries', () => {
    expect(randomPaymentBlessing('zh-CN', () => 0)).toBe(PAYMENT_BLESSINGS.zh[0])
    expect(randomPaymentBlessing('zh-CN', () => 0.999999)).toBe(PAYMENT_BLESSINGS.zh[39])
    expect(randomPaymentBlessing('en-US', () => 0)).toBe(PAYMENT_BLESSINGS.en[0])
    expect(randomPaymentBlessing('en-US', () => 0.999999)).toBe(PAYMENT_BLESSINGS.en[39])
  })
})
