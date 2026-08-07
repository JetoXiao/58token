<template>
  <div v-if="eligible" class="min-w-[105px] text-xs">
    <template v-if="balance">
      <div class="font-semibold text-gray-800 dark:text-gray-100" :title="balanceTitle">
        {{ formattedBalance }}
      </div>
      <div class="mt-0.5 text-[10px] text-gray-400">
        {{ balance.source === 'sub2api' ? 'Sub2API' : balance.source === 'newapi' ? 'New API' : balance.source }}
      </div>
    </template>
    <div
      v-else-if="state?.snapshot"
      class="text-gray-400"
      :title="t('admin.accounts.upstreamPricing.balanceUnavailableHint')"
    >
      {{ t('admin.accounts.upstreamPricing.balanceUnavailable') }}
    </div>
    <div
      v-else-if="state?.error"
      class="flex items-center gap-1 text-red-500 dark:text-red-400"
      :title="state.error"
    >
      <Icon name="exclamationCircle" size="xs" />
      <span>{{ t('admin.accounts.upstreamPricing.failedShort') }}</span>
    </div>
    <div v-else class="flex items-center gap-1 text-gray-400">
      <Icon name="refresh" size="xs" class="animate-spin" />
      <span>{{ t('admin.accounts.upstreamPricing.loading') }}</span>
    </div>
  </div>
  <span v-else class="text-sm text-gray-400">-</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AccountUpstreamPricingState } from '@/composables/useAccountUpstreamPricing'
import { accountUpstreamPricingKey } from '@/composables/useAccountUpstreamPricing'
import type { Account } from '@/types'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  account: Account
  state: AccountUpstreamPricingState | null
}>()
const { t } = useI18n()

const eligible = computed(() => accountUpstreamPricingKey(props.account) !== null)
const balance = computed(() => props.state?.snapshot?.balance || null)

const formattedBalance = computed(() => {
  const current = balance.value
  if (!current) return '-'
  if (current.unit === 'currency') {
    const currency = (current.currency || 'USD').toUpperCase()
    try {
      return new Intl.NumberFormat(undefined, {
        style: 'currency',
        currency,
        minimumFractionDigits: 2,
        maximumFractionDigits: 4
      }).format(current.amount)
    } catch {
      return `${currency} ${formatNumber(current.amount)}`
    }
  }
  return `${formatNumber(current.amount)} ${t('admin.accounts.upstreamPricing.balanceQuotaUnit')}`
})

const balanceTitle = computed(() => {
  const current = balance.value
  if (!current) return ''
  const lines = [formattedBalance.value]
  if (typeof current.used_amount === 'number') {
    lines.push(`${t('admin.accounts.upstreamPricing.balanceUsed')}: ${formatNumber(current.used_amount)}`)
  }
  if (current.checked_at) {
    lines.push(`${t('admin.accounts.upstreamPricing.balanceCheckedAt')}: ${new Date(current.checked_at).toLocaleString()}`)
  }
  return lines.join('\n')
})

function formatNumber(value: number): string {
  return value.toLocaleString(undefined, { maximumFractionDigits: 4 })
}
</script>
