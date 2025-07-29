import { djangoClient } from './http'

/**
 * 订单状态类型
 */
export type OrderStatus = 
  | 'pending_payment'
  | 'paid'
  | 'cancelled'
  | 'expired'
  | 'processing'
  | 'completed'

/**
 * 订单信息类型
 */
export interface Order {
  id: number
  user_id: number
  activity_id: number
  product_name: string
  seckill_price: string
  quantity: number
  total_amount: string
  status: OrderStatus
  payment_deadline: string | null
  remaining_time: number
  created_at: string
  updated_at: string
  can_pay: boolean
  can_cancel: boolean
}

/**
 * 订单API响应类型
 */
export interface OrderResponse {
  success: boolean
  message?: string
  order?: Order
  orders?: Order[]
  total?: number
}

/**
 * 创建订单请求类型
 */
export interface CreateOrderRequest {
  activity_id: number
  quantity?: number
}

/**
 * 订单API
 */
export const orderApi = {
  /**
   * 创建秒杀订单
   */
  async createSeckillOrder(data: CreateOrderRequest): Promise<{
    success: boolean
    message: string
    task_id?: string
    code?: string
  }> {
    return djangoClient.post('/orders/seckill/', data)
  },

  /**
   * 检查订单创建状态
   */
  async checkOrderStatus(taskId: string): Promise<{
    task_id: string
    status: 'pending' | 'completed' | 'failed'
    result?: any
    error?: string
  }> {
    return djangoClient.get(`/orders/status/${taskId}/`)
  },

  /**
   * 获取订单状态
   */
  async getOrderStatus(orderId: string): Promise<OrderResponse> {
    try {
      const response = await djangoClient.get(`/orders/${orderId}/`)
      return {
        success: true,
        order: response
      }
    } catch (error: any) {
      return {
        success: false,
        message: error.message || '获取订单状态失败'
      }
    }
  },

  /**
   * 获取用户订单列表
   */
  async getUserOrders(): Promise<OrderResponse> {
    try {
      const response = await djangoClient.get('/orders/')
      return {
        success: response.success || true,
        orders: response.orders || response,
        total: response.total || response.length
      }
    } catch (error: any) {
      return {
        success: false,
        message: error.message || '获取订单列表失败',
        orders: []
      }
    }
  },

  /**
   * 取消订单
   */
  async cancelOrder(orderId: number): Promise<{
    success: boolean
    message: string
    order_id?: number
  }> {
    return djangoClient.post(`/orders/${orderId}/cancel/`)
  },

  /**
   * 获取订单详情
   */
  async getOrderDetail(orderId: number): Promise<OrderResponse> {
    try {
      const response = await djangoClient.get(`/orders/${orderId}/`)
      return {
        success: true,
        order: response
      }
    } catch (error: any) {
      return {
        success: false,
        message: error.message || '获取订单详情失败'
      }
    }
  },

  /**
   * 轮询订单状态 - 专门用于轮询
   */
  async pollOrderStatus(orderId: string, attempt: number = 1): Promise<OrderResponse> {
    try {
      const response = await djangoClient.get(`/orders/${orderId}/`, {
        headers: {
          'X-Poll-Attempt': attempt.toString(),
          'X-Poll-Timestamp': Date.now().toString()
        },
        timeout: 5000 // 轮询请求使用较短的超时时间
      })
      
      return {
        success: true,
        order: response.order || response
      }
    } catch (error: any) {
      return {
        success: false,
        message: error.message || '轮询订单状态失败'
      }
    }
  },

  /**
   * 批量获取订单状态
   */
  async batchGetOrderStatus(orderIds: string[]): Promise<{
    success: boolean
    orders: Record<string, Order>
    errors: Record<string, string>
  }> {
    try {
      const response = await djangoClient.post('/orders/batch-status/', {
        order_ids: orderIds
      })
      
      return {
        success: true,
        orders: response.orders || {},
        errors: response.errors || {}
      }
    } catch (error: any) {
      return {
        success: false,
        orders: {},
        errors: orderIds.reduce((acc, id) => {
          acc[id] = error.message || '获取状态失败'
          return acc
        }, {} as Record<string, string>)
      }
    }
  }
}

export default orderApi
