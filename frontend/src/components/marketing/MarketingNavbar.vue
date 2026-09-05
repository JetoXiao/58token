<template>
  <header class="marketing-navbar relative z-20 px-4 py-4 sm:px-6">
    <nav
      class="mx-auto flex max-w-7xl items-center justify-between rounded-2xl border border-sky-100/80 bg-white/78 px-4 py-3 shadow-[0_18px_70px_rgba(15,23,42,0.06)] backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.04] dark:shadow-[0_18px_70px_rgba(0,0,0,0.28)]"
    >
      <router-link to="/home" class="flex min-w-0 items-center gap-3">
        <span class="flex h-9 w-9 items-center justify-center overflow-hidden rounded-xl border border-sky-100/80 bg-sky-50/70 dark:border-white/10 dark:bg-white/10">
          <img :src="siteLogo" alt="" class="h-full w-full object-contain" />
        </span>
        <span class="hidden min-w-0 sm:block">
          <span class="block truncate text-sm font-semibold tracking-tight text-slate-900 dark:text-white">{{ siteName }}</span>
          <span class="block truncate text-xs text-slate-500 dark:text-slate-400">{{ subtitle }}</span>
        </span>
      </router-link>

      <div class="flex items-center gap-2">
        <router-link
          v-if="modelMarketplaceTo && showModels"
          :to="modelMarketplaceTo"
          custom
          v-slot="{ href, navigate, isActive }"
        >
          <a
            :href="href"
            @click="navigate"
            :class="[
              'inline-flex h-10 max-w-[8.5rem] items-center gap-1.5 truncate rounded-xl border px-3 py-2 text-xs font-semibold transition sm:max-w-none sm:px-3.5 sm:text-sm',
              isActive
                ? 'border-sky-200 bg-sky-50 text-sky-700 shadow-[0_10px_28px_rgba(125,211,252,0.14)] dark:border-sky-400/40 dark:bg-sky-400/10 dark:text-sky-100 dark:shadow-[0_0_28px_rgba(125,211,252,0.12)]'
                : 'border-transparent text-slate-600 hover:border-sky-100 hover:bg-sky-50 hover:text-slate-900 dark:text-slate-300 dark:hover:border-white/10 dark:hover:bg-white/10 dark:hover:text-white',
            ]"
          >
            <span
              :class="[
                'h-1.5 w-1.5 shrink-0 rounded-full',
                isActive ? 'bg-sky-400 dark:bg-sky-300' : 'bg-slate-300 dark:bg-slate-600',
              ]"
            ></span>
            <span class="truncate">{{ modelMarketplaceLabel }}</span>
          </a>
        </router-link>
        <router-link
          v-if="docsTo && showDocs"
          :to="docsTo"
          custom
          v-slot="{ href, navigate, isActive }"
        >
          <a
            :href="href"
            @click="navigate"
            :class="[
              'inline-flex h-10 max-w-[8.5rem] items-center gap-1.5 truncate rounded-xl border px-3 py-2 text-xs font-semibold transition sm:max-w-none sm:px-3.5 sm:text-sm',
              isActive
                ? 'border-sky-200 bg-sky-50 text-sky-700 shadow-[0_10px_28px_rgba(125,211,252,0.14)] dark:border-sky-400/40 dark:bg-sky-400/10 dark:text-sky-100 dark:shadow-[0_0_28px_rgba(125,211,252,0.12)]'
                : 'border-transparent text-slate-600 hover:border-sky-100 hover:bg-sky-50 hover:text-slate-900 dark:text-slate-300 dark:hover:border-white/10 dark:hover:bg-white/10 dark:hover:text-white',
            ]"
          >
            <span
              :class="[
                'h-1.5 w-1.5 shrink-0 rounded-full',
                isActive ? 'bg-sky-400 dark:bg-sky-300' : 'bg-slate-300 dark:bg-slate-600',
              ]"
            ></span>
            <span class="truncate">{{ docsLabel }}</span>
          </a>
        </router-link>
        <router-link
          v-if="resourcesTo && showResources"
          :to="resourcesTo"
          custom
          v-slot="{ href, navigate, isActive }"
        >
          <a
            :href="href"
            @click="navigate"
            :class="[
              'inline-flex h-10 max-w-[8.5rem] items-center gap-1.5 truncate rounded-xl border px-3 py-2 text-xs font-semibold transition sm:max-w-none sm:px-3.5 sm:text-sm',
              isActive
                ? 'border-primary-200 bg-primary-50 text-primary-700 shadow-[0_10px_28px_rgba(59,130,246,0.12)] dark:border-primary-400/40 dark:bg-primary-400/10 dark:text-primary-200'
                : 'border-transparent text-slate-600 hover:border-primary-100 hover:bg-primary-50 hover:text-primary-700 dark:text-slate-300 dark:hover:border-primary-400/20 dark:hover:bg-primary-400/10 dark:hover:text-primary-200',
            ]"
          >
            <span :class="['h-1.5 w-1.5 shrink-0 rounded-full', isActive ? 'bg-primary-500 dark:bg-primary-300' : 'bg-primary-300 dark:bg-primary-500/60']"></span>
            <span class="truncate">{{ resourcesLabel }}</span>
          </a>
        </router-link>
        <router-link
          v-if="partnerTo && showPartner"
          :to="partnerTo"
          custom
          v-slot="{ href, navigate, isActive }"
        >
          <a
            :href="href"
            @click="navigate"
            :class="[
              'inline-flex h-10 max-w-[7.5rem] items-center gap-1.5 truncate rounded-xl border px-3 py-2 text-xs font-semibold transition sm:max-w-none sm:px-3.5 sm:text-sm',
              isActive
                ? 'border-cyan-200 bg-cyan-50 text-cyan-700 shadow-[0_10px_28px_rgba(34,211,238,0.12)] dark:border-cyan-400/40 dark:bg-cyan-400/10 dark:text-cyan-100 dark:shadow-[0_0_28px_rgba(34,211,238,0.14)]'
                : 'border-transparent text-slate-600 hover:border-cyan-100 hover:bg-cyan-50 hover:text-cyan-700 dark:text-slate-300 dark:hover:border-cyan-400/20 dark:hover:bg-cyan-400/10 dark:hover:text-cyan-200',
            ]"
          >
            <span
              :class="[
                'h-1.5 w-1.5 shrink-0 rounded-full',
                isActive ? 'bg-cyan-400 dark:bg-cyan-300' : 'bg-cyan-300 dark:bg-cyan-500/60',
              ]"
            ></span>
            <span class="truncate">{{ partnerLabel }}</span>
          </a>
        </router-link>
        <slot name="tools" />
        <a
          v-if="docUrl && !docsTo && showDocs"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden h-10 items-center rounded-xl px-3 py-2 text-sm font-medium text-slate-600 transition hover:bg-sky-50 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-white/10 dark:hover:text-white md:inline-flex"
        >
          {{ docsLabel }}
        </a>
        <router-link
          :to="ctaTo"
          class="inline-flex h-10 items-center rounded-xl border border-transparent bg-gradient-to-r from-primary-500 via-cyan-400 to-sky-400 px-4 py-2 text-sm font-semibold text-white shadow-[0_16px_40px_rgba(20,184,166,0.16)] transition hover:-translate-y-0.5 hover:from-primary-400 hover:via-cyan-300 hover:to-sky-300 dark:border-white/15 dark:bg-white dark:text-slate-950 dark:shadow-[0_0_32px_rgba(255,255,255,0.16)] dark:hover:bg-slate-100"
        >
          {{ ctaLabel }}
        </router-link>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores'
import { BRAND_LOGO_URL } from '@/constants/brand'
import { normalizeMarketingNavItems } from '@/utils/marketingNav'

const props = defineProps<{
  siteName: string
  subtitle: string
  logo?: string
  docUrl?: string
  docsLabel: string
  docsTo?: string
  ctaTo: string
  ctaLabel: string
  modelMarketplaceTo?: string
  modelMarketplaceLabel?: string
  partnerTo?: string
  partnerLabel?: string
  resourcesTo?: string
  resourcesLabel?: string
  visibleItems?: string[]
}>()

const appStore = useAppStore()
const visibleNavItems = computed(() =>
  normalizeMarketingNavItems(props.visibleItems),
)
const siteLogo = computed(() => props.logo || appStore.siteLogo || BRAND_LOGO_URL)
const showModels = computed(() => visibleNavItems.value.includes('models'))
const showDocs = computed(() => visibleNavItems.value.includes('docs'))
const showPartner = computed(() => visibleNavItems.value.includes('partner'))
const showResources = computed(() => visibleNavItems.value.includes('resources'))
</script>

<style scoped>
.marketing-navbar {
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Noto Sans SC", sans-serif;
}
</style>
