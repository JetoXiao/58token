import { apiClient } from '../client'
import type { DownloadResource } from '../downloadResources'

export interface AdminDownloadResource extends DownloadResource {
  object_key: string
}

export type DownloadResourcePayload = Pick<AdminDownloadResource,
  'slug' | 'name_zh' | 'name_en' | 'description_zh' | 'description_en' | 'version' | 'platform' |
  'object_key' | 'file_name' | 'content_type' | 'checksum_sha256' | 'published' | 'sort_order'>

export interface DownloadResourceStorageConfig {
  endpoint: string
  region: string
  bucket: string
  access_key_id: string
  secret_access_key?: string
  prefix: string
  force_path_style: boolean
}

export interface DownloadResourceUploadURL {
  object_key: string
  upload_url: string
  expires_at: string
}

export interface DownloadRecord {
  id: number
  user_id?: number
  username?: string
  email?: string
  resource_id: number
  resource_name: string
  version: string
  ip: string
  user_agent: string
  referrer: string
  requested_at: string
  geo_country?: string
  geo_region?: string
  geo_city?: string
}

export interface DownloadRecordPage {
  items: DownloadRecord[]
  total: number
  page: number
  page_size: number
  pages: number
}

const downloadResourcesAPI = {
  async list(): Promise<AdminDownloadResource[]> {
    const { data } = await apiClient.get<{ items: AdminDownloadResource[] }>('/admin/download-resources')
    return Array.isArray(data?.items) ? data.items : []
  },
  async create(payload: DownloadResourcePayload): Promise<AdminDownloadResource> {
    const { data } = await apiClient.post<AdminDownloadResource>('/admin/download-resources', payload)
    return data
  },
  async update(id: number, payload: DownloadResourcePayload): Promise<AdminDownloadResource> {
    const { data } = await apiClient.put<AdminDownloadResource>(`/admin/download-resources/${id}`, payload)
    return data
  },
  async remove(id: number): Promise<void> {
    await apiClient.delete(`/admin/download-resources/${id}`)
  },
  async listDownloads(page = 1, pageSize = 20): Promise<DownloadRecordPage> {
    const { data } = await apiClient.get<DownloadRecordPage>('/admin/download-resources/downloads', { params: { page, page_size: pageSize } })
    return data
  },
  async lookupIP(ip: string): Promise<{ ip: string; country: string; region: string; city: string; country_code: string }> {
    const { data } = await apiClient.post('/admin/download-resources/ip-lookup', { ip })
    return data
  },
  async storage(): Promise<DownloadResourceStorageConfig | null> {
    const { data } = await apiClient.get<DownloadResourceStorageConfig | null>('/admin/download-resources/storage')
    return data
  },
  async saveStorage(payload: DownloadResourceStorageConfig): Promise<DownloadResourceStorageConfig> {
    const { data } = await apiClient.put<DownloadResourceStorageConfig>('/admin/download-resources/storage', payload)
    return data
  },
  async testStorage(payload: DownloadResourceStorageConfig): Promise<void> {
    await apiClient.post('/admin/download-resources/storage/test', payload)
  },
  async createUploadURL(fileName: string, contentType: string, sizeBytes: number): Promise<DownloadResourceUploadURL> {
    const { data } = await apiClient.post<DownloadResourceUploadURL>('/admin/download-resources/upload-url', {
      file_name: fileName,
      content_type: contentType,
      size_bytes: sizeBytes,
    })
    return data
  },
}

export default downloadResourcesAPI
