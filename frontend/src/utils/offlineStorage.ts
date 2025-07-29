/**
 * 离线存储管理器
 * 使用 IndexedDB 存储离线数据，支持数据同步
 */

// 数据库配置
const DB_NAME = 'FlashSkuOfflineDB'
const DB_VERSION = 1

// 对象存储名称
const STORES = {
  PRODUCTS: 'products',
  ACTIVITIES: 'activities',
  ORDERS: 'orders',
  USER_DATA: 'userData',
  SYNC_QUEUE: 'syncQueue',
  CACHE_META: 'cacheMeta'
} as const

// 数据类型定义
export interface OfflineProduct {
  id: string
  name: string
  price: number
  image: string
  description: string
  category: string
  stock: number
  updatedAt: number
}

export interface OfflineActivity {
  id: string
  productId: string
  title: string
  startTime: number
  endTime: number
  originalPrice: number
  seckillPrice: number
  stock: number
  maxPerUser: number
  status: 'pending' | 'active' | 'ended'
  updatedAt: number
}

export interface OfflineOrder {
  id: string
  activityId: string
  productId: string
  userId: string
  quantity: number
  totalPrice: number
  status: 'pending' | 'success' | 'failed'
  createdAt: number
  synced: boolean
}

export interface SyncQueueItem {
  id: string
  type: 'order' | 'user_action' | 'preference'
  data: any
  url: string
  method: string
  headers?: Record<string, string>
  createdAt: number
  retryCount: number
  maxRetries: number
}

export interface CacheMeta {
  key: string
  lastUpdated: number
  expiresAt: number
  size: number
}

class OfflineStorageManager {
  private db: IDBDatabase | null = null
  private initPromise: Promise<void> | null = null

  constructor() {
    this.initPromise = this.init()
  }

  /**
   * 初始化数据库
   */
  private async init(): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION)

      request.onerror = () => {
        reject(new Error('Failed to open IndexedDB'))
      }

      request.onsuccess = () => {
        this.db = request.result
        resolve()
      }

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result

        // 创建产品存储
        if (!db.objectStoreNames.contains(STORES.PRODUCTS)) {
          const productStore = db.createObjectStore(STORES.PRODUCTS, { keyPath: 'id' })
          productStore.createIndex('category', 'category', { unique: false })
          productStore.createIndex('updatedAt', 'updatedAt', { unique: false })
        }

        // 创建活动存储
        if (!db.objectStoreNames.contains(STORES.ACTIVITIES)) {
          const activityStore = db.createObjectStore(STORES.ACTIVITIES, { keyPath: 'id' })
          activityStore.createIndex('productId', 'productId', { unique: false })
          activityStore.createIndex('status', 'status', { unique: false })
          activityStore.createIndex('startTime', 'startTime', { unique: false })
        }

        // 创建订单存储
        if (!db.objectStoreNames.contains(STORES.ORDERS)) {
          const orderStore = db.createObjectStore(STORES.ORDERS, { keyPath: 'id' })
          orderStore.createIndex('userId', 'userId', { unique: false })
          orderStore.createIndex('synced', 'synced', { unique: false })
          orderStore.createIndex('createdAt', 'createdAt', { unique: false })
        }

        // 创建用户数据存储
        if (!db.objectStoreNames.contains(STORES.USER_DATA)) {
          db.createObjectStore(STORES.USER_DATA, { keyPath: 'key' })
        }

        // 创建同步队列存储
        if (!db.objectStoreNames.contains(STORES.SYNC_QUEUE)) {
          const syncStore = db.createObjectStore(STORES.SYNC_QUEUE, { keyPath: 'id' })
          syncStore.createIndex('type', 'type', { unique: false })
          syncStore.createIndex('createdAt', 'createdAt', { unique: false })
        }

        // 创建缓存元数据存储
        if (!db.objectStoreNames.contains(STORES.CACHE_META)) {
          const metaStore = db.createObjectStore(STORES.CACHE_META, { keyPath: 'key' })
          metaStore.createIndex('expiresAt', 'expiresAt', { unique: false })
        }
      }
    })
  }

  /**
   * 确保数据库已初始化
   */
  private async ensureInit(): Promise<void> {
    if (this.initPromise) {
      await this.initPromise
    }
    if (!this.db) {
      throw new Error('Database not initialized')
    }
  }

  /**
   * 执行事务
   */
  private async transaction<T>(
    storeNames: string | string[],
    mode: IDBTransactionMode,
    callback: (stores: IDBObjectStore | IDBObjectStore[]) => Promise<T>
  ): Promise<T> {
    await this.ensureInit()
    
    const transaction = this.db!.transaction(storeNames, mode)
    const stores = Array.isArray(storeNames)
      ? storeNames.map(name => transaction.objectStore(name))
      : transaction.objectStore(storeNames)

    return callback(stores)
  }

  /**
   * 存储产品数据
   */
  async storeProducts(products: OfflineProduct[]): Promise<void> {
    await this.transaction(STORES.PRODUCTS, 'readwrite', async (store) => {
      const promises = products.map(product => {
        const request = (store as IDBObjectStore).put({
          ...product,
          updatedAt: Date.now()
        })
        return new Promise<void>((resolve, reject) => {
          request.onsuccess = () => resolve()
          request.onerror = () => reject(request.error)
        })
      })
      await Promise.all(promises)
    })
  }

  /**
   * 获取产品数据
   */
  async getProducts(category?: string): Promise<OfflineProduct[]> {
    return this.transaction(STORES.PRODUCTS, 'readonly', async (store) => {
      return new Promise<OfflineProduct[]>((resolve, reject) => {
        let request: IDBRequest

        if (category) {
          const index = (store as IDBObjectStore).index('category')
          request = index.getAll(category)
        } else {
          request = (store as IDBObjectStore).getAll()
        }

        request.onsuccess = () => resolve(request.result)
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 存储活动数据
   */
  async storeActivities(activities: OfflineActivity[]): Promise<void> {
    await this.transaction(STORES.ACTIVITIES, 'readwrite', async (store) => {
      const promises = activities.map(activity => {
        const request = (store as IDBObjectStore).put({
          ...activity,
          updatedAt: Date.now()
        })
        return new Promise<void>((resolve, reject) => {
          request.onsuccess = () => resolve()
          request.onerror = () => reject(request.error)
        })
      })
      await Promise.all(promises)
    })
  }

  /**
   * 获取活动数据
   */
  async getActivities(status?: string): Promise<OfflineActivity[]> {
    return this.transaction(STORES.ACTIVITIES, 'readonly', async (store) => {
      return new Promise<OfflineActivity[]>((resolve, reject) => {
        let request: IDBRequest

        if (status) {
          const index = (store as IDBObjectStore).index('status')
          request = index.getAll(status)
        } else {
          request = (store as IDBObjectStore).getAll()
        }

        request.onsuccess = () => resolve(request.result)
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 存储订单数据
   */
  async storeOrder(order: OfflineOrder): Promise<void> {
    await this.transaction(STORES.ORDERS, 'readwrite', async (store) => {
      return new Promise<void>((resolve, reject) => {
        const request = (store as IDBObjectStore).put(order)
        request.onsuccess = () => resolve()
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 获取用户订单
   */
  async getUserOrders(userId: string): Promise<OfflineOrder[]> {
    return this.transaction(STORES.ORDERS, 'readonly', async (store) => {
      return new Promise<OfflineOrder[]>((resolve, reject) => {
        const index = (store as IDBObjectStore).index('userId')
        const request = index.getAll(userId)
        request.onsuccess = () => resolve(request.result)
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 获取未同步的订单
   */
  async getUnsyncedOrders(): Promise<OfflineOrder[]> {
    return this.transaction(STORES.ORDERS, 'readonly', async (store) => {
      return new Promise<OfflineOrder[]>((resolve, reject) => {
        const index = (store as IDBObjectStore).index('synced')
        const request = index.getAll(IDBKeyRange.only(false))
        request.onsuccess = () => resolve(request.result)
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 标记订单为已同步
   */
  async markOrderSynced(orderId: string): Promise<void> {
    await this.transaction(STORES.ORDERS, 'readwrite', async (store) => {
      return new Promise<void>((resolve, reject) => {
        const getRequest = (store as IDBObjectStore).get(orderId)
        getRequest.onsuccess = () => {
          const order = getRequest.result
          if (order) {
            order.synced = true
            const putRequest = (store as IDBObjectStore).put(order)
            putRequest.onsuccess = () => resolve()
            putRequest.onerror = () => reject(putRequest.error)
          } else {
            reject(new Error('Order not found'))
          }
        }
        getRequest.onerror = () => reject(getRequest.error)
      })
    })
  }

  /**
   * 添加到同步队列
   */
  async addToSyncQueue(item: Omit<SyncQueueItem, 'id' | 'createdAt' | 'retryCount'>): Promise<void> {
    const syncItem: SyncQueueItem = {
      ...item,
      id: `sync_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`,
      createdAt: Date.now(),
      retryCount: 0
    }

    await this.transaction(STORES.SYNC_QUEUE, 'readwrite', async (store) => {
      return new Promise<void>((resolve, reject) => {
        const request = (store as IDBObjectStore).put(syncItem)
        request.onsuccess = () => resolve()
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 获取同步队列
   */
  async getSyncQueue(): Promise<SyncQueueItem[]> {
    return this.transaction(STORES.SYNC_QUEUE, 'readonly', async (store) => {
      return new Promise<SyncQueueItem[]>((resolve, reject) => {
        const request = (store as IDBObjectStore).getAll()
        request.onsuccess = () => resolve(request.result)
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 移除同步队列项
   */
  async removeSyncQueueItem(id: string): Promise<void> {
    await this.transaction(STORES.SYNC_QUEUE, 'readwrite', async (store) => {
      return new Promise<void>((resolve, reject) => {
        const request = (store as IDBObjectStore).delete(id)
        request.onsuccess = () => resolve()
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 更新同步队列项重试次数
   */
  async updateSyncQueueItemRetry(id: string): Promise<void> {
    await this.transaction(STORES.SYNC_QUEUE, 'readwrite', async (store) => {
      return new Promise<void>((resolve, reject) => {
        const getRequest = (store as IDBObjectStore).get(id)
        getRequest.onsuccess = () => {
          const item = getRequest.result
          if (item) {
            item.retryCount += 1
            const putRequest = (store as IDBObjectStore).put(item)
            putRequest.onsuccess = () => resolve()
            putRequest.onerror = () => reject(putRequest.error)
          } else {
            reject(new Error('Sync queue item not found'))
          }
        }
        getRequest.onerror = () => reject(getRequest.error)
      })
    })
  }

  /**
   * 存储用户数据
   */
  async storeUserData(key: string, data: any): Promise<void> {
    await this.transaction(STORES.USER_DATA, 'readwrite', async (store) => {
      return new Promise<void>((resolve, reject) => {
        const request = (store as IDBObjectStore).put({
          key,
          data,
          updatedAt: Date.now()
        })
        request.onsuccess = () => resolve()
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 获取用户数据
   */
  async getUserData(key: string): Promise<any> {
    return this.transaction(STORES.USER_DATA, 'readonly', async (store) => {
      return new Promise<any>((resolve, reject) => {
        const request = (store as IDBObjectStore).get(key)
        request.onsuccess = () => {
          const result = request.result
          resolve(result ? result.data : null)
        }
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 清理过期缓存
   */
  async cleanupExpiredCache(): Promise<void> {
    const now = Date.now()
    
    await this.transaction(STORES.CACHE_META, 'readwrite', async (store) => {
      return new Promise<void>((resolve, reject) => {
        const index = (store as IDBObjectStore).index('expiresAt')
        const range = IDBKeyRange.upperBound(now)
        const request = index.openCursor(range)
        
        request.onsuccess = () => {
          const cursor = request.result
          if (cursor) {
            cursor.delete()
            cursor.continue()
          } else {
            resolve()
          }
        }
        
        request.onerror = () => reject(request.error)
      })
    })
  }

  /**
   * 获取存储使用情况
   */
  async getStorageUsage(): Promise<{ used: number; quota: number }> {
    if ('storage' in navigator && 'estimate' in navigator.storage) {
      const estimate = await navigator.storage.estimate()
      return {
        used: estimate.usage || 0,
        quota: estimate.quota || 0
      }
    }
    
    return { used: 0, quota: 0 }
  }

  /**
   * 清空所有数据
   */
  async clearAllData(): Promise<void> {
    await this.ensureInit()
    
    const storeNames = Object.values(STORES)
    await this.transaction(storeNames, 'readwrite', async (stores) => {
      const promises = (stores as IDBObjectStore[]).map(store => {
        return new Promise<void>((resolve, reject) => {
          const request = store.clear()
          request.onsuccess = () => resolve()
          request.onerror = () => reject(request.error)
        })
      })
      await Promise.all(promises)
    })
  }
}

/**
 * 离线同步管理器
 */
class OfflineSyncManager {
  private storage: OfflineStorageManager
  private syncInProgress = false
  private syncInterval: number | null = null

  constructor(storage: OfflineStorageManager) {
    this.storage = storage
    this.init()
  }

  private init() {
    // 监听网络状态变化
    window.addEventListener('online', () => {
      this.startSync()
    })

    // 定期同步（当在线时）
    this.startPeriodicSync()
  }

  /**
   * 开始同步
   */
  async startSync(): Promise<void> {
    if (this.syncInProgress || !navigator.onLine) {
      return
    }

    this.syncInProgress = true

    try {
      // 同步订单
      await this.syncOrders()

      // 同步队列中的其他数据
      await this.syncQueueItems()

      console.log('Offline sync completed successfully')
    } catch (error) {
      console.error('Offline sync failed:', error)
    } finally {
      this.syncInProgress = false
    }
  }

  /**
   * 同步订单
   */
  private async syncOrders(): Promise<void> {
    const unsyncedOrders = await this.storage.getUnsyncedOrders()

    for (const order of unsyncedOrders) {
      try {
        const response = await fetch('/api/orders/', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          },
          body: JSON.stringify({
            activity_id: order.activityId,
            quantity: order.quantity
          })
        })

        if (response.ok) {
          await this.storage.markOrderSynced(order.id)
          console.log('Order synced successfully:', order.id)
        } else {
          console.error('Failed to sync order:', order.id, response.status)
        }
      } catch (error) {
        console.error('Error syncing order:', order.id, error)
      }
    }
  }

  /**
   * 同步队列项
   */
  private async syncQueueItems(): Promise<void> {
    const queueItems = await this.storage.getSyncQueue()

    for (const item of queueItems) {
      if (item.retryCount >= item.maxRetries) {
        // 超过最大重试次数，移除项目
        await this.storage.removeSyncQueueItem(item.id)
        continue
      }

      try {
        const response = await fetch(item.url, {
          method: item.method,
          headers: item.headers,
          body: JSON.stringify(item.data)
        })

        if (response.ok) {
          await this.storage.removeSyncQueueItem(item.id)
          console.log('Sync queue item completed:', item.id)
        } else {
          await this.storage.updateSyncQueueItemRetry(item.id)
          console.error('Failed to sync queue item:', item.id, response.status)
        }
      } catch (error) {
        await this.storage.updateSyncQueueItemRetry(item.id)
        console.error('Error syncing queue item:', item.id, error)
      }
    }
  }

  /**
   * 开始定期同步
   */
  private startPeriodicSync(): void {
    if (this.syncInterval) {
      clearInterval(this.syncInterval)
    }

    // 每5分钟尝试同步一次
    this.syncInterval = window.setInterval(() => {
      if (navigator.onLine) {
        this.startSync()
      }
    }, 5 * 60 * 1000)
  }

  /**
   * 停止定期同步
   */
  stopPeriodicSync(): void {
    if (this.syncInterval) {
      clearInterval(this.syncInterval)
      this.syncInterval = null
    }
  }

  /**
   * 添加离线操作到队列
   */
  async addOfflineAction(
    type: string,
    url: string,
    method: string,
    data: any,
    headers?: Record<string, string>
  ): Promise<void> {
    await this.storage.addToSyncQueue({
      type: type as any,
      url,
      method,
      data,
      headers,
      maxRetries: 3
    })
  }
}

// 创建全局实例
export const offlineStorage = new OfflineStorageManager()
export const offlineSync = new OfflineSyncManager(offlineStorage)

export default offlineStorage
