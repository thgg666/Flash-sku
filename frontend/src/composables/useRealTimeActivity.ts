import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { useWebSocketMessage } from '@/composables/useWebSocket'
import { useActivityStore } from '@/stores/activity'
import { formatDateTime } from '@/utils'
import type { SeckillActivity } from '@/types'

// 活动状态更新消息类型
interface ActivityStatusMessage {
  activity_id: number
  status: 'pending' | 'active' | 'ended' | 'cancelled'
  start_time: string
  end_time: string
  remaining_time?: number // 剩余时间（秒）
  participant_count?: number // 参与人数
  message?: string // 状态变更消息
  timestamp: string
}

// 批量活动状态更新
interface BatchActivityStatusMessage {
  updates: ActivityStatusMessage[]
  timestamp: string
}

/**
 * 单个活动的实时状态更新
 */
export function useRealTimeActivity(activityId: number) {
  const activityStore = useActivityStore()
  
  // WebSocket消息监听
  const { 
    data: statusUpdate, 
    sendMessage, 
    isConnected 
  } = useWebSocketMessage<ActivityStatusMessage>('activity_status')
  
  // 状态
  const isSubscribed = ref(false)
  const lastStatusChange = ref<Date | null>(null)
  const statusHistory = ref<Array<{ time: Date; status: string; message?: string }>>([])
  
  // 计算属性
  const currentActivity = computed(() => {
    return activityStore.currentActivity || 
           activityStore.activities.find(a => a.id === activityId)
  })
  
  const hasRecentStatusChange = computed(() => {
    if (!lastStatusChange.value) return false
    const now = new Date()
    const diff = now.getTime() - lastStatusChange.value.getTime()
    return diff < 10000 // 10秒内有状态变更
  })
  
  // 订阅活动状态更新
  const subscribe = () => {
    if (!isConnected.value) {
      console.warn('WebSocket未连接，无法订阅活动状态更新')
      return
    }
    
    sendMessage({
      action: 'subscribe_activity',
      activity_id: activityId
    })
    
    isSubscribed.value = true
    console.log(`已订阅活动 ${activityId} 的状态更新`)
  }
  
  // 取消订阅
  const unsubscribe = () => {
    if (!isConnected.value) return
    
    sendMessage({
      action: 'unsubscribe_activity',
      activity_id: activityId
    })
    
    isSubscribed.value = false
    console.log(`已取消订阅活动 ${activityId} 的状态更新`)
  }
  
  // 处理活动状态更新
  const handleStatusUpdate = (update: ActivityStatusMessage) => {
    if (update.activity_id !== activityId) return
    
    const oldStatus = currentActivity.value?.status
    const newStatus = update.status
    
    // 更新活动信息
    if (currentActivity.value) {
      currentActivity.value.status = update.status
      currentActivity.value.start_time = update.start_time
      currentActivity.value.end_time = update.end_time
    }
    
    // 更新store中的活动信息
    const activityIndex = activityStore.activities.findIndex(a => a.id === activityId)
    if (activityIndex !== -1) {
      activityStore.activities[activityIndex].status = update.status
      activityStore.activities[activityIndex].start_time = update.start_time
      activityStore.activities[activityIndex].end_time = update.end_time
    }
    
    // 记录状态变更历史
    if (oldStatus !== newStatus) {
      statusHistory.value.push({
        time: new Date(),
        status: newStatus,
        message: update.message
      })
      
      // 只保留最近10条记录
      if (statusHistory.value.length > 10) {
        statusHistory.value = statusHistory.value.slice(-10)
      }
      
      // 显示状态变更通知
      showStatusChangeNotification(oldStatus, newStatus, update)
    }
    
    lastStatusChange.value = new Date()
  }
  
  // 显示状态变更通知
  const showStatusChangeNotification = (
    oldStatus: string | undefined, 
    newStatus: string, 
    update: ActivityStatusMessage
  ) => {
    const activityName = currentActivity.value?.name || `活动 ${activityId}`
    
    switch (newStatus) {
      case 'active':
        if (oldStatus === 'pending') {
          ElNotification({
            title: '活动开始',
            message: `${activityName} 已开始，快来抢购吧！`,
            type: 'success',
            duration: 8000,
            onClick: () => {
              // 可以跳转到活动详情页
              window.location.href = `/activity/${activityId}`
            }
          })
        }
        break
        
      case 'ended':
        if (oldStatus === 'active') {
          ElNotification({
            title: '活动结束',
            message: `${activityName} 已结束，感谢您的参与！`,
            type: 'warning',
            duration: 6000
          })
        }
        break
        
      case 'cancelled':
        ElNotification({
          title: '活动取消',
          message: `${activityName} 已被取消${update.message ? '：' + update.message : ''}`,
          type: 'error',
          duration: 10000
        })
        break
        
      default:
        if (update.message) {
          ElMessage.info(update.message)
        }
    }
  }
  
  // 监听WebSocket消息
  watch(statusUpdate, (update) => {
    if (update) {
      handleStatusUpdate(update)
    }
  })
  
  // 监听连接状态
  watch(isConnected, (connected) => {
    if (connected && !isSubscribed.value) {
      setTimeout(() => {
        subscribe()
      }, 1000)
    }
  })
  
  // 组件挂载时订阅
  onMounted(() => {
    if (isConnected.value) {
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
    currentActivity,
    isSubscribed: computed(() => isSubscribed.value),
    hasRecentStatusChange,
    statusHistory: computed(() => statusHistory.value),
    lastStatusChange: computed(() => lastStatusChange.value),
    
    // 方法
    subscribe,
    unsubscribe,
  }
}

/**
 * 批量活动的实时状态更新
 */
export function useRealTimeBatchActivity(activityIds: number[]) {
  const activityStore = useActivityStore()
  
  // WebSocket消息监听
  const { 
    data: batchUpdate, 
    sendMessage, 
    isConnected 
  } = useWebSocketMessage<BatchActivityStatusMessage>('batch_activity_status')
  
  // 状态
  const subscribedIds = ref<Set<number>>(new Set())
  const statusUpdates = ref<Map<number, ActivityStatusMessage>>(new Map())
  
  // 订阅批量活动状态更新
  const subscribeBatch = (ids: number[] = activityIds) => {
    if (!isConnected.value) return
    
    sendMessage({
      action: 'subscribe_batch_activity',
      activity_ids: ids
    })
    
    ids.forEach(id => subscribedIds.value.add(id))
    console.log(`已订阅 ${ids.length} 个活动的状态更新`)
  }
  
  // 取消批量订阅
  const unsubscribeBatch = (ids: number[] = activityIds) => {
    if (!isConnected.value) return
    
    sendMessage({
      action: 'unsubscribe_batch_activity',
      activity_ids: ids
    })
    
    ids.forEach(id => subscribedIds.value.delete(id))
    console.log(`已取消订阅 ${ids.length} 个活动的状态更新`)
  }
  
  // 处理批量状态更新
  const handleBatchUpdate = (update: BatchActivityStatusMessage) => {
    update.updates.forEach(statusUpdate => {
      // 更新本地状态
      statusUpdates.value.set(statusUpdate.activity_id, statusUpdate)
      
      // 更新store中的活动
      const activityIndex = activityStore.activities.findIndex(
        a => a.id === statusUpdate.activity_id
      )
      if (activityIndex !== -1) {
        activityStore.activities[activityIndex].status = statusUpdate.status
        activityStore.activities[activityIndex].start_time = statusUpdate.start_time
        activityStore.activities[activityIndex].end_time = statusUpdate.end_time
      }
    })
  }
  
  // 获取特定活动的状态信息
  const getActivityStatus = (activityId: number) => {
    return statusUpdates.value.get(activityId)
  }
  
  // 监听批量更新
  watch(batchUpdate, (update) => {
    if (update) {
      handleBatchUpdate(update)
    }
  })
  
  // 监听连接状态
  watch(isConnected, (connected) => {
    if (connected && subscribedIds.value.size === 0) {
      setTimeout(() => {
        subscribeBatch()
      }, 1000)
    }
  })
  
  // 组件挂载时订阅
  onMounted(() => {
    if (isConnected.value) {
      subscribeBatch()
    }
  })
  
  // 组件卸载时取消订阅
  onUnmounted(() => {
    if (subscribedIds.value.size > 0) {
      unsubscribeBatch(Array.from(subscribedIds.value))
    }
  })
  
  return {
    // 状态
    subscribedIds: computed(() => Array.from(subscribedIds.value)),
    statusUpdates: computed(() => statusUpdates.value),
    
    // 方法
    subscribeBatch,
    unsubscribeBatch,
    getActivityStatus,
  }
}
