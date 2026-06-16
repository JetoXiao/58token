import { apiClient } from './client'
import type { FreeQuotaSummary, RedeemTrialCardResponse } from '@/types'

export async function getSummary(): Promise<FreeQuotaSummary> {
  const { data } = await apiClient.get<FreeQuotaSummary>('/user/free-quota/summary')
  return data
}

export async function redeemTrialCard(code: string): Promise<RedeemTrialCardResponse> {
  const { data } = await apiClient.post<RedeemTrialCardResponse>('/redeem/trial', { code })
  return data
}

export const freeQuotaAPI = {
  getSummary,
  redeemTrialCard
}

export default freeQuotaAPI
