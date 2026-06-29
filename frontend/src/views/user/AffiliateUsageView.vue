<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-6">
        <div class="summary-tile">
          <p>{{ t('affiliate.usage.summaryInvitees') }}</p>
          <strong>{{ formatInteger(records.length ? pagination.total : 0) }}</strong>
        </div>
        <div class="summary-tile">
          <p>{{ t('affiliate.usage.summaryRequests') }}</p>
          <strong>{{ formatInteger(summary.total_requests) }}</strong>
        </div>
        <div class="summary-tile">
          <p>{{ t('affiliate.usage.summaryRecharge') }}</p>
          <strong>{{ formatCnyAmount(summary.total_recharge_amount) }}</strong>
        </div>
        <div class="summary-tile">
          <p>{{ t('affiliate.usage.summaryRebate') }}</p>
          <strong class="text-primary-600 dark:text-primary-400">{{ formatCnyAmount(summary.total_rebate_amount) }}</strong>
        </div>
        <div class="summary-tile">
          <p>{{ t('affiliate.usage.summarySettled') }}</p>
          <strong class="text-emerald-600 dark:text-emerald-400">{{ formatCnyAmount(summary.total_settled_amount) }}</strong>
        </div>
        <div class="summary-tile">
          <p>{{ t('affiliate.usage.summaryPending') }}</p>
          <strong class="text-amber-600 dark:text-amber-400">{{ formatCnyAmount(summary.total_pending_amount) }}</strong>
        </div>
      </div>

      <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-wrap items-center gap-3">
          <div class="relative w-full md:w-80">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="filters.search"
              class="input pl-10"
              :placeholder="t('affiliate.usage.searchPlaceholder')"
              @input="debounceLoad"
            />
          </div>
          <input v-model="filters.start_at" type="date" class="input w-full sm:w-44" @change="reloadFromFirstPage" />
          <input v-model="filters.end_at" type="date" class="input w-full sm:w-44" @change="reloadFromFirstPage" />
          <button class="btn btn-secondary px-3" :disabled="loading" :title="t('common.refresh')" @click="loadRecords">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button class="btn btn-secondary" type="button" @click="openSettlementRecordsDialog">
            <Icon name="clipboard" size="md" />
            <span>{{ t('affiliate.usage.settlementRecordsButton') }}</span>
          </button>
        </div>
      </div>

      <div class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
        <div class="overflow-x-auto">
          <table class="w-full min-w-[64rem] divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-900">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                  <button type="button" class="sort-header justify-start" @click="toggleSort('invitee')">
                    <span>{{ t('affiliate.usage.invitee') }}</span>
                    <Icon :name="sortIconName('invitee')" size="xs" />
                  </button>
                </th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                  <button type="button" class="sort-header justify-end" @click="toggleSort('requests')">
                    <span>{{ t('affiliate.usage.requests') }}</span>
                    <Icon :name="sortIconName('requests')" size="xs" />
                  </button>
                </th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                  <button type="button" class="sort-header justify-end" @click="toggleSort('total_tokens')">
                    <span>{{ t('affiliate.usage.totalTokens') }}</span>
                    <Icon :name="sortIconName('total_tokens')" size="xs" />
                  </button>
                </th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                  <button type="button" class="sort-header justify-end" @click="toggleSort('actual_cost')">
                    <span>{{ t('affiliate.usage.actualCost') }}</span>
                    <Icon :name="sortIconName('actual_cost')" size="xs" />
                  </button>
                </th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                  <button type="button" class="sort-header justify-end" @click="toggleSort('recharge_amount')">
                    <span>{{ t('affiliate.usage.rechargeAmount') }}</span>
                    <Icon :name="sortIconName('recharge_amount')" size="xs" />
                  </button>
                </th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                  <button type="button" class="sort-header justify-end" @click="toggleSort('rebate_amount')">
                    <span>{{ t('affiliate.usage.rebateAmount') }}</span>
                    <Icon :name="sortIconName('rebate_amount')" size="xs" />
                  </button>
                </th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ t('affiliate.usage.details') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-if="loading">
                <td colspan="7" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('common.loading') }}</td>
              </tr>
              <tr v-else-if="records.length === 0">
                <td colspan="7" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('empty.noData') }}</td>
              </tr>
              <tr v-for="row in records" v-else :key="row.invitee_id" class="hover:bg-gray-50 dark:hover:bg-dark-900/50">
                <td class="px-4 py-3">
                  <div class="font-mono text-sm text-gray-900 dark:text-white">#{{ row.invitee_id }}</div>
                  <div class="max-w-64 truncate text-sm font-medium text-gray-900 dark:text-white">{{ row.invitee_email || '-' }}</div>
                  <div class="max-w-64 truncate text-xs text-gray-500 dark:text-dark-400">{{ row.invitee_username || '-' }}</div>
                </td>
                <td class="px-4 py-3 text-right font-mono text-sm text-gray-700 dark:text-dark-200">{{ formatInteger(row.requests) }}</td>
                <td class="px-4 py-3 text-right font-mono text-sm text-gray-900 dark:text-white">{{ formatTokens(row.total_tokens) }}</td>
                <td class="px-4 py-3 text-right font-mono text-sm text-emerald-600 dark:text-emerald-400">{{ formatUsdAmount(row.actual_cost) }}</td>
                <td class="px-4 py-3 text-right font-mono text-sm text-sky-600 dark:text-sky-400">{{ formatCnyAmount(row.recharge_amount) }}</td>
                <td class="px-4 py-3 text-right font-mono text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatCnyAmount(row.rebate_amount) }}</td>
                <td class="px-4 py-3 text-right">
                  <button class="btn btn-secondary btn-sm" @click="openDetails(row)">
                    <Icon name="eye" size="sm" />
                    <span>{{ t('common.view') }}</span>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <BaseDialog :show="detailsDialog" :title="t('affiliate.usage.detailsTitle')" width="extra-wide" @close="closeDetails">
      <div v-if="selectedRecord" class="space-y-4">
        <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900">
          <div class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.usage.invitee') }}</div>
          <div class="mt-1 font-mono text-sm text-gray-900 dark:text-white">#{{ selectedRecord.invitee_id }} {{ selectedRecord.invitee_email || selectedRecord.invitee_username || '-' }}</div>
        </div>

        <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="w-full min-w-[48rem] divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ t('affiliate.usage.source') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ t('affiliate.usage.group') }}</th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ t('affiliate.usage.requests') }}</th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ t('affiliate.usage.profitDetailAmount') }}</th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ t('affiliate.usage.profitRate') }}</th>
                <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ t('affiliate.usage.rebateAmount') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="detail in selectedRecord.profit_details || []" :key="detailKey(detail)">
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-dark-200">
                  {{ detail.source === 'subscription' ? t('affiliate.usage.subscriptionSource') : detail.model }}
                </td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-dark-200">{{ detail.group_name || `#${detail.group_id}` }}</td>
                <td class="px-4 py-3 text-right font-mono text-sm text-gray-700 dark:text-dark-200">{{ formatInteger(detail.requests) }}</td>
                <td class="px-4 py-3 text-right font-mono text-sm text-emerald-600 dark:text-emerald-400">{{ formatProfitDetailActualCost(detail) }}</td>
                <td class="px-4 py-3 text-right font-mono text-sm text-gray-700 dark:text-dark-200">{{ formatPercent(detail.profit_rate_percent) }}</td>
                <td class="px-4 py-3 text-right font-mono text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatCnyAmount(detail.rebate_amount) }}</td>
              </tr>
              <tr v-if="!(selectedRecord.profit_details || []).length">
                <td colspan="6" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('empty.noData') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </BaseDialog>

    <BaseDialog
      :show="settlementRecordsDialog"
      :title="t('affiliate.usage.settlementRecordsTitle')"
      width="extra-wide"
      @close="closeSettlementRecordsDialog"
    >
      <div class="space-y-4">
        <div class="flex flex-wrap items-center gap-3">
          <div class="relative w-full md:w-80">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="settlementRecordFilters.search"
              class="input pl-10"
              :placeholder="t('affiliate.usage.settlementRecordSearch')"
              @input="debounceSettlementRecordsLoad"
            />
          </div>
          <input v-model="settlementRecordFilters.start_at" type="date" class="input w-full sm:w-44" @change="loadSettlementRecordsFromFirstPage" />
          <input v-model="settlementRecordFilters.end_at" type="date" class="input w-full sm:w-44" @change="loadSettlementRecordsFromFirstPage" />
          <button class="btn btn-secondary px-3" :disabled="settlementRecordsLoading" :title="t('common.refresh')" @click="loadSettlementRecords">
            <Icon name="refresh" size="md" :class="settlementRecordsLoading ? 'animate-spin' : ''" />
          </button>
        </div>

        <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <div class="overflow-x-auto">
            <table class="w-full min-w-[46rem] divide-y divide-gray-200 dark:divide-dark-700">
              <thead class="bg-gray-50 dark:bg-dark-900">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                    <button type="button" class="sort-header justify-start" @click="toggleSettlementSort('settled_on')">
                      <span>{{ t('affiliate.usage.settlementDate') }}</span>
                      <Icon :name="settlementSortIconName('settled_on')" size="xs" />
                    </button>
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                    <button type="button" class="sort-header justify-end" @click="toggleSettlementSort('amount')">
                      <span>{{ t('affiliate.usage.settlementAmount') }}</span>
                      <Icon :name="settlementSortIconName('amount')" size="xs" />
                    </button>
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                    <button type="button" class="sort-header justify-start" @click="toggleSettlementSort('created_at')">
                      <span>{{ t('affiliate.usage.settlementCreatedAt') }}</span>
                      <Icon :name="settlementSortIconName('created_at')" size="xs" />
                    </button>
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-dark-400">
                    {{ t('affiliate.usage.settlementNote') }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-if="settlementRecordsLoading">
                  <td colspan="4" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('common.loading') }}</td>
                </tr>
                <tr v-else-if="settlementRecords.length === 0">
                  <td colspan="4" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('empty.noData') }}</td>
                </tr>
                <tr v-for="record in settlementRecords" v-else :key="record.id" class="hover:bg-gray-50 dark:hover:bg-dark-900/50">
                  <td class="px-4 py-3 font-mono text-sm text-gray-900 dark:text-white">{{ formatDateOnly(record.settled_on) }}</td>
                  <td class="px-4 py-3 text-right font-mono text-sm font-semibold text-emerald-600 dark:text-emerald-400">{{ formatCnyAmount(record.amount) }}</td>
                  <td class="px-4 py-3 font-mono text-sm text-gray-700 dark:text-dark-200">{{ formatDateTime(record.created_at) }}</td>
                  <td class="px-4 py-3 text-sm text-gray-700 dark:text-dark-200">
                    <div class="max-w-xl whitespace-pre-wrap break-words">{{ record.note || '-' }}</div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <Pagination
          v-if="settlementRecordPagination.total > 0"
          :page="settlementRecordPagination.page"
          :total="settlementRecordPagination.total"
          :page-size="settlementRecordPagination.page_size"
          @update:page="handleSettlementRecordPageChange"
          @update:pageSize="handleSettlementRecordPageSizeChange"
        />
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import { userAPI } from '@/api/user'
import type {
  AffiliateSettlementRecord,
  AffiliateUsageDailyRecord,
  AffiliateUsageProfitDetail,
  AffiliateUsageSummary,
} from '@/api/admin/affiliates'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const records = ref<AffiliateUsageDailyRecord[]>([])
const summary = reactive<AffiliateUsageSummary>({
  total_requests: 0,
  total_tokens: 0,
  total_actual_cost: 0,
  total_account_cost: 0,
  total_net_profit: 0,
  total_recharge_amount: 0,
  total_rebate_amount: 0,
  total_settled_amount: 0,
  total_pending_amount: 0,
})
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const filters = reactive({
  search: '',
  start_at: defaultDateInput(-29),
  end_at: defaultDateInput(0),
})
const sortState = reactive({ sort_by: 'actual_cost', sort_order: 'desc' as 'asc' | 'desc' })

const detailsDialog = ref(false)
const selectedRecord = ref<AffiliateUsageDailyRecord | null>(null)
const settlementRecordsDialog = ref(false)
const settlementRecordsLoading = ref(false)
const settlementRecords = ref<AffiliateSettlementRecord[]>([])
const settlementRecordFilters = reactive({
  search: '',
  start_at: defaultDateInput(-29),
  end_at: defaultDateInput(0),
})
const settlementRecordPagination = reactive({ page: 1, page_size: 20, total: 0 })
const settlementRecordSort = reactive({ sort_by: 'settled_on', sort_order: 'desc' as 'asc' | 'desc' })
let searchTimer: ReturnType<typeof setTimeout> | null = null
let settlementRecordsTimer: ReturnType<typeof setTimeout> | null = null

async function loadRecords() {
  loading.value = true
  try {
    const res = await userAPI.listAffiliateUsage({
      page: pagination.page,
      page_size: pagination.page_size,
      search: filters.search,
      start_at: filters.start_at,
      end_at: filters.end_at,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      sort_by: sortState.sort_by,
      sort_order: sortState.sort_order,
      view: 'users',
    })
    records.value = res.items || []
    pagination.total = Number(res.total || 0)
    pagination.page = Number(res.page || pagination.page)
    pagination.page_size = Number(res.page_size || pagination.page_size)
    Object.assign(summary, res.summary || {})
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || t('affiliate.usage.loadFailed'))
  } finally {
    loading.value = false
  }
}

function debounceLoad() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(reloadFromFirstPage, 300)
}

function reloadFromFirstPage() {
  pagination.page = 1
  void loadRecords()
}

function handlePageChange(page: number) {
  pagination.page = page
  void loadRecords()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadRecords()
}

function toggleSort(key: string) {
  if (sortState.sort_by === key) {
    sortState.sort_order = sortState.sort_order === 'asc' ? 'desc' : 'asc'
  } else {
    sortState.sort_by = key
    sortState.sort_order = 'desc'
  }
  reloadFromFirstPage()
}

function sortIconName(key: string): 'sort' | 'chevronUp' | 'chevronDown' {
  if (sortState.sort_by !== key) return 'sort'
  return sortState.sort_order === 'asc' ? 'chevronUp' : 'chevronDown'
}

function openDetails(row: AffiliateUsageDailyRecord) {
  selectedRecord.value = row
  detailsDialog.value = true
}

function closeDetails() {
  detailsDialog.value = false
  selectedRecord.value = null
}

function openSettlementRecordsDialog() {
  settlementRecordFilters.search = ''
  settlementRecordFilters.start_at = filters.start_at
  settlementRecordFilters.end_at = filters.end_at
  settlementRecordPagination.page = 1
  settlementRecordsDialog.value = true
  void loadSettlementRecords()
}

function closeSettlementRecordsDialog() {
  settlementRecordsDialog.value = false
}

async function loadSettlementRecords() {
  settlementRecordsLoading.value = true
  try {
    const res = await userAPI.listAffiliateSettlements({
      page: settlementRecordPagination.page,
      page_size: settlementRecordPagination.page_size,
      search: settlementRecordFilters.search.trim() || undefined,
      start_at: settlementRecordFilters.start_at || undefined,
      end_at: settlementRecordFilters.end_at || undefined,
      sort_by: settlementRecordSort.sort_by,
      sort_order: settlementRecordSort.sort_order,
      timezone: userTimezone(),
    })
    settlementRecords.value = res.items || []
    settlementRecordPagination.total = Number(res.total || 0)
    settlementRecordPagination.page = Number(res.page || settlementRecordPagination.page)
    settlementRecordPagination.page_size = Number(res.page_size || settlementRecordPagination.page_size)
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || t('affiliate.usage.settlementRecordsLoadFailed'))
  } finally {
    settlementRecordsLoading.value = false
  }
}

function debounceSettlementRecordsLoad() {
  if (settlementRecordsTimer) clearTimeout(settlementRecordsTimer)
  settlementRecordsTimer = setTimeout(loadSettlementRecordsFromFirstPage, 300)
}

function loadSettlementRecordsFromFirstPage() {
  settlementRecordPagination.page = 1
  void loadSettlementRecords()
}

function handleSettlementRecordPageChange(page: number) {
  settlementRecordPagination.page = page
  void loadSettlementRecords()
}

function handleSettlementRecordPageSizeChange(pageSize: number) {
  settlementRecordPagination.page_size = pageSize
  settlementRecordPagination.page = 1
  void loadSettlementRecords()
}

function toggleSettlementSort(key: string) {
  if (settlementRecordSort.sort_by === key) {
    settlementRecordSort.sort_order = settlementRecordSort.sort_order === 'asc' ? 'desc' : 'asc'
  } else {
    settlementRecordSort.sort_by = key
    settlementRecordSort.sort_order = 'desc'
  }
  loadSettlementRecordsFromFirstPage()
}

function settlementSortIconName(key: string): 'sort' | 'chevronUp' | 'chevronDown' {
  if (settlementRecordSort.sort_by !== key) return 'sort'
  return settlementRecordSort.sort_order === 'asc' ? 'chevronUp' : 'chevronDown'
}

function detailKey(detail: AffiliateUsageProfitDetail): string {
  return `${detail.source || 'usage'}:${detail.group_id}:${detail.model}:${detail.net_profit}:${detail.actual_cost}`
}

function formatProfitDetailActualCost(detail: AffiliateUsageProfitDetail): string {
  return detail.source === 'subscription' ? formatCnyAmount(detail.actual_cost) : formatUsdAmount(detail.actual_cost)
}

function defaultDateInput(offsetDays: number): string {
  const date = new Date()
  date.setDate(date.getDate() + offsetDays)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function userTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  } catch {
    return 'UTC'
  }
}

function formatInteger(value: number | null | undefined): string {
  return Math.round(Number(value || 0)).toLocaleString()
}

function formatTokens(value: number | null | undefined): string {
  const n = Number(value || 0)
  if (n >= 1_000_000) return `${trimNumber(n / 1_000_000)}M`
  if (n >= 1_000) return `${trimNumber(n / 1_000)}K`
  return formatInteger(n)
}

function trimNumber(value: number): string {
  return value.toFixed(value >= 10 ? 1 : 2).replace(/\.0+$/, '').replace(/(\.\d*[1-9])0+$/, '$1')
}

function formatUsdAmount(value: number | null | undefined): string {
  const n = Number(value || 0)
  return `$${Math.abs(n).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 4 })}`
}

function formatCnyAmount(value: number | null | undefined): string {
  const n = Number(value || 0)
  return `${n < 0 ? '-' : ''}\u00a5${Math.abs(n).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 4 })}`
}

function formatPercent(value: number | null | undefined): string {
  return `${trimNumber(Number(value || 0))}%`
}

function formatDateOnly(value: string | null | undefined): string {
  if (!value) return '-'
  return String(value).slice(0, 10)
}

function formatDateTime(value: string | null | undefined): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  return date.toLocaleString()
}

onMounted(() => {
  void loadRecords()
})
</script>

<style scoped>
.summary-tile {
  min-height: 6.25rem;
  border-radius: 0.5rem;
  border: 1px solid rgb(229 231 235);
  background: white;
  padding: 1rem;
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}

.summary-tile p {
  font-size: 0.875rem;
  color: rgb(107 114 128);
}

.summary-tile strong {
  margin-top: 0.5rem;
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  color: rgb(17 24 39);
}

.sort-header {
  display: inline-flex;
  width: 100%;
  align-items: center;
  gap: 0.25rem;
  border: 0;
  background: transparent;
  padding: 0;
  font: inherit;
  color: inherit;
  transition: color 0.15s ease;
}

.sort-header:hover {
  color: rgb(17 24 39);
}

:global(.dark) .summary-tile {
  border-color: rgb(51 65 85);
  background: rgb(30 41 59);
}

:global(.dark) .summary-tile p {
  color: rgb(148 163 184);
}

:global(.dark) .summary-tile strong {
  color: white;
}

:global(.dark) .sort-header:hover {
  color: white;
}
</style>
