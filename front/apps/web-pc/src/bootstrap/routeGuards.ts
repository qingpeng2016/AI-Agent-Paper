import type { RouteLocationNormalized, Router } from 'vue-router'
import { reloadCatalogProducts } from '@/composables/useCatalogProducts'
import { reloadEnterpriseProducts } from '@/composables/useEnterpriseProducts'
import {
  hasNavInMemory,
  hydrateNavFromSession,
  reloadNavMenu,
} from '@/composables/useNavMenu'
import { prefetchProductDetail } from '@/composables/useProductDetail'
import {
  prefetchBlogArticle,
  reloadTutorialBlog,
} from '@/composables/useTutorialBlog'
import { setRouteNavigationLoading } from '@/composables/useRouteNavigationLoading'
import {
  hydrateRouteSessionCache,
  isColdNavigation,
  routeHasSessionCache,
} from '@/bootstrap/sessionRouteCache'

async function prepareRouteData(
  to: RouteLocationNormalized,
  from: RouteLocationNormalized,
  opts?: { soft?: boolean },
): Promise<void> {
  const name = to.name
  const soft = opts?.soft ?? !isColdNavigation(from)

  if (name === 'login' || name === 'register') {
    return
  }

  hydrateNavFromSession()
  const cold = isColdNavigation(from)
  const needNavFetch = cold || !hasNavInMemory()
  if (needNavFetch) {
    await reloadNavMenu({ soft: soft || hasNavInMemory() })
  }

  if (name === 'home' || name === 'member') {
    await reloadCatalogProducts({ soft })
  }
  if (name === 'product') {
    const slug = String(to.params.slug ?? '')
    if (slug) {
      await prefetchProductDetail(slug, { soft })
    }
  }
  if (name === 'enterprise') {
    await reloadEnterpriseProducts({ soft })
  }
  if (name === 'blog') {
    await reloadTutorialBlog({ soft })
  }
  if (name === 'blog-article') {
    const slug = String(to.params.slug ?? '')
    await reloadTutorialBlog({ soft })
    if (slug) {
      await prefetchBlogArticle(slug)
    }
  }
}

/**
 * 切页：原页面不动 + 顶栏 loading，接口返回后再换内容。
 * 刷新：session 还原上一屏后立即展示，再 soft 拉接口（体验与切页一致）。
 */
export function installRouteDataGuards(router: Router): void {
  router.beforeEach(async (to, from) => {
    if (from.matched.length > 0 && to.fullPath === from.fullPath) {
      return true
    }

    const cold = isColdNavigation(from)

    if (cold) {
      hydrateRouteSessionCache(to)
    }

    if (cold && routeHasSessionCache(to)) {
      setRouteNavigationLoading(true)
      void prepareRouteData(to, from, { soft: true }).finally(() => {
        setRouteNavigationLoading(false)
      })
      return true
    }

    setRouteNavigationLoading(true)
    try {
      await prepareRouteData(to, from, { soft: !cold })
      return true
    } finally {
      setRouteNavigationLoading(false)
    }
  })
}
