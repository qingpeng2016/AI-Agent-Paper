import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior(to, from, savedPosition) {
    const coldEntry = from.matched.length === 0
    // 首页：首次打开 / 刷新不要跟 hash 或历史滚动，首屏停在 Hero
    if (to.path === '/') {
      if (coldEntry || !to.hash) return { top: 0, left: 0 }
      return { el: to.hash, top: 80, behavior: 'smooth' }
    }
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash, top: 80, behavior: 'smooth' }
    // 会员中心等同一路径只改 ?tab= 时不滚回顶部，避免侧栏/整页上下跳
    if (from.path && to.path === from.path) return false
    return { top: 0, left: 0 }
  },
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/DefaultLayout.vue'),
      children: [
        { path: '', name: 'home', component: () => import('@/views/HomeView.vue') },
        {
          path: 'p/:slug',
          name: 'product',
          component: () => import('@/views/ProductDetailView.vue'),
        },
        {
          path: 'member',
          name: 'member',
          component: () => import('@/views/MemberCenterView.vue'),
        },
        {
          path: 'enterprise',
          name: 'enterprise',
          component: () => import('@/views/EnterpriseView.vue'),
        },
        {
          path: 'blog',
          name: 'blog',
          component: () => import('@/views/BlogView.vue'),
        },
        {
          path: 'blog/:slug',
          name: 'blog-article',
          component: () => import('@/views/BlogArticleView.vue'),
        },
      ],
    },
    {
      path: '/workbench',
      component: () => import('@/layouts/WorkbenchLayout.vue'),
      children: [
        {
          path: '',
          name: 'paper-workbench',
          component: () => import('@/views/paper/PaperWorkbenchView.vue'),
          meta: { requiresAuth: true },
        },
      ],
    },
    {
      path: '/',
      component: () => import('@/layouts/AuthLayout.vue'),
      children: [
        { path: 'login', name: 'login', component: () => import('@/views/LoginView.vue') },
        { path: 'register', name: 'register', component: () => import('@/views/RegisterView.vue') },
      ],
    },
  ],
})

export default router
