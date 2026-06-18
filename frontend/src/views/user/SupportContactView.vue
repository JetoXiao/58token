<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <section class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <p class="text-sm font-medium text-primary-600 dark:text-primary-400">
              {{ t('supportContact.eyebrow') }}
            </p>
            <h1 class="mt-2 text-2xl font-semibold text-gray-950 dark:text-white md:text-3xl">
              {{ config.title || t('supportContact.title') }}
            </h1>
            <p class="mt-3 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ config.description || t('supportContact.description') }}
            </p>
          </div>
          <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-300">
            <Icon name="chatBubble" size="lg" />
          </div>
        </div>
      </section>

      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <section
        v-else-if="!config.enabled"
        class="rounded-2xl border border-gray-200 bg-white p-12 text-center shadow-sm dark:border-dark-700 dark:bg-dark-800"
      >
        <Icon name="inbox" size="xl" class="mx-auto mb-4 text-gray-400" />
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('supportContact.disabledTitle') }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
          {{ t('supportContact.disabledDescription') }}
        </p>
      </section>

      <section
        v-else-if="visibleContacts.length === 0"
        class="rounded-2xl border border-gray-200 bg-white p-12 text-center shadow-sm dark:border-dark-700 dark:bg-dark-800"
      >
        <Icon name="inbox" size="xl" class="mx-auto mb-4 text-gray-400" />
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('supportContact.emptyTitle') }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
          {{ t('supportContact.emptyDescription') }}
        </p>
      </section>

      <section v-else class="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="(contact, index) in visibleContacts"
          :key="`${contact.name}-${index}`"
          class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <h2 class="truncate text-lg font-semibold text-gray-950 dark:text-white">
                {{ contact.name || t('supportContact.defaultName', { index: index + 1 }) }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ contact.work_hours || t('supportContact.workHoursUnknown') }}
              </p>
            </div>
            <span class="rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
              {{ t('supportContact.wechat') }}
            </span>
          </div>

          <div class="mt-5 overflow-hidden rounded-xl border border-gray-200 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-900">
            <div class="aspect-square w-full overflow-hidden rounded-lg bg-white dark:bg-dark-800">
              <img
                :src="contact.qr_image"
                :alt="contact.name || t('supportContact.qrAlt')"
                class="h-full w-full object-contain"
                loading="lazy"
              />
            </div>
          </div>

          <div class="mt-4 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
            <Icon name="clock" size="sm" />
            <span>{{ contact.work_hours || t('supportContact.workHoursUnknown') }}</span>
          </div>
        </article>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import type { SupportContactConfig, SupportContactItem } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)

const defaultConfig: SupportContactConfig = {
  enabled: true,
  title: '',
  description: '',
  contacts: [],
}

const config = computed<SupportContactConfig>(() => ({
  ...defaultConfig,
  ...(appStore.cachedPublicSettings?.support_contact_config ?? {}),
  contacts: Array.isArray(appStore.cachedPublicSettings?.support_contact_config?.contacts)
    ? appStore.cachedPublicSettings.support_contact_config.contacts
    : [],
}))

const visibleContacts = computed<SupportContactItem[]>(() =>
  config.value.contacts
    .filter((contact) => Boolean(contact.qr_image?.trim()))
    .slice(0, 3),
)

onMounted(async () => {
  if (appStore.publicSettingsLoaded) return
  loading.value = true
  try {
    await appStore.fetchPublicSettings()
  } finally {
    loading.value = false
  }
})
</script>
