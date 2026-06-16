import { apiClient } from '../client'
import type { FreeQuotaSettings, PaginatedResponse, TrialCard } from '@/types'

export interface TrialCardPayload {
  code: string
  name?: string
  amount: number
  max_redemptions: number
  per_user_limit?: number
  status?: string
  notes?: string
  expires_at?: string | null
}

export interface TrialCardUpdatePayload {
  name?: string
  amount?: number
  max_redemptions?: number
  per_user_limit?: number
  status?: string
  notes?: string
  expires_at?: string | null
  clear_expires_at?: boolean
}

export async function list(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<TrialCard>> {
  const { data } = await apiClient.get<PaginatedResponse<TrialCard>>('/admin/trial-cards', {
    params: {
      page,
      page_size: pageSize
    }
  })
  return data
}

export async function create(payload: TrialCardPayload): Promise<TrialCard> {
  const { data } = await apiClient.post<TrialCard>('/admin/trial-cards', payload)
  return data
}

export async function update(id: number, payload: TrialCardUpdatePayload): Promise<TrialCard> {
  const { data } = await apiClient.put<TrialCard>(`/admin/trial-cards/${id}`, payload)
  return data
}

export async function deleteCard(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/trial-cards/${id}`)
  return data
}

export async function getSettings(): Promise<FreeQuotaSettings> {
  const { data } = await apiClient.get<FreeQuotaSettings>('/admin/free-quota/settings')
  return data
}

export async function updateSettings(settings: FreeQuotaSettings): Promise<FreeQuotaSettings> {
  const { data } = await apiClient.put<FreeQuotaSettings>('/admin/free-quota/settings', settings)
  return data
}

export const trialCardsAPI = {
  list,
  create,
  update,
  delete: deleteCard,
  getSettings,
  updateSettings
}

export default trialCardsAPI
