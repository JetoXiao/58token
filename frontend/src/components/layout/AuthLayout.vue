<template>
  <div class="auth-shell relative flex min-h-screen items-center justify-center overflow-hidden p-4">
    <!-- Background -->
    <div
      class="absolute inset-0 bg-[radial-gradient(circle_at_14%_10%,rgba(96,165,250,0.16),transparent_28%),radial-gradient(circle_at_86%_12%,rgba(196,181,253,0.14),transparent_24%),radial-gradient(circle_at_78%_82%,rgba(251,146,60,0.08),transparent_20%),radial-gradient(circle_at_18%_82%,rgba(45,212,191,0.10),transparent_22%),linear-gradient(180deg,#fbfdff_0%,#f4f8fd_48%,#eef4fa_100%)] dark:bg-[radial-gradient(circle_at_14%_10%,rgba(56,189,248,0.16),transparent_28%),radial-gradient(circle_at_86%_12%,rgba(196,181,253,0.10),transparent_24%),radial-gradient(circle_at_18%_82%,rgba(45,212,191,0.08),transparent_22%),linear-gradient(180deg,#05060a_0%,#090c14_52%,#05060a_100%)]"
    ></div>

    <!-- Decorative Elements -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Gradient Orbs -->
      <div
        class="absolute -right-40 -top-40 h-80 w-80 rounded-full bg-primary-400/20 blur-3xl"
      ></div>
      <div
        class="absolute -bottom-40 -left-40 h-80 w-80 rounded-full bg-primary-500/15 blur-3xl"
      ></div>
      <div
        class="absolute left-1/2 top-1/2 h-96 w-96 -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary-300/10 blur-3xl"
      ></div>

      <!-- Grid Pattern -->
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(59,130,246,0.035)_1px,transparent_1px),linear-gradient(90deg,rgba(59,130,246,0.035)_1px,transparent_1px)] bg-[size:64px_64px] opacity-70"
      ></div>
    </div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <RouterLink
        to="/home"
        class="mb-8 block rounded-2xl text-center transition duration-200 hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-primary-500/40 focus:ring-offset-2 focus:ring-offset-transparent"
        :aria-label="siteName"
      >
        <template v-if="settingsLoaded">
          <h1 class="text-gradient mb-2 text-3xl font-bold">
            {{ siteName }}
          </h1>
          <p class="text-sm text-gray-500 dark:text-dark-400">
            {{ siteSubtitle }}
          </p>
        </template>
      </RouterLink>

      <!-- Card Container -->
      <div class="card-glass rounded-2xl p-8 shadow-glass">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs text-gray-400 dark:text-dark-500">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || '58Token')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || '让AI为我所用')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.text-gradient {
  @apply bg-gradient-to-r from-primary-600 to-primary-500 bg-clip-text text-transparent;
}
</style>
