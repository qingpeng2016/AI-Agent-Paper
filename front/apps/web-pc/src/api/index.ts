import {
  createEnterpriseApi,
  createInviteRebateApi,
  createInvoiceApi,
  createNotificationApi,
  createOrderApi,
  createSubscriptionApi,
  createProductApi,
  createTutorialApi,
  createUserApi,
  createApiKeyApi,
  createCouponApi,
} from '@ai-token-mall/shared'
import { getAuthToken } from '@/utils/auth-cookie'

const baseURL = import.meta.env.VITE_API_BASE_URL ?? ''

export const userApi = createUserApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const productApi = createProductApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const tutorialApi = createTutorialApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const enterpriseApi = createEnterpriseApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const orderApi = createOrderApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const subscriptionApi = createSubscriptionApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const invoiceApi = createInvoiceApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const notificationApi = createNotificationApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const inviteRebateApi = createInviteRebateApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const apiKeyApi = createApiKeyApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const couponApi = createCouponApi({
  baseURL,
  getToken: () => getAuthToken(),
})
