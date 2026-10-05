import { ref, type Ref } from 'vue'

const routeNavigationLoading = ref(false)
let loadingDepth = 0

export function useRouteNavigationLoading(): { active: Ref<boolean> } {
  return { active: routeNavigationLoading }
}

export function setRouteNavigationLoading(active: boolean): void {
  if (active) {
    loadingDepth++
    routeNavigationLoading.value = true
    return
  }
  loadingDepth = 0
  routeNavigationLoading.value = false
}

/** 会员中心等页内请求：与路由守卫共用顶栏进度条，支持并发计数 */
export function beginRouteNavigationLoading(): void {
  loadingDepth++
  routeNavigationLoading.value = true
}

export function endRouteNavigationLoading(): void {
  loadingDepth = Math.max(0, loadingDepth - 1)
  if (loadingDepth === 0) {
    routeNavigationLoading.value = false
  }
}
