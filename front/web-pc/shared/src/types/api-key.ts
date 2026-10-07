export type UserAPIKeyItem = {
  id: number
  key_type: string
  key_masked: string
  status: string
  user_subscription_id: number
  subscription_name: string
  limit_tokens: number
  used_tokens: number
  member_user_id?: number
  member_nickname?: string
  member_email?: string
  /** 仅创建子 Key 时返回一次 */
  api_key?: string
}

export type ApiTeamMemberItem = {
  id: number
  user_id: number
  nickname: string
  email: string
  status: string
}

export type ApiTeamAddableInviteeItem = {
  user_id: number
  nickname: string
  email: string
  registered_at: string
}

export type ApiTeamEnterpriseInquiryStatus = {
  has_inquiry: boolean
  inquiry?: {
    id: number
    company_name: string
    contact_name: string
    phone: string
    email?: string
    status: string
  }
}
