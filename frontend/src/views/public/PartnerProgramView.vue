<template>
  <div class="relative flex min-h-screen flex-col overflow-hidden bg-[#f7f8fb] text-slate-950 dark:bg-[#05060a] dark:text-white">
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute inset-0 bg-[linear-gradient(rgba(15,23,42,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(15,23,42,0.045)_1px,transparent_1px)] bg-[size:72px_72px] [mask-image:linear-gradient(to_bottom,black,transparent_82%)] dark:bg-[linear-gradient(rgba(255,255,255,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.045)_1px,transparent_1px)]"></div>
      <div class="absolute inset-x-0 top-0 h-72 bg-gradient-to-b from-emerald-100/70 via-white/20 to-transparent dark:from-emerald-950/35 dark:via-white/[0.03]"></div>
    </div>

    <MarketingNavbar
      :site-name="siteName"
      :subtitle="t('gateway.common.navSubtitle')"
      :logo="siteLogo"
      :doc-url="docUrl"
      docs-to="/docs"
      :docs-label="t('gateway.common.docs')"
      :cta-to="isAuthenticated ? dashboardPath : '/login'"
      :cta-label="isAuthenticated ? t('gateway.common.dashboard') : t('gateway.common.login')"
      model-marketplace-to="/available-channels"
      :model-marketplace-label="t('gateway.common.models')"
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
          class="h-10 rounded-xl border border-slate-200/70 bg-white/70 p-2 text-slate-600 transition hover:bg-white hover:text-slate-950 dark:border-white/10 dark:bg-white/[0.06] dark:text-slate-300 dark:hover:bg-white/12 dark:hover:text-white"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @click="toggleTheme"
        >
          <Icon v-if="isDark" name="sun" size="md" />
          <Icon v-else name="moon" size="md" />
        </button>
      </template>
    </MarketingNavbar>

    <main class="relative z-10 flex-1 px-4 pb-16 pt-6 sm:px-6">
      <section class="mx-auto grid max-w-7xl gap-8 rounded-[2rem] border border-slate-200/70 bg-white/72 p-5 shadow-[0_26px_90px_rgba(15,23,42,0.10)] backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045] dark:shadow-[0_26px_90px_rgba(0,0,0,0.30)] md:p-8 lg:grid-cols-[minmax(0,1fr)_430px] lg:items-center">
        <div>
          <div class="inline-flex items-center gap-2 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-900/20 dark:text-emerald-300">
            <Icon name="sparkles" size="sm" />
            {{ t('gateway.partner.eyebrow') }}
          </div>
          <h1 class="mt-5 max-w-4xl text-4xl font-semibold tracking-normal text-slate-950 dark:text-white md:text-6xl">
            {{ t('gateway.partner.title') }}
          </h1>
          <p class="mt-5 max-w-3xl text-base leading-8 text-slate-600 dark:text-slate-300">
            {{ t('gateway.partner.subtitle') }}
          </p>
          <div class="mt-7 flex flex-wrap gap-3">
            <RouterLink
              :to="primaryCtaTo"
              class="inline-flex items-center justify-center gap-2 rounded-2xl bg-slate-950 px-5 py-3 text-sm font-semibold text-white shadow-[0_18px_50px_rgba(15,23,42,0.18)] transition hover:-translate-y-0.5 hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
            >
              {{ t('gateway.partner.primaryCta') }}
              <Icon name="arrowRight" size="sm" />
            </RouterLink>
            <RouterLink
              to="/docs"
              class="inline-flex items-center justify-center rounded-2xl border border-slate-200/80 bg-white/75 px-5 py-3 text-sm font-semibold text-slate-700 shadow-sm backdrop-blur-xl transition hover:border-emerald-300 hover:text-emerald-700 dark:border-white/10 dark:bg-white/[0.05] dark:text-slate-200 dark:hover:border-emerald-700 dark:hover:text-emerald-300"
            >
              {{ t('gateway.partner.secondaryCta') }}
            </RouterLink>
          </div>
        </div>

        <div class="rounded-[1.5rem] border border-emerald-200/80 bg-emerald-50/65 p-5 shadow-sm backdrop-blur dark:border-emerald-900/50 dark:bg-emerald-900/15">
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700 dark:text-emerald-300">{{ t('gateway.partner.rateCard.eyebrow') }}</p>
          <div class="mt-4 text-4xl font-semibold leading-tight tracking-normal text-slate-950 dark:text-white md:text-5xl">{{ t('gateway.partner.rateCard.headline') }}</div>
          <p class="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-300">{{ t('gateway.partner.rateCard.copy') }}</p>
          <div class="mt-5 grid grid-cols-2 gap-3">
            <div v-for="stat in stats" :key="stat.label" class="rounded-2xl border border-white/70 bg-white/70 p-4 dark:border-white/10 dark:bg-white/[0.05]">
              <div class="text-2xl font-semibold text-slate-950 dark:text-white">{{ stat.value }}</div>
              <div class="mt-1 text-xs text-slate-500 dark:text-slate-400">{{ stat.label }}</div>
            </div>
          </div>
          <div class="mt-4 rounded-2xl border border-emerald-200 bg-white/65 p-4 text-sm leading-6 text-slate-600 dark:border-emerald-900/50 dark:bg-white/[0.04] dark:text-slate-300">
            {{ t('gateway.partner.rateCard.note') }}
          </div>
        </div>
      </section>

      <section class="mx-auto mt-6 grid max-w-7xl gap-4 md:grid-cols-3">
        <article v-for="item in valueProps" :key="item.title" class="rounded-[1.5rem] border border-slate-200/70 bg-white/75 p-5 shadow-sm backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045]">
          <div class="flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-white text-emerald-600 dark:border-white/10 dark:bg-white/[0.06] dark:text-emerald-300">
            <Icon :name="item.icon" size="md" />
          </div>
          <h2 class="mt-5 text-lg font-semibold text-slate-950 dark:text-white">{{ item.title }}</h2>
          <p class="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400">{{ item.copy }}</p>
        </article>
      </section>

      <section class="mx-auto mt-6 max-w-7xl rounded-[2rem] border border-slate-200/70 bg-white/75 p-5 shadow-sm backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045] md:p-8">
        <div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-600 dark:text-emerald-300">{{ t('gateway.partner.tiers.eyebrow') }}</p>
            <h2 class="mt-3 text-3xl font-semibold tracking-normal text-slate-950 dark:text-white">{{ t('gateway.partner.tiers.title') }}</h2>
          </div>
          <p class="max-w-xl text-sm leading-6 text-slate-600 dark:text-slate-400">{{ t('gateway.partner.tiers.copy') }}</p>
        </div>
        <div class="mt-6 grid gap-3 md:grid-cols-4">
          <div v-for="tier in tiers" :key="tier.name" class="rounded-2xl border border-slate-200 bg-white/80 p-4 dark:border-white/10 dark:bg-slate-950/35">
            <div class="text-sm font-semibold text-slate-950 dark:text-white">{{ tier.name }}</div>
            <div class="mt-3 text-lg font-semibold text-emerald-600 dark:text-emerald-300">{{ tier.benefit }}</div>
            <p class="mt-2 text-xs font-semibold text-slate-700 dark:text-slate-200">{{ tier.requirement }}</p>
            <p class="mt-2 text-xs leading-5 text-slate-500 dark:text-slate-400">{{ tier.copy }}</p>
          </div>
        </div>
      </section>

      <section class="mx-auto mt-6 grid max-w-7xl gap-6 lg:grid-cols-[0.95fr_1.05fr]">
        <div class="rounded-[2rem] border border-slate-200/70 bg-white/75 p-6 shadow-sm backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045]">
          <p class="text-sm font-semibold uppercase tracking-[0.18em] text-violet-600 dark:text-violet-300">{{ t('gateway.partner.flow.eyebrow') }}</p>
          <h2 class="mt-3 text-3xl font-semibold tracking-normal text-slate-950 dark:text-white">{{ t('gateway.partner.flow.title') }}</h2>
          <div class="mt-6 space-y-4">
            <div v-for="(step, index) in steps" :key="step.title" class="flex gap-4">
              <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-slate-950 text-sm font-semibold text-white dark:bg-white dark:text-slate-950">{{ index + 1 }}</div>
              <div>
                <h3 class="text-base font-semibold text-slate-950 dark:text-white">{{ step.title }}</h3>
                <p class="mt-1 text-sm leading-6 text-slate-600 dark:text-slate-400">{{ step.copy }}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-[2rem] border border-slate-200/70 bg-slate-950 p-6 text-white shadow-[0_26px_90px_rgba(15,23,42,0.16)] dark:border-white/10 dark:bg-white/[0.06]">
          <p class="text-sm font-semibold uppercase tracking-[0.18em] text-emerald-300">{{ t('gateway.partner.settlement.eyebrow') }}</p>
          <h2 class="mt-3 text-3xl font-semibold tracking-normal">{{ t('gateway.partner.settlement.title') }}</h2>
          <div class="mt-6 grid gap-3 sm:grid-cols-3">
            <div v-for="item in settlementItems" :key="item.title" class="rounded-2xl border border-white/10 bg-white/10 p-4">
              <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-white/10 text-emerald-200">
                <Icon :name="item.icon" size="sm" />
              </div>
              <div class="mt-4 text-base font-semibold">{{ item.title }}</div>
              <p class="mt-2 text-xs leading-5 text-slate-300">{{ item.copy }}</p>
            </div>
          </div>
          <p class="mt-5 text-sm leading-7 text-slate-300">{{ t('gateway.partner.settlement.copy') }}</p>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import MarketingNavbar from '@/components/marketing/MarketingNavbar.vue'
import { BRAND_LOGO_URL } from '@/constants/brand'
import userAPI from '@/api/user'
import type { AffiliatePartnerLevel, AffiliatePartnerTier } from '@/types'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const isDark = ref(document.documentElement.classList.contains('dark'))
const partnerTiers = ref<AffiliatePartnerTier[]>([])

const fallbackPartnerTiers: AffiliatePartnerTier[] = [
  { level: 'spark', name: 'Spark', rebate_rate_percent: 40, required_invitees: 10, next_required_invitees: 30 },
  { level: 'voyage', name: 'Voyage', rebate_rate_percent: 50, required_invitees: 30, next_required_invitees: 50 },
  { level: 'summit', name: 'Summit', rebate_rate_percent: 60, required_invitees: 50, next_required_invitees: 100 },
  { level: 'cocreate', name: 'Co-create', rebate_rate_percent: 70, required_invitees: 100, next_required_invitees: null }
]

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'UseAiForMe')
const siteLogo = computed(() => appStore.siteLogo || BRAND_LOGO_URL)
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const marketingNavItems = computed(() => appStore.cachedPublicSettings?.marketing_nav_items)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const primaryCtaTo = computed(() => isAuthenticated.value ? '/affiliate' : '/register')
const displayPartnerTiers = computed(() => partnerTiers.value.length > 0 ? partnerTiers.value : fallbackPartnerTiers)

const stats = computed(() => [
  { value: t('gateway.partner.stats.settlementValue'), label: t('gateway.partner.stats.settlement') },
  { value: t('gateway.partner.stats.lifetimeValue'), label: t('gateway.partner.stats.lifetime') }
])
const valueProps = computed<Array<{ title: string; copy: string; icon: 'dollar' | 'chart' | 'users' }>>(() => [
  { title: t('gateway.partner.valueProps.highRate.title'), copy: t('gateway.partner.valueProps.highRate.copy'), icon: 'dollar' },
  { title: t('gateway.partner.valueProps.realUsage.title'), copy: t('gateway.partner.valueProps.realUsage.copy'), icon: 'chart' },
  { title: t('gateway.partner.valueProps.easyShare.title'), copy: t('gateway.partner.valueProps.easyShare.copy'), icon: 'users' }
])
const tiers = computed(() => displayPartnerTiers.value.map((tier) => ({
  name: partnerLevelLabel(tier.level),
  benefit: t(`gateway.partner.tiers.items.${tier.level}.benefit`),
  requirement: t('affiliate.partner.tierRequirement', { count: tier.required_invitees }),
  copy: t(`gateway.partner.tiers.items.${tier.level}.copy`)
})))
const steps = computed(() => ['apply', 'invite', 'settle'].map((key) => ({
  title: t(`gateway.partner.flow.steps.${key}.title`),
  copy: t(`gateway.partner.flow.steps.${key}.copy`)
})))
const settlementItems = computed<Array<{ title: string; copy: string; icon: 'chart' | 'sync' | 'shield' }>>(() => [
  { title: t('gateway.partner.settlement.items.usage.title'), copy: t('gateway.partner.settlement.items.usage.copy'), icon: 'chart' },
  { title: t('gateway.partner.settlement.items.review.title'), copy: t('gateway.partner.settlement.items.review.copy'), icon: 'shield' },
  { title: t('gateway.partner.settlement.items.payout.title'), copy: t('gateway.partner.settlement.items.payout.copy'), icon: 'sync' }
])

function partnerLevelLabel(level: AffiliatePartnerLevel): string {
  return t(`affiliate.partner.levels.${level}`)
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  isDark.value = savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', isDark.value)
}

async function loadPartnerTiers() {
  try {
    const tiers = await userAPI.getPublicAffiliatePartnerTiers()
    partnerTiers.value = tiers.length > 0 ? tiers : fallbackPartnerTiers
  } catch (error) {
    partnerTiers.value = fallbackPartnerTiers
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
  loadPartnerTiers()
})
</script>
