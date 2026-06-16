import { defineStore } from 'pinia'
import { ref } from 'vue'
import { freeQuotaAPI } from '@/api'
import type { FreeQuotaSummary } from '@/types'

const EMPTY_SUMMARY: FreeQuotaSummary = {
  balance_amount: 0,
  free_quota_amount: 0,
  total_amount: 0
}

export const useFreeQuotaStore = defineStore('freeQuota', () => {
  const summary = ref<FreeQuotaSummary>({ ...EMPTY_SUMMARY })
  const loading = ref(false)
  let activePromise: Promise<FreeQuotaSummary> | null = null

  async function fetchSummary(force = false): Promise<FreeQuotaSummary> {
    if (activePromise && !force) {
      return activePromise
    }

    loading.value = true
    const request = freeQuotaAPI
      .getSummary()
      .then((data) => {
        summary.value = data || { ...EMPTY_SUMMARY }
        return summary.value
      })
      .catch((error) => {
        console.error('Failed to fetch free quota summary:', error)
        throw error
      })
      .finally(() => {
        if (activePromise === request) {
          activePromise = null
          loading.value = false
        }
      })

    activePromise = request
    return request
  }

  function applyTrialGrant(amount: number, balanceAmount?: number | null) {
    const safeAmount = Number(amount || 0)
    const balance = typeof balanceAmount === 'number' ? balanceAmount : summary.value.balance_amount
    summary.value = {
      balance_amount: balance,
      free_quota_amount: Number(summary.value.free_quota_amount || 0) + safeAmount,
      total_amount: balance + Number(summary.value.free_quota_amount || 0) + safeAmount
    }
  }

  function setBalanceAmount(balanceAmount: number) {
    const balance = Number(balanceAmount || 0)
    summary.value = {
      ...summary.value,
      balance_amount: balance,
      total_amount: balance + Number(summary.value.free_quota_amount || 0)
    }
  }

  function clear() {
    activePromise = null
    summary.value = { ...EMPTY_SUMMARY }
    loading.value = false
  }

  return {
    summary,
    loading,
    fetchSummary,
    applyTrialGrant,
    setBalanceAmount,
    clear
  }
})
