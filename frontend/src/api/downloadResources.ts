import { apiClient } from './client'

export interface DownloadResource {
  id: number
  slug: string
  name_zh: string
  name_en: string
  description_zh: string
  description_en: string
  version: string
  platform: string
  file_name: string
  content_type: string
  size_bytes: number
  checksum_sha256: string
  published: boolean
  sort_order: number
  download_count: number
  uploaded_at: string
  created_at: string
  updated_at: string
}

export interface DownloadAuthorization {
  url: string
}

const downloadResourcesAPI = {
  async list(): Promise<DownloadResource[]> {
    const { data } = await apiClient.get<{ items: DownloadResource[] }>('/public/download-resources')
    return Array.isArray(data?.items) ? data.items : []
  },

  async authorizeDownload(id: number): Promise<DownloadAuthorization> {
    const { data } = await apiClient.post<DownloadAuthorization>(`/public/download-resources/${id}/download`)
    return data
  },
}

export default downloadResourcesAPI
