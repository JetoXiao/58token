<template>
  <section class="mx-auto max-w-7xl px-4 py-10 sm:px-6">
    <div class="mb-8 max-w-2xl">
      <p class="text-sm font-medium text-cyan-600 dark:text-cyan-300">
        {{ t('home.marketing.platformEyebrow') }}
      </p>
      <h2 class="mt-3 text-3xl font-semibold tracking-tight text-gray-950 dark:text-white sm:text-4xl">
        {{ t('home.marketing.platformTitle') }}
      </h2>
      <p class="mt-4 text-gray-600 dark:text-slate-400">
        {{ t('home.marketing.platformDescription') }}
      </p>
    </div>

    <div class="grid gap-4 lg:grid-cols-6">
      <article
        v-for="item in items"
        :key="item.title"
        :class="[
          'group rounded-[1.5rem] border border-gray-200/70 bg-white/70 p-6 backdrop-blur-xl transition duration-300 hover:-translate-y-1 hover:border-cyan-400/40 hover:bg-white dark:border-white/10 dark:bg-white/[0.04] dark:hover:border-cyan-300/30 dark:hover:bg-white/[0.07]',
          item.wide ? 'lg:col-span-3' : 'lg:col-span-2'
        ]"
      >
        <div class="mb-8 flex h-11 w-11 items-center justify-center rounded-2xl bg-gradient-to-br from-cyan-100 to-violet-100 text-cyan-700 ring-1 ring-gray-900/5 dark:from-cyan-300/20 dark:to-violet-400/20 dark:text-cyan-200 dark:ring-white/10">
          {{ item.icon }}
        </div>
        <h3 class="text-lg font-semibold text-gray-950 dark:text-white">{{ item.title }}</h3>
        <p class="mt-3 text-sm leading-6 text-gray-600 dark:text-slate-400">
          {{ item.description }}
        </p>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const items = computed(() => [
  {
    title: t('home.marketing.bento.unifiedGateway.title'),
    description: t('home.marketing.bento.unifiedGateway.description'),
    icon: '01',
    wide: true
  },
  {
    title: t('home.marketing.bento.accountPool.title'),
    description: t('home.marketing.bento.accountPool.description'),
    icon: '02',
    wide: true
  },
  {
    title: t('home.marketing.bento.realtimeBilling.title'),
    description: t('home.marketing.bento.realtimeBilling.description'),
    icon: '03'
  },
  {
    title: t('home.marketing.bento.accessControl.title'),
    description: t('home.marketing.bento.accessControl.description'),
    icon: '04'
  },
  {
    title: t('home.marketing.bento.opsDashboard.title'),
    description: t('home.marketing.bento.opsDashboard.description'),
    icon: '05'
  }
])
</script>
