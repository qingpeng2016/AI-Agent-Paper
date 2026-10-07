import { createInviteRebateApi, createUserApi } from '@ai-token-mall/shared'
import { getAuthToken } from '@/utils/auth-cookie'

const baseURL = import.meta.env.VITE_API_BASE_URL ?? ''

export const userApi = createUserApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const inviteRebateApi = createInviteRebateApi({
  baseURL,
  getToken: () => getAuthToken(),
})
