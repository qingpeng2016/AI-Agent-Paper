import { appendShareFeature, type CatalogProduct, type UpstreamName } from '@ai-token-mall/shared'

export type { CatalogProduct }
export { heroSubKeyShareLine, subKeyShareFeature } from '@ai-token-mall/shared'

/** 全部套餐 — 胶囊筛选（mock 回退；线上以接口 categories[].id / name 为准） */
export const catalogFilterPills: {
  value: 'all' | string
  label: string
  dot?: string
  activeBg?: string
}[] = [
  { value: 'all', label: '全部' },
  { value: '1', label: 'ChatGPT', dot: '#10a37f', activeBg: '#10a37f' },
  { value: '2', label: 'Claude', dot: '#d97757', activeBg: '#d97757' },
  { value: '3', label: 'Grok', dot: '#0f172a', activeBg: '#334155' },
  { value: '4', label: 'Gemini', dot: '#4285f4', activeBg: '#4285f4' },
  { value: '5', label: 'Cursor', dot: '#7c3aed', activeBg: '#7c3aed' },
  { value: '6', label: 'Perplexity', dot: '#0d9488', activeBg: '#0d9488' },
]

export const mockCategorySlugById: Record<number, string> = {
  1: 'openai',
  2: 'anthropic',
  3: 'xai',
  4: 'gemini',
  5: 'cursor',
  6: 'perplexity',
}

export const heroQuickLinks = [
  { name: 'ChatGPT', price: '¥89', upstream: 'openai' as UpstreamName },
  { name: 'Claude', price: '¥219', upstream: 'anthropic' as UpstreamName },
  { name: 'Gemini', price: '¥189', upstream: 'gemini' as UpstreamName },
  { name: 'Grok', price: '¥790', upstream: 'xai' as UpstreamName },
  { name: 'Cursor', price: '¥198', upstream: 'openai' as UpstreamName, note: 'BYOK 教程' },
  { name: 'Perplexity', price: '¥119', upstream: 'perplexity' as UpstreamName },
]

export const trustStats = [
  { value: '低于官方', label: '同档套餐单价' },
  { value: '支持子 Key', label: '多人共用' },
  { value: '100%', label: '稳定性' },
  { value: '客服', label: '在线客服' },
]

export const heroChecklist = [
  '同档服务 · 低于官网公开价',
  '支付宝 / 微信快捷支付',
  '无需海外信用卡',
  '约 1 分钟自助开通',
  '失败 100% 退款',
]

export const compareSection = {
  title: '我们的能力',
  mallName: 'AI Plan',
  diyName: '官网直购',
  mallCta: '支付宝 / 微信 · 低价开通',
  diyCta: '查看全部套餐 →',
}

export const compareRows = [
  {
    icon: 'pay' as const,
    label: '支付方式',
    us: '支付宝 / 微信 国内直付',
    them: '海外信用卡 / 复杂绑卡',
  },
  {
    icon: 'chip' as const,
    label: '调用单价',
    us: '同模型档位 · 低于官网公开价',
    them: '美元标价 + 汇率与支付附加成本',
  },
  {
    icon: 'terminal' as const,
    label: '使用方式',
    us: '下单即用，Cursor / Claude Code 等按教程配置',
    them: '多平台账号与网络需自行解决',
  },
  {
    icon: 'shield' as const,
    label: '额度周期',
    us: '套餐额度明码标注，订阅期内可续费',
    them: '各平台计费规则分散难估',
  },
  {
    icon: 'devices' as const,
    label: '用量可见',
    us: '会员中心查看剩余额度，支持子账号分账',
    them: '官方后台分散、难汇总',
  },
  {
    icon: 'invoice' as const,
    label: '企业报销',
    us: '可开发票 · 对公采购通道',
    them: '多数不支持国内专票',
  },
  {
    icon: 'support' as const,
    label: '售后客服',
    us: '中文客服 · 快速响应',
    them: '英文邮件 · 响应慢',
  },
]

export const brandCards = [
  {
    key: 'openai',
    icon: 'G',
    title: 'ChatGPT 通道',
    accent: '#10a37f',
    tags: ['GPT Go', 'Plus 月卡', 'Pro 5X / 20X', 'Codex'],
    slogan: '最热门之选',
  },
  {
    key: 'anthropic',
    icon: 'C',
    title: 'Claude 通道',
    accent: '#d97757',
    tags: ['Claude Pro', 'Max 5X', 'Max 20X', 'Claude Code'],
    slogan: '编程 · 长文首选',
  },
  {
    key: 'xai',
    icon: 'X',
    title: 'Grok 通道',
    accent: '#0f172a',
    tags: ['SuperGrok', 'Heavy', 'Imagine'],
    slogan: '实时联网创意',
  },
  {
    key: 'gemini',
    icon: '◆',
    title: 'Gemini 通道',
    accent: '#4285f4',
    tags: ['AI Pro', 'Ultra', 'Nano / Veo'],
    slogan: '谷歌多模态',
  },
  {
    key: 'cursor',
    icon: '>_',
    title: 'Cursor 配置',
    accent: '#7c3aed',
    tags: ['Pro 教程', 'Pro+', 'Cloud Agents'],
    slogan: 'BYOK 自定义模型',
  },
  {
    key: 'perplexity',
    icon: 'P',
    title: 'Perplexity',
    accent: '#0d9488',
    tags: ['Pro', 'Pro Search', '多模型'],
    slogan: '搜索研究首选',
  },
]

export const reviewSummary = {
  score: 5,
  count: 2539,
}

export const reviews = [
  {
    stars: 5,
    text: '对比官网账单，同档 Pro 这边单价明显更低，支付宝付完一分钟就能用，不用折腾海外卡。',
    user: '林**',
    sku: 'Pro 20X 月卡',
    date: '2026-06-16',
  },
  {
    stars: 5,
    text: '团队几个人一起用，子账号分账清楚；比自己在官网买省一截，剩余额度在会员中心随时看。',
    user: '周**',
    sku: 'Plus 月卡',
    date: '2026-06-10',
  },
  {
    stars: 5,
    text: '重度写代码额度跟得上，价格比官方公开价划算，超额了提示也清楚。',
    user: '高*',
    sku: 'Pro 5X 月卡',
    date: '2026-06-19',
  },
  {
    stars: 5,
    text: '企业批量买了席位，发票流程顺；同样模型档位，财务对比过比走美元官网省。',
    user: '许**',
    sku: '企业采购',
    date: '2026-06-05',
  },
  {
    stars: 5,
    text: '支持很快就到账了，全程丝滑好用',
    user: '罗**',
    sku: 'GPT Go',
    date: '2026-06-13',
  },
  {
    stars: 5,
    text: '标价就是到手价，没有隐藏费用；Claude Code 按教程配好就能用，已推荐给组里。',
    user: '叶*',
    sku: 'Claude Pro',
    date: '2026-06-21',
  },
]

export const faqs = [
  {
    q: '为什么说低于官方？',
    a: '同款模型与额度档位，我们的套餐标价通常低于你在官网直购的等效花费（不含海外绑卡、汇率等成本）。页面价格即到手价，以各 SKU 说明为准。',
  },
  {
    q: '没有海外信用卡，能买吗？',
    a: '可以。支持支付宝 / 微信；购买的是本平台的套餐与使用额度，自助开通后在会员中心按教程使用即可。',
  },
  {
    q: '和官网买有什么区别？',
    a: '主要差别在国内支付更省事、价格更透明，会员中心可统一查看额度与续费；具体可用模型与额度以所购套餐为准。',
  },
  {
    q: '多久到账？失败怎么办？',
    a: '支付成功通常约 1 分钟内开通。若失败会原路退款，不会重复扣款，有疑问可联系右下角客服。',
  },
  {
    q: 'Go、Plus、Pro 20X 怎么选？',
    a: '轻量使用选 Go；日常选 Plus；高频或团队选 Pro 5X / 20X。各卡片上写了月度额度与适用场景。',
  },
  {
    q: '支持团队分账吗？',
    a: '支持。主账号购买后可在会员中心开通子账号，共享套餐额度并单独查看用量。',
  },
  {
    q: '企业发票怎么开？',
    a: '下单可勾选企业开票；付款后在会员中心填写发票抬头与税号。',
  },
]

const mockProductsRaw: CatalogProduct[] = [
  {
    id: 1,
    sku_code: 'OAI-GO-M',
    card_title: 'ChatGPT Go 月卡',
    card_subtitle: '入门档额度，日常对话与轻量开发够用。',
    card_features: ['{limit_tokens} tokens / 月', 'gpt-4o-mini 可用', 'ChatGPT 线路', '约 1 分钟开通'],
    share_seats: 3,
    products_category_name: 'openai',
    sku_product_name: 'GPT GO 5X',
    limit_tokens: 2_000_000,
    rpm_limit: 60,
    allowed_models: ['gpt-4o-mini', 'gpt-4o'],
    billing_period: 'month',
    period_days: 30,
    price: 89,
    currency: 'CNY',
    sort: 10,
    status: 'on_sale',
  },
  {
    id: 9,
    sku_code: 'PPX-PRO-M',
    card_title: 'Perplexity Pro 月卡',
    card_subtitle: 'Pro Search + 多模型检索，研究类 Agent 首选。',
    card_features: ['{limit_tokens} tokens / 月', 'sonar / sonar-pro', 'Pro Search', '低价研究首选'],
    share_seats: 3,
    products_category_name: 'perplexity',
    sku_product_name: 'Perplexity Pro',
    limit_tokens: 5_000_000,
    rpm_limit: 50,
    allowed_models: ['sonar', 'sonar-pro'],
    billing_period: 'month',
    period_days: 30,
    price: 119,
    currency: 'CNY',
    sort: 12,
    status: 'on_sale',
  },
  {
    id: 10,
    sku_code: 'OAI-PLUS-M',
    card_title: 'ChatGPT Plus 月卡',
    card_subtitle: '主力模型组合，含 Codex 与 Images 档位白名单。',
    card_features: ['{limit_tokens} tokens / 月', 'gpt-4o + o1-mini', 'Codex 可用', '热销性价比'],
    share_seats: 5,
    products_category_name: 'openai',
    sku_product_name: 'GPT PRO 5X',
    limit_tokens: 8_000_000,
    rpm_limit: 90,
    allowed_models: ['gpt-4o', 'o1-mini'],
    billing_period: 'month',
    period_days: 30,
    price: 178,
    currency: 'CNY',
    hot_tag_name: '热销',
    sort: 8,
    status: 'on_sale',
  },
  {
    id: 6,
    sku_code: 'GEM-PRO-M',
    card_title: 'Gemini AI Pro 月卡',
    card_subtitle: 'Pro + Flash 组合，多模态与长上下文。',
    card_features: ['{limit_tokens} tokens / 月', '1.5 Pro / Flash', 'Nano 出图', 'Gemini 全系列'],
    share_seats: 5,
    products_category_name: 'gemini',
    sku_product_name: 'Gemini Pro Pool',
    limit_tokens: 10_000_000,
    rpm_limit: 90,
    allowed_models: ['gemini-1.5-pro', 'gemini-1.5-flash'],
    billing_period: 'month',
    period_days: 30,
    price: 189,
    currency: 'CNY',
    sort: 14,
    status: 'on_sale',
  },
  {
    id: 11,
    sku_code: 'CUR-PRO-M',
    card_title: 'Cursor 配置包 月卡',
    card_subtitle: '含使用额度 + Cursor 配置教程（搭配编辑器自助使用）。',
    card_features: ['额度 + 教程', 'Cloud Agents 说明', 'Auto 模式说明', '按文档自助配置'],
    share_seats: 5,
    products_category_name: 'cursor',
    sku_product_name: 'Mixed Pool',
    limit_tokens: 6_000_000,
    rpm_limit: 80,
    allowed_models: ['gpt-4o'],
    billing_period: 'month',
    period_days: 30,
    price: 198,
    currency: 'CNY',
    hot_tag_name: '热销',
    sort: 13,
    status: 'on_sale',
  },
  {
    id: 3,
    sku_code: 'ANT-PRO-M',
    card_title: 'Claude Pro 月卡',
    card_subtitle: 'Sonnet / Haiku 全系，支持 Claude Code。',
    card_features: ['{limit_tokens} tokens / 月', '200K 上下文', 'Claude 全系列', 'Claude Code 可用'],
    share_seats: 5,
    products_category_name: 'anthropic',
    sku_product_name: 'Claude Pro 5X',
    limit_tokens: 8_000_000,
    rpm_limit: 80,
    allowed_models: ['claude-3-5-sonnet-20241022'],
    billing_period: 'month',
    period_days: 30,
    price: 219,
    currency: 'CNY',
    hot_tag_name: '热销',
    sort: 7,
    status: 'on_sale',
  },
  {
    id: 2,
    sku_code: 'OAI-PRO20-M',
    card_title: 'ChatGPT Pro 20X 月卡',
    card_subtitle: '约 20 倍配额，团队满血与重度 Codex。',
    card_features: ['{limit_tokens} tokens / 月', '{rpm_limit} RPM', 'o1-mini 白名单', '专业 / 团队首选'],
    share_seats: 20,
    products_category_name: 'openai',
    sku_product_name: 'GPT PRO 20X',
    limit_tokens: 20_000_000,
    rpm_limit: 120,
    allowed_models: ['gpt-4o', 'o1-mini'],
    billing_period: 'month',
    period_days: 30,
    price: 1799,
    currency: 'CNY',
    hot_tag_name: '旗舰',
    sort: 5,
    status: 'on_sale',
  },
  {
    id: 12,
    sku_code: 'OAI-PRO5-M',
    card_title: 'ChatGPT Pro 5X 月卡',
    card_subtitle: 'Pro 5X，主力旗舰组合，约 5 倍配额。',
    card_features: ['{limit_tokens} tokens / 月', '旗舰模型可用', 'Codex 重度可用', '约 1 分钟开通'],
    share_seats: 10,
    products_category_name: 'openai',
    sku_product_name: 'GPT PRO 5X',
    limit_tokens: 12_000_000,
    rpm_limit: 100,
    allowed_models: ['gpt-4o', 'o1-mini'],
    billing_period: 'month',
    period_days: 30,
    price: 899,
    currency: 'CNY',
    sort: 6,
    status: 'on_sale',
  },
  {
    id: 4,
    sku_code: 'ANT-MAX-M',
    card_title: 'Claude Max 5X 月卡',
    card_subtitle: '约 5 倍 Pro 用量，重度编程友好。',
    card_features: ['{limit_tokens} tokens / 月', 'Opus 可选', 'Claude Code', 'Max 专线'],
    share_seats: 10,
    products_category_name: 'anthropic',
    sku_product_name: 'Claude Max 10X',
    limit_tokens: 15_000_000,
    rpm_limit: 100,
    allowed_models: ['claude-3-opus-20240229'],
    billing_period: 'month',
    period_days: 30,
    price: 999,
    currency: 'CNY',
    sort: 16,
    status: 'on_sale',
  },
  {
    id: 5,
    sku_code: 'XAI-GROK-M',
    card_title: 'Grok Super 3 个月',
    card_subtitle: 'DeepSearch + 联网 + Imagine 线路。',
    card_features: ['{limit_tokens} tokens / 周期', 'grok-2 系列', 'Grok 全能力', '3 个月包'],
    share_seats: 5,
    products_category_name: 'xai',
    sku_product_name: 'Grok 5X',
    limit_tokens: 6_000_000,
    rpm_limit: 60,
    allowed_models: ['grok-2'],
    billing_period: 'month',
    period_days: 30,
    price: 790,
    currency: 'CNY',
    sort: 18,
    status: 'on_sale',
  },
]

export const mockProducts: CatalogProduct[] = mockProductsRaw.map((p) => ({
  ...p,
  card_features: appendShareFeature(p.card_features, p.share_seats),
}))

/** 首页 Hero 固定文案（不随商品接口变化，避免加载后闪动） */
export const SUB_KEY_HERO_LINE = '支持开通子 Key，3–20 人按档共用'
