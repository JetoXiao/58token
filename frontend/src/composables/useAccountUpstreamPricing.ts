import { reactive } from 'vue'
import { adminAPI } from '@/api/admin'
import type { UpstreamPricingSnapshot } from '@/api/admin/accounts'
import type { Account } from '@/types'

export interface AccountUpstreamPricingState {
  loading: boolean
  snapshot: UpstreamPricingSnapshot | null
  error: string
  updatedAt: number
}

interface UpstreamPricingLoaderOptions {
  staleMs?: number
  concurrency?: number
  fetchPricing?: (accountId: number) => Promise<UpstreamPricingSnapshot>
  now?: () => number
}

const DEFAULT_STALE_MS = 30 * 1000
const DEFAULT_CONCURRENCY = 4

export function accountUpstreamPricingKey(account: Account): string | null {
  if (account.type !== 'apikey') return null
  const raw = account.credentials?.base_url
  if (typeof raw !== 'string' || !raw.trim()) return null
  const groupKey = account.credentials?.upstream_group_key
  if (typeof groupKey !== 'string' || !groupKey.trim()) return null
  return String(account.id)
}

function errorMessage(error: any): string {
  return error?.response?.data?.detail || error?.response?.data?.message || error?.message || 'Failed to fetch upstream pricing'
}

export function useAccountUpstreamPricing(options: UpstreamPricingLoaderOptions = {}) {
  const states = reactive<Record<string, AccountUpstreamPricingState>>({})
  const staleMs = options.staleMs ?? DEFAULT_STALE_MS
  const concurrency = Math.max(1, options.concurrency ?? DEFAULT_CONCURRENCY)
  const fetchPricing = options.fetchPricing ?? adminAPI.accounts.getUpstreamPricing
  const now = options.now ?? Date.now

  function stateFor(account: Account): AccountUpstreamPricingState | null {
    const key = accountUpstreamPricingKey(account)
    return key ? states[key] || null : null
  }

  async function load(accounts: Account[], force = false): Promise<void> {
    const representatives = new Map<string, Account>()
    for (const account of accounts) {
      const key = accountUpstreamPricingKey(account)
      if (key && !representatives.has(key)) representatives.set(key, account)
    }

    const currentTime = now()
    const pending = [...representatives.entries()].filter(([key]) => {
      const state = states[key]
      if (!state) return true
      if (state.loading) return false
      return force || currentTime - state.updatedAt >= staleMs
    })

    let cursor = 0
    const worker = async () => {
      while (cursor < pending.length) {
        const [key, account] = pending[cursor++]
        const previous = states[key]
        states[key] = {
          loading: true,
          snapshot: previous?.snapshot || null,
          error: '',
          updatedAt: previous?.updatedAt || 0
        }
        try {
          const snapshot = await fetchPricing(account.id)
          states[key] = { loading: false, snapshot, error: '', updatedAt: now() }
        } catch (error) {
          states[key] = {
            loading: false,
            snapshot: previous?.snapshot || null,
            error: errorMessage(error),
            updatedAt: now()
          }
        }
      }
    }

    await Promise.all(Array.from({ length: Math.min(concurrency, pending.length) }, () => worker()))
  }

  return { states, stateFor, load }
}
