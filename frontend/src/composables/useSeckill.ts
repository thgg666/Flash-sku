import { ref, computed, onUnmounted, readonly } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { seckillApi } from '@/api/seckill'
import { useOrderPolling } from './useOrderPolling'
import type { SeckillActivity } from '@/types'

/**
 * 秒杀状态枚举
 */
export enum SeckillStatus {
  IDLE = 'idle',           // 空闲状态
  PREPARING = 'preparing', // 准备中
  REQUESTING = 'requesting', // 请求中
  SUCCESS = 'success',     // 成功
  FAILED = 'failed',       // 失败
  CANCELLED = 'cancelled'  // 已取消
}

/**
 * 秒杀结果类型
 */
export interface SeckillResult {
  success: boolean
  message: string
  orderId?: string
  requestId?: string
  code?: string
  timestamp?: number
}

/**
 * 秒杀组合式函数
 */
export function useSeckill(activity?: SeckillActivity) {
  // 响应式状态
  const status = ref<SeckillStatus>(SeckillStatus.IDLE)
  const loading = ref(false)
  const result = ref<SeckillResult | null>(null)
  const error = ref<string | null>(null)
  const requestId = ref<string | null>(null)
  const countdown = ref(0)
  const orderId = ref<string | null>(null)

  // 防重复点击相关
  const lastClickTime = ref(0)
  const clickCooldown = 1000 // 1秒冷却时间

  // 定时器
  let statusCheckTimer: number | null = null
  let countdownTimer: number | null = null

  // 订单轮询
  const orderPolling = useOrderPolling({
    interval: 2000,
    maxAttempts: 30,
    onSuccess: (order) => {
      console.log('订单轮询成功:', order)
      ElNotification({
        title: '订单状态更新',
        message: '订单已创建成功，请及时支付',
        type: 'success',
        duration: 5000
      })
    },
    onError: (error) => {
      console.error('订单轮询失败:', error)
      ElMessage.error('订单状态查询失败')
    },
    onStatusChange: (status, order) => {
      console.log('订单状态变化:', status, order)
    }
  })

  // 计算属性
  const canParticipate = computed(() => {
    return status.value === SeckillStatus.IDLE && 
           !loading.value && 
           activity?.status === 'active'
  })

  const isProcessing = computed(() => {
    return status.value === SeckillStatus.REQUESTING || 
           status.value === SeckillStatus.PREPARING
  })

  const buttonText = computed(() => {
    switch (status.value) {
      case SeckillStatus.PREPARING:
        return '准备中...'
      case SeckillStatus.REQUESTING:
        return '抢购中...'
      case SeckillStatus.SUCCESS:
        return '抢购成功'
      case SeckillStatus.FAILED:
        return '抢购失败'
      case SeckillStatus.CANCELLED:
        return '已取消'
      default:
        return '立即抢购'
    }
  })

  const buttonType = computed(() => {
    switch (status.value) {
      case SeckillStatus.SUCCESS:
        return 'success'
      case SeckillStatus.FAILED:
        return 'danger'
      case SeckillStatus.CANCELLED:
        return 'info'
      default:
        return 'primary'
    }
  })

  /**
   * 防重复点击检查
   */
  const checkClickCooldown = (): boolean => {
    const now = Date.now()
    if (now - lastClickTime.value < clickCooldown) {
      ElMessage.warning('请勿频繁点击')
      return false
    }
    lastClickTime.value = now
    return true
  }

  /**
   * 参与秒杀
   */
  const participate = async (activityId?: number): Promise<SeckillResult> => {
    if (!activityId && !activity?.id) {
      throw new Error('活动ID不能为空')
    }

    const targetActivityId = activityId || activity!.id

    // 检查点击冷却
    if (!checkClickCooldown()) {
      return {
        success: false,
        message: '请勿频繁点击'
      }
    }

    // 检查是否可以参与
    if (!canParticipate.value) {
      const message = loading.value ? '请求正在处理中' : '当前无法参与秒杀'
      return {
        success: false,
        message
      }
    }

    try {
      // 重置状态
      error.value = null
      result.value = null
      
      // 设置状态
      status.value = SeckillStatus.PREPARING
      loading.value = true

      ElMessage.info('正在提交秒杀请求...')

      // 调用秒杀API
      status.value = SeckillStatus.REQUESTING
      const response = await seckillApi.participate(targetActivityId, {
        retryCount: 2,
        retryDelay: 500,
        timeout: 3000
      })

      // 保存请求ID
      requestId.value = response.request_id || null

      // 处理响应
      if (response.code === 'SUCCESS') {
        status.value = SeckillStatus.SUCCESS
        result.value = {
          success: true,
          message: response.message || '抢购成功！',
          orderId: response.order_id,
          requestId: response.request_id,
          code: response.code,
          timestamp: response.timestamp
        }

        ElNotification({
          title: '抢购成功',
          message: '恭喜您抢购成功！请及时完成支付。',
          type: 'success',
          duration: 5000
        })

        // 如果有订单ID，开始轮询订单状态
        if (response.order_id) {
          orderId.value = response.order_id
          orderPolling.startPolling(response.order_id)
        }

      } else {
        status.value = SeckillStatus.FAILED
        const errorMessage = getErrorMessage(response.code, response.message)
        
        result.value = {
          success: false,
          message: errorMessage,
          code: response.code,
          timestamp: response.timestamp
        }

        ElMessage.error(errorMessage)
      }

      return result.value

    } catch (err: any) {
      status.value = SeckillStatus.FAILED
      const errorMessage = err.message || '网络错误，请重试'
      error.value = errorMessage

      result.value = {
        success: false,
        message: errorMessage
      }

      ElMessage.error(errorMessage)
      return result.value

    } finally {
      loading.value = false
      
      // 3秒后重置状态
      setTimeout(() => {
        if (status.value !== SeckillStatus.SUCCESS) {
          reset()
        }
      }, 3000)
    }
  }

  /**
   * 获取错误信息
   */
  const getErrorMessage = (code: string, defaultMessage: string): string => {
    const errorMessages: Record<string, string> = {
      'RATE_LIMIT': '请求过于频繁，请稍后再试',
      'SOLD_OUT': '商品已售罄',
      'LIMIT_EXCEEDED': '超出购买限制',
      'ACTIVITY_ENDED': '活动已结束',
      'INSUFFICIENT_STOCK': '库存不足',
      'FAILURE': '抢购失败，请重试'
    }
    
    return errorMessages[code] || defaultMessage || '抢购失败'
  }



  /**
   * 取消秒杀请求
   */
  const cancel = async (): Promise<void> => {
    if (!requestId.value || !isProcessing.value) {
      return
    }

    try {
      await seckillApi.cancelSeckill(requestId.value)
      status.value = SeckillStatus.CANCELLED
      ElMessage.info('已取消秒杀请求')
    } catch (err: any) {
      console.error('取消秒杀请求失败:', err)
      ElMessage.error('取消请求失败')
    }
  }

  /**
   * 重置状态
   */
  const reset = (): void => {
    status.value = SeckillStatus.IDLE
    loading.value = false
    result.value = null
    error.value = null
    requestId.value = null
    countdown.value = 0
    
    // 清除定时器
    if (statusCheckTimer) {
      clearInterval(statusCheckTimer)
      statusCheckTimer = null
    }
    if (countdownTimer) {
      clearInterval(countdownTimer)
      countdownTimer = null
    }
  }

  /**
   * 开始倒计时
   */
  const startCountdown = (seconds: number): void => {
    countdown.value = seconds
    
    if (countdownTimer) {
      clearInterval(countdownTimer)
    }
    
    countdownTimer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        clearInterval(countdownTimer!)
        countdownTimer = null
      }
    }, 1000)
  }

  // 组件卸载时清理
  onUnmounted(() => {
    reset()
  })

  return {
    // 状态
    status: readonly(status),
    loading: readonly(loading),
    result: readonly(result),
    error: readonly(error),
    requestId: readonly(requestId),
    countdown: readonly(countdown),
    orderId: readonly(orderId),

    // 计算属性
    canParticipate,
    isProcessing,
    buttonText,
    buttonType,

    // 方法
    participate,
    cancel,
    reset,
    startCountdown,

    // 订单轮询相关
    orderPolling: {
      isPolling: orderPolling.isPolling,
      currentOrder: orderPolling.currentOrder,
      orderStatus: orderPolling.orderStatus,
      remainingTime: orderPolling.remainingTime,
      startPolling: orderPolling.startPolling,
      stopPolling: orderPolling.stopPolling,
      reset: orderPolling.reset
    }
  }
}
