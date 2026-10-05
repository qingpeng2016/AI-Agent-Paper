import type { CatalogProduct } from '@/mocks/home'
import { subKeyShareFeature } from '@/mocks/home'

function appendShareBullet(bullets: string[], seats: number): string[] {
  const line = subKeyShareFeature(seats)
  if (bullets.some((f) => f.includes('人共用'))) return bullets
  return [...bullets, line]
}

export type ProductDetailContent = {
  eyebrow: string
  heroTitle: string
  heroLead: string
  heroBullets: string[]
  heroTags: string[]
  audiences: { title: string; desc: string }[]
  compareTitle?: string
  compareBody?: string
  steps: { title: string; desc: string }[]
  ctaTitle: string
  ctaSubtitle: string
  faqs: { q: string; a: string }[]
  related: { label: string; slug: string }[]
}

const BRAND_EYEBROW: Record<string, string> = {
  openai: 'ChatGPT',
  anthropic: 'Claude',
  xai: 'Grok',
  gemini: 'Gemini',
  perplexity: 'Perplexity',
  cursor: 'Cursor',
}

const OVERRIDES: Partial<Record<string, Partial<ProductDetailContent>>> = {
  'cursor-pro-plus': {
    eyebrow: 'Cursor · 重度版',
    heroTitle: 'Cursor Pro+ 会员套餐',
    heroLead:
      '面向重度编程与大型重构，额度约为 Pro 的 3 倍。支持前沿模型与 Max Mode，Cloud Agents、MCP 等能力按套餐说明配置。支付宝 / 微信自助下单，约 1 分钟到账。',
    heroBullets: [
      '约 3 倍 Pro 的 Agent 额度，减少触顶',
      '前沿模型 + Max Mode 重度可用',
      'Cloud Agents 并行 + MCP / skills / hooks',
      '支付宝 / 微信自助支付，约 1 分钟开通',
    ],
    audiences: [
      { title: '重度编程', desc: '长时间使用前沿模型 / Max Mode，减少撞额度。' },
      { title: '大型重构', desc: '整库理解、多文件改写更从容。' },
      { title: '常触顶的 Pro 用户', desc: '一周多次触顶，升级后立刻缓解。' },
      { title: '高强度个人', desc: '强度高但尚未到全天跑 Agent 的量级。' },
    ],
    compareTitle: 'Pro+ 还是 Ultra？',
    compareBody:
      '若一周多次触顶、重度使用前沿模型 / Max Mode，Pro+（约 3× 额度）通常最划算。几乎全天跑 Cloud Agents、多仓库并行或团队场景，再考虑 Ultra（约 20×）。日常轻度使用可继续 Pro 更省。',
    related: [
      { label: 'Cursor Pro 套餐', slug: 'cursor-pro' },
      { label: '查看全部 Cursor 套餐', slug: 'cursor-pro' },
    ],
  },
  'chatgpt-plus': {
    eyebrow: 'ChatGPT · 主力档',
    heroTitle: 'ChatGPT Plus 会员套餐',
    heroLead:
      '主力模型组合，适合日常对话、写作与轻量开发。支付宝 / 微信自助下单，无需海外信用卡，约 1 分钟开通后在会员中心查看额度。',
    compareTitle: 'Plus 还是 Pro 档位？',
    compareBody:
      '日常个人使用 Plus 性价比最高；需要更高 RPM、更多 tokens 或 Codex 重度场景，可选 Pro 5X / 20X 档位。',
    related: [
      { label: 'GPT Go 入门', slug: 'gpt-go' },
      { label: 'ChatGPT Pro 5X', slug: 'chatgpt-pro-5x' },
    ],
  },
}

export function getProductDetail(slug: string, product: CatalogProduct): ProductDetailContent {
  const brand = BRAND_EYEBROW[product.products_category_name] ?? 'AI 套餐'
  const base: ProductDetailContent = {
    eyebrow: `${brand} · ${product.sku_product_name}`,
    heroTitle: product.card_title.replace(/月卡$/, '会员套餐'),
    heroLead: `${product.card_subtitle} 支付宝 / 微信自助下单，约 1 分钟开通；额度与订单可在会员中心查看。`,
    heroBullets: product.card_features.map((f) => f),
    heroTags: ['支付宝 / 微信', '无需海外卡', '自助约 1 分钟'],
    audiences: [
      { title: '个人用户', desc: '希望低于官网价、国内支付自助开通。' },
      { title: '日常办公', desc: '写作、翻译、资料整理等高频使用。' },
      { title: '开发者', desc: '需要稳定额度与明确套餐边界。' },
      { title: '小团队', desc: '可先单账号试用，批量走企业采购。' },
    ],
    steps: [
      { title: '选择套餐', desc: '确认本页套餐与价格，点击立即购买。' },
      { title: '登录并支付', desc: '使用支付宝或微信完成付款。' },
      { title: '开始使用', desc: '约 1 分钟到账，在会员中心查看额度与订单。' },
    ],
    ctaTitle: `开通 ${product.sku_product_name}`,
    ctaSubtitle: '选好套餐并完成支付后，在会员中心查看额度与使用说明。',
    faqs: [
      {
        q: '国内能开通吗？需要海外卡吗？',
        a: '可以。支持支付宝 / 微信，无需海外信用卡。',
      },
      {
        q: '多久到账？',
        a: '通常约 1 分钟自动开通，可在会员中心查看状态。',
      },
      {
        q: '失败怎么办？',
        a: '支付异常或未到账请联系在线客服，核实后按规则处理。',
      },
      {
        q: '和官网套餐有什么关系？',
        a: '本站为独立第三方 AI 服务平台，提供会员套餐与开通服务，与商标持有人无隶属关系。',
      },
    ],
    related: [{ label: '查看全部套餐', slug: 'chatgpt-plus' }],
  }

  const patch = OVERRIDES[slug]
  const merged = patch
    ? { ...base, ...patch, audiences: patch.audiences ?? base.audiences }
    : base
  return {
    ...merged,
    heroBullets: appendShareBullet(
      merged.heroBullets ?? base.heroBullets,
      product.share_seats,
    ),
  }
}
