import { ref } from 'vue'
import type { UserProfile } from '@ai-token-mall/shared'
import { clearAuthToken, getAuthToken, hasAuthToken } from '@/utils/auth-cookie'

const STORAGE_KEY = 'atm_user'
export const PENDING_BUY_KEY = 'atm_pending_buy'
/** 首页新人券弹窗：当前 SPA 会话内点「稍后再说/关闭」后不再展示（不用 sessionStorage，避免退出后仍被挡住） */
let registerCouponPromoDismissed = false

export function isRegisterCouponPromoDismissed(): boolean {
  return registerCouponPromoDismissed
}

export function dismissRegisterCouponPromoForSession(): void {
  registerCouponPromoDismissed = true
}

export function clearRegisterCouponPromoDismiss(): void {
  registerCouponPromoDismissed = false
}

/** 退出登录后递增，用于首页新人券弹窗强制重新挂载、重新拉取 */
export const guestPromoRemountKey = ref(0)

/** 登录态变更时递增，供首页 computed 刷新（localStorage 非响应式） */
export const authSessionRevision = ref(0)

function bumpAuthRevision() {
  authSessionRevision.value++
}

function readStoredProfile(): UserProfile | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const profile = JSON.parse(raw) as UserProfile
    return profile?.id ? profile : null
  } catch {
    return null
  }
}

export function getSessionUser(): UserProfile | null {
  // atm_token 可能为 HttpOnly，JS 读不到 Cookie 时仍以本地资料恢复登录态
  return readStoredProfile()
}

export function isLoggedIn(): boolean {
  return hasAuthToken() || readStoredProfile() != null
}

export function setSessionUser(profile: UserProfile): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(profile))
  bumpAuthRevision()
}

/** 退出登录：清空 Cookie 中的 atm_token，并移除本地用户信息 */
export function clearSessionUser(): void {
  clearAuthToken()
  localStorage.removeItem(STORAGE_KEY)
  clearRegisterCouponPromoDismiss()
  if (typeof sessionStorage !== 'undefined') {
    sessionStorage.removeItem('atm_register_coupon_promo_dismiss')
  }
  guestPromoRemountKey.value++
  bumpAuthRevision()
}

export function userAccountLabel(user: UserProfile): string {
  if (user.email) return user.email
  if (user.phone) return user.phone
  if (user.nickname) return user.nickname
  return `用户 #${user.id}`
}

/** 供调试：当前 cookie token */
export function sessionToken(): string | null {
  return getAuthToken()
}
