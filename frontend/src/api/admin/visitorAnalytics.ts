import { apiClient } from '../client'

export interface VisitorAnalyticsOverview {
  page_views: number
  unique_visitors: number
  unique_ips: number
  active_channels: number
}

export interface VisitorTrendPoint {
  date: string
  page_views: number
  unique_visitors: number
  unique_ips: number
}

export interface VisitorChannel {
  id: number
  name: string
  code: string
  destination_path: string
  description: string
  active: boolean
  created_at: string
  updated_at: string
}

export interface VisitorChannelStats {
  id: number | null
  name: string
  code: string
  destination_path: string
  active: boolean
  page_views: number
  unique_visitors: number
  unique_ips: number
}

export interface VisitorEvent {
  id: number
  user_id?: number
  username?: string
  email?: string
  channel_name: string
  channel_code: string
  visitor_id: string
  session_id: string
  ip: string
  country_code: string
  path: string
  referrer: string
  landing_url: string
  user_agent: string
  language: string
  screen: string
  is_bot: boolean
  occurred_at: string
  geo_country: string
  geo_region: string
  geo_city: string
  geo_resolved_at?: string
}

export interface VisitorAnalyticsSettings {
  enabled: boolean
  retention_days: number
  updated_at: string
}

export interface IPGeolocation {
  ip: string
  country: string
  country_code: string
  region: string
  city: string
  timezone: string
  latitude?: number
  longitude?: number
  provider: string
  resolved_at: string
  expires_at: string
}

export interface VisitorDateParams {
  start_date: string
  end_date: string
}

export interface VisitorEventParams extends VisitorDateParams {
  page: number
  page_size: number
  channel_code?: string
  ip?: string
  search?: string
}

export interface PaginatedVisitorEvents {
  items: VisitorEvent[]
  total: number
  page: number
  page_size: number
  pages: number
}

export type VisitorChannelPayload = Pick<VisitorChannel, 'name' | 'code' | 'destination_path' | 'description' | 'active'>

const visitorAnalyticsAPI = {
  async overview(params: VisitorDateParams): Promise<VisitorAnalyticsOverview> {
    const { data } = await apiClient.get<VisitorAnalyticsOverview>('/admin/visitor-analytics/overview', { params })
    return data
  },
  async trend(params: VisitorDateParams): Promise<VisitorTrendPoint[]> {
    const { data } = await apiClient.get<{ items: VisitorTrendPoint[] }>('/admin/visitor-analytics/trend', { params })
    return data.items
  },
  async channelStats(params: VisitorDateParams): Promise<VisitorChannelStats[]> {
    const { data } = await apiClient.get<{ items: VisitorChannelStats[] }>('/admin/visitor-analytics/channel-stats', { params })
    return data.items
  },
  async events(params: VisitorEventParams): Promise<PaginatedVisitorEvents> {
    const { data } = await apiClient.get<PaginatedVisitorEvents>('/admin/visitor-analytics/events', { params })
    return data
  },
  async channels(): Promise<VisitorChannel[]> {
    const { data } = await apiClient.get<{ items: VisitorChannel[] }>('/admin/visitor-analytics/channels')
    return data.items
  },
  async createChannel(payload: VisitorChannelPayload): Promise<VisitorChannel> {
    const { data } = await apiClient.post<VisitorChannel>('/admin/visitor-analytics/channels', payload)
    return data
  },
  async updateChannel(id: number, payload: VisitorChannelPayload): Promise<VisitorChannel> {
    const { data } = await apiClient.put<VisitorChannel>(`/admin/visitor-analytics/channels/${id}`, payload)
    return data
  },
  async deleteChannel(id: number): Promise<void> {
    await apiClient.delete(`/admin/visitor-analytics/channels/${id}`)
  },
  async lookupIP(ip: string): Promise<IPGeolocation> {
    const { data } = await apiClient.post<IPGeolocation>('/admin/visitor-analytics/ip-lookup', { ip })
    return data
  },
  async settings(): Promise<VisitorAnalyticsSettings> {
    const { data } = await apiClient.get<VisitorAnalyticsSettings>('/admin/visitor-analytics/settings')
    return data
  },
  async updateSettings(payload: Pick<VisitorAnalyticsSettings, 'enabled' | 'retention_days'>): Promise<VisitorAnalyticsSettings> {
    const { data } = await apiClient.put<VisitorAnalyticsSettings>('/admin/visitor-analytics/settings', payload)
    return data
  },
}

export default visitorAnalyticsAPI
