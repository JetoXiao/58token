<template>
  <AppLayout>
    <div class="visitor-analytics-page">
    <section class="analytics-hero">
      <div class="hero-copy">
        <span class="eyebrow">{{ t('admin.visitorAnalytics.eyebrow') }}</span>
        <h1>{{ t('admin.visitorAnalytics.heroTitle') }}</h1>
        <p>{{ t('admin.visitorAnalytics.heroDescription') }}</p>
      </div>
      <div class="hero-actions">
        <div class="date-control">
          <span>{{ t('admin.visitorAnalytics.dateRange') }}</span>
          <div class="date-inputs">
            <input v-model="filters.start_date" type="date" :max="filters.end_date" />
            <i>—</i>
            <input v-model="filters.end_date" type="date" :min="filters.start_date" />
          </div>
        </div>
        <button class="primary-button" :disabled="loading" @click="refreshAll">
          <Icon name="refresh" size="sm" />
          {{ t('admin.visitorAnalytics.refresh') }}
        </button>
      </div>
      <div class="privacy-note">
        <Icon name="shield" size="sm" />
        {{ t('admin.visitorAnalytics.privacyNote') }}
      </div>
    </section>

    <section class="metric-grid" aria-live="polite">
      <article v-for="metric in metrics" :key="metric.label" class="metric-card">
        <div :class="['metric-icon', metric.tone]"><Icon :name="metric.icon" size="md" /></div>
        <div>
          <span>{{ metric.label }}</span>
          <strong>{{ numberFormatter.format(metric.value) }}</strong>
        </div>
      </article>
    </section>

    <section class="analytics-grid">
      <article class="panel trend-panel">
        <div class="panel-heading">
          <div>
            <h2>{{ t('admin.visitorAnalytics.trendTitle') }}</h2>
            <p>{{ t('admin.visitorAnalytics.trendDescription') }}</p>
          </div>
          <div class="trend-legend"><span class="bar-dot" />{{ t('admin.visitorAnalytics.pageViews') }} <span class="line-dot" />{{ t('admin.visitorAnalytics.uniqueVisitors') }}</div>
        </div>
        <div v-if="loading" class="chart-skeleton" />
        <div v-else-if="trend.length" class="trend-chart">
          <div v-for="point in trend" :key="point.date" class="trend-column" :title="`${point.date}: ${point.page_views} / ${point.unique_visitors}`">
            <div class="bar-track">
              <div class="visitor-marker" :style="{ bottom: markerHeight(point.unique_visitors) }" />
              <div class="view-bar" :style="{ height: barHeight(point.page_views) }" />
            </div>
            <span>{{ shortDate(point.date) }}</span>
          </div>
        </div>
        <div v-else class="empty-chart">{{ t('admin.visitorAnalytics.noTrend') }}</div>
      </article>

      <article class="panel settings-panel">
        <div class="panel-heading">
          <div>
            <h2>{{ t('admin.visitorAnalytics.trackingSettings') }}</h2>
            <p>{{ t('admin.visitorAnalytics.trackingDescription') }}</p>
          </div>
        </div>
        <label class="setting-row">
          <div>
            <strong>{{ t('admin.visitorAnalytics.enabled') }}</strong>
            <span>{{ settings.enabled ? t('admin.visitorAnalytics.enabled') : t('admin.visitorAnalytics.disabled') }}</span>
          </div>
          <input v-model="settings.enabled" class="switch-input" type="checkbox" :disabled="!canManage" />
        </label>
        <label class="retention-field">
          <span>{{ t('admin.visitorAnalytics.retention') }}</span>
          <select v-model.number="settings.retention_days" :disabled="!canManage">
            <option v-for="days in retentionOptions" :key="days" :value="days">{{ t('admin.visitorAnalytics.days', { count: days }) }}</option>
          </select>
        </label>
        <button v-if="canManage" class="secondary-button full" :disabled="savingSettings" @click="saveSettings">
          {{ t('admin.visitorAnalytics.saveSettings') }}
        </button>
      </article>
    </section>

    <section class="panel channel-panel">
      <div class="panel-heading channel-heading">
        <div>
          <h2>{{ t('admin.visitorAnalytics.channelPerformance') }}</h2>
          <p>{{ t('admin.visitorAnalytics.channelDescription') }}</p>
        </div>
        <button v-if="canManage" class="secondary-button" @click="openChannelManager">
          <Icon name="plus" size="sm" />
          {{ t('admin.visitorAnalytics.manageChannels') }}
        </button>
      </div>
      <div v-if="channelStats.length" class="channel-grid">
        <article v-for="channel in channelStats" :key="channel.code" class="channel-card">
          <div class="channel-topline">
            <span class="channel-avatar">{{ channel.code === 'direct' ? '↗' : channel.name.slice(0, 1).toUpperCase() }}</span>
            <div><strong>{{ channel.code === 'direct' ? t('admin.visitorAnalytics.direct') : channel.name }}</strong><code>{{ channel.code }}</code></div>
            <span :class="['status-pill', channel.active ? 'active' : 'inactive']">{{ channel.active ? t('admin.visitorAnalytics.enabled') : t('admin.visitorAnalytics.disabled') }}</span>
          </div>
          <div class="channel-metrics">
            <div><span>{{ t('admin.visitorAnalytics.pageViews') }}</span><strong>{{ numberFormatter.format(channel.page_views) }}</strong></div>
            <div><span>{{ t('admin.visitorAnalytics.uniqueVisitors') }}</span><strong>{{ numberFormatter.format(channel.unique_visitors) }}</strong></div>
            <div><span>{{ t('admin.visitorAnalytics.uniqueIPs') }}</span><strong>{{ numberFormatter.format(channel.unique_ips) }}</strong></div>
          </div>
          <button v-if="channel.code !== 'direct'" class="copy-link" @click="copyChannelLink(channel.code, channel.destination_path)">
            <span>{{ channelLink(channel.code, channel.destination_path) }}</span>
            <Icon name="copy" size="sm" />
          </button>
        </article>
      </div>
      <div v-else class="empty-inline">{{ t('admin.visitorAnalytics.noChannels') }}</div>
    </section>

    <section class="panel records-panel">
      <div class="panel-heading records-heading">
        <div>
          <h2>{{ t('admin.visitorAnalytics.visitRecords') }}</h2>
          <p>{{ t('admin.visitorAnalytics.recordsDescription') }}</p>
        </div>
        <div class="record-filters">
          <div class="search-control">
            <Icon name="search" size="sm" />
            <input v-model="eventFilters.search" :placeholder="t('admin.visitorAnalytics.searchPlaceholder')" @keyup.enter="reloadEvents" />
          </div>
          <select v-model="eventFilters.channel_code" @change="reloadEvents">
            <option value="">{{ t('admin.visitorAnalytics.allChannels') }}</option>
            <option value="direct">{{ t('admin.visitorAnalytics.direct') }}</option>
            <option v-for="channel in channels" :key="channel.id" :value="channel.code">{{ channel.name }}</option>
          </select>
        </div>
      </div>
      <div class="table-shell">
        <table>
          <thead><tr>
            <th>{{ t('admin.visitorAnalytics.time') }}</th>
            <th>{{ t('admin.visitorAnalytics.channel') }}</th>
            <th>{{ t('admin.visitorAnalytics.ip') }}</th>
            <th>{{ t('admin.visitorAnalytics.location') }}</th>
            <th>{{ t('admin.visitorAnalytics.page') }}</th>
            <th>{{ t('admin.visitorAnalytics.referrer') }}</th>
            <th>{{ t('admin.visitorAnalytics.device') }}</th>
          </tr></thead>
          <tbody>
            <tr v-for="event in events.items" :key="event.id">
              <td class="time-cell">{{ formatDateTime(event.occurred_at) }}</td>
              <td><span class="source-pill">{{ event.channel_code === 'direct' ? t('admin.visitorAnalytics.direct') : event.channel_name }}</span></td>
              <td><button class="ip-button" @click="lookupEventIP(event.ip)">{{ event.ip }}</button></td>
              <td>
                <div v-if="event.geo_country" class="location-cell"><strong>{{ event.geo_country }}</strong><span>{{ [event.geo_region, event.geo_city].filter(Boolean).join(' · ') }}</span></div>
                <button v-else class="lookup-button" @click="lookupEventIP(event.ip)">{{ t('admin.visitorAnalytics.lookup') }}</button>
              </td>
              <td><span class="path-cell" :title="event.landing_url">{{ event.path }}</span></td>
              <td><span class="referrer-cell" :title="event.referrer">{{ hostname(event.referrer) || t('admin.visitorAnalytics.sourceUnknown') }}</span></td>
              <td><div class="device-cell"><span>{{ event.language || '—' }} · {{ event.screen || '—' }}</span><em v-if="event.is_bot">{{ t('admin.visitorAnalytics.bot') }}</em></div></td>
            </tr>
          </tbody>
        </table>
        <div v-if="!events.items.length && !loadingEvents" class="table-empty">{{ t('admin.visitorAnalytics.noRecords') }}</div>
      </div>
      <div class="pagination-row">
        <span>{{ t('admin.visitorAnalytics.pageOf', { page: events.page, pages: events.pages, total: events.total }) }}</span>
        <div>
          <button :disabled="events.page <= 1" @click="changePage(events.page - 1)">{{ t('admin.visitorAnalytics.previous') }}</button>
          <button :disabled="events.page >= events.pages" @click="changePage(events.page + 1)">{{ t('admin.visitorAnalytics.next') }}</button>
        </div>
      </div>
    </section>

    <section class="panel ip-panel">
      <div class="ip-copy">
        <span class="eyebrow">IP INTELLIGENCE</span>
        <h2>{{ t('admin.visitorAnalytics.ipLookup') }}</h2>
        <p>{{ t('admin.visitorAnalytics.ipLookupDescription') }}</p>
        <div class="ip-search">
          <input v-model="ipQuery" :placeholder="t('admin.visitorAnalytics.ipPlaceholder')" @keyup.enter="lookupIP" />
          <button class="primary-button" :disabled="lookingUpIP || !ipQuery" @click="lookupIP">{{ t('admin.visitorAnalytics.lookupAction') }}</button>
        </div>
      </div>
      <div v-if="ipResult" class="ip-result">
        <div class="result-head"><span>{{ ipResult.ip }}</span><strong>{{ [ipResult.city, ipResult.country].filter(Boolean).join(', ') }}</strong></div>
        <dl>
          <div><dt>{{ t('admin.visitorAnalytics.country') }}</dt><dd>{{ ipResult.country || '—' }} <small>{{ ipResult.country_code }}</small></dd></div>
          <div><dt>{{ t('admin.visitorAnalytics.regionCity') }}</dt><dd>{{ [ipResult.region, ipResult.city].filter(Boolean).join(' / ') || '—' }}</dd></div>
          <div><dt>{{ t('admin.visitorAnalytics.timezone') }}</dt><dd>{{ ipResult.timezone || '—' }}</dd></div>
          <div><dt>{{ t('admin.visitorAnalytics.coordinates') }}</dt><dd>{{ coordinates(ipResult) }}</dd></div>
        </dl>
      </div>
      <div v-else class="ip-placeholder-art"><Icon name="globe" size="xl" /></div>
    </section>

    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="showChannelManager" class="modal-backdrop" @click.self="closeChannelManager">
          <div class="channel-modal" role="dialog" aria-modal="true">
            <div class="modal-heading">
              <div><span class="eyebrow">ATTRIBUTION LINKS</span><h2>{{ t('admin.visitorAnalytics.manageChannels') }}</h2></div>
              <button class="icon-button" :aria-label="t('common.close')" @click="closeChannelManager"><Icon name="x" size="sm" /></button>
            </div>
            <div class="modal-body-grid">
              <div class="configured-channels">
                <button class="add-channel-card" @click="startCreateChannel">+ {{ t('admin.visitorAnalytics.createChannel') }}</button>
                <article v-for="channel in channels" :key="channel.id" :class="['configured-channel', { selected: channelForm.id === channel.id }]" @click="editChannel(channel)">
                  <span>{{ channel.name.slice(0, 1).toUpperCase() }}</span>
                  <div><strong>{{ channel.name }}</strong><code>?ref={{ channel.code }}</code></div>
                  <em>{{ channel.active ? t('admin.visitorAnalytics.enabled') : t('admin.visitorAnalytics.disabled') }}</em>
                </article>
              </div>
              <form class="channel-form" @submit.prevent="saveChannel">
                <h3>{{ channelForm.id ? t('admin.visitorAnalytics.editChannel') : t('admin.visitorAnalytics.createChannel') }}</h3>
                <label><span>{{ t('admin.visitorAnalytics.channelName') }}</span><input v-model="channelForm.name" required maxlength="100" /></label>
                <label><span>{{ t('admin.visitorAnalytics.channelCode') }}</span><input v-model="channelForm.code" :disabled="Boolean(channelForm.id)" required maxlength="64" pattern="[a-z0-9][a-z0-9_-]*" /><small>{{ t('admin.visitorAnalytics.codeHint') }}</small></label>
                <label><span>{{ t('admin.visitorAnalytics.destinationPath') }}</span><input v-model="channelForm.destination_path" required /><small>{{ t('admin.visitorAnalytics.destinationHint') }}</small></label>
                <label><span>{{ t('admin.visitorAnalytics.channelNotes') }}</span><textarea v-model="channelForm.description" rows="3" maxlength="500" /></label>
                <label class="form-toggle"><span>{{ t('admin.visitorAnalytics.enabled') }}</span><input v-model="channelForm.active" class="switch-input" type="checkbox" /></label>
                <div class="form-actions">
                  <button v-if="channelForm.id" type="button" class="danger-button" @click="deleteChannel">{{ t('admin.visitorAnalytics.delete') }}</button>
                  <span />
                  <button type="button" class="secondary-button" @click="closeChannelManager">{{ t('admin.visitorAnalytics.cancel') }}</button>
                  <button type="submit" class="primary-button" :disabled="savingChannel">{{ t('admin.visitorAnalytics.save') }}</button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import visitorAnalyticsAPI, {
  type IPGeolocation,
  type PaginatedVisitorEvents,
  type VisitorAnalyticsOverview,
  type VisitorAnalyticsSettings,
  type VisitorChannel,
  type VisitorChannelStats,
  type VisitorTrendPoint,
} from '@/api/admin/visitorAnalytics'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const canManage = computed(() => !authStore.isReadonlyAdmin)

const toLocalISODate = (date: Date) => {
  const shifted = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return shifted.toISOString().slice(0, 10)
}
const today = new Date()
const weekAgo = new Date(today)
weekAgo.setDate(today.getDate() - 6)

const filters = reactive({ start_date: toLocalISODate(weekAgo), end_date: toLocalISODate(today) })
const eventFilters = reactive({ search: '', channel_code: '' })
const overview = ref<VisitorAnalyticsOverview>({ page_views: 0, unique_visitors: 0, unique_ips: 0, active_channels: 0 })
const trend = ref<VisitorTrendPoint[]>([])
const channelStats = ref<VisitorChannelStats[]>([])
const channels = ref<VisitorChannel[]>([])
const settings = reactive<VisitorAnalyticsSettings>({ enabled: true, retention_days: 90, updated_at: '' })
const events = reactive<PaginatedVisitorEvents>({ items: [], total: 0, page: 1, page_size: 20, pages: 1 })
const loading = ref(false)
const loadingEvents = ref(false)
const savingSettings = ref(false)
const savingChannel = ref(false)
const lookingUpIP = ref(false)
const showChannelManager = ref(false)
const ipQuery = ref('')
const ipResult = ref<IPGeolocation | null>(null)
const retentionOptions = [30, 60, 90, 180, 365, 730]

const channelForm = reactive({ id: 0, name: '', code: '', destination_path: '/home', description: '', active: true })
const numberFormatter = computed(() => new Intl.NumberFormat(locale.value === 'zh' ? 'zh-CN' : 'en-US')).value
const maxTrend = computed(() => Math.max(1, ...trend.value.flatMap((point) => [point.page_views, point.unique_visitors])))

const metrics = computed(() => [
  { label: t('admin.visitorAnalytics.pageViews'), value: overview.value.page_views, tone: 'violet', icon: 'chart' as const },
  { label: t('admin.visitorAnalytics.uniqueVisitors'), value: overview.value.unique_visitors, tone: 'cyan', icon: 'users' as const },
  { label: t('admin.visitorAnalytics.uniqueIPs'), value: overview.value.unique_ips, tone: 'green', icon: 'globe' as const },
  { label: t('admin.visitorAnalytics.activeChannels'), value: overview.value.active_channels, tone: 'amber', icon: 'link' as const },
])

async function refreshAll() {
  loading.value = true
  try {
    const [summary, trendItems, stats, channelItems, settingsItem] = await Promise.all([
      visitorAnalyticsAPI.overview(filters), visitorAnalyticsAPI.trend(filters), visitorAnalyticsAPI.channelStats(filters),
      visitorAnalyticsAPI.channels(), visitorAnalyticsAPI.settings(),
    ])
    overview.value = summary
    trend.value = trendItems
    channelStats.value = stats
    channels.value = channelItems
    Object.assign(settings, settingsItem)
    await reloadEvents()
  } catch {
    appStore.showError(t('admin.visitorAnalytics.requestFailed'))
  } finally {
    loading.value = false
  }
}

async function reloadEvents() {
  loadingEvents.value = true
  try {
    const result = await visitorAnalyticsAPI.events({ ...filters, page: events.page, page_size: events.page_size,
      channel_code: eventFilters.channel_code || undefined, search: eventFilters.search.trim() || undefined })
    Object.assign(events, result)
  } catch {
    appStore.showError(t('admin.visitorAnalytics.requestFailed'))
  } finally {
    loadingEvents.value = false
  }
}

function changePage(page: number) { events.page = page; void reloadEvents() }
function barHeight(value: number) { return `${Math.max(value ? 8 : 0, (value / maxTrend.value) * 100)}%` }
function markerHeight(value: number) { return `calc(${Math.min(96, (value / maxTrend.value) * 100)}% - 4px)` }
function shortDate(value: string) { return value.slice(5).replace('-', '/') }
function formatDateTime(value: string) { return new Intl.DateTimeFormat(locale.value === 'zh' ? 'zh-CN' : 'en-US', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }).format(new Date(value)) }
function hostname(value: string) { try { return value ? new URL(value).hostname : '' } catch { return value } }
function channelLink(code: string, destination: string) { const url = new URL(destination || '/home', window.location.origin); url.searchParams.set('ref', code); return url.toString() }
async function copyChannelLink(code: string, destination: string) { await navigator.clipboard.writeText(channelLink(code, destination)); appStore.showSuccess(t('admin.visitorAnalytics.linkCopied')) }
function coordinates(result: IPGeolocation) { return result.latitude == null || result.longitude == null ? '—' : `${result.latitude.toFixed(4)}, ${result.longitude.toFixed(4)}` }

async function saveSettings() {
  savingSettings.value = true
  try { Object.assign(settings, await visitorAnalyticsAPI.updateSettings(settings)); appStore.showSuccess(t('admin.visitorAnalytics.settingsSaved')) }
  catch { appStore.showError(t('admin.visitorAnalytics.operationFailed')) }
  finally { savingSettings.value = false }
}

function resetChannelForm() { Object.assign(channelForm, { id: 0, name: '', code: '', destination_path: '/home', description: '', active: true }) }
function openChannelManager() { if (!canManage.value) return; resetChannelForm(); showChannelManager.value = true }
function closeChannelManager() { showChannelManager.value = false }
function startCreateChannel() { resetChannelForm() }
function editChannel(channel: VisitorChannel) { Object.assign(channelForm, channel) }
async function saveChannel() {
  savingChannel.value = true
  try {
    const payload = { name: channelForm.name, code: channelForm.code.toLowerCase(), destination_path: channelForm.destination_path, description: channelForm.description, active: channelForm.active }
    if (channelForm.id) await visitorAnalyticsAPI.updateChannel(channelForm.id, payload)
    else await visitorAnalyticsAPI.createChannel(payload)
    appStore.showSuccess(t('admin.visitorAnalytics.channelSaved')); resetChannelForm(); channels.value = await visitorAnalyticsAPI.channels(); channelStats.value = await visitorAnalyticsAPI.channelStats(filters)
  } catch { appStore.showError(t('admin.visitorAnalytics.operationFailed')) }
  finally { savingChannel.value = false }
}
async function deleteChannel() {
  if (!channelForm.id || !window.confirm(t('admin.visitorAnalytics.deleteConfirm'))) return
  try { await visitorAnalyticsAPI.deleteChannel(channelForm.id); appStore.showSuccess(t('admin.visitorAnalytics.channelDeleted')); resetChannelForm(); channels.value = await visitorAnalyticsAPI.channels(); channelStats.value = await visitorAnalyticsAPI.channelStats(filters) }
  catch { appStore.showError(t('admin.visitorAnalytics.operationFailed')) }
}

async function lookupIP() {
  if (!ipQuery.value.trim()) return
  lookingUpIP.value = true
  try { ipResult.value = await visitorAnalyticsAPI.lookupIP(ipQuery.value.trim()) }
  catch { appStore.showError(t('admin.visitorAnalytics.operationFailed')) }
  finally { lookingUpIP.value = false }
}
async function lookupEventIP(ip: string) { ipQuery.value = ip; await lookupIP(); document.querySelector('.ip-panel')?.scrollIntoView({ behavior: 'smooth', block: 'center' }); await reloadEvents() }

watch(() => [filters.start_date, filters.end_date], () => { events.page = 1; void refreshAll() })
onMounted(refreshAll)
</script>

<style scoped>
.visitor-analytics-page{--ink:#111827;--muted:#6b7280;--line:#e7eaf1;display:flex;flex-direction:column;gap:24px;padding:32px 28px 56px;color:var(--ink);background:radial-gradient(circle at 90% 0,rgba(118,92,255,.11),transparent 27%),#f7f8fc;min-height:100%}.analytics-hero,.panel,.metric-card{background:rgba(255,255,255,.88);border:1px solid rgba(224,227,237,.86);box-shadow:0 18px 45px rgba(78,87,119,.08)}.analytics-hero{position:relative;display:grid;grid-template-columns:minmax(0,1fr) auto;gap:26px;padding:38px;border-radius:28px;overflow:hidden}.analytics-hero:after{content:"";position:absolute;width:360px;height:360px;border-radius:50%;right:-170px;top:-230px;background:linear-gradient(145deg,#655cff,#9a5cff);opacity:.14}.hero-copy{max-width:780px;z-index:1}.eyebrow{display:block;color:#665cff;font-size:12px;font-weight:800;letter-spacing:.18em;margin-bottom:12px}.hero-copy h1{font-size:clamp(30px,3vw,48px);line-height:1.08;letter-spacing:-.04em;margin:0 0 15px}.hero-copy p,.panel-heading p,.ip-copy p{color:var(--muted);font-size:15px;line-height:1.7;margin:0}.hero-actions{display:flex;align-items:flex-start;gap:12px;z-index:1}.date-control{padding:10px 14px;border:1px solid var(--line);border-radius:16px;background:#fff}.date-control>span{display:block;color:#9ca3af;font-size:11px;font-weight:700;margin-bottom:5px}.date-inputs{display:flex;align-items:center;gap:7px}.date-inputs input{border:0;background:transparent;font-weight:650;color:var(--ink);outline:0}.date-inputs i{font-style:normal;color:#c3c7d2}.primary-button,.secondary-button,.danger-button{display:inline-flex;align-items:center;justify-content:center;gap:8px;border:0;border-radius:14px;padding:13px 18px;font-weight:750;cursor:pointer;transition:.2s}.primary-button{color:#fff;background:linear-gradient(135deg,#5f6aff,#8756f6);box-shadow:0 10px 24px rgba(99,82,245,.22)}.secondary-button{color:#30374a;background:#fff;border:1px solid var(--line)}.danger-button{color:#dc3545;background:#fff0f1}.primary-button:hover,.secondary-button:hover{transform:translateY(-1px)}button:disabled{opacity:.5;cursor:not-allowed;transform:none}.primary-button svg,.secondary-button svg,.copy-link svg,.search-control svg{width:18px;height:18px;fill:none;stroke:currentColor;stroke-width:1.8;stroke-linecap:round;stroke-linejoin:round}.privacy-note{grid-column:1/-1;display:flex;align-items:center;gap:9px;padding-top:19px;border-top:1px solid var(--line);color:#7b8190;font-size:12px}.privacy-note svg{width:17px;height:17px;fill:none;stroke:#6d64f8;stroke-width:1.8}.metric-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:16px}.metric-card{display:flex;align-items:center;gap:16px;padding:22px;border-radius:22px}.metric-icon{display:grid;place-items:center;width:48px;height:48px;border-radius:15px}.metric-icon :deep(svg){width:22px;height:22px;fill:none;stroke:currentColor;stroke-width:1.8;stroke-linecap:round;stroke-linejoin:round}.metric-icon.violet{background:#efedff;color:#6b5cff}.metric-icon.cyan{background:#e5f8ff;color:#0b9bc2}.metric-icon.green{background:#e7fbf1;color:#0aa56f}.metric-icon.amber{background:#fff5df;color:#e39920}.metric-card span{display:block;color:#8b91a0;font-size:12px;font-weight:700;margin-bottom:4px}.metric-card strong{font-size:29px;letter-spacing:-.04em}.analytics-grid{display:grid;grid-template-columns:minmax(0,1.8fr) minmax(280px,.7fr);gap:20px}.panel{border-radius:24px;padding:26px}.panel-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:20px;margin-bottom:24px}.panel-heading h2,.ip-copy h2{font-size:20px;margin:0 0 6px;letter-spacing:-.02em}.trend-legend{display:flex;align-items:center;gap:6px;color:#858b99;font-size:11px;white-space:nowrap}.bar-dot,.line-dot{width:8px;height:8px;border-radius:3px;background:#6a62f7}.line-dot{margin-left:7px;border-radius:50%;background:#13b985}.trend-chart{display:grid;grid-template-columns:repeat(auto-fit,minmax(32px,1fr));gap:10px;height:240px}.trend-column{display:flex;flex-direction:column;align-items:center;min-width:0}.bar-track{position:relative;display:flex;align-items:flex-end;justify-content:center;flex:1;width:100%;border-bottom:1px solid #eceef5;background:linear-gradient(to top,rgba(103,92,247,.035),transparent)}.view-bar{width:min(28px,65%);min-height:0;border-radius:7px 7px 2px 2px;background:linear-gradient(to top,#655af2,#8b7cff);transition:height .35s}.visitor-marker{position:absolute;z-index:2;width:8px;height:8px;border-radius:50%;border:2px solid #fff;background:#0dbb82;box-shadow:0 0 0 2px rgba(13,187,130,.16)}.trend-column>span{font-size:10px;color:#969caa;margin-top:10px}.chart-skeleton,.empty-chart{height:240px;border-radius:15px;background:linear-gradient(90deg,#f2f3f7 25%,#fafafd 50%,#f2f3f7 75%);background-size:200% 100%;animation:pulse 1.4s infinite}.empty-chart{display:grid;place-items:center;background:#fafafd;color:#9ca3af;animation:none}.settings-panel{display:flex;flex-direction:column}.setting-row{display:flex;align-items:center;justify-content:space-between;padding:18px 0;border-top:1px solid var(--line);border-bottom:1px solid var(--line)}.setting-row strong,.setting-row span{display:block}.setting-row span{color:#8b91a0;font-size:12px;margin-top:4px}.switch-input{appearance:none;width:44px;height:25px;border-radius:999px;background:#d7dae3;position:relative;cursor:pointer;transition:.2s}.switch-input:after{content:"";position:absolute;width:19px;height:19px;border-radius:50%;background:#fff;left:3px;top:3px;box-shadow:0 2px 6px #0002;transition:.2s}.switch-input:checked{background:#6a5cf6}.switch-input:checked:after{transform:translateX(19px)}.retention-field{display:flex;flex-direction:column;gap:8px;margin:18px 0}.retention-field span,.channel-form label>span{font-size:12px;color:#72798a;font-weight:700}.retention-field select,.record-filters select,.channel-form input,.channel-form textarea,.ip-search input{border:1px solid var(--line);border-radius:13px;padding:12px 14px;background:#fff;color:var(--ink);outline:none}.full{width:100%;margin-top:auto}.channel-heading{align-items:center}.channel-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px}.channel-card{border:1px solid var(--line);border-radius:19px;padding:18px;background:#fff;min-width:0}.channel-topline{display:flex;align-items:center;gap:11px}.channel-avatar{display:grid;place-items:center;width:38px;height:38px;border-radius:12px;background:#f0efff;color:#675cf6;font-weight:800}.channel-topline>div{min-width:0;flex:1}.channel-topline strong,.channel-topline code{display:block;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.channel-topline code{font-size:11px;color:#969baa;margin-top:3px}.status-pill,.source-pill{font-size:10px;border-radius:999px;padding:5px 8px;font-style:normal}.status-pill.active{background:#e8fbf2;color:#079968}.status-pill.inactive{background:#f1f2f5;color:#8b91a0}.channel-metrics{display:grid;grid-template-columns:repeat(3,1fr);gap:8px;margin:18px 0 14px}.channel-metrics div{padding:10px;border-radius:12px;background:#f8f8fb}.channel-metrics span,.channel-metrics strong{display:block}.channel-metrics span{font-size:9px;color:#959aa8;white-space:nowrap}.channel-metrics strong{font-size:18px;margin-top:3px}.copy-link{display:flex;align-items:center;gap:8px;width:100%;border:0;border-radius:10px;background:#f6f6fa;color:#636a79;padding:10px;cursor:pointer}.copy-link span{flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;text-align:left;font-size:10px}.copy-link svg{flex:0 0 auto;width:15px}.empty-inline,.table-empty{padding:48px;text-align:center;color:#969baa}.records-heading{align-items:center}.record-filters{display:flex;gap:10px}.search-control{display:flex;align-items:center;gap:8px;border:1px solid var(--line);border-radius:13px;padding:0 12px;background:#fff}.search-control svg{color:#9ba1af}.search-control input{border:0;outline:0;padding:12px 0;min-width:210px}.record-filters select{min-width:150px}.table-shell{overflow-x:auto;border:1px solid var(--line);border-radius:17px}.table-shell table{width:100%;min-width:1120px;border-collapse:collapse}.table-shell th{padding:13px 14px;background:#f7f8fb;color:#89909e;font-size:10px;text-align:left;letter-spacing:.04em;text-transform:uppercase}.table-shell td{padding:15px 14px;border-top:1px solid #eceef4;font-size:12px;vertical-align:middle}.time-cell{white-space:nowrap;color:#737989}.source-pill{background:#efedff;color:#665bf4;font-weight:700}.ip-button,.lookup-button{border:0;background:transparent;color:#5f5be8;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;cursor:pointer}.lookup-button{font-family:inherit;border:1px solid #dad8ff;border-radius:9px;padding:6px 9px}.location-cell strong,.location-cell span{display:block;white-space:nowrap}.location-cell span{color:#9298a7;font-size:10px;margin-top:3px}.path-cell,.referrer-cell{display:block;max-width:170px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.path-cell{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;color:#333a49}.referrer-cell{color:#737989}.device-cell{white-space:nowrap;color:#828897}.device-cell em{margin-left:5px;font-style:normal;color:#d37a17}.pagination-row{display:flex;align-items:center;justify-content:space-between;padding-top:16px;color:#858b99;font-size:12px}.pagination-row div{display:flex;gap:8px}.pagination-row button{border:1px solid var(--line);border-radius:10px;background:#fff;padding:8px 12px;color:#4c5362}.ip-panel{display:grid;grid-template-columns:1fr 1fr;gap:30px;align-items:center;background:linear-gradient(135deg,#151827,#25263b);color:#fff;border:0;overflow:hidden}.ip-copy p{color:#aeb3c4}.ip-search{display:flex;gap:10px;margin-top:22px}.ip-search input{flex:1;background:#ffffff10;border-color:#ffffff1c;color:#fff}.ip-result{border:1px solid #ffffff18;border-radius:18px;background:#ffffff0a;padding:20px}.result-head{display:flex;justify-content:space-between;gap:15px;padding-bottom:15px;border-bottom:1px solid #ffffff14}.result-head span{font-family:monospace;color:#aaa6ff}.ip-result dl{display:grid;grid-template-columns:repeat(2,1fr);gap:16px;margin:18px 0 0}.ip-result dt{color:#999fb0;font-size:10px;text-transform:uppercase}.ip-result dd{margin:5px 0 0;font-size:13px}.ip-result small{color:#777e90}.ip-placeholder-art{position:relative;display:grid;place-items:center;min-height:190px}.ip-placeholder-art span{display:grid;place-items:center;width:105px;height:105px;border:1px solid #887cff;border-radius:50%;font-size:26px;font-weight:800;color:#a9a2ff;box-shadow:0 0 60px #735fff42}.ip-placeholder-art i{position:absolute;width:160px;height:160px;border:1px solid #ffffff10;border-radius:50%}.ip-placeholder-art i:nth-child(2){width:220px;height:220px}.ip-placeholder-art i:nth-child(3){width:280px;height:280px}.modal-backdrop{position:fixed;inset:0;z-index:100;background:rgba(22,25,39,.48);backdrop-filter:blur(10px);display:grid;place-items:center;padding:20px}.channel-modal{width:min(980px,95vw);max-height:88vh;overflow:auto;border-radius:26px;background:#f9f9fc;box-shadow:0 30px 90px #1e22355c}.modal-heading{display:flex;align-items:center;justify-content:space-between;padding:25px 27px;border-bottom:1px solid var(--line)}.modal-heading h2{margin:0}.icon-button{border:0;background:#eceef4;width:38px;height:38px;border-radius:12px;font-size:24px;cursor:pointer}.modal-body-grid{display:grid;grid-template-columns:320px 1fr;min-height:500px}.configured-channels{padding:22px;border-right:1px solid var(--line);display:flex;flex-direction:column;gap:9px}.add-channel-card,.configured-channel{border:1px dashed #c9c6ff;border-radius:14px;padding:14px;background:#fff;color:#655bf2;text-align:left;cursor:pointer}.configured-channel{display:flex;align-items:center;gap:10px;border-style:solid;border-color:var(--line);color:var(--ink)}.configured-channel.selected{border-color:#8175ff;background:#f4f2ff}.configured-channel>span{display:grid;place-items:center;width:34px;height:34px;border-radius:10px;background:#efedff;color:#655bf2;font-weight:800}.configured-channel>div{flex:1;min-width:0}.configured-channel strong,.configured-channel code{display:block;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.configured-channel code{color:#9298a7;font-size:10px;margin-top:3px}.configured-channel em{font-style:normal;font-size:9px;color:#8a90a0}.channel-form{padding:28px;display:flex;flex-direction:column;gap:16px}.channel-form h3{margin:0 0 4px}.channel-form label{display:flex;flex-direction:column;gap:7px}.channel-form small{color:#9ba0ae}.channel-form textarea{resize:vertical}.channel-form input:focus,.channel-form textarea:focus,.ip-search input:focus{border-color:#776af8;box-shadow:0 0 0 3px rgba(119,106,248,.12)}.form-toggle{flex-direction:row!important;align-items:center;justify-content:space-between}.form-actions{display:grid;grid-template-columns:auto 1fr auto auto;gap:10px;margin-top:auto}.modal-fade-enter-active,.modal-fade-leave-active{transition:.2s}.modal-fade-enter-from,.modal-fade-leave-to{opacity:0}.modal-fade-enter-from .channel-modal,.modal-fade-leave-to .channel-modal{transform:translateY(12px) scale(.98)}@keyframes pulse{to{background-position:-200% 0}}@media(max-width:1100px){.metric-grid{grid-template-columns:repeat(2,1fr)}.analytics-grid{grid-template-columns:1fr}.channel-grid{grid-template-columns:repeat(2,1fr)}.hero-actions{flex-direction:column}}@media(max-width:760px){.visitor-analytics-page{padding:18px 14px 40px}.analytics-hero{grid-template-columns:1fr;padding:25px}.hero-actions{width:100%}.date-control,.primary-button{width:100%}.date-inputs{justify-content:space-between}.metric-grid,.channel-grid{grid-template-columns:1fr}.panel{padding:20px}.panel-heading,.records-heading,.channel-heading{flex-direction:column}.record-filters{width:100%;flex-direction:column}.search-control input{min-width:0;width:100%}.ip-panel{grid-template-columns:1fr}.modal-body-grid{grid-template-columns:1fr}.configured-channels{border-right:0;border-bottom:1px solid var(--line);max-height:240px;overflow:auto}.form-actions{grid-template-columns:1fr 1fr}.form-actions span{display:none}.trend-legend{display:none}}:global(.dark) .visitor-analytics-page{--ink:#f4f5fb;--muted:#949bad;--line:#292d3b;background:radial-gradient(circle at 90% 0,rgba(112,88,255,.16),transparent 28%),#0c0f18}:global(.dark) .analytics-hero,:global(.dark) .panel,:global(.dark) .metric-card{background:rgba(17,20,31,.9);border-color:#272b39}:global(.dark) .date-control,:global(.dark) .retention-field select,:global(.dark) .record-filters select,:global(.dark) .search-control,:global(.dark) .channel-card,:global(.dark) .secondary-button,:global(.dark) .pagination-row button{background:#151925;border-color:#2d3140;color:var(--ink)}:global(.dark) .metric-icon.violet,:global(.dark) .channel-avatar{background:#292443}:global(.dark) .channel-metrics div,:global(.dark) .copy-link,:global(.dark) .empty-chart{background:#151925}:global(.dark) .table-shell th{background:#151925}:global(.dark) .table-shell td{border-color:#272b39}:global(.dark) .channel-modal{background:#11141f}:global(.dark) .configured-channel,:global(.dark) .add-channel-card,:global(.dark) .channel-form input,:global(.dark) .channel-form textarea{background:#171b28;border-color:#2b3040;color:var(--ink)}
/* Keep all channel form actions visible on ordinary laptop viewports. */
.channel-modal{height:min(720px,88vh);display:flex;flex-direction:column;overflow:hidden}
.modal-heading{padding:20px 27px;flex:0 0 auto}
.modal-body-grid{flex:1;min-height:0}
.configured-channels,.channel-form{overflow:auto}
.channel-form{padding:24px 28px;gap:13px}

/* Keep this operational page compact and aligned with the rest of the admin UI. */
.visitor-analytics-page {
  background: transparent;
  gap: 18px;
  padding: 0 0 32px;
}
.analytics-hero,
.panel,
.metric-card {
  border-color: #e5e7eb;
  border-radius: 8px;
  box-shadow: none;
}
.analytics-hero {
  padding: 22px;
}
.analytics-hero::after {
  display: none;
}
.hero-copy h1 {
  font-size: 24px;
  letter-spacing: 0;
  line-height: 1.3;
  margin-bottom: 8px;
}
.hero-copy p,
.panel-heading p,
.ip-copy p {
  font-size: 14px;
  line-height: 1.6;
}
.eyebrow {
  letter-spacing: 0;
  margin-bottom: 6px;
}
.date-control,
.primary-button,
.secondary-button,
.danger-button,
.metric-card,
.metric-icon,
.panel,
.channel-card,
.channel-avatar,
.channel-metrics div,
.copy-link,
.search-control,
.table-shell,
.lookup-button,
.ip-result,
.channel-modal,
.icon-button,
.add-channel-card,
.configured-channel,
.configured-channel > span,
.retention-field select,
.record-filters select,
.channel-form input,
.channel-form textarea,
.ip-search input {
  border-radius: 8px;
}
.primary-button {
  background: #2563eb;
  box-shadow: none;
}
.primary-button:hover {
  background: #1d4ed8;
}
.primary-button:hover,
.secondary-button:hover {
  transform: none;
}
.metric-card {
  padding: 16px;
}
.metric-card strong {
  font-size: 24px;
  letter-spacing: 0;
}
.panel {
  padding: 20px;
}
.view-bar {
  background: #4f46e5;
}
.bar-track,
.chart-skeleton {
  background: #f9fafb;
}
.ip-panel {
  background: #111827;
}
.ip-placeholder-art {
  color: #a5b4fc;
}
.ip-placeholder-art span,
.ip-placeholder-art i {
  display: none;
}
:global(.dark) .visitor-analytics-page {
  background: transparent;
}
:global(.dark) .analytics-hero,
:global(.dark) .panel,
:global(.dark) .metric-card {
  border-color: #272b39;
  box-shadow: none;
}
</style>
