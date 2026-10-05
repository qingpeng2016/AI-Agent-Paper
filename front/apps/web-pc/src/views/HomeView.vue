<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import PurchaseModal from '@/components/checkout/PurchaseModal.vue'
import {
  getSessionUser,
  isLoggedIn,
  PENDING_BUY_KEY,
} from '@/composables/useSessionUser'
import CompareRowIcon from '@/components/home/CompareRowIcon.vue'
import ProductCard from '@/components/home/ProductCard.vue'
import ProductCardSkeleton from '@/components/home/ProductCardSkeleton.vue'
import type { CatalogProduct } from '@ai-token-mall/shared'
import { useCatalogProducts } from '@/composables/useCatalogProducts'
import {
  compareRows,
  compareSection,
  faqs,
  reviewSummary,
  reviews,
  trustStats,
  heroChecklist,
  SUB_KEY_HERO_LINE,
} from '@/mocks/home'
import { productDetailPath } from '@/mocks/productRoutes'

type FilterKey = 'all' | string

const router = useRouter()
const route = useRoute()
const {
  products: catalogProducts,
  catalogFilterPills,
  loading: catalogLoading,
} = useCatalogProducts()

const catalogSkeletonCount = 6
const filter = ref<FilterKey>('all')
const selectedProductId = ref<number | null>(null)
const purchaseOpen = ref(false)
const purchaseProduct = ref<CatalogProduct | null>(null)
const sessionUser = ref(getSessionUser())
const activePillStyle = computed(() => {
  const pill = catalogFilterPills.value.find((p) => p.value === filter.value)
  if (!pill?.activeBg || filter.value === 'all') {
    return { background: 'var(--atm-gradient)', color: '#fff' }
  }
  return { background: pill.activeBg, color: '#fff' }
})

const filteredProducts = computed(() => {
  let list = [...catalogProducts.value]
  if (filter.value !== 'all') {
    list = list.filter((p) => p.products_category_id === Number(filter.value))
  }
  list.sort((a, b) => a.sort - b.sort)
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

function onOpenProduct(p: CatalogProduct) {
  selectedProductId.value = p.id
  router.push(productDetailPath(p.sku_code))
}

function openPurchase(p: CatalogProduct) {
  purchaseProduct.value = p
  purchaseOpen.value = true
}

function onBuy(p: CatalogProduct) {
  selectedProductId.value = p.id
  sessionUser.value = getSessionUser()
  if (!isLoggedIn()) {
    sessionStorage.setItem(PENDING_BUY_KEY, String(p.id))
    ElMessage.warning('请先登录后再购买')
    router.push({ path: '/login', query: { redirect: '/' } })
    return
  }
  openPurchase(p)
}

function resumePendingPurchase() {
  if (route.path !== '/') return
  const raw = sessionStorage.getItem(PENDING_BUY_KEY)
  if (!raw || !isLoggedIn()) return
  sessionStorage.removeItem(PENDING_BUY_KEY)
  const id = Number(raw)
  const p = catalogProducts.value.find((item) => item.id === id)
  if (p) {
    selectedProductId.value = p.id
    openPurchase(p)
  }
}

function syncCatalogFilterFromRoute() {
  const cat = route.query.cat
  if (cat === 'all') {
    filter.value = 'all'
    return
  }
  if (typeof cat === 'string' && cat !== '') {
    const exists = catalogFilterPills.value.some((p) => p.value === cat)
    if (exists) filter.value = cat
  }
}

function scrollToCatalogSection() {
  nextTick(() => {
    document.getElementById('catalog')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  })
}

onMounted(() => {
  const hasCatalogIntent = route.hash === '#catalog' || route.query.cat !== undefined
  if (hasCatalogIntent) {
    syncCatalogFilterFromRoute()
    scrollToCatalogSection()
  } else if (window.location.hash === '#catalog') {
    history.replaceState(null, '', `${window.location.pathname}${window.location.search}`)
    window.scrollTo(0, 0)
  }
  sessionUser.value = getSessionUser()
  resumePendingPurchase()
})

watch(
  () => [route.query.cat, route.hash] as const,
  () => {
    if (route.path !== '/') return
    if (route.hash !== '#catalog' && route.query.cat === undefined) return
    syncCatalogFilterFromRoute()
    scrollToCatalogSection()
  },
)

watch(catalogFilterPills, () => {
  if (route.path === '/' && route.query.cat !== undefined) {
    syncCatalogFilterFromRoute()
  }
})

watch(catalogProducts, () => {
  resumePendingPurchase()
})

watch(
  () => route.fullPath,
  () => {
    sessionUser.value = getSessionUser()
    resumePendingPurchase()
  },
)

watch(purchaseOpen, (open) => {
  if (open) sessionUser.value = getSessionUser()
})

function reviewInitial(user: string) {
  return user.charAt(0)
}
</script>

<template>
  <div class="home">
    <section class="hero-band">
      <div class="atm-container">
        <div class="hero-inner">
          <div class="hero-center">
            <h1>
              国内低价开通
              <span class="grad">ChatGPT / Claude / Cursor</span>
              等 AI 服务
            </h1>
            <p class="hero-highlight">{{ SUB_KEY_HERO_LINE }}</p>
            <p class="sub">
              <strong>支付宝 / 微信</strong>自助下单；团队批量与开票见
              <RouterLink to="/enterprise">企业采购</RouterLink>。
            </p>
            <div class="hero-actions">
              <a href="#catalog" class="atm-btn-primary">查看全部套餐 →</a>
              <RouterLink to="/enterprise" class="atm-btn-ghost">企业采购 / 开票</RouterLink>
            </div>
            <ul class="hero-checks">
              <li v-for="(item, i) in heroChecklist" :key="i">{{ item }}</li>
            </ul>
          </div>
        </div>
      </div>
      <!-- 与下方套餐目录同宽：全宽 atm-container -->
      <div class="atm-container trust-bar-wrap">
        <div class="trust-bar">
          <div v-for="(t, i) in trustStats" :key="i" class="trust-cell">
            <strong>{{ t.value }}</strong>
            <span>{{ t.label }}</span>
          </div>
        </div>
      </div>
    </section>

    <section id="catalog" class="catalog-section atm-container">
      <header class="catalog-head">
        <h2>全部套餐</h2>
      </header>
      <template v-if="catalogLoading && !catalogProducts.length">
        <div class="catalog-pills catalog-pills--sk" aria-hidden="true">
          <span v-for="i in 6" :key="i" class="catalog-pill-sk" />
        </div>
        <div class="catalog-grid" aria-busy="true">
          <ProductCardSkeleton v-for="i in catalogSkeletonCount" :key="`sk-${i}`" />
        </div>
      </template>
      <template v-else>
        <div class="catalog-pills">
          <button
            v-for="pill in catalogFilterPills"
            :key="pill.label"
            type="button"
            class="catalog-pill"
            :class="{ active: filter === pill.value }"
            :style="filter === pill.value ? activePillStyle : undefined"
            @click="filter = pill.value"
          >
            <span v-if="pill.dot && filter !== pill.value" class="pill-dot" :style="{ background: pill.dot }" />
            {{ pill.label }}
          </button>
        </div>
        <div class="catalog-grid">
          <ProductCard
            v-for="p in filteredProducts"
            :key="p.id"
            :product="p"
            :selected="selectedProductId === p.id"
            @open="onOpenProduct"
            @buy="onBuy"
          />
        </div>
      </template>
      <p v-if="!catalogLoading && !filteredProducts.length" class="catalog-empty">
        该品牌暂无套餐
      </p>
    </section>

    <section class="compare-section atm-container">
      <header class="compare-head">
        <h2 class="compare-title">{{ compareSection.title }}</h2>
      </header>

      <div class="compare-cards">
        <div class="compare-card compare-card--mall">
          <span class="compare-badge">推荐</span>
          <div class="compare-card-inner">
            <h3 class="compare-card-name">{{ compareSection.mallName }}</h3>
            <ul class="compare-list">
              <li v-for="row in compareRows" :key="'mall-' + row.label">
                <span class="compare-row-icon compare-row-icon--mall" aria-hidden="true">
                  <CompareRowIcon :name="row.icon" />
                </span>
                <div class="compare-row-body">
                  <span class="compare-row-label">{{ row.label }}</span>
                  <span class="compare-row-value compare-row-value--ok">
                    <span class="mark ok" aria-hidden="true">✓</span>
                    {{ row.us }}
                  </span>
                </div>
              </li>
            </ul>
            <a href="#catalog" class="compare-cta compare-cta--primary">
              {{ compareSection.mallCta }}
            </a>
          </div>
        </div>

        <div class="compare-card compare-card--diy">
          <div class="compare-card-inner">
            <h3 class="compare-card-name compare-card-name--muted">{{ compareSection.diyName }}</h3>
            <ul class="compare-list">
              <li v-for="row in compareRows" :key="'diy-' + row.label">
                <span class="compare-row-icon compare-row-icon--diy" aria-hidden="true">
                  <CompareRowIcon :name="row.icon" />
                </span>
                <div class="compare-row-body">
                  <span class="compare-row-label">{{ row.label }}</span>
                  <span class="compare-row-value compare-row-value--no">
                    <span class="mark no" aria-hidden="true">✕</span>
                    {{ row.them }}
                  </span>
                </div>
              </li>
            </ul>
            <a href="#catalog" class="compare-cta compare-cta--ghost">
              {{ compareSection.diyCta }}
            </a>
          </div>
        </div>
      </div>
    </section>

    <section class="reviews-section">
      <div class="atm-container">
        <header class="reviews-head">
          <div class="reviews-head-text">
            <h2 class="reviews-title">用户怎么说</h2>
          </div>
          <div class="reviews-score" aria-label="综合评分">
            <span class="reviews-score-num">{{ reviewSummary.score }}</span>
            <div class="reviews-score-meta">
              <span class="reviews-stars" aria-hidden="true">★★★★★</span>
              <p>综合评分 · 基于 {{ reviewSummary.count }} 条用户评价</p>
            </div>
          </div>
        </header>

        <div class="reviews-grid">
          <article v-for="(r, i) in reviews" :key="i" class="review-card">
            <span class="reviews-stars review-card-stars" aria-hidden="true">★★★★★</span>
            <p class="review-card-text">{{ r.text }}</p>
            <footer class="review-card-foot">
              <span class="review-avatar" aria-hidden="true">{{ reviewInitial(r.user) }}</span>
              <div>
                <strong>{{ r.user }}</strong>
                <span>{{ r.sku }} · {{ r.date }}</span>
              </div>
            </footer>
          </article>
        </div>
      </div>
    </section>

    <!-- FAQ 双列静态，非折叠列表 -->
    <section id="faq" class="faq-section atm-container">
      <h2 class="section-head">FAQ</h2>
      <div class="faq-grid">
        <div v-for="(f, i) in faqs" :key="i" class="faq-block">
          <h4>{{ f.q }}</h4>
          <p>{{ f.a }}</p>
        </div>
      </div>
    </section>

    <PurchaseModal
      v-model:open="purchaseOpen"
      :product="purchaseProduct"
      :user="sessionUser"
    />
  </div>
</template>

<style scoped>
.home {
  background: var(--atm-bg);
}

.hero-band {
  padding: 40px 0 32px;
}
.hero-inner {
  max-width: 815px;
  margin: 0 auto;
}
.hero-center {
  text-align: center;
}
.trust-bar-wrap {
  margin-top: 28px;
}
.trust-bar {
  width: 100%;
}
.hero-center h1 {
  margin: 0 0 16px;
  font-size: clamp(22px, 3.6vw, 32px);
  font-weight: 800;
  line-height: 1.3;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}
.grad {
  background: var(--atm-gradient);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.hero-highlight {
  margin: 0 auto 14px;
  max-width: 640px;
  font-size: clamp(18px, 2.4vw, 20px);
  font-weight: 600;
  line-height: 1.65;
  background: var(--atm-gradient);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.sub {
  max-width: 593px;
  margin: 0 auto 24px;
  font-size: 16px;
  line-height: 1.7;
  color: var(--atm-text-muted);
}
.sub a {
  color: var(--atm-primary);
  font-weight: 500;
  text-decoration: none;
}
.sub a:hover {
  text-decoration: underline;
}
.hero-actions {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 16px;
}
.hero-checks {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px 24px;
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: 13px;
  color: var(--atm-text-muted);
}
.hero-checks li::before {
  content: '✓ ';
  color: var(--atm-primary);
  font-weight: 700;
}

.trust-bar-wrap .trust-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 8px 32px rgba(30, 27, 75, 0.08);
  overflow: hidden;
}
@media (max-width: 639px) {
  .trust-bar-wrap .trust-bar {
    grid-template-columns: repeat(2, 1fr);
  }
}
.trust-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 22px 16px;
  text-align: center;
  border-right: 1px solid #f1f5f9;
}
.trust-cell:last-child {
  border-right: none;
}
@media (max-width: 639px) {
  .trust-cell {
    border-right: none;
    border-bottom: 1px solid #f1f5f9;
  }
  .trust-cell:nth-child(odd) {
    border-right: 1px solid #f1f5f9;
  }
  .trust-cell:nth-last-child(-n + 2) {
    border-bottom: none;
  }
}
.trust-cell strong {
  font-size: 22px;
  font-weight: 800;
  color: var(--atm-primary);
  margin-bottom: 6px;
}
.trust-cell span {
  font-size: 13px;
  color: #64748b;
}

.catalog-section {
  /* 勿用 padding 简写，否则会冲掉 .atm-container 的左右 24px */
  padding-top: 48px;
  padding-bottom: 56px;
}
.catalog-head h2 {
  margin: 0 0 24px;
  font-size: 26px;
  font-weight: 800;
  color: var(--atm-text);
}
.catalog-pills {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  margin-bottom: 28px;
  padding: 10px 12px;
  background: #fff;
  border-radius: 999px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}
.catalog-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  font-size: 14px;
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
.catalog-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}
@media (min-width: 640px) {
  .catalog-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (min-width: 960px) {
  .catalog-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
.catalog-pills--sk {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 20px;
}

.catalog-pill-sk {
  display: inline-block;
  width: 72px;
  height: 36px;
  border-radius: 999px;
  background: linear-gradient(90deg, #eef1f6 0%, #e2e8f0 45%, #eef1f6 90%);
  background-size: 200% 100%;
  animation: catalog-pill-shimmer 1.2s ease-in-out infinite;
}

@keyframes catalog-pill-shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: -100% 0;
  }
}

.catalog-empty {
  text-align: center;
  color: var(--atm-text-muted);
  padding: 48px;
}

.section-head {
  margin: 0 0 24px;
  font-size: 22px;
  font-weight: 700;
}
.compare-section {
  padding-bottom: 56px;
}
.compare-head {
  text-align: left;
  margin-bottom: 32px;
}
.compare-title {
  margin: 0;
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.compare-cards {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
  align-items: stretch;
  padding-top: 12px;
}
@media (min-width: 900px) {
  .compare-cards {
    grid-template-columns: 1fr 1fr;
    gap: 24px;
  }
}
.compare-card {
  position: relative;
  min-width: 0;
}
.compare-card--mall {
  padding: 2px;
  border-radius: 20px;
  background: linear-gradient(135deg, #7c3aed 0%, #6366f1 50%, #818cf8 100%);
  box-shadow: 0 12px 40px rgba(124, 58, 237, 0.22);
}
.compare-card--diy {
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  background: #fff;
}
.compare-badge {
  position: absolute;
  top: 0;
  left: 50%;
  z-index: 1;
  transform: translate(-50%, -50%);
  padding: 4px 14px;
  font-size: 12px;
  font-weight: 700;
  color: #fff;
  background: var(--atm-gradient);
  border-radius: 999px;
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.35);
}
.compare-card-inner {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 28px 22px 24px;
  background: #fff;
  border-radius: 18px;
}
.compare-card--diy .compare-card-inner {
  border-radius: 19px;
}
.compare-card-name {
  margin: 8px 0 20px;
  font-size: 18px;
  font-weight: 800;
  text-align: center;
}
.compare-card-name--muted {
  color: #64748b;
}
.compare-list {
  flex: 1;
  margin: 0 0 24px;
  padding: 0;
  list-style: none;
}
.compare-list li {
  display: flex;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid #f1f5f9;
}
.compare-list li:last-child {
  border-bottom: none;
}
.compare-row-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
}
.compare-row-icon :deep(svg) {
  width: 22px;
  height: 22px;
}
.compare-row-icon--mall {
  color: #6366f1;
  background: #eef2ff;
}
.compare-row-icon--diy {
  color: #94a3b8;
  background: #f8fafc;
}
.compare-row-body {
  min-width: 0;
}
.compare-row-label {
  display: block;
  margin-bottom: 4px;
  font-size: 12px;
  color: var(--atm-text-muted);
}
.compare-row-value {
  display: block;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
}
.compare-row-value--ok {
  color: var(--atm-text);
}
.compare-row-value--no {
  color: #64748b;
  font-weight: 500;
}
.mark {
  margin-right: 4px;
  font-weight: 800;
}
.mark.ok {
  color: #16a34a;
}
.mark.no {
  color: #ef4444;
}
.compare-cta {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 14px 20px;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
  border-radius: 999px;
  transition: transform 0.15s, box-shadow 0.15s;
}
.compare-cta--primary {
  color: #fff !important;
  background: var(--atm-gradient);
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.35);
}
.compare-cta--primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 22px rgba(124, 58, 237, 0.45);
}
.compare-cta--ghost {
  color: var(--atm-text);
  background: #fff;
  border: 1px solid #e2e8f0;
}
.compare-cta--ghost:hover {
  border-color: #cbd5e1;
  background: #fafafa;
}

.reviews-section {
  padding: 56px 0 80px;
}
.reviews-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px 32px;
  margin-bottom: 28px;
}
.reviews-title {
  margin: 0 0 6px;
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.reviews-score {
  display: flex;
  align-items: center;
  gap: 12px;
}
.reviews-score-num {
  font-size: 40px;
  font-weight: 800;
  line-height: 1;
  color: var(--atm-primary);
  letter-spacing: -0.03em;
}
.reviews-score-meta p {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--atm-text-muted);
}
.reviews-stars {
  font-size: 14px;
  letter-spacing: 2px;
  color: #f59e0b;
}
.reviews-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}
@media (min-width: 640px) {
  .reviews-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (min-width: 960px) {
  .reviews-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
.review-card {
  display: flex;
  flex-direction: column;
  padding: 22px 22px 18px;
  background: #fff;
  border: 1px solid #ede9fe;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(91, 33, 182, 0.06);
}
.review-card-stars {
  margin-bottom: 12px;
}
.review-card-text {
  flex: 1;
  margin: 0 0 18px;
  font-size: 14px;
  line-height: 1.65;
  color: var(--atm-text);
}
.review-card-foot {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: var(--atm-text-muted);
}
.review-card-foot strong {
  display: block;
  font-size: 13px;
  color: var(--atm-text);
}
.review-avatar {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  background: var(--atm-gradient);
  border-radius: 50%;
}
.faq-section {
  padding-top: 8px;
  padding-bottom: 48px;
}
.faq-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}
@media (min-width: 768px) {
  .faq-grid {
    grid-template-columns: 1fr 1fr;
  }
}
.faq-block h4 {
  margin: 0 0 8px;
  font-size: 15px;
}
.faq-block p {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--atm-text-muted);
}

</style>
