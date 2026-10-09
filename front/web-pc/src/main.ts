import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { installAuthRouteGuard } from './bootstrap/routeGuards'
import './styles/global.css'
import './styles/theme.css'
import './styles/paper-message-box.css'
import './styles/auth-form.css'

if (typeof history !== 'undefined' && 'scrollRestoration' in history) {
  history.scrollRestoration = 'manual'
}

const app = createApp(App)
app.use(createPinia())
app.use(router)
installAuthRouteGuard(router)
app.use(ElementPlus, { locale: zhCn })
app.mount('#app')
