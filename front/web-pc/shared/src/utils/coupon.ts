import type { UserCouponItem } from '../types/coupon'

function parseYmdStartOfDay(ymd: string): Date | null {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(ymd.trim())
  if (!m) return null
  return new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]), 0, 0, 0, 0)
}

function parseYmdEndOfDay(ymd: string): Date | null {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(ymd.trim())
  if (!m) return null
  return new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]), 23, 59, 59, 999)
}

/** 当前订单小计下是否可用（与后端 ResolveForCheckout 规则对齐） */
export function isCouponEligibleForSubtotal(c: UserCouponItem, subtotal: number, now = new Date()): boolean {
  if (c.status !== 'available') return false
  const min = parseFloat(c.min_order_amount)
  if (!Number.isFinite(subtotal) || subtotal < min) return false
  const from = parseYmdStartOfDay(c.valid_from)
  const until = parseYmdEndOfDay(c.valid_until)
  if (from && now < from) return false
  if (until && now > until) return false
  return couponDiscountAmount(c, subtotal) > 0
}

export function couponDiscountAmount(c: UserCouponItem, subtotal: number): number {
  if (!Number.isFinite(subtotal) || subtotal <= 0) return 0
  const min = parseFloat(c.min_order_amount)
  if (subtotal < min) return 0
  let off = 0
  if (c.discount_type === 'percent') {
    off = (subtotal * parseFloat(c.discount_value)) / 100
  } else {
    off = parseFloat(c.discount_value)
  }
  if (!Number.isFinite(off) || off <= 0) return 0
  off = Math.round(off * 100) / 100
  return Math.min(off, subtotal)
}

export function pickBestCoupon(coupons: UserCouponItem[], subtotal: number): UserCouponItem | null {
  let best: UserCouponItem | null = null
  let bestOff = 0
  for (const c of coupons) {
    if (!isCouponEligibleForSubtotal(c, subtotal)) continue
    const off = couponDiscountAmount(c, subtotal)
    if (off > bestOff) {
      bestOff = off
      best = c
    }
  }
  return best
}

export function couponOptionLabel(c: UserCouponItem, subtotal: number): string {
  const title = c.title || c.campaign_code || '优惠券'
  const off = couponDiscountAmount(c, subtotal)
  if (c.discount_type === 'percent') {
    return `${title} · ${parseFloat(c.discount_value)}% 减 ¥${off.toFixed(2)}`
  }
  return `${title} · 减 ¥${off.toFixed(2)}`
}
