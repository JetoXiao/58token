<template>
  <AppLayout>
    <div class="space-y-6 pb-12">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">首字分析</h1>
          <p class="mt-2 max-w-3xl text-sm text-gray-500 dark:text-dark-400">
            统计 Time To First Token、慢因占比、Top 模型/账号/分组/API Key，并给出可操作建议。当前页面只做观测，不改变调度。
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <select v-model="filters.timeRange" class="h-10 rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 dark:border-dark-700 dark:bg-dark-800 dark:text-white">
            <option value="30m">30 分钟</option>
            <option value="1h">1 小时</option>
            <option value="6h">6 小时</option>
            <option value="24h">24 小时</option>
            <option value="7d">7 天</option>
          </select>
          <label class="inline-flex h-10 items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-700 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200">
            <span>慢阈值</span>
            <input v-model.number="filters.slowThresholdMs" type="number" min="100" step="100" class="w-20 bg-transparent text-right outline-none" />
            <span>ms</span>
          </label>
          <button
            type="button"
            class="inline-flex h-10 items-center gap-2 rounded-lg bg-primary-600 px-4 text-sm font-medium text-white transition-colors hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="loading"
            @click="fetchData"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            刷新
          </button>
        </div>
      </div>

      <div v-if="errorMessage" class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300">
        {{ errorMessage }}
      </div>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard title="首字样本" :value="formatInt(summary.first_token_sample_count)" :hint="`请求 ${formatInt(summary.request_count)} 次`" />
        <MetricCard title="TTFT P95" :value="formatMs(summary.ttft?.p95_ms)" :hint="`P99 ${formatMs(summary.ttft?.p99_ms)} / Max ${formatMs(summary.ttft?.max_ms)}`" :tone="thresholdTone(summary.ttft?.p95_ms)" />
        <MetricCard title="慢请求占比" :value="formatPercent(summary.slow_rate)" :hint="`慢请求 ${formatInt(summary.slow_request_count)} 次，阈值 ${formatInt(data?.slow_threshold_ms ?? filters.slowThresholdMs)}ms`" :tone="summary.slow_rate >= 0.3 ? 'warning' : 'normal'" />
        <MetricCard title="上游前 P95" :value="formatMs(summary.pre_upstream?.p95_ms)" :hint="`auth + routing + queue/conn，Avg ${formatMs(summary.pre_upstream?.avg_ms)}`" :tone="thresholdTone(summary.pre_upstream?.p95_ms)" />
      </div>

      <section class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col gap-1">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">请求来源占比</h2>
          <p class="text-sm text-gray-500 dark:text-dark-400">区分自有号池和上游/API Key 路径，先判断慢请求主要落在哪类账号链路。</p>
        </div>
        <div v-if="routeSources.length" class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-3">
          <div v-for="item in routeSources" :key="item.key" class="rounded-lg border border-gray-100 p-3 dark:border-dark-700">
            <div class="flex items-center justify-between gap-3">
              <div>
                <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ routeSourceLabel(item.key) }}</div>
                <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ item.key }}</div>
              </div>
              <div class="text-right">
                <div class="text-lg font-semibold text-gray-900 dark:text-white">{{ formatPercent(item.share) }}</div>
                <div class="text-xs text-gray-500">{{ formatInt(item.count) }} 次</div>
              </div>
            </div>
            <div class="mt-3 flex items-center justify-between text-xs text-gray-500 dark:text-dark-400">
              <span>慢请求 {{ formatInt(item.slowCount) }} 次</span>
              <span>慢占比 {{ formatPercent(item.slowRate) }}</span>
            </div>
            <div class="mt-3 h-2 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
              <div class="h-full rounded bg-emerald-500" :style="{ width: `${Math.min(item.share * 100, 100)}%` }"></div>
            </div>
          </div>
        </div>
        <EmptyState v-else class="mt-6" :title="'暂无来源数据'" :description="'当前窗口内没有可统计的请求来源。'" />
      </section>

      <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <section class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800 lg:col-span-2">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">慢因占比</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">按请求记录里的阶段耗时和首字样本做规则归因。</p>
            </div>
          </div>
          <div v-if="reasons.length" class="space-y-3">
            <div v-for="item in reasons" :key="item.reason" class="rounded-lg border border-gray-100 p-3 dark:border-dark-700">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ reasonLabel(item.reason) }}</div>
                  <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ item.suggestion }}</div>
                </div>
                <div class="text-right">
                  <div class="text-lg font-semibold text-gray-900 dark:text-white">{{ formatPercent(item.share) }}</div>
                  <div class="text-xs text-gray-500">{{ formatInt(item.count) }} 次</div>
                </div>
              </div>
              <div class="mt-3 h-2 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
                <div class="h-full rounded bg-primary-500" :style="{ width: `${Math.min(item.share * 100, 100)}%` }"></div>
              </div>
            </div>
          </div>
          <EmptyState v-else :title="'暂无慢因数据'" :description="'当前窗口没有可归因的首字样本。'" />
        </section>

        <section class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">系统建议</h2>
          <div v-if="recommendations.length" class="mt-4 space-y-3">
            <div v-for="item in recommendations" :key="`${item.title}-${item.message}`" class="rounded-lg border p-3" :class="recommendationClass(item.severity)">
              <div class="text-sm font-semibold">{{ item.title }}</div>
              <div class="mt-1 text-sm">{{ item.message }}</div>
              <div class="mt-2 text-xs opacity-80">{{ item.action }}</div>
            </div>
          </div>
          <EmptyState v-else class="mt-6" :title="'暂无建议'" :description="'数据不足或未发现明显集中慢因。'" />
        </section>
      </div>

      <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <TopTable title="Top 模型" :items="data?.top_models ?? []" />
        <TopTable title="Top 账号" :items="data?.top_accounts ?? []" />
        <TopTable title="Top 分组" :items="data?.top_groups ?? []" />
        <TopTable title="Top API Key" :items="data?.top_api_keys ?? []" />
      </div>

      <section class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">账号参与调度</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">按当前窗口的观测分组和平台列出上游 APIKey 账号，区分已选中、候选未选中和未进入分组。</p>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
            <thead>
              <tr class="text-left text-xs font-semibold uppercase text-gray-500 dark:text-dark-400">
                <th class="px-3 py-2">账号</th>
                <th class="px-3 py-2">参与状态</th>
                <th class="px-3 py-2">分组</th>
                <th class="px-3 py-2">请求</th>
                <th class="px-3 py-2">慢占比</th>
                <th class="px-3 py-2">P95</th>
                <th class="px-3 py-2">状态</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="item in accountParticipation" :key="item.account_id" class="text-gray-700 dark:text-dark-200">
                <td class="px-3 py-2">
                  <div class="font-medium text-gray-900 dark:text-white">{{ item.account_name || item.account_id }}</div>
                  <div class="mt-1 text-xs text-gray-500">{{ item.platform }} / {{ item.account_type }}</div>
                </td>
                <td class="px-3 py-2">{{ participationReasonLabel(item.participation_reason) }}</td>
                <td class="px-3 py-2 text-xs text-gray-500">{{ formatGroupIDs(item.observed_group_ids) }}</td>
                <td class="px-3 py-2">{{ formatInt(item.request_count) }}</td>
                <td class="px-3 py-2">{{ formatPercent(item.slow_rate) }}</td>
                <td class="px-3 py-2">{{ formatMs(item.p95_ttft_ms) }}</td>
                <td class="px-3 py-2 text-xs">
                  <span :class="item.schedulable && item.status === 'active' ? 'text-emerald-600 dark:text-emerald-300' : 'text-amber-600 dark:text-amber-300'">
                    {{ item.status }} / {{ item.schedulable ? 'schedulable' : 'unschedulable' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <EmptyState v-if="!accountParticipation.length" class="mt-6" :title="'暂无账号参与数据'" :description="'当前窗口内还没有可关联的上游账号。'" />
      </section>

      <section class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center justify-between">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">慢请求明细</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">展示当前窗口内首字最慢的请求样本，便于按 request_id 溯源。</p>
          </div>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
            <thead>
              <tr class="text-left text-xs font-semibold uppercase text-gray-500 dark:text-dark-400">
                <th class="px-3 py-2">时间</th>
                <th class="px-3 py-2">模型</th>
                <th class="px-3 py-2">来源</th>
                <th class="px-3 py-2">账号</th>
                <th class="px-3 py-2">分组</th>
                <th class="px-3 py-2">TTFT</th>
                <th class="px-3 py-2">慢因</th>
                <th class="px-3 py-2">上游前/上游</th>
                <th class="px-3 py-2">调度</th>
                <th class="px-3 py-2">request_id</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in slowRequests" :key="`${row.request_id}-${row.created_at}`" class="text-gray-700 dark:text-dark-200">
                <td class="whitespace-nowrap px-3 py-2 text-xs text-gray-500">{{ formatDateTime(row.created_at) }}</td>
                <td class="px-3 py-2">{{ row.model || '-' }}</td>
                <td class="whitespace-nowrap px-3 py-2">{{ routeSourceLabel(row.route_source) }}</td>
                <td class="px-3 py-2">{{ row.account_name || row.account_id || '-' }}</td>
                <td class="px-3 py-2">{{ row.group_name || row.group_id || '-' }}</td>
                <td class="whitespace-nowrap px-3 py-2 font-semibold text-gray-900 dark:text-white">{{ formatMs(row.first_token_ms) }}</td>
                <td class="px-3 py-2">
                  <div>{{ reasonLabel(row.slow_reason) }}</div>
                  <div v-if="row.slow_detail" class="mt-1 text-xs text-gray-500">{{ row.slow_detail }}</div>
                </td>
                <td class="px-3 py-2 text-xs text-gray-500">
                  <div>pre {{ formatMs(row.pre_upstream_ms) }} / upstream {{ formatMs(row.upstream_latency_ms) }}</div>
                  <div>auth {{ formatMs(row.auth_latency_ms) }} / route {{ formatMs(row.routing_latency_ms) }}</div>
                  <div>queue {{ formatMs(row.queue_wait_ms) }} / conn {{ formatMs(row.conn_pick_ms) }}</div>
                </td>
                <td class="px-3 py-2 text-xs text-gray-500">
                  <div>{{ schedulerSummary(row) }}</div>
                  <div v-if="row.scheduler_reason" class="mt-1">{{ schedulerReasonLabel(row.scheduler_reason) }}</div>
                  <div v-if="schedulerCandidates(row).length" class="mt-2 space-y-1">
                    <div
                      v-for="candidate in schedulerCandidates(row)"
                      :key="`${row.request_id}-${candidate.account_id}`"
                      :class="candidate.selected ? 'font-semibold text-emerald-600 dark:text-emerald-300' : 'text-gray-500 dark:text-dark-400'"
                    >
                      {{ schedulerCandidateLabel(candidate) }}
                    </div>
                  </div>
                </td>
                <td class="max-w-[220px] truncate px-3 py-2 font-mono text-xs text-gray-500">{{ row.request_id || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <EmptyState v-if="!slowRequests.length" class="mt-6" :title="'暂无慢请求'" :description="'当前窗口内没有超过阈值的首字样本。'" />
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { opsAPI, type OpsTTFTAnalysisResponse, type OpsTTFTSlowRequest, type OpsTTFTTopItem } from '@/api/admin/ops'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from './utils/opsFormatters'

const appStore = useAppStore()
const loading = ref(false)
const errorMessage = ref('')
const data = ref<OpsTTFTAnalysisResponse | null>(null)
const filters = reactive({
  timeRange: '24h',
  slowThresholdMs: 1000,
  limit: 10
})

const summary = computed(() => data.value?.summary ?? {
  request_count: 0,
  first_token_sample_count: 0,
  slow_request_count: 0,
  slow_rate: 0,
  cache_hit_count: 0,
  cache_hit_rate: 0,
  by_route_source: {},
  route_sources: [],
  ttft: {},
  duration: {},
  pre_upstream: {}
})
const reasons = computed(() => data.value?.reasons ?? [])
const recommendations = computed(() => data.value?.recommendations ?? [])
const slowRequests = computed(() => data.value?.slow_requests ?? [])
const accountParticipation = computed(() => data.value?.account_participation ?? [])

interface SchedulerCandidateDiagnostic {
  account_id?: number
  account_name?: string
  account_type?: string
  priority?: number
  concurrency?: number
  load_factor?: number
  current_concurrency?: number
  waiting_count?: number
  load_rate?: number
  score?: number
  error_rate?: number
  ttft_ms?: number
  has_ttft?: boolean
  rank?: number
  order?: number
  in_top_k?: boolean
  selected?: boolean
  last_used_at?: string
}

const routeSources = computed(() => {
  const routeSourceItems = summary.value.route_sources ?? []
  if (routeSourceItems.length) {
    return routeSourceItems
      .map((item) => ({
        key: item.source,
        count: Number(item.count || 0),
        share: Number(item.share || 0),
        slowCount: Number(item.slow_count || 0),
        slowRate: Number(item.slow_rate || 0)
      }))
      .sort((a, b) => b.count - a.count)
  }
  const source = summary.value.by_route_source ?? {}
  const total = Object.values(source).reduce((sum, count) => sum + Number(count || 0), 0)
  return Object.entries(source)
    .map(([key, count]) => ({
      key,
      count: Number(count || 0),
      share: total > 0 ? Number(count || 0) / total : 0,
      slowCount: 0,
      slowRate: 0
    }))
    .sort((a, b) => b.count - a.count)
})

async function fetchData() {
  loading.value = true
  errorMessage.value = ''
  try {
    data.value = await opsAPI.getTTFTAnalysis({
      time_range: filters.timeRange as any,
      slow_threshold_ms: filters.slowThresholdMs,
      limit: filters.limit
    })
  } catch (err: any) {
    errorMessage.value = err?.response?.data?.detail || err?.message || '加载首字分析失败'
    appStore.showError(errorMessage.value)
  } finally {
    loading.value = false
  }
}

watch(() => [filters.timeRange], () => {
  fetchData()
})

onMounted(fetchData)

function formatInt(value?: number | null) {
  return Number(value ?? 0).toLocaleString()
}

function formatMs(value?: number | null) {
  if (value == null) return '-'
  return `${Math.round(value)}ms`
}

function formatPercent(value?: number | null) {
  return `${((value ?? 0) * 100).toFixed(1)}%`
}

function thresholdTone(value?: number | null) {
  if (value == null) return 'normal'
  if (value >= 2000) return 'danger'
  if (value >= 1000) return 'warning'
  return 'normal'
}

function reasonLabel(reason: string) {
  const labels: Record<string, string> = {
    normal: '正常',
    unknown: '数据不足',
    cache_hit: '缓存命中',
    account_queue_slow: '账号池排队慢',
    connection_pick_slow: '连接选择慢',
    routing_slow: '路由/账号选择慢',
    upstream_ttft_slow: '上游首字慢',
    response_flush_slow: '响应尾部慢（非首字前）',
    platform_overhead_slow: '平台前置处理慢'
  }
  return labels[reason] || reason || '-'
}

function schedulerSummary(row: OpsTTFTSlowRequest) {
  const parts: string[] = []
  if (row.scheduler_layer) {
    parts.push(schedulerLayerLabel(row.scheduler_layer))
  }
  if (row.scheduler_candidate_count != null) {
    parts.push(`候选 ${formatInt(row.scheduler_candidate_count)}`)
  }
  if (row.scheduler_top_k != null) {
    parts.push(`TopK ${formatInt(row.scheduler_top_k)}`)
  }
  if (row.scheduler_latency_ms != null) {
    parts.push(`调度 ${formatMs(row.scheduler_latency_ms)}`)
  }
  return parts.length ? parts.join(' / ') : '-'
}

function schedulerLayerLabel(layer: string) {
  const labels: Record<string, string> = {
    previous_response_id: '上文粘性',
    session_hash: '会话粘性',
    load_balance: '负载选择'
  }
  return labels[layer] || layer
}

function schedulerReasonLabel(reason: string) {
  const labels: Record<string, string> = {
    previous_response_sticky: '沿用 previous_response_id 绑定账号',
    session_sticky_hit: '命中会话粘性账号',
    session_sticky: '按会话粘性选择账号',
    legacy_session_sticky_hit: 'legacy 调度命中会话粘性账号',
    load_balance_top_k: '按优先级/负载/队列/TTFT 评分进入 TopK 后选择',
    load_balance_priority_load_queue_ttft: '按优先级/负载/队列/TTFT 评分选择'
  }
  return labels[reason] || reason
}

function schedulerCandidates(row: OpsTTFTSlowRequest): SchedulerCandidateDiagnostic[] {
  const raw = row.request_params?.scheduler_candidates
  if (!Array.isArray(raw)) return []
  return raw
    .filter((item): item is Record<string, unknown> => !!item && typeof item === 'object')
    .map(item => ({
      account_id: numberField(item.account_id),
      account_name: stringField(item.account_name),
      account_type: stringField(item.account_type),
      priority: numberField(item.priority),
      concurrency: numberField(item.concurrency),
      load_factor: numberField(item.load_factor),
      current_concurrency: numberField(item.current_concurrency),
      waiting_count: numberField(item.waiting_count),
      load_rate: numberField(item.load_rate),
      score: numberField(item.score),
      error_rate: numberField(item.error_rate),
      ttft_ms: numberField(item.ttft_ms),
      has_ttft: Boolean(item.has_ttft),
      rank: numberField(item.rank),
      order: numberField(item.order),
      in_top_k: Boolean(item.in_top_k),
      selected: Boolean(item.selected),
      last_used_at: stringField(item.last_used_at)
    }))
}

function schedulerCandidateLabel(candidate: SchedulerCandidateDiagnostic) {
  const name = candidate.account_name || (candidate.account_id != null ? `#${candidate.account_id}` : '-')
  const prefix = candidate.selected ? '选中' : '候选'
  const parts = [
    `${prefix} ${name}`,
    candidate.account_id != null ? `#${candidate.account_id}` : '',
    candidate.rank != null ? `rank ${formatInt(candidate.rank)}` : '',
    candidate.order != null ? `order ${formatInt(candidate.order)}` : '',
    candidate.priority != null ? `pri ${formatInt(candidate.priority)}` : '',
    candidate.load_rate != null ? `load ${formatInt(candidate.load_rate)}%` : '',
    candidate.waiting_count != null ? `wait ${formatInt(candidate.waiting_count)}` : '',
    candidate.current_concurrency != null && candidate.concurrency != null ? `conc ${formatInt(candidate.current_concurrency)}/${formatInt(candidate.concurrency)}` : '',
    candidate.score != null ? `score ${candidate.score.toFixed(3)}` : '',
    candidate.has_ttft && candidate.ttft_ms != null ? `ttft ${formatMs(candidate.ttft_ms)}` : ''
  ]
  return parts.filter(Boolean).join(' / ')
}

function numberField(value: unknown): number | undefined {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    if (Number.isFinite(parsed)) return parsed
  }
  return undefined
}

function stringField(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() !== '' ? value : undefined
}

function participationReasonLabel(reason: string) {
  const labels: Record<string, string> = {
    selected: '已被选中',
    candidate_not_selected: '候选但未选中',
    not_in_observed_group: '未进入观测分组',
    not_active: '账号非 active',
    unschedulable: '不可调度'
  }
  return labels[reason] || reason || '-'
}

function formatGroupIDs(groupIDs?: number[] | null) {
  if (!groupIDs?.length) return '-'
  return groupIDs.join(', ')
}

function routeSourceLabel(source?: string | null) {
  const labels: Record<string, string> = {
    own_pool: '自有号池',
    upstream: '上游/API Key',
    cache: '缓存命中',
    unknown: '未知'
  }
  const key = source || 'unknown'
  return labels[key] || key
}

function recommendationClass(severity: string) {
  if (severity === 'critical') return 'border-red-200 bg-red-50 text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300'
  if (severity === 'warning') return 'border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-200'
  return 'border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-900/50 dark:bg-blue-900/20 dark:text-blue-300'
}

const MetricCard = defineComponent({
  props: {
    title: { type: String, required: true },
    value: { type: String, required: true },
    hint: { type: String, default: '' },
    tone: { type: String, default: 'normal' }
  },
  setup(props) {
    return () => h('div', { class: 'rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800' }, [
      h('p', { class: 'text-sm font-medium text-gray-500 dark:text-dark-400' }, props.title),
      h('p', { class: ['mt-3 text-3xl font-semibold', props.tone === 'danger' ? 'text-red-600 dark:text-red-300' : props.tone === 'warning' ? 'text-amber-600 dark:text-amber-300' : 'text-gray-900 dark:text-white'] }, props.value),
      h('p', { class: 'mt-2 text-xs text-gray-500 dark:text-dark-400' }, props.hint)
    ])
  }
})

const TopTable = defineComponent({
  props: {
    title: { type: String, required: true },
    items: { type: Array as () => OpsTTFTTopItem[], default: () => [] }
  },
  setup(props) {
    return () => h('section', { class: 'rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800' }, [
      h('h2', { class: 'text-base font-semibold text-gray-900 dark:text-white' }, props.title),
      props.items.length
        ? h('div', { class: 'mt-4 overflow-x-auto' }, [
            h('table', { class: 'min-w-full text-sm' }, [
              h('thead', [
                h('tr', { class: 'text-left text-xs font-semibold uppercase text-gray-500' }, [
                  h('th', { class: 'px-2 py-2' }, '对象'),
                  h('th', { class: 'px-2 py-2' }, '请求'),
                  h('th', { class: 'px-2 py-2' }, '慢占比'),
                  h('th', { class: 'px-2 py-2' }, 'P95'),
                  h('th', { class: 'px-2 py-2' }, '建议')
                ])
              ]),
              h('tbody', props.items.map((item) =>
                h('tr', { class: 'border-t border-gray-100 dark:border-dark-700' }, [
                  h('td', { class: 'px-2 py-2 text-gray-900 dark:text-white' }, item.label || item.key),
                  h('td', { class: 'px-2 py-2 text-gray-600 dark:text-dark-300' }, formatInt(item.count)),
                  h('td', { class: 'px-2 py-2 text-gray-600 dark:text-dark-300' }, formatPercent(item.slow_rate)),
                  h('td', { class: 'px-2 py-2 text-gray-600 dark:text-dark-300' }, formatMs(item.p95_ttft_ms)),
                  h('td', { class: 'px-2 py-2 text-xs text-gray-500' }, item.suggestion || '-')
                ])
              ))
            ])
          ])
        : h(EmptyState, { class: 'mt-6', title: '暂无数据', description: '当前窗口内没有足够样本。' })
    ])
  }
})
</script>
