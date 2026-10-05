import type { Router } from 'vue-router'
import { initTracker, pageEnter, setCurrentPage, trackMetaForRoute } from '@ai-token-mall/shared'
import { getAuthToken } from '@/utils/auth-cookie'

export function installTracking(router: Router) {
  const baseURL = import.meta.env.VITE_API_BASE_URL ?? ''
  initTracker({
    baseURL,
    getToken: () => getAuthToken(),
    enabled: import.meta.env.VITE_TRACKING_ENABLED !== 'false',
    channel: 'web',
  })

  router.afterEach((to) => {
    const meta = trackMetaForRoute(to.name, to.fullPath)
    setCurrentPage(meta)
    pageEnter(meta)
  })
}
