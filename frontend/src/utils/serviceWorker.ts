/**
 * Service Worker管理器
 * 用于注册、更新和管理Service Worker
 */

// Service Worker状态
export type ServiceWorkerStatus = 'unsupported' | 'installing' | 'installed' | 'updating' | 'updated' | 'error'

// Service Worker事件
export interface ServiceWorkerEvents {
  statusChange: (status: ServiceWorkerStatus) => void
  updateAvailable: () => void
  updateReady: () => void
  cacheUpdated: (cacheName: string) => void
  offline: () => void
  online: () => void
}

// 缓存信息
export interface CacheInfo {
  [cacheName: string]: {
    count: number
    urls: string[]
  }
}

/**
 * Service Worker管理器类
 */
export class ServiceWorkerManager {
  private registration: ServiceWorkerRegistration | null = null
  private status: ServiceWorkerStatus = 'unsupported'
  private listeners: Partial<ServiceWorkerEvents> = {}
  private updateCheckInterval: number | null = null

  constructor() {
    this.initNetworkStatusListener()
  }

  /**
   * 注册Service Worker
   */
  async register(swUrl = '/sw.js'): Promise<boolean> {
    if (!('serviceWorker' in navigator)) {
      this.setStatus('unsupported')
      return false
    }

    try {
      this.setStatus('installing')
      
      this.registration = await navigator.serviceWorker.register(swUrl, {
        scope: '/'
      })

      this.setupRegistrationListeners()
      this.setStatus('installed')
      
      // 定期检查更新
      this.startUpdateCheck()
      
      console.log('Service Worker registered successfully')
      return true
    } catch (error) {
      console.error('Service Worker registration failed:', error)
      this.setStatus('error')
      return false
    }
  }

  /**
   * 注销Service Worker
   */
  async unregister(): Promise<boolean> {
    if (!this.registration) {
      return false
    }

    try {
      const result = await this.registration.unregister()
      this.registration = null
      this.stopUpdateCheck()
      console.log('Service Worker unregistered successfully')
      return result
    } catch (error) {
      console.error('Service Worker unregistration failed:', error)
      return false
    }
  }

  /**
   * 检查更新
   */
  async checkForUpdate(): Promise<boolean> {
    if (!this.registration) {
      return false
    }

    try {
      await this.registration.update()
      return true
    } catch (error) {
      console.error('Service Worker update check failed:', error)
      return false
    }
  }

  /**
   * 跳过等待，立即激活新的Service Worker
   */
  async skipWaiting(): Promise<void> {
    if (!this.registration?.waiting) {
      return
    }

    this.sendMessage({ type: 'SKIP_WAITING' })
  }

  /**
   * 获取缓存信息
   */
  async getCacheInfo(): Promise<CacheInfo | null> {
    if (!this.registration?.active) {
      return null
    }

    return new Promise((resolve) => {
      const messageChannel = new MessageChannel()
      
      messageChannel.port1.onmessage = (event) => {
        if (event.data.type === 'CACHE_INFO') {
          resolve(event.data.payload)
        }
      }

      this.sendMessage(
        { type: 'GET_CACHE_INFO' },
        [messageChannel.port2]
      )
    })
  }

  /**
   * 清理缓存
   */
  async clearCache(cacheName?: string): Promise<void> {
    if (!this.registration?.active) {
      return
    }

    return new Promise((resolve) => {
      const messageChannel = new MessageChannel()
      
      messageChannel.port1.onmessage = (event) => {
        if (event.data.type === 'CACHE_CLEARED') {
          resolve()
        }
      }

      this.sendMessage(
        { type: 'CLEAR_CACHE', payload: { cacheName } },
        [messageChannel.port2]
      )
    })
  }

  /**
   * 预缓存URL列表
   */
  async precacheUrls(urls: string[]): Promise<void> {
    if (!this.registration?.active) {
      return
    }

    return new Promise((resolve) => {
      const messageChannel = new MessageChannel()
      
      messageChannel.port1.onmessage = (event) => {
        if (event.data.type === 'PRECACHE_COMPLETE') {
          resolve()
        }
      }

      this.sendMessage(
        { type: 'PRECACHE_URLS', payload: { urls } },
        [messageChannel.port2]
      )
    })
  }

  /**
   * 添加事件监听器
   */
  on<K extends keyof ServiceWorkerEvents>(event: K, listener: ServiceWorkerEvents[K]): void {
    this.listeners[event] = listener
  }

  /**
   * 移除事件监听器
   */
  off<K extends keyof ServiceWorkerEvents>(event: K): void {
    delete this.listeners[event]
  }

  /**
   * 获取当前状态
   */
  getStatus(): ServiceWorkerStatus {
    return this.status
  }

  /**
   * 检查是否支持Service Worker
   */
  isSupported(): boolean {
    return 'serviceWorker' in navigator
  }

  /**
   * 检查是否已安装
   */
  isInstalled(): boolean {
    return this.registration !== null
  }

  /**
   * 设置状态
   */
  private setStatus(status: ServiceWorkerStatus): void {
    if (this.status !== status) {
      this.status = status
      this.listeners.statusChange?.(status)
    }
  }

  /**
   * 设置注册监听器
   */
  private setupRegistrationListeners(): void {
    if (!this.registration) return

    // 监听安装事件
    this.registration.addEventListener('updatefound', () => {
      const newWorker = this.registration!.installing
      if (!newWorker) return

      this.setStatus('updating')

      newWorker.addEventListener('statechange', () => {
        switch (newWorker.state) {
          case 'installed':
            if (navigator.serviceWorker.controller) {
              // 有新版本可用
              this.listeners.updateAvailable?.()
            } else {
              // 首次安装完成
              this.setStatus('installed')
            }
            break
          case 'activated':
            this.setStatus('updated')
            this.listeners.updateReady?.()
            break
        }
      })
    })

    // 监听控制器变化
    navigator.serviceWorker.addEventListener('controllerchange', () => {
      window.location.reload()
    })
  }

  /**
   * 发送消息给Service Worker
   */
  private sendMessage(message: any, transfer?: Transferable[]): void {
    if (!this.registration?.active) {
      return
    }

    if (transfer) {
      this.registration.active.postMessage(message, { transfer })
    } else {
      this.registration.active.postMessage(message)
    }
  }

  /**
   * 开始定期检查更新
   */
  private startUpdateCheck(): void {
    // 每30分钟检查一次更新
    this.updateCheckInterval = window.setInterval(() => {
      this.checkForUpdate()
    }, 30 * 60 * 1000)
  }

  /**
   * 停止定期检查更新
   */
  private stopUpdateCheck(): void {
    if (this.updateCheckInterval) {
      clearInterval(this.updateCheckInterval)
      this.updateCheckInterval = null
    }
  }

  /**
   * 初始化网络状态监听
   */
  private initNetworkStatusListener(): void {
    window.addEventListener('online', () => {
      this.listeners.online?.()
    })

    window.addEventListener('offline', () => {
      this.listeners.offline?.()
    })
  }
}

/**
 * 缓存策略工具
 */
export class CacheStrategy {
  /**
   * 缓存优先策略
   */
  static async cacheFirst(request: Request, cacheName: string): Promise<Response> {
    const cache = await caches.open(cacheName)
    const cachedResponse = await cache.match(request)
    
    if (cachedResponse) {
      return cachedResponse
    }

    const networkResponse = await fetch(request)
    if (networkResponse.ok) {
      cache.put(request, networkResponse.clone())
    }
    
    return networkResponse
  }

  /**
   * 网络优先策略
   */
  static async networkFirst(request: Request, cacheName: string): Promise<Response> {
    try {
      const networkResponse = await fetch(request)
      if (networkResponse.ok) {
        const cache = await caches.open(cacheName)
        cache.put(request, networkResponse.clone())
      }
      return networkResponse
    } catch (error) {
      const cache = await caches.open(cacheName)
      const cachedResponse = await cache.match(request)
      if (cachedResponse) {
        return cachedResponse
      }
      throw error
    }
  }

  /**
   * 仅缓存策略
   */
  static async cacheOnly(request: Request, cacheName: string): Promise<Response> {
    const cache = await caches.open(cacheName)
    const cachedResponse = await cache.match(request)
    
    if (!cachedResponse) {
      throw new Error('No cached response available')
    }
    
    return cachedResponse
  }

  /**
   * 仅网络策略
   */
  static async networkOnly(request: Request): Promise<Response> {
    return fetch(request)
  }

  /**
   * 过期重新验证策略
   */
  static async staleWhileRevalidate(request: Request, cacheName: string): Promise<Response> {
    const cache = await caches.open(cacheName)
    const cachedResponse = await cache.match(request)
    
    // 后台更新缓存
    const networkResponsePromise = fetch(request).then(response => {
      if (response.ok) {
        cache.put(request, response.clone())
      }
      return response
    })

    // 如果有缓存，立即返回缓存
    if (cachedResponse) {
      return cachedResponse
    }

    // 否则等待网络响应
    return networkResponsePromise
  }
}

// 全局Service Worker管理器实例
let globalSWManager: ServiceWorkerManager | null = null

/**
 * 初始化Service Worker
 */
export function initServiceWorker(): ServiceWorkerManager {
  if (!globalSWManager) {
    globalSWManager = new ServiceWorkerManager()
    
    // 自动注册
    globalSWManager.register().then(success => {
      if (success) {
        console.log('Service Worker initialized successfully')
      }
    })
  }
  
  return globalSWManager
}

/**
 * 获取Service Worker管理器
 */
export function getServiceWorkerManager(): ServiceWorkerManager | null {
  return globalSWManager
}
