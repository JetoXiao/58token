<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <!-- Current Balance Card -->
      <div class="card overflow-hidden">
        <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
          <div
            class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm"
          >
            <Icon name="creditCard" size="xl" class="text-white" />
          </div>
          <p class="text-sm font-medium text-primary-100">{{ t('redeem.currentBalance') }}</p>
          <p class="mt-2 text-4xl font-bold text-white">
            {{ formatCurrency(totalBalance) }}
          </p>
          <div class="mt-2 flex justify-center gap-4 text-sm text-primary-100">
            <span>{{ t('common.balance') }} {{ formatCurrency(paidBalance) }}</span>
            <span>{{ t('common.freeQuota') }} {{ formatCurrency(freeQuotaAmount) }}</span>
          </div>
          <p class="mt-2 text-sm text-primary-100">
            {{ t('redeem.concurrency') }}: {{ user?.concurrency || 0 }} {{ t('redeem.requests') }}
          </p>
        </div>
      </div>

      <!-- Redeem Forms -->
      <div class="grid gap-6 lg:grid-cols-2">
      <div class="card">
        <div class="p-6">
          <form @submit.prevent="handleRedeem" class="space-y-5">
            <div>
              <label for="code" class="input-label">
                {{ t('redeem.redeemCodeLabel') }}
              </label>
              <div class="relative mt-1">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
                  <Icon name="gift" size="md" class="text-gray-400 dark:text-dark-500" />
                </div>
                <input
                  id="code"
                  v-model="redeemCode"
                  type="text"
                  required
                  :placeholder="t('redeem.redeemCodePlaceholder')"
                  :disabled="submitting"
                  class="input py-3 pl-12 text-lg"
                />
              </div>
              <p class="input-hint">
                {{ t('redeem.redeemCodeHint') }}
              </p>
            </div>

            <button
              type="submit"
              :disabled="!redeemCode || submitting"
              class="btn btn-primary w-full py-3"
            >
              <svg
                v-if="submitting"
                class="-ml-1 mr-2 h-5 w-5 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <Icon v-else name="checkCircle" size="md" class="mr-2" />
              {{ submitting ? t('redeem.redeeming') : t('redeem.redeemButton') }}
            </button>
          </form>
        </div>
      </div>

      <div class="card">
        <div class="p-6">
          <form @submit.prevent="handleTrialRedeem" class="space-y-5">
            <div>
              <label for="trial-code" class="input-label">
                {{ t('redeem.trialCardLabel') }}
              </label>
              <div class="relative mt-1">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
                  <Icon name="beaker" size="md" class="text-gray-400 dark:text-dark-500" />
                </div>
                <input
                  id="trial-code"
                  v-model="trialCode"
                  type="text"
                  required
                  :placeholder="t('redeem.trialCardPlaceholder')"
                  :disabled="submittingTrial"
                  class="input py-3 pl-12 text-lg"
                />
              </div>
              <p class="input-hint">
                {{ t('redeem.trialCardHint') }}
              </p>
            </div>

            <button
              type="submit"
              :disabled="!trialCode || submittingTrial"
              class="btn btn-primary w-full py-3"
            >
              <svg
                v-if="submittingTrial"
                class="-ml-1 mr-2 h-5 w-5 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <Icon v-else name="checkCircle" size="md" class="mr-2" />
              {{ submittingTrial ? t('redeem.redeeming') : t('redeem.trialCardButton') }}
            </button>
          </form>
        </div>
      </div>
      </div>

      <!-- Success Message -->
      <transition name="fade">
        <div
          v-if="redeemResult"
          class="card border-emerald-200 bg-emerald-50 dark:border-emerald-800/50 dark:bg-emerald-900/20"
        >
          <div class="p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-900/30"
              >
                <Icon name="checkCircle" size="md" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-emerald-800 dark:text-emerald-300">
                  {{ t('redeem.redeemSuccess') }}
                </h3>
                <div class="mt-2 text-sm text-emerald-700 dark:text-emerald-400">
                  <p>{{ redeemResult.message }}</p>
                  <div class="mt-3 space-y-1">
                    <p v-if="redeemResult.type === 'balance'" class="font-medium">
                      {{ t('redeem.added') }}: ${{ redeemResult.value.toFixed(2) }}
                    </p>
                    <p v-else-if="redeemResult.type === 'concurrency'" class="font-medium">
                      {{ t('redeem.added') }}: {{ redeemResult.value }}
                      {{ t('redeem.concurrentRequests') }}
                    </p>
                    <p v-else-if="redeemResult.type === 'subscription'" class="font-medium">
                      {{ t('redeem.subscriptionAssigned') }}
                      <span v-if="redeemResult.group_name"> - {{ redeemResult.group_name }}</span>
                      <span v-if="redeemResult.validity_days">
                        ({{
                          t('redeem.subscriptionDays', { days: redeemResult.validity_days })
                        }})</span
                      >
                    </p>
                    <p v-if="redeemResult.new_balance !== undefined">
                      {{ t('redeem.newBalance') }}:
                      <span class="font-semibold">${{ redeemResult.new_balance.toFixed(2) }}</span>
                    </p>
                    <p v-if="redeemResult.new_concurrency !== undefined">
                      {{ t('redeem.newConcurrency') }}:
                      <span class="font-semibold"
                        >{{ redeemResult.new_concurrency }} {{ t('redeem.requests') }}</span
                      >
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <transition name="fade">
        <div
          v-if="trialResult"
          class="card border-blue-200 bg-blue-50 dark:border-blue-800/50 dark:bg-blue-900/20"
        >
          <div class="p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-blue-100 dark:bg-blue-900/30"
              >
                <Icon name="beaker" size="md" class="text-blue-600 dark:text-blue-400" />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-blue-800 dark:text-blue-300">
                  {{ t('redeem.trialCardSuccess') }}
                </h3>
                <div class="mt-2 space-y-1 text-sm text-blue-700 dark:text-blue-400">
                  <p class="font-medium">
                    {{ t('redeem.added') }}: ${{ trialResult.free_quota.amount.toFixed(2) }}
                  </p>
                  <p>
                    {{ t('redeem.trialCardRemaining') }}:
                    ${{ trialResult.free_quota.remaining_amount.toFixed(2) }}
                  </p>
                  <p>
                    {{ t('redeem.trialCardAllowedGroups') }}:
                    {{ formatGroupNames(trialResult.free_quota.allowed_group_ids) }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Error Message -->
      <transition name="fade">
        <div
          v-if="errorMessage"
          class="card border-red-200 bg-red-50 dark:border-red-800/50 dark:bg-red-900/20"
        >
          <div class="p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-red-100 dark:bg-red-900/30"
              >
                <Icon
                  name="exclamationCircle"
                  size="md"
                  class="text-red-600 dark:text-red-400"
                />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-red-800 dark:text-red-300">
                  {{ t('redeem.redeemFailed') }}
                </h3>
                <p class="mt-2 text-sm text-red-700 dark:text-red-400">
                  {{ errorMessage }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Information Card -->
      <div
        class="card border-primary-200 bg-primary-50 dark:border-primary-800/50 dark:bg-primary-900/20"
      >
        <div class="p-6">
          <div class="flex items-start gap-4">
            <div
              class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30"
            >
              <Icon name="infoCircle" size="md" class="text-primary-600 dark:text-primary-400" />
            </div>
            <div class="flex-1">
              <h3 class="text-sm font-semibold text-primary-800 dark:text-primary-300">
                {{ t('redeem.aboutCodes') }}
              </h3>
              <ul
                class="mt-2 list-inside list-disc space-y-1 text-sm text-primary-700 dark:text-primary-400"
              >
                <li>{{ t('redeem.codeRule1') }}</li>
                <li>{{ t('redeem.codeRule2') }}</li>
                <li>
                  {{ t('redeem.codeRule3') }}
                  <span
                    v-if="contactInfo"
                    class="ml-1.5 inline-flex items-center rounded-md bg-primary-200/50 px-2 py-0.5 text-xs font-medium text-primary-800 dark:bg-primary-800/40 dark:text-primary-200"
                  >
                    {{ contactInfo }}
                  </span>
                </li>
                <li>{{ t('redeem.codeRule4') }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Activity -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('redeem.recentActivity') }}
          </h2>
        </div>
        <div class="p-6">
          <!-- Loading State -->
          <div v-if="loadingHistory" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
          </div>

          <!-- History List -->
          <div v-else-if="history.length > 0" class="space-y-3">
            <div
              v-for="item in history"
              :key="item.id"
              class="flex items-center justify-between rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
            >
              <div class="flex items-center gap-4">
                <div
                  :class="[
                    'flex h-10 w-10 items-center justify-center rounded-xl',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'bg-emerald-100 dark:bg-emerald-900/30'
                        : 'bg-red-100 dark:bg-red-900/30'
                      : isSubscriptionType(item.type)
                        ? 'bg-purple-100 dark:bg-purple-900/30'
                        : item.value >= 0
                          ? 'bg-blue-100 dark:bg-blue-900/30'
                          : 'bg-orange-100 dark:bg-orange-900/30'
                  ]"
                >
                  <!-- 余额类型图标 -->
                  <Icon
                    v-if="isBalanceType(item.type)"
                    name="dollar"
                    size="md"
                    :class="
                      item.value >= 0
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-red-600 dark:text-red-400'
                    "
                  />
                  <!-- 订阅类型图标 -->
                  <Icon
                    v-else-if="isSubscriptionType(item.type)"
                    name="badge"
                    size="md"
                    class="text-purple-600 dark:text-purple-400"
                  />
                  <!-- 并发类型图标 -->
                  <Icon
                    v-else
                    name="bolt"
                    size="md"
                    :class="
                      item.value >= 0
                        ? 'text-blue-600 dark:text-blue-400'
                        : 'text-orange-600 dark:text-orange-400'
                    "
                  />
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ getHistoryItemTitle(item) }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-dark-400">
                    {{ formatDateTime(item.used_at) }}
                  </p>
                </div>
              </div>
              <div class="text-right">
                <p
                  :class="[
                    'text-sm font-semibold',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-red-600 dark:text-red-400'
                      : isSubscriptionType(item.type)
                        ? 'text-purple-600 dark:text-purple-400'
                        : item.value >= 0
                          ? 'text-blue-600 dark:text-blue-400'
                          : 'text-orange-600 dark:text-orange-400'
                  ]"
                >
                  {{ formatHistoryValue(item) }}
                </p>
                <p
                  v-if="!isAdminAdjustment(item.type)"
                  class="font-mono text-xs text-gray-400 dark:text-dark-500"
                >
                  {{ item.code.slice(0, 8) }}...
                </p>
                <p v-else class="text-xs text-gray-400 dark:text-dark-500">
                  {{ t('redeem.adminAdjustment') }}
                </p>
                <!-- Display notes for admin adjustments -->
                <p
                  v-if="item.notes"
                  class="mt-1 text-xs text-gray-500 dark:text-dark-400 italic max-w-[200px] truncate"
                  :title="item.notes"
                >
                  {{ item.notes }}
                </p>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="empty-state py-8">
            <div
              class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800"
            >
              <Icon name="clock" size="xl" class="text-gray-400 dark:text-dark-500" />
            </div>
            <p class="text-sm text-gray-500 dark:text-dark-400">
              {{ t('redeem.historyWillAppear') }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useFreeQuotaStore } from '@/stores/freeQuota'
import { redeemAPI, freeQuotaAPI, authAPI, userGroupsAPI, type RedeemHistoryItem } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { Group, RedeemTrialCardResponse } from '@/types'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()
const freeQuotaStore = useFreeQuotaStore()

const user = computed(() => authStore.user)
const paidBalance = computed(() => freeQuotaStore.summary.balance_amount || user.value?.balance || 0)
const freeQuotaAmount = computed(() => freeQuotaStore.summary.free_quota_amount || 0)
const totalBalance = computed(() => freeQuotaStore.summary.total_amount || paidBalance.value + freeQuotaAmount.value)

const redeemCode = ref('')
const submitting = ref(false)
const trialCode = ref('')
const submittingTrial = ref(false)
const redeemResult = ref<{
  message: string
  type: string
  value: number
  new_balance?: number
  new_concurrency?: number
  group_name?: string
  validity_days?: number
} | null>(null)
const trialResult = ref<RedeemTrialCardResponse | null>(null)
const errorMessage = ref('')

// History data
const history = ref<RedeemHistoryItem[]>([])
const loadingHistory = ref(false)
const contactInfo = ref('')
const availableGroups = ref<Group[]>([])
const groupNameMap = computed(() => {
  const map = new Map<number, string>()
  for (const group of availableGroups.value) {
    map.set(group.id, group.name)
  }
  return map
})

const trialCardErrorMessages: Record<string, string> = {
  TRIAL_CARD_NOT_FOUND: '\u4f53\u9a8c\u5361\u4e0d\u5b58\u5728',
  'trial card not found': '\u4f53\u9a8c\u5361\u4e0d\u5b58\u5728',
  TRIAL_CARD_INACTIVE: '\u4f53\u9a8c\u5361\u5df2\u7981\u7528',
  'trial card is inactive': '\u4f53\u9a8c\u5361\u5df2\u7981\u7528',
  TRIAL_CARD_EXPIRED: '\u4f53\u9a8c\u5361\u5df2\u8fc7\u671f',
  'trial card has expired': '\u4f53\u9a8c\u5361\u5df2\u8fc7\u671f',
  TRIAL_CARD_EXHAUSTED: '\u4f53\u9a8c\u5361\u53ef\u5151\u6362\u6b21\u6570\u5df2\u7528\u5b8c',
  'trial card has no remaining redemptions': '\u4f53\u9a8c\u5361\u53ef\u5151\u6362\u6b21\u6570\u5df2\u7528\u5b8c',
  TRIAL_CARD_USER_LIMIT: '\u5df2\u8fbe\u5230\u8be5\u4f53\u9a8c\u5361\u7684\u5151\u6362\u6b21\u6570\u9650\u5236',
  'trial card redemption limit reached': '\u5df2\u8fbe\u5230\u8be5\u4f53\u9a8c\u5361\u7684\u5151\u6362\u6b21\u6570\u9650\u5236'
}

const getTrialCardErrorMessage = (error: any) => {
  const candidates = [
    error?.code,
    error?.reason,
    error?.message,
    error?.response?.data?.code,
    error?.response?.data?.reason,
    error?.response?.data?.message,
    error?.response?.data?.detail
  ]
  for (const candidate of candidates) {
    const key = String(candidate || '').trim()
    if (key && trialCardErrorMessages[key]) {
      return trialCardErrorMessages[key]
    }
  }
  return error?.message || error?.response?.data?.detail || t('redeem.failedToRedeemTrialCard')
}

// Helper functions for history display
const isBalanceType = (type: string) => {
  return type === 'balance' || type === 'admin_balance'
}

const isSubscriptionType = (type: string) => {
  return type === 'subscription'
}

const isAdminAdjustment = (type: string) => {
  return type === 'admin_balance' || type === 'admin_concurrency'
}

const getHistoryItemTitle = (item: RedeemHistoryItem) => {
  if (item.type === 'balance') {
    return t('redeem.balanceAddedRedeem')
  } else if (item.type === 'admin_balance') {
    return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
  } else if (item.type === 'concurrency') {
    return t('redeem.concurrencyAddedRedeem')
  } else if (item.type === 'admin_concurrency') {
    return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
  } else if (item.type === 'subscription') {
    return t('redeem.subscriptionAssigned')
  }
  return t('common.unknown')
}

const formatHistoryValue = (item: RedeemHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  } else if (isSubscriptionType(item.type)) {
    // 订阅类型显示有效天数和分组名称
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}${t('redeem.days')} - ${groupName}` : `${days}${t('redeem.days')}`
  } else {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${item.value} ${t('redeem.requests')}`
  }
}

const fetchHistory = async () => {
  loadingHistory.value = true
  try {
    history.value = await redeemAPI.getHistory()
  } catch (error) {
    console.error('Failed to fetch history:', error)
  } finally {
    loadingHistory.value = false
  }
}

const fetchAvailableGroups = async () => {
  try {
    availableGroups.value = await userGroupsAPI.getAvailable()
  } catch (error) {
    console.error('Failed to fetch available groups:', error)
  }
}

const formatGroupNames = (ids: number[]) => {
  if (!ids.length) {
    return t('common.none')
  }
  return ids.map((id) => groupNameMap.value.get(id) || `#${id}`).join(', ')
}

const formatCurrency = (value: number) => {
  return `$${Number(value || 0).toFixed(2)}`
}

const handleRedeem = async () => {
  if (!redeemCode.value.trim()) {
    appStore.showError(t('redeem.pleaseEnterCode'))
    return
  }

  submitting.value = true
  errorMessage.value = ''
  redeemResult.value = null

  try {
    const result = await redeemAPI.redeem(redeemCode.value.trim())

    redeemResult.value = result

    // Refresh user data to get updated balance/concurrency
    await authStore.refreshUser()
    if (authStore.user?.balance != null) {
      freeQuotaStore.setBalanceAmount(authStore.user.balance)
    }
    await freeQuotaStore.fetchSummary(true).catch((error) => {
      console.error('Failed to refresh free quota summary after redeem:', error)
    })

    // If subscription type, immediately refresh subscription status
    if (result.type === 'subscription') {
      try {
        await subscriptionStore.fetchActiveSubscriptions(true) // force refresh
      } catch (error) {
        console.error('Failed to refresh subscriptions after redeem:', error)
        appStore.showWarning(t('redeem.subscriptionRefreshFailed'))
      }
    }

    // Clear the input
    redeemCode.value = ''

    // Refresh history
    await fetchHistory()

    // Show success toast
    appStore.showSuccess(t('redeem.codeRedeemSuccess'))
  } catch (error: any) {
    errorMessage.value = error.response?.data?.detail || t('redeem.failedToRedeem')

    appStore.showError(t('redeem.redeemFailed'))
  } finally {
    submitting.value = false
  }
}

const handleTrialRedeem = async () => {
  if (!trialCode.value.trim()) {
    appStore.showError(t('redeem.pleaseEnterTrialCode'))
    return
  }

  submittingTrial.value = true
  errorMessage.value = ''
  trialResult.value = null

  try {
    const result = await freeQuotaAPI.redeemTrialCard(trialCode.value.trim())
    trialResult.value = result
    trialCode.value = ''
    freeQuotaStore.applyTrialGrant(result.free_quota.remaining_amount, authStore.user?.balance)
    await authStore.refreshUser()
    if (authStore.user?.balance != null) {
      freeQuotaStore.setBalanceAmount(authStore.user.balance)
    }
    await freeQuotaStore.fetchSummary(true)
    await fetchAvailableGroups()
    appStore.showSuccess(t('redeem.trialCardRedeemSuccess'))
  } catch (error: any) {
    errorMessage.value = getTrialCardErrorMessage(error)
    appStore.showError(t('redeem.redeemFailed'))
  } finally {
    submittingTrial.value = false
  }
}

onMounted(async () => {
  fetchHistory()
  fetchAvailableGroups()
  freeQuotaStore.fetchSummary().catch((error) => {
    console.error('Failed to fetch free quota summary:', error)
  })
  try {
    const settings = await authAPI.getPublicSettings()
    contactInfo.value = settings.contact_info || ''
  } catch (error) {
    console.error('Failed to load contact info:', error)
  }
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
