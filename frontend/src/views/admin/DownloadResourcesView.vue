<template>
  <AppLayout>
    <div class="resource-admin">
      <header class="page-header">
        <div>
          <p class="eyebrow"><Icon name="inbox" size="sm" />{{ t('admin.downloadResources.eyebrow') }}</p>
          <h1>{{ t('admin.downloadResources.title') }}</h1>
          <p>{{ t('admin.downloadResources.description') }}</p>
        </div>
        <div class="page-actions">
          <button class="button secondary" type="button" :disabled="loading" @click="refresh">
            <Icon name="refresh" size="sm" :class="{ spinning: loading }" />{{ t('common.refresh') }}
          </button>
          <button v-if="canManage" class="button primary" type="button" @click="openCreate">
            <Icon name="plus" size="sm" />{{ t('admin.downloadResources.addResource') }}
          </button>
        </div>
      </header>

      <section class="metrics" aria-live="polite">
        <article><span>{{ t('admin.downloadResources.totalResources') }}</span><strong>{{ resources.length }}</strong></article>
        <article><span>{{ t('admin.downloadResources.publishedResources') }}</span><strong>{{ publishedCount }}</strong></article>
        <article><span>{{ t('admin.downloadResources.totalDownloads') }}</span><strong>{{ numberFormatter.format(totalDownloads) }}</strong></article>
        <article><span>{{ t('admin.downloadResources.storageStatus') }}</span><strong :class="storageConfigured ? 'success' : 'muted'">{{ storageConfigured ? t('admin.downloadResources.storageReady') : t('admin.downloadResources.storageMissing') }}</strong></article>
      </section>

      <section v-if="canManage" class="panel storage-panel">
        <div class="panel-heading">
          <div><h2>{{ t('admin.downloadResources.storageTitle') }}</h2><p>{{ t('admin.downloadResources.storageDescription') }}</p></div>
          <span :class="['status-chip', storageConfigured ? 'ready' : 'missing']">{{ storageConfigured ? t('admin.downloadResources.storageReady') : t('admin.downloadResources.storageMissing') }}</span>
        </div>
        <div class="storage-grid">
          <label><span>{{ t('admin.downloadResources.endpoint') }}</span><input v-model.trim="storage.endpoint" :disabled="!canManage" placeholder="https://ACCOUNT_ID.r2.cloudflarestorage.com" /></label>
          <label><span>{{ t('admin.downloadResources.bucket') }}</span><input v-model.trim="storage.bucket" :disabled="!canManage" placeholder="downloads" /></label>
          <label><span>{{ t('admin.downloadResources.accessKey') }}</span><input v-model.trim="storage.access_key_id" :disabled="!canManage" autocomplete="off" /></label>
          <label><span>{{ t('admin.downloadResources.secretKey') }}</span><input v-model.trim="storage.secret_access_key" :disabled="!canManage" autocomplete="new-password" type="password" :placeholder="storageConfigured ? t('admin.downloadResources.secretUnchanged') : ''" /></label>
          <label><span>{{ t('admin.downloadResources.prefix') }}</span><input v-model.trim="storage.prefix" :disabled="!canManage" placeholder="downloads/" /></label>
          <label><span>{{ t('admin.downloadResources.region') }}</span><input v-model.trim="storage.region" :disabled="!canManage" placeholder="auto" /></label>
        </div>
        <label class="checkbox-row"><input v-model="storage.force_path_style" type="checkbox" :disabled="!canManage" /><span>{{ t('admin.downloadResources.forcePathStyle') }}</span></label>
        <div class="cors-guide">
          <div>
            <strong>{{ t('admin.downloadResources.corsTitle') }}</strong>
            <p>{{ t('admin.downloadResources.corsDescription') }}</p>
          </div>
          <button class="button secondary" type="button" @click="copyCORSConfig"><Icon name="copy" size="sm" />{{ t('admin.downloadResources.copyCORS') }}</button>
          <pre>{{ corsConfig }}</pre>
        </div>
        <div v-if="canManage" class="storage-actions">
          <button class="button secondary" type="button" :disabled="testingStorage" @click="testStorage"><Icon name="shield" size="sm" />{{ testingStorage ? t('common.loading') : t('admin.downloadResources.testConnection') }}</button>
          <button class="button primary" type="button" :disabled="savingStorage" @click="saveStorage"><Icon name="check" size="sm" />{{ savingStorage ? t('common.saving') : t('common.save') }}</button>
        </div>
      </section>

      <section class="panel resource-panel">
        <div class="panel-heading"><div><h2>{{ t('admin.downloadResources.resourceList') }}</h2><p>{{ t('admin.downloadResources.resourceListDescription') }}</p></div></div>
        <div class="table-shell">
          <table>
            <thead><tr><th>{{ t('admin.downloadResources.resource') }}</th><th>{{ t('downloads.platform') }}</th><th>{{ t('downloads.size') }}</th><th>{{ t('downloads.updated') }}</th><th>{{ t('downloads.downloads') }}</th><th>{{ t('admin.downloadResources.visibility') }}</th><th>{{ t('common.actions') }}</th></tr></thead>
            <tbody>
              <tr v-for="item in resources" :key="item.id">
                <td><strong>{{ displayName(item) }}</strong><span class="subline">{{ item.version || '-' }} - {{ item.file_name }}</span></td>
                <td>{{ item.platform || '-' }}</td><td>{{ formatBytes(item.size_bytes) }}</td><td>{{ formatDate(item.uploaded_at) }}</td><td>{{ numberFormatter.format(item.download_count) }}</td>
                <td><span :class="['status-chip', item.published ? 'ready' : 'missing']">{{ item.published ? t('admin.downloadResources.published') : t('admin.downloadResources.draft') }}</span></td>
                <td><div v-if="canManage" class="row-actions"><button class="icon-button" type="button" :title="t('common.edit')" @click="openEdit(item)"><Icon name="edit" size="sm" /></button><button class="icon-button danger" type="button" :title="t('common.delete')" @click="removeResource(item)"><Icon name="trash" size="sm" /></button></div></td>
              </tr>
            </tbody>
          </table>
          <div v-if="!loading && !resources.length" class="empty-table">{{ t('admin.downloadResources.noResources') }}</div>
        </div>
      </section>

      <section class="panel records-panel">
        <div class="panel-heading"><div><h2>{{ t('admin.downloadResources.downloadRecords') }}</h2><p>{{ t('admin.downloadResources.downloadRecordsDescription') }}</p></div><button class="button secondary" type="button" :disabled="loadingRecords" @click="loadRecords"><Icon name="refresh" size="sm" />{{ t('common.refresh') }}</button></div>
        <div class="table-shell compact"><table><thead><tr><th>{{ t('admin.downloadResources.requestedAt') }}</th><th>{{ t('admin.downloadResources.resource') }}</th><th>{{ t('admin.downloadResources.identity') }}</th><th>{{ t('admin.downloadResources.location') }}</th><th>{{ t('admin.downloadResources.referrer') }}</th></tr></thead><tbody><tr v-for="record in records.items" :key="record.id"><td>{{ formatDateTime(record.requested_at) }}</td><td>{{ record.resource_name }} <span class="subline">{{ record.version }}</span></td><td><div class="identity-ip-cell"><span v-if="record.username || record.email" class="user-cell">{{ record.username || record.email }}</span><div class="ip-cell"><code>{{ record.ip }}</code><button class="inline-lookup" :disabled="lookingUpIP === record.ip" :title="t('admin.downloadResources.lookup')" @click="lookupIP(record.ip)"><Icon name="search" size="xs" /></button></div></div></td><td><div v-if="record.geo_country || record.geo_region || record.geo_city" class="location-cell"><strong>{{ record.geo_country }}</strong><span>{{ [record.geo_region, record.geo_city].filter(Boolean).join(' · ') }}</span></div><button v-else class="lookup-button" :disabled="lookingUpIP === record.ip" @click="lookupIP(record.ip)">{{ lookingUpIP === record.ip ? t('common.loading') : t('admin.downloadResources.lookup') }}</button></td><td class="truncate" :title="record.referrer">{{ record.referrer || '-' }}</td></tr></tbody></table><div v-if="!loadingRecords && !records.items.length" class="empty-table">{{ t('admin.downloadResources.noRecords') }}</div></div>
      </section>
    </div>

    <div v-if="editorOpen" class="modal-backdrop" @click.self="closeEditor">
      <form class="editor-modal" @submit.prevent="saveResource">
        <header><div><h2>{{ editing ? t('admin.downloadResources.editResource') : t('admin.downloadResources.addResource') }}</h2><p>{{ t('admin.downloadResources.resourceFormDescription') }}</p></div><button type="button" class="icon-button" @click="closeEditor"><Icon name="x" size="sm" /></button></header>
        <section class="upload-box"><div><strong>{{ t('admin.downloadResources.file') }}</strong><p>{{ uploadSummary }}</p></div><label v-if="canManage" class="button secondary upload-trigger"><Icon name="upload" size="sm" />{{ uploading ? t('admin.downloadResources.uploading') : t('admin.downloadResources.uploadFile') }}<input type="file" :disabled="uploading" @change="uploadFile" /></label></section>
        <div v-if="uploading" class="upload-progress" role="status" aria-live="polite">
          <div class="upload-progress-heading">
            <div><strong>{{ t('admin.downloadResources.uploadProgress') }}</strong><span>{{ uploadProgressPercent }}%</span></div>
            <button class="icon-button" type="button" :title="t('admin.downloadResources.cancelUpload')" :aria-label="t('admin.downloadResources.cancelUpload')" @click="cancelUpload"><Icon name="x" size="sm" /></button>
          </div>
          <div class="upload-progress-track" role="progressbar" :aria-label="t('admin.downloadResources.uploadProgress')" aria-valuemin="0" aria-valuemax="100" :aria-valuenow="uploadProgressPercent"><span :style="{ width: `${uploadProgressPercent}%` }" /></div>
          <p>{{ uploadTransferSummary }}</p>
        </div>
        <p v-if="autoMetadataMessage" class="metadata-message"><Icon name="check" size="sm" />{{ autoMetadataMessage }}</p>
        <div class="form-grid">
          <label><span>{{ t('admin.downloadResources.slug') }}</span><input v-model.trim="form.slug" required readonly pattern="[a-z0-9-]+" :placeholder="t('admin.downloadResources.slugAuto')" /></label>
          <label><span>{{ t('downloads.version') }}</span><input v-model.trim="form.version" :placeholder="t('admin.downloadResources.versionAuto')" /></label>
          <label><span>{{ t('admin.downloadResources.nameZh') }}</span><input v-model.trim="form.name_zh" /></label>
          <label><span>{{ t('admin.downloadResources.nameEn') }}</span><input v-model.trim="form.name_en" /></label>
          <label><span>{{ t('downloads.platform') }}</span><input v-model.trim="form.platform" placeholder="Windows / macOS / Linux" /></label>
          <label><span>{{ t('admin.downloadResources.sortOrder') }}</span><input v-model.number="form.sort_order" type="number" /></label>
          <label class="wide"><span>{{ t('admin.downloadResources.descriptionZh') }}</span><textarea v-model.trim="form.description_zh" rows="2" /></label>
          <label class="wide"><span>{{ t('admin.downloadResources.descriptionEn') }}</span><textarea v-model.trim="form.description_en" rows="2" /></label>
        </div>
        <div class="form-grid"><label class="wide"><span>{{ t('admin.downloadResources.objectKey') }}</span><input v-model.trim="form.object_key" required /></label><label><span>{{ t('admin.downloadResources.fileName') }}</span><input v-model.trim="form.file_name" required /></label><label><span>{{ t('admin.downloadResources.contentType') }}</span><input v-model.trim="form.content_type" placeholder="application/octet-stream" /></label><label class="wide"><span>{{ t('downloads.checksum') }}</span><input v-model.trim="form.checksum_sha256" maxlength="64" /></label></div>
        <label class="checkbox-row"><input v-model="form.published" type="checkbox" /><span>{{ t('admin.downloadResources.published') }}</span></label>
        <p v-if="formError" class="form-error">{{ formError }}</p>
        <footer><button class="button secondary" type="button" @click="closeEditor">{{ t('common.cancel') }}</button><button class="button primary" :disabled="savingResource || uploading" type="submit">{{ savingResource ? t('common.saving') : t('common.save') }}</button></footer>
      </form>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import downloadResourcesAPI, { type AdminDownloadResource, type DownloadRecordPage, type DownloadResourcePayload, type DownloadResourceStorageConfig } from '@/api/admin/downloadResources'
import { useAppStore, useAuthStore } from '@/stores'
import { formatBytes } from '@/utils/format'
import { createUniqueDownloadResourceSlug, inspectDownloadResourceFile } from '@/utils/downloadResourceMetadata'
import { useClipboard } from '@/composables/useClipboard'

const MAX_UPLOAD_BYTES = 2 * 1024 * 1024 * 1024
const { t, locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const resources = ref<AdminDownloadResource[]>([])
const records = ref<DownloadRecordPage>({ items: [], total: 0, page: 1, page_size: 20, pages: 1 })
const lookingUpIP = ref('')
const loading = ref(true); const loadingRecords = ref(false); const savingStorage = ref(false); const testingStorage = ref(false)
const editorOpen = ref(false); const editing = ref<AdminDownloadResource | null>(null); const uploading = ref(false); const savingResource = ref(false); const formError = ref('')
const autoMetadataMessage = ref('')
const uploadProgress = ref(0); const uploadedBytes = ref(0); const uploadTotalBytes = ref(0); const uploadSpeedBytesPerSecond = ref(0)
const activeUploadController = ref<AbortController | null>(null)
const storage = reactive<DownloadResourceStorageConfig>({ endpoint: '', region: 'auto', bucket: '', access_key_id: '', secret_access_key: '', prefix: 'downloads/', force_path_style: true })
const form = reactive<DownloadResourcePayload>(emptyForm())
const canManage = computed(() => !authStore.isReadonlyAdmin)
const storageConfigured = computed(() => Boolean(storage.endpoint && storage.bucket && storage.access_key_id))
const publishedCount = computed(() => resources.value.filter((item) => item.published).length)
const totalDownloads = computed(() => resources.value.reduce((sum, item) => sum + item.download_count, 0))
const numberFormatter = computed(() => new Intl.NumberFormat(locale.value.startsWith('zh') ? 'zh-CN' : 'en-US'))
const uploadSummary = computed(() => form.file_name || t('admin.downloadResources.fileHint'))
const uploadProgressPercent = computed(() => Math.min(100, Math.max(0, Math.round(uploadProgress.value))))
const uploadTransferSummary = computed(() => {
  const total = uploadTotalBytes.value || 0
  const transferred = total > 0 ? `${formatBytes(uploadedBytes.value)} / ${formatBytes(total)}` : formatBytes(uploadedBytes.value)
  return uploadSpeedBytesPerSecond.value > 0 ? `${transferred} · ${formatBytes(uploadSpeedBytesPerSecond.value)}/s` : transferred
})
const corsConfig = computed(() => JSON.stringify([{
  AllowedOrigins: Array.from(new Set([window.location.origin, 'https://useaifor.me', 'https://useaifor.fun'])),
  AllowedMethods: ['PUT'],
  AllowedHeaders: ['Content-Type'],
  ExposeHeaders: ['ETag'],
  MaxAgeSeconds: 3600,
}], null, 2))

function emptyForm(): DownloadResourcePayload { return { slug: '', name_zh: '', name_en: '', description_zh: '', description_en: '', version: '', platform: '', object_key: '', file_name: '', content_type: 'application/octet-stream', checksum_sha256: '', published: true, sort_order: 0 } }
function displayName(item: AdminDownloadResource) { return locale.value.startsWith('zh') ? item.name_zh || item.name_en || item.file_name : item.name_en || item.name_zh || item.file_name }
function formatDate(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? '-' : new Intl.DateTimeFormat(locale.value.startsWith('zh') ? 'zh-CN' : 'en-US', { year: 'numeric', month: 'short', day: 'numeric' }).format(date) }
function formatDateTime(value: string) { const date = new Date(value); return Number.isNaN(date.getTime()) ? '-' : new Intl.DateTimeFormat(locale.value.startsWith('zh') ? 'zh-CN' : 'en-US', { dateStyle: 'short', timeStyle: 'short' }).format(date) }
function assignStorage(value: DownloadResourceStorageConfig | null) { Object.assign(storage, { endpoint: value?.endpoint || '', region: value?.region || 'auto', bucket: value?.bucket || '', access_key_id: value?.access_key_id || '', secret_access_key: '', prefix: value?.prefix || 'downloads/', force_path_style: value?.force_path_style ?? true }) }
async function refresh() { loading.value = true; try { resources.value = await downloadResourcesAPI.list(); if (canManage.value) assignStorage(await downloadResourcesAPI.storage()) } finally { loading.value = false } }
async function loadRecords() { loadingRecords.value = true; try { records.value = await downloadResourcesAPI.listDownloads() } finally { loadingRecords.value = false } }
async function lookupIP(ip: string) {
  if (!ip || lookingUpIP.value) return
  lookingUpIP.value = ip
  try {
    const result = await downloadResourcesAPI.lookupIP(ip)
    for (const record of records.value.items) {
      if (record.ip === ip) Object.assign(record, { geo_country: result.country, geo_region: result.region, geo_city: result.city })
    }
  } catch { appStore.showError(t('admin.downloadResources.lookupFailed')) }
  finally { lookingUpIP.value = '' }
}
async function saveStorage() { savingStorage.value = true; try { const updated = await downloadResourcesAPI.saveStorage({ ...storage }); assignStorage(updated); appStore.showSuccess(t('admin.downloadResources.storageSaved')) } catch (error) { appStore.showError(requestErrorMessage(error, t('admin.downloadResources.saveFailed'))) } finally { savingStorage.value = false } }
async function testStorage() { testingStorage.value = true; try { await downloadResourcesAPI.testStorage({ ...storage }); appStore.showSuccess(t('admin.downloadResources.testSuccess')) } catch (error) { appStore.showError(requestErrorMessage(error, t('admin.downloadResources.testFailed'))) } finally { testingStorage.value = false } }
function requestErrorMessage(error: unknown, fallback: string) { return (error as { message?: string })?.message || fallback }
function uploadErrorMessage(error: unknown) {
  const status = (error as { response?: { status?: number } })?.response?.status
  if (status) return t('admin.downloadResources.uploadFailedStatus', { status })
  const message = (error as { message?: string })?.message || ''
  return /failed to fetch|networkerror|load failed/i.test(message)
    ? t('admin.downloadResources.uploadFailed')
    : message || t('admin.downloadResources.uploadFailed')
}
async function copyCORSConfig() { await copyToClipboard(corsConfig.value, t('admin.downloadResources.corsCopied')) }
function resetUploadProgress(total = 0) { uploadProgress.value = 0; uploadedBytes.value = 0; uploadTotalBytes.value = total; uploadSpeedBytesPerSecond.value = 0 }
function cancelUpload() { activeUploadController.value?.abort(); uploading.value = false }
function closeEditor() { cancelUpload(); editorOpen.value = false }
function openCreate() { editing.value = null; Object.assign(form, emptyForm()); formError.value = ''; autoMetadataMessage.value = ''; resetUploadProgress(); editorOpen.value = true }
function openEdit(item: AdminDownloadResource) { editing.value = item; Object.assign(form, { slug: item.slug, name_zh: item.name_zh, name_en: item.name_en, description_zh: item.description_zh, description_en: item.description_en, version: item.version, platform: item.platform, object_key: item.object_key, file_name: item.file_name, content_type: item.content_type, checksum_sha256: item.checksum_sha256, published: item.published, sort_order: item.sort_order }); formError.value = ''; autoMetadataMessage.value = ''; resetUploadProgress(); editorOpen.value = true }
async function uploadFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  if (file.size > MAX_UPLOAD_BYTES) { formError.value = t('admin.downloadResources.fileTooLarge'); input.value = ''; return }
  const controller = new AbortController()
  activeUploadController.value = controller
  resetUploadProgress(file.size)
  uploading.value = true; formError.value = ''; autoMetadataMessage.value = ''
  try {
    const metadata = await inspectDownloadResourceFile(file)
    if (controller.signal.aborted) return
    if (!editing.value) {
      form.slug = createUniqueDownloadResourceSlug(metadata.slug, resources.value.map((item) => item.slug))
      if (!form.version) form.version = metadata.version
      if (!form.platform) form.platform = metadata.platform
      if (!form.name_zh) form.name_zh = metadata.name
      if (!form.name_en) form.name_en = metadata.name
    }
    autoMetadataMessage.value = metadata.version
      ? t('admin.downloadResources.metadataDetected', { platform: metadata.platform || '-', version: metadata.version })
      : t('admin.downloadResources.metadataFallback')

    const contentType = file.type || 'application/octet-stream'
    const upload = await downloadResourcesAPI.createUploadURL(file.name, contentType, file.size)
    const startedAt = performance.now()
    await axios.put(upload.upload_url, file, {
      headers: { 'Content-Type': contentType },
      signal: controller.signal,
      onUploadProgress: (event) => {
        const total = event.total || file.size
        const loaded = Math.min(event.loaded, total)
        const elapsedSeconds = Math.max((performance.now() - startedAt) / 1000, 0.001)
        uploadTotalBytes.value = total
        uploadedBytes.value = loaded
        uploadProgress.value = total > 0 ? (loaded / total) * 100 : 0
        uploadSpeedBytesPerSecond.value = loaded / elapsedSeconds
      },
    })
    uploadedBytes.value = file.size; uploadProgress.value = 100
    form.object_key = upload.object_key; form.file_name = file.name; form.content_type = contentType
  } catch (error) { if (!axios.isCancel(error) && (error as { name?: string })?.name !== 'AbortError') formError.value = uploadErrorMessage(error) }
  finally { if (activeUploadController.value === controller) { activeUploadController.value = null; uploading.value = false } input.value = '' }
}
async function saveResource() { savingResource.value = true; formError.value = ''; try { if (editing.value) await downloadResourcesAPI.update(editing.value.id, { ...form }); else await downloadResourcesAPI.create({ ...form }); editorOpen.value = false; await refresh() } catch { formError.value = t('admin.downloadResources.saveFailed') } finally { savingResource.value = false } }
async function removeResource(item: AdminDownloadResource) { if (!canManage.value || !window.confirm(t('admin.downloadResources.deleteConfirm', { name: displayName(item) }))) return; try { await downloadResourcesAPI.remove(item.id); await refresh() } catch { return } }
onMounted(async () => { await Promise.all([refresh(), loadRecords()]) })
onBeforeUnmount(cancelUpload)
</script>

<style scoped>
.cors-guide { position: relative; margin-top: 16px; border: 1px solid #cbd5e1; border-radius: 7px; background: #f8fafc; padding: 14px; }.cors-guide > div { padding-right: 130px; }.cors-guide strong { color: #334155; font-size: 13px; }.cors-guide p { margin: 4px 0 0; color: #64748b; font-size: 12px; line-height: 1.5; }.cors-guide .button { position: absolute; top: 12px; right: 12px; }.cors-guide pre { max-height: 210px; overflow: auto; margin: 12px 0 0; border-radius: 6px; background: #0f172a; padding: 12px; color: #d1fae5; font-size: 11px; line-height: 1.55; white-space: pre-wrap; word-break: break-word; }:global(.dark) .cors-guide { border-color: #334155; background: #0f172a; }:global(.dark) .cors-guide strong { color: #e2e8f0; }:global(.dark) .cors-guide pre { background: #020617; }@media (max-width: 620px) { .cors-guide > div { padding-right: 0; }.cors-guide .button { position: static; width: 100%; margin-top: 12px; } }
.resource-admin { max-width: 1480px; margin: 0 auto; padding: 8px 0 36px; }.page-header, .panel-heading, .storage-actions, .page-actions, .row-actions, .checkbox-row { display: flex; align-items: center; }.page-header, .panel-heading { justify-content: space-between; gap: 22px; }.eyebrow { display: flex; align-items: center; gap: 7px; margin: 0; color: #0f766e; font-size: 12px; font-weight: 700; letter-spacing: .07em; }.page-header h1 { margin: 8px 0 0; font-size: 28px; color: #172033; }.page-header p:not(.eyebrow), .panel-heading p { margin: 6px 0 0; color: #64748b; font-size: 14px; line-height: 1.5; }.page-actions, .storage-actions { gap: 9px; }.button { display: inline-flex; align-items: center; justify-content: center; gap: 7px; border: 1px solid transparent; border-radius: 7px; padding: 9px 13px; cursor: pointer; font: inherit; font-size: 13px; font-weight: 650; }.button:disabled { cursor: wait; opacity: .6; }.button.primary { background: #0f766e; color: white; }.button.secondary { border-color: #d7e0e9; background: white; color: #475569; }.metrics { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-top: 24px; }.metrics article, .panel { border: 1px solid #dbe3ed; border-radius: 8px; background: #fff; }.metrics article { padding: 16px; }.metrics span { display: block; color: #64748b; font-size: 12px; }.metrics strong { display: block; margin-top: 7px; color: #1e293b; font-size: 22px; }.metrics strong.success { color: #059669; font-size: 15px; }.metrics strong.muted { color: #94a3b8; font-size: 15px; }.panel { margin-top: 16px; padding: 20px; }.panel-heading h2 { margin: 0; color: #1e293b; font-size: 17px; }.storage-grid, .form-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 13px; margin-top: 18px; }.storage-grid label, .form-grid label { display: grid; gap: 6px; color: #475569; font-size: 12px; font-weight: 600; }.storage-grid input, .form-grid input, .form-grid textarea { min-width: 0; border: 1px solid #cbd5e1; border-radius: 6px; background: white; padding: 9px 10px; color: #1e293b; font: inherit; font-size: 13px; outline: none; }.storage-grid input:focus, .form-grid input:focus, .form-grid textarea:focus { border-color: #14b8a6; box-shadow: 0 0 0 3px rgba(20,184,166,.12); }.checkbox-row { gap: 8px; margin-top: 15px; color: #475569; font-size: 13px; }.storage-actions { justify-content: flex-end; margin-top: 16px; }.status-chip { display: inline-flex; align-items: center; border-radius: 999px; padding: 4px 8px; font-size: 11px; font-weight: 650; }.status-chip.ready { background: #d1fae5; color: #047857; }.status-chip.missing { background: #f1f5f9; color: #64748b; }.table-shell { margin-top: 17px; overflow-x: auto; border: 1px solid #e2e8f0; border-radius: 7px; }.table-shell table { width: 100%; min-width: 900px; border-collapse: collapse; }.table-shell th { background: #f8fafc; color: #64748b; font-size: 12px; font-weight: 650; text-align: left; }.table-shell td, .table-shell th { padding: 12px; border-bottom: 1px solid #e2e8f0; vertical-align: middle; }.table-shell td { color: #475569; font-size: 13px; }.table-shell tbody tr:last-child td { border-bottom: 0; }.table-shell td strong { display: block; color: #1e293b; }.subline { display: block; margin-top: 3px; color: #94a3b8; font-size: 11px; }.row-actions { gap: 5px; }.icon-button { display: grid; width: 31px; height: 31px; place-items: center; border: 1px solid #dbe3ed; border-radius: 6px; background: white; color: #475569; cursor: pointer; }.icon-button.danger { color: #dc2626; }.empty-table { padding: 38px; color: #94a3b8; font-size: 14px; text-align: center; }.truncate { max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }.modal-backdrop { position: fixed; z-index: 60; inset: 0; display: grid; place-items: center; padding: 18px; background: rgba(15,23,42,.5); }.editor-modal { width: min(780px, 100%); max-height: calc(100vh - 36px); overflow-y: auto; border-radius: 8px; background: white; padding: 22px; box-shadow: 0 24px 80px rgba(15,23,42,.3); }.editor-modal header, .editor-modal footer { display: flex; align-items: start; justify-content: space-between; gap: 16px; }.editor-modal h2 { margin: 0; color: #1e293b; font-size: 19px; }.editor-modal header p { margin: 5px 0 0; color: #64748b; font-size: 13px; }.form-grid .wide { grid-column: span 2; }.upload-box { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-top: 16px; border: 1px dashed #99f6e4; border-radius: 7px; background: #f0fdfa; padding: 13px; }.upload-box strong { color: #115e59; font-size: 13px; }.upload-box p { margin: 4px 0 0; color: #64748b; font-size: 12px; }.upload-trigger { position: relative; overflow: hidden; }.upload-trigger input { position: absolute; inset: 0; cursor: pointer; opacity: 0; }.editor-modal footer { justify-content: flex-end; margin-top: 21px; }.form-error { margin: 13px 0 0; color: #dc2626; font-size: 13px; }.spinning { animation: spin 1s linear infinite; }@keyframes spin { to { transform: rotate(360deg); } }:global(.dark) .page-header h1, :global(.dark) .panel-heading h2, :global(.dark) .metrics strong, :global(.dark) .table-shell td strong, :global(.dark) .editor-modal h2 { color: #f8fafc; }:global(.dark) .metrics article, :global(.dark) .panel, :global(.dark) .editor-modal { border-color: #334155; background: #111827; }:global(.dark) .button.secondary, :global(.dark) .icon-button, :global(.dark) .storage-grid input, :global(.dark) .form-grid input, :global(.dark) .form-grid textarea { border-color: #334155; background: #0f172a; color: #dbeafe; }:global(.dark) .table-shell { border-color: #334155; }:global(.dark) .table-shell th { background: #0f172a; }:global(.dark) .table-shell td, :global(.dark) .table-shell th { border-color: #334155; }:global(.dark) .upload-box { border-color: #115e59; background: rgba(6,78,59,.22); }@media (max-width: 900px) { .metrics { grid-template-columns: repeat(2, 1fr); }.storage-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.form-grid { grid-template-columns: 1fr; }.form-grid .wide { grid-column: auto; } }@media (max-width: 620px) { .page-header, .panel-heading { align-items: start; flex-direction: column; }.page-actions { width: 100%; }.page-actions .button { flex: 1; }.storage-grid, .metrics { grid-template-columns: 1fr; }.storage-actions { justify-content: stretch; }.storage-actions .button { flex: 1; }.upload-box { align-items: start; flex-direction: column; }.upload-trigger { width: 100%; } }
.metadata-message { display: flex; align-items: center; gap: 6px; margin: 9px 0 -4px; color: #047857; font-size: 12px; }
.upload-progress { margin-top: 10px; border: 1px solid #ccfbf1; border-radius: 7px; background: #f8fffe; padding: 11px 12px; }.upload-progress-heading, .upload-progress-heading > div { display: flex; align-items: center; }.upload-progress-heading { justify-content: space-between; gap: 12px; }.upload-progress-heading > div { gap: 8px; }.upload-progress-heading strong { color: #115e59; font-size: 12px; }.upload-progress-heading span { color: #0f766e; font-size: 12px; font-variant-numeric: tabular-nums; }.upload-progress-heading .icon-button { width: 27px; height: 27px; }.upload-progress-track { height: 7px; overflow: hidden; margin-top: 8px; border-radius: 4px; background: #dbe7e5; }.upload-progress-track span { display: block; height: 100%; border-radius: inherit; background: #0d9488; transition: width .15s ease-out; }.upload-progress p { margin: 7px 0 0; color: #64748b; font-size: 11px; font-variant-numeric: tabular-nums; }:global(.dark) .upload-progress { border-color: #134e4a; background: rgba(6,78,59,.16); }:global(.dark) .upload-progress-heading strong { color: #99f6e4; }:global(.dark) .upload-progress-heading span { color: #5eead4; }:global(.dark) .upload-progress-track { background: #334155; }
.form-grid input[readonly] { background: #f8fafc; color: #64748b; }
.identity-ip-cell { display: grid; gap: 3px; }
.ip-cell { display: flex; align-items: center; gap: 6px; white-space: nowrap; }.inline-lookup { display: grid; width: 24px; height: 24px; place-items: center; border: 1px solid #99f6e4; border-radius: 6px; background: #f0fdfa; color: #0f766e; cursor: pointer; }.inline-lookup:disabled, .lookup-button:disabled { cursor: wait; opacity: .55; }.lookup-button { border: 1px solid #99f6e4; border-radius: 6px; background: #f0fdfa; padding: 5px 8px; color: #0f766e; cursor: pointer; font-size: 11px; }.user-cell { display: block; max-width: 190px; overflow: hidden; color: #334155; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }.anonymous-cell { color: #94a3b8; }.location-cell strong, .location-cell span { display: block; white-space: nowrap; }.location-cell strong { color: #334155; }.location-cell span { margin-top: 2px; color: #94a3b8; font-size: 11px; }
:global(.dark) .form-grid input[readonly] { background: #111827; color: #94a3b8; }
</style>
