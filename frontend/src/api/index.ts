// 导出HTTP客户端
export { djangoClient, goClient, djangoApi, goApi } from './http'

// 导出认证API
export { authApi } from './auth'

// 导出秒杀相关API
export { 
  productApi, 
  activityApi, 
  seckillApi, 
  websocketApi 
} from './seckill'

// 统一API对象
export const api = {
  auth: () => import('./auth').then(m => m.authApi),
  product: () => import('./seckill').then(m => m.productApi),
  activity: () => import('./seckill').then(m => m.activityApi),
  seckill: () => import('./seckill').then(m => m.seckillApi),
  websocket: () => import('./seckill').then(m => m.websocketApi),
}

// 默认导出
export default api
