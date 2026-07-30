<template>
  <header class="marketing-navbar relative z-20 px-4 py-4 sm:px-6">
    <nav
      class="mx-auto flex max-w-7xl items-center justify-between rounded-2xl border border-gray-200/70 bg-white/70 px-4 py-3 shadow-[0_18px_70px_rgba(15,23,42,0.08)] backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.04] dark:shadow-[0_18px_70px_rgba(0,0,0,0.28)]"
    >
      <router-link to="/home" class="flex min-w-0 items-center gap-3">
        <span class="flex h-9 w-9 items-center justify-center overflow-hidden rounded-xl border border-gray-200/80 bg-white/80 dark:border-white/10 dark:bg-white/10">
          <img :src="logo || BRAND_LOGO_URL" alt="" class="h-full w-full object-contain" />
        </span>
        <span class="hidden min-w-0 sm:block">
          <span class="block truncate text-sm font-semibold tracking-tight text-gray-950 dark:text-white">{{ siteName }}</span>
          <span class="block truncate text-xs text-gray-500 dark:text-slate-400">{{ subtitle }}</span>
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
                ? 'border-primary-300 bg-primary-50 text-primary-700 shadow-[0_10px_28px_rgba(20,184,166,0.16)] dark:border-primary-400/40 dark:bg-primary-400/10 dark:text-primary-200 dark:shadow-[0_0_28px_rgba(45,212,191,0.12)]'
                : 'border-transparent text-gray-600 hover:border-gray-200 hover:bg-gray-100 hover:text-gray-950 dark:text-slate-300 dark:hover:border-white/10 dark:hover:bg-white/10 dark:hover:text-white',
            ]"
          >
            <span
              :class="[
                'h-1.5 w-1.5 shrink-0 rounded-full',
                isActive ? 'bg-primary-500 dark:bg-primary-300' : 'bg-gray-300 dark:bg-slate-600',
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
                ? 'border-primary-300 bg-primary-50 text-primary-700 shadow-[0_10px_28px_rgba(20,184,166,0.16)] dark:border-primary-400/40 dark:bg-primary-400/10 dark:text-primary-200 dark:shadow-[0_0_28px_rgba(45,212,191,0.12)]'
                : 'border-transparent text-gray-600 hover:border-gray-200 hover:bg-gray-100 hover:text-gray-950 dark:text-slate-300 dark:hover:border-white/10 dark:hover:bg-white/10 dark:hover:text-white',
            ]"
          >
            <span
              :class="[
                'h-1.5 w-1.5 shrink-0 rounded-full',
                isActive ? 'bg-primary-500 dark:bg-primary-300' : 'bg-gray-300 dark:bg-slate-600',
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
                ? 'border-emerald-300 bg-emerald-50 text-emerald-700 shadow-[0_10px_28px_rgba(16,185,129,0.16)] dark:border-emerald-400/40 dark:bg-emerald-400/10 dark:text-emerald-200'
                : 'border-transparent text-gray-600 hover:border-emerald-200 hover:bg-emerald-50 hover:text-emerald-700 dark:text-slate-300 dark:hover:border-emerald-400/20 dark:hover:bg-emerald-400/10 dark:hover:text-emerald-200',
            ]"
          >
            <span :class="['h-1.5 w-1.5 shrink-0 rounded-full', isActive ? 'bg-emerald-500 dark:bg-emerald-300' : 'bg-emerald-300 dark:bg-emerald-500/60']"></span>
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
                ? 'border-violet-300 bg-violet-50 text-violet-700 shadow-[0_10px_28px_rgba(124,58,237,0.16)] dark:border-violet-400/40 dark:bg-violet-400/10 dark:text-violet-200 dark:shadow-[0_0_28px_rgba(124,58,237,0.14)]'
                : 'border-transparent text-gray-600 hover:border-violet-200 hover:bg-violet-50 hover:text-violet-700 dark:text-slate-300 dark:hover:border-violet-400/20 dark:hover:bg-violet-400/10 dark:hover:text-violet-200',
            ]"
          >
            <span
              :class="[
                'h-1.5 w-1.5 shrink-0 rounded-full',
                isActive ? 'bg-violet-500 dark:bg-violet-300' : 'bg-violet-300 dark:bg-violet-500/60',
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
          class="hidden h-10 items-center rounded-xl px-3 py-2 text-sm font-medium text-gray-600 transition hover:bg-gray-100 hover:text-gray-950 dark:text-slate-300 dark:hover:bg-white/10 dark:hover:text-white md:inline-flex"
        >
          {{ docsLabel }}
        </a>
        <router-link
          :to="ctaTo"
          class="inline-flex h-10 items-center rounded-xl border border-gray-900/10 bg-gray-950 px-4 py-2 text-sm font-semibold text-white shadow-[0_16px_40px_rgba(15,23,42,0.16)] transition hover:-translate-y-0.5 hover:bg-gray-800 dark:border-white/15 dark:bg-white dark:text-slate-950 dark:shadow-[0_0_32px_rgba(255,255,255,0.16)] dark:hover:bg-slate-100"
        >
          {{ ctaLabel }}
        </router-link>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
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

const visibleNavItems = computed(() =>
  normalizeMarketingNavItems(props.visibleItems),
)
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
