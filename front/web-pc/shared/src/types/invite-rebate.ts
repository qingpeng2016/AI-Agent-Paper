export type InviteRebateTier = {
  level_label: string
  min_valid_invites: number
  min_invitee_paid_amount: number | string
  rate_percent: number | string
  is_current: boolean
}

export type InviteRebateOverview = {
  promo_domain_url: string
  current_level_label: string
  current_rate_percent: number | string
  valid_invite_count: number
  invitee_paid_total: number | string
  commission_balance: number | string
  tiers: InviteRebateTier[]
  notes: string[]
}

export type InviteRebateMember = {
  id: number
  nickname: string
  email: string
  phone: string
  status: 'registered' | 'valid' | string
  registered_at: string
}

export type InviteRebateMemberList = {
  items: InviteRebateMember[]
  total: number
}

export type InviteCommissionRecord = {
  id: number
  invitee_nickname: string
  invitee_email: string
  order_no: string
  product_name: string
  order_amount: number | string
  rebate_amount: number | string
  created_at: string
}

export type InviteCommissionRecordListPage = {
  items: InviteCommissionRecord[]
  total: number
  page: number
  page_size: number
}

export type InviteWithdrawalRecord = {
  id: number
  amount: number | string
  channel: 'alipay' | 'wechat' | string
  status: 'pending' | 'completed' | 'failed' | string
  created_at: string
}

export type InviteWithdrawalListPage = {
  items: InviteWithdrawalRecord[]
  total: number
  page: number
  page_size: number
}

export type InvitePayoutConfig = {
  alipay_qr_data_url: string
  wechat_qr_data_url: string
  alipay_configured: boolean
  wechat_configured: boolean
}

export type CommissionTransferToBalanceResult = {
  transferred_amount: string
  commission_balance: string
  wallet_balance: string
}

export type CreateCommissionWithdrawalResult = {
  withdrawal: InviteWithdrawalRecord
  commission_balance: string
}
