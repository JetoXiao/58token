<template>
  <div class="card p-4">
    <div class="mb-1 flex items-center justify-between gap-3">
      <div>
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.dashboard.balanceDeductionTitle') }}</h3>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.dashboard.balanceDeductionDescription') }}</p>
      </div>
      <span class="text-sm font-semibold text-emerald-600 dark:text-emerald-400">${{ formatCost(totalCost) }}</span>
    </div>
    <div v-if="loading" class="flex h-64 items-center justify-center"><LoadingSpinner /></div>
    <div v-else-if="points.length && chartData" class="h-64"><Chart type="bar" :data="chartData" :options="chartOptions" /></div>
    <div v-else class="flex h-64 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('admin.dashboard.noDataAvailable') }}</div>
    <div v-if="points.length" class="mt-3 max-h-44 overflow-y-auto rounded border border-gray-100 text-xs dark:border-gray-700">
      <div v-for="point in points" :key="point.date" class="border-b border-gray-100 last:border-0 dark:border-gray-700">
        <div class="flex items-center justify-between bg-gray-50 px-2 py-1 font-medium dark:bg-gray-800"><span>{{ point.date }}</span><span>${{ formatCost(point.total_actual_cost) }}</span></div>
        <div v-for="user in point.users" :key="`${point.date}-${user.user_id}`" class="flex items-center justify-between px-2 py-1 text-gray-500 dark:text-gray-400"><span class="truncate pr-2">{{ displayName(user) }}</span><span class="whitespace-nowrap">${{ formatCost(user.actual_cost) }} · {{ user.requests }} {{ t('admin.dashboard.requests') }}</span></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip
} from 'chart.js'
import { Chart } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { DailyBalanceDeductionPoint } from '@/types'

ChartJS.register(CategoryScale, LinearScale, BarElement, LineElement, PointElement, Tooltip, Legend)

const { t } = useI18n()
const props = defineProps<{ points: DailyBalanceDeductionPoint[]; loading?: boolean }>()

const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316']
const points = computed(() => props.points || [])
const totalCost = computed(() => points.value.reduce((sum, point) => sum + Number(point.total_actual_cost || 0), 0))

const displayName = (user: { username?: string; email?: string; user_id: number }) =>
  user.username?.trim() || user.email?.trim() || `#${user.user_id}`

const selectedUsers = computed(() => {
  const totals = new Map<number, { name: string; cost: number }>()
  points.value.forEach((point) => point.users?.forEach((user) => {
    const existing = totals.get(user.user_id)
    totals.set(user.user_id, { name: displayName(user), cost: (existing?.cost || 0) + Number(user.actual_cost || 0) })
  }))
  return Array.from(totals.entries()).sort((a, b) => b[1].cost - a[1].cost).slice(0, colors.length)
})

const chartData = computed(() => {
  if (!points.value.length) return null
  const labels = points.value.map((point) => point.date)
  const datasets: any[] = selectedUsers.value.map(([userId, user], index) => ({
    type: 'bar',
    label: user.name,
    stack: 'users',
    backgroundColor: colors[index],
    borderRadius: 3,
    data: points.value.map((point) => point.users?.find((item) => item.user_id === userId)?.actual_cost || 0)
  }))
  datasets.push({
    type: 'bar',
    label: t('admin.dashboard.otherUsers'),
    stack: 'users',
    backgroundColor: '#cbd5e1',
    borderRadius: 3,
    data: points.value.map((point) => {
      const known = new Set(selectedUsers.value.map(([userId]) => userId))
      const selected = point.users?.filter((user) => known.has(user.user_id)).reduce((sum, user) => sum + Number(user.actual_cost || 0), 0) || 0
      return Math.max(0, Number(point.total_actual_cost || 0) - selected)
    })
  })
  datasets.push({
    type: 'line',
    label: t('admin.dashboard.platformTotal'),
    data: points.value.map((point) => point.total_actual_cost),
    borderColor: '#0f766e',
    backgroundColor: '#0f766e',
    borderWidth: 2,
    pointRadius: 3,
    tension: 0.25,
    yAxisID: 'cost'
  })
  return { labels, datasets }
})

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index' as const, intersect: false },
  plugins: {
    legend: { position: 'bottom' as const, labels: { usePointStyle: true, boxWidth: 8, font: { size: 10 } } },
    tooltip: { callbacks: { label: (context: any) => `${context.dataset.label}: $${formatCost(Number(context.raw || 0))}` } }
  },
  scales: {
    x: { stacked: true },
    y: { stacked: true, beginAtZero: true, ticks: { callback: (value: string | number) => `$${formatCost(Number(value))}` } },
    cost: { display: false, beginAtZero: true }
  }
}))

const formatCost = (value: number) => value >= 1000 ? `${(value / 1000).toFixed(2)}K` : value >= 1 ? value.toFixed(2) : value.toFixed(4)
</script>
