<template>
  <header class="relative z-20 px-4 py-4 sm:px-6">
    <nav
      class="mx-auto flex max-w-7xl items-center justify-between rounded-2xl border border-gray-200/70 bg-white/70 px-4 py-3 shadow-[0_18px_70px_rgba(15,23,42,0.08)] backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.04] dark:shadow-[0_18px_70px_rgba(0,0,0,0.28)]"
    >
      <router-link to="/home" class="flex min-w-0 items-center gap-3">
        <span class="flex h-9 w-9 items-center justify-center overflow-hidden rounded-xl border border-gray-200/80 bg-white/80 dark:border-white/10 dark:bg-white/10">
          <img :src="logo || '/logo.png'" alt="" class="h-full w-full object-contain" />
        </span>
        <span class="hidden min-w-0 sm:block">
          <span class="block truncate text-sm font-semibold tracking-tight text-gray-950 dark:text-white">{{ siteName }}</span>
          <span class="block truncate text-xs text-gray-500 dark:text-slate-400">{{ subtitle }}</span>
        </span>
      </router-link>

      <div class="flex items-center gap-2">
        <slot name="tools" />
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hidden rounded-xl px-3 py-2 text-sm font-medium text-gray-600 transition hover:bg-gray-100 hover:text-gray-950 dark:text-slate-300 dark:hover:bg-white/10 dark:hover:text-white md:inline-flex"
        >
          {{ docsLabel }}
        </a>
        <router-link
          :to="ctaTo"
          class="inline-flex items-center rounded-xl border border-gray-900/10 bg-gray-950 px-4 py-2 text-sm font-semibold text-white shadow-[0_16px_40px_rgba(15,23,42,0.16)] transition hover:-translate-y-0.5 hover:bg-gray-800 dark:border-white/15 dark:bg-white dark:text-slate-950 dark:shadow-[0_0_32px_rgba(255,255,255,0.16)] dark:hover:bg-slate-100"
        >
          {{ ctaLabel }}
        </router-link>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
defineProps<{
  siteName: string
  subtitle: string
  logo?: string
  docUrl?: string
  docsLabel: string
  ctaTo: string
  ctaLabel: string
}>()
</script>
