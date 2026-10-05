import type { RouteLocationNormalized } from 'vue-router'
import { hydrateCatalogFromSession, hasCatalogSessionCache } from '@/composables/useCatalogProducts'
import { hydrateEnterpriseProductsFromSession, hasEnterpriseProductsSessionCache } from '@/composables/useEnterpriseProducts'
import { hydrateNavFromSession, hasNavSessionCache } from '@/composables/useNavMenu'
import {
  hydrateProductDetailFromSession,
  hasProductDetailSessionCache,
} from '@/composables/useProductDetail'
import {
  hydrateTutorialBlogFromSession,
  hydrateTutorialArticleFromSession,
  hasTutorialArticleSessionCache,
  hasTutorialBlogSessionCache,
} from '@/composables/useTutorialBlog'

/** 刷新前用 session 还原上一屏，再 soft 拉接口（与切页「原页不动 + 顶栏 loading」一致） */
export function hydrateRouteSessionCache(to: RouteLocationNormalized): void {
  hydrateNavFromSession()
  const name = to.name

  if (name === 'home') {
    hydrateCatalogFromSession()
  }
  if (name === 'enterprise') {
    hydrateEnterpriseProductsFromSession()
  }
  if (name === 'blog') {
    hydrateTutorialBlogFromSession()
  }
  if (name === 'blog-article') {
    hydrateTutorialBlogFromSession()
    const slug = String(to.params.slug ?? '')
    if (slug) hydrateTutorialArticleFromSession(slug)
  }
  if (name === 'product') {
    const slug = String(to.params.slug ?? '')
    if (slug) hydrateProductDetailFromSession(slug)
  }
}

export function routeHasSessionCache(to: RouteLocationNormalized): boolean {
  const name = to.name
  if (name === 'login' || name === 'register') {
    return true
  }
  if (!hasNavSessionCache()) {
    return false
  }
  if (name === 'home') {
    return hasCatalogSessionCache()
  }
  if (name === 'enterprise') {
    return hasEnterpriseProductsSessionCache()
  }
  if (name === 'blog') {
    return hasTutorialBlogSessionCache()
  }
  if (name === 'blog-article') {
    const slug = String(to.params.slug ?? '')
    return hasTutorialBlogSessionCache() && (!slug || hasTutorialArticleSessionCache(slug))
  }
  if (name === 'product') {
    const slug = String(to.params.slug ?? '')
    return slug ? hasProductDetailSessionCache(slug) : false
  }
  return true
}

export function isColdNavigation(from: RouteLocationNormalized): boolean {
  return from.matched.length === 0
}
