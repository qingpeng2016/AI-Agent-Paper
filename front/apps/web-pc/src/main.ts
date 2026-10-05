import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

// 主题色对齐品牌紫
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import './styles/global.css'
import './styles/theme.css'
import './styles/auth-form.css'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { installRouteDataGuards } from '@/bootstrap/routeGuards'
import { installTracking } from '@/tracking/installTracking'

if (typeof history !== 'undefined' && 'scrollRestoration' in history) {
  history.scrollRestoration = 'manual'
}

const app = createApp(App)
app.use(createPinia())
app.use(router)
installRouteDataGuards(router)
installTracking(router)
app.use(ElementPlus, { locale: zhCn })
app.mount('#app')
