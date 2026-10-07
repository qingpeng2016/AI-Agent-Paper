/** 登录态 token（Cookie `atm_token`，由后端 /users/login Set-Cookie 写入） */
export const AUTH_TOKEN_COOKIE = 'atm_token'

const MAX_AGE_SEC = 60 * 60 * 24 * 30

export function getAuthToken(): string | null {
  if (typeof document === 'undefined') return null
  const match = document.cookie.match(
    new RegExp(`(?:^|;\\s*)${AUTH_TOKEN_COOKIE}=([^;]*)`),
  )
  return match ? decodeURIComponent(match[1]) : null
}

export function setAuthToken(value: string): void {
  if (!value) return
  document.cookie = `${AUTH_TOKEN_COOKIE}=${encodeURIComponent(value)}; Path=/; Max-Age=${MAX_AGE_SEC}; SameSite=Lax`
}

/** 退出登录：删除 Cookie 中的 token */
export function clearAuthToken(): void {
  document.cookie = `${AUTH_TOKEN_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`
}

export function hasAuthToken(): boolean {
  return Boolean(getAuthToken())
}
