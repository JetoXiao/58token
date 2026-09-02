<template>
  <div class="relative flex min-h-screen flex-col overflow-hidden bg-[#f7f8fb] text-gray-950 dark:bg-[#05060a] dark:text-white">
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute left-1/2 top-[-18rem] h-[42rem] w-[42rem] -translate-x-1/2 rounded-full bg-cyan-300/30 blur-3xl dark:bg-cyan-400/20"></div>
      <div class="absolute right-[-14rem] top-24 h-[34rem] w-[34rem] rounded-full bg-violet-300/25 blur-3xl dark:bg-violet-500/18"></div>
      <div class="absolute bottom-[-18rem] left-[-10rem] h-[38rem] w-[38rem] rounded-full bg-emerald-300/20 blur-3xl dark:bg-emerald-400/10"></div>
      <div class="absolute inset-0 bg-[linear-gradient(rgba(15,23,42,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(15,23,42,0.045)_1px,transparent_1px)] bg-[size:72px_72px] [mask-image:radial-gradient(circle_at_top,black,transparent_78%)] dark:bg-[linear-gradient(rgba(255,255,255,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.045)_1px,transparent_1px)]"></div>
    </div>

    <MarketingNavbar
      :site-name="siteName"
      :subtitle="siteSubtitle"
      :logo="siteLogo"
      :doc-url="docUrl"
      docs-to="/docs"
      :docs-label="t('nav.integrationDocs')"
      :cta-to="isAuthenticated ? dashboardPath : '/login'"
      :cta-label="isAuthenticated ? t('home.dashboard') : t('home.login')"
      model-marketplace-to="/available-channels"
      :model-marketplace-label="t('nav.availableChannels')"
      partner-to="/partners"
      :partner-label="t('gateway.common.partner')"
      resources-to="/downloads"
      :resources-label="t('gateway.common.resources')"
      :visible-items="marketingNavItems"
    >
      <template #tools>
        <LocaleSwitcher />
        <button
          type="button"
          class="rounded-xl border border-gray-200/70 bg-white/70 p-2 text-gray-600 transition hover:bg-white hover:text-gray-950 dark:border-white/10 dark:bg-white/[0.04] dark:text-slate-300 dark:hover:bg-white/10 dark:hover:text-white"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @click="toggleTheme"
        >
          <Icon v-if="isDark" name="sun" size="md" />
          <Icon v-else name="moon" size="md" />
        </button>
      </template>
    </MarketingNavbar>

    <main class="relative z-10 flex-1 px-4 pb-14 pt-6 sm:px-6">
      <section class="mx-auto max-w-7xl overflow-hidden rounded-[2rem] border border-gray-200/70 bg-white/72 p-5 shadow-[0_26px_90px_rgba(15,23,42,0.10)] backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045] dark:shadow-[0_26px_90px_rgba(0,0,0,0.30)] md:p-8">
        <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_360px] lg:items-end">
          <div>
            <div class="inline-flex items-center gap-2 rounded-full border border-primary-200 bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700 dark:border-primary-800/60 dark:bg-primary-900/20 dark:text-primary-300">
              <Icon name="sparkles" size="sm" />
              {{ t('availableChannels.hero.eyebrow') }}
            </div>
            <h1 class="mt-5 max-w-4xl text-4xl font-semibold tracking-normal text-gray-950 dark:text-white md:text-6xl">
              {{ t('availableChannels.title') }}
            </h1>
            <p class="mt-5 max-w-3xl text-base leading-8 text-gray-600 dark:text-slate-300">
              {{ t('availableChannels.hero.subtitle') }}
            </p>
            <div class="mt-6 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-slate-400">
              <span class="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 font-medium text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-900/20 dark:text-emerald-300">
                {{ t('availableChannels.hero.exchangeRate', { rate: usdToCnyLabel }) }}
              </span>
              <span class="rounded-full border border-gray-200 bg-white/70 px-3 py-1 dark:border-white/10 dark:bg-white/5">
                {{ t('availableChannels.hero.unit') }}
              </span>
            </div>
            <div class="mt-4 max-w-3xl rounded-2xl border border-emerald-200/80 bg-white/62 p-4 text-sm leading-6 text-gray-600 shadow-sm backdrop-blur dark:border-emerald-900/50 dark:bg-white/[0.04] dark:text-slate-300">
              <p class="font-semibold text-emerald-700 dark:text-emerald-300">{{ t('availableChannels.hero.pricingFormula') }}</p>
              <p class="mt-1 text-xs text-gray-500 dark:text-slate-400">{{ pricingExample }}</p>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <StatCard :label="t('availableChannels.stats.models')" :value="uniqueModelCount" />
            <StatCard :label="t('availableChannels.stats.providers')" :value="totalProviderCount" />
            <StatCard :label="t('availableChannels.stats.groups')" :value="totalGroupCount" />
            <StatCard :label="t('availableChannels.stats.priced')" :value="pricedModelCount" />
          </div>
        </div>
      </section>

      <div class="mx-auto mt-6 grid max-w-7xl gap-6 xl:grid-cols-[280px_minmax(0,1fr)]">
        <aside class="rounded-[1.5rem] border border-gray-200/70 bg-white/75 p-5 shadow-sm backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045] xl:sticky xl:top-6 xl:self-start">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-950 dark:text-white">{{ t('availableChannels.filters.title') }}</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-slate-400">{{ t('availableChannels.filters.description') }}</p>
            </div>
            <button type="button" class="rounded-xl border border-gray-200 px-3 py-2 text-xs font-medium text-gray-600 transition hover:border-primary-300 hover:text-primary-600 dark:border-white/10 dark:text-slate-300 dark:hover:border-primary-700 dark:hover:text-primary-300" @click="resetFilters">
              {{ t('availableChannels.filters.reset') }}
            </button>
          </div>

          <div class="mt-5 space-y-6">
            <FilterBlock :title="t('availableChannels.filters.providers')" :options="providerOptions" :selected="selectedProvider ? [selectedProvider] : []" @toggle="selectProvider" />
            <FilterBlock :title="t('availableChannels.filters.groups')" :options="groupOptions" :selected="selectedGroup ? [selectedGroup] : []" @toggle="selectGroup" />
            <FilterBlock :title="t('availableChannels.filters.capabilities')" :options="capabilityOptions" :selected="selectedCapability ? [selectedCapability] : []" @toggle="selectCapability" />
          </div>
        </aside>

        <section class="min-w-0 space-y-5">
          <div class="rounded-[1.5rem] border border-gray-200/70 bg-white/75 p-4 shadow-sm backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045]">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div class="relative min-w-0 flex-1">
                <Icon name="search" size="md" class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 dark:text-slate-500" />
                <input
                  v-model="searchQuery"
                  type="text"
                  :placeholder="t('availableChannels.searchPlaceholder')"
                  class="h-12 w-full rounded-2xl border border-gray-200 bg-white/80 pl-11 pr-4 text-sm text-gray-950 outline-none transition placeholder:text-gray-400 focus:border-primary-300 focus:ring-4 focus:ring-primary-100 dark:border-white/10 dark:bg-white/[0.05] dark:text-white dark:placeholder:text-slate-500 dark:focus:border-primary-700 dark:focus:ring-primary-950/40"
                />
              </div>
              <div class="flex flex-wrap items-center gap-3">
                <SegmentedControl :items="currencyItems" :model-value="currencyMode" @update:model-value="currencyMode = $event as CurrencyMode" />
                <SegmentedControl :items="viewItems" :model-value="viewMode" @update:model-value="viewMode = $event as ViewMode" />
                <button type="button" class="inline-flex h-12 items-center justify-center rounded-2xl border border-gray-200 bg-white/80 px-4 text-sm font-medium text-gray-700 shadow-sm transition hover:border-primary-300 hover:text-primary-600 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/10 dark:bg-white/[0.05] dark:text-slate-200 dark:hover:border-primary-700 dark:hover:text-primary-300" :disabled="loading" @click="loadPricingConfig">
                  <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
                </button>
              </div>
            </div>
            <div class="mt-4 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-slate-400">
              <span>{{ t('availableChannels.results', { count: filteredModels.length }) }}</span>
              <span v-if="activeFilterCount > 0" class="text-gray-300 dark:text-slate-700">/</span>
              <span v-if="activeFilterCount > 0">{{ t('availableChannels.activeFilters', { count: activeFilterCount }) }}</span>
              <span v-if="sourceVersion" class="ml-auto truncate">{{ t('availableChannels.dataVersion', { version: sourceVersion }) }}</span>
            </div>
          </div>

          <div v-if="loading" class="rounded-[1.5rem] border border-gray-200/70 bg-white/75 py-16 text-center shadow-sm dark:border-white/10 dark:bg-white/[0.045]">
            <Icon name="refresh" size="lg" class="mx-auto animate-spin text-primary-500" />
            <p class="mt-3 text-sm text-gray-500 dark:text-slate-400">{{ t('common.loading') }}</p>
          </div>

          <div v-else-if="filteredModels.length === 0" class="rounded-[1.5rem] border border-gray-200/70 bg-white/75 py-16 text-center shadow-sm dark:border-white/10 dark:bg-white/[0.045]">
            <Icon name="inbox" size="xl" class="mx-auto mb-3 h-12 w-12 text-gray-400" />
            <p class="text-sm text-gray-500 dark:text-slate-400">{{ t('availableChannels.empty') }}</p>
          </div>

          <div v-else-if="viewMode === 'cards'" class="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3">
            <article v-for="item in filteredModels" :key="`${item.modelName}-${item.group}`" class="group rounded-[1.5rem] border border-gray-200/70 bg-white/82 p-5 shadow-sm backdrop-blur-xl transition duration-200 hover:-translate-y-0.5 hover:border-primary-300 hover:shadow-lg hover:shadow-primary-100/60 dark:border-white/10 dark:bg-white/[0.045] dark:hover:border-primary-800 dark:hover:shadow-primary-950/20">
              <div class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start">
                <div class="flex min-w-0 items-start gap-3">
                  <span class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border border-gray-200 bg-gray-50 shadow-sm dark:border-white/10 dark:bg-white/5">
                    <ModelIcon :model="item.modelName" size="24px" />
                  </span>
                  <div class="min-w-0 flex-1">
                    <div class="flex min-w-0 items-start gap-2">
                      <h3 class="min-w-0 max-w-full break-words text-base font-semibold leading-6 text-gray-950 dark:text-white" :title="item.modelName">{{ item.modelName }}</h3>
                      <button type="button" class="shrink-0 rounded-lg p-1 text-gray-400 transition hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-white/10 dark:hover:text-primary-300" :title="t('availableChannels.copyModel')" @click="copyModel(item.modelName)">
                        <Icon name="copy" size="sm" />
                      </button>
                    </div>
                    <p class="mt-1 truncate text-xs text-gray-500 dark:text-slate-400">{{ item.vendorName }} / {{ item.group }} · {{ item.groupMultiplier }}x</p>
                  </div>
                </div>
                <div class="flex shrink-0 flex-col items-end gap-2">
                  <span class="rounded-full border border-gray-200 px-2.5 py-1 text-xs font-medium text-gray-600 dark:border-white/10 dark:text-slate-300">
                    {{ item.billingLabel }}
                  </span>
                  <span class="rounded-full border border-emerald-200 bg-emerald-50 px-2.5 py-1 text-xs font-semibold text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-900/20 dark:text-emerald-300">
                    {{ itemDiscountLabel(item) }}
                  </span>
                </div>
              </div>

              <p class="mt-4 line-clamp-2 text-sm leading-6 text-gray-500 dark:text-slate-400">
                {{ item.description }}
              </p>

              <div class="mt-4 flex flex-wrap gap-1.5">
                <span v-for="tag in item.capabilities.slice(0, 5)" :key="`${item.modelName}-${item.group}-${tag}`" class="rounded-full border border-gray-200 bg-gray-50 px-2 py-1 text-[11px] font-medium text-gray-600 dark:border-white/10 dark:bg-white/5 dark:text-slate-300">
                  {{ localizedCapability(tag) }}
                </span>
              </div>

              <div class="mt-5 grid gap-2">
                <PriceRowView v-for="row in priceRows(item)" :key="row.key" :row="row" />
              </div>
            </article>
          </div>

          <div v-else class="overflow-hidden rounded-[1.5rem] border border-gray-200/70 bg-white/82 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.045]">
            <div class="overflow-x-auto">
              <table class="min-w-[980px] w-full border-collapse text-sm">
                <thead class="bg-gray-50/80 text-xs font-medium uppercase text-gray-500 dark:bg-white/[0.04] dark:text-slate-400">
                  <tr>
                    <th class="px-5 py-4 text-left">{{ t('availableChannels.table.model') }}</th>
                    <th class="px-5 py-4 text-left">{{ t('availableChannels.table.provider') }}</th>
                    <th class="px-5 py-4 text-left">{{ t('availableChannels.table.groups') }}</th>
                    <th class="px-5 py-4 text-left">{{ t('availableChannels.pricing.inputPrice') }}</th>
                    <th class="px-5 py-4 text-left">{{ t('availableChannels.pricing.outputPrice') }}</th>
                    <th class="px-5 py-4 text-left">{{ t('availableChannels.pricing.cacheWritePrice') }}</th>
                    <th class="px-5 py-4 text-left">{{ t('availableChannels.pricing.cacheReadPrice') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in filteredModels" :key="`row-${item.modelName}-${item.group}`" class="border-t border-gray-100 transition hover:bg-gray-50/70 dark:border-white/10 dark:hover:bg-white/[0.04]">
                    <td class="px-5 py-4">
                      <button type="button" class="flex min-w-0 items-center gap-3 font-semibold text-gray-950 hover:text-primary-600 dark:text-white dark:hover:text-primary-300" @click="copyModel(item.modelName)">
                        <ModelIcon :model="item.modelName" size="20px" />
                        <span class="break-words">{{ item.modelName }}</span>
                      </button>
                    </td>
                    <td class="px-5 py-4 text-gray-600 dark:text-slate-300">{{ item.vendorName }}</td>
                    <td class="px-5 py-4 text-gray-600 dark:text-slate-300">{{ item.group }}</td>
                    <td class="px-5 py-4">{{ tablePrice(item, 'input') }}</td>
                    <td class="px-5 py-4">{{ tablePrice(item, 'output') }}</td>
                    <td class="px-5 py-4">{{ tablePrice(item, 'cacheWrite') }}</td>
                    <td class="px-5 py-4">{{ tablePrice(item, 'cacheRead') }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref, type PropType } from 'vue'
import axios from 'axios'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import MarketingNavbar from '@/components/marketing/MarketingNavbar.vue'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { BRAND_LOGO_URL } from '@/constants/brand'
import { normalizeMarketplaceModelName } from '@/utils/marketplaceModel'
import type { PaymentConfig } from '@/types/payment'

type CurrencyMode = 'usd' | 'cny'
type ViewMode = 'cards' | 'table'
type PriceKey = 'input' | 'output' | 'cacheWrite' | 'cacheRead'
type OfficialPrice = number | PriceTier[] | null

interface PriceTier {
  label: string
  value: number
}

interface FilterOption {
  value: string
  label: string
  count: number
}

interface PricingModel {
  model_name: string
  description?: string
  tags?: string
  vendor_id: number
  quota_type: number
  model_price: number
  supported_endpoint_types?: string[]
}

interface PricingResponse {
  success?: boolean
  data?: PricingModel[]
  pricing_version?: string
}

interface MarketplaceItemResponse {
  id: number
  model_name: string
  pricing_aliases?: string[]
  vendor_name: string
  groups: string[]
  tags: string[]
  endpoints: string[]
  description?: string
  official_prices: Record<PriceKey, OfficialPrice>
  sort_order: number
  enabled: boolean
}

interface MarketplaceResponse {
  items: MarketplaceItemResponse[]
}

interface MarketplaceModel {
  modelName: string
  pricingAliases: string[]
  vendorName: string
  group: string
  groupMultiplier: number
  capabilities: string[]
  billingLabel: string
  description: string
  prices: Record<PriceKey, OfficialPrice>
  requestPrice: number | null
  searchText: string
}

interface PriceRow {
  key: string
  label: string
  platform: string
  official: string
  discount: string
  unit: string
}

const FALLBACK_PRICING: PricingResponse = { success: true, pricing_version: 'fallback', data: [] }

const DEFAULT_MARKETPLACE_GROUP_MULTIPLIERS: Record<string, number> = {}
const MARKETPLACE_MODEL_DESCRIPTION_KEYS: Record<string, string> = {
  'claude-fable-5': 'claudeFable5',
  'claude-opus-4-8': 'claudeOpus48'
}
const MARKETPLACE_CAPABILITY_KEYS: Record<string, string> = {
  Reasoning: 'reasoning',
  Tools: 'tools',
  Files: 'files',
  Vision: 'vision',
  'Computer Use': 'computerUse',
  'Adaptive Thinking': 'adaptiveThinking',
  'Fast Mode': 'fastMode',
  '1M': 'context1m',
  '128K': 'output128k',
  'OpenAI API': 'openaiApi',
  anthropic: 'anthropic',
  openai: 'openai'
}
const RETIRED_MARKETPLACE_MODELS = new Set(['gpt-5.2', 'gpt-5.3-codex'])

const FilterBlock = defineComponent({
  name: 'FilterBlock',
  props: {
    title: { type: String, required: true },
    options: { type: Array as PropType<FilterOption[]>, required: true },
    selected: { type: Array as PropType<string[]>, required: true }
  },
  emits: ['toggle'],
  setup(props, { emit }) {
    return () => h('div', [
      h('h3', { class: 'mb-3 text-xs font-semibold uppercase text-gray-500 dark:text-slate-400' }, props.title),
      h('div', { class: 'flex flex-wrap gap-2 xl:flex-col' }, props.options.map((option) => {
        const active = props.selected.includes(option.value)
        return h('button', {
          key: option.value,
          type: 'button',
          class: [
            'flex items-center justify-between gap-3 rounded-xl border px-3 py-2 text-left text-sm font-medium transition',
            active
              ? 'border-primary-500 bg-primary-100 text-primary-800 shadow-[0_0_0_3px_rgba(20,184,166,0.16)] ring-1 ring-primary-300 dark:border-primary-400 dark:bg-primary-500/20 dark:text-primary-100 dark:shadow-[0_0_0_3px_rgba(45,212,191,0.16)] dark:ring-primary-500/60'
              : 'border-gray-200 bg-white/70 text-gray-600 hover:border-primary-200 hover:text-primary-600 dark:border-white/10 dark:bg-white/[0.04] dark:text-slate-300 dark:hover:border-primary-800 dark:hover:text-primary-300'
          ],
          onClick: () => emit('toggle', option.value)
        }, [
          h('span', { class: 'truncate' }, option.label),
          h('span', { class: 'shrink-0 rounded-full bg-gray-100 px-2 py-0.5 text-[11px] text-gray-500 dark:bg-white/10 dark:text-slate-400' }, option.count)
        ])
      }))
    ])
  }
})

const StatCard = defineComponent({
  name: 'StatCard',
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number], required: true }
  },
  setup(props) {
    return () => h('div', { class: 'rounded-2xl border border-gray-200 bg-white/80 p-4 dark:border-white/10 dark:bg-white/[0.05]' }, [
      h('p', { class: 'text-xs font-medium text-gray-500 dark:text-slate-400' }, props.label),
      h('p', { class: 'mt-2 text-2xl font-semibold text-gray-950 dark:text-white' }, props.value)
    ])
  }
})

const SegmentedControl = defineComponent({
  name: 'SegmentedControl',
  props: {
    items: { type: Array as PropType<Array<{ value: string; label: string }>>, required: true },
    modelValue: { type: String, required: true }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('div', { class: 'inline-flex rounded-2xl border border-gray-200 bg-gray-50 p-1 dark:border-white/10 dark:bg-white/[0.04]' },
      props.items.map((item) => h('button', {
        key: item.value,
        type: 'button',
        class: [
          'rounded-xl px-3 py-2 text-sm font-medium transition',
          props.modelValue === item.value
            ? 'bg-white text-gray-950 shadow-sm dark:bg-white/10 dark:text-white'
            : 'text-gray-500 hover:text-gray-900 dark:text-slate-400 dark:hover:text-white'
        ],
        onClick: () => emit('update:modelValue', item.value)
      }, item.label))
    )
  }
})

const PriceRowView = defineComponent({
  name: 'PriceRowView',
  props: {
    row: { type: Object as PropType<PriceRow>, required: true }
  },
  setup(props) {
    return () => h('div', { class: 'rounded-xl border border-gray-200 bg-gray-50/80 p-3 dark:border-white/10 dark:bg-white/[0.04]' }, [
      h('div', { class: 'flex items-center justify-between gap-3' }, [
        h('span', { class: 'text-xs font-medium text-gray-500 dark:text-slate-400' }, props.row.label),
        h('span', { class: 'text-sm font-semibold text-gray-950 dark:text-white' }, props.row.platform)
      ]),
      h('div', { class: 'mt-1 flex items-center justify-between gap-3' }, [
        h('span', { class: 'text-[11px] text-gray-400 dark:text-slate-500' }, props.row.unit),
        h('span', { class: 'text-xs font-medium text-emerald-700 dark:text-emerald-300' }, props.row.official)
      ])
    ])
  }
})

const { t, te, locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const rawPricing = ref<PricingResponse>(FALLBACK_PRICING)
const marketplaceItems = ref<MarketplaceItemResponse[]>([])
const loading = ref(false)
const searchQuery = ref('')
const selectedProvider = ref('')
const selectedGroup = ref('')
const selectedCapability = ref('')
const viewMode = ref<ViewMode>('cards')
const currencyMode = ref<CurrencyMode>('usd')
const usdToCnyRate = ref(7)
const marketplaceGroupMultipliers = ref<Record<string, number>>({ ...DEFAULT_MARKETPLACE_GROUP_MULTIPLIERS })
const isDark = ref(document.documentElement.classList.contains('dark'))

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'UseAiForMe')
const siteLogo = computed(() => appStore.siteLogo || BRAND_LOGO_URL)
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || t('home.heroSubtitle'))
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const marketingNavItems = computed(() => appStore.cachedPublicSettings?.marketing_nav_items)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const usdToCnyLabel = computed(() => trimNumber(usdToCnyRate.value, 2))
const sourceVersion = computed(() => rawPricing.value.pricing_version || '')
const isChineseLocale = computed(() => String(locale.value).startsWith('zh'))
const pricingExample = computed(() => {
  const official = 5
  const multiplier = selectedGroup.value ? groupMultiplierFor(selectedGroup.value) : 1
  const price = platformPrice(official, multiplier)
  return t('availableChannels.hero.pricingExample', {
    official: `$${trimNumber(official, 2)}`,
    multiplier: `${trimNumber(multiplier, 2)}x`,
    rate: usdToCnyLabel.value,
    price: price == null ? '-' : `$${price.toFixed(2)}`
  })
})
const currencyItems = computed(() => [
  { value: 'usd', label: t('availableChannels.currency.usd') },
  { value: 'cny', label: t('availableChannels.currency.cny') }
])
const viewItems = computed(() => [
  { value: 'cards', label: t('availableChannels.view.cards') },
  { value: 'table', label: t('availableChannels.view.table') }
])

const marketplaceModels = computed<MarketplaceModel[]>(() => marketplaceItems.value.filter(isVisibleMarketplaceItem).flatMap((definition) => {
  const modelName = normalizeMarketplaceModelName(definition.model_name)
  const pricingAliases = modelPricingAliases(modelName, definition.pricing_aliases)
  const pricingModel = rawPricing.value.data?.find((item) => pricingAliases.includes(item.model_name.toLowerCase()))
  const capabilities = parseCapabilities(definition.tags, pricingModel?.supported_endpoint_types || [...definition.endpoints])
  const description = localizedModelDescription(definition, pricingModel)
  const prices = normalizeOfficialPrices(definition.official_prices)
  const localizedCapabilities = capabilities.map(localizedCapability)

  return definition.groups.map((group) => {
    const groupMultiplier = groupMultiplierFor(group)
    return {
      modelName,
      pricingAliases,
      vendorName: definition.vendor_name,
      group,
      groupMultiplier,
      capabilities,
      billingLabel: t('availableChannels.pricing.billingModeToken'),
      description,
      prices,
      requestPrice: null,
      searchText: [definition.model_name, ...pricingAliases, definition.vendor_name, group, ...capabilities, ...localizedCapabilities, ...definition.endpoints].join(' ').toLowerCase()
    }
  })
}))

const uniqueModelCount = computed(() => new Set(marketplaceModels.value.map((model) => model.modelName)).size)
const pricedModelCount = computed(() => new Set(marketplaceModels.value.filter((model) => Object.values(model.prices).some((price) => price != null)).map((model) => model.modelName)).size)
const totalProviderCount = computed(() => new Set(marketplaceModels.value.map((model) => model.vendorName)).size)
const totalGroupCount = computed(() => new Set(marketplaceModels.value.map((model) => model.group)).size)
const providerOrder = computed(() => orderedUnique(marketplaceItems.value.map((item) => item.vendor_name)))
const groupOrder = computed(() => orderedUnique(marketplaceItems.value.flatMap((item) => item.groups)))
const providerOptions = computed(() => toOptions(countUniqueModelsBy(marketplaceModels.value, (model) => model.vendorName), providerOrder.value))
const groupOptions = computed(() => {
  const models = selectedProvider.value ? marketplaceModels.value.filter((model) => model.vendorName === selectedProvider.value) : marketplaceModels.value
  return toOptions(countUniqueModelsBy(models, (model) => model.group), groupOrder.value)
})
const capabilityOptions = computed(() => toOptions(
  countUniqueModelsBy(filterModelsByPrimarySelection(marketplaceModels.value), (model) => model.capabilities),
  [],
  localizedCapability
))
const activeFilterCount = computed(() => (selectedProvider.value ? 1 : 0) + (selectedGroup.value ? 1 : 0) + (selectedCapability.value ? 1 : 0))
const filteredModels = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return filterModelsByPrimarySelection(marketplaceModels.value).filter((model) => {
    if (q && !model.searchText.includes(q)) return false
    if (selectedCapability.value && !model.capabilities.includes(selectedCapability.value)) return false
    return true
  })
})

function countUniqueModelsBy(models: MarketplaceModel[], keyFn: (model: MarketplaceModel) => string | string[]): Map<string, number> {
  const buckets = new Map<string, Set<string>>()
  models.forEach((model) => {
    const rawKeys = keyFn(model)
    const keys = Array.isArray(rawKeys) ? rawKeys : [rawKeys]
    keys.forEach((key) => {
      if (!key) return
      if (!buckets.has(key)) buckets.set(key, new Set<string>())
      buckets.get(key)?.add(model.modelName)
    })
  })
  return new Map(Array.from(buckets.entries()).map(([key, models]) => [key, models.size]))
}

function orderedUnique(values: string[]): string[] {
  const seen = new Set<string>()
  const result: string[] = []
  values.forEach((value) => {
    if (!value || seen.has(value)) return
    seen.add(value)
    result.push(value)
  })
  return result
}

function filterModelsByPrimarySelection(models: MarketplaceModel[]): MarketplaceModel[] {
  return models.filter((model) => {
    if (selectedProvider.value && model.vendorName !== selectedProvider.value) return false
    if (selectedGroup.value && model.group !== selectedGroup.value) return false
    return true
  })
}

function toOptions(counts: Map<string, number>, order: string[] = [], labelFn: (value: string) => string = (value) => value): FilterOption[] {
  return Array.from(counts.entries())
    .sort((a, b) => {
      const orderA = order.indexOf(a[0])
      const orderB = order.indexOf(b[0])
      if (orderA >= 0 || orderB >= 0) {
        if (orderA < 0) return 1
        if (orderB < 0) return -1
        return orderA - orderB
      }
      return b[1] - a[1] || a[0].localeCompare(b[0])
    })
    .map(([value, count]) => ({ value, label: labelFn(value), count }))
}

function selectProvider(value: string) {
  selectedProvider.value = value
  const nextGroup = groupOptions.value.find((option) => option.value === selectedGroup.value)?.value || groupOptions.value[0]?.value || ''
  selectedGroup.value = nextGroup
}

function selectGroup(value: string) {
  selectedGroup.value = value
}

function selectCapability(value: string) {
  selectedCapability.value = selectedCapability.value === value ? '' : value
}

function resetFilters() {
  searchQuery.value = ''
  selectedProvider.value = providerOptions.value[0]?.value || ''
  selectedGroup.value = groupOptions.value[0]?.value || ''
  selectedCapability.value = ''
}

function parseCapabilities(tags: string[], endpoints: string[]): string[] {
  const tagItems = tags.map((tag) => tag.trim()).filter(Boolean)
  const endpointItems = endpoints.map((endpoint) => endpoint === 'openai' ? 'OpenAI API' : endpoint)
  return Array.from(new Set([...tagItems, ...endpointItems]))
}

function localizedModelDescription(definition: MarketplaceItemResponse, pricingModel?: PricingModel): string {
  const modelName = normalizeMarketplaceModelName(definition.model_name).toLowerCase()
  const exactKey = MARKETPLACE_MODEL_DESCRIPTION_KEYS[modelName]
  if (exactKey && te(`availableChannels.modelDescriptions.${exactKey}`)) {
    return t(`availableChannels.modelDescriptions.${exactKey}`)
  }

  const rawDescription = definition.description || pricingModel?.description || ''
  if (isChineseLocale.value && hasLatinWords(rawDescription)) {
    return modelDescription(modelName)
  }
  return rawDescription || modelDescription(modelName)
}

function localizedCapability(value: string): string {
  const key = MARKETPLACE_CAPABILITY_KEYS[value]
  if (!key) return value
  const localeKey = `availableChannels.capabilityLabels.${key}`
  return te(localeKey) ? t(localeKey) : value
}

function hasLatinWords(value: string): boolean {
  return /[A-Za-z]{3,}/.test(value)
}

function modelPricingAliases(modelName: string, configuredAliases: readonly string[] | undefined): string[] {
  const aliases = [modelName, ...(configuredAliases || []), modelName.replace(/-(\d+)\.(\d+)(?=$|-)/g, '-$1-$2')]
  return Array.from(new Set(aliases.map((alias) => alias.toLowerCase())))
}

function normalizeOfficialPrices(raw: Partial<Record<PriceKey, OfficialPrice>> | undefined): Record<PriceKey, OfficialPrice> {
  return {
    input: normalizeOfficialPrice(raw?.input),
    output: normalizeOfficialPrice(raw?.output),
    cacheWrite: normalizeOfficialPrice(raw?.cacheWrite),
    cacheRead: normalizeOfficialPrice(raw?.cacheRead)
  }
}

function normalizeOfficialPrice(raw: OfficialPrice | undefined): OfficialPrice {
  if (raw == null) return null
  if (typeof raw === 'number') return Number.isFinite(raw) ? raw : null
  if (!Array.isArray(raw)) return null
  const tiers = raw
    .map((tier) => ({ label: String(tier.label || '').trim(), value: Number(tier.value) }))
    .filter((tier) => tier.label && Number.isFinite(tier.value))
  return tiers.length ? tiers : null
}

function modelDescription(modelName: string): string {
  if (modelName.includes('claude')) return t('availableChannels.modelDescriptions.anthropic')
  if (modelName.includes('gpt') || modelName.includes('codex')) return t('availableChannels.modelDescriptions.openai')
  return t('availableChannels.modelDescriptions.default')
}

function groupMultiplierFor(group: string): number {
  const configured = Number(marketplaceGroupMultipliers.value[group])
  return Number.isFinite(configured) && configured > 0 ? configured : 1
}

function itemDiscountLabel(item: MarketplaceModel): string {
  const official = firstOfficialPrice(item.prices)
  return discountLabel(official, platformPrice(official, item.groupMultiplier))
}

function priceRows(item: MarketplaceModel): PriceRow[] {
  return [
    makePriceRow(item, 'input', t('availableChannels.pricing.inputPrice'), item.prices.input),
    makePriceRow(item, 'output', t('availableChannels.pricing.outputPrice'), item.prices.output),
    makePriceRow(item, 'cacheWrite', t('availableChannels.pricing.cacheWritePrice'), item.prices.cacheWrite),
    makePriceRow(item, 'cacheRead', t('availableChannels.pricing.cacheReadPrice'), item.prices.cacheRead)
  ].filter((row) => row.platform !== '-')
}

function makePriceRow(item: MarketplaceModel, key: string, label: string, value: OfficialPrice): PriceRow {
  const firstValue = firstPriceValue(value)
  return {
    key,
    label,
    platform: formatPlatformPrice(value, item.groupMultiplier),
    official: value == null ? t('availableChannels.pricing.noOfficialPrice') : t('availableChannels.pricing.officialPrice', { price: formatOfficialPrice(value) }),
    discount: discountLabel(firstValue, platformPrice(firstValue, item.groupMultiplier)),
    unit: key === 'request' ? t('availableChannels.pricing.unitPerRequest') : t('availableChannels.pricing.unitPerMillion')
  }
}

function tablePrice(item: MarketplaceModel, key: PriceKey): string {
  return formatPlatformPrice(item.prices[key], item.groupMultiplier)
}

function formatPrice(value: number | null): string {
  if (value == null || !Number.isFinite(value)) return '-'
  if (currencyMode.value === 'cny') return `\u00A5${trimNumber(value * usdToCnyRate.value, 4)}`
  return `$${value.toFixed(2)}`
}

function formatOfficialPrice(value: OfficialPrice): string {
  if (value == null) return '-'
  if (typeof value === 'number') return formatPrice(value)
  return value.map((tier) => `${tier.label} ${formatPrice(tier.value)}`).join(' / ')
}

function formatPlatformPrice(value: OfficialPrice, multiplier: number): string {
  if (value == null) return '-'
  if (typeof value === 'number') return formatPrice(platformPrice(value, multiplier))
  return value.map((tier) => `${tier.label} ${formatPrice(platformPrice(tier.value, multiplier))}`).join(' / ')
}

function firstPriceValue(value: OfficialPrice): number | null {
  if (value == null) return null
  if (typeof value === 'number') return Number.isFinite(value) ? value : null
  return value.find((tier) => Number.isFinite(tier.value))?.value ?? null
}

function firstOfficialPrice(prices: Record<PriceKey, OfficialPrice>): number | null {
  for (const value of Object.values(prices)) {
    const first = firstPriceValue(value)
    if (first != null) return first
  }
  return null
}

function platformPrice(officialPrice: number | null, multiplier: number): number | null {
  if (officialPrice == null || !Number.isFinite(officialPrice) || usdToCnyRate.value <= 0) return null
  return officialPrice * multiplier / usdToCnyRate.value
}

function discountLabel(officialPrice: number | null, platformPriceValue: number | null): string {
  if (officialPrice == null || platformPriceValue == null || officialPrice <= 0) {
    return t('availableChannels.pricing.noOfficialPrice')
  }
  const ratio = platformPriceValue / officialPrice
  const discount = Math.max(0, Math.min(100, (1 - ratio) * 100))
  return t('availableChannels.pricing.discountOff', { discount: trimNumber(discount, 1) })
}

function trimNumber(value: number, maxDigits = 6): string {
  if (!Number.isFinite(value)) return '0'
  const fixed = value >= 100 ? value.toFixed(2) : value.toFixed(maxDigits)
  return fixed.replace(/\.?0+$/, '')
}

async function copyModel(modelName: string) {
  await copyToClipboard(modelName, t('common.copiedToClipboard'))
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  isDark.value = savedTheme === 'dark' || (!savedTheme && prefersDark)
  document.documentElement.classList.toggle('dark', isDark.value)
}

async function loadPricingConfig() {
  loading.value = true
  try {
    const [paymentConfig, pricing] = await Promise.all([
      loadMarketplacePaymentConfig(),
      axios.get<PricingResponse>('/api/v1/public/model-pricing', { timeout: 12000 }).catch(() => ({ data: FALLBACK_PRICING })),
      loadMarketplaceItems()
    ])
    const configuredRate = Number(paymentConfig?.usdt_cny_exchange_rate)
    if (Number.isFinite(configuredRate) && configuredRate > 0) usdToCnyRate.value = configuredRate
    marketplaceGroupMultipliers.value = normalizeMarketplaceGroupMultipliers(paymentConfig?.marketplace_group_multipliers)
    rawPricing.value = pricing.data?.data?.length ? pricing.data : FALLBACK_PRICING
    syncDefaultSelections()
  } catch (error) {
    console.error('Failed to load model marketplace config:', error)
    rawPricing.value = FALLBACK_PRICING
    marketplaceItems.value = []
  } finally {
    loading.value = false
  }
}

async function loadMarketplaceItems(): Promise<void> {
  const response = await axios.get('/api/v1/public/model-marketplace', { timeout: 12000 })
  const payload = unwrapMarketplacePayload(response.data)
  marketplaceItems.value = payload.items.filter(isVisibleMarketplaceItem)
}

function unwrapMarketplacePayload(payload: unknown): MarketplaceResponse {
  if (!payload || typeof payload !== 'object') return { items: [] }
  const record = payload as Record<string, unknown>
  const data = record.data
  if (data && typeof data === 'object') {
    const dataRecord = data as Record<string, unknown>
    return { items: Array.isArray(dataRecord.items) ? dataRecord.items as MarketplaceItemResponse[] : [] }
  }
  return { items: Array.isArray(record.items) ? record.items as MarketplaceItemResponse[] : [] }
}

function isVisibleMarketplaceItem(item: MarketplaceItemResponse): boolean {
  return item.enabled !== false && !RETIRED_MARKETPLACE_MODELS.has(item.model_name.trim().toLowerCase())
}

function syncDefaultSelections() {
  if (!providerOptions.value.some((option) => option.value === selectedProvider.value)) {
    selectedProvider.value = providerOptions.value[0]?.value || ''
  }
  if (!groupOptions.value.some((option) => option.value === selectedGroup.value)) {
    selectedGroup.value = groupOptions.value[0]?.value || ''
  }
}

async function loadMarketplacePaymentConfig(): Promise<PaymentConfig | null> {
  const token = localStorage.getItem('auth_token')
  if (!token) return null

  try {
    const response = await axios.get('/api/v1/payment/config', {
      timeout: 12000,
      withCredentials: true,
      headers: { Authorization: `Bearer ${token}` }
    })
    return unwrapPaymentConfig(response.data)
  } catch {
    return null
  }
}

function unwrapPaymentConfig(payload: unknown): PaymentConfig | null {
  if (!payload || typeof payload !== 'object') return null
  const record = payload as Record<string, unknown>
  const data = record.data
  if (data && typeof data === 'object') return data as PaymentConfig
  if ('usdt_cny_exchange_rate' in record || 'marketplace_group_multipliers' in record) {
    return record as unknown as PaymentConfig
  }
  return null
}

function normalizeMarketplaceGroupMultipliers(raw: Record<string, number> | undefined): Record<string, number> {
  const groupNames = groupOrder.value
  return groupNames.reduce<Record<string, number>>((acc, group) => {
    const value = Number(raw?.[group])
    acc[group] = Number.isFinite(value) && value > 0 ? value : DEFAULT_MARKETPLACE_GROUP_MULTIPLIERS[group] || 1
    return acc
  }, {})
}

onMounted(() => {
  initTheme()
  loadPricingConfig()
})
</script>
