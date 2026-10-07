export type RegisterCouponPromo = {
  active: boolean
  title: string
  subtitle: string
  discount_label: string
  valid_days: number
  campaign_code: string
}

export type UserCouponItem = {
  id: number
  coupon_code: string
  campaign_code: string
  title: string
  discount_type: 'fixed_amount' | 'percent' | string
  discount_value: string
  min_order_amount: string
  status: 'available' | 'used' | 'expired' | string
  valid_from: string
  valid_until: string
}

export type UserCouponList = {
  items: UserCouponItem[]
}
