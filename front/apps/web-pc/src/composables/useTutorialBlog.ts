import { ref, type Ref } from 'vue'
import {
  readSessionCache,
  writeSessionCache,
  type TutorialArticle,
  type TutorialCategory,
} from '@ai-token-mall/shared'
import { tutorialApi } from '@/api'
import {
  blogCategories,
  blogPageMeta,
  blogPosts,
  getBlogPost,
  type BlogPost,
} from '@/mocks/blog'
import type { ReloadOptions } from '@/composables/reloadOptions'

export type TutorialFilterCategory = {
  id: string
  label: string
}

const pageTitle = ref('')
const categories = ref<TutorialFilterCategory[]>([])
const articles = ref<BlogPost[]>([])
const loading = ref(false)

const TUTORIAL_BLOG_CACHE_KEY = 'atm:tutorial-blog'
const tutorialArticleCacheKey = (slug: string) => `atm:tutorial-article:${slug}`

type TutorialBlogCache = {
  pageTitle: string
  categories: TutorialFilterCategory[]
  articles: BlogPost[]
}

let loadSeq = 0
let prefetchedSlug: string | null = null
let prefetchedPost: BlogPost | null = null

export function hydrateTutorialBlogFromSession(): void {
  const cached = readSessionCache<TutorialBlogCache>(TUTORIAL_BLOG_CACHE_KEY)
  if (!cached?.articles?.length) return
  pageTitle.value = cached.pageTitle
  categories.value = cached.categories
  articles.value = cached.articles
}

export function hasTutorialBlogSessionCache(): boolean {
  return !!readSessionCache<TutorialBlogCache>(TUTORIAL_BLOG_CACHE_KEY)?.articles?.length
}

function persistTutorialBlogToSession(): void {
  if (!articles.value.length) return
  writeSessionCache(TUTORIAL_BLOG_CACHE_KEY, {
    pageTitle: pageTitle.value,
    categories: categories.value,
    articles: articles.value,
  })
}

export function hydrateTutorialArticleFromSession(slug: string): void {
  const cached = readSessionCache<{ post: BlogPost }>(tutorialArticleCacheKey(slug))
  if (!cached?.post) return
  prefetchedPost = cached.post
  prefetchedSlug = slug
}

export function hasTutorialArticleSessionCache(slug: string): boolean {
  return !!readSessionCache<{ post: BlogPost }>(tutorialArticleCacheKey(slug))?.post
}

function fromMock() {
  pageTitle.value = blogPageMeta.title
  categories.value = blogCategories.map((c) => ({ id: c.id, label: c.label }))
  articles.value = blogPosts
}

function mapApiArticle(a: TutorialArticle): BlogPost {
  return {
    slug: a.slug,
    category: a.category_code as BlogPost['category'],
    categoryLabel: a.category_name,
    date: a.date,
    title: a.title,
    excerpt: a.excerpt,
    body: a.body ?? [],
  }
}

function applyApiList(data: {
  title?: string
  categories?: TutorialCategory[]
  articles?: TutorialArticle[]
}) {
  pageTitle.value = data.title || blogPageMeta.title
  categories.value = [
    { id: 'all', label: '全部' },
    ...(data.categories ?? [])
      .slice()
      .sort((a, b) => a.sort - b.sort)
      .map((c) => ({ id: c.code, label: c.name })),
  ]
  articles.value = (data.articles ?? []).map(mapApiArticle)
}

export async function reloadTutorialBlog(
  options?: ReloadOptions,
): Promise<void> {
  const seq = ++loadSeq
  const soft = options?.soft === true
  loading.value = !soft || articles.value.length === 0
  if (!soft) {
    pageTitle.value = ''
    categories.value = []
    articles.value = []
  }
  try {
    const data = await tutorialApi.list()
    if (seq !== loadSeq) return
    applyApiList(data)
    persistTutorialBlogToSession()
  } catch {
    if (seq !== loadSeq) return
    fromMock()
    persistTutorialBlogToSession()
  } finally {
    if (seq === loadSeq) loading.value = false
  }
}

export function findCachedBlogPost(slug: string): BlogPost | undefined {
  return articles.value.find((a) => a.slug === slug)
}

function persistTutorialArticleToSession(slug: string, post: BlogPost): void {
  writeSessionCache(tutorialArticleCacheKey(slug), { post })
}

export async function prefetchBlogArticle(slug: string): Promise<void> {
  try {
    const data = await tutorialApi.detail(slug)
    if (data?.slug) {
      prefetchedPost = mapApiArticle(data)
      prefetchedSlug = slug
      persistTutorialArticleToSession(slug, prefetchedPost)
      return
    }
  } catch {
    /* fallback in view */
  }
  const fromList = findCachedBlogPost(slug)
  if (fromList?.body?.length) {
    prefetchedPost = fromList
    prefetchedSlug = slug
    persistTutorialArticleToSession(slug, fromList)
    return
  }
  prefetchedPost = getBlogPost(slug) ?? null
  prefetchedSlug = prefetchedPost ? slug : null
  if (prefetchedPost) {
    persistTutorialArticleToSession(slug, prefetchedPost)
  }
}

/** 路由守卫预取成功后，详情页直接消费，避免再闪 skeleton */
export function takePrefetchedBlogPost(slug: string): BlogPost | null {
  if (prefetchedSlug === slug && prefetchedPost) {
    const post = prefetchedPost
    prefetchedSlug = null
    prefetchedPost = null
    return post
  }
  return null
}

export function useTutorialBlog(): {
  pageTitle: Ref<string>
  categories: Ref<TutorialFilterCategory[]>
  articles: Ref<BlogPost[]>
  loading: Ref<boolean>
  reload: () => Promise<void>
} {
  return {
    pageTitle,
    categories,
    articles,
    loading,
    reload: reloadTutorialBlog,
  }
}
