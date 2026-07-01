import { apiClient } from '../client'
import type { HelpCenterAttachment, HelpCenterConfig } from '@/types'

export interface AdminHelpCenterResponse {
  draft: HelpCenterConfig
  published: HelpCenterConfig
}

export async function get(): Promise<AdminHelpCenterResponse> {
  const { data } = await apiClient.get<AdminHelpCenterResponse>('/admin/help-center')
  return data
}

export async function saveDraft(config: HelpCenterConfig): Promise<HelpCenterConfig> {
  const { data } = await apiClient.put<{ draft: HelpCenterConfig }>('/admin/help-center/draft', { config })
  return data.draft
}

export async function publishDraft(): Promise<HelpCenterConfig> {
  const { data } = await apiClient.post<{ published: HelpCenterConfig }>('/admin/help-center/publish')
  return data.published
}

export async function uploadAttachment(file: File): Promise<HelpCenterAttachment> {
  const formData = new FormData()
  formData.append('file', file)
  const { data } = await apiClient.post<HelpCenterAttachment>('/admin/help-center/attachments', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

export const helpCenterAPI = {
  get,
  saveDraft,
  publishDraft,
  uploadAttachment,
}

export default helpCenterAPI
