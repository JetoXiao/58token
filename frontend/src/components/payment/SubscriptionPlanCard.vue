<template>
  <article
    :class="[
      'group relative flex h-full flex-col rounded-2xl border bg-white p-5 text-left shadow-sm transition-all dark:bg-dark-900',
      'hover:-translate-y-0.5 hover:shadow-lg',
      isSelected
        ? 'border-primary-400 bg-primary-50/70 ring-4 ring-primary-100 dark:border-primary-500 dark:bg-primary-900/20 dark:ring-primary-950'
        : 'border-gray-200 hover:border-primary-200 dark:border-dark-700 dark:hover:border-primary-800',
    ]"
  >
    <div class="flex flex-1 flex-col">
      <div class="mb-5 flex items-start justify-between gap-3">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <span :class="['rounded-full px-2.5 py-1 text-[11px] font-bold', badgeLightClass]">
              {{ pLabel }}
            </span>
            <span
              v-if="isRenewal"
              class="rounded-full bg-accent-100 px-2.5 py-1 text-[11px] font-bold text-accent-700 dark:bg-accent-900/40 dark:text-accent-300"
            >
              {{ t('payment.renewNow') }}
            </span>
            <span
              v-if="limitedOfferActive"
              class="rounded-full bg-accent-100 px-2.5 py-1 text-[11px] font-bold text-accent-700 dark:bg-accent-900/30 dark:text-accent-300"
            >
              {{ t('payment.planCard.limitedOffer') }}
            </span>
          </div>
          <h3 class="mt-4 text-lg font-black leading-tight text-gray-950 dark:text-white">{{ plan.name }}</h3>
          <p v-if="plan.description" class="mt-2 text-sm leading-6 text-gray-500 line-clamp-2 dark:text-gray-400">
            {{ plan.description }}
          </p>
        </div>
        <span v-if="hasSavings" :class="['shrink-0 rounded-full px-2.5 py-1 text-[11px] font-bold', discountClass]">
          {{ discountText }}
        </span>
      </div>

      <div class="mb-5">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.planCard.currentPrice') }}</p>
        <div class="mt-1 flex flex-wrap items-end justify-between gap-3">
          <div class="flex flex-wrap items-baseline gap-x-1.5 gap-y-1">
            <span class="text-sm font-bold text-gray-400 dark:text-gray-500">$</span>
            <span :class="['text-4xl font-black tracking-tight', textClass]">{{ priceDisplay }}</span>
            <span class="pb-1 text-sm text-gray-500 dark:text-gray-400">/ {{ validitySuffix }}</span>
          </div>
          <div v-if="hasOriginalPrice || limitedOfferActive" class="text-right">
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ limitedOfferActive ? t('payment.planCard.restorePrice') : t('payment.planCard.originalPrice') }}</p>
            <p class="text-sm font-semibold text-gray-400 line-through dark:text-gray-500">
              {{ formatCurrency(limitedOfferActive ? regularPrice : plan.original_price) }}
            </p>
            <p v-if="hasSavings" class="mt-1 text-xs font-bold text-primary-700 dark:text-primary-300">
              {{ savingsText }}
            </p>
          </div>
        </div>
        <div
          v-if="limitedOfferActive"
          class="mt-3 rounded-xl border border-accent-200 bg-accent-50 px-3 py-2.5 text-xs font-semibold text-accent-700 dark:border-accent-900/50 dark:bg-accent-900/20 dark:text-accent-200"
        >
          {{ t('payment.planCard.limitedOfferUntil', { time: limitedOfferEndText, price: formatCurrency(regularPrice) }) }}
        </div>
      </div>

      <div v-if="quotaItems.length > 0" class="mb-3 grid gap-3 sm:grid-cols-3">
        <div
          v-for="item in quotaItems"
          :key="item.key"
          class="rounded-xl border border-gray-200 bg-white/80 p-3 dark:border-dark-700 dark:bg-dark-800/80"
        >
          <div class="flex items-center gap-1.5 text-xs font-medium text-gray-500 dark:text-gray-400">
            <Icon :name="item.icon" size="xs" :class="iconClass" :stroke-width="2" />
            <span>{{ item.label }}</span>
          </div>
          <p class="mt-2 text-lg font-black text-gray-950 dark:text-white">{{ item.value }}</p>
        </div>
      </div>

      <div v-else class="mb-3 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2.5 dark:border-dark-700 dark:bg-dark-800/80">
        <div class="flex items-center justify-between gap-3 text-sm">
          <span class="text-gray-500 dark:text-gray-400">{{ t('payment.planCard.quota') }}</span>
          <span class="font-semibold text-gray-900 dark:text-white">{{ t('payment.planCard.unlimited') }}</span>
        </div>
      </div>

      <div
        v-if="workdayFriendly"
        class="mb-4 rounded-xl border border-primary-200 bg-primary-50 px-3 py-2.5 dark:border-primary-900/50 dark:bg-primary-900/20"
      >
        <div class="flex items-center gap-2 text-sm font-bold text-primary-800 dark:text-primary-200">
          <Icon name="calendar" size="sm" :stroke-width="2" />
          <span>{{ t('payment.planCard.workdayFriendly') }}</span>
        </div>
        <p class="mt-1 text-xs leading-relaxed text-primary-700/90 dark:text-primary-200/80">
          {{ t('payment.planCard.workdayFriendlyDesc', { daily: formatCurrency(plan.daily_limit_usd), weekly: formatCurrency(plan.weekly_limit_usd) }) }}
        </p>
      </div>

      <div class="mb-4 flex flex-wrap items-center gap-2 text-xs">
        <span class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-1 font-semibold text-gray-600 dark:bg-dark-800 dark:text-gray-300">
          <Icon name="bolt" size="xs" :class="iconClass" :stroke-width="2" />
          {{ t('payment.planCard.rate') }} {{ rateDisplay }}
        </span>
        <span class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-1 font-semibold text-gray-600 dark:bg-dark-800 dark:text-gray-300">
          <Icon name="clock" size="xs" :class="iconClass" :stroke-width="2" />
          {{ t('payment.planCard.validFor', { duration: validityDuration }) }}
        </span>
        <template v-if="modelScopeLabels.length > 0">
          <span
            v-for="scope in modelScopeLabels"
            :key="scope"
            class="rounded-full bg-gray-100 px-2.5 py-1 font-semibold text-gray-600 dark:bg-dark-800 dark:text-gray-300"
          >
            {{ scope }}
          </span>
        </template>
      </div>

      <div v-if="displayFeatures.length > 0" class="mb-4 space-y-2 border-t border-gray-100 pt-4 dark:border-dark-700">
        <div v-for="feature in displayFeatures" :key="feature" class="flex items-start gap-2">
          <Icon name="checkCircle" size="sm" :class="['mt-0.5 flex-shrink-0', iconClass]" :stroke-width="2" />
          <span class="text-sm leading-relaxed text-gray-600 dark:text-gray-300">{{ feature }}</span>
        </div>
      </div>

      <div v-if="hasSavings" class="mb-4 rounded-xl bg-primary-50 px-3 py-2 text-xs font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">
        <div class="flex items-center gap-2">
          <Icon name="sparkles" size="sm" :stroke-width="2" />
          <span>{{ t('payment.planCard.dealHint', { original: formatCurrency(savingsComparePrice), price: formatCurrency(plan.price) }) }}</span>
        </div>
      </div>

      <div class="flex-1" />

      <button
        type="button"
        :class="['w-full rounded-xl py-3 text-sm font-bold transition-all active:scale-[0.98]', isSelected ? 'bg-gray-900 text-white hover:bg-gray-800 dark:bg-white dark:text-gray-950 dark:hover:bg-gray-100' : btnClass]"
        @click="emit('select', plan)"
      >
        <span class="inline-flex items-center justify-center gap-2">
          <Icon name="creditCard" size="sm" :stroke-width="2" />
          {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
        </span>
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import {
  platformBadgeLightClass,
  platformTextClass,
  platformIconClass,
  platformButtonClass,
  platformDiscountClass,
  platformLabel,
} from '@/utils/platformColors'

type QuotaIcon = 'calendar' | 'clock' | 'chart'

interface QuotaItem {
  key: string
  label: string
  value: string
  icon: QuotaIcon
}

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[]; selected?: boolean }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const isSelected = computed(() => props.selected === true)
const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

const badgeLightClass = computed(() => platformBadgeLightClass(platform.value))
const textClass = computed(() => platformTextClass(platform.value))
const iconClass = computed(() => platformIconClass(platform.value))
const btnClass = computed(() => platformButtonClass(platform.value))
const discountClass = computed(() => platformDiscountClass(platform.value))
const pLabel = computed(() => platformLabel(platform.value))

function normalizeUnit(unit: string | undefined): string {
  const value = (unit || 'day').toLowerCase()
  if (value === 'days') return 'day'
  if (value === 'weeks') return 'week'
  if (value === 'months') return 'month'
  if (value === 'years') return 'year'
  return value
}

function formatMoney(value: number | null | undefined): string {
  const amount = Number(value ?? 0)
  if (!Number.isFinite(amount)) return '0'
  return amount.toFixed(2).replace(/\.?0+$/, '')
}

function formatCurrency(value: number | null | undefined): string {
  return `$${formatMoney(value)}`
}

const priceDisplay = computed(() => formatMoney(props.plan.price))
const regularPrice = computed(() => props.plan.regular_price || props.plan.price)
const limitedOfferActive = computed(() => {
  if (props.plan.limited_offer_active != null) return props.plan.limited_offer_active
  if (!props.plan.limited_offer_price || !props.plan.limited_offer_expires_at) return false
  return new Date(props.plan.limited_offer_expires_at).getTime() > Date.now()
})
const limitedOfferEndText = computed(() => {
  const value = props.plan.limited_offer_expires_at
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
})
const hasOriginalPrice = computed(() => (props.plan.original_price ?? 0) > 0)
const savingsComparePrice = computed(() => {
  const originalPrice = props.plan.original_price ?? 0
  if (originalPrice > props.plan.price) return originalPrice
  if (limitedOfferActive.value && regularPrice.value > props.plan.price) return regularPrice.value
  return 0
})
const hasSavings = computed(() =>
  savingsComparePrice.value > props.plan.price
)

const discountText = computed(() => {
  if (!hasSavings.value) return ''
  const pct = Math.round((1 - props.plan.price / savingsComparePrice.value) * 100)
  return pct > 0 ? t('payment.planCard.discountPercent', { percent: pct }) : ''
})

const savingsText = computed(() => {
  if (!hasSavings.value) return ''
  return t('payment.planCard.saveAmount', {
    amount: formatCurrency(savingsComparePrice.value - props.plan.price),
  })
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return `x${Number(rate.toPrecision(10))}`
})

const quotaItems = computed<QuotaItem[]>(() => {
  const items: QuotaItem[] = []
  if (props.plan.daily_limit_usd != null) {
    items.push({
      key: 'daily',
      label: t('payment.planCard.dailyLimit'),
      value: formatCurrency(props.plan.daily_limit_usd),
      icon: 'calendar',
    })
  }
  if (props.plan.weekly_limit_usd != null) {
    items.push({
      key: 'weekly',
      label: t('payment.planCard.weeklyLimit'),
      value: formatCurrency(props.plan.weekly_limit_usd),
      icon: 'clock',
    })
  }
  if (props.plan.monthly_limit_usd != null) {
    items.push({
      key: 'monthly',
      label: t('payment.planCard.monthlyLimit'),
      value: formatCurrency(props.plan.monthly_limit_usd),
      icon: 'chart',
    })
  }
  return items
})

const workdayFriendly = computed(() => {
  const daily = props.plan.daily_limit_usd
  const weekly = props.plan.weekly_limit_usd
  if (daily == null || weekly == null || daily <= 0 || weekly <= 0) return false
  return true
})

const MODEL_SCOPE_LABELS: Record<string, string> = {
  claude: 'Claude',
  gemini_text: 'Gemini',
  gemini_image: 'Imagen',
}

const modelScopeLabels = computed(() => {
  if (platform.value !== 'antigravity') return []
  const scopes = props.plan.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => MODEL_SCOPE_LABELS[s] || s)
})

const displayFeatures = computed(() => props.plan.features || [])

const validitySuffix = computed(() => {
  const u = normalizeUnit(props.plan.validity_unit)
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${props.plan.validity_days}${t('payment.days')}`
})

const validityDuration = computed(() => {
  const days = props.plan.validity_days
  const unit = normalizeUnit(props.plan.validity_unit)
  if (unit === 'month') return t('payment.planCard.durationMonths', { count: days })
  if (unit === 'week') return t('payment.planCard.durationWeeks', { count: days })
  if (unit === 'year') return t('payment.planCard.durationYears', { count: days })
  return t('payment.planCard.durationDays', { count: days })
})
</script>
