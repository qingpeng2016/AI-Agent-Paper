/** 内存 + sessionStorage，用于 SWR：先展示缓存，后台再刷新 */

export function readSessionCache<T>(key: string): T | null {
  if (typeof sessionStorage === 'undefined') return null
  try {
    const raw = sessionStorage.getItem(key)
    if (!raw) return null
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

export function writeSessionCache(key: string, value: unknown): void {
  if (typeof sessionStorage === 'undefined') return
  try {
    sessionStorage.setItem(key, JSON.stringify(value))
  } catch {
    /* quota / private mode */
  }
}
