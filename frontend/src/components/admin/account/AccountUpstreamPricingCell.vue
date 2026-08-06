<template>
  <div v-if="eligible" class="min-w-[170px] max-w-[320px] text-xs">
    <template v-if="state?.snapshot">
      <div v-if="groupEntries.length" class="flex flex-wrap gap-1" :title="groupTitle">
        <span class="w-full text-[10px] text-gray-500 dark:text-gray-400">
          {{ snapshotScopeLabel }}
        </span>
        <span
          v-for="group in groupEntries"
          :key="group.key"
          class="rounded-md bg-primary-50 px-1.5 py-0.5 font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
        >
          {{ group.name }} {{ formatRatio(group.ratio) }}x
        </span>
      </div>
      <div v-else class="text-gray-400">
        -
      </div>
    </template>
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
  <span v-else-if="isApiKeyWithBaseUrl" class="text-xs text-amber-600 dark:text-amber-400">
    {{ t('admin.accounts.upstreamPricing.configureGroup') }}
  </span>
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
const isApiKeyWithBaseUrl = computed(() => props.account.type === 'apikey' && typeof props.account.credentials?.base_url === 'string' && Boolean(String(props.account.credentials.base_url).trim()))

const groupEntries = computed(() => {
  const snapshot = props.state?.snapshot
  if (!snapshot) return []
  return Object.entries(snapshot.group_ratios)
    .map(([key, ratio]) => ({ key, ratio, name: snapshot.group_names[key] || key }))
    .sort((a, b) => a.name.localeCompare(b.name))
})

const groupTitle = computed(() => groupEntries.value
  .map(group => `${group.name}: ${formatRatio(group.ratio)}x`)
  .join('\n'))

const snapshotScopeLabel = computed(() => props.state?.snapshot?.ratio_scope === 'effective'
  ? t('admin.accounts.upstreamPricing.effectiveScope')
  : t('admin.accounts.upstreamPricing.baseScope'))

function formatRatio(value: number): string {
  return value.toLocaleString(undefined, { maximumFractionDigits: 6 })
}
</script>
