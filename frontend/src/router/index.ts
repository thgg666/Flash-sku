import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { storage } from '@/utils'
import type { RouteMeta } from '@/types'

// 路由配置
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/SimpleHome.vue'),
    meta: {
      title: '首页',
      requiresAuth: false,
    } as RouteMeta,
  },
  {
    path: '/about',
    name: 'about',
    component: () => import('@/views/AboutView.vue'),
    meta: {
      title: '关于我们',
      requiresAuth: false,
    } as RouteMeta,
  },
  {
    path: '/activities',
    name: 'activities',
    component: () => import('@/views/ActivityListView.vue'),
    meta: {
      title: '秒杀活动',
      requiresAuth: false,
    } as RouteMeta,
  },
  {
    path: '/activity/:id',
    name: 'activity-detail',
    component: () => import('@/views/ActivityDetailView.vue'),
    meta: {
      title: '活动详情',
      requiresAuth: false,
    } as RouteMeta,
  },
  {
    path: '/seckill-demo',
    name: 'seckill-demo',
    component: () => import('@/views/SeckillDemoView.vue'),
    meta: {
      title: '秒杀演示',
      requiresAuth: false,
    } as RouteMeta,
  },
  {
    path: '/auth',
    name: 'auth',
    redirect: '/auth/login',
    children: [
      {
        path: 'login',
        name: 'login',
        component: () => import('@/views/auth/LoginView.vue'),
        meta: {
          title: '用户登录',
          requiresAuth: false,
        } as RouteMeta,
      },
      {
        path: 'register',
        name: 'register',
        component: () => import('@/views/auth/RegisterView.vue'),
        meta: {
          title: '用户注册',
          requiresAuth: false,
        } as RouteMeta,
      },
      {
        path: 'verify-email',
        name: 'verify-email',
        component: () => import('@/views/auth/VerifyEmailView.vue'),
        meta: {
          title: '邮箱验证',
          requiresAuth: false,
        } as RouteMeta,
      },
      {
        path: 'forgot-password',
        name: 'forgot-password',
        component: () => import('@/views/auth/ForgotPasswordView.vue'),
        meta: {
          title: '忘记密码',
          requiresAuth: false,
        } as RouteMeta,
      },
      {
        path: 'reset-password',
        name: 'reset-password',
        component: () => import('@/views/auth/ResetPasswordView.vue'),
        meta: {
          title: '重置密码',
          requiresAuth: false,
        } as RouteMeta,
      },
    ],
  },
  {
    path: '/user',
    name: 'user',
    redirect: '/user/profile',
    meta: {
      requiresAuth: true,
    } as RouteMeta,
    children: [
      {
        path: 'profile',
        name: 'user-profile',
        component: () => import('@/views/user/ProfileView.vue'),
        meta: {
          title: '个人中心',
          requiresAuth: true,
        } as RouteMeta,
      },
    ],
  },
  // 其他路由将在后续任务中添加
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

// 开发环境专用路由
if (import.meta.env.DEV) {
  routes.push({
    path: '/dev',
    name: 'dev',
    children: [
      {
        path: 'device-test',
        name: 'device-test',
        component: () => import('@/views/DeviceTestView.vue'),
        meta: {
          title: '设备兼容性测试',
          requiresAuth: false,
        } as RouteMeta,
      },
      {
        path: 'browser-test',
        name: 'browser-test',
        component: () => import('@/views/BrowserTestView.vue'),
        meta: {
          title: '浏览器兼容性测试',
          requiresAuth: false,
        } as RouteMeta,
      },
    ],
  })
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  },
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta?.title) {
    document.title = `${to.meta.title} - ${import.meta.env.VITE_APP_TITLE}`
  }

  // 检查是否需要认证
  if (to.meta?.requiresAuth) {
    const token = storage.get('access_token')
    if (!token) {
      // 未登录，跳转到登录页
      next({
        name: 'login',
        query: { redirect: to.fullPath },
      })
      return
    }
  }

  // 如果已登录用户访问登录/注册页，重定向到首页
  if (['login', 'register'].includes(to.name as string)) {
    const token = storage.get('access_token')
    if (token) {
      next({ name: 'home' })
      return
    }
  }

  next()
})

export default router
