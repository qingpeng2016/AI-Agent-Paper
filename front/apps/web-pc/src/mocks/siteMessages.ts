import type { MemberTab } from '@/mocks/member'

export type SiteMessageCategory = 'order' | 'subscription' | 'billing' | 'system'

export type SiteMessage = {
  id: number
  category: SiteMessageCategory
  title: string
  body: string
  createdAt: string
  readAt: string | null
  /** 点击后跳转会员中心 Tab（演示） */
  linkTab?: MemberTab
}

export const siteMessageCategoryLabel: Record<SiteMessageCategory, string> = {
  order: '订单',
  subscription: '套餐',
  billing: '发票/资金',
  system: '系统',
}

/** 站内消息演示数据（未接 API） */
export const mockSiteMessages: SiteMessage[] = [
  {
    id: 1,
    category: 'order',
    title: '订单支付成功',
    body: '订单 TST20260930001 已支付 ¥299.00，套餐资源正在开通，请稍后在「我的套餐」查看。',
    createdAt: '2026-09-30 11:20',
    readAt: null,
    linkTab: 'orders',
  },
  {
    id: 2,
    category: 'subscription',
    title: 'API Key 已下发',
    body: '您的 Claude 中转套餐已激活，主 Key 可在「API Key · 我的 Key」中复制使用。',
    createdAt: '2026-09-30 11:21',
    readAt: null,
    linkTab: 'api-keys',
  },
  {
    id: 3,
    category: 'subscription',
    title: '套餐即将到期',
    body: '「GPT 中转 · 专业版」将于 3 天后到期，建议提前续费以免影响调用。',
    createdAt: '2026-09-29 09:00',
    readAt: null,
    linkTab: 'plans',
  },
  {
    id: 4,
    category: 'billing',
    title: '企业发票待补充信息',
    body: '订单 TST20260928008 已标记需要企业开票，请完善发票抬头后联系客服或等待开具。',
    createdAt: '2026-09-28 16:40',
    readAt: '2026-09-28 18:02',
    linkTab: 'invoices',
  },
  {
    id: 5,
    category: 'billing',
    title: '余额支付成功',
    body: '已从账户余额扣款 ¥50.00 完成套餐续费，可在「资金流水」查看明细。',
    createdAt: '2026-09-27 14:15',
    readAt: '2026-09-27 14:16',
    linkTab: 'account',
  },
  {
    id: 6,
    category: 'system',
    title: '密码修改成功',
    body: '您的登录密码已于 2026-09-26 20:10 修改。如非本人操作，请立即联系客服。',
    createdAt: '2026-09-26 20:10',
    readAt: '2026-09-26 20:11',
    linkTab: 'settings',
  },
  {
    id: 7,
    category: 'order',
    title: '订单已取消',
    body: '订单 TST20260925003 超时未支付，已自动关闭。如需购买请重新下单。',
    createdAt: '2026-09-25 12:00',
    readAt: '2026-09-25 12:05',
    linkTab: 'orders',
  },
  {
    id: 8,
    category: 'system',
    title: '欢迎使用 AI Token Mall',
    body: '完成首次充值或购买套餐后，即可在会员中心管理 Key、订单与发票。祝使用愉快。',
    createdAt: '2026-09-20 10:00',
    readAt: '2026-09-20 10:01',
    linkTab: 'overview',
  },
]
