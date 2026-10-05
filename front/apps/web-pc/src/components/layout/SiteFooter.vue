<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import SiteLogo from '@/components/brand/SiteLogo.vue'
import { SITE_NAME } from '@/constants/brand'
import {
  reloadCatalogProducts,
  useCatalogProducts,
} from '@/composables/useCatalogProducts'
import {
  reloadTutorialBlog,
  useTutorialBlog,
} from '@/composables/useTutorialBlog'
import { blogCategories } from '@/mocks/blog'
import { catalogFilterPills as mockCatalogPills, mockProducts } from '@/mocks/home'
import { productDetailPath } from '@/mocks/productRoutes'

/** 与原先「工具 · 帮助」列链接数量接近 */
const FOOTER_TUTORIAL_CATEGORY_LIMIT = 5

const aboutLinks = [
  { label: '关于我们', href: '#' },
  { label: '联系我们', href: '#' },
  { label: '企业采购', href: '/enterprise' },
  { label: '分销合作', href: '#' },
  { label: '服务条款与隐私', href: '#' },
]

const { categories: tutorialCategories } = useTutorialBlog()
const { categories: productCategories } = useCatalogProducts()

onMounted(() => {
  if (!productCategories.value.length) {
    void reloadCatalogProducts({ soft: true })
  }
  if (!tutorialCategories.value.some((c) => c.id !== 'all')) {
    void reloadTutorialBlog({ soft: true })
  }
})

const chatgptLinks = computed(() => {
  const chatgptCategory =
    productCategories.value.find((c) => c.name === 'ChatGPT') ??
    productCategories.value.find((c) =>
      c.products.some((p) => p.products_category_name === 'openai'),
    )

  const products = chatgptCategory
    ? [...chatgptCategory.products].sort((a, b) => a.sort - b.sort)
    : mockProducts
        .filter((p) => p.products_category_name === 'openai')
        .sort((a, b) => a.sort - b.sort)

  return products.map((p) => ({
    label: p.card_title || p.sku_product_name,
    to: productDetailPath(p.sku_code),
  }))
})

const moreAiLinks = computed(() => {
  const sorted =
    productCategories.value.length > 0
      ? [...productCategories.value].sort((a, b) => a.sort - b.sort)
      : mockCatalogPills
          .filter((p) => p.value !== 'all')
          .map((p) => ({
            id: Number(p.value),
            name: p.label,
            sort: Number(p.value),
          }))

  return sorted.map((c) => ({
    label: c.name,
    to: {
      path: '/',
      hash: '#catalog',
      query: { cat: String(c.id) },
    },
  }))
})

const toolsLinks = computed(() => {
  const fromState = tutorialCategories.value.filter((c) => c.id !== 'all')
  const source =
    fromState.length > 0
      ? fromState
      : blogCategories.map((c) => ({ id: c.id, label: c.label }))

  return source.slice(0, FOOTER_TUTORIAL_CATEGORY_LIMIT).map((c) => ({
    label: c.label,
    to: { path: '/blog', query: { cat: c.id } },
  }))
})

const footerColumns = computed(() => [
  { title: 'CHATGPT', links: chatgptLinks.value, kind: 'router' as const },
  { title: '更多 AI', links: moreAiLinks.value, kind: 'router' as const },
  { title: '工具 · 帮助', links: toolsLinks.value, kind: 'router' as const },
  { title: '关于', links: aboutLinks, kind: 'href' as const },
])
</script>

<template>
  <footer class="site-footer">
    <div class="atm-container footer-grid">
      <div class="footer-brand">
        <SiteLogo variant="footer" class="footer-logo" />
      </div>

      <div v-for="(col, idx) in footerColumns" :key="idx" class="footer-col">
        <h3 class="footer-col-title">{{ col.title }}</h3>
        <ul class="footer-col-list">
          <li v-for="link in col.links" :key="link.label">
            <RouterLink
              v-if="col.kind === 'router' && 'to' in link"
              :to="link.to"
              class="footer-link"
            >
              {{ link.label }}
            </RouterLink>
            <a v-else-if="'href' in link" :href="link.href" class="footer-link">{{ link.label }}</a>
          </li>
        </ul>
      </div>
    </div>

    <div class="footer-legal">
      <div class="atm-container footer-legal-inner">
        <p>
          © 2026 {{ SITE_NAME }} · 独立第三方 AI 服务平台，与 OpenAI 等商标持有人无隶属或授权关系
        </p>
        <p class="footer-legal-muted">
          ChatGPT、OpenAI 为 OpenAI, Inc. 商标；名称仅作描述性（nominative）说明用途。
        </p>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.site-footer {
  background: #0c1222;
  color: #94a3b8;
  font-size: 13px;
  line-height: 1.5;
}

.footer-grid {
  display: grid;
  grid-template-columns: minmax(200px, 1.4fr) repeat(4, minmax(0, 1fr));
  gap: 32px 24px;
  padding: 48px 24px 40px;
}

@media (max-width: 1024px) {
  .footer-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
  .footer-brand {
    grid-column: 1 / -1;
    max-width: 420px;
  }
}

@media (max-width: 640px) {
  .footer-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 28px 20px;
    padding: 36px 16px 32px;
  }
}

.footer-logo {
  display: inline-flex;
  margin-bottom: 12px;
}

.footer-col-title {
  margin: 0 0 14px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: #64748b;
}

.footer-col-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.footer-col-list li + li {
  margin-top: 10px;
}

.footer-link {
  color: #cbd5e1;
  text-decoration: none;
  transition: color 0.15s;
}

.footer-link:hover {
  color: #fff;
}

.footer-legal {
  border-top: 1px solid #1e293b;
  background: #080d18;
}

.footer-legal-inner {
  padding: 20px 24px 28px;
  text-align: center;
}

.footer-legal p {
  margin: 0;
  font-size: 12px;
  color: #64748b;
  line-height: 1.6;
}

.footer-legal-muted {
  margin-top: 8px !important;
  font-size: 11px !important;
  color: #475569 !important;
}
</style>
