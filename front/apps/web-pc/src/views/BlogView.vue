<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import BlogListSkeleton from '@/components/blog/BlogListSkeleton.vue'
import { useTutorialBlog } from '@/composables/useTutorialBlog'

const route = useRoute()
const router = useRouter()
const { pageTitle, categories, articles, loading } = useTutorialBlog()

function parseCategory(q: unknown): string {
  const id = typeof q === 'string' ? q : 'all'
  return categories.value.some((c) => c.id === id) ? id : 'all'
}

const activeCategory = ref('all')

watch(
  categories,
  () => {
    activeCategory.value = parseCategory(route.query.cat)
  },
  { immediate: true },
)

watch(
  () => route.query.cat,
  (cat) => {
    activeCategory.value = parseCategory(cat)
  },
)

const filteredPosts = computed(() => {
  if (activeCategory.value === 'all') return articles.value
  return articles.value.filter((p) => p.category === activeCategory.value)
})

function setCategory(id: string) {
  activeCategory.value = id
  const query = id === 'all' ? {} : { cat: id }
  router.replace({ path: '/blog', query })
}
</script>

<template>
  <div class="blog-page">
    <div class="atm-container blog-container">
      <nav class="blog-breadcrumb" aria-label="面包屑">
        <RouterLink to="/">首页</RouterLink>
        <span class="sep">/</span>
        <span class="current">教程</span>
      </nav>

      <BlogListSkeleton v-if="loading && !articles.length" />

      <template v-else>
        <header class="blog-hero">
          <h1 class="blog-title">{{ pageTitle }}</h1>
        </header>

        <div class="blog-filters" role="tablist" aria-label="文章分类">
          <button
            v-for="cat in categories"
            :key="cat.id"
            type="button"
            class="filter-pill"
            :class="{ 'filter-pill--active': activeCategory === cat.id }"
            role="tab"
            :aria-selected="activeCategory === cat.id"
            @click="setCategory(cat.id)"
          >
            {{ cat.label }}
          </button>
        </div>

        <div v-if="filteredPosts.length" class="blog-grid">
          <RouterLink
            v-for="post in filteredPosts"
            :key="post.slug"
            :to="{ name: 'blog-article', params: { slug: post.slug } }"
            class="blog-card"
          >
            <div class="blog-card-meta">
              <span class="blog-card-tag">{{ post.categoryLabel }}</span>
              <time class="blog-card-date" :datetime="post.date">{{ post.date }}</time>
            </div>
            <h2 class="blog-card-title">{{ post.title }}</h2>
            <p class="blog-card-excerpt">{{ post.excerpt }}</p>
          </RouterLink>
        </div>

        <p v-else class="blog-empty">该分类暂无文章，试试「全部」。</p>
      </template>
    </div>
  </div>
</template>

<style scoped>
.blog-page {
  padding: 28px 0 72px;
  background: var(--atm-bg-soft);
  min-height: 60vh;
}

.blog-container {
  max-width: 1080px;
}

.blog-breadcrumb {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  font-size: 14px;
  color: var(--atm-text-muted);
}

.blog-breadcrumb a {
  color: var(--atm-text-muted);
  text-decoration: none;
}

.blog-breadcrumb a:hover {
  color: var(--atm-primary);
}

.blog-breadcrumb .current {
  color: var(--atm-text);
}

.blog-breadcrumb .sep {
  opacity: 0.45;
}

.blog-hero {
  margin-bottom: 28px;
}

.blog-title {
  margin: 0;
  font-size: clamp(26px, 4vw, 34px);
  font-weight: 800;
  color: var(--atm-text);
  letter-spacing: -0.02em;
}

.blog-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 28px;
}

.filter-pill {
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-text-muted);
  background: #fff;
  border: 1px solid #e8ecf4;
  border-radius: 999px;
  cursor: pointer;
  transition:
    color 0.15s ease,
    border-color 0.15s ease,
    background 0.15s ease;
}

.filter-pill:hover {
  border-color: #d5dbe8;
  color: var(--atm-text);
}

.filter-pill--active {
  color: #fff;
  background: var(--atm-gradient);
  border-color: transparent;
}

.blog-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

@media (min-width: 720px) {
  .blog-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.blog-card {
  display: block;
  padding: 22px;
  text-decoration: none;
  color: inherit;
  background: #fff;
  border: 1px solid #e8ecf4;
  border-radius: 16px;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.blog-card:hover {
  border-color: #d5dbe8;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.06);
  transform: translateY(-2px);
}

.blog-card-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 12px;
}

.blog-card-tag {
  padding: 3px 8px;
  font-weight: 700;
  color: var(--atm-primary);
  background: #f5f3ff;
  border-radius: 999px;
}

.blog-card-date {
  color: var(--atm-text-muted);
}

.blog-card-title {
  margin: 0 0 10px;
  font-size: 18px;
  font-weight: 800;
  line-height: 1.35;
  color: var(--atm-text);
}

.blog-card-excerpt {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: var(--atm-text-muted);
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.blog-empty {
  text-align: center;
  color: var(--atm-text-muted);
  padding: 48px 16px;
}
</style>
