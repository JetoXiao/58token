/**
 * Shared utility functions for payment order display.
 * Used by AdminOrderDetail, AdminOrderTable, AdminRefundDialog, AdminOrdersView, etc.
 */

import type { PaymentOrder } from '@/types/payment'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'

const STATUS_BADGE_MAP: Record<string, string> = {
  PENDING: 'badge-warning',
  PAID: 'badge-info',
  RECHARGING: 'badge-info',
  COMPLETED: 'badge-success',
  EXPIRED: 'badge-secondary',
  CANCELLED: 'badge-secondary',
  FAILED: 'badge-danger',
  REFUND_REQUESTED: 'badge-warning',
  REFUNDING: 'badge-warning',
  PARTIALLY_REFUNDED: 'badge-warning',
  REFUNDED: 'badge-info',
  REFUND_FAILED: 'badge-danger',
}

const REFUNDABLE_STATUSES = ['COMPLETED', 'PARTIALLY_REFUNDED', 'REFUND_REQUESTED', 'REFUND_FAILED']

export function statusBadgeClass(status: string): string {
  return STATUS_BADGE_MAP[status] || 'badge-secondary'
}

export function canRefund(status: string): boolean {
  return REFUNDABLE_STATUSES.includes(status)
}

export function formatOrderDateTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

export function isUsdtOrder(row: PaymentOrder): boolean {
  return row.payment_type === 'usdt' || row.payment_type === 'infini' || normalizePaymentCurrency(row.currency) === 'USDT'
}

export function orderPayCurrency(row: PaymentOrder): string {
  if (isUsdtOrder(row)) {
    return 'USDT'
  }
  return normalizePaymentCurrency(row.currency)
}

export function formatOrderPayAmount(row: PaymentOrder, value: number = row.pay_amount, locale?: string): string {
  return formatPaymentAmount(Number(value) || 0, orderPayCurrency(row), locale)
}

export function formatOrderCreditedAmount(row: PaymentOrder, locale?: string): string {
  return row.order_type === 'balance'
    ? formatPaymentAmount(row.amount, 'USD', locale)
    : formatPaymentAmount(row.amount, orderPayCurrency(row), locale)
}
