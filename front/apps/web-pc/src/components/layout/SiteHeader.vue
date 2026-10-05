<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import SiteLogo from '@/components/brand/SiteLogo.vue'
import { authSessionRevision, isLoggedIn } from '@/composables/useSessionUser'
import NavMegaMenuSkeleton from '@/components/layout/NavMegaMenuSkeleton.vue'
import { reloadNavMenu, useNavMenu } from '@/composables/useNavMenu'
import { isNavDropdown, type NavMenuEntry } from '@/mocks/nav'
import SiteNotificationBell from '@/components/layout/SiteNotificationBell.vue'

const { megaMenu, brandMenus, loading: navLoading } = useNavMenu()

const navBrandEntries = computed((): NavMenuEntry[] => [
  ...brandMenus.value.map((b) => ({
    id: b.id,
    label: b.label,
    hotTagName: b.hot_tag_name,
    items: b.items.map((i) => ({
      label: i.label,
      href: i.href,
      price: i.price,
    })),
  })),
  {
    id: 'enterprise',
    label: '企业采购',
    plainLink: true,
    to: '/enterprise',
  },
])

const route = useRoute()
const loggedIn = ref(isLoggedIn())

function syncAuthState() {
  authSessionRevision.value
  loggedIn.value = isLoggedIn()
}

onMounted(() => {
  syncAuthState()
  if (!brandMenus.value.length) {
    void reloadNavMenu({ soft: true })
  }
})
watch(() => route.fullPath, syncAuthState)

const megaOpen = ref(false)
const openDrop = ref<string | null>(null)

function openMega() {
  megaOpen.value = true
  openDrop.value = null
}

function closeAll() {
  megaOpen.value = false
  openDrop.value = null
}

function openBrandDrop(id: string) {
  openDrop.value = id
  megaOpen.value = false
}

function isSpaNav(href: string) {
  return href.startsWith('/p/') || href.startsWith('/#')
}
</script>

<template>
  <header class="site-header">
    <div class="atm-container header-row">
      <SiteLogo variant="header" />

      <div class="nav-shell" @mouseleave="closeAll">
        <nav class="main-nav" aria-label="主导航">
          <div class="nav-item nav-item--mega" @mouseenter="openMega">
            <a
              href="/"
              class="nav-link nav-link--mega"
              :class="{ 'nav-link--active': megaOpen }"
              @click.prevent="openMega"
            >
              套餐购买
              <span class="nav-chevron" aria-hidden="true">▾</span>
            </a>

            <div
              v-show="megaOpen"
              class="mega-popover"
              @mouseenter="openMega"
            >
              <div class="mega-popover-inner">
                <NavMegaMenuSkeleton v-if="navLoading && !megaMenu.length && !brandMenus.length" />
                <div v-else class="mega-grid">
                  <div v-for="col in megaMenu" :key="col.title" class="mega-col">
                    <h3 class="mega-col-title">{{ col.title }}</h3>
                    <ul class="mega-list">
                      <li v-for="item in col.items" :key="item.label">
                        <RouterLink
                          v-if="isSpaNav(item.href)"
                          :to="item.href"
                          class="mega-link"
                          @click="closeAll"
                        >
                          <span class="mega-link-label">{{ item.label }}</span>
                          <span class="mega-link-price">{{ item.price }}</span>
                        </RouterLink>
                        <a
                          v-else
                          :href="item.href"
                          class="mega-link"
                          @click="closeAll"
                        >
                          <span class="mega-link-label">{{ item.label }}</span>
                          <span class="mega-link-price">{{ item.price }}</span>
                        </a>
                      </li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <template v-for="brand in navBrandEntries" :key="brand.id">
            <RouterLink
              v-if="!isNavDropdown(brand)"
              :to="brand.to"
              class="nav-link"
              :class="{ 'nav-link--active': route.path === brand.to }"
            >
              {{ brand.label }}
            </RouterLink>
            <div
              v-else
              class="nav-item"
              @mouseenter="openBrandDrop(brand.id)"
            >
              <button
                type="button"
                class="nav-link nav-link-btn"
                :class="{ 'nav-link--active': openDrop === brand.id }"
              >
                {{ brand.label }}
                <span v-if="brand.hotTagName" class="nav-hot-tag">{{ brand.hotTagName }}</span>
                <span class="nav-chevron" aria-hidden="true">▾</span>
              </button>
              <div
                v-show="openDrop === brand.id"
                class="nav-drop-panel"
                @mouseenter="openBrandDrop(brand.id)"
              >
                <div class="nav-drop-panel-inner">
                  <template v-for="item in brand.items" :key="item.label">
                    <RouterLink
                      v-if="isSpaNav(item.href)"
                      :to="item.href"
                      class="nav-drop-link"
                      @click="closeAll"
                    >
                      <span class="nav-drop-label">{{ item.label }}</span>
                      <span v-if="item.price" class="nav-drop-price">{{ item.price }}</span>
                    </RouterLink>
                    <a
                      v-else
                      :href="item.href"
                      class="nav-drop-link"
                      @click="closeAll"
                    >
                      <span class="nav-drop-label">{{ item.label }}</span>
                      <span v-if="item.price" class="nav-drop-price">{{ item.price }}</span>
                    </a>
                  </template>
                </div>
              </div>
            </div>
          </template>

          <RouterLink
            to="/blog"
            class="nav-link nav-link-plain"
            :class="{ 'nav-link--active': route.path.startsWith('/blog') }"
          >
            <span class="nav-link-plain-text">教程 FAQ</span>
          </RouterLink>
        </nav>
      </div>

      <div class="actions">
        <template v-if="loggedIn">
          <SiteNotificationBell />
          <RouterLink to="/member" class="atm-btn-primary btn-sm">会员中心</RouterLink>
        </template>
        <template v-else>
          <RouterLink to="/login" class="link-muted">登录</RouterLink>
          <RouterLink to="/register" class="atm-btn-primary btn-sm">注册</RouterLink>
        </template>
      </div>
    </div>
  </header>
</template>

<style scoped>
.site-header {
  position: sticky;
  top: 0;
  z-index: 200;
  background: var(--atm-bg);
  border-bottom: 1px solid #ede9fe;
}

.header-row {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 56px;
}

.nav-shell {
  position: relative;
  flex: 1;
  display: none;
  justify-content: flex-start;
  min-width: 0;
}

@media (min-width: 1024px) {
  .nav-shell {
    display: flex;
  }
}

.main-nav {
  display: flex;
  align-items: center;
  gap: 2px;
}

.nav-item {
  position: relative;
}

.nav-item--mega {
  position: relative;
}

.nav-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  font-size: 14px;
  font-weight: 500;
  color: var(--atm-text);
  text-decoration: none;
  border-radius: 999px;
  white-space: nowrap;
  transition:
    background 0.15s,
    color 0.15s;
}

.nav-link-btn {
  border: none;
  background: transparent;
  cursor: pointer;
  font-family: inherit;
}

.nav-link-plain {
  font-weight: 600;
}

.nav-link-plain-text {
  background: var(--atm-gradient);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.nav-link:hover,
.nav-link--active {
  color: var(--atm-primary);
  background: #ede9fe;
}

.nav-hot-tag {
  padding: 1px 6px;
  font-size: 10px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(90deg, #f97316, #fb923c);
  border-radius: 999px;
  line-height: 1.4;
}

.nav-chevron {
  font-size: 10px;
  color: var(--atm-text-muted);
}

.nav-panel-loading {
  margin: 0;
  padding: 12px 8px;
  font-size: 13px;
  color: var(--atm-text-muted);
}

.nav-drop-panel {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  min-width: 240px;
  padding-top: 10px;
  background: transparent;
  border: none;
  box-shadow: none;
  z-index: 210;
}

.nav-drop-panel-inner {
  padding: 6px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 14px;
  box-shadow:
    0 4px 6px rgba(30, 27, 75, 0.04),
    0 20px 48px rgba(30, 27, 75, 0.14);
}

.nav-drop-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 9px 12px;
  font-size: 13px;
  color: var(--atm-text);
  text-decoration: none;
  border-radius: 8px;
}

.nav-drop-link:hover {
  background: #f5f3ff;
  color: var(--atm-primary);
}

.nav-drop-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
}

.nav-drop-price {
  flex-shrink: 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-primary);
  white-space: nowrap;
}

/* 浮层：顶部透明 padding 过桥，避免鼠标从菜单项滑向面板时误触发 mouseleave */
.mega-popover {
  position: absolute;
  top: 100%;
  left: 0;
  width: min(1064px, calc(100vw - 40px));
  padding-top: 10px;
  background: transparent;
  border: none;
  box-shadow: none;
  z-index: 210;
}

.mega-popover-inner {
  padding: 22px 26px 18px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow:
    0 4px 8px rgba(30, 27, 75, 0.04),
    0 24px 64px rgba(30, 27, 75, 0.16);
}

.nav-link--sk {
  display: inline-block;
  width: 72px;
  min-height: 1.25em;
  padding: 0;
  border: none;
  border-radius: 8px;
  background: linear-gradient(90deg, #eef1f6 0%, #e2e8f0 45%, #eef1f6 90%);
  background-size: 200% 100%;
  animation: nav-link-sk-shimmer 1.2s ease-in-out infinite;
  pointer-events: none;
}

@keyframes nav-link-sk-shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: -100% 0;
  }
}

.mega-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(158px, 1fr));
  gap: 18px 28px;
  align-items: start;
}

@media (min-width: 1100px) {
  .mega-grid {
    grid-template-columns: repeat(5, minmax(172px, 1fr));
    gap: 16px 20px;
  }

  @supports (grid-template-rows: subgrid) {
    .mega-grid {
      grid-template-rows: auto 1fr;
    }

    .mega-col {
      grid-row: span 2;
      display: grid;
      grid-template-rows: subgrid;
    }
  }
}

.mega-col-title {
  margin: 0 0 8px;
  padding: 0 8px;
  min-height: 20px;
  line-height: 20px;
  font-size: 13px;
  font-weight: 700;
  color: var(--atm-text);
}

.mega-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.mega-col {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  min-width: 0;
}

.mega-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 7px 8px;
  font-size: 13px;
  line-height: 1.4;
  color: #64748b;
  text-decoration: none;
  border-radius: 10px;
  transition:
    background 0.12s,
    color 0.12s;
}

.mega-link:hover {
  background: #f8fafc;
  color: var(--atm-text);
}

.mega-link-label {
  flex: 1;
  min-width: 0;
  font-weight: 500;
  color: var(--atm-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mega-link-price {
  flex-shrink: 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-primary);
  white-space: nowrap;
  padding-left: 4px;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
  flex-shrink: 0;
}

.link-muted {
  font-size: 14px;
  color: var(--atm-text-muted);
  text-decoration: none;
}

.link-muted:hover {
  color: var(--atm-primary);
}

.btn-sm {
  padding: 8px 16px;
  font-size: 14px;
}

</style>
