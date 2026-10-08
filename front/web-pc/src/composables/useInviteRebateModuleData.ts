import { reactive, ref } from 'vue'
import {
  createInviteRebateApi,
  parseMoney,
  type InvitePayoutConfig,
  type InviteRebateOverview,
} from '@ai-agent-paper/shared'
import { getAuthToken } from '@/utils/auth-cookie'

const baseURL = import.meta.env.VITE_API_BASE_URL ?? ''

export const inviteRebateApi = createInviteRebateApi({
  baseURL,
  getToken: () => getAuthToken(),
})

export const inviteRebateOverview = ref<InviteRebateOverview | null>(null)
export const inviteRebateCommissionAvailable = ref(0)
export const inviteRebatePayoutQr = reactive({ alipay: '', wechat: '' })
export const inviteRebatePayoutQrConfigured = reactive({ alipay: false, wechat: false })
export const inviteRebateOverviewLoaded = ref(false)

export function applyInviteRebatePayoutConfig(payout: InvitePayoutConfig) {
  inviteRebatePayoutQr.alipay = payout.alipay_qr_data_url ?? ''
  inviteRebatePayoutQr.wechat = payout.wechat_qr_data_url ?? ''
  inviteRebatePayoutQrConfigured.alipay = Boolean(
    payout.alipay_configured ?? inviteRebatePayoutQr.alipay,
  )
  inviteRebatePayoutQrConfigured.wechat = Boolean(
    payout.wechat_configured ?? inviteRebatePayoutQr.wechat,
  )
}

let reloadSeq = 0

/** 侧栏切到「邀请返利」时重新拉 overview + 收款码配置 */
export async function reloadInviteRebateModuleData(): Promise<void> {
  const seq = ++reloadSeq
  inviteRebateOverviewLoaded.value = false
  const ov = await inviteRebateApi.overview()
  if (seq !== reloadSeq) return
  inviteRebateOverview.value = ov
  inviteRebateCommissionAvailable.value = parseMoney(ov.commission_balance)
  const payout = await inviteRebateApi.payoutConfig()
  if (seq !== reloadSeq) return
  applyInviteRebatePayoutConfig(payout)
  inviteRebateOverviewLoaded.value = true
}
