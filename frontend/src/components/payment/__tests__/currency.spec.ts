import { describe, expect, it } from 'vitest'
import { formatPaymentAmount } from '../currency'

describe('formatPaymentAmount', () => {
  it('uses the currency default fraction digits', () => {
    expect(formatPaymentAmount(100, 'JPY', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'KRW', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'HKD', 'en-US')).toContain('.00')
  })

  it('keeps stablecoin paid amount precision for reconciliation', () => {
    expect(formatPaymentAmount(0.1429, 'USDT', 'en-US')).toBe('0.1429 USDT')
    expect(formatPaymentAmount(71.4286, 'USDT', 'en-US')).toBe('71.4286 USDT')
  })
})
