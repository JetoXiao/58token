<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else
    class="gateway-page relative min-h-screen overflow-hidden bg-transparent text-slate-950 transition-colors duration-500 dark:bg-[#050509] dark:text-white"
  >
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_10%_12%,rgba(214,231,255,0.96)_0%,rgba(214,231,255,0.78)_11%,transparent_30%),radial-gradient(circle_at_49%_2%,rgba(228,220,255,0.95)_0%,rgba(228,220,255,0.6)_11%,transparent_28%),radial-gradient(circle_at_88%_12%,rgba(255,216,196,0.96)_0%,rgba(255,216,196,0.5)_12%,transparent_30%),radial-gradient(circle_at_14%_88%,rgba(215,242,255,0.88)_0%,rgba(215,242,255,0.36)_12%,transparent_30%),radial-gradient(circle_at_90%_90%,rgba(255,231,219,0.76)_0%,rgba(255,231,219,0.24)_10%,transparent_28%),linear-gradient(180deg,#fcfdff_0%,#f4f8ff_40%,#eef3fb_100%)] dark:bg-[radial-gradient(circle_at_10%_12%,rgba(56,189,248,0.2),transparent_30%),radial-gradient(circle_at_49%_2%,rgba(196,181,253,0.18),transparent_28%),radial-gradient(circle_at_88%_12%,rgba(251,146,60,0.14),transparent_26%),radial-gradient(circle_at_14%_88%,rgba(45,212,191,0.14),transparent_30%),linear-gradient(180deg,#050509_0%,#080a12_52%,#050509_100%)]"></div>
      <div class="absolute left-1/2 top-[42%] h-[34rem] w-[48rem] -translate-x-1/2 -translate-y-1/2 rounded-full bg-white/18 blur-[140px] dark:bg-white/[0.05]"></div>
      <div class="absolute inset-0 bg-[linear-gradient(rgba(148,163,184,0.022)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.022)_1px,transparent_1px)] bg-[size:72px_72px] opacity-16 [mask-image:radial-gradient(circle_at_top,black,transparent_74%)] dark:bg-[linear-gradient(rgba(255,255,255,0.04)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.04)_1px,transparent_1px)]"></div>
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
          @click="toggleTheme"
          class="rounded-xl border border-slate-200/70 bg-white/70 p-2 text-slate-600 transition hover:bg-white hover:text-slate-950 dark:border-white/10 dark:bg-white/[0.06] dark:text-slate-300 dark:hover:bg-white/12 dark:hover:text-white"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
        >
          <Icon v-if="isDark" name="sun" size="md" />
          <Icon v-else name="moon" size="md" />
        </button>
      </template>
    </MarketingNavbar>

    <main class="relative z-10">
      <section class="mx-auto grid min-h-[calc(100vh-6rem)] max-w-7xl items-center gap-12 px-4 pb-20 pt-12 sm:px-6 lg:grid-cols-[0.95fr_1.05fr] lg:pt-6">
        <div class="animate-fade-in">
          <div class="inline-flex items-center gap-2 rounded-full border border-primary-300/50 bg-white/70 px-3 py-1.5 text-xs font-semibold text-primary-700 shadow-sm backdrop-blur-xl dark:border-primary-400/30 dark:bg-white/[0.06] dark:text-primary-200 dark:shadow-[0_0_34px_rgba(59,130,246,0.18)]">
            <span class="h-1.5 w-1.5 rounded-full bg-primary-400 shadow-[0_0_14px_rgba(59,130,246,0.9)]"></span>
            {{ t('gateway.home.hero.eyebrow') }}
          </div>
          <h1 class="mt-7 max-w-4xl text-5xl font-semibold leading-[1.02] tracking-normal text-slate-950 dark:text-white sm:text-6xl lg:text-7xl">
            {{ t('gateway.home.hero.title') }}
          </h1>
          <p class="mt-6 max-w-2xl text-lg leading-8 text-slate-600 dark:text-slate-300">
            {{ t('gateway.home.hero.subtitle') }}
          </p>
          <div class="mt-8 flex flex-wrap gap-3">
            <RouterLink
              :to="isAuthenticated ? '/keys' : '/register'"
              class="inline-flex items-center justify-center gap-2 rounded-2xl bg-slate-950 px-5 py-3 text-sm font-semibold text-white shadow-[0_18px_50px_rgba(15,23,42,0.18)] transition hover:-translate-y-0.5 hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:shadow-[0_0_38px_rgba(255,255,255,0.18)]"
            >
              {{ t('gateway.home.hero.primary') }}
              <Icon name="arrowRight" size="sm" />
            </RouterLink>
            <RouterLink
              to="/available-channels"
              class="inline-flex items-center justify-center gap-2 rounded-2xl border border-slate-200/80 bg-white/70 px-5 py-3 text-sm font-semibold text-slate-700 shadow-sm backdrop-blur-xl transition hover:-translate-y-0.5 hover:border-cyan-300 hover:text-cyan-700 dark:border-white/10 dark:bg-white/[0.06] dark:text-slate-200 dark:hover:border-cyan-400/40 dark:hover:text-cyan-200"
            >
              {{ t('gateway.home.hero.secondary') }}
            </RouterLink>
          </div>
          <div class="mt-8 grid max-w-2xl grid-cols-2 gap-3 sm:grid-cols-4">
            <div v-for="metric in metrics" :key="metric.label" class="rounded-2xl border border-slate-200/70 bg-white/65 p-4 backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.045]">
              <div class="text-2xl font-semibold text-slate-950 dark:text-white">{{ metric.value }}</div>
              <div class="mt-1 text-xs leading-5 text-slate-500 dark:text-slate-400">{{ metric.label }}</div>
            </div>
          </div>
        </div>

        <div class="relative mx-auto w-full max-w-[680px]">
          <div class="routing-visual relative aspect-square rounded-[2rem] border border-slate-200/70 bg-white/55 shadow-[0_30px_120px_rgba(15,23,42,0.18)] backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045] dark:shadow-[0_0_120px_rgba(6,182,212,0.12)]">
            <div class="absolute inset-8 rounded-[1.5rem] border border-slate-200/50 dark:border-white/10"></div>
            <div class="absolute left-5 top-5 z-10 rounded-2xl border border-slate-200/70 bg-white/70 px-3 py-2 text-[11px] font-semibold text-slate-600 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/45 dark:text-slate-300">
              <span class="mr-2 inline-block h-1.5 w-1.5 rounded-full bg-primary-400 shadow-[0_0_10px_rgba(59,130,246,0.9)]"></span>
              {{ t('gateway.home.routeVisual.ingress') }}
            </div>
            <div class="absolute right-5 top-5 z-10 rounded-2xl border border-cyan-200/80 bg-cyan-50/70 px-3 py-2 text-[11px] font-semibold text-cyan-800 shadow-sm backdrop-blur-xl dark:border-cyan-400/20 dark:bg-cyan-400/10 dark:text-cyan-100">
              {{ t('gateway.home.routeVisual.latency') }}
            </div>
            <div class="absolute bottom-5 left-5 z-10 rounded-2xl border border-accent-200/80 bg-accent-50/70 px-3 py-2 text-[11px] font-semibold text-accent-800 shadow-sm backdrop-blur-xl dark:border-accent-400/20 dark:bg-accent-400/10 dark:text-accent-100">
              {{ t('gateway.home.routeVisual.policy') }}
            </div>
            <div class="absolute bottom-5 right-5 z-10 rounded-2xl border border-slate-200/70 bg-white/70 px-3 py-2 text-[11px] font-semibold text-slate-600 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/45 dark:text-slate-300">
              {{ t('gateway.home.routeVisual.failover') }}
            </div>

            <div class="request-stack absolute left-1/2 top-[18%] z-10 w-44 -translate-x-1/2 rounded-2xl border border-slate-200/70 bg-white/72 p-3 shadow-[0_18px_50px_rgba(15,23,42,0.10)] backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/40">
              <div class="mb-2 flex items-center justify-between text-[10px] font-semibold text-slate-500 dark:text-slate-400">
                <span>POST /v1/chat</span>
                <span>200</span>
              </div>
              <div class="space-y-1.5">
                <span class="block h-1.5 w-11/12 rounded-full bg-slate-200 dark:bg-white/12"></span>
                <span class="block h-1.5 w-8/12 rounded-full bg-cyan-300/70"></span>
                <span class="block h-1.5 w-10/12 rounded-full bg-accent-300/70"></span>
              </div>
            </div>

            <div class="absolute left-1/2 top-1/2 z-20 flex h-36 w-36 -translate-x-1/2 -translate-y-1/2 flex-col items-center justify-center rounded-[1.75rem] border border-cyan-300/60 bg-slate-950 text-white shadow-[0_0_58px_rgba(6,182,212,0.34)] dark:border-cyan-300/40 dark:bg-white/[0.08]">
              <Icon name="server" size="lg" class="text-cyan-300" />
              <span class="mt-3 text-sm font-semibold">API Gateway</span>
              <span class="mt-1 text-[11px] text-cyan-100/80">OpenAI-compatible</span>
              <span class="absolute -top-2 h-3 w-3 rounded-full bg-cyan-300 shadow-[0_0_18px_rgba(103,232,249,0.95)]"></span>
              <span class="absolute -bottom-2 h-3 w-3 rounded-full bg-cyan-300 shadow-[0_0_18px_rgba(103,232,249,0.95)]"></span>
            </div>

            <div v-for="node in modelNodes" :key="node.name" :class="['model-node', node.position]">
              <span :class="['node-mark', node.color]">{{ node.short }}</span>
              <span class="text-sm font-semibold">{{ node.name }}</span>
              <span class="text-[11px] text-slate-500 dark:text-slate-400">{{ node.caption }}</span>
            </div>

            <svg class="absolute inset-0 h-full w-full" viewBox="0 0 600 600" aria-hidden="true">
              <defs>
                <linearGradient id="routeGlow" x1="0" x2="1">
                  <stop offset="0%" stop-color="#7C3AED" />
                  <stop offset="100%" stop-color="#06B6D4" />
                </linearGradient>
                <linearGradient id="routeSoft" x1="0" x2="1">
                  <stop offset="0%" stop-color="#A78BFA" stop-opacity=".36" />
                  <stop offset="100%" stop-color="#67E8F9" stop-opacity=".36" />
                </linearGradient>
              </defs>
              <circle cx="300" cy="300" r="218" class="orbit-line" />
              <circle cx="300" cy="300" r="142" class="orbit-line orbit-line-inner" />
              <path d="M106 114 C184 170 216 222 300 300" class="route-soft-line" />
              <path d="M300 300 C372 230 428 204 496 112" class="route-soft-line" />
              <path d="M300 300 C382 380 430 418 498 500" class="route-soft-line" />
              <path d="M300 300 C222 368 168 418 102 496" class="route-soft-line" />
              <path v-for="path in routePaths" :key="path" :d="path" class="route-line" />
            </svg>

            <div class="packet packet-1"></div>
            <div class="packet packet-2"></div>
            <div class="packet packet-3"></div>
            <div class="packet packet-4"></div>
            <div class="packet packet-5"></div>
            <div class="packet packet-6"></div>
          </div>
        </div>
      </section>

      <section class="mx-auto max-w-7xl px-4 py-10 sm:px-6">
        <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <article v-for="feature in features" :key="feature.title" class="group rounded-2xl border border-slate-200/70 bg-white/70 p-5 backdrop-blur-2xl transition duration-300 hover:-translate-y-1 hover:border-primary-300 hover:shadow-[0_20px_70px_rgba(59,130,246,0.14)] dark:border-white/10 dark:bg-white/[0.045] dark:hover:border-cyan-400/40 dark:hover:shadow-[0_0_55px_rgba(6,182,212,0.12)]">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-white text-cyan-600 dark:border-white/10 dark:bg-white/[0.06] dark:text-cyan-300">
              <Icon :name="feature.icon" size="md" />
            </div>
            <h3 class="mt-5 text-base font-semibold text-slate-950 dark:text-white">{{ feature.title }}</h3>
            <p class="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400">{{ feature.copy }}</p>
          </article>
        </div>
      </section>

      <section class="mx-auto grid max-w-7xl gap-6 px-4 py-14 sm:px-6 lg:grid-cols-[0.9fr_1.1fr] lg:items-center">
        <div>
          <p class="text-sm font-semibold uppercase tracking-[0.22em] text-cyan-600 dark:text-cyan-300">{{ t('gateway.home.sdk.eyebrow') }}</p>
          <h2 class="mt-4 text-3xl font-semibold tracking-normal text-slate-950 dark:text-white sm:text-5xl">{{ t('gateway.home.sdk.title') }}</h2>
          <p class="mt-5 max-w-xl text-base leading-8 text-slate-600 dark:text-slate-300">{{ t('gateway.home.sdk.copy') }}</p>
        </div>
        <div class="overflow-hidden rounded-2xl border border-slate-200/80 bg-slate-950 shadow-[0_28px_90px_rgba(15,23,42,0.22)] dark:border-white/10">
          <div class="flex items-center justify-between border-b border-white/10 px-4 py-3">
            <div class="flex gap-2">
              <span class="h-3 w-3 rounded-full bg-red-400"></span>
              <span class="h-3 w-3 rounded-full bg-accent-300"></span>
              <span class="h-3 w-3 rounded-full bg-primary-400"></span>
            </div>
            <button class="rounded-lg border border-white/10 bg-white/10 px-3 py-1.5 text-xs font-semibold text-slate-100 transition hover:bg-white/15" @click="copyCode">
              {{ copied ? t('common.copied') : t('common.copy') }}
            </button>
          </div>
          <pre class="overflow-x-auto p-5 text-sm leading-7 text-slate-100"><code>{{ sdkCode }}</code></pre>
        </div>
      </section>

      <section class="mx-auto max-w-7xl px-4 py-14 sm:px-6">
        <div class="grid gap-6 lg:grid-cols-[1fr_1.2fr]">
          <div class="rounded-[2rem] border border-slate-200/70 bg-white/70 p-6 backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045]">
            <p class="text-sm font-semibold uppercase tracking-[0.22em] text-primary-600 dark:text-primary-300">{{ t('gateway.home.dashboard.eyebrow') }}</p>
            <h2 class="mt-4 text-3xl font-semibold tracking-normal text-slate-950 dark:text-white">{{ t('gateway.home.dashboard.title') }}</h2>
            <p class="mt-4 text-sm leading-7 text-slate-600 dark:text-slate-400">{{ t('gateway.home.dashboard.copy') }}</p>
          </div>
          <div class="rounded-[2rem] border border-slate-200/70 bg-white/75 p-4 shadow-[0_26px_90px_rgba(15,23,42,0.12)] backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.05]">
            <div class="grid gap-3 sm:grid-cols-4">
              <div v-for="card in dashboardCards" :key="card.label" class="rounded-2xl border border-slate-200/70 bg-white/70 p-4 dark:border-white/10 dark:bg-slate-950/35">
                <div class="text-xs text-slate-500 dark:text-slate-400">{{ card.label }}</div>
                <div class="mt-2 text-xl font-semibold text-slate-950 dark:text-white">{{ card.value }}</div>
              </div>
            </div>
            <div class="mt-4 grid gap-4 lg:grid-cols-[1fr_0.8fr]">
              <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-4 dark:border-white/10 dark:bg-slate-950/35">
                <div class="mb-4 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
                  <span>{{ t('gateway.home.dashboard.traffic') }}</span>
                  <span>24h</span>
                </div>
                <div class="flex h-36 items-end gap-2">
                  <span v-for="bar in trafficBars" :key="bar" class="flex-1 rounded-t-lg bg-gradient-to-t from-primary-500 to-cyan-400 opacity-80" :style="{ height: `${bar}%` }"></span>
                </div>
              </div>
              <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-4 dark:border-white/10 dark:bg-slate-950/35">
                <div class="text-xs text-slate-500 dark:text-slate-400">{{ t('gateway.home.dashboard.models') }}</div>
                <div class="mt-4 space-y-3">
                  <div v-for="model in modelShare" :key="model.name">
                    <div class="flex justify-between text-xs"><span>{{ model.name }}</span><span>{{ model.value }}%</span></div>
                    <div class="mt-1 h-1.5 rounded-full bg-slate-200 dark:bg-white/10">
                      <div class="h-full rounded-full bg-gradient-to-r from-primary-500 to-cyan-400" :style="{ width: `${model.value}%` }"></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="mx-auto max-w-7xl px-4 py-14 sm:px-6">
        <div class="rounded-[2rem] border border-slate-200/70 bg-white/70 p-8 text-center shadow-[0_26px_90px_rgba(15,23,42,0.10)] backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045] dark:shadow-[0_0_90px_rgba(124,58,237,0.12)]">
          <p class="text-sm font-semibold uppercase tracking-[0.22em] text-cyan-600 dark:text-cyan-300">{{ t('gateway.home.pricing.eyebrow') }}</p>
          <h2 class="mt-4 text-3xl font-semibold tracking-normal text-slate-950 dark:text-white sm:text-5xl">{{ t('gateway.home.pricing.title') }}</h2>
          <p class="mx-auto mt-5 max-w-2xl text-base leading-8 text-slate-600 dark:text-slate-300">{{ t('gateway.home.pricing.copy') }}</p>
          <div class="mt-7 flex flex-wrap justify-center gap-3">
            <span v-for="item in pricingPills" :key="item" class="rounded-full border border-slate-200 bg-white/70 px-4 py-2 text-sm font-medium text-slate-600 dark:border-white/10 dark:bg-white/[0.06] dark:text-slate-300">{{ item }}</span>
          </div>
        </div>
      </section>

      <section class="mx-auto max-w-5xl px-4 py-20 text-center sm:px-6">
        <h2 class="text-4xl font-semibold tracking-normal text-slate-950 dark:text-white sm:text-6xl">{{ t('gateway.home.cta.title') }}</h2>
        <p class="mx-auto mt-5 max-w-2xl text-base leading-8 text-slate-600 dark:text-slate-300">{{ t('gateway.home.cta.copy') }}</p>
        <RouterLink
          :to="isAuthenticated ? '/keys' : '/register'"
          class="mt-8 inline-flex items-center justify-center gap-2 rounded-2xl bg-slate-950 px-6 py-3 text-sm font-semibold text-white shadow-[0_18px_50px_rgba(15,23,42,0.18)] transition hover:-translate-y-0.5 hover:bg-slate-800 dark:bg-white dark:text-slate-950"
        >
          {{ t('gateway.home.cta.button') }}
          <Icon name="key" size="sm" />
        </RouterLink>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import MarketingNavbar from '@/components/marketing/MarketingNavbar.vue'
import { BRAND_LOGO_URL } from '@/constants/brand'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || '58Token')
const siteLogo = computed(() => appStore.siteLogo || BRAND_LOGO_URL)
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const marketingNavItems = computed(() => appStore.cachedPublicSettings?.marketing_nav_items)
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})
const isDark = ref(document.documentElement.classList.contains('dark'))
const copied = ref(false)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')

const metrics = computed(() => [
  { value: '99.99%', label: t('gateway.home.metrics.availability') },
  { value: '100+', label: t('gateway.home.metrics.models') },
  { value: 'OpenAI', label: t('gateway.home.metrics.compatible') },
  { value: 'PayGo', label: t('gateway.home.metrics.paygo') }
])

const modelNodes = [
  { name: 'GPT', short: 'G', caption: 'OpenAI', color: 'from-primary-400 to-cyan-400', position: 'node-top' },
  { name: 'Claude', short: 'C', caption: 'Anthropic', color: 'from-accent-300 to-primary-400', position: 'node-right' },
  { name: 'Gemini', short: 'G', caption: 'Google', color: 'from-cyan-300 to-primary-500', position: 'node-bottom' },
  { name: 'DeepSeek', short: 'D', caption: 'DeepSeek', color: 'from-cyan-300 to-primary-500', position: 'node-left' }
]
const routePaths = [
  'M300 300 C300 210 300 160 300 92',
  'M300 300 C390 300 442 260 508 190',
  'M300 300 C300 390 300 442 300 508',
  'M300 300 C210 300 160 340 92 410'
]
const sdkCode = `const client = new OpenAI({
  apiKey: process.env.API_KEY,
  baseURL: "https://api.58token.vip/v1"
})`
const features = computed<Array<{ title: string; copy: string; icon: 'key' | 'sync' | 'chart' | 'shield' }>>(() => [
  { title: t('gateway.home.features.key.title'), copy: t('gateway.home.features.key.copy'), icon: 'key' },
  { title: t('gateway.home.features.routing.title'), copy: t('gateway.home.features.routing.copy'), icon: 'sync' },
  { title: t('gateway.home.features.analytics.title'), copy: t('gateway.home.features.analytics.copy'), icon: 'chart' },
  { title: t('gateway.home.features.failover.title'), copy: t('gateway.home.features.failover.copy'), icon: 'shield' }
])
const dashboardCards = computed(() => [
  { label: t('gateway.home.dashboard.tokens'), value: '18.4M' },
  { label: t('gateway.home.dashboard.spend'), value: '$42.18' },
  { label: t('gateway.home.dashboard.requests'), value: '128K' },
  { label: t('gateway.home.dashboard.errorRate'), value: '0.04%' }
])
const trafficBars = [42, 58, 46, 64, 78, 62, 84, 70, 92, 74, 88, 96]
const modelShare = [
  { name: 'GPT', value: 42 },
  { name: 'Claude', value: 31 },
  { name: 'Gemini', value: 18 },
  { name: 'DeepSeek', value: 9 }
]
const pricingPills = computed(() => [
  t('gateway.home.pricing.simple'),
  t('gateway.home.pricing.transparent'),
  t('gateway.home.pricing.noLockin')
])

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

async function copyCode() {
  await navigator.clipboard.writeText(sdkCode)
  copied.value = true
  window.setTimeout(() => { copied.value = false }, 1500)
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
})
</script>

<style scoped>
.gateway-page {
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Noto Sans SC", sans-serif;
}

.routing-visual::before {
  content: "";
  position: absolute;
  inset: 10%;
  border-radius: 999px;
  background: conic-gradient(from 120deg, rgba(124, 58, 237, 0.22), rgba(6, 182, 212, 0.28), rgba(124, 58, 237, 0.22));
  filter: blur(42px);
  animation: glowPulse 5s ease-in-out infinite;
}

.model-node {
  position: absolute;
  z-index: 18;
  display: flex;
  min-width: 128px;
  flex-direction: column;
  gap: 2px;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.32);
  background: rgba(255, 255, 255, 0.76);
  padding: 14px;
  color: rgb(15 23 42);
  box-shadow: 0 18px 54px rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(18px);
  animation: floatNode 5s ease-in-out infinite;
}

:global(.dark) .model-node {
  border-color: rgba(255,255,255,0.12);
  background: rgba(255,255,255,0.07);
  color: white;
}

.node-mark {
  display: flex;
  height: 34px;
  width: 34px;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  background-image: linear-gradient(135deg, var(--tw-gradient-stops));
  color: white;
  font-size: 12px;
  font-weight: 800;
  box-shadow: 0 0 24px rgba(6, 182, 212, 0.24);
}

.node-top { left: 50%; top: 7%; transform: translateX(-50%); }
.node-right { right: 5%; top: 28%; animation-delay: .4s; }
.node-bottom { left: 50%; bottom: 7%; transform: translateX(-50%); animation-delay: .8s; }
.node-left { left: 5%; top: 58%; animation-delay: 1.2s; }

.route-line {
  fill: none;
  stroke: url(#routeGlow);
  stroke-width: 1.7;
  stroke-dasharray: 8 12;
  opacity: .8;
  animation: routeFlow 2.8s linear infinite;
  filter: drop-shadow(0 0 8px rgba(6, 182, 212, .45));
}

.route-soft-line {
  fill: none;
  stroke: url(#routeSoft);
  stroke-width: 1.1;
  stroke-dasharray: 3 16;
  animation: routeFlow 7s linear infinite reverse;
}

.orbit-line {
  fill: none;
  stroke: rgba(14, 165, 233, .18);
  stroke-width: 1;
  stroke-dasharray: 2 12;
}

.orbit-line-inner {
  stroke: rgba(124, 58, 237, .18);
  animation: routeFlow 8s linear infinite;
}

.request-stack {
  animation: floatPanel 6s ease-in-out infinite;
}

.packet {
  position: absolute;
  z-index: 24;
  height: 8px;
  width: 8px;
  border-radius: 999px;
  background: #67e8f9;
  box-shadow: 0 0 22px rgba(103, 232, 249, 0.95);
}
.packet-1 { animation: packetTop 3s linear infinite; }
.packet-2 { animation: packetRight 3.4s linear infinite; animation-delay: .45s; }
.packet-3 { animation: packetBottom 3.2s linear infinite; animation-delay: .9s; }
.packet-4 { animation: packetLeft 3.7s linear infinite; animation-delay: 1.2s; }
.packet-5 { animation: packetArcA 4.2s linear infinite; animation-delay: .2s; background: #a78bfa; box-shadow: 0 0 22px rgba(167, 139, 250, .9); }
.packet-6 { animation: packetArcB 4.6s linear infinite; animation-delay: 1.6s; background: #34d399; box-shadow: 0 0 22px rgba(52, 211, 153, .85); }

.token-particle {
  position: absolute;
  height: 5px;
  width: 5px;
  border-radius: 999px;
  background: rgba(6, 182, 212, .8);
  box-shadow: 0 0 18px rgba(6, 182, 212, .7);
}
.particle-a { left: 12%; top: 22%; animation: drift 10s linear infinite; }
.particle-b { right: 18%; top: 38%; animation: drift 13s linear infinite reverse; }
.particle-c { left: 42%; bottom: 18%; animation: drift 11s linear infinite; }

@keyframes routeFlow { to { stroke-dashoffset: -40; } }
@keyframes floatNode { 0%,100% { margin-top: 0; } 50% { margin-top: -9px; } }
@keyframes floatPanel { 0%,100% { transform: translate(-50%, 0); } 50% { transform: translate(-50%, -8px); } }
@keyframes glowPulse { 0%,100% { opacity: .75; transform: scale(.96); } 50% { opacity: 1; transform: scale(1.05); } }
@keyframes drift { from { transform: translate3d(0,0,0); } to { transform: translate3d(90px, -120px, 0); opacity: .1; } }
@keyframes packetTop { from { left: 50%; top: 50%; } to { left: 50%; top: 18%; } }
@keyframes packetRight { from { left: 50%; top: 50%; } to { left: 80%; top: 31%; } }
@keyframes packetBottom { from { left: 50%; top: 50%; } to { left: 50%; top: 79%; } }
@keyframes packetLeft { from { left: 50%; top: 50%; } to { left: 20%; top: 66%; } }
@keyframes packetArcA { 0% { left: 18%; top: 20%; opacity: 0; } 15% { opacity: 1; } 60% { left: 50%; top: 50%; opacity: 1; } 100% { left: 82%; top: 23%; opacity: 0; } }
@keyframes packetArcB { 0% { left: 84%; top: 78%; opacity: 0; } 15% { opacity: 1; } 58% { left: 50%; top: 50%; opacity: 1; } 100% { left: 17%; top: 72%; opacity: 0; } }

@media (max-width: 640px) {
  .model-node { min-width: 104px; padding: 10px; }
  .node-right { right: 2%; }
  .node-left { left: 2%; }
  .request-stack { top: 16%; width: 9.5rem; }
}
</style>
