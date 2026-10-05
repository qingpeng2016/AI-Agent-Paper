<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { formatCny } from '@ai-token-mall/shared'
import PurchaseModal from '@/components/checkout/PurchaseModal.vue'
import {
  getSessionUser,
  isLoggedIn,
  PENDING_BUY_KEY,
} from '@/composables/useSessionUser'
import {
  prefetchProductDetail,
  takePrefetchedProductDetail,
  useProductDetail,
} from '@/composables/useProductDetail'
import type { CatalogProduct } from '@ai-token-mall/shared'
import ProductDetailSkeleton from '@/components/home/ProductDetailSkeleton.vue'

const route = useRoute()
const router = useRouter()

const purchaseOpen = ref(false)
const sessionUser = ref(getSessionUser())
const { product, detail, loading } = useProductDetail()

const slug = computed(() => String(route.params.slug ?? ''))

async function loadForSlug(s: string) {
  if (!s) return
  if (takePrefetchedProductDetail(s)) return
  await prefetchProductDetail(s)
  if (!product.value) {
    router.replace('/')
  }
}

watch(
  slug,
  (s) => {
    void loadForSlug(s)
  },
  { immediate: true },
)

function onBuy(p: CatalogProduct) {
  sessionUser.value = getSessionUser()
  if (!isLoggedIn()) {
    sessionStorage.setItem(PENDING_BUY_KEY, String(p.id))
    ElMessage.warning('请先登录后再购买')
    router.push({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  purchaseOpen.value = true
}

const openFaqs = ref<string[]>([])
</script>

<template>
  <ProductDetailSkeleton v-if="loading && !product" />
  <div
    v-else-if="product && detail"
    class="product-page"
    :class="{ 'product-page--busy': loading && !!product }"
  >
    <div class="atm-container">
      <section class="hero-card">
        <div class="hero-main">
          <span class="eyebrow">{{ detail.eyebrow }}</span>
          <h1>{{ detail.heroTitle }}</h1>
          <p class="lead">{{ detail.heroLead }}</p>
          <ul class="hero-bullets">
            <li v-for="(b, i) in detail.heroBullets" :key="i">
              <span class="check" aria-hidden="true">✓</span>
              <span class="feature-text">{{ b }}</span>
            </li>
          </ul>
          <button type="button" class="atm-btn-primary hero-scroll" @click="onBuy(product)">
            立即购买
          </button>
          <div class="hero-tags">
            <span v-for="(t, i) in detail.heroTags" :key="i">{{ t }}</span>
          </div>
        </div>

        <aside id="buy-card" class="buy-card">
          <h2 class="buy-card-title">{{ product.card_title }}</h2>
          <p class="buy-card-price">{{ formatCny(product.price) }}</p>
          <p class="buy-card-desc">{{ product.card_subtitle }}</p>
          <ul class="buy-card-features">
            <li v-for="(f, i) in product.card_features" :key="i">
              <span class="check" aria-hidden="true">✓</span>
              <span class="feature-text">{{ f }}</span>
            </li>
          </ul>
          <button type="button" class="buy-card-cta" @click="onBuy(product)">立即购买</button>
        </aside>
      </section>

      <section class="section">
        <h2 class="section-title">{{ product.sku_product_name }} 适合谁</h2>
        <div class="audience-grid">
          <article v-for="(a, i) in detail.audiences" :key="i" class="audience-card">
            <h3>{{ a.title }}</h3>
            <p>{{ a.desc }}</p>
          </article>
        </div>
      </section>

      <section v-if="detail.compareTitle && detail.compareBody" class="section">
        <h2 class="section-title">{{ detail.compareTitle }}</h2>
        <p class="compare-body">{{ detail.compareBody }}</p>
      </section>

      <section class="section">
        <h2 class="section-title">怎么开通</h2>
        <ol class="steps">
          <li v-for="(s, i) in detail.steps" :key="i">
            <span class="step-num">{{ i + 1 }}</span>
            <div>
              <strong>{{ s.title }}</strong>
              <p>{{ s.desc }}</p>
            </div>
          </li>
        </ol>
      </section>

      <section class="section">
        <h2 class="section-title">常见问题</h2>
        <el-collapse v-model="openFaqs" class="faq-collapse">
          <el-collapse-item
            v-for="(f, i) in detail.faqs"
            :key="i"
            :title="f.q"
            :name="String(i)"
          >
            <p class="faq-a">{{ f.a }}</p>
          </el-collapse-item>
        </el-collapse>
      </section>

      <nav v-if="detail.related.length" class="related">
        <RouterLink
          v-for="(r, i) in detail.related"
          :key="i"
          :to="`/p/${r.slug}`"
        >
          {{ r.label }} →
        </RouterLink>
      </nav>
    </div>

    <PurchaseModal
      v-model:open="purchaseOpen"
      :product="product"
      :user="sessionUser"
    />
  </div>
</template>

<style scoped>
.product-page {
  padding: 28px 0 56px;
  background: var(--atm-bg);
}

.product-page--busy {
  opacity: 0.96;
  transition: opacity 0.2s ease;
}

.hero-card {
  display: grid;
  gap: 28px;
  padding: 32px 28px;
  margin-bottom: 40px;
  background: #fff;
  border: 1px solid #ede9fe;
  border-radius: 24px;
  box-shadow: var(--atm-shadow-lg);
}

@media (min-width: 960px) {
  .hero-card {
    grid-template-columns: 1fr minmax(280px, 340px);
    gap: 36px;
    padding: 40px 36px;
  }
}

.eyebrow {
  display: inline-block;
  margin-bottom: 12px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-primary);
  background: #f5f3ff;
  border-radius: 999px;
}

.hero-main h1 {
  margin: 0 0 16px;
  font-size: clamp(26px, 4vw, 34px);
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.2;
  color: var(--atm-text);
}

.lead {
  margin: 0 0 20px;
  font-size: 15px;
  line-height: 1.65;
  color: var(--atm-text-muted);
}

.hero-bullets {
  margin: 0 0 24px;
  padding: 0;
  list-style: none;
}

.hero-bullets li {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 12px;
  line-height: 1.45;
}

.hero-bullets .feature-text {
  font-size: 15px;
  font-weight: 600;
  color: var(--atm-text);
}

.check {
  flex-shrink: 0;
  margin-top: 2px;
  font-size: 15px;
  color: #22c55e;
  font-weight: 800;
}

.hero-scroll {
  margin-bottom: 16px;
}

.hero-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.hero-tags span {
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 500;
  color: var(--atm-primary);
  background: #f5f3ff;
  border-radius: 999px;
}

.buy-card {
  padding: 24px 22px;
  background: #fff;
  border: 2px solid #c4b5fd;
  border-radius: 20px;
}

.buy-card-title {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 700;
  color: var(--atm-text);
}

.buy-card-price {
  margin: 0 0 12px;
  font-size: 36px;
  font-weight: 800;
  color: var(--atm-primary);
  letter-spacing: -0.02em;
}

.buy-card-desc {
  margin: 0 0 14px;
  font-size: 13px;
  line-height: 1.55;
  color: var(--atm-text-muted);
}

.buy-card-features {
  margin: 0 0 20px;
  padding: 12px 14px;
  list-style: none;
  background: #f8f7ff;
  border-radius: 12px;
  border: 1px solid #ede9fe;
}

.buy-card-features li {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 10px;
}

.buy-card-features li:last-child {
  margin-bottom: 0;
}

.buy-card-features .feature-text {
  font-size: 15px;
  font-weight: 600;
  line-height: 1.4;
  color: var(--atm-text);
}

.buy-card-cta {
  width: 100%;
  padding: 14px;
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  background: var(--atm-gradient);
  border: none;
  border-radius: 14px;
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.35);
}

.section {
  margin-bottom: 40px;
}

.section-title {
  margin: 0 0 20px;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.audience-grid {
  display: grid;
  gap: 16px;
}

@media (min-width: 768px) {
  .audience-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.audience-card {
  padding: 20px 22px;
  background: #fff;
  border: 1px solid #ede9fe;
  border-radius: 16px;
}

.audience-card h3 {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 700;
  color: var(--atm-text);
}

.audience-card p {
  margin: 0;
  font-size: 14px;
  line-height: 1.55;
  color: var(--atm-text-muted);
}

.compare-body {
  margin: 0;
  padding: 20px 22px;
  font-size: 15px;
  line-height: 1.65;
  color: var(--atm-text-muted);
  background: #fff;
  border: 1px solid #ede9fe;
  border-radius: 16px;
}

.steps {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.steps li {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  padding: 18px 20px;
  background: #fff;
  border: 1px solid #ede9fe;
  border-radius: 14px;
}

.step-num {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  background: var(--atm-gradient);
  border-radius: 50%;
}

.steps strong {
  display: block;
  margin-bottom: 4px;
  font-size: 15px;
  color: var(--atm-text);
}

.steps p {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: var(--atm-text-muted);
}

.faq-collapse {
  border: none;
  background: transparent;
}

.faq-collapse :deep(.el-collapse-item) {
  margin-bottom: 10px;
  overflow: hidden;
  background: #fff;
  border: 1px solid #ede9fe;
  border-radius: 12px;
}

.faq-collapse :deep(.el-collapse-item__header) {
  padding: 0 18px;
  font-weight: 600;
  color: var(--atm-text);
  border: none;
}

.faq-collapse :deep(.el-collapse-item__wrap) {
  border: none;
}

.faq-a {
  margin: 0;
  padding: 0 18px 16px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--atm-text-muted);
}

.related {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  padding-top: 8px;
}

.related a {
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-primary);
  text-decoration: none;
}

.related a:hover {
  text-decoration: underline;
}
</style>
