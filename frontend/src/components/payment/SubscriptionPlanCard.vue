<template>
  <div
    :class="[
      'group relative flex h-full flex-col overflow-hidden rounded-2xl border bg-white shadow-sm transition-all dark:bg-dark-800',
      'hover:-translate-y-0.5 hover:shadow-xl',
      borderClass,
    ]"
  >
    <div :class="['h-1.5', accentClass]" />

    <div class="flex flex-1 flex-col p-5">
      <div class="mb-4 flex flex-wrap items-center gap-2">
        <span :class="['rounded-full px-2.5 py-1 text-[11px] font-semibold', badgeLightClass]">
          {{ pLabel }}
        </span>
        <span v-if="hasSavings" :class="['rounded-full px-2.5 py-1 text-[11px] font-semibold', discountClass]">
          {{ discountText }}
        </span>
        <span
          v-if="isRenewal"
          class="rounded-full bg-amber-100 px-2.5 py-1 text-[11px] font-semibold text-amber-700 dark:bg-amber-900/40 dark:text-amber-300"
        >
          {{ t('payment.renewNow') }}
        </span>
      </div>

      <div class="mb-4">
        <h3 class="text-lg font-bold leading-tight text-gray-900 dark:text-white">{{ plan.name }}</h3>
        <p v-if="plan.description" class="mt-1 text-sm leading-relaxed text-gray-500 line-clamp-2 dark:text-dark-400">
          {{ plan.description }}
        </p>
      </div>

      <div class="mb-4 rounded-xl border border-gray-100 bg-gray-50/80 p-4 dark:border-dark-700 dark:bg-dark-700/40">
        <div class="flex items-end justify-between gap-3">
          <div class="min-w-0">
            <p class="text-xs font-medium uppercase text-gray-400 dark:text-dark-400">
              {{ t('payment.planCard.currentPrice') }}
            </p>
            <div class="mt-1 flex flex-wrap items-baseline gap-x-1.5 gap-y-1">
              <span class="text-sm font-bold text-gray-400 dark:text-dark-500">$</span>
              <span :class="['text-4xl font-extrabold leading-none', textClass]">{{ priceDisplay }}</span>
              <span class="text-sm text-gray-500 dark:text-gray-400">/ {{ validitySuffix }}</span>
            </div>
          </div>
          <div v-if="hasOriginalPrice" class="shrink-0 text-right">
            <p class="text-xs text-gray-400 dark:text-dark-500">{{ t('payment.planCard.originalPrice') }}</p>
            <p class="text-sm font-medium text-gray-400 line-through dark:text-dark-500">
              {{ formatCurrency(plan.original_price) }}
            </p>
            <p v-if="hasSavings" :class="['mt-1 inline-flex rounded-full px-2 py-0.5 text-[11px] font-semibold', discountClass]">
              {{ savingsText }}
            </p>
          </div>
        </div>
      </div>

      <div v-if="quotaItems.length > 0" class="mb-3 grid grid-cols-1 gap-2 sm:grid-cols-3">
        <div
          v-for="item in quotaItems"
          :key="item.key"
          class="rounded-xl border border-gray-100 bg-white px-3 py-2.5 dark:border-dark-700 dark:bg-dark-800/80"
        >
          <div class="flex items-center gap-1.5 text-[11px] font-medium text-gray-500 dark:text-dark-400">
            <Icon :name="item.icon" size="xs" :class="iconClass" :stroke-width="2" />
            <span>{{ item.label }}</span>
          </div>
          <p class="mt-1 text-base font-bold text-gray-900 dark:text-white">{{ item.value }}</p>
        </div>
      </div>

      <div v-else class="mb-3 rounded-xl border border-gray-100 bg-gray-50 px-3 py-2.5 dark:border-dark-700 dark:bg-dark-700/40">
        <div class="flex items-center justify-between gap-3 text-sm">
          <span class="text-gray-500 dark:text-dark-400">{{ t('payment.planCard.quota') }}</span>
          <span class="font-semibold text-gray-900 dark:text-white">{{ t('payment.planCard.unlimited') }}</span>
        </div>
      </div>

      <div
        v-if="workdayFriendly"
        class="mb-4 rounded-xl border border-amber-200 bg-amber-50 px-3 py-2.5 dark:border-amber-900/50 dark:bg-amber-900/20"
      >
        <div class="flex items-center gap-2 text-sm font-semibold text-amber-800 dark:text-amber-200">
          <Icon name="calendar" size="sm" :stroke-width="2" />
          <span>{{ t('payment.planCard.workdayFriendly') }}</span>
        </div>
        <p class="mt-1 text-xs leading-relaxed text-amber-700/90 dark:text-amber-200/80">
          {{ t('payment.planCard.workdayFriendlyDesc', { daily: formatCurrency(plan.daily_limit_usd), weekly: formatCurrency(plan.weekly_limit_usd) }) }}
        </p>
      </div>

      <div class="mb-4 flex flex-wrap items-center gap-2 text-xs">
        <span class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-1 font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">
          <Icon name="bolt" size="xs" :class="iconClass" :stroke-width="2" />
          {{ t('payment.planCard.rate') }} {{ rateDisplay }}
        </span>
        <span class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-1 font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">
          <Icon name="clock" size="xs" :class="iconClass" :stroke-width="2" />
          {{ t('payment.planCard.validFor', { duration: validityDuration }) }}
        </span>
        <template v-if="modelScopeLabels.length > 0">
          <span
            v-for="scope in modelScopeLabels"
            :key="scope"
            class="rounded-full bg-gray-100 px-2.5 py-1 font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300"
          >
            {{ scope }}
          </span>
        </template>
      </div>

      <div v-if="displayFeatures.length > 0" class="mb-4 space-y-2">
        <div v-for="feature in displayFeatures" :key="feature" class="flex items-start gap-2">
          <Icon name="checkCircle" size="sm" :class="['mt-0.5 flex-shrink-0', iconClass]" :stroke-width="2" />
          <span class="text-sm leading-relaxed text-gray-600 dark:text-gray-300">{{ feature }}</span>
        </div>
      </div>

      <div v-if="hasSavings" class="mb-4 rounded-xl bg-emerald-50 px-3 py-2 text-xs font-medium text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300">
        <div class="flex items-center gap-2">
          <Icon name="sparkles" size="sm" :stroke-width="2" />
          <span>{{ t('payment.planCard.dealHint', { original: formatCurrency(plan.original_price), price: formatCurrency(plan.price) }) }}</span>
        </div>
      </div>

      <div class="flex-1" />

      <button
        type="button"
        :class="['w-full rounded-xl py-3 text-sm font-semibold transition-all active:scale-[0.98]', btnClass]"
        @click="emit('select', plan)"
      >
        <span class="inline-flex items-center justify-center gap-2">
          <Icon name="creditCard" size="sm" :stroke-width="2" />
          {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import {
  platformAccentBarClass,
  platformBadgeLightClass,
  platformBorderClass,
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

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

const accentClass = computed(() => platformAccentBarClass(platform.value))
const borderClass = computed(() => platformBorderClass(platform.value))
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
const hasOriginalPrice = computed(() => (props.plan.original_price ?? 0) > 0)
const hasSavings = computed(() =>
  hasOriginalPrice.value && (props.plan.original_price ?? 0) > props.plan.price
)

const discountText = computed(() => {
  if (!hasSavings.value) return ''
  const original = props.plan.original_price ?? 0
  const pct = Math.round((1 - props.plan.price / original) * 100)
  return pct > 0 ? t('payment.planCard.discountPercent', { percent: pct }) : ''
})

const savingsText = computed(() => {
  if (!hasSavings.value) return ''
  return t('payment.planCard.saveAmount', {
    amount: formatCurrency((props.plan.original_price ?? 0) - props.plan.price),
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
  return weekly >= daily * 3 && weekly <= daily * 5
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
