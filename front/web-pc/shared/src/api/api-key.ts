import type { ApiEnvelope } from '../types/user'
import type {
  ApiTeamAddableInviteeItem,
  ApiTeamEnterpriseInquiryStatus,
  ApiTeamMemberItem,
  UserAPIKeyItem,
} from '../types/api-key'
import { createHttpClient, type HttpClientOptions } from './http'

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createApiKeyApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async listMainKeys(): Promise<UserAPIKeyItem[]> {
      const res = await http.get<ApiEnvelope<UserAPIKeyItem[]>>('/api/v1/users/api-keys/main')
      return unwrap(res) ?? []
    },
    async listTeamKeys(): Promise<UserAPIKeyItem[]> {
      const res = await http.get<ApiEnvelope<UserAPIKeyItem[]>>('/api/v1/users/api-keys/team')
      return unwrap(res) ?? []
    },
    async createSubKey(body: {
      user_subscription_id: number
      member_user_id: number
      limit_tokens: number
    }): Promise<UserAPIKeyItem> {
      const res = await http.post<ApiEnvelope<UserAPIKeyItem>>('/api/v1/users/api-keys/sub', body)
      return unwrap(res)
    },
    async updateSubKeyLimit(id: number, limit_tokens: number): Promise<void> {
      await http.patch<ApiEnvelope<{ ok: boolean }>>(`/api/v1/users/api-keys/${id}/limit`, {
        limit_tokens,
      })
    },
    async listTeamMembers(): Promise<ApiTeamMemberItem[]> {
      const res = await http.get<ApiEnvelope<{ items: ApiTeamMemberItem[] }>>(
        '/api/v1/users/api-team/members',
      )
      return unwrap(res).items ?? []
    },
    async listAddableInvitees(): Promise<ApiTeamAddableInviteeItem[]> {
      const res = await http.get<ApiEnvelope<{ items: ApiTeamAddableInviteeItem[] }>>(
        '/api/v1/users/api-team/addable-invitees',
      )
      return unwrap(res).items ?? []
    },
    async addTeamMember(user_id: number): Promise<void> {
      const uid = Number(user_id)
      if (!Number.isFinite(uid) || uid <= 0) {
        throw new Error('请选择有效的成员')
      }
      await http.post<ApiEnvelope<{ ok: boolean }>>('/api/v1/users/api-team/members', {
        user_id: uid,
      })
    },
    async getEnterpriseInquiry(): Promise<ApiTeamEnterpriseInquiryStatus> {
      const res = await http.get<ApiEnvelope<ApiTeamEnterpriseInquiryStatus>>(
        '/api/v1/users/api-team/enterprise-inquiry',
      )
      return unwrap(res)
    },
    async submitEnterpriseInquiry(body: {
      company_name: string
    }): Promise<{ id: number; message?: string }> {
      const res = await http.post<ApiEnvelope<{ id: number; message?: string }>>(
        '/api/v1/users/api-team/enterprise-inquiry',
        body,
      )
      return unwrap(res)
    },
  }
}
