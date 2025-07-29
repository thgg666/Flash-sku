import { djangoClient, goClient } from './http'
import type {
  SeckillActivity,
  Product,
  Category,
  StockInfo,
  PaginatedResponse
} from '@/types'

/**
 * 商品相关API (Django)
 */
export const productApi = {
  /**
   * 获取商品列表
   */
  getProducts(params?: {
    page?: number
    page_size?: number
    category_id?: number
    search?: string
  }): Promise<PaginatedResponse<Product>> {
    return djangoClient.get('/products/', { params })
  },

  /**
   * 获取商品详情
   */
  getProduct(id: number): Promise<Product> {
    return djangoClient.get(`/products/${id}/`)
  },

  /**
   * 获取商品分类列表
   */
  getCategories(): Promise<Category[]> {
    return djangoClient.get('/categories/')
  },

  /**
   * 获取分类详情
   */
  getCategory(id: number): Promise<Category> {
    return djangoClient.get(`/categories/${id}/`)
  }
}

/**
 * 秒杀活动相关API (Django)
 */
export const activityApi = {
  /**
   * 获取秒杀活动列表
   */
  getActivities(params?: {
    page?: number
    page_size?: number
    status?: 'pending' | 'active' | 'ended' | 'cancelled'
    product_id?: number
    search?: string
  }): Promise<PaginatedResponse<SeckillActivity>> {
    return djangoClient.get('/activities/', { params })
  },

  /**
   * 获取活动详情
   */
  getActivity(id: number): Promise<SeckillActivity> {
    return djangoClient.get(`/activities/${id}/`)
  },

  /**
   * 获取即将开始的活动
   */
  getUpcomingActivities(limit = 10): Promise<SeckillActivity[]> {
    return djangoClient.get('/activities/upcoming/', { 
      params: { limit } 
    })
  },

  /**
   * 获取正在进行的活动
   */
  getActiveActivities(limit = 10): Promise<SeckillActivity[]> {
    return djangoClient.get('/activities/active/', { 
      params: { limit } 
    })
  },

  /**
   * 获取热门活动
   */
  getHotActivities(limit = 10): Promise<SeckillActivity[]> {
    return djangoClient.get('/activities/hot/', { 
      params: { limit } 
    })
  },

  /**
   * 搜索活动
   */
  searchActivities(query: string, params?: {
    page?: number
    page_size?: number
  }): Promise<PaginatedResponse<SeckillActivity>> {
    return djangoClient.get('/activities/search/', { 
      params: { q: query, ...params } 
    })
  }
}

/**
 * 秒杀核心API (Go服务)
 */
export const seckillApi = {
  /**
   * 参与秒杀 - 增强版本，支持防重复点击和重试
   */
  async participate(activityId: number, options?: {
    retryCount?: number
    retryDelay?: number
    timeout?: number
  }): Promise<{
    code: 'SUCCESS' | 'FAILURE' | 'RATE_LIMIT' | 'SOLD_OUT' | 'LIMIT_EXCEEDED' | 'ACTIVITY_ENDED' | 'INSUFFICIENT_STOCK'
    message: string
    order_id?: string
    quantity?: number
    amount?: string
    data?: any
    request_id?: string
    timestamp?: number
  }> {
    const {
      retryCount = 3,
      retryDelay = 1000,
      timeout = 5000
    } = options || {}

    // 生成请求ID，防止重复提交
    const requestId = `seckill_${activityId}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`

    // 检查是否有正在进行的相同请求
    const pendingKey = `seckill_pending_${activityId}`
    if (sessionStorage.getItem(pendingKey)) {
      throw new Error('秒杀请求正在处理中，请勿重复点击')
    }

    // 标记请求开始
    sessionStorage.setItem(pendingKey, requestId)

    try {
      let lastError: Error | null = null

      for (let attempt = 0; attempt <= retryCount; attempt++) {
        try {
          const response = await goClient.post(`/seckill/${activityId}`, {
            request_id: requestId,
            timestamp: Date.now()
          }, {
            timeout,
            headers: {
              'X-Request-ID': requestId,
              'X-Retry-Attempt': attempt.toString()
            }
          })

          // 成功响应，清除pending标记
          sessionStorage.removeItem(pendingKey)

          return {
            ...response,
            request_id: requestId,
            timestamp: Date.now()
          }
        } catch (error: any) {
          lastError = error

          // 如果是业务错误（非网络错误），不重试
          if (error.response?.status === 400 || error.response?.status === 409) {
            break
          }

          // 如果不是最后一次尝试，等待后重试
          if (attempt < retryCount) {
            await new Promise(resolve => setTimeout(resolve, retryDelay * (attempt + 1)))
          }
        }
      }

      // 所有重试都失败，清除pending标记并抛出错误
      sessionStorage.removeItem(pendingKey)
      throw lastError || new Error('秒杀请求失败')

    } catch (error) {
      // 确保清除pending标记
      sessionStorage.removeItem(pendingKey)
      throw error
    }
  },

  /**
   * 检查秒杀请求状态
   */
  async checkSeckillStatus(requestId: string): Promise<{
    status: 'pending' | 'processing' | 'success' | 'failed'
    message: string
    order_id?: string
    error_code?: string
  }> {
    return goClient.get(`/seckill/status/${requestId}`)
  },

  /**
   * 取消秒杀请求
   */
  async cancelSeckill(requestId: string): Promise<{
    success: boolean
    message: string
  }> {
    return goClient.post(`/seckill/cancel/${requestId}`)
  },

  /**
   * 获取实时库存信息
   */
  getStock(activityId: number): Promise<StockInfo> {
    return goClient.get(`/seckill/stock/${activityId}`)
  },

  /**
   * 批量获取库存信息
   */
  getBatchStock(activityIds: number[]): Promise<Record<number, StockInfo>> {
    return goClient.post('/seckill/stock/batch', { activity_ids: activityIds })
  },

  /**
   * 获取用户参与记录
   */
  getUserParticipation(activityId: number): Promise<{
    participated: boolean
    participation_count: number
    max_allowed: number
    can_participate: boolean
  }> {
    return goClient.get(`/seckill/user/${activityId}`)
  },

  /**
   * 获取秒杀统计信息
   */
  getStatistics(activityId: number): Promise<{
    total_participants: number
    success_count: number
    success_rate: number
    peak_qps: number
    average_response_time: number
  }> {
    return goClient.get(`/seckill/stats/${activityId}`)
  },

  /**
   * 预热活动缓存
   */
  warmupActivity(activityId: number): Promise<{ message: string }> {
    return goClient.post(`/seckill/warmup/${activityId}`)
  },

  /**
   * 获取系统健康状态
   */
  getHealthStatus(): Promise<{
    status: 'healthy' | 'degraded' | 'unhealthy'
    timestamp: string
    redis: boolean
    database: boolean
    rabbitmq: boolean
    services: Record<string, boolean>
  }> {
    return goClient.get('/health')
  },

  /**
   * 获取限流状态
   */
  getRateLimitStatus(): Promise<{
    global_limit: number
    global_remaining: number
    ip_limit: number
    ip_remaining: number
    user_limit: number
    user_remaining: number
    reset_time: number
  }> {
    return goClient.get('/seckill/rate-limit/status')
  }
}

/**
 * WebSocket相关API
 */
export const websocketApi = {
  /**
   * 获取WebSocket连接URL
   */
  getWebSocketUrl(): string {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsHost = import.meta.env.VITE_WS_HOST || window.location.host
    return `${wsProtocol}//${wsHost}/ws/seckill/`
  },

  /**
   * 获取活动专用WebSocket URL
   */
  getActivityWebSocketUrl(activityId: number): string {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsHost = import.meta.env.VITE_WS_HOST || window.location.host
    return `${wsProtocol}//${wsHost}/ws/seckill/${activityId}/`
  }
}

/**
 * 导出所有API
 */
export default {
  product: productApi,
  activity: activityApi,
  seckill: seckillApi,
  websocket: websocketApi
}
