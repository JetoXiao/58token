<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(360px,420px)]">
        <div class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.trialCards.settingsTitle') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('admin.trialCards.settingsDescription') }}
            </p>
          </div>
          <form class="space-y-5 p-6" @submit.prevent="saveSettings">
            <label class="flex items-start gap-3">
              <input
                v-model="settingsForm.invitation_enabled"
                type="checkbox"
                class="mt-0.5 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
              <span>
                <span class="block text-sm font-medium text-gray-900 dark:text-white">
                  {{ t('admin.trialCards.invitationEnabled') }}
                </span>
                <span class="block text-xs text-gray-500 dark:text-dark-400">
                  {{ t('admin.trialCards.invitationEnabledHint') }}
                </span>
              </span>
            </label>

            <div>
              <label class="input-label">{{ t('admin.trialCards.invitationAmount') }}</label>
              <input
                v-model.number="settingsForm.invitation_amount"
                type="number"
                step="0.01"
                min="0"
                class="input"
              />
            </div>

            <div>
              <label class="input-label">{{ t('admin.trialCards.freeGroups') }}</label>
              <div class="mt-2 grid gap-2 sm:grid-cols-2">
                <label
                  v-for="group in groups"
                  :key="group.id"
                  class="flex items-center gap-2 rounded-lg border border-gray-200 px-3 py-2 text-sm dark:border-dark-700"
                >
                  <input
                    v-model="settingsForm.group_ids"
                    type="checkbox"
                    :value="group.id"
                    class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                  />
                  <span class="truncate text-gray-700 dark:text-gray-200">{{ group.name }}</span>
                </label>
              </div>
              <p v-if="groups.length === 0" class="input-hint">{{ t('admin.trialCards.noGroups') }}</p>
            </div>

            <div class="space-y-3">
              <label class="flex items-start gap-3">
                <input
                  v-model="settingsForm.show_locked_groups"
                  type="checkbox"
                  class="mt-0.5 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                />
                <span>
                  <span class="block text-sm font-medium text-gray-900 dark:text-white">
                    {{ t('admin.trialCards.showLockedGroups') }}
                  </span>
                  <span class="block text-xs text-gray-500 dark:text-dark-400">
                    {{ t('admin.trialCards.showLockedGroupsHint') }}
                  </span>
                </span>
              </label>
              <label class="flex items-start gap-3">
                <input
                  v-model="settingsForm.transfer_on_payment"
                  type="checkbox"
                  class="mt-0.5 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                />
                <span>
                  <span class="block text-sm font-medium text-gray-900 dark:text-white">
                    {{ t('admin.trialCards.transferOnPayment') }}
                  </span>
                  <span class="block text-xs text-gray-500 dark:text-dark-400">
                    {{ t('admin.trialCards.transferOnPaymentHint') }}
                  </span>
                </span>
              </label>
            </div>

            <div class="flex justify-end">
              <button type="submit" class="btn btn-primary" :disabled="savingSettings">
                <Icon name="check" size="md" class="mr-2" />
                {{ savingSettings ? t('common.submitting') : t('common.save') }}
              </button>
            </div>
          </form>
        </div>

        <div class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.trialCards.createTitle') }}
            </h2>
          </div>
          <form class="space-y-4 p-6" @submit.prevent="createCard">
            <div>
              <label class="input-label">{{ t('admin.trialCards.code') }}</label>
              <input v-model.trim="createForm.code" type="text" required class="input font-mono" />
            </div>
            <div>
              <label class="input-label">{{ t('common.name') }}</label>
              <input v-model.trim="createForm.name" type="text" class="input" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label">{{ t('admin.trialCards.amount') }}</label>
                <input v-model.number="createForm.amount" type="number" step="0.01" min="0.01" required class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.trialCards.maxRedemptions') }}</label>
                <input v-model.number="createForm.max_redemptions" type="number" min="1" required class="input" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label">{{ t('admin.trialCards.perUserLimit') }}</label>
                <input v-model.number="createForm.per_user_limit" type="number" min="1" class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('common.status') }}</label>
                <select v-model="createForm.status" class="input">
                  <option value="active">{{ t('common.active') }}</option>
                  <option value="inactive">{{ t('common.inactive') }}</option>
                </select>
              </div>
            </div>
            <div>
              <label class="input-label">{{ t('admin.trialCards.expiresAt') }}</label>
              <select v-model="createForm.expires_in_days" class="input">
                <option :value="1">{{ t('admin.trialCards.expiryOneDay') }}</option>
                <option :value="3">{{ t('admin.trialCards.expiryThreeDays') }}</option>
                <option :value="7">{{ t('admin.trialCards.expirySevenDays') }}</option>
                <option :value="null">{{ t('admin.trialCards.neverExpires') }}</option>
              </select>
            </div>
            <div>
              <label class="input-label">{{ t('admin.trialCards.notes') }}</label>
              <textarea v-model.trim="createForm.notes" rows="2" class="input"></textarea>
            </div>
            <button type="submit" class="btn btn-primary w-full" :disabled="creating">
              <Icon name="plus" size="md" class="mr-2" />
              {{ creating ? t('common.submitting') : t('common.create') }}
            </button>
          </form>
        </div>
      </div>

      <TablePageLayout>
        <template #filters>
          <div class="flex justify-end">
            <button class="btn btn-secondary" :disabled="loading" @click="loadCards">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </template>

        <template #table>
          <DataTable :columns="columns" :data="cards" :loading="loading">
            <template #cell-code="{ value }">
              <code class="font-mono text-sm text-gray-900 dark:text-gray-100">{{ value }}</code>
            </template>
            <template #cell-amount="{ value }">
              <span class="font-medium">${{ Number(value || 0).toFixed(2) }}</span>
            </template>
            <template #cell-redemptions="{ row }">
              <span>{{ row.redeemed_count }} / {{ row.max_redemptions }}</span>
            </template>
            <template #cell-remaining_redemptions="{ row }">
              <span>{{ Math.max(row.max_redemptions - row.redeemed_count, 0) }}</span>
            </template>
            <template #cell-status="{ value }">
              <span :class="['badge', value === 'active' ? 'badge-success' : 'badge-gray']">
                {{ value === 'active' ? t('common.active') : t('common.inactive') }}
              </span>
            </template>
            <template #cell-expires_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">
                {{ value ? formatDateTime(value) : t('admin.trialCards.neverExpires') }}
              </span>
            </template>
            <template #cell-created_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">
                {{ formatDateTime(value) }}
              </span>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center space-x-1">
                <button
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600"
                  :title="t('common.edit')"
                  @click="startEdit(row)"
                >
                  <Icon name="edit" size="sm" />
                </button>
                <button
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
                  :title="t('admin.trialCards.deactivate')"
                  @click="deleteCard(row)"
                >
                  <Icon name="ban" size="sm" />
                </button>
              </div>
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
    </div>

    <BaseDialog
      :show="!!editingCard"
      :title="t('admin.trialCards.editTitle')"
      width="wide"
      @close="editingCard = null"
    >
      <form v-if="editingCard" id="edit-trial-card-form" class="space-y-4" @submit.prevent="updateCard">
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('common.name') }}</label>
            <input v-model.trim="editForm.name" type="text" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('common.status') }}</label>
            <select v-model="editForm.status" class="input">
              <option value="active">{{ t('common.active') }}</option>
              <option value="inactive">{{ t('common.inactive') }}</option>
            </select>
          </div>
          <div>
            <label class="input-label">{{ t('admin.trialCards.amount') }}</label>
            <input v-model.number="editForm.amount" type="number" step="0.01" min="0.01" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.trialCards.maxRedemptions') }}</label>
            <input v-model.number="editForm.max_redemptions" type="number" min="1" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.trialCards.perUserLimit') }}</label>
            <input v-model.number="editForm.per_user_limit" type="number" min="1" class="input" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.trialCards.expiresAt') }}</label>
            <select v-model="editForm.expires_in_days" class="input">
              <option :value="1">{{ t('admin.trialCards.expiryOneDay') }}</option>
              <option :value="3">{{ t('admin.trialCards.expiryThreeDays') }}</option>
              <option :value="7">{{ t('admin.trialCards.expirySevenDays') }}</option>
              <option :value="null">{{ t('admin.trialCards.neverExpires') }}</option>
            </select>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('admin.trialCards.notes') }}</label>
          <textarea v-model.trim="editForm.notes" rows="3" class="input"></textarea>
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="editingCard = null">
          {{ t('common.cancel') }}
        </button>
        <button type="submit" form="edit-trial-card-form" class="btn btn-primary" :disabled="updating">
          {{ updating ? t('common.submitting') : t('common.save') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import type { AdminGroup, FreeQuotaSettings, TrialCard } from '@/types'

interface Column {
  key: string
  label: string
}

interface TrialCardForm {
  code: string
  name: string
  amount: number
  max_redemptions: number
  per_user_limit: number
  status: string
  notes: string
  expires_in_days: number | null
}

const { t } = useI18n()
const appStore = useAppStore()

const groups = ref<AdminGroup[]>([])
const cards = ref<TrialCard[]>([])
const loading = ref(false)
const creating = ref(false)
const updating = ref(false)
const savingSettings = ref(false)
const editingCard = ref<TrialCard | null>(null)

const settingsForm = reactive<FreeQuotaSettings>({
  invitation_enabled: false,
  invitation_amount: 0,
  group_ids: [],
  show_locked_groups: false,
  transfer_on_payment: false
})

const createForm = reactive<TrialCardForm>({
  code: '',
  name: '',
  amount: 1,
  max_redemptions: 20,
  per_user_limit: 1,
  status: 'active',
  notes: '',
  expires_in_days: null
})

const editForm = reactive<Omit<TrialCardForm, 'code'>>({
  name: '',
  amount: 1,
  max_redemptions: 1,
  per_user_limit: 1,
  status: 'active',
  notes: '',
  expires_in_days: null
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})

const columns = computed<Column[]>(() => [
  { key: 'code', label: t('admin.trialCards.code') },
  { key: 'name', label: t('common.name') },
  { key: 'amount', label: t('admin.trialCards.amount') },
  { key: 'redemptions', label: t('admin.trialCards.redemptions') },
  { key: 'remaining_redemptions', label: t('admin.trialCards.remainingRedemptions') },
  { key: 'per_user_limit', label: t('admin.trialCards.perUserLimit') },
  { key: 'status', label: t('common.status') },
  { key: 'expires_at', label: t('admin.trialCards.expiresAt') },
  { key: 'created_at', label: t('admin.trialCards.createdAt') },
  { key: 'actions', label: t('common.actions') }
])

const toApiDate = (days: number | null): string | null => {
  if (!days) return null
  const date = new Date()
  date.setDate(date.getDate() + days)
  return date.toISOString()
}

const toExpiryPreset = (value?: string | null): number | null => {
  if (!value) return null
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return null
  const days = Math.ceil((date.getTime() - Date.now()) / 86_400_000)
  if (days <= 0) return 1
  if (days <= 1) return 1
  if (days <= 3) return 3
  return 7
}

const loadGroups = async () => {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch (error) {
    console.error('Failed to load groups:', error)
    appStore.showError(t('admin.trialCards.loadGroupsFailed'))
  }
}

const loadSettings = async () => {
  try {
    const settings = await adminAPI.trialCards.getSettings()
    settingsForm.invitation_enabled = !!settings.invitation_enabled
    settingsForm.invitation_amount = settings.invitation_amount
    settingsForm.group_ids = [...(settings.group_ids || [])]
    settingsForm.show_locked_groups = settings.show_locked_groups
    settingsForm.transfer_on_payment = settings.transfer_on_payment
  } catch (error) {
    console.error('Failed to load free quota settings:', error)
    appStore.showError(t('admin.trialCards.loadSettingsFailed'))
  }
}

const loadCards = async () => {
  loading.value = true
  try {
    const response = await adminAPI.trialCards.list(pagination.page, pagination.page_size)
    cards.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error) {
    console.error('Failed to load trial cards:', error)
    appStore.showError(t('admin.trialCards.loadCardsFailed'))
  } finally {
    loading.value = false
  }
}

const saveSettings = async () => {
  savingSettings.value = true
  try {
    await adminAPI.trialCards.updateSettings({
      invitation_enabled: settingsForm.invitation_enabled,
      invitation_amount: Number(settingsForm.invitation_amount || 0),
      group_ids: [...settingsForm.group_ids],
      show_locked_groups: settingsForm.show_locked_groups,
      transfer_on_payment: settingsForm.transfer_on_payment
    })
    appStore.showSuccess(t('common.saved'))
  } catch (error) {
    console.error('Failed to save free quota settings:', error)
    appStore.showError(t('admin.trialCards.saveSettingsFailed'))
  } finally {
    savingSettings.value = false
  }
}

const resetCreateForm = () => {
  createForm.code = ''
  createForm.name = ''
  createForm.amount = Number(settingsForm.invitation_amount || 1)
  createForm.max_redemptions = 20
  createForm.per_user_limit = 1
  createForm.status = 'active'
  createForm.notes = ''
  createForm.expires_in_days = null
}

const createCard = async () => {
  if (!settingsForm.group_ids.length) {
    appStore.showError(t('admin.trialCards.groupsRequired'))
    return
  }
  creating.value = true
  try {
    await adminAPI.trialCards.create({
      code: createForm.code.trim(),
      name: createForm.name.trim(),
      amount: Number(createForm.amount),
      max_redemptions: Number(createForm.max_redemptions),
      per_user_limit: Number(createForm.per_user_limit || 1),
      status: createForm.status,
      notes: createForm.notes.trim(),
      expires_at: toApiDate(createForm.expires_in_days)
    })
    appStore.showSuccess(t('admin.trialCards.createSuccess'))
    resetCreateForm()
    await loadCards()
  } catch (error: any) {
    console.error('Failed to create trial card:', error)
    appStore.showError(error.message || t('admin.trialCards.createFailed'))
  } finally {
    creating.value = false
  }
}

const startEdit = (card: TrialCard) => {
  editingCard.value = card
  editForm.name = card.name || ''
  editForm.amount = card.amount
  editForm.max_redemptions = card.max_redemptions
  editForm.per_user_limit = card.per_user_limit || 1
  editForm.status = card.status || 'active'
  editForm.notes = card.notes || ''
  editForm.expires_in_days = toExpiryPreset(card.expires_at)
}

const updateCard = async () => {
  if (!editingCard.value) return
  updating.value = true
  try {
    await adminAPI.trialCards.update(editingCard.value.id, {
      name: editForm.name.trim(),
      amount: Number(editForm.amount),
      max_redemptions: Number(editForm.max_redemptions),
      per_user_limit: Number(editForm.per_user_limit || 1),
      status: editForm.status,
      notes: editForm.notes.trim(),
      expires_at: toApiDate(editForm.expires_in_days),
      clear_expires_at: !editForm.expires_in_days
    })
    editingCard.value = null
    appStore.showSuccess(t('common.saved'))
    await loadCards()
  } catch (error: any) {
    console.error('Failed to update trial card:', error)
    appStore.showError(error.message || t('admin.trialCards.updateFailed'))
  } finally {
    updating.value = false
  }
}

const deleteCard = async (card: TrialCard) => {
  if (!window.confirm(t('admin.trialCards.deleteConfirm', { code: card.code }))) return
  try {
    await adminAPI.trialCards.delete(card.id)
    appStore.showSuccess(t('admin.trialCards.deactivateSuccess'))
    await loadCards()
  } catch (error: any) {
    console.error('Failed to deactivate trial card:', error)
    appStore.showError(error.message || t('admin.trialCards.deleteFailed'))
  }
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadCards()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadCards()
}

onMounted(async () => {
  await Promise.all([loadGroups(), loadSettings(), loadCards()])
  resetCreateForm()
})
</script>
