<template>
  <div :class="['downloads-page', { 'downloads-page-embedded': embedded }]">
    <MarketingNavbar
      v-if="!embedded"
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
          class="theme-button"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @click="toggleTheme"
        >
          <Icon v-if="isDark" name="sun" size="md" />
          <Icon v-else name="moon" size="md" />
        </button>
      </template>
    </MarketingNavbar>

    <main class="downloads-main">
      <section class="downloads-hero">
        <div>
          <span class="eyebrow"><Icon name="download" size="sm" />{{ t('downloads.eyebrow') }}</span>
          <h1>{{ t('downloads.title') }}</h1>
          <p>{{ t('downloads.subtitle') }}</p>
        </div>
        <div class="hero-status">
          <Icon name="shield" size="lg" />
          <div>
            <strong>{{ t('downloads.protectionTitle') }}</strong>
            <span>{{ t('downloads.protectionDescription') }}</span>
          </div>
        </div>
      </section>

      <section class="catalog-section" aria-live="polite">
        <div class="catalog-heading">
          <div>
            <h2>{{ t('downloads.catalogTitle') }}</h2>
            <p>{{ t('downloads.catalogDescription') }}</p>
          </div>
          <button v-if="loadError" class="reload-button" type="button" @click="loadResources">
            <Icon name="refresh" size="sm" />
            {{ t('downloads.retry') }}
          </button>
        </div>

        <div v-if="loading" class="resource-list resource-list-loading">
          <div v-for="index in 3" :key="index" class="resource-skeleton"></div>
        </div>
        <div v-else-if="loadError" class="empty-state">
          <Icon name="exclamationCircle" size="lg" />
          <p>{{ t('downloads.loadFailed') }}</p>
        </div>
        <div v-else-if="!resources.length" class="empty-state">
          <Icon name="inbox" size="lg" />
          <p>{{ t('downloads.noResources') }}</p>
        </div>
        <div v-else class="resource-list">
          <article v-for="resource in resources" :key="resource.id" class="resource-row">
            <div class="resource-icon"><Icon name="cube" size="lg" /></div>
            <div class="resource-copy">
              <div class="resource-title-row">
                <h3>{{ resourceName(resource) }}</h3>
                <span v-if="resource.version" class="version-chip">{{ resource.version }}</span>
              </div>
              <p v-if="resourceDescription(resource)">{{ resourceDescription(resource) }}</p>
              <dl class="resource-meta">
                <div v-if="resource.platform"><dt>{{ t('downloads.platform') }}</dt><dd>{{ resource.platform }}</dd></div>
                <div><dt>{{ t('downloads.size') }}</dt><dd>{{ formatBytes(resource.size_bytes) }}</dd></div>
                <div><dt>{{ t('downloads.updated') }}</dt><dd>{{ formatDate(resource.uploaded_at) }}</dd></div>
                <div><dt>{{ t('downloads.downloads') }}</dt><dd>{{ numberFormatter.format(resource.download_count) }}</dd></div>
              </dl>
              <div v-if="resource.checksum_sha256" class="checksum" :title="resource.checksum_sha256">
                <span>{{ t('downloads.checksum') }}</span>
                <code>{{ resource.checksum_sha256 }}</code>
              </div>
            </div>
            <button
              class="download-button"
              type="button"
              :disabled="downloadingID === resource.id"
              @click="startDownload(resource)"
            >
              <Icon :name="downloadingID === resource.id ? 'refresh' : 'download'" size="sm" :class="{ spinning: downloadingID === resource.id }" />
              {{ downloadingID === resource.id ? t('downloads.downloading') : t('downloads.download') }}
            </button>
          </article>
        </div>
        <p v-if="downloadError" class="download-error" role="alert">{{ downloadError }}</p>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import downloadResourcesAPI, { type DownloadResource } from '@/api/downloadResources'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import MarketingNavbar from '@/components/marketing/MarketingNavbar.vue'
import { BRAND_LOGO_URL } from '@/constants/brand'
import { useAppStore, useAuthStore } from '@/stores'
import { formatBytes } from '@/utils/format'

const { t, locale } = useI18n()
const { embedded = false } = defineProps<{ embedded?: boolean }>()
const appStore = useAppStore()
const authStore = useAuthStore()
const resources = ref<DownloadResource[]>([])
const loading = ref(true)
const loadError = ref(false)
const downloadError = ref('')
const downloadingID = ref<number | null>(null)
const isDark = ref(false)

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'UseAiForMe')
const siteLogo = computed(() => appStore.siteLogo || BRAND_LOGO_URL)
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const marketingNavItems = computed(() => appStore.cachedPublicSettings?.marketing_nav_items)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const numberFormatter = computed(() => new Intl.NumberFormat(locale.value.startsWith('zh') ? 'zh-CN' : 'en-US'))

function resourceName(resource: DownloadResource): string {
  return locale.value.startsWith('zh')
    ? resource.name_zh || resource.name_en || resource.file_name
    : resource.name_en || resource.name_zh || resource.file_name
}

function resourceDescription(resource: DownloadResource): string {
  return locale.value.startsWith('zh')
    ? resource.description_zh || resource.description_en
    : resource.description_en || resource.description_zh
}

function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return new Intl.DateTimeFormat(locale.value.startsWith('zh') ? 'zh-CN' : 'en-US', {
    year: 'numeric', month: 'short', day: 'numeric',
  }).format(date)
}

async function loadResources() {
  loading.value = true
  loadError.value = false
  try {
    resources.value = await downloadResourcesAPI.list()
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
}

async function startDownload(resource: DownloadResource) {
  if (downloadingID.value !== null) return
  downloadingID.value = resource.id
  downloadError.value = ''
  try {
    const authorization = await downloadResourcesAPI.authorizeDownload(resource.id)
    const link = document.createElement('a')
    link.href = authorization.url
    link.rel = 'noopener'
    link.download = resource.file_name
    document.body.appendChild(link)
    link.click()
    link.remove()
    resource.download_count += 1
  } catch {
    downloadError.value = t('downloads.downloadFailed')
  } finally {
    downloadingID.value = null
  }
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const saved = localStorage.getItem('theme')
  isDark.value = saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', isDark.value)
}

onMounted(async () => {
  if (!embedded) initTheme()
  await Promise.allSettled([appStore.fetchPublicSettings(), loadResources()])
})
</script>

<style scoped>
.downloads-page { min-height: 100vh; background: #f7f8fb; color: #111827; }
.downloads-page-embedded { min-height: 0; background: transparent; }
.downloads-page-embedded .downloads-main { width: 100%; padding-top: 0; }
.downloads-page :deep(.marketing-navbar) { background: linear-gradient(180deg, rgba(236, 253, 245, .72), transparent); }
.downloads-main { width: min(1180px, calc(100% - 32px)); margin: 0 auto; padding: 28px 0 80px; }
.downloads-hero { display: flex; align-items: end; justify-content: space-between; gap: 32px; padding: 42px 0 38px; border-bottom: 1px solid #dbe3ed; }
.eyebrow { display: inline-flex; align-items: center; gap: 8px; color: #047857; font-size: 12px; font-weight: 700; letter-spacing: .08em; }
.downloads-hero h1 { margin: 14px 0 0; font-size: clamp(38px, 5vw, 62px); line-height: 1.05; font-weight: 700; }
.downloads-hero p { max-width: 630px; margin: 16px 0 0; color: #526174; font-size: 17px; line-height: 1.7; }
.hero-status { display: flex; max-width: 330px; gap: 14px; padding: 16px 0; color: #065f46; }
.hero-status > :first-child { flex: none; margin-top: 2px; }
.hero-status strong, .hero-status span { display: block; }
.hero-status span { margin-top: 4px; color: #526174; font-size: 13px; line-height: 1.55; }
.catalog-section { padding-top: 34px; }
.catalog-heading { display: flex; align-items: end; justify-content: space-between; gap: 16px; }
.catalog-heading h2 { margin: 0; font-size: 25px; }
.catalog-heading p { margin: 7px 0 0; color: #64748b; font-size: 14px; }
.resource-list { margin-top: 22px; border-top: 1px solid #dbe3ed; }
.resource-row { display: grid; grid-template-columns: 48px minmax(0, 1fr) auto; gap: 18px; align-items: center; padding: 24px 4px; border-bottom: 1px solid #dbe3ed; }
.resource-icon { display: grid; width: 48px; height: 48px; place-items: center; border-radius: 8px; background: #e9fbf4; color: #047857; }
.resource-title-row { display: flex; align-items: center; flex-wrap: wrap; gap: 9px; }
.resource-title-row h3 { margin: 0; font-size: 18px; }
.version-chip { border-radius: 999px; background: #eef2f7; padding: 3px 8px; color: #475569; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 11px; }
.resource-copy > p { margin: 6px 0 0; color: #64748b; font-size: 14px; line-height: 1.55; }
.resource-meta { display: flex; flex-wrap: wrap; gap: 8px 20px; margin: 14px 0 0; }
.resource-meta div { display: flex; gap: 6px; color: #64748b; font-size: 12px; }
.resource-meta dt { color: #94a3b8; }.resource-meta dd { margin: 0; color: #475569; font-weight: 600; }
.checksum { display: flex; min-width: 0; gap: 7px; margin-top: 12px; color: #94a3b8; font-size: 11px; }.checksum code { overflow: hidden; color: #64748b; text-overflow: ellipsis; white-space: nowrap; }
.download-button, .reload-button, .theme-button { display: inline-flex; align-items: center; justify-content: center; gap: 8px; border: 0; cursor: pointer; font: inherit; }
.download-button { min-width: 122px; border-radius: 8px; background: #0f766e; padding: 11px 15px; color: white; font-size: 14px; font-weight: 700; transition: background .2s, transform .2s; }.download-button:hover:not(:disabled) { background: #047857; transform: translateY(-1px); }.download-button:disabled { cursor: wait; opacity: .65; }
.reload-button { border: 1px solid #dbe3ed; border-radius: 8px; background: white; padding: 9px 12px; color: #475569; font-size: 13px; font-weight: 600; }.theme-button { border: 1px solid rgba(203,213,225,.7); border-radius: 10px; background: rgba(255,255,255,.7); padding: 8px; color: #475569; }
.empty-state { display: grid; min-height: 240px; place-content: center; gap: 12px; color: #64748b; text-align: center; }.empty-state p { margin: 0; }
.resource-list-loading { display: grid; gap: 12px; padding-top: 22px; border-top: 0; }.resource-skeleton { height: 108px; border-radius: 8px; background: linear-gradient(100deg, #edf2f7 35%, #f8fafc 50%, #edf2f7 65%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.download-error { margin: 16px 0 0; color: #b91c1c; font-size: 14px; }.spinning { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } } @keyframes shimmer { to { background-position: -200% 0; } }
:global(.dark) .downloads-page { background: #080b11; color: #f8fafc; }:global(.dark) .downloads-page :deep(.marketing-navbar) { background: linear-gradient(180deg, rgba(6, 78, 59, .18), transparent); }:global(.dark) .downloads-hero, :global(.dark) .resource-list, :global(.dark) .resource-row { border-color: #243142; }:global(.dark) .downloads-hero p, :global(.dark) .catalog-heading p, :global(.dark) .resource-copy > p, :global(.dark) .hero-status span { color: #a5b4c6; }:global(.dark) .resource-icon { background: rgba(6, 78, 59, .36); color: #6ee7b7; }:global(.dark) .version-chip { background: #1e293b; color: #cbd5e1; }:global(.dark) .resource-meta dt { color: #64748b; }:global(.dark) .resource-meta dd, :global(.dark) .checksum code { color: #cbd5e1; }:global(.dark) .reload-button, :global(.dark) .theme-button { border-color: #334155; background: #111827; color: #dbeafe; }:global(.dark) .resource-skeleton { background: linear-gradient(100deg, #111827 35%, #1e293b 50%, #111827 65%); background-size: 200% 100%; }
@media (max-width: 720px) { .downloads-main { width: min(100% - 28px, 1180px); padding-top: 12px; }.downloads-hero { display: block; padding: 30px 0; }.hero-status { max-width: none; margin-top: 24px; border-top: 1px solid #dbe3ed; }.resource-row { grid-template-columns: 42px minmax(0, 1fr); gap: 14px; }.resource-icon { width: 42px; height: 42px; }.download-button { grid-column: 2; width: 100%; }.catalog-heading { align-items: start; }.resource-meta { gap: 7px 14px; }.checksum { max-width: 100%; }:global(.dark) .hero-status { border-color: #243142; } }
:global(.dark) .downloads-page-embedded { background: transparent; }
</style>
