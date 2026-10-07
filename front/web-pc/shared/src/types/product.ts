/** 与表 products 字段对齐（展示/下单用） */

export type ProductsCategoryName =
  | 'openai'
  | 'anthropic'
  | 'xai'
  | 'gemini'
  | 'perplexity'
  | 'cursor'

/** @deprecated 使用 products_category_name */
export type UpstreamName = Exclude<ProductsCategoryName, 'cursor'>

export type BillingPeriod = 'month' | 'year' | 'once'
export type ProductStatus = 'on_sale' | 'off_sale'

export interface Product {
  id: number
  sku_code: string
  products_category_id?: number | null
  products_category_name: ProductsCategoryName | string
  sku_product_name: string
  limit_tokens: number
  rpm_limit: number
  tpm_limit?: number | null
  allowed_models: string[]
  billing_period: BillingPeriod
  /** 每个计费周期天数（订阅到期/续费/升档计算依据） */
  period_days?: number
  /** 售价（元，两位小数） */
  price: number | string
  currency: string
  /** 卡片右上角标签，空则不显示 */
  hot_tag_name?: string
  /** 分类内排序，对应表 products.sort_order */
  sort: number
  status: ProductStatus
}

/** 首页 / 选购弹窗卡片展示 */
export interface CatalogProduct extends Product {
  card_title: string
  card_subtitle: string
  card_features: string[]
  share_seats: number
}

export interface ProductCatalogCategory {
  id: number
  name: string
  dot_color?: string
  active_bg?: string
  sort: number
  /** 顶栏分类标签，空则不显示 */
  hot_tag_name?: string
  products: CatalogProduct[]
}

export interface ProductCatalogResponse {
  categories: ProductCatalogCategory[]
}

export interface HomeBanner {
  id: string
  title: string
  subtitle: string
  ctaText: string
  tone: 'primary' | 'success' | 'warning'
}
