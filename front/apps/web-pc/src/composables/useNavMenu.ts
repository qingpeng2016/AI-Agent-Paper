import { ref, type Ref } from 'vue'
import {
  readSessionCache,
  writeSessionCache,
  type NavBrandMenu,
  type NavMegaColumn,
  type NavMenuResponse,
} from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { navBrandDropdowns, navMegaMenu } from '@/mocks/nav'
import type { ReloadOptions } from '@/composables/reloadOptions'

const megaMenu = ref<NavMegaColumn[]>([])
const brandMenus = ref<NavBrandMenu[]>([])
const loading = ref(false)

const NAV_CACHE_KEY = 'atm:nav-menu'

let loadSeq = 0

export function hydrateNavFromSession(): void {
  const cached = readSessionCache<NavMenuResponse>(NAV_CACHE_KEY)
  if (!cached) return
  applyNavData(mapNavApiResponse(cached))
}

export function hasNavSessionCache(): boolean {
  const cached = readSessionCache<NavMenuResponse>(NAV_CACHE_KEY)
  return !!(cached?.brand_menus?.length || cached?.mega_menu?.length)
}

export function hasNavInMemory(): boolean {
  return brandMenus.value.length > 0 || megaMenu.value.length > 0
}

function persistNavToSession(): void {
  writeSessionCache(NAV_CACHE_KEY, {
    mega_menu: megaMenu.value,
    brand_menus: brandMenus.value,
  })
}

function mockBrandMenus(): NavBrandMenu[] {
  return navBrandDropdowns
    .filter((e): e is Extract<typeof e, { items: unknown }> => 'items' in e)
    .slice(0, 3)
    .map((e) => ({
      id: e.id,
      label: e.label,
      hot_tag_name: e.hotTagName,
      items: e.items.map((i) => ({
        label: i.label,
        price: i.price ?? '',
        href: i.href,
        featured: i.featured,
      })),
    }))
}

function applyMockNav() {
  megaMenu.value = navMegaMenu
  brandMenus.value = mockBrandMenus()
}

function mapNavApiResponse(data: NavMenuResponse | null | undefined): NavMenuResponse {
  if (!data) {
    return { mega_menu: [], brand_menus: [] }
  }
  return {
    mega_menu: data.mega_menu ?? [],
    brand_menus: (data.brand_menus ?? []).map((b) => ({
      id: b.id,
      label: b.label,
      hot_tag_name: b.hot_tag_name,
      items: b.items ?? [],
    })),
  }
}

function applyNavData(data: NavMenuResponse) {
  if (data.brand_menus.length === 0 && data.mega_menu.length === 0) {
    if (!hasNavInMemory()) {
      applyMockNav()
    }
    return
  }
  if (data.mega_menu.length > 0) {
    megaMenu.value = data.mega_menu
  }
  if (data.brand_menus.length > 0) {
    brandMenus.value = data.brand_menus
  }
}

export async function reloadNavMenu(options?: ReloadOptions): Promise<void> {
  const seq = ++loadSeq
  const soft = options?.soft === true
  const hasNav = brandMenus.value.length > 0 || megaMenu.value.length > 0
  /** 切页 / 刷新带缓存：保留当前菜单文案，后台更新，不进入 navLoading（避免顶栏白骨架） */
  const keepVisible = soft && (hasNav || hasNavSessionCache())

  if (keepVisible) {
    try {
      const raw = await productApi.navMenu()
      if (seq !== loadSeq) return
      applyNavData(mapNavApiResponse(raw))
      persistNavToSession()
    } catch {
      if (seq !== loadSeq) return
    }
    return
  }

  loading.value = !hasNav
  if (!soft) {
    megaMenu.value = []
    brandMenus.value = []
    loading.value = true
  }
  try {
    const raw = await productApi.navMenu()
    if (seq !== loadSeq) return
    applyNavData(mapNavApiResponse(raw))
    persistNavToSession()
  } catch {
    if (seq !== loadSeq) return
    applyMockNav()
    persistNavToSession()
  } finally {
    if (seq === loadSeq) {
      loading.value = false
    }
  }
}

export function useNavMenu(): {
  megaMenu: Ref<NavMegaColumn[]>
  brandMenus: Ref<NavBrandMenu[]>
  loading: Ref<boolean>
  reload: () => Promise<void>
} {
  return { megaMenu, brandMenus, loading, reload: reloadNavMenu }
}
