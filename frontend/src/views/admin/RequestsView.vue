<template>
  <AppLayout>
    <div class="space-y-5">
      <section class="border-b border-gray-200 pb-4 dark:border-dark-700">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ ui.title }}</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ ui.subtitle }}</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <button type="button" class="btn btn-secondary" :disabled="loading" @click="refreshAll">
              <Icon name="refresh" size="sm" class="mr-1.5" />
              {{ ui.refresh }}
            </button>
            <button type="button" class="btn btn-secondary" @click="resetFilters">
              <Icon name="x" size="sm" class="mr-1.5" />
              {{ ui.reset }}
            </button>
          </div>
        </div>
      </section>

      <section class="rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
        <div class="grid gap-3 p-4 md:grid-cols-2 xl:grid-cols-4">
          <label class="space-y-1">
            <span class="input-label">{{ ui.range }}</span>
            <select v-model="rangePreset" class="input">
              <option value="5m">{{ ui.last5m }}</option>
              <option value="30m">{{ ui.last30m }}</option>
              <option value="1h">{{ ui.last1h }}</option>
              <option value="6h">{{ ui.last6h }}</option>
              <option value="24h">{{ ui.last24h }}</option>
              <option value="7d">{{ ui.last7d }}</option>
              <option value="30d">{{ ui.last30dRange }}</option>
              <option value="custom">{{ ui.custom }}</option>
            </select>
          </label>

          <label v-if="rangePreset === 'custom'" class="space-y-1">
            <span class="input-label">{{ ui.startTime }}</span>
            <input v-model="customStart" type="datetime-local" class="input" />
          </label>

          <label v-if="rangePreset === 'custom'" class="space-y-1">
            <span class="input-label">{{ ui.endTime }}</span>
            <input v-model="customEnd" type="datetime-local" class="input" />
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.result }}</span>
            <select v-model="filters.kind" class="input">
              <option value="all">{{ ui.all }}</option>
              <option value="success">{{ ui.success }}</option>
              <option value="error">{{ ui.failed }}</option>
            </select>
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.sort }}</span>
            <select v-model="filters.sort" class="input">
              <option value="created_at_desc">{{ ui.sortCreated }}</option>
              <option value="duration_desc">{{ ui.sortDuration }}</option>
            </select>
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.search }}</span>
            <div class="relative">
              <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input
                v-model.trim="filters.q"
                type="search"
                class="input pl-9"
                :placeholder="ui.searchPlaceholder"
                @keyup.enter="applyFilters"
              />
            </div>
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.requestId }}</span>
            <input v-model.trim="filters.request_id" type="text" class="input font-mono text-xs" @keyup.enter="applyFilters" />
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.model }}</span>
            <select v-model="filters.model" class="input">
              <option value="">{{ ui.all }}</option>
              <option v-for="option in filterOptions.models" :key="`model:${option.value}`" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.platform }}</span>
            <select v-model="filters.platform" class="input">
              <option value="">{{ ui.all }}</option>
              <option v-for="option in filterOptions.platforms" :key="`platform:${option.value}`" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.userId }}</span>
            <select v-model="filters.user_id" class="input">
              <option value="">{{ ui.all }}</option>
              <option v-for="option in filterOptions.users" :key="`user:${option.value}`" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.apiKeyId }}</span>
            <select v-model="filters.api_key_id" class="input">
              <option value="">{{ ui.all }}</option>
              <option v-for="option in filterOptions.api_keys" :key="`api-key:${option.value}`" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.accountId }}</span>
            <select v-model="filters.account_id" class="input">
              <option value="">{{ ui.all }}</option>
              <option v-for="option in filterOptions.accounts" :key="`account:${option.value}`" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.groupId }}</span>
            <select v-model="filters.group_id" class="input">
              <option value="">{{ ui.all }}</option>
              <option v-for="option in filterOptions.groups" :key="`group:${option.value}`" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.minDuration }}</span>
            <input v-model.trim="filters.min_duration_ms" type="number" min="0" class="input" @keyup.enter="applyFilters" />
          </label>

          <label class="space-y-1">
            <span class="input-label">{{ ui.maxDuration }}</span>
            <input v-model.trim="filters.max_duration_ms" type="number" min="0" class="input" @keyup.enter="applyFilters" />
          </label>

          <div class="flex items-end">
            <button type="button" class="btn btn-primary w-full" :disabled="loading || filterOptionsLoading" @click="applyFilters">
              <Icon name="filter" size="sm" class="mr-1.5" />
              {{ ui.apply }}
            </button>
          </div>
        </div>
      </section>

      <section class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col gap-3 border-b border-gray-200 px-4 py-3 dark:border-dark-700 md:flex-row md:items-center md:justify-between">
          <div class="text-sm text-gray-600 dark:text-gray-300">
            {{ ui.total }} <span class="font-semibold text-gray-900 dark:text-white">{{ formatNumber(total) }}</span>
          </div>
          <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
            <span>{{ ui.pageSize }}</span>
            <select v-model.number="pageSize" class="input h-9 w-24">
              <option :value="10">10</option>
              <option :value="20">20</option>
              <option :value="50">50</option>
              <option :value="100">100</option>
            </select>
          </label>
        </div>

        <div v-if="loading" class="flex min-h-[280px] items-center justify-center">
          <div class="flex items-center gap-3 text-sm text-gray-500 dark:text-gray-400">
            <Icon name="refresh" size="md" class="animate-spin" />
            {{ ui.loading }}
          </div>
        </div>

        <div v-else-if="items.length === 0" class="px-4 py-16 text-center">
          <div class="text-sm font-medium text-gray-600 dark:text-gray-300">{{ ui.empty }}</div>
          <div class="mt-1 text-xs text-gray-400">{{ ui.emptyHint }}</div>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-900">
              <tr>
                <th class="table-head">{{ ui.time }}</th>
                <th class="table-head">{{ ui.result }}</th>
                <th class="table-head">{{ ui.identity }}</th>
                <th class="table-head">{{ ui.route }}</th>
                <th class="table-head">{{ ui.model }}</th>
                <th class="table-head">{{ ui.endpoint }}</th>
                <th class="table-head">{{ ui.outcome }}</th>
                <th class="table-head min-w-[300px]">{{ ui.params }}</th>
                <th class="table-head text-right">{{ ui.actions }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 dark:divide-dark-700">
              <template v-for="(row, index) in items" :key="rowKey(row, index)">
                <tr class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                  <td class="table-cell whitespace-nowrap text-xs text-gray-600 dark:text-gray-300">
                    {{ formatDateTime(row.created_at) || '-' }}
                  </td>
                  <td class="table-cell">
                    <span class="inline-flex rounded-full px-2 py-1 text-[11px] font-semibold" :class="kindBadgeClass(row.kind)">
                      {{ row.kind === 'error' ? ui.failed : ui.success }}
                    </span>
                  </td>
                  <td class="table-cell">
                    <div class="space-y-0.5 text-xs">
                      <div class="text-gray-700 dark:text-gray-200">{{ ui.userId }}: {{ row.user_id ?? '-' }}</div>
                      <div class="text-gray-500 dark:text-gray-400">{{ ui.apiKeyId }}: {{ row.api_key_id ?? '-' }}</div>
                      <div class="text-gray-500 dark:text-gray-400">{{ ui.accountId }}: {{ row.account_id ?? '-' }}</div>
                    </div>
                  </td>
                  <td class="table-cell">
                    <div class="space-y-1 text-xs">
                      <div class="font-semibold text-gray-800 dark:text-gray-100">{{ formatPlatform(row.platform) }}</div>
                      <div class="flex flex-wrap gap-1">
                        <span v-if="row.request_type" class="chip">{{ row.request_type }}</span>
                        <span class="chip">{{ row.stream ? 'stream' : 'sync' }}</span>
                      </div>
                    </div>
                  </td>
                  <td class="table-cell">
                    <div class="max-w-[260px] space-y-0.5 text-xs">
                      <div class="truncate font-medium text-gray-800 dark:text-gray-100" :title="row.model || ''">{{ row.model || '-' }}</div>
                      <div v-if="row.requested_model && row.requested_model !== row.model" class="truncate text-gray-500 dark:text-gray-400" :title="row.requested_model">
                        {{ ui.requested }}: {{ row.requested_model }}
                      </div>
                      <div v-if="row.upstream_model" class="truncate text-gray-500 dark:text-gray-400" :title="row.upstream_model">
                        {{ ui.upstream }}: {{ row.upstream_model }}
                      </div>
                    </div>
                  </td>
                  <td class="table-cell">
                    <div class="max-w-[280px] space-y-0.5 text-xs">
                      <div class="truncate text-gray-700 dark:text-gray-200" :title="row.inbound_endpoint || ''">
                        {{ ui.inbound }}: {{ row.inbound_endpoint || '-' }}
                      </div>
                      <div class="truncate text-gray-500 dark:text-gray-400" :title="row.upstream_endpoint || ''">
                        {{ ui.upstream }}: {{ row.upstream_endpoint || '-' }}
                      </div>
                    </div>
                  </td>
                  <td class="table-cell">
                    <div class="space-y-0.5 text-xs">
                      <div>{{ formatDuration(row.duration_ms) }}</div>
                      <div v-if="row.kind === 'error'" class="text-red-600 dark:text-red-300">
                        {{ row.status_code ?? '-' }} {{ row.severity ? `/ ${row.severity}` : '' }}
                      </div>
                      <div v-else class="text-gray-500 dark:text-gray-400">200</div>
                    </div>
                  </td>
                  <td class="table-cell">
                    <div v-if="hasParams(row.request_params)" class="flex max-w-[360px] flex-wrap gap-1.5">
                      <span v-for="param in summaryParams(row.request_params)" :key="param.key" class="param-chip" :title="`${param.key}: ${param.value}`">
                        <span class="text-gray-500 dark:text-gray-400">{{ param.key }}</span>
                        <span class="ml-1 font-medium text-gray-800 dark:text-gray-100">{{ param.value }}</span>
                      </span>
                      <span v-if="hiddenParamCount(row.request_params) > 0" class="param-chip text-gray-500 dark:text-gray-400">
                        +{{ hiddenParamCount(row.request_params) }}
                      </span>
                    </div>
                    <span v-else class="text-xs text-gray-400">-</span>
                  </td>
                  <td class="table-cell text-right">
                    <div class="flex justify-end gap-1.5">
                      <button
                        v-if="row.request_id"
                        type="button"
                        class="icon-button"
                        :title="ui.copyRequestId"
                        @click="copyText(row.request_id, ui.copied)"
                      >
                        <Icon name="copy" size="sm" />
                      </button>
                      <button
                        type="button"
                        class="icon-button"
                        :title="expandedKeys.has(rowKey(row, index)) ? ui.collapse : ui.expand"
                        @click="toggleExpanded(rowKey(row, index))"
                      >
                        <Icon name="chevronDown" size="sm" class="transition-transform" :class="{ 'rotate-180': expandedKeys.has(rowKey(row, index)) }" />
                      </button>
                    </div>
                  </td>
                </tr>
                <tr v-if="expandedKeys.has(rowKey(row, index))" class="bg-gray-50 dark:bg-dark-900/60">
                  <td colspan="9" class="px-4 py-4">
                    <div class="grid gap-4 lg:grid-cols-[minmax(260px,360px),1fr]">
                      <div class="space-y-2 text-xs text-gray-600 dark:text-gray-300">
                        <div v-for="item in detailRows(row)" :key="item.label" class="flex gap-2">
                          <span class="w-28 shrink-0 text-gray-400">{{ item.label }}</span>
                          <span class="min-w-0 break-all font-mono">{{ item.value }}</span>
                        </div>
                        <div v-if="row.kind === 'error' && row.message" class="pt-2">
                          <div class="mb-1 text-gray-400">{{ ui.errorMessage }}</div>
                          <div class="break-words rounded-md border border-red-200 bg-red-50 p-2 text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-200">
                            {{ row.message }}
                          </div>
                        </div>
                      </div>

                      <div>
                        <div class="mb-2 flex items-center justify-between">
                          <div class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ ui.paramsJson }}</div>
                          <button
                            type="button"
                            class="btn btn-secondary btn-sm"
                            :disabled="!hasParams(row.request_params)"
                            @click="copyText(formatParamsJson(row.request_params), ui.copied)"
                          >
                            <Icon name="copy" size="xs" class="mr-1.5" />
                            {{ ui.copy }}
                          </button>
                        </div>
                        <pre class="max-h-96 overflow-auto rounded-md border border-gray-200 bg-white p-3 text-xs leading-relaxed text-gray-800 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-100">{{ formatParamsJson(row.request_params) }}</pre>
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <Pagination
          v-if="total > 0"
          :page="page"
          :total="total"
          :page-size="pageSize"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  opsAPI,
  type OpsRequestDetail,
  type OpsRequestDetailsKind,
  type OpsRequestDetailsParams,
  type OpsRequestDetailsSort,
  type OpsRequestFilterOptions
} from '@/api/admin/ops'
import { useAppStore } from '@/stores'
import { formatDateTime, formatNumber } from '@/utils/format'

type RangePreset = NonNullable<OpsRequestDetailsParams['time_range']> | 'custom'

interface FilterState {
  kind: OpsRequestDetailsKind
  sort: OpsRequestDetailsSort
  q: string
  request_id: string
  model: string
  platform: string
  user_id: string
  api_key_id: string
  account_id: string
  group_id: string
  min_duration_ms: string
  max_duration_ms: string
}

const zh = {
  title: '调用请求',
  subtitle: '查看成功和失败的用户调用请求，以及脱敏后的入参摘要，包含图片生成尺寸等关键参数。',
  refresh: '刷新',
  reset: '重置',
  range: '时间范围',
  last5m: '最近 5 分钟',
  last30m: '最近 30 分钟',
  last1h: '最近 1 小时',
  last6h: '最近 6 小时',
  last24h: '最近 24 小时',
  last7d: '最近 7 天',
  last30dRange: '最近 30 天',
  custom: '自定义',
  startTime: '开始时间',
  endTime: '结束时间',
  result: '结果',
  all: '全部',
  success: '成功',
  failed: '失败',
  sort: '排序',
  sortCreated: '最新请求',
  sortDuration: '耗时最长',
  search: '搜索',
  searchPlaceholder: '模型、请求 ID、错误或参数',
  requestId: '请求 ID',
  model: '模型',
  platform: '平台',
  userId: '用户',
  apiKeyId: '密钥',
  accountId: '账号',
  groupId: '分组',
  minDuration: '最小耗时 ms',
  maxDuration: '最大耗时 ms',
  apply: '应用筛选',
  total: '总数',
  pageSize: '每页',
  loading: '加载中',
  empty: '没有匹配的请求',
  emptyHint: '调整时间范围或筛选条件后再试。',
  time: '时间',
  identity: '身份',
  route: '路由',
  requested: '请求',
  upstream: '上游',
  inbound: '入站',
  endpoint: '端点',
  outcome: '耗时 / 状态',
  params: '入参',
  actions: '操作',
  copyRequestId: '复制请求 ID',
  copy: '复制',
  copied: '已复制',
  copyFailed: '复制失败',
  expand: '展开',
  collapse: '收起',
  errorMessage: '错误信息',
  paramsJson: '完整入参摘要'
}

const en: typeof zh = {
  title: 'Requests',
  subtitle: 'Inspect successful and failed user calls with sanitized request parameters, including image generation sizes.',
  refresh: 'Refresh',
  reset: 'Reset',
  range: 'Time range',
  last5m: 'Last 5 min',
  last30m: 'Last 30 min',
  last1h: 'Last 1 hour',
  last6h: 'Last 6 hours',
  last24h: 'Last 24 hours',
  last7d: 'Last 7 days',
  last30dRange: 'Last 30 days',
  custom: 'Custom',
  startTime: 'Start time',
  endTime: 'End time',
  result: 'Result',
  all: 'All',
  success: 'Success',
  failed: 'Failed',
  sort: 'Sort',
  sortCreated: 'Newest',
  sortDuration: 'Slowest',
  search: 'Search',
  searchPlaceholder: 'Model, request ID, error, or params',
  requestId: 'Request ID',
  model: 'Model',
  platform: 'Platform',
  userId: 'User',
  apiKeyId: 'API key',
  accountId: 'Account',
  groupId: 'Group',
  minDuration: 'Min duration ms',
  maxDuration: 'Max duration ms',
  apply: 'Apply filters',
  total: 'Total',
  pageSize: 'Page size',
  loading: 'Loading',
  empty: 'No matching requests',
  emptyHint: 'Try a wider time range or fewer filters.',
  time: 'Time',
  identity: 'Identity',
  route: 'Route',
  requested: 'Requested',
  upstream: 'Upstream',
  inbound: 'Inbound',
  endpoint: 'Endpoint',
  outcome: 'Duration / Status',
  params: 'Params',
  actions: 'Actions',
  copyRequestId: 'Copy request ID',
  copy: 'Copy',
  copied: 'Copied',
  copyFailed: 'Copy failed',
  expand: 'Expand',
  collapse: 'Collapse',
  errorMessage: 'Error message',
  paramsJson: 'Full params JSON'
}

const { locale } = useI18n()
const appStore = useAppStore()

const ui = computed(() => String(locale.value).startsWith('zh') ? zh : en)
const loading = ref(false)
const filterOptionsLoading = ref(false)
const items = ref<OpsRequestDetail[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const rangePreset = ref<RangePreset>('1h')
const customEnd = ref(toLocalInputValue(new Date()))
const customStart = ref(toLocalInputValue(new Date(Date.now() - 60 * 60 * 1000)))
const expandedKeys = ref(new Set<string>())
const filterOptions = ref<OpsRequestFilterOptions>(emptyFilterOptions())

const filters = reactive<FilterState>({
  kind: 'all',
  sort: 'created_at_desc',
  q: '',
  request_id: '',
  model: '',
  platform: '',
  user_id: '',
  api_key_id: '',
  account_id: '',
  group_id: '',
  min_duration_ms: '',
  max_duration_ms: ''
})

const priorityParamKeys = [
  'model',
  'request_type',
  'stream',
  'image_size',
  'image_input_size',
  'image_output_size',
  'size',
  'quality',
  'n',
  'service_tier',
  'reasoning_effort',
  'messages_count',
  'input_items',
  'contents_count',
  'prompt_chars',
  'tool_types'
]

onMounted(() => {
  void refreshAll()
})

watch(pageSize, () => {
  page.value = 1
  void fetchRequests()
})

watch([rangePreset, () => filters.kind], () => {
  page.value = 1
  expandedKeys.value = new Set()
  void refreshAll()
})

function buildWindowParams(): Pick<OpsRequestDetailsParams, 'time_range' | 'start_time' | 'end_time' | 'kind'> {
  const params: Pick<OpsRequestDetailsParams, 'time_range' | 'start_time' | 'end_time' | 'kind'> = {
    kind: filters.kind
  }

  if (rangePreset.value === 'custom') {
    if (customStart.value) params.start_time = new Date(customStart.value).toISOString()
    if (customEnd.value) params.end_time = new Date(customEnd.value).toISOString()
  } else {
    params.time_range = rangePreset.value
  }

  return params
}

function buildParams(): OpsRequestDetailsParams {
  const params: OpsRequestDetailsParams = {
    ...buildWindowParams(),
    sort: filters.sort,
    page: page.value,
    page_size: pageSize.value
  }

  if (filters.platform) params.platform = filters.platform
  if (filters.model) params.model = filters.model
  if (filters.request_id) params.request_id = filters.request_id
  if (filters.q) params.q = filters.q

  assignPositiveNumber(params, 'user_id', filters.user_id)
  assignPositiveNumber(params, 'api_key_id', filters.api_key_id)
  assignPositiveNumber(params, 'account_id', filters.account_id)
  assignPositiveNumber(params, 'group_id', filters.group_id)
  assignNonNegativeNumber(params, 'min_duration_ms', filters.min_duration_ms)
  assignNonNegativeNumber(params, 'max_duration_ms', filters.max_duration_ms)

  return params
}

async function refreshAll() {
  await Promise.all([fetchFilterOptions(), fetchRequests()])
}

async function fetchFilterOptions() {
  filterOptionsLoading.value = true
  try {
    const res = await opsAPI.getRequestFilterOptions(buildWindowParams())
    filterOptions.value = normalizeFilterOptions(res)
    pruneMissingOptionSelections()
  } catch (error) {
    console.error('[RequestsView] failed to load request filter options', error)
    filterOptions.value = emptyFilterOptions()
  } finally {
    filterOptionsLoading.value = false
  }
}

async function fetchRequests() {
  loading.value = true
  try {
    const res = await opsAPI.listRequestDetails(buildParams())
    items.value = res.items || []
    total.value = res.total || 0
    page.value = res.page || page.value
    pageSize.value = res.page_size || pageSize.value
  } catch (error: any) {
    console.error('[RequestsView] failed to load request details', error)
    appStore.showError(error?.response?.data?.error || error?.message || 'Failed to load requests')
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  page.value = 1
  expandedKeys.value = new Set()
  void fetchRequests()
}

function resetFilters() {
  rangePreset.value = '1h'
  customEnd.value = toLocalInputValue(new Date())
  customStart.value = toLocalInputValue(new Date(Date.now() - 60 * 60 * 1000))
  Object.assign(filters, {
    kind: 'all',
    sort: 'created_at_desc',
    q: '',
    request_id: '',
    model: '',
    platform: '',
    user_id: '',
    api_key_id: '',
    account_id: '',
    group_id: '',
    min_duration_ms: '',
    max_duration_ms: ''
  })
  page.value = 1
  expandedKeys.value = new Set()
  void refreshAll()
}

function handlePageChange(nextPage: number) {
  page.value = nextPage
  expandedKeys.value = new Set()
  void fetchRequests()
}

function handlePageSizeChange(nextSize: number) {
  pageSize.value = nextSize
}

function rowKey(row: OpsRequestDetail, index: number): string {
  return `${row.kind}:${row.request_id || row.error_id || row.created_at}:${index}`
}

function toggleExpanded(key: string) {
  const next = new Set(expandedKeys.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expandedKeys.value = next
}

function hasParams(params: OpsRequestDetail['request_params']): params is Record<string, unknown> {
  return !!params && typeof params === 'object' && Object.keys(params).length > 0
}

function summaryParams(params: OpsRequestDetail['request_params']): Array<{ key: string; value: string }> {
  if (!hasParams(params)) return []
  const result: Array<{ key: string; value: string }> = []
  const used = new Set<string>()

  for (const key of priorityParamKeys) {
    if (!(key in params)) continue
    result.push({ key, value: formatParamValue(params[key]) })
    used.add(key)
    if (result.length >= 7) return result
  }

  for (const [key, value] of Object.entries(params)) {
    if (used.has(key)) continue
    result.push({ key, value: formatParamValue(value) })
    if (result.length >= 7) break
  }
  return result
}

function hiddenParamCount(params: OpsRequestDetail['request_params']): number {
  if (!hasParams(params)) return 0
  return Math.max(0, Object.keys(params).length - summaryParams(params).length)
}

function formatParamsJson(params: OpsRequestDetail['request_params']): string {
  if (!hasParams(params)) return '{}'
  return JSON.stringify(params, null, 2)
}

function formatParamValue(value: unknown): string {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'string') return value.length > 48 ? `${value.slice(0, 48)}...` : value
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  if (Array.isArray(value)) {
    const joined = value.map((item) => typeof item === 'object' ? JSON.stringify(item) : String(item)).join(', ')
    return joined.length > 48 ? `${joined.slice(0, 48)}...` : joined
  }
  const json = JSON.stringify(value)
  return json.length > 48 ? `${json.slice(0, 48)}...` : json
}

function detailRows(row: OpsRequestDetail): Array<{ label: string; value: string }> {
  return [
    { label: ui.value.requestId, value: row.request_id || '-' },
    { label: ui.value.userId, value: String(row.user_id ?? '-') },
    { label: ui.value.apiKeyId, value: String(row.api_key_id ?? '-') },
    { label: ui.value.accountId, value: String(row.account_id ?? '-') },
    { label: ui.value.groupId, value: String(row.group_id ?? '-') },
    { label: ui.value.platform, value: formatPlatform(row.platform) },
    { label: ui.value.result, value: row.kind === 'error' ? ui.value.failed : ui.value.success },
    { label: ui.value.model, value: row.model || '-' },
    { label: ui.value.requested, value: row.requested_model || '-' },
    { label: ui.value.upstream, value: row.upstream_model || '-' },
    { label: ui.value.inbound, value: row.inbound_endpoint || '-' },
    { label: ui.value.endpoint, value: row.upstream_endpoint || '-' },
    { label: ui.value.outcome, value: `${formatDuration(row.duration_ms)} / ${row.status_code ?? (row.kind === 'success' ? 200 : '-')}` }
  ]
}

function kindBadgeClass(kind: string): string {
  if (kind === 'error') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
}

function formatPlatform(platform?: string): string {
  const value = (platform || '').trim()
  return value ? value.toUpperCase() : '-'
}

function formatDuration(value?: number | null): string {
  return typeof value === 'number' ? `${value} ms` : '-'
}

async function copyText(text: string, message: string) {
  if (!text) return
  try {
    if (!navigator.clipboard?.writeText) throw new Error('clipboard unavailable')
    await navigator.clipboard.writeText(text)
    appStore.showSuccess(message)
  } catch {
    appStore.showWarning(ui.value.copyFailed)
  }
}

function toLocalInputValue(date: Date): string {
  const offsetMs = date.getTimezoneOffset() * 60 * 1000
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16)
}

function emptyFilterOptions(): OpsRequestFilterOptions {
  return {
    platforms: [],
    models: [],
    users: [],
    api_keys: [],
    accounts: [],
    groups: []
  }
}

function normalizeFilterOptions(value?: Partial<OpsRequestFilterOptions> | null): OpsRequestFilterOptions {
  return {
    platforms: value?.platforms || [],
    models: value?.models || [],
    users: value?.users || [],
    api_keys: value?.api_keys || [],
    accounts: value?.accounts || [],
    groups: value?.groups || []
  }
}

function pruneMissingOptionSelections() {
  filters.platform = keepIfOptionExists(filters.platform, filterOptions.value.platforms)
  filters.model = keepIfOptionExists(filters.model, filterOptions.value.models)
  filters.user_id = keepIfOptionExists(filters.user_id, filterOptions.value.users)
  filters.api_key_id = keepIfOptionExists(filters.api_key_id, filterOptions.value.api_keys)
  filters.account_id = keepIfOptionExists(filters.account_id, filterOptions.value.accounts)
  filters.group_id = keepIfOptionExists(filters.group_id, filterOptions.value.groups)
}

function keepIfOptionExists(value: string, options: Array<{ value: string }>): string {
  if (!value) return ''
  return options.some((option) => option.value === value) ? value : ''
}

function assignPositiveNumber<K extends keyof OpsRequestDetailsParams>(params: OpsRequestDetailsParams, key: K, value: string) {
  const parsed = Number(value)
  if (Number.isFinite(parsed) && parsed > 0) {
    ;(params as Record<string, unknown>)[key] = Math.trunc(parsed)
  }
}

function assignNonNegativeNumber<K extends keyof OpsRequestDetailsParams>(params: OpsRequestDetailsParams, key: K, value: string) {
  const parsed = Number(value)
  if (Number.isFinite(parsed) && parsed >= 0) {
    ;(params as Record<string, unknown>)[key] = Math.trunc(parsed)
  }
}
</script>

<style scoped>
.table-head {
  @apply px-4 py-3 text-left text-[11px] font-semibold uppercase text-gray-500 dark:text-gray-400;
}

.table-cell {
  @apply px-4 py-3 align-top;
}

.chip {
  @apply rounded-md bg-gray-100 px-1.5 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300;
}

.param-chip {
  @apply max-w-full truncate rounded-md border border-gray-200 bg-gray-50 px-2 py-1 text-[11px] dark:border-dark-600 dark:bg-dark-700;
}

.icon-button {
  @apply inline-flex h-8 w-8 items-center justify-center rounded-md border border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-dark-600 dark:text-gray-300 dark:hover:bg-dark-700;
}
</style>
