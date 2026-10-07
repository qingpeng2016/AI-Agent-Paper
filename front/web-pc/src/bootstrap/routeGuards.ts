import type { Router } from 'vue-router'
import { isLoggedIn } from '@/composables/useSessionUser'

/** 未登录访问需鉴权路由 → 登录页（带 redirect） */
export function installAuthRouteGuard(router: Router): void {
  router.beforeEach((to) => {
    const needsAuth = to.matched.some((record) => record.meta.requiresAuth === true)
    if (needsAuth && !isLoggedIn()) {
      return {
        path: '/login',
        query: { redirect: to.fullPath },
      }
    }
    if ((to.name === 'login' || to.name === 'register') && isLoggedIn()) {
      const redirect =
        typeof to.query.redirect === 'string' && to.query.redirect.startsWith('/')
          ? to.query.redirect
          : '/workbench'
      return { path: redirect }
    }
    return true
  })
}
