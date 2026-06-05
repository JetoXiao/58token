<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <template v-else-if="detail">
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="card p-5">
            <p class="flex items-center gap-1.5 text-sm text-gray-500 dark:text-dark-400">
              <Icon name="badge" size="sm" class="text-primary-500" />
              {{ t('affiliate.stats.partnerLevel') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ partnerLevelLabel(effectivePartnerLevel) }}
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">
              {{ currentTier ? t('affiliate.partner.currentBenefit') : t('affiliate.partner.normalBenefit') }}
            </p>
          </div>
          <div class="card p-5">
            <p class="flex items-center gap-1.5 text-sm text-gray-500 dark:text-dark-400">
              <Icon name="dollar" size="sm" class="text-primary-500" />
              {{ t('affiliate.stats.rewardMode') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-primary-600 dark:text-primary-400">
              {{ t(isPartner ? 'affiliate.stats.partnerRewardMode' : 'affiliate.stats.regularRewardMode') }}
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">
              {{ t(isPartner ? 'affiliate.stats.partnerRebateRateHint' : 'affiliate.stats.rebateRateHint') }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.invitedUsers') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCount(detail.aff_count) }}
            </p>
            <p v-if="nextTier" class="mt-1 text-xs text-gray-400 dark:text-dark-500">
              {{ t('affiliate.partner.nextHint', { count: Math.max(nextTier.required_invitees - detail.aff_count, 0), level: partnerLevelLabel(nextTier.level) }) }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">
              {{ t(isPartner ? 'affiliate.stats.partnerSettlement' : 'affiliate.stats.availableQuota') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
              <template v-if="isPartner">{{ t('affiliate.stats.manualSettlement') }}</template>
              <template v-else>{{ formatCurrency(detail.aff_quota) }}</template>
            </p>
            <p v-if="isPartner" class="mt-1 text-xs text-gray-400 dark:text-dark-500">
              {{ t('affiliate.stats.partnerSettlementHint') }}
            </p>
            <p v-else-if="detail.aff_frozen_quota > 0" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
              {{ t('affiliate.stats.frozenQuota') }}: {{ formatCurrency(detail.aff_frozen_quota) }}
            </p>
          </div>
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
            <div class="max-w-2xl">
              <div class="inline-flex items-center gap-2 rounded-full bg-primary-50 px-3 py-1 text-xs font-medium text-primary-700 dark:bg-primary-900/25 dark:text-primary-300">
                <Icon name="sparkles" size="xs" />
                <span>{{ t('affiliate.partner.badge') }}</span>
              </div>
              <h3 class="mt-3 text-lg font-semibold text-gray-900 dark:text-white">{{ t('affiliate.partner.title') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.partner.description') }}</p>
            </div>

            <div class="w-full rounded-lg border border-gray-200 p-4 dark:border-dark-700 lg:max-w-sm">
              <div class="flex items-center justify-between text-sm">
                <span class="font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.partner.progress') }}</span>
                <span class="text-gray-500 dark:text-dark-400">
                  {{ detail.aff_count }} / {{ progressTarget }}
                </span>
              </div>
              <div class="mt-3 h-2 rounded-full bg-gray-100 dark:bg-dark-800">
                <div
                  class="h-2 rounded-full bg-primary-500 transition-all"
                  :style="{ width: `${progressPercent}%` }"
                ></div>
              </div>
              <p class="mt-3 text-xs text-gray-500 dark:text-dark-400">
                {{ nextTier ? t('affiliate.partner.progressHint', { level: partnerLevelLabel(nextTier.level), count: nextTier.required_invitees }) : t('affiliate.partner.maxLevelHint') }}
              </p>
            </div>
          </div>

          <div class="mt-6 grid gap-3 md:grid-cols-4">
            <div
              v-for="tier in partnerTiers"
              :key="tier.level"
              class="rounded-lg border p-4"
              :class="[
                partnerTierActive(tier.level)
                  ? 'border-primary-300 bg-primary-50 dark:border-primary-800 dark:bg-primary-900/20'
                  : 'border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900/40'
              ]"
            >
              <div class="flex items-center justify-between gap-2">
                <p class="font-medium text-gray-900 dark:text-white">{{ partnerLevelLabel(tier.level) }}</p>
                <Icon v-if="partnerTierActive(tier.level)" name="checkCircle" size="sm" class="text-primary-500" />
              </div>
              <p class="mt-2 text-sm font-semibold text-primary-700 dark:text-primary-300">
                {{ partnerBenefitLabel(tier.level) }}
              </p>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                {{ t('affiliate.partner.tierRequirement', { count: tier.required_invitees }) }}
              </p>
              <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-400">
                {{ t(`affiliate.partner.tierBenefits.${tier.level}`) }}
              </p>
            </div>
          </div>
        </div>

        <div class="grid gap-6 lg:grid-cols-[1fr_380px]">
          <div class="card p-6">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.title') }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t(isPartner ? 'affiliate.partnerInviteDescription' : 'affiliate.description') }}</p>

            <div class="mt-5 grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.yourCode') }}</p>
                <div class="flex items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                  <code class="flex-1 truncate text-sm font-semibold text-gray-900 dark:text-white">{{ detail.aff_code }}</code>
                  <button class="btn btn-secondary btn-sm" @click="copyCode">
                    <Icon name="copy" size="sm" />
                    <span>{{ t('affiliate.copyCode') }}</span>
                  </button>
                </div>
              </div>

              <div class="space-y-2">
                <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.inviteLink') }}</p>
                <div class="flex items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                  <code class="flex-1 truncate text-sm text-gray-700 dark:text-gray-300">{{ inviteLink }}</code>
                  <button class="btn btn-secondary btn-sm" @click="copyInviteLink">
                    <Icon name="copy" size="sm" />
                    <span>{{ t('affiliate.copyLink') }}</span>
                  </button>
                </div>
              </div>
            </div>

            <div class="mt-5 rounded-lg border border-primary-200 bg-primary-50 p-4 dark:border-primary-900/40 dark:bg-primary-900/20">
              <p class="text-sm font-medium text-primary-800 dark:text-primary-200">{{ t('affiliate.tips.title') }}</p>
              <ul class="mt-2 space-y-1 text-sm text-primary-700 dark:text-primary-300">
                <li>1. {{ t('affiliate.tips.line1') }}</li>
                <li>2. {{ t(isPartner ? 'affiliate.tips.partnerLine2' : 'affiliate.tips.line2') }}</li>
                <li>3. {{ t(isPartner ? 'affiliate.tips.partnerLine3' : 'affiliate.tips.line3') }}</li>
                <li v-if="!isPartner && detail.aff_frozen_quota > 0">4. {{ t('affiliate.tips.line4') }}</li>
              </ul>
            </div>
          </div>

          <div class="card p-6">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0 flex-1">
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.partner.applyTitle') }}</h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ applicationStatusText }}</p>
              </div>
              <span
                v-if="latestApplication"
                class="badge shrink-0 whitespace-nowrap"
                :class="applicationBadgeClass(latestApplication.status)"
              >
                {{ applicationStatusLabel(latestApplication.status) }}
              </span>
            </div>

            <form class="mt-5 space-y-4" @submit.prevent="submitPartnerApplication">
              <div>
                <label class="input-label">{{ t('affiliate.partner.source') }}</label>
                <select v-model="applicationForm.source" class="input">
                  <option v-for="option in sourceOptions" :key="option" :value="option">
                    {{ t(`affiliate.partner.sources.${option}`) }}
                  </option>
                </select>
              </div>
              <div>
                <label class="input-label">{{ t('affiliate.partner.portalUrl') }}</label>
                <input v-model.trim="applicationForm.portal_url" type="url" class="input" placeholder="https://..." />
              </div>
              <div>
                <label class="input-label">{{ t('affiliate.partner.strengths') }}</label>
                <textarea v-model.trim="applicationForm.strengths" rows="4" class="input resize-none" :placeholder="t('affiliate.partner.strengthsPlaceholder')"></textarea>
              </div>
              <button
                type="submit"
                class="btn btn-primary w-full"
                :disabled="submittingApplication || hasPendingApplication"
              >
                <Icon v-if="submittingApplication" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="userPlus" size="sm" />
                <span>{{ submitButtonText }}</span>
              </button>
              <p v-if="latestApplication?.review_note" class="text-xs text-gray-500 dark:text-dark-400">
                {{ t('affiliate.partner.reviewNote') }}: {{ latestApplication.review_note }}
              </p>
            </form>
          </div>
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.transfer.title') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t(isPartner ? 'affiliate.transfer.partnerDescription' : 'affiliate.transfer.description') }}</p>
            </div>
            <button
              v-if="!isPartner"
              class="btn btn-primary"
              :disabled="transferring || detail.aff_quota <= 0"
              @click="transferQuota"
            >
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <Icon v-else name="dollar" size="sm" />
              <span>{{ transferring ? t('affiliate.transfer.transferring') : t('affiliate.transfer.button') }}</span>
            </button>
            <span v-else class="badge badge-primary">{{ t('affiliate.transfer.partnerBadge') }}</span>
          </div>
          <p v-if="isPartner" class="mt-3 text-sm text-primary-700 dark:text-primary-300">
            {{ t('affiliate.transfer.partnerHint') }}
          </p>
          <p v-else-if="detail.aff_quota <= 0" class="mt-3 text-sm text-amber-600 dark:text-amber-400">
            {{ t('affiliate.transfer.empty') }}
          </p>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.invitees.title') }}</h3>
          <div v-if="detail.invitees.length === 0" class="mt-4 rounded-lg border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('affiliate.invitees.empty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[560px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.email') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.username') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ t('affiliate.invitees.columns.recharge') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.joinedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in detail.invitees"
                  :key="item.user_id"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3 text-gray-900 dark:text-white">{{ item.email || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.username || '-' }}</td>
                  <td class="px-3 py-3 text-right">
                    <div
                      v-if="inviteeRechargeParts(item).length > 0"
                      class="flex flex-col items-end gap-1 font-medium text-emerald-600 dark:text-emerald-400 sm:flex-row sm:justify-end"
                    >
                      <span v-for="part in inviteeRechargeParts(item)" :key="part">{{ part }}</span>
                    </div>
                    <span v-else class="text-gray-400 dark:text-dark-500">-</span>
                  </td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import userAPI from '@/api/user'
import type {
  AffiliateInvitee,
  AffiliatePartnerApplicationStatus,
  AffiliatePartnerLevel,
  AffiliatePartnerTier,
  UserAffiliateDetail,
} from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const transferring = ref(false)
const submittingApplication = ref(false)
const detail = ref<UserAffiliateDetail | null>(null)

const sourceOptions = ['twitter', 'discord', 'telegram', 'community', 'other'] as const
const applicationForm = reactive({
  source: 'twitter',
  portal_url: '',
  strengths: '',
})

const partnerTiers = computed<AffiliatePartnerTier[]>(() => detail.value?.partner_tiers ?? [])
const currentTier = computed(() => detail.value?.partner_tier ?? null)
const latestApplication = computed(() => detail.value?.partner_application ?? null)
const effectivePartnerLevel = computed<AffiliatePartnerLevel>(() => currentTier.value?.level ?? detail.value?.partner_level ?? 'none')
const isPartner = computed(() => partnerLevelRank(effectivePartnerLevel.value) > 0)
const hasPendingApplication = computed(() => latestApplication.value?.status === 'pending')
const nextTier = computed(() => {
  const d = detail.value
  if (!d) return null
  const currentRank = partnerLevelRank(effectivePartnerLevel.value)
  if (currentRank > 0) {
    return partnerTiers.value.find((tier) => partnerLevelRank(tier.level) > currentRank) ?? null
  }
  return partnerTiers.value.find((tier) => tier.required_invitees > d.aff_count) ?? null
})
const progressTarget = computed(() => nextTier.value?.required_invitees ?? Math.max(detail.value?.aff_count ?? 0, 1))
const progressPercent = computed(() => {
  if (!detail.value) return 0
  return Math.min(100, Math.round((detail.value.aff_count / Math.max(progressTarget.value, 1)) * 100))
})

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

const applicationStatusText = computed(() => {
  const app = latestApplication.value
  if (!app) return t('affiliate.partner.applyDescription')
  if (app.status === 'pending') return t('affiliate.partner.pendingDescription')
  if (app.status === 'approved') return t('affiliate.partner.approvedDescription', { level: partnerLevelLabel(app.granted_level || effectivePartnerLevel.value) })
  return t('affiliate.partner.rejectedDescription')
})

const submitButtonText = computed(() => {
  if (submittingApplication.value) return t('affiliate.partner.submitting')
  if (hasPendingApplication.value) return t('affiliate.partner.pendingButton')
  return latestApplication.value ? t('affiliate.partner.reapplyButton') : t('affiliate.partner.submitButton')
})

function formatCount(value: number): string {
  return value.toLocaleString()
}

function inviteeRechargeParts(item: AffiliateInvitee): string[] {
  const parts: string[] = []
  if ((item.recharge_amount_cny ?? 0) > 0) {
    parts.push(formatCurrency(item.recharge_amount_cny, 'CNY'))
  }
  if ((item.recharge_amount_usdt ?? 0) > 0) {
    parts.push(formatCurrency(item.recharge_amount_usdt, 'USD'))
  }
  if (parts.length === 0 && (item.recharge_amount ?? 0) > 0) {
    parts.push(formatCurrency(item.recharge_amount, 'USD'))
  }
  return parts
}

function partnerLevelLabel(level?: AffiliatePartnerLevel | ''): string {
  const normalized = level || 'none'
  return t(`affiliate.partner.levels.${normalized}`)
}

function partnerBenefitLabel(level?: AffiliatePartnerLevel | ''): string {
  const normalized = level || 'none'
  return t(`affiliate.partner.benefitLabels.${normalized}`)
}

function partnerLevelRank(level?: AffiliatePartnerLevel | ''): number {
  const order: AffiliatePartnerLevel[] = ['none', 'spark', 'voyage', 'summit', 'cocreate']
  return order.indexOf((level || 'none') as AffiliatePartnerLevel)
}

function partnerTierActive(level: AffiliatePartnerLevel): boolean {
  return partnerLevelRank(effectivePartnerLevel.value) >= partnerLevelRank(level)
}

function applicationStatusLabel(status: AffiliatePartnerApplicationStatus): string {
  return t(`affiliate.partner.status.${status}`)
}

function applicationBadgeClass(status: AffiliatePartnerApplicationStatus): string {
  if (status === 'approved') return 'badge-success'
  if (status === 'rejected') return 'badge-danger'
  return 'badge-warning'
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) {
    loading.value = true
  }
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

async function submitPartnerApplication(): Promise<void> {
  if (hasPendingApplication.value || submittingApplication.value) return
  if (!applicationForm.portal_url.trim() || !applicationForm.strengths.trim()) {
    appStore.showError(t('affiliate.partner.formRequired'))
    return
  }
  submittingApplication.value = true
  try {
    const application = await userAPI.applyAffiliatePartner({
      source: applicationForm.source,
      portal_url: applicationForm.portal_url.trim(),
      strengths: applicationForm.strengths.trim(),
    })
    if (detail.value) {
      detail.value = {
        ...detail.value,
        partner_application: application,
      }
    }
    appStore.showSuccess(t('affiliate.partner.submitSuccess'))
    applicationForm.portal_url = ''
    applicationForm.strengths = ''
    void loadAffiliateDetail(true)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.partner.submitFailed')))
  } finally {
    submittingApplication.value = false
  }
}

async function transferQuota(): Promise<void> {
  if (!detail.value || isPartner.value || detail.value.aff_quota <= 0 || transferring.value) return
  transferring.value = true
  try {
    const resp = await userAPI.transferAffiliateQuota()
    appStore.showSuccess(t('affiliate.transfer.success', { amount: formatCurrency(resp.transferred_quota) }))
    await Promise.all([
      loadAffiliateDetail(true),
      authStore.refreshUser().catch(() => undefined),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.transferFailed')))
  } finally {
    transferring.value = false
  }
}

onMounted(() => {
  void loadAffiliateDetail()
})
</script>
