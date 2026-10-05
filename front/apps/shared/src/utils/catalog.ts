import type {
  CatalogProduct,
  ProductCatalogCategory,
  ProductCatalogResponse,
} from '../types/product'
import { formatTokenCount } from './format'

/** card_features 内占位符，与后端 expandCardFeatures 一致 */
export function expandCardFeaturePlaceholders(
  features: string[],
  ctx: Pick<CatalogProduct, 'limit_tokens' | 'rpm_limit'>,
): string[] {
  return features.map((line) => {
    if (!line.includes('{')) return line
    return line
      .replaceAll('{limit_tokens}', formatTokenCount(ctx.limit_tokens))
      .replaceAll('{rpm_limit}', String(ctx.rpm_limit))
  })
}

export function subKeyShareFeature(seats: number): string {
  return `支持${seats}人共用`
}

export function appendShareFeature(features: string[], seats: number): string[] {
  const line = subKeyShareFeature(seats)
  if (features.some((f) => f.includes('人共用'))) return features
  return [...features, line]
}

export function normalizeCatalogProduct(p: CatalogProduct): CatalogProduct {
  const card_features = expandCardFeaturePlaceholders(p.card_features, p)
  const legacy = p as CatalogProduct & { sort_order?: number }
  const sort = typeof p.sort === 'number' ? p.sort : legacy.sort_order ?? 0
  return {
    ...p,
    sort,
    card_features: appendShareFeature(card_features, p.share_seats),
  }
}

export function flattenCatalogProducts(
  catalog: ProductCatalogResponse,
): CatalogProduct[] {
  return catalog.categories.flatMap((c) =>
    c.products.map((p) =>
      normalizeCatalogProduct({
        ...p,
        products_category_id: p.products_category_id ?? c.id,
      }),
    ),
  )
}

export type CatalogFilterPill = {
  value: 'all' | string
  label: string
  dot?: string
  activeBg?: string
}

export function catalogFilterPillsFromCategories(
  categories: ProductCatalogCategory[],
): CatalogFilterPill[] {
  const sorted = [...categories].sort((a, b) => a.sort - b.sort)
  return [
    { value: 'all', label: '全部' },
    ...sorted.map((c) => ({
      value: String(c.id),
      label: c.name,
      dot: c.dot_color,
      activeBg: c.active_bg,
    })),
  ]
}

export function heroSubKeyShareLine(products: { share_seats: number }[]): string {
  if (products.length === 0) return ''
  const seats = products.map((p) => p.share_seats)
  const min = Math.min(...seats)
  const max = Math.max(...seats)
  if (min === max) return `支持开通子 Key，${subKeyShareFeature(min)}`
  return `支持开通子 Key，${min}–${max} 人按档共用`
}
