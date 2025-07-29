import { ref, computed, onUnmounted, readonly } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { orderApi } from '@/api/order'

/**
 * 订单状态枚举
 */
export enum OrderStatus {
  PENDING_PAYMENT = 'pending_payment',
  PAID = 'paid',
  CANCELLED = 'cancelled',
  EXPIRED = 'expired',
  PROCESSING = 'processing',
  COMPLETED = 'completed'
}

/**
 * 轮询配置
 */
interface PollingOptions {
  /** 轮询间隔（毫秒） */
  interval?: number
  /** 最大轮询次数 */
  maxAttempts?: number
  /** 超时时间（毫秒） */
  timeout?: number
  /** 是否自动开始轮询 */
  autoStart?: boolean
  /** 成功回调 */
  onSuccess?: (order: any) => void
  /** 失败回调 */
  onError?: (error: any) => void
  /** 状态变化回调 */
  onStatusChange?: (status: string, order: any) => void
}

/**
 * 订单轮询状态
 */
interface PollingState {
  isPolling: boolean
  attempts: number
  lastUpdate: number
  error: string | null
}

/**
 * 订单状态轮询组合式函数
 */
export function useOrderPolling(options: PollingOptions = {}) {
  const {
    interval = 2000,
    maxAttempts = 30,
    timeout = 60000,
    autoStart = false,
    onSuccess,
    onError,
    onStatusChange
  } = options

  // 响应式状态
  const pollingState = ref<PollingState>({
    isPolling: false,
    attempts: 0,
    lastUpdate: 0,
    error: null
  })

  const currentOrder = ref<any>(null)
  const orderStatus = ref<string>('')
  const remainingTime = ref(0)

  // 定时器
  let pollingTimer: number | null = null
  let timeoutTimer: number | null = null
  let countdownTimer: number | null = null

  // 计算属性
  const isPolling = computed(() => pollingState.value.isPolling)
  const canPoll = computed(() => !isPolling.value && pollingState.value.attempts < maxAttempts)
  const progress = computed(() => {
    if (maxAttempts === 0) return 0
    return Math.min((pollingState.value.attempts / maxAttempts) * 100, 100)
  })

  /**
   * 开始轮询订单状态
   */
  const startPolling = async (orderId: string): Promise<void> => {
    if (isPolling.value) {
      console.warn('轮询已在进行中')
      return
    }

    // 重置状态
    pollingState.value = {
      isPolling: true,
      attempts: 0,
      lastUpdate: Date.now(),
      error: null
    }

    currentOrder.value = null
    orderStatus.value = ''

    console.log(`开始轮询订单状态: ${orderId}`)

    // 设置超时定时器
    if (timeout > 0) {
      timeoutTimer = setTimeout(() => {
        stopPolling('轮询超时')
      }, timeout)
    }

    // 开始轮询
    await pollOrderStatus(orderId)
  }

  /**
   * 轮询订单状态
   */
  const pollOrderStatus = async (orderId: string): Promise<void> => {
    if (!isPolling.value) {
      return
    }

    try {
      pollingState.value.attempts++
      pollingState.value.lastUpdate = Date.now()

      console.log(`轮询订单状态 - 第${pollingState.value.attempts}次尝试`)

      // 调用API获取订单状态
      const response = await orderApi.pollOrderStatus(orderId, pollingState.value.attempts)
      
      if (response.success && response.order) {
        const order = response.order
        const newStatus = order.status
        
        // 更新订单信息
        currentOrder.value = order
        
        // 检查状态是否变化
        if (orderStatus.value !== newStatus) {
          const oldStatus = orderStatus.value
          orderStatus.value = newStatus
          
          console.log(`订单状态变化: ${oldStatus} -> ${newStatus}`)
          
          // 触发状态变化回调
          if (onStatusChange) {
            onStatusChange(newStatus, order)
          }
          
          // 显示状态变化通知
          showStatusNotification(newStatus, order)
        }

        // 更新剩余时间
        if (order.payment_deadline) {
          updateRemainingTime(order.payment_deadline)
        }

        // 检查是否需要停止轮询
        if (shouldStopPolling(newStatus)) {
          const message = getStatusMessage(newStatus)
          stopPolling(message)
          
          if (isSuccessStatus(newStatus)) {
            onSuccess?.(order)
          }
          return
        }

      } else {
        throw new Error(response.message || '获取订单状态失败')
      }

      // 继续轮询
      scheduleNextPoll(orderId)

    } catch (error: any) {
      console.error('轮询订单状态失败:', error)
      
      pollingState.value.error = error.message || '网络错误'
      
      // 如果达到最大尝试次数，停止轮询
      if (pollingState.value.attempts >= maxAttempts) {
        stopPolling('达到最大轮询次数')
        onError?.(error)
        return
      }

      // 继续轮询
      scheduleNextPoll(orderId)
    }
  }

  /**
   * 安排下次轮询
   */
  const scheduleNextPoll = (orderId: string): void => {
    if (!isPolling.value) {
      return
    }

    pollingTimer = setTimeout(() => {
      pollOrderStatus(orderId)
    }, interval)
  }

  /**
   * 停止轮询
   */
  const stopPolling = (reason?: string): void => {
    console.log(`停止轮询订单状态${reason ? `: ${reason}` : ''}`)
    
    pollingState.value.isPolling = false
    
    // 清除定时器
    if (pollingTimer) {
      clearTimeout(pollingTimer)
      pollingTimer = null
    }
    
    if (timeoutTimer) {
      clearTimeout(timeoutTimer)
      timeoutTimer = null
    }
    
    if (countdownTimer) {
      clearInterval(countdownTimer)
      countdownTimer = null
    }
  }

  /**
   * 判断是否应该停止轮询
   */
  const shouldStopPolling = (status: string): boolean => {
    const finalStatuses = [
      OrderStatus.PAID,
      OrderStatus.CANCELLED,
      OrderStatus.EXPIRED,
      OrderStatus.COMPLETED
    ]
    return finalStatuses.includes(status as OrderStatus)
  }

  /**
   * 判断是否为成功状态
   */
  const isSuccessStatus = (status: string): boolean => {
    return status === OrderStatus.PAID || status === OrderStatus.COMPLETED
  }

  /**
   * 获取状态消息
   */
  const getStatusMessage = (status: string): string => {
    const messages: Record<string, string> = {
      [OrderStatus.PENDING_PAYMENT]: '等待支付',
      [OrderStatus.PAID]: '支付成功',
      [OrderStatus.CANCELLED]: '订单已取消',
      [OrderStatus.EXPIRED]: '订单已过期',
      [OrderStatus.PROCESSING]: '处理中',
      [OrderStatus.COMPLETED]: '订单完成'
    }
    return messages[status] || '未知状态'
  }

  /**
   * 显示状态通知
   */
  const showStatusNotification = (status: string, order: any): void => {
    const message = getStatusMessage(status)
    
    switch (status) {
      case OrderStatus.PAID:
        ElNotification({
          title: '支付成功',
          message: '订单支付成功，感谢您的购买！',
          type: 'success',
          duration: 5000
        })
        break
        
      case OrderStatus.CANCELLED:
        ElMessage.warning(`订单已取消: ${message}`)
        break
        
      case OrderStatus.EXPIRED:
        ElMessage.error('订单已过期，请重新下单')
        break
        
      default:
        console.log(`订单状态更新: ${message}`)
    }
  }

  /**
   * 更新剩余时间
   */
  const updateRemainingTime = (deadline: string): void => {
    const deadlineTime = new Date(deadline).getTime()
    
    const updateCountdown = () => {
      const now = Date.now()
      const remaining = Math.max(0, deadlineTime - now)
      remainingTime.value = Math.floor(remaining / 1000)
      
      if (remaining <= 0) {
        if (countdownTimer) {
          clearInterval(countdownTimer)
          countdownTimer = null
        }
      }
    }
    
    updateCountdown()
    
    if (countdownTimer) {
      clearInterval(countdownTimer)
    }
    
    countdownTimer = setInterval(updateCountdown, 1000)
  }

  /**
   * 重置状态
   */
  const reset = (): void => {
    stopPolling()
    
    pollingState.value = {
      isPolling: false,
      attempts: 0,
      lastUpdate: 0,
      error: null
    }
    
    currentOrder.value = null
    orderStatus.value = ''
    remainingTime.value = 0
  }

  // 组件卸载时清理
  onUnmounted(() => {
    reset()
  })

  return {
    // 状态
    pollingState: readonly(pollingState),
    currentOrder: readonly(currentOrder),
    orderStatus: readonly(orderStatus),
    remainingTime: readonly(remainingTime),
    
    // 计算属性
    isPolling,
    canPoll,
    progress,
    
    // 方法
    startPolling,
    stopPolling,
    reset
  }
}
