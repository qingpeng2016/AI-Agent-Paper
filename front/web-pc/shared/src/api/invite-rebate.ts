import type { ApiEnvelope } from '../types/user'
import type {
  InviteCommissionRecordListPage,
  InvitePayoutConfig,
  CommissionTransferToBalanceResult,
  CreateCommissionWithdrawalResult,
  InviteRebateMemberList,
  InviteRebateOverview,
  InviteWithdrawalListPage,
} from '../types/invite-rebate'
import { createHttpClient, type HttpClientOptions } from './http'

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createInviteRebateApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async overview(): Promise<InviteRebateOverview> {
      const res = await http.get<ApiEnvelope<InviteRebateOverview>>('/api/v1/users/invite-rebate/overview')
      return unwrap(res)
    },
    async members(): Promise<InviteRebateMemberList> {
      const res = await http.get<ApiEnvelope<InviteRebateMemberList>>('/api/v1/users/invite-rebate/members')
      return unwrap(res)
    },
    async commissionRecords(params?: { page?: number; page_size?: number }): Promise<InviteCommissionRecordListPage> {
      const search = new URLSearchParams()
      if (params?.page != null) search.set('page', String(params.page))
      if (params?.page_size != null) search.set('page_size', String(params.page_size))
      const qs = search.toString()
      const path = qs
        ? `/api/v1/users/invite-rebate/commission-records?${qs}`
        : '/api/v1/users/invite-rebate/commission-records'
      const res = await http.get<ApiEnvelope<InviteCommissionRecordListPage>>(path)
      return unwrap(res)
    },
    async createWithdrawal(body: {
      amount: string
      channel: 'alipay' | 'wechat'
    }): Promise<CreateCommissionWithdrawalResult> {
      const res = await http.post<ApiEnvelope<CreateCommissionWithdrawalResult>>(
        '/api/v1/users/invite-rebate/withdrawals',
        body,
      )
      return unwrap(res)
    },
    async withdrawals(params?: { page?: number; page_size?: number }): Promise<InviteWithdrawalListPage> {
      const search = new URLSearchParams()
      if (params?.page != null) search.set('page', String(params.page))
      if (params?.page_size != null) search.set('page_size', String(params.page_size))
      const qs = search.toString()
      const path = qs
        ? `/api/v1/users/invite-rebate/withdrawals?${qs}`
        : '/api/v1/users/invite-rebate/withdrawals'
      const res = await http.get<ApiEnvelope<InviteWithdrawalListPage>>(path)
      return unwrap(res)
    },
    async payoutConfig(): Promise<InvitePayoutConfig> {
      const res = await http.get<ApiEnvelope<InvitePayoutConfig>>('/api/v1/users/invite-rebate/payout-config')
      return unwrap(res)
    },
    async transferCommissionToBalance(amount: string): Promise<CommissionTransferToBalanceResult> {
      const res = await http.post<ApiEnvelope<CommissionTransferToBalanceResult>>(
        '/api/v1/users/invite-rebate/commission/transfer-to-balance',
        { amount },
      )
      return unwrap(res)
    },
    async uploadPayoutQr(channel: 'alipay' | 'wechat', file: File): Promise<InvitePayoutConfig> {
      const form = new FormData()
      form.append('channel', channel)
      form.append('file', file)
      const res = await http.postForm<ApiEnvelope<InvitePayoutConfig>>(
        '/api/v1/users/invite-rebate/payout-config/upload',
        form,
      )
      return unwrap(res)
    },
  }
}
