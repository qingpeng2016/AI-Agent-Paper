import type { CatalogProduct } from './product'

export type ProductDetailAudience = {
  title: string
  desc: string
}

export type ProductDetailStep = {
  title: string
  desc: string
}

export type ProductDetailFaq = {
  q: string
  a: string
}

export type ProductDetailRelated = {
  label: string
  slug: string
}

export type ProductDetailContent = {
  eyebrow: string
  hero_title: string
  hero_lead: string
  hero_bullets: string[]
  hero_tags: string[]
  audiences: ProductDetailAudience[]
  compare_title?: string
  compare_body?: string
  steps: ProductDetailStep[]
  cta_title: string
  cta_subtitle: string
  faqs: ProductDetailFaq[]
  related: ProductDetailRelated[]
}

export type ProductDetailResponse = {
  slug: string
  product: CatalogProduct
  detail: ProductDetailContent
}
