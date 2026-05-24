<template>
  <section class="relative mx-auto grid max-w-7xl gap-12 px-4 pb-16 pt-12 sm:px-6 lg:grid-cols-[1.05fr_0.95fr] lg:pb-24 lg:pt-20">
    <div class="relative z-10 flex flex-col justify-center">
      <div class="mb-6 inline-flex w-fit items-center gap-2 rounded-full border border-cyan-500/20 bg-cyan-50/80 px-3 py-1 text-xs font-medium text-cyan-700 shadow-[0_0_30px_rgba(34,211,238,0.12)] backdrop-blur-xl dark:border-cyan-300/20 dark:bg-cyan-300/10 dark:text-cyan-100">
        <span class="h-1.5 w-1.5 rounded-full bg-emerald-300 shadow-[0_0_16px_rgba(110,231,183,0.9)]"></span>
        {{ t('home.marketing.eyebrow') }}
      </div>

      <h1 class="max-w-4xl text-5xl font-semibold tracking-tight text-gray-950 dark:text-white sm:text-6xl lg:text-7xl">
        {{ siteName }}
        <span class="block bg-gradient-to-r from-cyan-600 via-gray-950 to-violet-600 bg-clip-text text-transparent dark:from-cyan-200 dark:via-white dark:to-violet-200">
          {{ subtitle }}
        </span>
      </h1>

      <p class="mt-6 max-w-2xl text-base leading-8 text-gray-600 dark:text-slate-300 sm:text-lg">
        {{ t('home.marketing.heroCopy') }}
      </p>

      <div class="mt-9 flex flex-col gap-3 sm:flex-row">
        <router-link
          :to="primaryTo"
          class="inline-flex items-center justify-center rounded-2xl bg-gray-950 px-6 py-3 text-sm font-semibold text-white shadow-[0_18px_48px_rgba(15,23,42,0.2)] transition hover:-translate-y-0.5 hover:bg-gray-800 dark:bg-white dark:text-slate-950 dark:shadow-[0_0_48px_rgba(255,255,255,0.18)] dark:hover:bg-slate-100"
        >
          {{ primaryLabel }}
        </router-link>
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center justify-center rounded-2xl border border-gray-200 bg-white/70 px-6 py-3 text-sm font-semibold text-gray-900 backdrop-blur-xl transition hover:-translate-y-0.5 hover:bg-white dark:border-white/10 dark:bg-white/[0.04] dark:text-white dark:hover:bg-white/10"
        >
          {{ t('home.marketing.docsCta') }}
        </a>
      </div>
    </div>

    <div class="relative">
      <div class="absolute -inset-8 rounded-[2rem] bg-gradient-to-tr from-cyan-500/20 via-fuchsia-500/10 to-white/10 blur-3xl"></div>
      <div class="relative overflow-hidden rounded-[1.75rem] border border-gray-200/70 bg-white/75 p-4 shadow-[0_30px_120px_rgba(15,23,42,0.16)] backdrop-blur-2xl dark:border-white/10 dark:bg-slate-950/70 dark:shadow-[0_30px_120px_rgba(0,0,0,0.46)]">
        <div class="mb-4 flex items-center justify-between">
          <div class="flex gap-2">
            <span class="h-3 w-3 rounded-full bg-red-400/80"></span>
            <span class="h-3 w-3 rounded-full bg-amber-300/80"></span>
            <span class="h-3 w-3 rounded-full bg-emerald-300/80"></span>
          </div>
          <span class="rounded-full border border-gray-200 px-3 py-1 text-xs text-gray-500 dark:border-white/10 dark:text-slate-400">{{ t('home.marketing.liveRouting') }}</span>
        </div>

        <div class="grid gap-3 md:grid-cols-2">
          <div class="rounded-2xl border border-gray-200/70 bg-white/70 p-4 dark:border-white/10 dark:bg-white/[0.04]">
            <p class="text-xs uppercase tracking-[0.2em] text-gray-500 dark:text-slate-500">{{ t('home.marketing.requests') }}</p>
            <p class="mt-3 text-4xl font-semibold text-gray-950 dark:text-white">2.4M</p>
            <p class="mt-2 text-sm text-emerald-600 dark:text-emerald-300">{{ t('home.marketing.weeklyGrowth') }}</p>
          </div>
          <div class="rounded-2xl border border-gray-200/70 bg-white/70 p-4 dark:border-white/10 dark:bg-white/[0.04]">
            <p class="text-xs uppercase tracking-[0.2em] text-gray-500 dark:text-slate-500">{{ t('home.marketing.latency') }}</p>
            <p class="mt-3 text-4xl font-semibold text-gray-950 dark:text-white">184ms</p>
            <p class="mt-2 text-sm text-cyan-600 dark:text-cyan-300">{{ t('home.marketing.p95Stable') }}</p>
          </div>
        </div>

        <div class="mt-3 rounded-2xl border border-gray-200/70 bg-gray-950 p-4 font-mono text-xs text-slate-300 dark:border-white/10 dark:bg-black/30">
          <p><span class="text-cyan-300">$</span> curl /v1/messages --model claude</p>
          <p class="mt-2 text-slate-500">{{ t('home.marketing.routingRequest') }}</p>
          <p class="mt-2 text-emerald-300">200 OK <span class="text-slate-400">provider=openai cost=$0.0041</span></p>
        </div>

        <div class="mt-3 grid gap-3 sm:grid-cols-3">
          <div v-for="item in ['Claude', 'OpenAI', 'Gemini']" :key="item" class="rounded-2xl border border-gray-200/70 bg-white/70 p-3 text-sm text-gray-700 dark:border-white/10 dark:bg-white/[0.035] dark:text-slate-300">
            <span class="mr-2 inline-block h-2 w-2 rounded-full bg-emerald-300"></span>{{ item }}
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  siteName: string
  subtitle: string
  primaryTo: string
  primaryLabel: string
  docUrl?: string
}>()

const { t } = useI18n()
</script>
