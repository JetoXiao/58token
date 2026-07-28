import { describe, expect, it } from 'vitest'
import en from '../locales/en'
import zh from '../locales/zh'

describe('payment blessing locales', () => {
  it('defines localized payment blessing copy at the payment namespace', () => {
    expect(zh.payment.blessing.title).toBe('感谢你的支持')
    expect(en.payment.blessing.title).toBe('Thank you for your support')
  })
})
