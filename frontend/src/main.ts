// 导入polyfills (必须在最前面)
import '@/utils/polyfills'

// 导入样式
import './assets/main.css'
import '@/styles/global.scss'
import '@/styles/mobile.scss'
import '@/styles/tablet.scss'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

// Element Plus
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

import App from './App.vue'
import router from './router'
import { initBrowserCompatibility } from '@/utils/browserCompatibility'
import { initAccessibility } from '@/utils/accessibility'
import { initPerformanceMonitoring } from '@/utils/performance'
import { initResourceOptimization } from '@/utils/resourceOptimization'
import { initServiceWorker } from '@/utils/serviceWorker'
import { vLazyLoad } from '@/directives/lazyLoad'

// 初始化浏览器兼容性检测
initBrowserCompatibility()

// 初始化可访问性功能
initAccessibility()

// 初始化性能监控
initPerformanceMonitoring()

// 初始化资源优化
initResourceOptimization({
  baseUrl: import.meta.env.VITE_CDN_BASE_URL || '',
  fallbackUrls: [
    import.meta.env.VITE_CDN_FALLBACK_URL || ''
  ].filter(Boolean)
})

// 初始化Service Worker
if (import.meta.env.PROD) {
  initServiceWorker()
}

const app = createApp(App)

// 注册Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 配置Element Plus中文
app.use(ElementPlus, {
  locale: zhCn,
})

const pinia = createPinia()
app.use(pinia)
app.use(router)

// 注册全局指令
app.directive('lazy-load', vLazyLoad)

// 初始化认证状态
import { useAuthStore } from '@/stores/auth'
const authStore = useAuthStore()
// 重新启用认证初始化（只读取本地存储，不发起网络请求）
authStore.initAuth()

app.mount('#app')
