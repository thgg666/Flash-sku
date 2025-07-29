/**
 * 离线数据同步 Composable
 * 管理离线数据的存储、同步和状态
 */

import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { offlineStorage, offlineSync, type OfflineOrder } from '@/utils/offlineStorage'

export function useOfflineSync() {
  // 响应式状态
  const isOnline = ref(navigator.onLine)
  const isSyncing = ref(false)
  const pendingOrders = ref<OfflineOrder[]>([])
  const pendingSyncCount = ref(0)
  const lastSyncTime = ref<number | null>(null)
  const syncError = ref<string | null>(null)

  // 计算属性
  const hasPendingSync = computed(() => pendingSyncCount.value > 0)
  const canSync = computed(() => isOnline.value && !isSyncing.value)

  /**
   * 创建离线订单
   */
  const createOfflineOrder = async (orderData: {
    activityId: string
    productId: string
    quantity: number
    totalPrice: number
  }) => {
    const order: OfflineOrder = {
      id: `offline_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      activityId: orderData.activityId,
      productId: orderData.productId,
      userId: getCurrentUserId(),
      quantity: orderData.quantity,
      totalPrice: orderData.totalPrice,
      status: 'pending',
      createdAt: Date.now(),
      synced: false
    }

    try {
      await offlineStorage.storeOrder(order)
      await updatePendingCounts()
      
      ElMessage.info('订单已保存到离线队列，将在网络恢复时自动同步')
      
      // 如果在线，立即尝试同步
      if (isOnline.value) {
        await syncPendingData()
      }
      
      return order
    } catch (error) {
      console.error('Failed to create offline order:', error)
      throw new Error('保存离线订单失败')
    }
  }

  /**
   * 同步待处理数据
   */
  const syncPendingData = async () => {
    if (!canSync.value) {
      return
    }

    isSyncing.value = true
    syncError.value = null

    try {
      await offlineSync.startSync()
      await updatePendingCounts()
      
      lastSyncTime.value = Date.now()
      
      if (pendingSyncCount.value === 0) {
        ElMessage.success('所有数据已同步完成')
      }
    } catch (error) {
      console.error('Sync failed:', error)
      syncError.value = error instanceof Error ? error.message : '同步失败'
      ElMessage.error('数据同步失败，请稍后重试')
    } finally {
      isSyncing.value = false
    }
  }

  /**
   * 更新待同步数量
   */
  const updatePendingCounts = async () => {
    try {
      const [orders, syncQueue] = await Promise.all([
        offlineStorage.getUnsyncedOrders(),
        offlineStorage.getSyncQueue()
      ])
      
      pendingOrders.value = orders
      pendingSyncCount.value = orders.length + syncQueue.length
    } catch (error) {
      console.error('Failed to update pending counts:', error)
    }
  }

  /**
   * 获取用户订单历史（包括离线订单）
   */
  const getUserOrderHistory = async (userId?: string) => {
    const currentUserId = userId || getCurrentUserId()
    
    try {
      const offlineOrders = await offlineStorage.getUserOrders(currentUserId)
      
      // 这里可以与在线API结合，获取完整的订单历史
      // const onlineOrders = await fetchOnlineOrders(currentUserId)
      
      return offlineOrders.sort((a, b) => b.createdAt - a.createdAt)
    } catch (error) {
      console.error('Failed to get user order history:', error)
      return []
    }
  }

  /**
   * 缓存产品数据
   */
  const cacheProductData = async (products: any[]) => {
    try {
      const offlineProducts = products.map(product => ({
        id: product.id.toString(),
        name: product.name,
        price: product.price,
        image: product.image,
        description: product.description,
        category: product.category?.name || '',
        stock: product.stock,
        updatedAt: Date.now()
      }))
      
      await offlineStorage.storeProducts(offlineProducts)
    } catch (error) {
      console.error('Failed to cache product data:', error)
    }
  }

  /**
   * 缓存活动数据
   */
  const cacheActivityData = async (activities: any[]) => {
    try {
      const offlineActivities = activities.map(activity => ({
        id: activity.id.toString(),
        productId: activity.product_id.toString(),
        title: activity.title,
        startTime: new Date(activity.start_time).getTime(),
        endTime: new Date(activity.end_time).getTime(),
        originalPrice: activity.original_price,
        seckillPrice: activity.seckill_price,
        stock: activity.stock,
        maxPerUser: activity.max_per_user,
        status: activity.status,
        updatedAt: Date.now()
      }))
      
      await offlineStorage.storeActivities(offlineActivities)
    } catch (error) {
      console.error('Failed to cache activity data:', error)
    }
  }

  /**
   * 获取缓存的产品数据
   */
  const getCachedProducts = async (category?: string) => {
    try {
      return await offlineStorage.getProducts(category)
    } catch (error) {
      console.error('Failed to get cached products:', error)
      return []
    }
  }

  /**
   * 获取缓存的活动数据
   */
  const getCachedActivities = async (status?: string) => {
    try {
      return await offlineStorage.getActivities(status)
    } catch (error) {
      console.error('Failed to get cached activities:', error)
      return []
    }
  }

  /**
   * 添加用户操作到同步队列
   */
  const addUserActionToQueue = async (action: {
    type: string
    url: string
    method: string
    data: any
    headers?: Record<string, string>
  }) => {
    try {
      await offlineSync.addOfflineAction(
        action.type,
        action.url,
        action.method,
        action.data,
        action.headers
      )
      
      await updatePendingCounts()
      
      // 如果在线，立即尝试同步
      if (isOnline.value) {
        await syncPendingData()
      }
    } catch (error) {
      console.error('Failed to add user action to queue:', error)
      throw new Error('添加操作到同步队列失败')
    }
  }

  /**
   * 清理过期缓存
   */
  const cleanupExpiredCache = async () => {
    try {
      await offlineStorage.cleanupExpiredCache()
    } catch (error) {
      console.error('Failed to cleanup expired cache:', error)
    }
  }

  /**
   * 获取存储使用情况
   */
  const getStorageUsage = async () => {
    try {
      return await offlineStorage.getStorageUsage()
    } catch (error) {
      console.error('Failed to get storage usage:', error)
      return { used: 0, quota: 0 }
    }
  }

  /**
   * 获取当前用户ID
   */
  const getCurrentUserId = (): string => {
    // 这里应该从认证状态中获取用户ID
    const userStr = localStorage.getItem('user')
    if (userStr) {
      try {
        const user = JSON.parse(userStr)
        return user.id?.toString() || 'anonymous'
      } catch {
        return 'anonymous'
      }
    }
    return 'anonymous'
  }

  /**
   * 网络状态变化处理
   */
  const handleOnlineStatusChange = () => {
    isOnline.value = navigator.onLine
    
    if (isOnline.value && hasPendingSync.value) {
      // 网络恢复且有待同步数据时，延迟同步
      setTimeout(() => {
        syncPendingData()
      }, 1000)
    }
  }

  // 生命周期管理
  onMounted(() => {
    // 监听网络状态变化
    window.addEventListener('online', handleOnlineStatusChange)
    window.addEventListener('offline', handleOnlineStatusChange)
    
    // 初始化数据
    updatePendingCounts()
    
    // 定期清理过期缓存
    const cleanupInterval = setInterval(cleanupExpiredCache, 60 * 60 * 1000) // 每小时清理一次
    
    onUnmounted(() => {
      clearInterval(cleanupInterval)
    })
  })

  onUnmounted(() => {
    window.removeEventListener('online', handleOnlineStatusChange)
    window.removeEventListener('offline', handleOnlineStatusChange)
  })

  return {
    // 状态
    isOnline,
    isSyncing,
    pendingOrders,
    pendingSyncCount,
    lastSyncTime,
    syncError,
    
    // 计算属性
    hasPendingSync,
    canSync,
    
    // 方法
    createOfflineOrder,
    syncPendingData,
    updatePendingCounts,
    getUserOrderHistory,
    cacheProductData,
    cacheActivityData,
    getCachedProducts,
    getCachedActivities,
    addUserActionToQueue,
    cleanupExpiredCache,
    getStorageUsage
  }
}
