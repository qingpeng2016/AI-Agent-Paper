/** 与后端 domain/rest 对齐，随 API 演进 */

export interface UserProfile {
  id: number
  email?: string | null
  phone?: string | null
  nickname?: string | null
  status?: string
  wallet_balance?: number | string
  commission_balance?: number | string
  created_at?: string
  last_login_at?: string | null
}

export type UserWalletFlowItem = {
  id: number
  type: 'recharge' | 'pay' | 'refund' | 'commission' | 'withdraw' | string
  amount: number | string
  remark: string
  created_at: string
}

export type UserWalletFlowListPage = {
  items: UserWalletFlowItem[]
  total: number
  page: number
  page_size: number
}

export interface RegisterRequest {
  email?: string
  phone?: string
  password: string
  confirm_password: string
  /** 注册页浏览器 host，如 www.niceboxs.com */
  registration_host?: string
}

export interface LoginRequest {
  email?: string
  phone?: string
  password: string
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
  confirm_password: string
}

export interface LoginResponse {
  user: UserProfile
  token?: string
  register_coupons_granted?: number
}

/** 与 common/dederi/gin/response.ApiResp 一致 */
export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}
