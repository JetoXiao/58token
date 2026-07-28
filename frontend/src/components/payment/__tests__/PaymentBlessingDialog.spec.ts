import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick, ref } from 'vue'
import PaymentBlessingDialog from '../PaymentBlessingDialog.vue'
import { PAYMENT_BLESSINGS } from '../paymentBlessings'

const locale = ref('zh')
const messages: Record<string, Record<string, string>> = {
  zh: {
    'payment.blessing.eyebrow': '一份小小的心意',
    'payment.blessing.title': '感谢你的支持',
    'payment.blessing.description': '支付已顺利完成，送你一句此刻的祝福。',
    'payment.blessing.accept': '收下这份祝福',
    'payment.blessing.closeLabel': '关闭祝福',
  },
  en: {
    'payment.blessing.eyebrow': 'A little something for you',
    'payment.blessing.title': 'Thank you for your support',
    'payment.blessing.description': 'Your payment is complete. Here is a wish for the road ahead.',
    'payment.blessing.accept': 'Keep this wish',
    'payment.blessing.closeLabel': 'Close this wish',
  },
}

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    locale,
    t: (key: string) => messages[locale.value]?.[key] || key,
  }),
}))

describe('PaymentBlessingDialog', () => {
  beforeEach(() => {
    window.sessionStorage.clear()
    document.body.innerHTML = ''
    document.body.style.overflow = ''
    locale.value = 'zh'
  })

  afterEach(() => {
    document.body.innerHTML = ''
    document.body.style.overflow = ''
  })

  it('switches both interface copy and blessing with the locale', async () => {
    const wrapper = mount(PaymentBlessingDialog, {
      attachTo: document.body,
      props: { modelValue: true, orderKey: 'order-localized' },
    })

    expect(document.body.textContent).toContain('感谢你的支持')
    expect(PAYMENT_BLESSINGS.zh.some((line) => document.body.textContent?.includes(line))).toBe(true)

    locale.value = 'en'
    await nextTick()

    expect(document.body.textContent).toContain('Thank you for your support')
    expect(PAYMENT_BLESSINGS.en.some((line) => document.body.textContent?.includes(line))).toBe(true)
    wrapper.unmount()
  })

  it('only shows once per order in the same browser session', async () => {
    window.sessionStorage.setItem('useaiforme:payment-blessing:order-seen', '1')
    const wrapper = mount(PaymentBlessingDialog, {
      props: { modelValue: true, orderKey: 'order-seen' },
    })

    await nextTick()
    expect(wrapper.emitted('update:modelValue')).toEqual([[false]])
  })
})
