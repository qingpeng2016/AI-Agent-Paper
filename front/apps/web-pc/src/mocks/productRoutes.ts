import type { CatalogProduct } from '@ai-token-mall/shared'

/** URL slug → sku_code */
export const SLUG_TO_SKU: Record<string, string> = {
  'gpt-go': 'OAI-GO-M',
  'chatgpt-plus': 'OAI-PLUS-M',
  'chatgpt-pro-5x': 'OAI-PRO5-M',
  'chatgpt-pro-20x': 'OAI-PRO20-M',
  'claude-pro': 'ANT-PRO-M',
  'claude-max-5x': 'ANT-MAX-M',
  'grok-super': 'XAI-GROK-M',
  'gemini-pro': 'GEM-PRO-M',
  'cursor-pro': 'CUR-PRO-M',
  'cursor-pro-plus': 'CUR-PRO-M',
  'perplexity-pro': 'PPX-PRO-M',
}

export const SKU_TO_SLUG: Record<string, string> = {
  'OAI-GO-M': 'gpt-go',
  'OAI-PLUS-M': 'chatgpt-plus',
  'OAI-PRO5-M': 'chatgpt-pro-5x',
  'OAI-PRO20-M': 'chatgpt-pro-20x',
  'ANT-PRO-M': 'claude-pro',
  'ANT-MAX-M': 'claude-max-5x',
  'XAI-GROK-M': 'grok-super',
  'GEM-PRO-M': 'gemini-pro',
  'CUR-PRO-M': 'cursor-pro',
  'PPX-PRO-M': 'perplexity-pro',
}

/** 同一 SKU 多个 slug 时，取 canonical（先注册的） */
export function productDetailPath(skuCode: string): string {
  const slug = SKU_TO_SLUG[skuCode] ?? skuCode.toLowerCase().replace(/_/g, '-')
  return `/p/${slug}`
}

export function findProductBySlug(
  slug: string,
  products: CatalogProduct[],
): CatalogProduct | undefined {
  const sku = SLUG_TO_SKU[slug]
  if (!sku) return undefined
  return products.find((p) => p.sku_code === sku)
}

export function canonicalSlugForProduct(product: CatalogProduct): string {
  const slug = SKU_TO_SLUG[product.sku_code]
  return slug ?? product.sku_code.toLowerCase()
}
