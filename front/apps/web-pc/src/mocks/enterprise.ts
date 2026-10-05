export const enterpriseHero = {
  eyebrow: '团队 · 对公 · 开票',
  title: '企业采购与团队开通',
}

/** Hero 左侧企业通道要点 */
export const enterpriseChannelHighlights = [
  '席位清单 · CSV / 表格导入',
  '订单汇总 · 管理员会员中心',
  '续费提醒 · 到期前 7 天',
  '专属 1V1 · SVIP 客服对接群',
]

export const enterpriseStats = [
  { value: '500+', label: '服务企业 / 团队' },
  { value: '9 折起', label: '10 席以上批量优惠' },
  { value: '阶梯价', label: '席位越多单价越低' },
  { value: '1 对 1', label: '企业顾问跟进' },
]

export type EnterpriseBenefitIconName =
  | 'batch'
  | 'bank'
  | 'invoice'
  | 'account'
  | 'contract'
  | 'support'

export const enterpriseBenefits: {
  icon: EnterpriseBenefitIconName
  title: string
  desc: string
}[] = [
  {
    icon: 'batch',
    title: '批量开通',
    desc: '一次下单多个席位，按 SKU 分配至不同邮箱账号，避免个人重复报销。',
  },
  {
    icon: 'bank',
    title: '对公转账',
    desc: '提供对公账户与付款备注规则，到账后人工 + 系统自动履约，可配合 PO 流程。',
  },
  {
    icon: 'invoice',
    title: '增值税发票',
    desc: '下单勾选企业开票或走采购单，付款后在会员中心填写抬头与税号，电子普票 / 专票按政策开具。',
  },
  {
    icon: 'account',
    title: '充到本人账号',
    desc: '不代收密码。每个席位对应员工已注册账号，订阅周期与会员中心订单可对账。',
  },
  {
    icon: 'contract',
    title: '合同与资质',
    desc: '可提供平台服务说明、报价单与合作框架协议模板，满足内审与采购留档。',
  },
  {
    icon: 'support',
    title: '专属客服',
    desc: '企业微信 / 邮件专线，处理加急开通、续费提醒、席位变更与异常订单。',
  },
]

export const enterpriseSteps = [
  {
    step: '01',
    title: '提交需求',
    desc: '填写公司信息、预计席位数与目标产品（Plus / Pro / Claude / Cursor 等）。',
  },
  {
    step: '02',
    title: '确认方案',
    desc: '顾问 1 个工作日内回复报价、周期与开票方式，必要时出具报价单 PDF。',
  },
  {
    step: '03',
    title: '对公 / 在线支付',
    desc: '对公转账或聚合支付；大额可分期开通，到账后开始履约。',
  },
  {
    step: '04',
    title: '开通与开票',
    desc: '按清单充至各账号，订单汇总在管理员会员中心；开票信息在线提交。',
  },
]

export type EnterprisePlanTier = {
  id: string
  name: string
  badge?: string
  priceHint: string
  seats: string
  features: string[]
  cta: string
  featured?: boolean
}

export const enterprisePlans: EnterprisePlanTier[] = [
  {
    id: 'starter',
    name: '团队试用',
    priceHint: '¥5,000 起',
    seats: '3–10 席位',
    features: ['混合 SKU 组合', '在线支付为主', '电子普票', '标准客服响应'],
    cta: '适合小团队试点',
  },
  {
    id: 'business',
    name: '标准企业',
    badge: '常用',
    priceHint: '按 SKU 阶梯价',
    seats: '10–100 席位',
    features: ['对公转账 + 合同', 'Dedicated 顾问', '批量账号清单导入', '续费日历提醒'],
    cta: '适合部门统一采购',
    featured: true,
  },
  {
    id: 'custom',
    name: '定制方案',
    priceHint: '面议',
    seats: '100+ 或跨品牌',
    features: ['年度框架价', '多主体开票', '分阶段开通', 'SLA 与异常升级通道'],
    cta: '适合集团 / 多子公司',
  },
]

export const enterpriseInvoiceNotes = [
  '个人下单也可在支付时勾选「企业开票」，价格含约 6% 开票服务费，无需当场填抬头。',
  '付款成功后到会员中心「发票」填写单位名称、税号与邮箱，审核后电子发送。',
  '对公采购可走报价单 → 转账 → 批量开通，发票信息与合同主体保持一致。',
]

export const enterpriseFaqs = [
  {
    q: '一定要走对公吗？',
    a: '10 席位以内可用支付宝 / 微信批量下单；对公更适合有内审、合同与专票需求的企业。',
  },
  {
    q: '能否开增值税专用发票？',
    a: '在符合当地税务政策与平台资质前提下支持；提交需求时注明票种，顾问会确认材料与税率。',
  },
  {
    q: '账号归属算谁的？',
    a: '订阅充至员工自行注册并持有的 OpenAI / Anthropic / Cursor 等账号，企业仅持有采购与订单凭证。',
  },
  {
    q: '失败或封号怎么处理？',
    a: '按平台公示的退款与质保规则执行；企业单可集中由管理员会员中心查看订单状态并联系顾问。',
  },
]

export const enterpriseContact = {
  email: 'enterprise@aiplan.example',
  hours: '工作日 9:30–18:30（北京时间）',
  wechatHint: '提交表单后顾问会通过您留下的手机 / 邮箱联系，也可备注「需加企业微信」。',
}
