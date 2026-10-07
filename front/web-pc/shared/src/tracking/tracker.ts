/**
 * 用户端埋点（对齐 fgmm FgTracking / AppTracker）
 * POST /api/v1/tracking/events · 异步批量，不阻塞 UI
 */
import type { TrackEventBase, TrackEventItem, TrackerInitOptions } from './types'

const VISITOR_KEY = 'atm_track_visitor_id'
const SESSION_KEY = 'atm_track_session_id'
const TRACK_LOCALE = 'zh-Hans'

let opts: TrackerInitOptions = { baseURL: '', enabled: true }
const queue: TrackEventItem[] = []
let flushTimer: ReturnType<typeof setTimeout> | null = null
let lastPageEnterKey = ''
let lastPageEnterAt = 0
let currentPage = { page_id: '', page_path: '', page_title: '', referrer: '' }
let clickBound = false
/** 与 fgmm 一致：仅关联用户点击后 8s 内的接口，避免首屏自动 GET 刷屏 */
let lastClickContext: { element_id: string; element_name: string; at: number } | null =
  null
const API_TRACK_WINDOW_MS = 8000

function uuid(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  return `v-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

function visitorId(): string {
  try {
    let id = localStorage.getItem(VISITOR_KEY)
    if (!id) {
      id = uuid()
      localStorage.setItem(VISITOR_KEY, id)
    }
    return id
  } catch {
    return uuid()
  }
}

function sessionId(): string {
  try {
    let id = sessionStorage.getItem(SESSION_KEY)
    if (!id) {
      id = uuid()
      sessionStorage.setItem(SESSION_KEY, id)
    }
    return id
  } catch {
    return uuid()
  }
}

function locale(): string {
  return opts.locale?.() ?? TRACK_LOCALE
}

function channel(): string {
  return opts.channel ?? 'web'
}

function deviceContext(): Pick<
  TrackEventItem,
  'device_type' | 'screen_width' | 'screen_height'
> {
  if (typeof window === 'undefined') {
    return { device_type: 'unknown', screen_width: 0, screen_height: 0 }
  }
  const w = window.innerWidth || 0
  const h = window.innerHeight || 0
  let deviceType = 'desktop'
  if (w > 0 && w < 768) deviceType = 'mobile'
  else if (w < 1024) deviceType = 'tablet'
  return { device_type: deviceType, screen_width: w, screen_height: h }
}

function basePayload(): TrackEventBase {
  return {
    visitor_id: visitorId(),
    session_id: sessionId(),
    channel: channel(),
    app_version: opts.appVersion ?? '',
    locale: locale(),
    event_at: new Date().toISOString(),
    page_id: currentPage.page_id,
    page_path: currentPage.page_path,
    page_title: currentPage.page_title,
    referrer: currentPage.referrer,
    ...deviceContext(),
  }
}

function buildUrl(): string {
  const base = (opts.baseURL || '').replace(/\/$/, '')
  return `${base}/api/v1/tracking/events`
}

function authHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    Accept: 'application/json',
    'Content-Type': 'application/json',
    'Accept-Language': locale(),
  }
  const token = opts.getToken?.()
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }
  return headers
}

function sendPayload(events: TrackEventItem[]) {
  if (!opts.enabled || !events.length || typeof fetch === 'undefined') return
  const body = JSON.stringify({ events })
  const url = buildUrl()
  const hasAuth = Boolean(opts.getToken?.())
  if (!hasAuth && typeof navigator !== 'undefined' && navigator.sendBeacon) {
    try {
      const blob = new Blob([body], { type: 'application/json' })
      if (navigator.sendBeacon(url, blob)) return
    } catch {
      /* fetch fallback */
    }
  }
  void fetch(url, {
    method: 'POST',
    headers: authHeaders(),
    body,
    credentials: 'include',
    keepalive: true,
  }).catch(() => {})
}

function scheduleFlush() {
  if (flushTimer != null) return
  flushTimer = setTimeout(() => {
    flushTimer = null
    flush()
  }, 300)
}

function flush() {
  if (!queue.length) return
  const batch = queue.splice(0, 20)
  sendPayload(batch)
  if (queue.length) scheduleFlush()
}

function enqueue(event: TrackEventItem) {
  if (!opts.enabled) return
  queue.push(event)
  scheduleFlush()
}

export function initTracker(options: TrackerInitOptions) {
  opts = { enabled: true, ...options }
  if (typeof document !== 'undefined' && !clickBound) {
    clickBound = true
    document.addEventListener(
      'click',
      (ev) => {
        const target = ev.target as HTMLElement | null
        if (!target) return
        const el = target.closest(
          'button, a, [role="button"], [data-track-click], input[type="button"], input[type="submit"], .atm-btn-primary',
        ) as HTMLElement | null
        if (!el || el.closest('[data-track-ignore]')) return
        trackClick(el)
      },
      true,
    )
    window.addEventListener('pagehide', flush)
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'hidden') flush()
    })
  }
}

export function setCurrentPage(meta: {
  page_id: string
  page_path: string
  page_title: string
  referrer?: string
}) {
  currentPage = {
    page_id: meta.page_id,
    page_path: meta.page_path,
    page_title: meta.page_title,
    referrer: meta.referrer ?? (typeof document !== 'undefined' ? document.referrer : '') ?? '',
  }
}

export function pageEnter(extra?: Partial<TrackEventItem>) {
  const pageId = extra?.page_id ?? currentPage.page_id
  const pagePath = extra?.page_path ?? currentPage.page_path
  const key = `${pageId}|${pagePath}`
  const now = Date.now()
  if (key && key === lastPageEnterKey && now - lastPageEnterAt < 3000) {
    return
  }
  lastPageEnterKey = key
  lastPageEnterAt = now
  enqueue({
    ...basePayload(),
    event_type: 'page_view',
    action: 'enter',
    ...extra,
  })
}

export function click(elementId: string, elementName: string, targetUrl = '') {
  const id = elementId.slice(0, 128)
  const name = elementName.slice(0, 128)
  lastClickContext = { element_id: id, element_name: name, at: Date.now() }
  enqueue({
    ...basePayload(),
    event_type: 'click',
    action: 'click',
    element_id: id,
    element_name: name,
    target_url: targetUrl.slice(0, 512),
  })
}

function trackClick(el: HTMLElement) {
  const elementId =
    el.getAttribute('data-track-id') ||
    el.id ||
    el.getAttribute('name') ||
    el.className.split(/\s+/)[0] ||
    'click'
  const elementName =
    el.getAttribute('data-track-name') ||
    (el.textContent || '').trim().slice(0, 128) ||
    elementId
  let targetUrl = ''
  if (el instanceof HTMLAnchorElement && el.href) {
    targetUrl = el.href
  }
  click(elementId, elementName, targetUrl)
}

export function apiCall(method: string, apiPath: string, apiParams?: unknown) {
  if (apiPath.startsWith('/tracking/')) return
  if (
    !lastClickContext ||
    Date.now() - lastClickContext.at >= API_TRACK_WINDOW_MS
  ) {
    return
  }
  enqueue({
    ...basePayload(),
    event_type: 'click',
    action: 'click',
    element_id: lastClickContext.element_id,
    element_name: lastClickContext.element_name,
    api_method: method.toUpperCase(),
    api_path: apiPath,
    api_params: apiParams,
  })
}

export function flushTracking() {
  flush()
}
