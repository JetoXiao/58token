import { apiClient } from './client'
import type { HelpCenterConfig } from '@/types'

export interface HelpCenterResponse {
  config: HelpCenterConfig
  key_prompt_dismissed: boolean
  help_center_key_prompt_dismissed: boolean
}

export async function getHelpCenter(): Promise<HelpCenterResponse> {
  const { data } = await apiClient.get<HelpCenterResponse>('/help-center')
  return data
}

export async function dismissKeyCreatedPrompt(): Promise<{ dismissed: boolean }> {
  const { data } = await apiClient.post<{ dismissed: boolean }>('/help-center/key-created-prompt/dismiss')
  return data
}

export const helpCenterAPI = {
  get: getHelpCenter,
  dismissKeyCreatedPrompt,
}

export default helpCenterAPI
