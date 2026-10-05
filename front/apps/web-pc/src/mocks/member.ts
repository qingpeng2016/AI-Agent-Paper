/** 会员中心 UI mock（待接 API） */

export type MemberTab =
  | 'overview'
  | 'messages'
  | 'plans'
  | 'api-keys'
  | 'orders'
  | 'coupons'
  | 'account'
  | 'invoices'
  | 'sub-accounts'
  | 'settings'

export const memberNav: { id: MemberTab; label: string }[] = [
  { id: 'overview', label: '概览' },
  { id: 'coupons', label: '优惠券' },
  { id: 'plans', label: '我的套餐' },
  { id: 'api-keys', label: 'API Key' },
  { id: 'sub-accounts', label: '邀请返利' },
  { id: 'orders', label: '订单记录' },
  { id: 'account', label: '资金流水' },
  { id: 'invoices', label: '发票管理' },
  { id: 'settings', label: '账户设置' },
]

export type MockSubscription = {
  id: number
  /** 对应 products.id，续费下单用 */
  productId: number
  productName: string
  skuLabel: string
  status: 'active' | 'expired' | 'suspended'
  limitTokens: number
  usedTokens: number
  expiresAt: string
  periodEnd: string
  /** 该套餐专属 API Key（展示用脱敏；复制为演示完整值） */
  apiKeyMasked: string
  apiKeyCopyValue: string
  apiKeyLastUsedAt: string | null
}

export type MockOrder = {
  orderNo: string
  productName: string
  quantity: number
  totalAmount: number
  status: 'pending_payment' | 'completed' | 'failed' | 'cancelled'
  createdAt: string
  enterpriseInvoice: boolean
}

export type MockWalletTx = {
  id: string
  type: 'recharge' | 'pay' | 'refund'
  amount: number
  remark: string
  createdAt: string
}

export type MockInvoice = {
  id: number
  orderNo: string
  title: string
  amount: number
  status: 'pending' | 'issued' | 'failed'
  createdAt: string
}

export type MockSubAccount = {
  id: number
  nickname: string
  email: string
  usedTokens: number
  limitTokens: number
  status: 'active' | 'disabled'
}

/** 分配给团队成员的子 Key（从主套餐 Key 派生，单独计费到该成员） */
export type MockTeamSubKey = {
  id: number
  subscriptionId: number
  subscriptionName: string
  memberId: number
  memberNickname: string
  memberEmail: string
  apiKeyMasked: string
  apiKeyCopyValue: string
  /** 本周期已用 tokens */
  usedTokens: number
  /** 子 Key 用量上限，默认与关联主 Key 套餐总量一致 */
  limitTokens: number
  status: 'active' | 'revoked'
  createdAt: string
}

export type MockInviteRebateRecord = {
  id: number
  inviteeNickname: string
  inviteeEmail: string
  orderNo: string
  productName: string
  orderAmount: number
  /** 返利金额（元），已计入佣金 */
  rebateAmount: number
  createdAt: string
}

export const mockMemberOverview = {
  balance: 128,
  /** 邀请返利累计佣金（元，可提现 / 抵扣，演示） */
  commission: 35.6,
}

export type MockWithdrawalRecord = {
  id: number
  amount: number
  channel: 'alipay' | 'wechat'
  status: 'pending' | 'completed' | 'failed'
  createdAt: string
}

export const mockWithdrawalRecords: MockWithdrawalRecord[] = [
  {
    id: 1,
    amount: 20,
    channel: 'alipay',
    status: 'completed',
    createdAt: '2026-03-20 16:08',
  },
]

export const withdrawalStatusLabel: Record<MockWithdrawalRecord['status'], string> = {
  pending: '处理中',
  completed: '已到账',
  failed: '失败',
}

export const withdrawalChannelLabel: Record<MockWithdrawalRecord['channel'], string> = {
  alipay: '支付宝',
  wechat: '微信',
}

export type MockInviteRebateTier = {
  levelLabel: string
  minInvites: number
  ratePercent: number
}

export type MockInviteRebatePolicy = {
  currentLevelLabel: string
  currentRatePercent: number
  validInviteCount: number
  nextLevelLabel: string | null
  nextLevelRatePercent: number | null
  invitesToNextLevel: number | null
  tiers: MockInviteRebateTier[]
  notes: string[]
}

/** 专属推广域名根地址（演示，接 API 后按用户下发） */
export const mockPromoDomainBase = 'https://go.aiplan.com/i'

/** 当前登录用户的返佣政策（演示） */
export const mockInviteRebatePolicy: MockInviteRebatePolicy = {
  currentLevelLabel: '标准推广',
  currentRatePercent: 5,
  validInviteCount: 3,
  nextLevelLabel: '高级推广',
  nextLevelRatePercent: 8,
  invitesToNextLevel: 2,
  tiers: [
    { levelLabel: '入门推广', minInvites: 0, ratePercent: 3 },
    { levelLabel: '标准推广', minInvites: 3, ratePercent: 5 },
    { levelLabel: '高级推广', minInvites: 5, ratePercent: 8 },
    { levelLabel: '合伙人', minInvites: 20, ratePercent: 12 },
  ],
  notes: [
    '返佣比例按「邀请下级人数」自动升级，下级为注册时已绑定到您账号的用户。',
    '受邀用户每笔已支付订单，按实付金额 × 当前返佣比例计算返利，支付成功后即时计入「佣金」。',
    '佣金可划转到余额或者提现。',
  ],
}

export const mockInviteRebateRecords: MockInviteRebateRecord[] = [
  {
    id: 1,
    inviteeNickname: '同事 A',
    inviteeEmail: 'teama@example.com',
    orderNo: 'AP202603270001',
    productName: 'ChatGPT Plus 月卡',
    orderAmount: 178,
    rebateAmount: 8.9,
    createdAt: '2026-03-27 14:25',
  },
  {
    id: 2,
    inviteeNickname: '新用户 B',
    inviteeEmail: 'userb@example.com',
    orderNo: 'AP202603200015',
    productName: 'GPT Go 月卡',
    orderAmount: 89,
    rebateAmount: 4.45,
    createdAt: '2026-03-20 11:02',
  },
  {
    id: 3,
    inviteeNickname: '新用户 B',
    inviteeEmail: 'userb@example.com',
    orderNo: 'AP202603150032',
    productName: 'Claude Pro 月卡',
    orderAmount: 219,
    rebateAmount: 10.95,
    createdAt: '2026-03-15 09:12',
  },
]

export const mockSubscriptions: MockSubscription[] = [
  {
    id: 1,
    productId: 10,
    productName: 'ChatGPT Plus 月卡',
    skuLabel: 'OAI-PLUS-M',
    status: 'active',
    limitTokens: 8_000_000,
    usedTokens: 3_240_000,
    expiresAt: '2026-04-26',
    periodEnd: '2026-04-01',
    apiKeyMasked: 'ap_live_8f2a••••••7k9m',
    apiKeyCopyValue: 'ap_live_8f2a9b3c7d4e5f6g7h9m',
    apiKeyLastUsedAt: '2026-03-28 09:12',
  },
  {
    id: 2,
    productId: 3,
    productName: 'Claude Pro 月卡',
    skuLabel: 'ANT-PRO-M',
    status: 'active',
    limitTokens: 8_000_000,
    usedTokens: 1_100_000,
    expiresAt: '2026-04-10',
    periodEnd: '2026-04-01',
    apiKeyMasked: 'ap_live_3c9b••••••2p4q',
    apiKeyCopyValue: 'ap_live_3c9b1a2b3c4d5e6f2p4q',
    apiKeyLastUsedAt: '2026-03-27 18:40',
  },
]

export const mockOrders: MockOrder[] = [
  {
    orderNo: 'AP202603270001',
    productName: 'ChatGPT Plus 月卡',
    quantity: 1,
    totalAmount: 178,
    status: 'completed',
    createdAt: '2026-03-27 14:20',
    enterpriseInvoice: false,
  },
  {
    orderNo: 'AP202603150032',
    productName: 'Claude Pro 月卡',
    quantity: 1,
    totalAmount: 219,
    status: 'completed',
    createdAt: '2026-03-15 09:08',
    enterpriseInvoice: true,
  },
  {
    orderNo: 'AP202603280008',
    productName: 'GPT Go 月卡',
    quantity: 1,
    totalAmount: 89,
    status: 'pending_payment',
    createdAt: '2026-03-28 10:00',
    enterpriseInvoice: false,
  },
]

export const mockWalletTx: MockWalletTx[] = [
  {
    id: 'tx1',
    type: 'recharge',
    amount: 500,
    remark: '支付宝充值',
    createdAt: '2026-03-10 11:30',
  },
  {
    id: 'tx2',
    type: 'pay',
    amount: -178,
    remark: '订单 AP202603270001',
    createdAt: '2026-03-27 14:21',
  },
  {
    id: 'tx3',
    type: 'pay',
    amount: -219,
    remark: '订单 AP202603150032',
    createdAt: '2026-03-15 09:09',
  },
]

export const mockInvoices: MockInvoice[] = [
  {
    id: 1,
    orderNo: 'AP202603150032',
    title: '某某科技有限公司',
    amount: 232.14,
    status: 'issued',
    createdAt: '2026-03-16',
  },
]

/** 通过邀请链接完成注册的用户（可加入 API 团队） */
export type MockInvitedUser = {
  id: number
  nickname: string
  email: string
  registeredAt: string
}

export const mockInvitedUsers: MockInvitedUser[] = [
  {
    id: 101,
    nickname: '同事 A',
    email: 'teama@example.com',
    registeredAt: '2026-03-20',
  },
  {
    id: 102,
    nickname: '新用户 B',
    email: 'userb@example.com',
    registeredAt: '2026-03-18',
  },
  {
    id: 103,
    nickname: '待加入 C',
    email: 'userc@example.com',
    registeredAt: '2026-03-29',
  },
]

/** API Key 团队初始成员（演示） */
export const mockApiTeamMembers: MockSubAccount[] = [
  {
    id: 101,
    nickname: '同事 A',
    email: 'teama@example.com',
    usedTokens: 420_000,
    limitTokens: 2_000_000,
    status: 'active',
  },
]

/** @deprecated 演示兼容，请用 mockInvitedUsers / mockApiTeamMembers */
export const mockSubAccounts: MockSubAccount[] = mockApiTeamMembers

export const mockTeamSubKeys: MockTeamSubKey[] = [
  {
    id: 1,
    subscriptionId: 1,
    subscriptionName: 'ChatGPT Plus 月卡',
    memberId: 101,
    memberNickname: '同事 A',
    memberEmail: 'teama@example.com',
    apiKeyMasked: 'ap_sub_7k2m••••••x8p1',
    apiKeyCopyValue: 'ap_sub_7k2m4n5p6q7r8s9tx8p1',
    usedTokens: 420_000,
    limitTokens: 8_000_000,
    status: 'active',
    createdAt: '2026-03-28',
  },
]

export const orderStatusLabel: Record<MockOrder['status'], string> = {
  pending_payment: '待支付',
  completed: '已完成',
  failed: '失败',
  cancelled: '已取消',
}

export function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`
  return String(n)
}
