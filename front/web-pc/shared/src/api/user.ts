import type {
  ApiEnvelope,
  ChangePasswordRequest,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  UserProfile,
  UserWalletFlowListPage,
} from '../types/user'
import { createHttpClient, type HttpClientOptions } from './http'

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createUserApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async register(body: RegisterRequest): Promise<LoginResponse> {
      const res = await http.post<ApiEnvelope<LoginResponse>>('/api/v1/users/register', body)
      return unwrap(res)
    },
    async login(body: LoginRequest): Promise<LoginResponse> {
      const res = await http.post<ApiEnvelope<LoginResponse>>('/api/v1/users/login', body)
      return unwrap(res)
    },
    async logout(): Promise<void> {
      await http.post<ApiEnvelope<null>>('/api/v1/users/logout', {})
    },
    async me(): Promise<UserProfile> {
      const res = await http.get<ApiEnvelope<UserProfile>>('/api/v1/users/me')
      return unwrap(res)
    },
    async changePassword(body: ChangePasswordRequest): Promise<void> {
      await http.post<ApiEnvelope<null>>('/api/v1/users/me/password', body)
    },
    async walletFlows(params?: { page?: number; page_size?: number }): Promise<UserWalletFlowListPage> {
      const search = new URLSearchParams()
      if (params?.page != null) search.set('page', String(params.page))
      if (params?.page_size != null) search.set('page_size', String(params.page_size))
      const qs = search.toString()
      const path = qs ? `/api/v1/users/wallet-flows?${qs}` : '/api/v1/users/wallet-flows'
      const res = await http.get<ApiEnvelope<UserWalletFlowListPage>>(path)
      return unwrap(res)
    },
  }
}
