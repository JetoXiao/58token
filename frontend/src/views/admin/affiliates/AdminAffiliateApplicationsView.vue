<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="relative w-full md:w-80">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="filters.search"
              type="text"
              class="input pl-10"
              :placeholder="t('admin.affiliates.applications.searchPlaceholder')"
              @input="debounceLoad"
            />
          </div>
          <select v-model="filters.status" class="input w-full sm:w-40" @change="reloadFromFirstPage">
            <option value="">{{ t('admin.affiliates.applications.allStatus') }}</option>
            <option value="pending">{{ t('affiliate.partner.status.pending') }}</option>
            <option value="approved">{{ t('affiliate.partner.status.approved') }}</option>
            <option value="rejected">{{ t('affiliate.partner.status.rejected') }}</option>
          </select>
          <button class="btn btn-secondary px-2 md:px-3" :disabled="loading" :title="t('common.refresh')" @click="loadApplications">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="applications" :loading="loading">
          <template #cell-user="{ row }">
            <div class="space-y-0.5">
              <div class="font-mono text-xs text-gray-500 dark:text-dark-400">#{{ row.user_id }}</div>
              <div class="text-sm font-medium text-gray-900 dark:text-white">{{ row.email || '-' }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ row.username || '-' }}</div>
            </div>
          </template>
          <template #cell-current_level="{ row }">
            <span :class="['badge', row.current_level && row.current_level !== 'none' ? 'badge-purple' : 'badge-gray']">
              {{ partnerLevelLabel(row.current_level || 'none') }}
            </span>
          </template>
          <template #cell-source="{ row }">
            <div class="space-y-1">
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ sourceLabel(row.source) }}</span>
              <a
                :href="row.portal_url"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex max-w-56 items-center gap-1 truncate text-xs text-primary-600 hover:underline dark:text-primary-400"
              >
                <span class="truncate">{{ row.portal_url }}</span>
                <Icon name="externalLink" size="xs" />
              </a>
            </div>
          </template>
          <template #cell-strengths="{ row }">
            <p class="max-w-md whitespace-pre-line text-sm text-gray-700 dark:text-gray-300">{{ row.strengths }}</p>
          </template>
          <template #cell-status="{ row }">
            <span :class="['badge', applicationBadgeClass(row.status)]">
              {{ t(`affiliate.partner.status.${row.status}`) }}
            </span>
          </template>
          <template #cell-created_at="{ row }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(row.created_at) }}</span>
          </template>
          <template #cell-actions="{ row }">
            <button
              v-if="row.status === 'pending'"
              class="btn btn-primary btn-sm"
              type="button"
              @click="openReviewDialog(row)"
            >
              <Icon name="check" size="sm" />
              <span>{{ t('admin.affiliates.applications.review') }}</span>
            </button>
            <span v-else class="text-sm text-gray-400 dark:text-dark-500">{{ row.review_note || '-' }}</span>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="reviewDialog"
      :title="t('admin.affiliates.applications.reviewTitle')"
      width="normal"
      @close="closeReviewDialog"
    >
      <div v-if="selectedApplication" class="space-y-5">
        <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800">
          <div class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedApplication.email || `#${selectedApplication.user_id}` }}</div>
          <div class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ sourceLabel(selectedApplication.source) }}</div>
          <p class="mt-3 whitespace-pre-line text-sm text-gray-700 dark:text-gray-300">{{ selectedApplication.strengths }}</p>
        </div>

        <div>
          <label class="input-label">{{ t('admin.affiliates.applications.reviewResult') }}</label>
          <div class="inline-flex rounded-lg border border-gray-200 bg-white p-1 dark:border-dark-700 dark:bg-dark-800">
            <button
              type="button"
              class="rounded-md px-3 py-1.5 text-sm font-medium transition"
              :class="reviewForm.status === 'approved' ? 'bg-primary-600 text-white' : 'text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-700'"
              @click="reviewForm.status = 'approved'"
            >
              {{ t('affiliate.partner.status.approved') }}
            </button>
            <button
              type="button"
              class="rounded-md px-3 py-1.5 text-sm font-medium transition"
              :class="reviewForm.status === 'rejected' ? 'bg-red-600 text-white' : 'text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-700'"
              @click="reviewForm.status = 'rejected'"
            >
              {{ t('affiliate.partner.status.rejected') }}
            </button>
          </div>
        </div>

        <div v-if="reviewForm.status === 'approved'">
          <label class="input-label">{{ t('admin.affiliates.applications.grantedLevel') }}</label>
          <select v-model="reviewForm.granted_level" class="input">
            <option v-for="tier in partnerTiers" :key="tier.level" :value="tier.level">
              {{ partnerLevelLabel(tier.level) }} · {{ tier.rebate_rate_percent }}%
            </option>
          </select>
        </div>

        <div>
          <label class="input-label">{{ t('admin.affiliates.applications.reviewNote') }}</label>
          <textarea v-model.trim="reviewForm.review_note" rows="3" class="input resize-none"></textarea>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeReviewDialog">{{ t('common.cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="reviewing" @click="submitReview">
            <Icon v-if="reviewing" name="refresh" size="sm" class="animate-spin" />
            <span>{{ t('common.confirm') }}</span>
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores/app'
import {
  affiliatesAPI,
  type ListPartnerApplicationsParams,
} from '@/api/admin/affiliates'
import type {
  AffiliatePartnerApplication,
  AffiliatePartnerApplicationStatus,
  AffiliatePartnerLevel,
  AffiliatePartnerTier,
} from '@/types'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'

type ReviewStatus = Exclude<AffiliatePartnerApplicationStatus, 'pending'>
type GrantableLevel = Exclude<AffiliatePartnerLevel, 'none'>

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const reviewing = ref(false)
const applications = ref<AffiliatePartnerApplication[]>([])
const partnerTiers = ref<AffiliatePartnerTier[]>([])
const filters = reactive({ search: '', status: '' as ListPartnerApplicationsParams['status'] })
const pagination = reactive({ page: 1, page_size: 20, total: 0, pages: 1 })
const reviewDialog = ref(false)
const selectedApplication = ref<AffiliatePartnerApplication | null>(null)
const reviewForm = reactive({
  status: 'approved' as ReviewStatus,
  granted_level: 'spark' as GrantableLevel,
  review_note: '',
})
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const columns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.affiliates.records.user'), sortable: false },
  { key: 'current_level', label: t('admin.affiliates.applications.currentLevel'), sortable: false },
  { key: 'source', label: t('admin.affiliates.applications.source'), sortable: false },
  { key: 'strengths', label: t('admin.affiliates.applications.strengths'), sortable: false },
  { key: 'status', label: t('admin.affiliates.applications.status'), sortable: false },
  { key: 'created_at', label: t('admin.affiliates.applications.createdAt'), sortable: false },
  { key: 'actions', label: t('admin.affiliates.applications.actions'), sortable: false },
])

function partnerLevelLabel(level?: AffiliatePartnerLevel | ''): string {
  return t(`affiliate.partner.levels.${level || 'none'}`)
}

function sourceLabel(source: string): string {
  return t(`affiliate.partner.sources.${source}`, source)
}

function applicationBadgeClass(status: AffiliatePartnerApplicationStatus): string {
  if (status === 'approved') return 'badge-success'
  if (status === 'rejected') return 'badge-danger'
  return 'badge-warning'
}

function buildParams(): ListPartnerApplicationsParams {
  return {
    page: pagination.page,
    page_size: pagination.page_size,
    search: filters.search,
    status: filters.status,
  }
}

async function loadApplications(): Promise<void> {
  loading.value = true
  try {
    const res = await affiliatesAPI.listPartnerApplications(buildParams())
    applications.value = res.items
    pagination.total = res.total
    pagination.pages = res.pages
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('admin.affiliates.errors.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function loadPartnerTiers(): Promise<void> {
  try {
    partnerTiers.value = await affiliatesAPI.listPartnerTiers()
    if (partnerTiers.value.length > 0) {
      reviewForm.granted_level = partnerTiers.value[0].level as GrantableLevel
    }
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('admin.affiliates.errors.loadFailed')))
  }
}

function reloadFromFirstPage(): void {
  pagination.page = 1
  void loadApplications()
}

function debounceLoad(): void {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(reloadFromFirstPage, 300)
}

function handlePageChange(page: number): void {
  pagination.page = Math.max(1, Math.min(page, pagination.pages || 1))
  void loadApplications()
}

function handlePageSizeChange(pageSize: number): void {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadApplications()
}

function openReviewDialog(app: AffiliatePartnerApplication): void {
  selectedApplication.value = app
  reviewForm.status = 'approved'
  reviewForm.granted_level = partnerTiers.value[0]?.level as GrantableLevel || 'spark'
  reviewForm.review_note = ''
  reviewDialog.value = true
}

function closeReviewDialog(): void {
  reviewDialog.value = false
  selectedApplication.value = null
}

async function submitReview(): Promise<void> {
  if (!selectedApplication.value || reviewing.value) return
  reviewing.value = true
  try {
    await affiliatesAPI.reviewPartnerApplication(selectedApplication.value.id, {
      status: reviewForm.status,
      granted_level: reviewForm.status === 'approved' ? reviewForm.granted_level : undefined,
      review_note: reviewForm.review_note || undefined,
    })
    appStore.showSuccess(t('admin.affiliates.applications.reviewSuccess'))
    closeReviewDialog()
    await loadApplications()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('admin.affiliates.applications.reviewFailed')))
  } finally {
    reviewing.value = false
  }
}

onMounted(() => {
  void loadPartnerTiers()
  void loadApplications()
})
</script>
