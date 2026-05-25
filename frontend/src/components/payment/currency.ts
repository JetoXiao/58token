export const DEFAULT_PAYMENT_CURRENCY = 'CNY'
const STABLECOIN_MAX_FRACTION_DIGITS = 8

export function normalizePaymentCurrency(currency?: string | null): string {
  const normalized = String(currency || '').trim().toUpperCase()
  if (normalized === 'USDT') return normalized
  return /^[A-Z]{3}$/.test(normalized) ? normalized : DEFAULT_PAYMENT_CURRENCY
}

function paymentCurrencyFractionDigits(currency: string): number {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
    }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

export function formatPaymentAmount(amount: number, currency?: string | null, locale?: string): string {
  const normalized = normalizePaymentCurrency(currency)
  if (normalized === 'USDT') {
    const value = Number.isFinite(amount) ? amount : 0
    return `${new Intl.NumberFormat(locale || undefined, {
      minimumFractionDigits: 0,
      maximumFractionDigits: STABLECOIN_MAX_FRACTION_DIGITS,
    }).format(value)} USDT`
  }
  const fractionDigits = paymentCurrencyFractionDigits(normalized)
  try {
    return new Intl.NumberFormat(locale || undefined, {
      style: 'currency',
      currency: normalized,
      currencyDisplay: 'narrowSymbol',
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits,
    }).format(Number.isFinite(amount) ? amount : 0)
  } catch {
    return `${normalized} ${(Number.isFinite(amount) ? amount : 0).toFixed(fractionDigits)}`
  }
}
