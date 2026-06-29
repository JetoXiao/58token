<template>
  <div
    class="group relative"
    @mouseenter="showProcess($event)"
    @mouseleave="hideProcess"
  >
    <div
      class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-emerald-50 transition-colors group-hover:bg-emerald-100 dark:bg-emerald-900/30 dark:group-hover:bg-emerald-900/60"
      :title="t('usage.billingProcess')"
    >
      <Icon
        name="calculator"
        size="xs"
        class="text-emerald-500 group-hover:text-emerald-600 dark:text-emerald-400 dark:group-hover:text-emerald-300"
      />
    </div>
  </div>

  <Teleport to="body">
    <div
      v-if="visible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{ left: position.x + 'px', top: position.y + 'px' }"
    >
      <div
        class="overflow-auto whitespace-normal rounded-lg border border-gray-700 bg-gray-900 px-3 py-2.5 text-xs text-white shadow-xl dark:border-gray-600 dark:bg-gray-800"
        style="width: min(430px, calc(100vw - 24px)); max-height: min(72vh, 560px);"
      >
        <div class="mb-2 flex items-center justify-between gap-4 border-b border-gray-700 pb-1.5">
          <span class="font-semibold text-gray-200">{{ t('usage.billingProcess') }}</span>
          <span class="rounded bg-emerald-500/15 px-1.5 py-0.5 font-medium text-emerald-300">
            {{ billingModeLabel }}
          </span>
        </div>

        <div class="space-y-2">
          <div class="space-y-1.5">
            <div v-if="row.request_id" class="grid grid-cols-[76px_minmax(0,1fr)] gap-3">
              <span class="text-gray-400">{{ t('admin.usage.requestId') }}</span>
              <span class="break-all font-mono text-[11px] text-gray-100">{{ row.request_id }}</span>
            </div>
            <div v-if="row.model" class="grid grid-cols-[76px_minmax(0,1fr)] gap-3">
              <span class="text-gray-400">{{ t('usage.model') }}</span>
              <span class="break-all font-medium text-gray-100">{{ row.model }}</span>
            </div>
            <div v-if="row.inbound_endpoint" class="grid grid-cols-[76px_minmax(0,1fr)] gap-3">
              <span class="text-gray-400">{{ t('usage.path') }}</span>
              <span class="break-all text-gray-100">{{ row.inbound_endpoint }}</span>
            </div>
          </div>

          <div class="border-t border-gray-700 pt-2">
            <div class="mb-1 text-xs font-semibold text-gray-300">{{ t('usage.priceDetails') }}</div>
            <div class="space-y-1.5">
              <div
                v-for="detail in priceDetails"
                :key="detail.label"
                class="flex items-center justify-between gap-4"
              >
                <span class="text-gray-400">{{ detail.label }}</span>
                <span class="text-right font-medium" :class="detail.className">{{ detail.value }}</span>
              </div>
            </div>
          </div>

          <div class="border-t border-gray-700 pt-2">
            <div class="mb-1 text-xs font-semibold text-gray-300">{{ t('usage.billingFormula') }}</div>
            <div class="rounded-md bg-gray-800/80 px-2 py-1.5 text-gray-100">
              <div
                v-for="line in formulaLines"
                :key="line"
                class="break-words leading-relaxed"
              >
                {{ line }}
              </div>
            </div>
          </div>

          <div class="space-y-1.5 border-t border-gray-700 pt-2">
            <div class="flex items-center justify-between gap-6">
              <span class="text-gray-400">{{ t('usage.serviceTier') }}</span>
              <span class="font-semibold text-cyan-300">{{ serviceTierLabel }}</span>
            </div>
            <div class="flex items-center justify-between gap-6">
              <span class="text-gray-400">{{ t('usage.original') }}</span>
              <span class="font-medium text-white">{{ formatMoney(totalCost) }}</span>
            </div>
            <div class="flex items-center justify-between gap-6">
              <span class="text-gray-400">{{ t('usage.groupMultiplier') }}</span>
              <span class="font-semibold text-blue-400">{{ formatRate(groupRate) }}</span>
            </div>
            <div class="flex items-center justify-between gap-6">
              <span class="text-gray-400">{{ t('usage.userBilled') }}</span>
              <span class="font-semibold text-green-400">{{ formatMoney(actualCost) }}</span>
            </div>
          </div>

          <div v-if="showAccountBilling" class="space-y-1.5 border-t border-gray-700 pt-2">
            <div class="mb-1 text-xs font-semibold text-gray-300">{{ t('usage.accountBillingProcess') }}</div>
            <div class="rounded-md bg-gray-800/80 px-2 py-1.5 text-gray-100">
              <div class="break-words leading-relaxed">{{ accountFormula }}</div>
            </div>
          </div>

          <div class="border-t border-gray-700 pt-2 text-[11px] leading-relaxed text-gray-400">
            {{ t('usage.billingReferenceNote') }}
          </div>
        </div>

        <div
          v-if="placement === 'right'"
          class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-gray-900 border-t-transparent dark:border-r-gray-800"
        ></div>
        <div
          v-else
          class="absolute left-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-l-[6px] border-t-[6px] border-b-transparent border-l-gray-900 border-t-transparent dark:border-l-gray-800"
        ></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUsageLog, UsageLog } from '@/types'
import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_PER_REQUEST,
  BILLING_MODE_TOKEN,
  getBillingModeLabel,
} from '@/utils/billingMode'
import { formatMultiplier } from '@/utils/formatters'
import { getUsageServiceTierLabel } from '@/utils/usageServiceTier'

type BillingRow = UsageLog | AdminUsageLog

interface PriceDetail {
  label: string
  value: string
  className: string
}

interface TokenPart {
  label: string
  tokens: number
  cost: number
  className: string
}

const props = withDefaults(
  defineProps<{
    row: BillingRow
    showAccount?: boolean
  }>(),
  {
    showAccount: false
  }
)

const { t } = useI18n()

const visible = ref(false)
const position = ref({ x: 0, y: 0 })
const placement = ref<'right' | 'left'>('right')
const row = computed(() => props.row)

const isFiniteNumber = (value: unknown): value is number => {
  return typeof value === 'number' && Number.isFinite(value)
}

const numberValue = (value: number | null | undefined): number => {
  return isFiniteNumber(value) ? value : 0
}

const formatMoney = (value: number | null | undefined, digits = 6): string => {
  return `$${numberValue(value).toFixed(digits)}`
}

const formatRate = (value: number | null | undefined): string => {
  return `${formatMultiplier(numberValue(value) || 0)}x`
}

const formatTokenCount = (value: number): string => {
  return Math.max(0, Math.round(numberValue(value))).toLocaleString()
}

const pricePerMillion = (cost: number, tokens: number): number | null => {
  if (tokens <= 0 || cost <= 0) {
    return null
  }
  return (cost / tokens) * 1_000_000
}

const formatTokenPrice = (cost: number, tokens: number): string => {
  const price = pricePerMillion(cost, tokens)
  if (price == null) {
    return '-'
  }
  return `${formatMoney(price)} ${t('usage.perMillionTokens')}`
}

const isImageUsage = computed(() => numberValue(row.value.image_count) > 0)

const displayBillingMode = computed(() => {
  if (isImageUsage.value) {
    return BILLING_MODE_IMAGE
  }
  return row.value.billing_mode || BILLING_MODE_TOKEN
})

const billingModeLabel = computed(() => getBillingModeLabel(displayBillingMode.value, t))
const serviceTierLabel = computed(() => getUsageServiceTierLabel(row.value.service_tier, t))
const totalCost = computed(() => numberValue(row.value.total_cost))
const actualCost = computed(() => numberValue(row.value.actual_cost))
const groupRate = computed(() => numberValue(row.value.rate_multiplier) || 1)

const tokenParts = computed<TokenPart[]>(() => {
  const current = row.value
  const parts: TokenPart[] = []

  const addPart = (label: string, tokens: number, cost: number, className: string) => {
    if (tokens > 0 || cost > 0) {
      parts.push({ label, tokens, cost, className })
    }
  }

  const imageOutputTokens = numberValue(current.image_output_tokens)
  const textOutputTokens = Math.max(numberValue(current.output_tokens) - imageOutputTokens, 0)

  addPart(t('usage.promptTokens'), numberValue(current.input_tokens), numberValue(current.input_cost), 'text-sky-300')
  addPart(t('usage.completionTokens'), textOutputTokens, numberValue(current.output_cost), 'text-violet-300')
  addPart(
    t('usage.cacheCreation'),
    numberValue(current.cache_creation_tokens),
    numberValue(current.cache_creation_cost),
    'text-amber-300'
  )
  addPart(t('usage.cacheRead'), numberValue(current.cache_read_tokens), numberValue(current.cache_read_cost), 'text-blue-300')
  addPart(
    t('usage.imageOutputTokens'),
    imageOutputTokens,
    numberValue(current.image_output_cost),
    'text-pink-300'
  )

  return parts
})

const knownComponentCost = computed(() => {
  return tokenParts.value.reduce((sum, part) => sum + part.cost, 0)
})

const residualCost = computed(() => {
  const residual = totalCost.value - knownComponentCost.value
  return Math.abs(residual) >= 0.0000005 ? residual : 0
})

const unitCount = computed(() => {
  if (displayBillingMode.value === BILLING_MODE_IMAGE) {
    return Math.max(1, numberValue(row.value.image_count))
  }
  return 1
})

const unitPrice = computed(() => {
  const count = unitCount.value
  if (count <= 0) {
    return 0
  }
  const price = totalCost.value / count
  return Number.isFinite(price) ? price : 0
})

const priceDetails = computed<PriceDetail[]>(() => {
  const details: PriceDetail[] = []

  if (displayBillingMode.value === BILLING_MODE_TOKEN) {
    for (const part of tokenParts.value) {
      details.push({
        label: `${part.label}${t('usage.priceSuffix')}`,
        value: formatTokenPrice(part.cost, part.tokens),
        className: part.className
      })
    }
    if (residualCost.value > 0) {
      details.push({
        label: t('usage.otherBillingItems'),
        value: formatMoney(residualCost.value),
        className: 'text-gray-100'
      })
    }
  } else if (displayBillingMode.value === BILLING_MODE_IMAGE) {
    details.push({
      label: t('usage.imageUnitPrice'),
      value: `${formatMoney(unitPrice.value)} / ${t('usage.imageUnit')}`,
      className: 'text-sky-300'
    })
    details.push({
      label: t('usage.imageBillingSize'),
      value: row.value.image_size || t('usage.imageSizeNotRecorded'),
      className: 'text-white'
    })
  } else if (displayBillingMode.value === BILLING_MODE_PER_REQUEST) {
    details.push({
      label: t('usage.unitPrice'),
      value: formatMoney(unitPrice.value),
      className: 'text-sky-300'
    })
  }

  details.push({
    label: t('usage.groupMultiplier'),
    value: formatRate(groupRate.value),
    className: 'text-blue-300'
  })

  return details
})

const tokenFormulaLine = computed(() => {
  const segments = tokenParts.value.map((part) => {
    const price = pricePerMillion(part.cost, part.tokens)
    if (price == null) {
      return `${part.label} ${formatMoney(part.cost)}`
    }
    return `${part.label} ${formatTokenCount(part.tokens)} tokens / 1M tokens * ${formatMoney(price)}`
  })

  if (residualCost.value !== 0) {
    segments.push(`${t('usage.otherBillingItems')} ${formatMoney(residualCost.value)}`)
  }

  if (segments.length === 0) {
    return `${t('usage.standardSubtotal')} ${formatMoney(totalCost.value)}`
  }

  return segments.join(' + ')
})

const formulaLines = computed(() => {
  if (displayBillingMode.value === BILLING_MODE_IMAGE) {
    return [
      `${t('usage.imageCount')} ${formatTokenCount(unitCount.value)} * ${t('usage.imageUnitPrice')} ${formatMoney(unitPrice.value)}`,
      `${t('usage.standardSubtotal')} ${formatMoney(totalCost.value)} * ${t('usage.groupMultiplier')} ${formatRate(groupRate.value)} = ${formatMoney(actualCost.value)}`
    ]
  }

  if (displayBillingMode.value === BILLING_MODE_PER_REQUEST) {
    return [
      `${t('usage.requestCount')} 1 * ${t('usage.unitPrice')} ${formatMoney(unitPrice.value)}`,
      `${t('usage.standardSubtotal')} ${formatMoney(totalCost.value)} * ${t('usage.groupMultiplier')} ${formatRate(groupRate.value)} = ${formatMoney(actualCost.value)}`
    ]
  }

  return [
    tokenFormulaLine.value,
    `${t('usage.standardSubtotal')} ${formatMoney(totalCost.value)} * ${t('usage.groupMultiplier')} ${formatRate(groupRate.value)} = ${formatMoney(actualCost.value)}`
  ]
})

const adminRow = computed(() => row.value as AdminUsageLog)
const accountRate = computed(() => {
  return isFiniteNumber(adminRow.value.account_rate_multiplier) ? adminRow.value.account_rate_multiplier : null
})
const accountBaseCost = computed(() => {
  return isFiniteNumber(adminRow.value.account_stats_cost)
    ? adminRow.value.account_stats_cost
    : totalCost.value
})
const accountBilledCost = computed(() => accountBaseCost.value * (accountRate.value ?? 1))
const showAccountBilling = computed(() => props.showAccount && accountRate.value != null)
const accountFormula = computed(() => {
  return `${t('usage.accountCost')} ${formatMoney(accountBaseCost.value)} * ${t('usage.accountMultiplier')} ${formatRate(accountRate.value ?? 1)} = ${formatMoney(accountBilledCost.value)}`
})

const showProcess = (event: MouseEvent) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  const tooltipWidth = Math.min(430, Math.max(280, window.innerWidth - 24))
  const rightX = rect.right + 8

  if (rightX + tooltipWidth > window.innerWidth - 12) {
    placement.value = 'left'
    position.value.x = Math.max(12, rect.left - tooltipWidth - 8)
  } else {
    placement.value = 'right'
    position.value.x = rightX
  }

  position.value.y = rect.top + rect.height / 2
  visible.value = true
}

const hideProcess = () => {
  visible.value = false
}
</script>
