<template>
  <div class="card p-4">
    <div class="mb-4">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.dashboard.userUsageHierarchyTitle') }}</h3>
      <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.dashboard.userUsageHierarchyDescription') }}</p>
    </div>
    <div v-if="loading" class="flex h-40 items-center justify-center"><LoadingSpinner /></div>
    <div v-else-if="users.length" class="overflow-x-auto">
      <table class="w-full min-w-[920px] text-sm">
        <thead><tr class="border-b border-gray-200 text-left text-xs text-gray-500 dark:border-gray-700 dark:text-gray-400">
          <th class="px-3 py-2">{{ t('admin.dashboard.user') }}</th><th class="px-3 py-2">{{ t('admin.dashboard.usage') }}</th><th class="px-3 py-2">{{ t('admin.dashboard.requests') }}</th><th class="px-3 py-2">{{ t('admin.dashboard.tokenBreakdown') }}</th><th class="px-3 py-2">{{ t('admin.dashboard.actual') }}</th>
        </tr></thead>
        <tbody>
          <template v-for="user in users" :key="user.user_id">
            <tr class="cursor-pointer border-b border-gray-100 hover:bg-gray-50 dark:border-gray-800 dark:hover:bg-gray-800/50" @click="toggleUser(user.user_id)">
              <td class="px-3 py-3 font-medium text-gray-900 dark:text-white"><span class="mr-2 text-gray-400">{{ isUserExpanded(user.user_id) ? '▾' : '▸' }}</span>{{ user.username || user.email || `#${user.user_id}` }}<span v-if="user.username && user.email" class="ml-2 text-xs font-normal text-gray-400">{{ user.email }}</span></td>
              <td class="px-3 py-3 font-semibold">{{ formatTokens(user.total_tokens) }}</td><td class="px-3 py-3">{{ formatNumber(user.requests) }}</td><td class="px-3 py-3 text-xs text-gray-500">{{ tokenSummary(user) }}</td><td class="px-3 py-3 text-primary-600">${{ formatCost(user.actual_cost) }}</td>
            </tr>
            <tr v-if="isUserExpanded(user.user_id)"><td colspan="5" class="bg-gray-50 px-3 py-3 dark:bg-gray-900/40">
              <div v-for="group in user.groups" :key="`${user.user_id}-${group.group_id}`" class="mb-2 rounded border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800">
                <button class="flex w-full items-center justify-between px-3 py-2 text-left" @click.stop="toggleGroup(user.user_id, group.group_id)"><span><span class="mr-2 text-gray-400">{{ isGroupExpanded(user.user_id, group.group_id) ? '▾' : '▸' }}</span><span class="font-medium">{{ group.group_name }}</span></span><span class="text-xs text-gray-500">{{ formatTokens(group.total_tokens) }} · {{ group.requests }} {{ t('admin.dashboard.requests') }} · ${{ formatCost(group.actual_cost) }}</span></button>
                <div v-if="isGroupExpanded(user.user_id, group.group_id)" class="border-t border-gray-100 px-3 pb-2 dark:border-gray-700">
                  <div v-for="model in group.models" :key="model.model" class="grid grid-cols-[minmax(180px,1fr)_90px_100px_1fr_90px] gap-2 border-b border-gray-100 py-2 text-xs last:border-0 dark:border-gray-700"><span class="truncate font-medium" :title="model.model">{{ model.model }}</span><span>{{ model.requests }} {{ t('admin.dashboard.requests') }}</span><span>{{ formatTokens(model.total_tokens) }}</span><span class="text-gray-500">{{ tokenSummary(model) }}</span><span class="text-primary-600">${{ formatCost(model.actual_cost) }}</span></div>
                </div>
              </div>
            </td></tr>
          </template>
        </tbody>
      </table>
    </div>
    <div v-else class="flex h-40 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('admin.dashboard.noDataAvailable') }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { UserUsageHierarchyItem } from '@/types'

defineProps<{ users: UserUsageHierarchyItem[]; loading?: boolean }>()
const { t } = useI18n()
const expandedUsers = ref(new Set<number>())
const expandedGroups = ref(new Set<string>())
const isUserExpanded = (id: number) => expandedUsers.value.has(id)
const isGroupExpanded = (userId: number, groupId: number) => expandedGroups.value.has(`${userId}:${groupId}`)
const toggleUser = (id: number) => { const next = new Set(expandedUsers.value); next.has(id) ? next.delete(id) : next.add(id); expandedUsers.value = next }
const toggleGroup = (userId: number, groupId: number) => { const key = `${userId}:${groupId}`; const next = new Set(expandedGroups.value); next.has(key) ? next.delete(key) : next.add(key); expandedGroups.value = next }
const formatTokens = (value: number) => value >= 1_000_000_000 ? `${(value / 1_000_000_000).toFixed(2)}B` : value >= 1_000_000 ? `${(value / 1_000_000).toFixed(2)}M` : value >= 1_000 ? `${(value / 1_000).toFixed(2)}K` : value.toLocaleString()
const formatNumber = (value: number) => value.toLocaleString()
const formatCost = (value: number) => value >= 1 ? value.toFixed(2) : value.toFixed(4)
const tokenSummary = (item: { input_tokens: number; output_tokens: number; cache_creation_tokens: number; cache_read_tokens: number }) => `I ${formatTokens(item.input_tokens)} · O ${formatTokens(item.output_tokens)} · CR ${formatTokens(item.cache_read_tokens)} · CW ${formatTokens(item.cache_creation_tokens)}`
</script>
