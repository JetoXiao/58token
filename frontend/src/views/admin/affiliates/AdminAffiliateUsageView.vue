<template>
  <AppLayout>
    <TablePageLayout class="affiliate-usage-layout">
      <template #actions>
        <div
          class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4"
          :class="viewMode === 'groups' ? 'xl:grid-cols-5' : 'xl:grid-cols-4'"
        >
          <SummaryStat
            :label="t('admin.affiliates.usage.summaryRequests')"
            :value="formatInteger(summary.total_requests)"
            icon="chart"
          />
          <SummaryStat
            :label="t('admin.affiliates.usage.summaryTokens')"
            :value="formatTokens(summary.total_tokens)"
            icon="database"
          />
          <SummaryStat
            :label="t('admin.affiliates.usage.summaryCost')"
            :value="formatCost(summary.total_actual_cost)"
            icon="creditCard"
          />
          <SummaryStat
            :label="usageLabel('summaryNetProfit')"
            :value="formatCost(summary.total_net_profit)"
            icon="chart"
          />
          <SummaryStat
            v-if="viewMode === 'groups'"
            :label="t('admin.affiliates.usage.summaryRebate')"
            :value="formatRebateAmount(summary.total_rebate_amount)"
            icon="creditCard"
          />
        </div>
      </template>

      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="inline-flex h-11 rounded-lg border border-gray-200 bg-white p-1 dark:border-dark-700 dark:bg-dark-800">
            <button
              type="button"
              class="rounded-md px-3 text-sm font-medium transition"
              :class="viewMode === 'users' ? 'bg-primary-600 text-white shadow-sm' : 'text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-700'"
              @click="setViewMode('users')"
            >
              {{ t('admin.affiliates.usage.viewUsers') }}
            </button>
            <button
              type="button"
              class="rounded-md px-3 text-sm font-medium transition"
              :class="viewMode === 'groups' ? 'bg-primary-600 text-white shadow-sm' : 'text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-700'"
              @click="setViewMode('groups')"
            >
              {{ t('admin.affiliates.usage.viewGroups') }}
            </button>
          </div>

          <div class="relative w-full md:w-72">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="filters.search"
              type="text"
              class="input pl-10"
              :placeholder="t('admin.affiliates.usage.searchPlaceholder')"
              @input="debounceLoad"
            />
          </div>

          <div class="relative w-full sm:w-64">
            <input
              v-model="inviterQuery"
              type="text"
              class="input pr-10"
              :placeholder="t('admin.affiliates.usage.inviterFilter')"
              @input="onFilterUserInput('inviter')"
              @focus="onFilterUserInput('inviter')"
              @keydown.enter.prevent="applyTypedFilterUser('inviter')"
              @blur="hideFilterResultsLater('inviter')"
            />
            <button
              v-if="filters.inviter_id"
              type="button"
              class="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-200"
              :title="t('common.reset')"
              @click="clearFilterUser('inviter')"
            >
              <Icon name="x" size="sm" />
            </button>
            <UserSearchResults
              v-if="showInviterResults"
              :users="filterInviterResults"
              :loading="filterInviterLoading"
              @select="selectFilterUser('inviter', $event)"
            />
          </div>

          <div class="relative w-full sm:w-64">
            <input
              v-model="inviteeQuery"
              type="text"
              class="input pr-10"
              :placeholder="t('admin.affiliates.usage.inviteeFilter')"
              @input="onFilterUserInput('invitee')"
              @focus="onFilterUserInput('invitee')"
              @keydown.enter.prevent="applyTypedFilterUser('invitee')"
              @blur="hideFilterResultsLater('invitee')"
            />
            <button
              v-if="filters.invitee_id"
              type="button"
              class="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-200"
              :title="t('common.reset')"
              @click="clearFilterUser('invitee')"
            >
              <Icon name="x" size="sm" />
            </button>
            <UserSearchResults
              v-if="showInviteeResults"
              :users="filterInviteeResults"
              :loading="filterInviteeLoading"
              @select="selectFilterUser('invitee', $event)"
            />
          </div>

          <input
            v-model="filters.start_at"
            type="date"
            class="input w-full sm:w-44"
            :title="t('admin.affiliates.records.startAt')"
            @change="reloadFromFirstPage"
          />
          <input
            v-model="filters.end_at"
            type="date"
            class="input w-full sm:w-44"
            :title="t('admin.affiliates.records.endAt')"
            @change="reloadFromFirstPage"
          />

          <button class="btn btn-secondary px-2 md:px-3" :disabled="loading" :title="t('common.refresh')" @click="loadRecords">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button class="btn btn-primary" type="button" @click="openAssignDialog">
            <Icon name="userPlus" size="md" />
            <span>{{ t('admin.affiliates.usage.assignButton') }}</span>
          </button>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="records"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="actual_cost"
          default-sort-order="desc"
          sort-storage-key="admin-affiliate-usage-v2-table-sort"
          @sort="handleSort"
        >
          <template #cell-user="{ row }">
            <UserCell
              :id="row.invitee_id"
              :email="row.invitee_email"
              :username="row.invitee_username"
              clickable
              @click="openAssignDialogForRecord(row)"
            />
          </template>
          <template #cell-inviter="{ row }">
            <UnassignedCell v-if="row.unassigned" />
            <UserCell v-else :id="row.inviter_id" :email="row.inviter_email" :username="row.inviter_username" />
          </template>
          <template #cell-invitee_count="{ row }">
            <button
              type="button"
              class="inline-flex h-9 items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 text-sm font-medium text-gray-700 shadow-sm transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200 dark:hover:border-primary-700 dark:hover:bg-primary-900/20 dark:hover:text-primary-300 dark:focus:ring-offset-dark-900"
              @click.stop="openMembersDialog(row)"
            >
              <Icon name="users" size="sm" />
              <span>{{ t('admin.affiliates.usage.showMembers', { count: groupMemberCount(row) }) }}</span>
            </button>
          </template>
          <template #cell-requests="{ row }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ formatInteger(row.requests) }}</span>
          </template>
          <template #cell-total_tokens="{ row }">
            <span class="text-sm font-medium text-gray-900 dark:text-white">{{ formatTokens(row.total_tokens) }}</span>
          </template>
          <template #cell-actual_cost="{ row }">
            <span class="text-sm font-semibold text-emerald-600 dark:text-emerald-400">{{ formatCost(row.actual_cost) }}</span>
          </template>
          <template #cell-net_profit="{ row }">
            <button
              type="button"
              class="inline-flex h-9 items-center gap-2 rounded-lg border border-indigo-100 bg-indigo-50 px-3 text-sm font-semibold text-indigo-700 transition hover:border-indigo-200 hover:bg-indigo-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:border-indigo-900/50 dark:bg-indigo-900/20 dark:text-indigo-300 dark:hover:border-indigo-700 dark:hover:bg-indigo-900/30 dark:focus:ring-offset-dark-900"
              @click.stop="openProfitDetailsDialog(row)"
            >
              <Icon name="eye" size="sm" />
              <span>{{ formatCost(row.net_profit) }}</span>
            </button>
          </template>
          <template #cell-recharge_amount="{ row }">
            <span class="text-sm font-semibold text-sky-600 dark:text-sky-400">{{ formatRechargeAmount(row.recharge_amount) }}</span>
          </template>
          <template #cell-rebate_rate="{ row }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ row.unassigned ? '-' : formatPercent(row.rebate_rate_percent) }}</span>
          </template>
          <template #cell-rebate_amount="{ row }">
            <span class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatRebateAmount(row.rebate_amount) }}</span>
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
      :show="membersDialog"
      :title="t('admin.affiliates.usage.groupMembers')"
      width="extra-wide"
      @close="closeMembersDialog"
    >
      <div v-if="selectedMemberGroup" class="space-y-4">
        <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/60">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="min-w-0 space-y-2">
              <div class="text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                {{ t('admin.affiliates.records.inviter') }}
              </div>
              <UnassignedCell v-if="selectedMemberGroup.unassigned" />
              <UserCell
                v-else
                :id="selectedMemberGroup.inviter_id"
                :email="selectedMemberGroup.inviter_email"
                :username="selectedMemberGroup.inviter_username"
              />
            </div>
            <div class="grid grid-cols-2 gap-3 text-right sm:grid-cols-4 lg:min-w-[32rem]">
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.usage.inviteeCount') }}</div>
                <div class="mt-1 font-mono text-base font-semibold text-gray-900 dark:text-white">{{ formatInteger(selectedMemberGroup.invitee_count) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.usage.requests') }}</div>
                <div class="mt-1 font-mono text-base font-semibold text-gray-900 dark:text-white">{{ formatInteger(selectedMemberGroup.requests) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.usage.actualCost') }}</div>
                <div class="mt-1 font-mono text-base font-semibold text-emerald-600 dark:text-emerald-400">{{ formatCost(selectedMemberGroup.actual_cost) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ usageLabel('netProfit') }}</div>
                <button
                  type="button"
                  class="mt-1 inline-flex items-center justify-end gap-1.5 font-mono text-base font-semibold text-indigo-600 transition hover:text-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:text-indigo-400 dark:hover:text-indigo-300 dark:focus:ring-offset-dark-900"
                  @click="openProfitDetailsDialog(selectedMemberGroup)"
                >
                  <Icon name="eye" size="xs" />
                  <span>{{ formatCost(selectedMemberGroup.net_profit) }}</span>
                </button>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.usage.rechargeAmount') }}</div>
                <div class="mt-1 font-mono text-base font-semibold text-sky-600 dark:text-sky-400">{{ formatRechargeAmount(selectedMemberGroup.recharge_amount) }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="hidden overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700 md:block">
          <div class="max-h-[60vh] overflow-auto">
            <table class="w-full min-w-[46rem] divide-y divide-gray-200 dark:divide-dark-700">
              <thead class="sticky top-0 z-10 bg-gray-50 dark:bg-dark-800">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.records.user') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.requests') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.totalTokens') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.actualCost') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ usageLabel('netProfit') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.rechargeAmount') }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
                <tr v-for="member in selectedMemberGroup.members || []" :key="member.invitee_id" class="hover:bg-gray-50 dark:hover:bg-dark-800">
                  <td class="px-4 py-3">
                    <UserCell
                      :id="member.invitee_id"
                      :email="member.invitee_email"
                      :username="member.invitee_username"
                      clickable
                      @click="openAssignDialogForRecord(member)"
                    />
                  </td>
                  <td class="px-4 py-3 text-right font-mono text-sm text-gray-700 dark:text-dark-200">{{ formatInteger(member.requests) }}</td>
                  <td class="px-4 py-3 text-right font-mono text-sm font-medium text-gray-900 dark:text-white">{{ formatTokens(member.total_tokens) }}</td>
                  <td class="px-4 py-3 text-right font-mono text-sm font-semibold text-emerald-600 dark:text-emerald-400">{{ formatCost(member.actual_cost) }}</td>
                  <td class="px-4 py-3 text-right">
                    <button
                      type="button"
                      class="inline-flex items-center justify-end gap-1.5 font-mono text-sm font-semibold text-indigo-600 transition hover:text-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:text-indigo-400 dark:hover:text-indigo-300 dark:focus:ring-offset-dark-900"
                      @click="openProfitDetailsDialog(member)"
                    >
                      <Icon name="eye" size="xs" />
                      <span>{{ formatCost(member.net_profit) }}</span>
                    </button>
                  </td>
                  <td class="px-4 py-3 text-right font-mono text-sm font-semibold text-sky-600 dark:text-sky-400">{{ formatRechargeAmount(member.recharge_amount) }}</td>
                </tr>
                <tr v-if="!(selectedMemberGroup.members || []).length">
                  <td colspan="6" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-dark-400">
                    {{ t('empty.noData') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="space-y-3 md:hidden">
          <div
            v-for="member in selectedMemberGroup.members || []"
            :key="member.invitee_id"
            class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900"
          >
            <div class="flex items-start justify-between gap-3">
              <UserCell
                :id="member.invitee_id"
                :email="member.invitee_email"
                :username="member.invitee_username"
                clickable
                @click="openAssignDialogForRecord(member)"
              />
              <div class="shrink-0 text-right">
                <div class="text-sm font-semibold text-emerald-600 dark:text-emerald-400">{{ formatCost(member.actual_cost) }}</div>
                <div class="text-xs font-medium text-sky-600 dark:text-sky-400">{{ formatRechargeAmount(member.recharge_amount) }}</div>
              </div>
            </div>
            <div class="mt-3 grid grid-cols-2 gap-3 text-xs text-gray-500 dark:text-dark-400">
              <div>
                <div>{{ t('admin.affiliates.usage.requests') }}</div>
                <div class="mt-0.5 font-mono text-gray-900 dark:text-white">{{ formatInteger(member.requests) }}</div>
              </div>
              <div>
                <div>{{ t('admin.affiliates.usage.totalTokens') }}</div>
                <div class="mt-0.5 font-mono text-gray-900 dark:text-white">{{ formatTokens(member.total_tokens) }}</div>
              </div>
              <div>
                <div>{{ t('admin.affiliates.usage.actualCost') }}</div>
                <div class="mt-0.5 font-mono text-emerald-600 dark:text-emerald-400">{{ formatCost(member.actual_cost) }}</div>
              </div>
              <div>
                <div>{{ usageLabel('netProfit') }}</div>
                <button
                  type="button"
                  class="mt-0.5 inline-flex items-center gap-1 font-mono text-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:text-indigo-400 dark:focus:ring-offset-dark-900"
                  @click="openProfitDetailsDialog(member)"
                >
                  <Icon name="eye" size="xs" />
                  <span>{{ formatCost(member.net_profit) }}</span>
                </button>
              </div>
              <div>
                <div>{{ t('admin.affiliates.usage.rechargeAmount') }}</div>
                <div class="mt-0.5 font-mono text-sky-600 dark:text-sky-400">{{ formatRechargeAmount(member.recharge_amount) }}</div>
              </div>
            </div>
          </div>
          <div v-if="!(selectedMemberGroup.members || []).length" class="rounded-lg border border-gray-200 bg-white py-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400">
            {{ t('empty.noData') }}
          </div>
        </div>
      </div>

      <template #footer>
        <button class="btn btn-secondary" type="button" @click="closeMembersDialog">
          {{ t('common.close') }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="profitDetailsDialog"
      :title="t('admin.affiliates.usage.profitDetailsTitle')"
      width="extra-wide"
      @close="closeProfitDetailsDialog"
    >
      <div v-if="selectedProfitRecord" class="space-y-4">
        <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/60">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="min-w-0 space-y-2">
              <div class="text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                {{ t('admin.affiliates.usage.profitDetailsSubject') }}
              </div>
              <UnassignedCell v-if="selectedProfitRecord.unassigned && !selectedProfitRecord.invitee_id" />
              <UserCell
                v-else
                :id="profitDetailsSubjectId(selectedProfitRecord)"
                :email="profitDetailsSubjectEmail(selectedProfitRecord)"
                :username="profitDetailsSubjectUsername(selectedProfitRecord)"
              />
            </div>
            <div class="grid grid-cols-2 gap-3 text-right sm:grid-cols-4 lg:min-w-[34rem]">
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.usage.requests') }}</div>
                <div class="mt-1 font-mono text-base font-semibold text-gray-900 dark:text-white">{{ formatInteger(selectedProfitRecord.requests) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.usage.actualCost') }}</div>
                <div class="mt-1 font-mono text-base font-semibold text-emerald-600 dark:text-emerald-400">{{ formatCost(selectedProfitRecord.actual_cost) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ usageLabel('netProfit') }}</div>
                <div class="mt-1 font-mono text-base font-semibold text-indigo-600 dark:text-indigo-400">{{ formatCost(selectedProfitRecord.net_profit) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.usage.rebateAmount') }}</div>
                <div class="mt-1 font-mono text-base font-semibold text-primary-600 dark:text-primary-400">{{ formatRebateAmount(selectedProfitRecord.rebate_amount) }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="hidden overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700 md:block">
          <div class="max-h-[60vh] overflow-auto">
            <table class="w-full min-w-[58rem] divide-y divide-gray-200 dark:divide-dark-700">
              <thead class="sticky top-0 z-10 bg-gray-50 dark:bg-dark-800">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.profitDetailGroup') }}
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.profitDetailModel') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.requests') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.totalTokens') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.actualCost') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.profitDetailRate') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ usageLabel('netProfit') }}
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.rebateAmount') }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
                <tr v-for="detail in selectedProfitDetails" :key="profitDetailKey(detail)" class="hover:bg-gray-50 dark:hover:bg-dark-800">
                  <td class="px-4 py-3 text-sm font-medium text-gray-900 dark:text-white">{{ profitDetailGroupLabel(detail) }}</td>
                  <td class="px-4 py-3 font-mono text-sm text-gray-700 dark:text-dark-200">{{ detail.model || '-' }}</td>
                  <td class="px-4 py-3 text-right font-mono text-sm text-gray-700 dark:text-dark-200">{{ formatInteger(detail.requests) }}</td>
                  <td class="px-4 py-3 text-right font-mono text-sm font-medium text-gray-900 dark:text-white">{{ formatTokens(detail.total_tokens) }}</td>
                  <td class="px-4 py-3 text-right font-mono text-sm font-semibold text-emerald-600 dark:text-emerald-400">{{ formatCost(detail.actual_cost) }}</td>
                  <td class="px-4 py-3 text-right font-mono text-sm text-gray-700 dark:text-dark-200">{{ formatPercent(detail.profit_rate_percent) }}</td>
                  <td class="px-4 py-3 text-right font-mono text-sm font-semibold text-indigo-600 dark:text-indigo-400">{{ formatCost(detail.net_profit) }}</td>
                  <td class="px-4 py-3 text-right font-mono text-sm font-semibold text-primary-600 dark:text-primary-400">{{ formatRebateAmount(detail.rebate_amount) }}</td>
                </tr>
                <tr v-if="selectedProfitDetails.length === 0">
                  <td colspan="8" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-dark-400">
                    {{ t('admin.affiliates.usage.profitDetailsEmpty') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="space-y-3 md:hidden">
          <div
            v-for="detail in selectedProfitDetails"
            :key="profitDetailKey(detail)"
            class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ profitDetailGroupLabel(detail) }}</div>
                <div class="mt-1 truncate font-mono text-xs text-gray-500 dark:text-dark-400">{{ detail.model || '-' }}</div>
              </div>
              <div class="shrink-0 text-right">
                <div class="font-mono text-sm font-semibold text-indigo-600 dark:text-indigo-400">{{ formatCost(detail.net_profit) }}</div>
                <div class="mt-1 font-mono text-xs text-primary-600 dark:text-primary-400">{{ formatRebateAmount(detail.rebate_amount) }}</div>
              </div>
            </div>
            <div class="mt-3 grid grid-cols-2 gap-3 text-xs text-gray-500 dark:text-dark-400">
              <div>
                <div>{{ t('admin.affiliates.usage.requests') }}</div>
                <div class="mt-0.5 font-mono text-gray-900 dark:text-white">{{ formatInteger(detail.requests) }}</div>
              </div>
              <div>
                <div>{{ t('admin.affiliates.usage.totalTokens') }}</div>
                <div class="mt-0.5 font-mono text-gray-900 dark:text-white">{{ formatTokens(detail.total_tokens) }}</div>
              </div>
              <div>
                <div>{{ t('admin.affiliates.usage.actualCost') }}</div>
                <div class="mt-0.5 font-mono text-emerald-600 dark:text-emerald-400">{{ formatCost(detail.actual_cost) }}</div>
              </div>
              <div>
                <div>{{ t('admin.affiliates.usage.profitDetailRate') }}</div>
                <div class="mt-0.5 font-mono text-gray-900 dark:text-white">{{ formatPercent(detail.profit_rate_percent) }}</div>
              </div>
            </div>
          </div>
          <div v-if="selectedProfitDetails.length === 0" class="rounded-lg border border-gray-200 bg-white py-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400">
            {{ t('admin.affiliates.usage.profitDetailsEmpty') }}
          </div>
        </div>
      </div>

      <template #footer>
        <button class="btn btn-secondary" type="button" @click="closeProfitDetailsDialog">
          {{ t('common.close') }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="assignDialog"
      :title="t('admin.affiliates.usage.assignTitle')"
      width="wide"
      @close="closeAssignDialog"
    >
      <div class="grid min-h-[28rem] gap-4 md:grid-cols-2">
        <div class="space-y-2">
          <label class="block text-sm font-medium text-gray-700 dark:text-dark-300">
            {{ t('admin.affiliates.records.inviter') }}
          </label>
          <div class="relative">
            <input
              v-model="assignForm.inviterQuery"
              type="text"
              class="input"
              :placeholder="t('admin.affiliates.usage.userSearchPlaceholder')"
              @input="onAssignUserInput('inviter')"
              @focus="onAssignUserInput('inviter')"
              @keydown.enter.prevent="applyTypedAssignUser('inviter')"
              @blur="hideAssignResultsLater('inviter')"
            />
            <UserSearchResults
              v-if="showAssignInviterResults"
              :users="assignForm.inviterResults"
              :loading="assignForm.inviterLoading"
              @select="selectAssignUser('inviter', $event)"
            />
          </div>
          <SelectedUserCard v-if="assignForm.inviter" :user="assignForm.inviter" />
        </div>

        <div class="space-y-2">
          <label class="block text-sm font-medium text-gray-700 dark:text-dark-300">
            {{ t('admin.affiliates.records.invitee') }}
          </label>
          <div class="relative">
            <input
              v-model="assignForm.inviteeQuery"
              type="text"
              class="input"
              :placeholder="t('admin.affiliates.usage.userSearchPlaceholder')"
              @input="onAssignUserInput('invitee')"
              @focus="onAssignUserInput('invitee')"
              @keydown.enter.prevent="applyTypedAssignUser('invitee')"
              @blur="hideAssignResultsLater('invitee')"
            />
            <UserSearchResults
              v-if="showAssignInviteeResults"
              :users="assignForm.inviteeResults"
              :loading="assignForm.inviteeLoading"
              @select="selectAssignUser('invitee', $event)"
            />
          </div>
          <SelectedUserCard v-if="assignForm.invitee" :user="assignForm.invitee" />
        </div>
      </div>

      <template #footer>
        <button class="btn btn-secondary" type="button" @click="closeAssignDialog">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          type="button"
          :disabled="assignForm.submitting || !assignForm.inviter || !assignForm.invitee"
          @click="submitAssignment"
        >
          <Icon name="sync" size="md" :class="assignForm.submitting ? 'animate-spin' : ''" />
          <span>{{ t('admin.affiliates.usage.assignSubmit') }}</span>
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, type PropType } from 'vue'
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
  type AffiliateUsageDailyRecord,
  type AffiliateUsageProfitDetail,
  type AffiliateUsageSummary,
  type ListAffiliateUsageParams,
  type SimpleUser,
} from '@/api/admin/affiliates'
import { extractI18nErrorMessage } from '@/utils/apiError'

type UserRole = 'inviter' | 'invitee'
type UsageView = 'users' | 'groups'

const { t, locale } = useI18n()
const appStore = useAppStore()

const usageLabelFallbacks = {
  netProfit: {
    zh: '净利润',
    en: 'Net Profit',
  },
  summaryNetProfit: {
    zh: '区间净利润',
    en: 'Range Net Profit',
  },
} as const

type UsageLabelKey = keyof typeof usageLabelFallbacks

function usageLabel(key: UsageLabelKey): string {
  const path = `admin.affiliates.usage.${key}`
  const value = t(path)
  if (value !== path) return value
  const lang = String(locale.value || '').toLowerCase().startsWith('zh') ? 'zh' : 'en'
  return usageLabelFallbacks[key][lang]
}

const loading = ref(false)
const viewMode = ref<UsageView>('users')
const records = ref<AffiliateUsageDailyRecord[]>([])
const membersDialog = ref(false)
const selectedMemberGroup = ref<AffiliateUsageDailyRecord | null>(null)
const profitDetailsDialog = ref(false)
const selectedProfitRecord = ref<AffiliateUsageDailyRecord | null>(null)
const selectedProfitDetails = computed(() => {
  const record = selectedProfitRecord.value
  if (!record) return []
  if (record.profit_details?.length) return record.profit_details
  return aggregateProfitDetails(record.members || [])
})
const summary = reactive<AffiliateUsageSummary>({
  total_requests: 0,
  total_tokens: 0,
  total_actual_cost: 0,
  total_account_cost: 0,
  total_net_profit: 0,
  total_rebate_amount: 0,
})
const filters = reactive({
  search: '',
  start_at: defaultDateInput(-29),
  end_at: defaultDateInput(0),
  inviter_id: 0,
  invitee_id: 0,
})
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const sortState = reactive({ sort_by: 'actual_cost', sort_order: 'desc' as 'asc' | 'desc' })

const inviterQuery = ref('')
const inviteeQuery = ref('')
const filterInviterResults = ref<SimpleUser[]>([])
const filterInviteeResults = ref<SimpleUser[]>([])
const filterInviterLoading = ref(false)
const filterInviteeLoading = ref(false)
const showInviterResults = ref(false)
const showInviteeResults = ref(false)

const assignDialog = ref(false)
const assignForm = reactive({
  inviterQuery: '',
  inviteeQuery: '',
  inviter: null as SimpleUser | null,
  invitee: null as SimpleUser | null,
  inviterResults: [] as SimpleUser[],
  inviteeResults: [] as SimpleUser[],
  inviterLoading: false,
  inviteeLoading: false,
  submitting: false,
})
const showAssignInviterResults = ref(false)
const showAssignInviteeResults = ref(false)

let reloadTimer: ReturnType<typeof setTimeout> | null = null
let filterInviterTimer: ReturnType<typeof setTimeout> | null = null
let filterInviteeTimer: ReturnType<typeof setTimeout> | null = null
let assignInviterTimer: ReturnType<typeof setTimeout> | null = null
let assignInviteeTimer: ReturnType<typeof setTimeout> | null = null

const columns = computed<Column[]>(() => [
  ...(viewMode.value === 'users'
    ? [
        { key: 'user', label: t('admin.affiliates.records.user'), sortable: true },
        { key: 'inviter', label: t('admin.affiliates.records.inviter'), sortable: true },
      ]
    : [
        { key: 'inviter', label: t('admin.affiliates.records.inviter'), sortable: true },
        { key: 'invitee_count', label: t('admin.affiliates.usage.inviteeCount'), sortable: true },
      ]),
  { key: 'requests', label: t('admin.affiliates.usage.requests'), sortable: true },
  { key: 'total_tokens', label: t('admin.affiliates.usage.totalTokens'), sortable: true },
  { key: 'actual_cost', label: t('admin.affiliates.usage.actualCost'), sortable: true },
  { key: 'net_profit', label: usageLabel('netProfit'), sortable: true },
  { key: 'recharge_amount', label: t('admin.affiliates.usage.rechargeAmount'), sortable: true },
  ...(viewMode.value === 'groups'
    ? [
        { key: 'rebate_rate', label: t('admin.affiliates.usage.rebateRate'), sortable: true },
        { key: 'rebate_amount', label: t('admin.affiliates.usage.rebateAmount'), sortable: true },
      ]
    : []),
])

function userTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch {
    return 'UTC'
  }
}

function buildParams(): ListAffiliateUsageParams {
  return {
    page: pagination.page,
    page_size: pagination.page_size,
    search: filters.search.trim() || undefined,
    start_at: filters.start_at || undefined,
    end_at: filters.end_at || undefined,
    inviter_id: filters.inviter_id || undefined,
    invitee_id: filters.invitee_id || undefined,
    view: viewMode.value,
    sort_by: sortState.sort_by,
    sort_order: sortState.sort_order,
    timezone: userTimezone(),
  }
}

async function loadRecords() {
  loading.value = true
  try {
    const res = await affiliatesAPI.listUsageDailyRecords(buildParams())
    records.value = res.items || []
    if (viewMode.value !== 'groups') {
      closeMembersDialog()
    }
    pagination.total = res.total || 0
    summary.total_requests = res.summary?.total_requests || 0
    summary.total_tokens = res.summary?.total_tokens || 0
    summary.total_actual_cost = res.summary?.total_actual_cost || 0
    summary.total_account_cost = res.summary?.total_account_cost || 0
    summary.total_net_profit = res.summary?.total_net_profit || 0
    summary.total_rebate_amount = res.summary?.total_rebate_amount || 0
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function debounceLoad() {
  if (reloadTimer) clearTimeout(reloadTimer)
  reloadTimer = setTimeout(() => reloadFromFirstPage(), 300)
}

function reloadFromFirstPage() {
  pagination.page = 1
  void loadRecords()
}

function handlePageChange(page: number) {
  pagination.page = page
  void loadRecords()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  void loadRecords()
}

function handleSort(key: string, order: 'asc' | 'desc') {
  sortState.sort_by = key
  sortState.sort_order = order
  pagination.page = 1
  void loadRecords()
}

function setViewMode(mode: UsageView) {
  if (viewMode.value === mode) return
  viewMode.value = mode
  closeMembersDialog()
  closeProfitDetailsDialog()
  sortState.sort_by = 'actual_cost'
  sortState.sort_order = 'desc'
  pagination.page = 1
  void loadRecords()
}

function groupMemberCount(row: AffiliateUsageDailyRecord): number {
  return row.members?.length || Number(row.invitee_count || 0)
}

function openMembersDialog(row: AffiliateUsageDailyRecord) {
  selectedMemberGroup.value = row
  membersDialog.value = true
}

function closeMembersDialog() {
  membersDialog.value = false
  selectedMemberGroup.value = null
}

async function openProfitDetailsDialog(row: AffiliateUsageDailyRecord) {
  const target = resolveProfitDetailsRecord(row)
  selectedProfitRecord.value = target
  profitDetailsDialog.value = true

  const current = selectedProfitRecord.value
  const hasCurrentDetails = !!current?.profit_details?.length || aggregateProfitDetails(current?.members || []).length > 0
  if (hasCurrentDetails) return

  try {
    const res = await affiliatesAPI.listUsageDailyRecords(buildParams())
    const latest = findMatchingProfitDetailsRecord(row, collectProfitDetailsCandidates(res.items || []))
    if (latest && profitDetailsDialog.value && selectedProfitRecord.value && isSameProfitDetailsRecord(selectedProfitRecord.value, target)) {
      selectedProfitRecord.value = latest
    }
  } catch {
    // Keep the current snapshot if refresh fails; the dialog still shows the summary block.
  }
}

function closeProfitDetailsDialog() {
  profitDetailsDialog.value = false
  selectedProfitRecord.value = null
}

function profitDetailsSubjectId(row: AffiliateUsageDailyRecord): number {
  return row.invitee_id || row.inviter_id || 0
}

function profitDetailsSubjectEmail(row: AffiliateUsageDailyRecord): string {
  return row.invitee_id ? row.invitee_email : row.inviter_email
}

function profitDetailsSubjectUsername(row: AffiliateUsageDailyRecord): string {
  return row.invitee_id ? row.invitee_username : row.inviter_username
}

function profitDetailGroupLabel(detail: AffiliateUsageProfitDetail): string {
  if (detail.group_name) return detail.group_name
  if (detail.group_id > 0) return `#${detail.group_id}`
  return t('admin.affiliates.usage.unassignedGroup')
}

function profitDetailKey(detail: AffiliateUsageProfitDetail): string {
  return `${detail.group_id}:${detail.group_name}:${detail.model}:${detail.profit_rate_percent}`
}

function resolveProfitDetailsRecord(row: AffiliateUsageDailyRecord): AffiliateUsageDailyRecord {
  if (row.profit_details?.length || row.members?.length) return row

  return findMatchingProfitDetailsRecord(row, collectProfitDetailsCandidates([
    ...records.value,
    ...(selectedMemberGroup.value?.members || []),
  ])) || row
}

function collectProfitDetailsCandidates(source: AffiliateUsageDailyRecord[]): AffiliateUsageDailyRecord[] {
  const candidates: AffiliateUsageDailyRecord[] = []
  for (const item of source) {
    candidates.push(item)
    for (const member of item.members || []) {
      candidates.push(member)
    }
  }
  return candidates
}

function findMatchingProfitDetailsRecord(
  target: AffiliateUsageDailyRecord,
  source: AffiliateUsageDailyRecord[],
): AffiliateUsageDailyRecord | null {
  return source.find((item) => (
    item.inviter_id === target.inviter_id
    && item.invitee_id === target.invitee_id
    && item.unassigned === target.unassigned
  )) || null
}

function isSameProfitDetailsRecord(left: AffiliateUsageDailyRecord, right: AffiliateUsageDailyRecord): boolean {
  return left.inviter_id === right.inviter_id
    && left.invitee_id === right.invitee_id
    && left.unassigned === right.unassigned
}

function aggregateProfitDetails(records: AffiliateUsageDailyRecord[]): AffiliateUsageProfitDetail[] {
  const byKey = new Map<string, AffiliateUsageProfitDetail>()
  for (const record of records) {
    for (const detail of record.profit_details || []) {
      const key = profitDetailKey(detail)
      const existing = byKey.get(key)
      if (existing) {
        existing.requests += Number(detail.requests || 0)
        existing.total_tokens += Number(detail.total_tokens || 0)
        existing.actual_cost += Number(detail.actual_cost || 0)
        existing.net_profit += Number(detail.net_profit || 0)
        existing.rebate_amount += Number(detail.rebate_amount || 0)
      } else {
        byKey.set(key, {
          group_id: detail.group_id,
          group_name: detail.group_name,
          model: detail.model,
          requests: Number(detail.requests || 0),
          total_tokens: Number(detail.total_tokens || 0),
          actual_cost: Number(detail.actual_cost || 0),
          profit_rate_percent: Number(detail.profit_rate_percent || 0),
          net_profit: Number(detail.net_profit || 0),
          rebate_amount: Number(detail.rebate_amount || 0),
        })
      }
    }
  }
  return Array.from(byKey.values()).sort((left, right) => (
    Number(right.net_profit || 0) - Number(left.net_profit || 0)
    || Number(right.actual_cost || 0) - Number(left.actual_cost || 0)
    || profitDetailGroupLabel(left).localeCompare(profitDetailGroupLabel(right))
    || String(left.model || '').localeCompare(String(right.model || ''))
  ))
}

function onFilterUserInput(role: UserRole) {
  const timer = role === 'inviter' ? filterInviterTimer : filterInviteeTimer
  if (timer) clearTimeout(timer)
  const run = () => searchFilterUsers(role)
  if (role === 'inviter') {
    filterInviterTimer = setTimeout(run, 250)
  } else {
    filterInviteeTimer = setTimeout(run, 250)
  }
}

async function searchFilterUsers(role: UserRole) {
  const query = (role === 'inviter' ? inviterQuery.value : inviteeQuery.value).trim()
  const showRef = role === 'inviter' ? showInviterResults : showInviteeResults
  const loadingRef = role === 'inviter' ? filterInviterLoading : filterInviteeLoading
  const resultsRef = role === 'inviter' ? filterInviterResults : filterInviteeResults
  showRef.value = true
  loadingRef.value = true
  try {
    resultsRef.value = await affiliatesAPI.lookupUsers(query)
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    loadingRef.value = false
  }
}

function selectFilterUser(role: UserRole, user: SimpleUser) {
  if (role === 'inviter') {
    filters.inviter_id = user.id
    inviterQuery.value = userLabel(user)
    showInviterResults.value = false
  } else {
    filters.invitee_id = user.id
    inviteeQuery.value = userLabel(user)
    showInviteeResults.value = false
  }
  reloadFromFirstPage()
}

async function applyTypedFilterUser(role: UserRole) {
  const query = (role === 'inviter' ? inviterQuery.value : inviteeQuery.value).trim()
  const showRef = role === 'inviter' ? showInviterResults : showInviteeResults
  const loadingRef = role === 'inviter' ? filterInviterLoading : filterInviteeLoading
  const resultsRef = role === 'inviter' ? filterInviterResults : filterInviteeResults
  showRef.value = true
  loadingRef.value = true
  try {
    const users = await affiliatesAPI.lookupUsers(query)
    resultsRef.value = users
    const picked = pickBestUser(query, users)
    if (picked) selectFilterUser(role, picked)
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    loadingRef.value = false
  }
}

function clearFilterUser(role: UserRole) {
  if (role === 'inviter') {
    filters.inviter_id = 0
    inviterQuery.value = ''
    filterInviterResults.value = []
    showInviterResults.value = false
  } else {
    filters.invitee_id = 0
    inviteeQuery.value = ''
    filterInviteeResults.value = []
    showInviteeResults.value = false
  }
  reloadFromFirstPage()
}

function hideFilterResultsLater(role: UserRole) {
  setTimeout(() => {
    if (role === 'inviter') {
      showInviterResults.value = false
    } else {
      showInviteeResults.value = false
    }
  }, 160)
}

function openAssignDialog() {
  resetAssignForm()
  assignDialog.value = true
}

function closeAssignDialog() {
  assignDialog.value = false
  showAssignInviterResults.value = false
  showAssignInviteeResults.value = false
}

function openAssignDialogForRecord(row: AffiliateUsageDailyRecord) {
  resetAssignForm()
  const invitee = usageRecordUser(row, 'invitee')
  if (invitee) selectAssignUser('invitee', invitee)

  const inviter = usageRecordUser(row, 'inviter')
  if (inviter) selectAssignUser('inviter', inviter)

  assignDialog.value = true
}

function resetAssignForm() {
  assignForm.inviterQuery = ''
  assignForm.inviteeQuery = ''
  assignForm.inviter = null
  assignForm.invitee = null
  assignForm.inviterResults = []
  assignForm.inviteeResults = []
  assignForm.inviterLoading = false
  assignForm.inviteeLoading = false
  showAssignInviterResults.value = false
  showAssignInviteeResults.value = false
}

function onAssignUserInput(role: UserRole) {
  if (role === 'inviter' && assignForm.inviter && assignForm.inviterQuery !== userLabel(assignForm.inviter)) {
    assignForm.inviter = null
  }
  if (role === 'invitee' && assignForm.invitee && assignForm.inviteeQuery !== userLabel(assignForm.invitee)) {
    assignForm.invitee = null
  }
  const timer = role === 'inviter' ? assignInviterTimer : assignInviteeTimer
  if (timer) clearTimeout(timer)
  const run = () => searchAssignUsers(role)
  if (role === 'inviter') {
    assignInviterTimer = setTimeout(run, 250)
  } else {
    assignInviteeTimer = setTimeout(run, 250)
  }
}

async function searchAssignUsers(role: UserRole) {
  const query = (role === 'inviter' ? assignForm.inviterQuery : assignForm.inviteeQuery).trim()
  const showRef = role === 'inviter' ? showAssignInviterResults : showAssignInviteeResults
  showRef.value = true
  if (role === 'inviter') assignForm.inviterLoading = true
  else assignForm.inviteeLoading = true
  try {
    const users = await affiliatesAPI.lookupUsers(query)
    if (role === 'inviter') {
      assignForm.inviterResults = users
    } else {
      assignForm.inviteeResults = users
    }
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    if (role === 'inviter') assignForm.inviterLoading = false
    else assignForm.inviteeLoading = false
  }
}

async function applyTypedAssignUser(role: UserRole) {
  const query = (role === 'inviter' ? assignForm.inviterQuery : assignForm.inviteeQuery).trim()
  const showRef = role === 'inviter' ? showAssignInviterResults : showAssignInviteeResults
  showRef.value = true
  if (role === 'inviter') assignForm.inviterLoading = true
  else assignForm.inviteeLoading = true
  try {
    const users = await affiliatesAPI.lookupUsers(query)
    if (role === 'inviter') {
      assignForm.inviterResults = users
    } else {
      assignForm.inviteeResults = users
    }
    const picked = pickBestUser(query, users)
    if (picked) selectAssignUser(role, picked)
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    if (role === 'inviter') assignForm.inviterLoading = false
    else assignForm.inviteeLoading = false
  }
}

function selectAssignUser(role: UserRole, user: SimpleUser) {
  if (role === 'inviter') {
    assignForm.inviter = user
    assignForm.inviterQuery = userLabel(user)
    assignForm.inviterResults = []
    showAssignInviterResults.value = false
  } else {
    assignForm.invitee = user
    assignForm.inviteeQuery = userLabel(user)
    assignForm.inviteeResults = []
    showAssignInviteeResults.value = false
  }
}

function hideAssignResultsLater(role: UserRole) {
  setTimeout(() => {
    if (role === 'inviter') {
      showAssignInviterResults.value = false
    } else {
      showAssignInviteeResults.value = false
    }
  }, 160)
}

async function submitAssignment() {
  if (!assignForm.inviter || !assignForm.invitee) return
  if (assignForm.inviter.id === assignForm.invitee.id) {
    appStore.showError(t('admin.affiliates.usage.selfAssignError'))
    return
  }
  assignForm.submitting = true
  try {
    const res = await affiliatesAPI.assignInviter({
      inviter_id: assignForm.inviter.id,
      invitee_id: assignForm.invitee.id,
    })
    appStore.showSuccess(res.changed ? t('admin.affiliates.usage.assignSuccess') : t('admin.affiliates.usage.assignNoChange'))
    assignDialog.value = false
    filters.inviter_id = assignForm.inviter.id
    inviterQuery.value = userLabel(assignForm.inviter)
    pagination.page = 1
    await loadRecords()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    assignForm.submitting = false
  }
}

function userLabel(user: SimpleUser): string {
  return `#${user.id} ${user.email || user.username || '-'}`
}

function usageRecordUser(row: AffiliateUsageDailyRecord, role: UserRole): SimpleUser | null {
  const id = role === 'inviter' ? row.inviter_id : row.invitee_id
  if (!id) return null
  return {
    id,
    email: role === 'inviter' ? row.inviter_email : row.invitee_email,
    username: role === 'inviter' ? row.inviter_username : row.invitee_username,
  }
}

function pickBestUser(query: string, users: SimpleUser[]): SimpleUser | null {
  if (users.length === 0) return null
  const normalized = query.trim().replace(/^#/, '').toLowerCase()
  if (!normalized) return users[0]
  if (/^\d+$/.test(normalized)) {
    const id = Number(normalized)
    const exactID = users.find((user) => user.id === id)
    if (exactID) return exactID
  }
  const exactText = users.find((user) => {
    const email = (user.email || '').toLowerCase()
    const username = (user.username || '').toLowerCase()
    return email === normalized || username === normalized
  })
  return exactText || users[0]
}

function defaultDateInput(offsetDays: number): string {
  const date = new Date()
  date.setDate(date.getDate() + offsetDays)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function formatInteger(value: number | null | undefined): string {
  return Math.round(Number(value || 0)).toLocaleString()
}

function formatTokens(value: number | null | undefined): string {
  const n = Number(value || 0)
  if (n >= 1_000_000_000) return `${trimNumber(n / 1_000_000_000)}B`
  if (n >= 1_000_000) return `${trimNumber(n / 1_000_000)}M`
  if (n >= 1_000) return `${trimNumber(n / 1_000)}K`
  return formatInteger(n)
}

function trimNumber(value: number): string {
  return value.toFixed(value >= 10 ? 1 : 2).replace(/\.0+$/, '').replace(/(\.\d*[1-9])0+$/, '$1')
}

function formatCost(value: number | null | undefined): string {
  return formatUsdAmount(value)
}

function formatRechargeAmount(value: number | null | undefined): string {
  return formatCnyAmount(value)
}

function formatRebateAmount(value: number | null | undefined): string {
  return formatCnyAmount(value)
}

function formatUsdAmount(value: number | null | undefined): string {
  const n = Number(value || 0)
  const amount = Math.abs(n).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 4,
  })
  return `${n < 0 ? '-' : ''}$${amount}`
}

function formatCnyAmount(value: number | null | undefined): string {
  const n = Number(value || 0)
  const amount = Math.abs(n).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 4,
  })
  return `${n < 0 ? '-' : ''}\u00a5${amount}`
}

function formatPercent(value: number | null | undefined): string {
  const n = Number(value || 0)
  return `${trimNumber(n)}%`
}

const UserCell = defineComponent({
  props: {
    id: { type: Number, required: true },
    email: { type: String, default: '' },
    username: { type: String, default: '' },
    clickable: { type: Boolean, default: false },
  },
  emits: ['click'],
  setup(cellProps, { emit }) {
    const children = () => [
      h('div', { class: 'font-mono text-sm text-gray-900 dark:text-white' }, `#${cellProps.id}`),
      h('div', { class: 'max-w-56 truncate text-sm font-medium text-gray-900 dark:text-white' }, cellProps.email || '-'),
      h('div', { class: 'max-w-56 truncate text-sm text-gray-500 dark:text-dark-400' }, cellProps.username || '-'),
    ]

    return () => cellProps.clickable
      ? h('button', {
        type: 'button',
        class: 'space-y-0.5 rounded-md transition hover:bg-primary-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:hover:bg-primary-900/20 dark:focus:ring-offset-dark-900',
        style: { textAlign: 'inherit' },
        onClick: () => emit('click'),
      }, children())
      : h('div', { class: 'space-y-0.5' }, children())
  },
})

const UnassignedCell = defineComponent({
  setup() {
    return () => h('div', { class: 'inline-flex rounded-md bg-amber-50 px-2.5 py-1 text-sm font-medium text-amber-700 dark:bg-amber-900/20 dark:text-amber-300' }, t('admin.affiliates.usage.unassignedGroup'))
  },
})

const SummaryStat = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
    icon: { type: String as PropType<'chart' | 'creditCard' | 'database' | 'dollar'>, required: true },
  },
  setup(statProps) {
    return () => h('div', { class: 'min-h-[6.25rem] rounded-lg border border-gray-200 bg-white p-3 shadow-sm dark:border-dark-700 dark:bg-dark-800 sm:p-4' }, [
      h('div', { class: 'flex items-start justify-between gap-3' }, [
        h('div', { class: 'min-w-0 text-sm leading-5 text-gray-500 dark:text-dark-400' }, statProps.label),
        h('div', { class: 'shrink-0 rounded-lg bg-primary-50 p-2 text-primary-600 dark:bg-primary-900/20 dark:text-primary-400' }, [
          h(Icon, { name: statProps.icon, size: 'md' }),
        ]),
      ]),
      h('div', { class: 'mt-2 truncate text-2xl font-semibold text-gray-900 dark:text-white' }, statProps.value),
    ])
  },
})

const UserSearchResults = defineComponent({
  props: {
    users: { type: Array as PropType<SimpleUser[]>, required: true },
    loading: { type: Boolean, default: false },
  },
  emits: ['select'],
  setup(resultProps, { emit }) {
    return () => h('div', {
      class: 'absolute left-0 right-0 top-full z-50 mt-2 max-h-72 overflow-y-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-900',
    }, resultProps.loading
      ? [h('div', { class: 'px-3 py-3 text-sm text-gray-500 dark:text-dark-400' }, t('common.loading'))]
      : resultProps.users.length === 0
        ? [h('div', { class: 'px-3 py-3 text-sm text-gray-500 dark:text-dark-400' }, t('common.noData'))]
        : resultProps.users.map((user) => h('button', {
          key: user.id,
          type: 'button',
          class: 'block w-full px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-dark-800',
          onMouseDown: (event: MouseEvent) => event.preventDefault(),
          onClick: () => emit('select', user),
        }, [
          h('div', { class: 'font-mono text-xs text-gray-500 dark:text-dark-400' }, `#${user.id}`),
          h('div', { class: 'truncate text-sm font-medium text-gray-900 dark:text-white' }, user.email || '-'),
          h('div', { class: 'truncate text-xs text-gray-500 dark:text-dark-400' }, user.username || '-'),
        ])))
  },
})

const SelectedUserCard = defineComponent({
  props: {
    user: { type: Object as PropType<SimpleUser>, required: true },
  },
  setup(cardProps) {
    return () => h('div', { class: 'rounded-lg border border-gray-100 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-800' }, [
      h('div', { class: 'font-mono text-xs text-gray-500 dark:text-dark-400' }, `#${cardProps.user.id}`),
      h('div', { class: 'mt-1 truncate text-sm font-medium text-gray-900 dark:text-white' }, cardProps.user.email || '-'),
      h('div', { class: 'truncate text-xs text-gray-500 dark:text-dark-400' }, cardProps.user.username || '-'),
    ])
  },
})

onMounted(() => {
  void loadRecords()
})
</script>

<style scoped>
@media (min-width: 1024px) {
  :deep(.affiliate-usage-layout) {
    gap: 1rem;
  }

  :deep(.affiliate-usage-layout .layout-section-scrollable) {
    min-height: 24rem;
  }
}
</style>
