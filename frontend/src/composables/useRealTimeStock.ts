import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useWebSocketMessage } from '@/composables/useWebSocket'
import { useActivityStore } from '@/stores/activity'
import type { StockInfo } from '@/types'

// 实时库存更新消息类型
interface StockUpdateMessage {
  activity_id: number
  available_stock: number
  total_stock: number
  status: 'normal' | 'low_stock' | 'out_of_stock'
  activity_status: 'pending' | 'active' | 'ended'
  last_updated: string
  change_amount?: number // 库存变化量
  user_count?: number // 当前在线用户数
}

// 批量库存更新消息
interface BatchStockUpdateMessage {
  updates: StockUpdateMessage[]
  timestamp: string
}

/**
 * 单个活动的实时库存更新
 */
export function useRealTimeStock(activityId: number) {
  const activityStore = useActivityStore()
  
  // WebSocket消息监听
  const { 
    data: stockUpdate, 
    sendMessage, 
    isConnected 
  } = useWebSocketMessage<StockUpdateMessage>('stock_update')
  
  // 状态
  const currentStock = ref<StockInfo | null>(null)
  const isSubscribed = ref(false)
  const lastUpdateTime = ref<Date | null>(null)
  const changeHistory = ref<Array<{ time: Date; change: number; stock: number }>>([])
  
  // 计算属性
  const stockInfo = computed(() => {
    return currentStock.value || activityStore.getActivityStockInfo(activityId)
  })
  
  const hasRecentChange = computed(() => {
    if (!lastUpdateTime.value) return false
    const now = new Date()
    const diff = now.getTime() - lastUpdateTime.value.getTime()
    return diff < 5000 // 5秒内有更新
  })
  
  const stockTrend = computed(() => {
    if (changeHistory.value.length < 2) return 'stable'
    const recent = changeHistory.value.slice(-3)
    const totalChange = recent.reduce((sum, item) => sum + item.change, 0)
    
    if (totalChange < -10) return 'decreasing_fast'
    if (totalChange < -3) return 'decreasing'
    if (totalChange > 3) return 'increasing'
    return 'stable'
  })
  
  // 订阅库存更新
  const subscribe = () => {
    if (!isConnected.value) {
      console.warn('WebSocket未连接，无法订阅库存更新')
      return
    }
    
    sendMessage({
      action: 'subscribe',
      activity_id: activityId
    })
    
    isSubscribed.value = true
    console.log(`已订阅活动 ${activityId} 的库存更新`)
  }
  
  // 取消订阅
  const unsubscribe = () => {
    if (!isConnected.value) return
    
    sendMessage({
      action: 'unsubscribe',
      activity_id: activityId
    })
    
    isSubscribed.value = false
    console.log(`已取消订阅活动 ${activityId} 的库存更新`)
  }
  
  // 请求当前库存
  const requestCurrentStock = () => {
    if (!isConnected.value) return
    
    sendMessage({
      action: 'get_stock',
      activity_id: activityId
    })
  }
  
  // 处理库存更新
  const handleStockUpdate = (update: StockUpdateMessage) => {
    if (update.activity_id !== activityId) return
    
    const oldStock = currentStock.value?.available_stock || 0
    const newStock = update.available_stock
    const change = newStock - oldStock
    
    // 更新当前库存
    currentStock.value = {
      activity_id: update.activity_id,
      available_stock: update.available_stock,
      total_stock: update.total_stock,
      status: update.status,
      activity_status: update.activity_status,
      last_updated: update.last_updated
    }
    
    // 更新store中的库存信息
    activityStore.stockInfos[activityId] = currentStock.value
    
    // 记录变化历史
    if (change !== 0) {
      changeHistory.value.push({
        time: new Date(),
        change,
        stock: newStock
      })
      
      // 只保留最近20条记录
      if (changeHistory.value.length > 20) {
        changeHistory.value = changeHistory.value.slice(-20)
      }
      
      // 显示库存变化提示
      if (change < 0) {
        const absChange = Math.abs(change)
        if (absChange >= 10) {
          ElMessage.info(`库存快速减少 ${absChange} 件，剩余 ${newStock} 件`)
        }
      }
      
      // 库存告急提示
      if (update.status === 'low_stock' && oldStock > newStock) {
        ElMessage.warning(`库存紧张！仅剩 ${newStock} 件`)
      } else if (update.status === 'out_of_stock') {
        ElMessage.error('商品已售罄！')
      }
    }
    
    lastUpdateTime.value = new Date()
  }
  
  // 监听WebSocket消息
  watch(stockUpdate, (update) => {
    if (update) {
      handleStockUpdate(update)
    }
  })
  
  // 监听连接状态
  watch(isConnected, (connected) => {
    if (connected && !isSubscribed.value) {
      // 连接建立后自动订阅
      setTimeout(() => {
        subscribe()
        requestCurrentStock()
      }, 1000)
    }
  })
  
  // 组件挂载时订阅
  onMounted(() => {
    if (isConnected.value) {
      subscribe()
      requestCurrentStock()
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
    stockInfo,
    isSubscribed: computed(() => isSubscribed.value),
    hasRecentChange,
    stockTrend,
    changeHistory: computed(() => changeHistory.value),
    lastUpdateTime: computed(() => lastUpdateTime.value),
    
    // 方法
    subscribe,
    unsubscribe,
    requestCurrentStock,
  }
}

/**
 * 批量活动的实时库存更新
 */
export function useRealTimeBatchStock(activityIds: number[]) {
  const activityStore = useActivityStore()
  
  // WebSocket消息监听
  const { 
    data: batchUpdate, 
    sendMessage, 
    isConnected 
  } = useWebSocketMessage<BatchStockUpdateMessage>('batch_stock_update')
  
  // 状态
  const subscribedIds = ref<Set<number>>(new Set())
  const stockUpdates = ref<Map<number, StockUpdateMessage>>(new Map())
  
  // 订阅批量库存更新
  const subscribeBatch = (ids: number[] = activityIds) => {
    if (!isConnected.value) return
    
    sendMessage({
      action: 'subscribe_batch',
      activity_ids: ids
    })
    
    ids.forEach(id => subscribedIds.value.add(id))
    console.log(`已订阅 ${ids.length} 个活动的库存更新`)
  }
  
  // 取消批量订阅
  const unsubscribeBatch = (ids: number[] = activityIds) => {
    if (!isConnected.value) return
    
    sendMessage({
      action: 'unsubscribe_batch',
      activity_ids: ids
    })
    
    ids.forEach(id => subscribedIds.value.delete(id))
    console.log(`已取消订阅 ${ids.length} 个活动的库存更新`)
  }
  
  // 处理批量库存更新
  const handleBatchUpdate = (update: BatchStockUpdateMessage) => {
    update.updates.forEach(stockUpdate => {
      // 更新本地状态
      stockUpdates.value.set(stockUpdate.activity_id, stockUpdate)
      
      // 更新store
      activityStore.stockInfos[stockUpdate.activity_id] = {
        activity_id: stockUpdate.activity_id,
        available_stock: stockUpdate.available_stock,
        total_stock: stockUpdate.total_stock,
        status: stockUpdate.status,
        activity_status: stockUpdate.activity_status,
        last_updated: stockUpdate.last_updated
      }
    })
  }
  
  // 获取特定活动的库存信息
  const getStockInfo = (activityId: number) => {
    return stockUpdates.value.get(activityId) || 
           activityStore.getActivityStockInfo(activityId)
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
    stockUpdates: computed(() => stockUpdates.value),
    
    // 方法
    subscribeBatch,
    unsubscribeBatch,
    getStockInfo,
  }
}
