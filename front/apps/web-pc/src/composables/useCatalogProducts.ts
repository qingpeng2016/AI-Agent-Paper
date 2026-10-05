import { computed, ref, type ComputedRef, type Ref } from 'vue'
import {
  catalogFilterPillsFromCategories,
  flattenCatalogProducts,
  normalizeCatalogProduct,
  readSessionCache,
  writeSessionCache,
  type CatalogFilterPill,
  type CatalogProduct,
  type ProductCatalogCategory,
} from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { catalogFilterPills, mockCategorySlugById, mockProducts } from '@/mocks/home'
import type { ReloadOptions } from '@/composables/reloadOptions'

const categories = ref<ProductCatalogCategory[]>([])
const loading = ref(false)

const CATALOG_CACHE_KEY = 'atm:catalog'

let loadSeq = 0

export function hydrateCatalogFromSession(): void {
  const cached = readSessionCache<{ categories: ProductCatalogCategory[] }>(CATALOG_CACHE_KEY)
  if (cached?.categories?.length) {
    categories.value = cached.categories
  }
}

export function hasCatalogSessionCache(): boolean {
  return !!readSessionCache<{ categories: ProductCatalogCategory[] }>(CATALOG_CACHE_KEY)
    ?.categories?.length
}

function persistCatalogToSession(): void {
  if (categories.value.length) {
    writeSessionCache(CATALOG_CACHE_KEY, { categories: categories.value })
  }
}

function mockCatalogCategories(): ProductCatalogCategory[] {
  const pills = catalogFilterPills.filter((p) => p.value !== 'all')
  return pills.map((pill) => {
    const id = Number(pill.value)
    const slug = mockCategorySlugById[id]
    const products = mockProducts
      .filter((p) => p.products_category_name === slug)
      .map((p) =>
        normalizeCatalogProduct({ ...p, products_category_id: id }),
      )
    return {
      id,
      name: pill.label,
      dot_color: pill.dot,
      active_bg: pill.activeBg,
      sort: id * 10,
      products,
    }
  })
}

export async function reloadCatalogProducts(
  options?: ReloadOptions,
): Promise<void> {
  const seq = ++loadSeq
  const soft = options?.soft === true
  loading.value = !soft || categories.value.length === 0
  if (!soft) {
    categories.value = []
  }
  try {
    const data = await productApi.list()
    if (seq !== loadSeq) return
    categories.value = (data.categories ?? []).map((c) => ({
      ...c,
      products: (c.products ?? []).map((p) => normalizeCatalogProduct(p)),
    }))
    persistCatalogToSession()
  } catch {
    if (seq !== loadSeq) return
    categories.value = mockCatalogCategories()
    persistCatalogToSession()
  } finally {
    if (seq === loadSeq) loading.value = false
  }
}

export function useCatalogProducts(): {
  categories: Ref<ProductCatalogCategory[]>
  catalogFilterPills: ComputedRef<CatalogFilterPill[]>
  products: ComputedRef<CatalogProduct[]>
  loading: Ref<boolean>
  reload: () => Promise<void>
} {
  const catalogFilterPillsComputed = computed(() => {
    if (categories.value.length) {
      return catalogFilterPillsFromCategories(categories.value)
    }
    return []
  })

  const products = computed(() =>
    flattenCatalogProducts({ categories: categories.value }),
  )

  return {
    categories,
    catalogFilterPills: catalogFilterPillsComputed,
    products,
    loading,
    reload: reloadCatalogProducts,
  }
}
