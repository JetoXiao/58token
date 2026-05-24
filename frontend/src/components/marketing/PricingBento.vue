<template>
  <section class="mx-auto max-w-7xl px-4 py-14 sm:px-6">
    <div class="mb-4 overflow-hidden rounded-[1.75rem] border border-cyan-200/70 bg-white/80 p-6 shadow-[0_24px_90px_rgba(8,145,178,0.14)] backdrop-blur-xl dark:border-cyan-300/20 dark:bg-cyan-300/[0.06] dark:shadow-[0_24px_90px_rgba(34,211,238,0.10)] sm:p-7">
      <div class="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
        <div class="min-w-0">
          <p class="text-sm font-medium text-cyan-700 dark:text-cyan-300">
            {{ t('home.marketing.rechargeEyebrow') }}
          </p>
          <h3 class="mt-2 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white sm:text-3xl">
            {{ t('home.marketing.rechargeTitle') }}
          </h3>
          <p class="mt-3 max-w-3xl text-sm leading-6 text-gray-600 dark:text-slate-400 sm:text-base">
            {{ t('home.marketing.rechargeDescription') }}
          </p>
        </div>
        <div class="shrink-0 rounded-2xl border border-gray-200 bg-white/80 px-5 py-4 text-left dark:border-white/10 dark:bg-slate-950/60">
          <p class="text-xs font-medium uppercase text-gray-500 dark:text-slate-500">
            {{ t('home.marketing.rechargeRateLabel') }}
          </p>
          <p class="mt-2 text-3xl font-semibold text-gray-950 dark:text-white">
            {{ t('home.marketing.rechargeRate') }}
          </p>
          <p class="mt-1 text-sm text-cyan-700 dark:text-cyan-300">
            {{ t('home.marketing.rechargeExample') }}
          </p>
        </div>
      </div>
    </div>

    <div class="grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
      <div class="rounded-[1.75rem] border border-gray-200/70 bg-white/70 p-8 backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.04]">
        <p class="text-sm font-medium text-violet-600 dark:text-violet-300">
          {{ t('home.marketing.pricingEyebrow') }}
        </p>
        <h2 class="mt-3 text-3xl font-semibold tracking-tight text-gray-950 dark:text-white sm:text-4xl">
          {{ t('home.marketing.pricingTitle') }}
        </h2>
        <p class="mt-4 text-gray-600 dark:text-slate-400">
          {{ t('home.marketing.pricingDescription') }}
        </p>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <article
          v-for="plan in plans"
          :key="plan.name"
          class="rounded-[1.75rem] border border-gray-200/70 bg-white/75 p-6 shadow-[0_24px_80px_rgba(15,23,42,0.10)] backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/60 dark:shadow-[0_24px_80px_rgba(0,0,0,0.28)]"
        >
          <div class="flex items-center justify-between gap-3">
            <h3 class="min-w-0 truncate text-lg font-semibold text-gray-950 dark:text-white">
              {{ plan.name }}
            </h3>
            <span class="shrink-0 rounded-full border border-gray-200 px-3 py-1 text-xs text-gray-500 dark:border-white/10 dark:text-slate-400">
              {{ plan.badge }}
            </span>
          </div>
          <p class="mt-5 text-4xl font-semibold text-gray-950 dark:text-white">{{ plan.price }}</p>
          <p class="mt-3 text-sm leading-6 text-gray-600 dark:text-slate-400">{{ plan.copy }}</p>
          <div class="mt-6 h-px bg-gradient-to-r from-transparent via-gray-200 to-transparent dark:via-white/15"></div>
          <ul class="mt-6 space-y-3 text-sm text-gray-700 dark:text-slate-300">
            <li v-for="feature in plan.features" :key="feature" class="flex gap-2">
              <span class="text-cyan-600 dark:text-cyan-300">✓</span>
              <span>{{ feature }}</span>
            </li>
          </ul>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t, tm, rt } = useI18n()

const plans = computed(() =>
  ['developer', 'team'].map((key) => {
    const rawFeatures = tm(`home.marketing.plans.${key}.features`) as unknown
    const features = Array.isArray(rawFeatures)
      ? rawFeatures.map((feature: unknown) => rt(feature as any))
      : []
    return {
      name: t(`home.marketing.plans.${key}.name`),
      badge: t(`home.marketing.plans.${key}.badge`),
      price: t(`home.marketing.plans.${key}.price`),
      copy: t(`home.marketing.plans.${key}.copy`),
      features
    }
  })
)
</script>
