<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import ProductCard from '@/components/home/ProductCard.vue'
import ProductCardSkeleton from '@/components/home/ProductCardSkeleton.vue'
import type { CatalogProduct } from '@ai-token-mall/shared'
import { useCatalogProducts } from '@/composables/useCatalogProducts'
import { productDetailPath } from '@/mocks/productRoutes'

type FilterKey = 'all' | string

const props = withDefaults(
  defineProps<{
    open: boolean
    /** 升档等场景：只展示该分类商品，并隐藏顶部分类 pill */
    productsCategoryId?: number | null
    /** 升档：仅展示 limit_tokens 严格大于该值的商品（同分类内比档位） */
    upgradeBaselineLimitTokens?: number | null
  }>(),
  { productsCategoryId: null, upgradeBaselineLimitTokens: null },
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  buy: [product: CatalogProduct]
}>()

const router = useRouter()
const {
  products: catalogProducts,
  catalogFilterPills,
  loading: catalogLoading,
} = useCatalogProducts()

const catalogSkeletonCount = 6
const filter = ref<FilterKey>('all')
const selectedProductId = ref<number | null>(null)

const activePillStyle = computed(() => {
  const pill = catalogFilterPills.value.find((p) => p.value === filter.value)
  if (!pill?.activeBg || filter.value === 'all') {
    return { background: 'var(--atm-gradient)', color: '#fff' }
  }
  return { background: pill.activeBg, color: '#fff' }
})

const categoryLocked = computed(
  () => props.productsCategoryId != null && props.productsCategoryId > 0,
)

const upgradeMode = computed(
  () =>
    categoryLocked.value &&
    props.upgradeBaselineLimitTokens != null &&
    props.upgradeBaselineLimitTokens >= 0,
)

const modalTitle = computed(() =>
  categoryLocked.value ? '选择升级档位' : '全部套餐',
)

const emptyHint = computed(() => {
  if (upgradeMode.value) {
    return '没有更高套餐，请联系管理员手动处理。'
  }
  return '该品牌暂无套餐'
})

const filteredProducts = computed(() => {
  let list = [...catalogProducts.value]
  if (categoryLocked.value) {
    list = list.filter((p) => p.products_category_id === props.productsCategoryId)
  } else if (filter.value !== 'all') {
    list = list.filter((p) => p.products_category_id === Number(filter.value))
  }
  if (upgradeMode.value) {
    const baseline = props.upgradeBaselineLimitTokens ?? 0
    list = list.filter((p) => p.limit_tokens > baseline)
    list.sort((a, b) => a.limit_tokens - b.limit_tokens || a.sort - b.sort)
  } else {
    list.sort((a, b) => a.sort - b.sort)
  }
  return list
})

watch(
  filteredProducts,
  (list) => {
    if (catalogLoading.value) return
    selectedProductId.value = list[0]?.id ?? null
  },
  { immediate: true },
)

watch(
  () => props.open,
  (visible) => {
    if (visible && !categoryLocked.value) {
      filter.value = 'all'
    }
  },
)

function close() {
  emit('update:open', false)
}

function onOpenProduct(p: CatalogProduct) {
  selectedProductId.value = p.id
  emit('update:open', false)
  router.push(productDetailPath(p.sku_code))
}

function onBuy(p: CatalogProduct) {
  selectedProductId.value = p.id
  emit('buy', p)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open) close()
}
</script>

<template>
  <Teleport to="body">
    <Transition name="catalog-modal-fade">
      <div
        v-if="open"
        class="catalog-modal-backdrop"
        aria-hidden="false"
        @click.self="close"
        @keydown="onKeydown"
      >
        <div
          class="catalog-modal-panel"
          role="dialog"
          aria-modal="true"
          aria-labelledby="catalog-picker-title"
        >
          <header class="catalog-modal-head">
            <h2 id="catalog-picker-title">{{ modalTitle }}</h2>
            <button type="button" class="catalog-modal-close" aria-label="关闭" @click="close">
              ×
            </button>
          </header>

          <div v-if="!categoryLocked" class="catalog-pills">
            <button
              v-for="pill in catalogFilterPills"
              :key="pill.label"
              type="button"
              class="catalog-pill"
              :class="{ active: filter === pill.value }"
              :style="filter === pill.value ? activePillStyle : undefined"
              @click="filter = pill.value"
            >
              <span
                v-if="pill.dot && filter !== pill.value"
                class="pill-dot"
                :style="{ background: pill.dot }"
              />
              {{ pill.label }}
            </button>
          </div>

          <div class="catalog-modal-body">
            <div class="catalog-grid" :aria-busy="catalogLoading">
              <template v-if="catalogLoading">
                <ProductCardSkeleton v-for="i in catalogSkeletonCount" :key="`sk-${i}`" />
              </template>
              <template v-else>
                <ProductCard
                  v-for="p in filteredProducts"
                  :key="p.id"
                  :product="p"
                  :selected="selectedProductId === p.id"
                  @open="onOpenProduct"
                  @buy="onBuy"
                />
              </template>
            </div>
            <p v-if="!catalogLoading && !filteredProducts.length" class="catalog-empty">
              {{ emptyHint }}
            </p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.catalog-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 16px;
  overflow-y: auto;
  background: rgba(15, 10, 40, 0.52);
  backdrop-filter: blur(4px);
}

.catalog-modal-panel {
  width: 100%;
  max-width: 920px;
  padding: 18px 20px 22px;
  background: var(--atm-bg, #f5f3ff);
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(30, 27, 75, 0.22);
}

.catalog-modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.catalog-modal-head h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 800;
  color: var(--atm-text);
}

.catalog-modal-close {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  font-size: 26px;
  line-height: 1;
  color: var(--atm-text-muted);
  background: #fff;
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 12px;
  cursor: pointer;
  transition:
    color 0.15s ease,
    border-color 0.15s ease;
}

.catalog-modal-close:hover {
  color: var(--atm-primary-dark);
  border-color: rgba(124, 58, 237, 0.28);
}

.catalog-pills {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  margin-bottom: 16px;
  padding: 8px 10px;
  background: #fff;
  border-radius: 999px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.catalog-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  font-size: 13px;
  font-weight: 500;
  color: var(--atm-text);
  background: transparent;
  border: none;
  border-radius: 999px;
  cursor: pointer;
}

.catalog-pill.active {
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.pill-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.catalog-modal-body {
  max-height: min(62vh, 580px);
  overflow-y: auto;
  padding-right: 4px;
}

.catalog-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}

@media (min-width: 640px) {
  .catalog-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 860px) {
  .catalog-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

.catalog-empty {
  text-align: center;
  color: var(--atm-text-muted);
  padding: 48px 16px;
}

.catalog-modal-fade-enter-active,
.catalog-modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.catalog-modal-fade-enter-active .catalog-modal-panel,
.catalog-modal-fade-leave-active .catalog-modal-panel {
  transition: transform 0.2s ease;
}

.catalog-modal-fade-enter-from,
.catalog-modal-fade-leave-to {
  opacity: 0;
}

.catalog-modal-fade-enter-from .catalog-modal-panel,
.catalog-modal-fade-leave-to .catalog-modal-panel {
  transform: translateY(12px);
}
</style>
