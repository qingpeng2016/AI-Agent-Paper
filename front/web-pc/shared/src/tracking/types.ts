export type TrackEventBase = {
  visitor_id?: string
  session_id?: string
  channel?: string
  app_version?: string
  page_id?: string
  page_path?: string
  page_title?: string
  element_id?: string
  element_name?: string
  target_url?: string
  api_method?: string
  api_path?: string
  api_params?: unknown
  locale?: string
  device_type?: string
  device_model?: string
  os_name?: string
  os_version?: string
  screen_width?: number
  screen_height?: number
  referrer?: string
  extra?: unknown
  event_at?: string
}

export type TrackEventItem = TrackEventBase & {
  event_type: 'page_view' | 'click' | string
  action: 'enter' | 'click' | string
}

export type ReportTrackEventsBody = {
  events: TrackEventItem[]
}

export type ReportTrackEventsResult = {
  accepted: number
}

export type TrackerInitOptions = {
  baseURL: string
  getToken?: () => string | null
  enabled?: boolean
  channel?: string
  appVersion?: string
  locale?: () => string
}
