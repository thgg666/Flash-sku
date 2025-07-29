import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { useWebSocketMessage } from '@/composables/useWebSocket'
import { useAuthStore } from '@/stores/auth'
import type { SeckillResult } from '@/types'

// 实时用户反馈消息类型
interface UserFeedbackMessage {
  type: 'seckill_result' | 'queue_update' | 'system_message' | 'activity_reminder'
  user_id: number
  data: any
  timestamp: string
}

// 秒杀结果反馈
interface SeckillResultFeedback {
  activity_id: number
  result: SeckillResult
  queue_position?: number
  estimated_wait_time?: number
}

// 排队状态更新
interface QueueUpdateFeedback {
  activity_id: number
  queue_position: number
  estimated_wait_time: number
  total_queue_length: number
  progress_percentage: number
}

// 系统消息
interface SystemMessageFeedback {
  message: string
  level: 'info' | 'warning' | 'error' | 'success'
  action?: {
    text: string
    url: string
  }
}

// 活动提醒
interface ActivityReminderFeedback {
  activity_id: number
  activity_name: string
  reminder_type: 'starting_soon' | 'ending_soon' | 'low_stock'
  message: string
  time_remaining?: number
}

/**
 * 实时用户反馈
 */
export function useRealTimeFeedback() {
  const authStore = useAuthStore()
  
  // WebSocket消息监听
  const { 
    data: feedbackMessage, 
    sendMessage, 
    isConnected 
  } = useWebSocketMessage<UserFeedbackMessage>('user_feedback')
  
  // 状态
  const isSubscribed = ref(false)
  const feedbackHistory = ref<UserFeedbackMessage[]>([])
  const unreadCount = ref(0)
  
  // 计算属性
  const hasUnreadFeedback = computed(() => unreadCount.value > 0)
  
  const recentFeedback = computed(() => {
    return feedbackHistory.value.slice(-10).reverse()
  })
  
  // 订阅用户反馈
  const subscribe = () => {
    if (!isConnected.value || !authStore.isAuthenticated) {
      console.warn('WebSocket未连接或用户未登录，无法订阅用户反馈')
      return
    }
    
    sendMessage({
      action: 'subscribe_user_feedback',
      user_id: authStore.user?.id
    })
    
    isSubscribed.value = true
    console.log('已订阅用户反馈')
  }
  
  // 取消订阅
  const unsubscribe = () => {
    if (!isConnected.value) return
    
    sendMessage({
      action: 'unsubscribe_user_feedback',
      user_id: authStore.user?.id
    })
    
    isSubscribed.value = false
    console.log('已取消订阅用户反馈')
  }
  
  // 处理反馈消息
  const handleFeedbackMessage = (message: UserFeedbackMessage) => {
    // 只处理当前用户的消息
    if (message.user_id !== authStore.user?.id) return
    
    // 添加到历史记录
    feedbackHistory.value.push(message)
    
    // 只保留最近50条记录
    if (feedbackHistory.value.length > 50) {
      feedbackHistory.value = feedbackHistory.value.slice(-50)
    }
    
    // 增加未读计数
    unreadCount.value++
    
    // 根据消息类型处理
    switch (message.type) {
      case 'seckill_result':
        handleSeckillResult(message.data as SeckillResultFeedback)
        break
      case 'queue_update':
        handleQueueUpdate(message.data as QueueUpdateFeedback)
        break
      case 'system_message':
        handleSystemMessage(message.data as SystemMessageFeedback)
        break
      case 'activity_reminder':
        handleActivityReminder(message.data as ActivityReminderFeedback)
        break
    }
  }
  
  // 处理秒杀结果
  const handleSeckillResult = (data: SeckillResultFeedback) => {
    const result = data.result
    
    if (result.code === 'SUCCESS') {
      ElNotification({
        title: '抢购成功！',
        message: `恭喜您成功抢购商品，订单号：${result.order_id}`,
        type: 'success',
        duration: 10000,
        onClick: () => {
          if (result.order_id) {
            window.location.href = `/user/orders/${result.order_id}`
          }
        }
      })
    } else {
      let message = '抢购失败'
      let type: 'warning' | 'error' = 'warning'
      
      switch (result.code) {
        case 'SOLD_OUT':
          message = '商品已售罄，下次要快一点哦'
          break
        case 'LIMIT_EXCEEDED':
          message = '超出限购数量'
          break
        case 'RATE_LIMIT':
          message = '请求过于频繁，请稍后再试'
          type = 'error'
          break
        default:
          message = result.message || '抢购失败，请重试'
      }
      
      ElNotification({
        title: '抢购失败',
        message,
        type,
        duration: 6000
      })
    }
  }
  
  // 处理排队状态更新
  const handleQueueUpdate = (data: QueueUpdateFeedback) => {
    const { queue_position, estimated_wait_time, progress_percentage } = data
    
    if (queue_position <= 10) {
      ElMessage.info(`排队中，您前面还有 ${queue_position} 人，即将轮到您！`)
    } else if (progress_percentage >= 50) {
      ElMessage.info(`排队进度 ${progress_percentage}%，预计等待 ${estimated_wait_time} 秒`)
    }
  }
  
  // 处理系统消息
  const handleSystemMessage = (data: SystemMessageFeedback) => {
    const { message, level, action } = data
    
    let type: 'success' | 'warning' | 'info' | 'error' = 'info'
    switch (level) {
      case 'success':
        type = 'success'
        break
      case 'warning':
        type = 'warning'
        break
      case 'error':
        type = 'error'
        break
    }
    
    ElNotification({
      title: '系统消息',
      message,
      type,
      duration: level === 'error' ? 0 : 8000,
      onClick: action ? () => {
        window.location.href = action.url
      } : undefined
    })
  }
  
  // 处理活动提醒
  const handleActivityReminder = (data: ActivityReminderFeedback) => {
    const { activity_name, reminder_type, message, time_remaining } = data
    
    let title = '活动提醒'
    let type: 'info' | 'warning' = 'info'
    
    switch (reminder_type) {
      case 'starting_soon':
        title = '活动即将开始'
        type = 'warning'
        break
      case 'ending_soon':
        title = '活动即将结束'
        type = 'warning'
        break
      case 'low_stock':
        title = '库存告急'
        type = 'warning'
        break
    }
    
    ElNotification({
      title,
      message: `${activity_name}: ${message}`,
      type,
      duration: 10000,
      onClick: () => {
        window.location.href = `/activity/${data.activity_id}`
      }
    })
  }
  
  // 标记消息为已读
  const markAsRead = (messageId?: string) => {
    if (messageId) {
      // 标记特定消息为已读
      // 这里可以发送已读确认到服务器
    } else {
      // 标记所有消息为已读
      unreadCount.value = 0
    }
  }
  
  // 清除历史记录
  const clearHistory = () => {
    feedbackHistory.value = []
    unreadCount.value = 0
  }
  
  // 监听WebSocket消息
  watch(feedbackMessage, (message) => {
    if (message) {
      handleFeedbackMessage(message)
    }
  })
  
  // 监听连接状态
  watch(isConnected, (connected) => {
    if (connected && !isSubscribed.value && authStore.isAuthenticated) {
      setTimeout(() => {
        subscribe()
      }, 1000)
    }
  })
  
  // 监听认证状态
  watch(() => authStore.isAuthenticated, (authenticated) => {
    if (authenticated && isConnected.value && !isSubscribed.value) {
      subscribe()
    } else if (!authenticated && isSubscribed.value) {
      unsubscribe()
    }
  })
  
  // 组件挂载时订阅
  onMounted(() => {
    if (isConnected.value && authStore.isAuthenticated) {
      subscribe()
    }
  })
  
  // 组件卸载时取消订阅
  onUnmounted(() => {
    if (isSubscribed.value) {
      unsubscribe()
    }
  })
  
  return {
    // 状态
    isSubscribed: computed(() => isSubscribed.value),
    feedbackHistory: computed(() => feedbackHistory.value),
    recentFeedback,
    unreadCount: computed(() => unreadCount.value),
    hasUnreadFeedback,
    
    // 方法
    subscribe,
    unsubscribe,
    markAsRead,
    clearHistory,
  }
}
