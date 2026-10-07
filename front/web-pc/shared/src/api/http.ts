import { apiCall as trackApiCall } from '../tracking/tracker'

export type HttpClientOptions = {
  baseURL: string
  getToken?: () => string | null
  /** 请求超时（毫秒），默认 12s */
  timeoutMs?: number
  /** 是否上报 API 埋点，默认 true */
  tracking?: boolean
}

const DEFAULT_TIMEOUT_MS = 12_000

function fetchWithTimeout(
  url: string,
  init: RequestInit,
  timeoutMs: number,
): Promise<Response> {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeoutMs)
  return fetch(url, { ...init, signal: controller.signal }).finally(() =>
    clearTimeout(timer),
  )
}

export function createHttpClient(options: HttpClientOptions) {
  const { baseURL, getToken, timeoutMs = DEFAULT_TIMEOUT_MS, tracking = true } = options

  async function request<T>(
    path: string,
    init: RequestInit & { json?: unknown } = {},
  ): Promise<T> {
    const headers = new Headers(init.headers)
    if (init.json !== undefined) {
      headers.set('Content-Type', 'application/json')
    }
    const token = getToken?.()
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }

    let res: Response
    try {
      res = await fetchWithTimeout(
        `${baseURL}${path}`,
        {
          ...init,
          headers,
          credentials: 'include',
          body: init.json !== undefined ? JSON.stringify(init.json) : init.body,
        },
        timeoutMs,
      )
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        throw new Error('请求超时，请检查后端是否已启动')
      }
      throw err
    }

    const text = await res.text()
    let parsed: unknown
    try {
      parsed = text ? JSON.parse(text) : null
    } catch {
      throw new Error(`Invalid JSON: ${text.slice(0, 200)}`)
    }

    const envelope = parsed as { code?: number | string; message?: string }
    if (!res.ok) {
      throw new Error(envelope.message ?? res.statusText)
    }
    const bizCode =
      envelope.code === undefined || envelope.code === null
        ? 200
        : Number(envelope.code)
    if (!Number.isNaN(bizCode) && bizCode !== 200) {
      throw new Error(envelope.message ?? `业务错误 ${bizCode}`)
    }
    if (tracking && !path.startsWith('/api/v1/tracking/')) {
      const method = (init.method || 'GET').toUpperCase()
      const apiParams =
        init.json !== undefined && typeof init.json === 'object'
          ? init.json
          : undefined
      trackApiCall(method, path, apiParams)
    }
    return parsed as T
  }

  return {
    get: <T>(path: string) => request<T>(path, { method: 'GET' }),
    post: <T>(path: string, json: unknown) => request<T>(path, { method: 'POST', json }),
    put: <T>(path: string, json: unknown) => request<T>(path, { method: 'PUT', json }),
    patch: <T>(path: string, json: unknown) => request<T>(path, { method: 'PATCH', json }),
    postForm: <T>(path: string, form: FormData) =>
      request<T>(path, { method: 'POST', body: form }),
  }
}
