import { ref, type Ref } from 'vue'
import {
  normalizeCatalogProduct,
  readSessionCache,
  writeSessionCache,
  type CatalogProduct,
  type ProductDetailResponse,
} from '@ai-token-mall/shared'
import { productApi } from '@/api'
import {
  getProductDetail,
  type ProductDetailContent as MockProductDetailContent,
} from '@/mocks/productDetails'
import { findProductBySlug } from '@/mocks/productRoutes'
import { mockProducts } from '@/mocks/home'
import type { ReloadOptions } from '@/composables/reloadOptions'

export type ProductDetailViewContent = MockProductDetailContent

const productDetailCacheKey = (slug: string) => `atm:product-detail:${slug}`

type ProductDetailCache = {
  product: CatalogProduct
  detail: ProductDetailViewContent
}

const product = ref<CatalogProduct | null>(null)
const detail = ref<ProductDetailViewContent | null>(null)
const loading = ref(false)

let loadSeq = 0
let prefetchedSlug: string | null = null

function mapApiDetail(data: ProductDetailResponse): ProductDetailCache {
  const p = normalizeCatalogProduct(data.product)
  const d = data.detail
  return {
    product: p,
    detail: {
      eyebrow: d.eyebrow,
      heroTitle: d.hero_title,
      heroLead: d.hero_lead,
      heroBullets: d.hero_bullets ?? [],
      heroTags: d.hero_tags ?? [],
      audiences: d.audiences ?? [],
      compareTitle: d.compare_title,
      compareBody: d.compare_body,
      steps: d.steps ?? [],
      ctaTitle: d.cta_title,
      ctaSubtitle: d.cta_subtitle,
      faqs: d.faqs ?? [],
      related: d.related ?? [],
    },
  }
}

function applyDetailCache(cached: ProductDetailCache, slug: string) {
  product.value = cached.product
  detail.value = cached.detail
  prefetchedSlug = slug
}

function persistProductDetailToSession(slug: string, cached: ProductDetailCache) {
  writeSessionCache(productDetailCacheKey(slug), cached)
}

export function hydrateProductDetailFromSession(slug: string): void {
  const cached = readSessionCache<ProductDetailCache>(productDetailCacheKey(slug))
  if (cached?.product && cached.detail) {
    applyDetailCache(cached, slug)
  }
}

export function hasProductDetailSessionCache(slug: string): boolean {
  return !!readSessionCache<ProductDetailCache>(productDetailCacheKey(slug))?.product
}

function applyMock(slug: string) {
  const fromList = findProductBySlug(slug, mockProducts)
  if (!fromList) {
    product.value = null
    detail.value = null
    return
  }
  product.value = fromList
  detail.value = getProductDetail(slug, fromList)
}

export async function prefetchProductDetail(
  slug: string,
  options?: ReloadOptions,
): Promise<void> {
  const s = slug.trim()
  if (!s) return
  const seq = ++loadSeq
  const soft = options?.soft === true
  const hasData = product.value && detail.value && prefetchedSlug === s
  loading.value = !soft || !hasData
  try {
    const data = await productApi.detailBySlug(s)
    if (seq !== loadSeq) return
    const mapped = mapApiDetail(data)
    applyDetailCache(mapped, s)
    persistProductDetailToSession(s, mapped)
  } catch {
    if (seq !== loadSeq) return
    if (!soft || !hasData) {
      applyMock(s)
      if (product.value && detail.value) {
        persistProductDetailToSession(s, {
          product: product.value,
          detail: detail.value,
        })
      }
    }
    prefetchedSlug = product.value ? s : null
  } finally {
    if (seq === loadSeq) loading.value = false
  }
}

export function takePrefetchedProductDetail(slug: string): boolean {
  if (prefetchedSlug === slug && product.value && detail.value) {
    prefetchedSlug = null
    return true
  }
  return false
}

export function useProductDetail(): {
  product: Ref<CatalogProduct | null>
  detail: Ref<ProductDetailViewContent | null>
  loading: Ref<boolean>
} {
  return { product, detail, loading }
}
