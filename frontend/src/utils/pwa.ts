/**
 * PWA 工具类
 * 管理 Service Worker、离线功能和应用安装
 */

export interface PWAInstallPromptEvent extends Event {
  readonly platforms: string[]
  readonly userChoice: Promise<{
    outcome: 'accepted' | 'dismissed'
    platform: string
  }>
  prompt(): Promise<void>
}

export interface BeforeInstallPromptEvent extends Event {
  readonly platforms: string[]
  readonly userChoice: Promise<{
    outcome: 'accepted' | 'dismissed'
    platform: string
  }>
  prompt(): Promise<void>
}

// PWA 状态
export interface PWAStatus {
  isInstalled: boolean
  isInstallable: boolean
  isStandalone: boolean
  hasServiceWorker: boolean
  isOnline: boolean
}

// Service Worker 消息类型
export interface SWMessage {
  type: string
  payload?: any
}

class PWAManager {
  private deferredPrompt: BeforeInstallPromptEvent | null = null
  private swRegistration: ServiceWorkerRegistration | null = null
  private updateAvailable = false
  private callbacks: Map<string, Function[]> = new Map()

  constructor() {
    this.init()
  }

  /**
   * 初始化 PWA 管理器
   */
  private async init() {
    // 注册 Service Worker
    await this.registerServiceWorker()
    
    // 监听安装提示事件
    this.listenForInstallPrompt()
    
    // 监听网络状态变化
    this.listenForNetworkChanges()
    
    // 监听应用安装事件
    this.listenForAppInstalled()
  }

  /**
   * 注册 Service Worker
   */
  async registerServiceWorker(): Promise<ServiceWorkerRegistration | null> {
    if (!('serviceWorker' in navigator)) {
      console.warn('Service Worker not supported')
      return null
    }

    try {
      const registration = await navigator.serviceWorker.register('/sw.js', {
        scope: '/'
      })

      this.swRegistration = registration

      console.log('Service Worker registered successfully:', registration.scope)

      // 监听 Service Worker 更新
      registration.addEventListener('updatefound', () => {
        const newWorker = registration.installing
        if (newWorker) {
          newWorker.addEventListener('statechange', () => {
            if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
              this.updateAvailable = true
              this.emit('updateAvailable', registration)
            }
          })
        }
      })

      // 监听 Service Worker 消息
      navigator.serviceWorker.addEventListener('message', (event) => {
        this.handleServiceWorkerMessage(event.data)
      })

      return registration
    } catch (error) {
      console.error('Service Worker registration failed:', error)
      return null
    }
  }

  /**
   * 监听安装提示事件
   */
  private listenForInstallPrompt() {
    window.addEventListener('beforeinstallprompt', (e) => {
      e.preventDefault()
      this.deferredPrompt = e as BeforeInstallPromptEvent
      this.emit('installPromptAvailable', e)
    })
  }

  /**
   * 监听网络状态变化
   */
  private listenForNetworkChanges() {
    window.addEventListener('online', () => {
      this.emit('networkStatusChanged', { online: true })
    })

    window.addEventListener('offline', () => {
      this.emit('networkStatusChanged', { online: false })
    })
  }

  /**
   * 监听应用安装事件
   */
  private listenForAppInstalled() {
    window.addEventListener('appinstalled', () => {
      this.deferredPrompt = null
      this.emit('appInstalled')
    })
  }

  /**
   * 处理 Service Worker 消息
   */
  private handleServiceWorkerMessage(message: SWMessage) {
    switch (message.type) {
      case 'CACHE_UPDATED':
        this.emit('cacheUpdated', message.payload)
        break
      case 'OFFLINE_READY':
        this.emit('offlineReady')
        break
      case 'UPDATE_AVAILABLE':
        this.emit('updateAvailable', message.payload)
        break
      default:
        console.log('Unknown SW message:', message)
    }
  }

  /**
   * 获取 PWA 状态
   */
  getStatus(): PWAStatus {
    return {
      isInstalled: this.isInstalled(),
      isInstallable: this.isInstallable(),
      isStandalone: this.isStandalone(),
      hasServiceWorker: 'serviceWorker' in navigator,
      isOnline: navigator.onLine
    }
  }

  /**
   * 检查应用是否已安装
   */
  isInstalled(): boolean {
    return this.isStandalone() || 
           localStorage.getItem('pwa-installed') === 'true'
  }

  /**
   * 检查应用是否可安装
   */
  isInstallable(): boolean {
    return this.deferredPrompt !== null
  }

  /**
   * 检查是否在独立模式运行
   */
  isStandalone(): boolean {
    return window.matchMedia('(display-mode: standalone)').matches ||
           (window.navigator as any).standalone === true
  }

  /**
   * 显示安装提示
   */
  async showInstallPrompt(): Promise<boolean> {
    if (!this.deferredPrompt) {
      throw new Error('Install prompt not available')
    }

    try {
      await this.deferredPrompt.prompt()
      const { outcome } = await this.deferredPrompt.userChoice
      
      this.deferredPrompt = null
      
      if (outcome === 'accepted') {
        localStorage.setItem('pwa-installed', 'true')
        return true
      }
      
      return false
    } catch (error) {
      console.error('Install prompt failed:', error)
      throw error
    }
  }

  /**
   * 更新 Service Worker
   */
  async updateServiceWorker(): Promise<void> {
    if (!this.swRegistration) {
      throw new Error('Service Worker not registered')
    }

    try {
      await this.swRegistration.update()
      
      if (this.swRegistration.waiting) {
        // 通知 Service Worker 跳过等待
        this.postMessageToSW({ type: 'SKIP_WAITING' })
      }
    } catch (error) {
      console.error('Service Worker update failed:', error)
      throw error
    }
  }

  /**
   * 向 Service Worker 发送消息
   */
  postMessageToSW(message: SWMessage): void {
    if (!this.swRegistration || !this.swRegistration.active) {
      console.warn('Service Worker not active')
      return
    }

    this.swRegistration.active.postMessage(message)
  }

  /**
   * 获取缓存信息
   */
  async getCacheInfo(): Promise<any> {
    return new Promise((resolve) => {
      const channel = new MessageChannel()
      
      channel.port1.onmessage = (event) => {
        if (event.data.type === 'CACHE_INFO') {
          resolve(event.data.payload)
        }
      }

      this.postMessageToSW({
        type: 'GET_CACHE_INFO'
      })
    })
  }

  /**
   * 清理缓存
   */
  async clearCache(cacheName?: string): Promise<void> {
    return new Promise((resolve) => {
      const channel = new MessageChannel()
      
      channel.port1.onmessage = (event) => {
        if (event.data.type === 'CACHE_CLEARED') {
          resolve()
        }
      }

      this.postMessageToSW({
        type: 'CLEAR_CACHE',
        payload: { cacheName }
      })
    })
  }

  /**
   * 预缓存 URL 列表
   */
  async precacheUrls(urls: string[]): Promise<void> {
    return new Promise((resolve) => {
      const channel = new MessageChannel()
      
      channel.port1.onmessage = (event) => {
        if (event.data.type === 'PRECACHE_COMPLETE') {
          resolve()
        }
      }

      this.postMessageToSW({
        type: 'PRECACHE_URLS',
        payload: { urls }
      })
    })
  }

  /**
   * 检查网络连接
   */
  async checkNetworkConnection(): Promise<boolean> {
    if (!navigator.onLine) {
      return false
    }

    try {
      const response = await fetch('/favicon.ico', {
        method: 'HEAD',
        cache: 'no-cache'
      })
      return response.ok
    } catch {
      return false
    }
  }

  /**
   * 事件监听器管理
   */
  on(event: string, callback: Function): void {
    if (!this.callbacks.has(event)) {
      this.callbacks.set(event, [])
    }
    this.callbacks.get(event)!.push(callback)
  }

  off(event: string, callback: Function): void {
    const callbacks = this.callbacks.get(event)
    if (callbacks) {
      const index = callbacks.indexOf(callback)
      if (index > -1) {
        callbacks.splice(index, 1)
      }
    }
  }

  private emit(event: string, data?: any): void {
    const callbacks = this.callbacks.get(event)
    if (callbacks) {
      callbacks.forEach(callback => callback(data))
    }
  }

  /**
   * 获取设备信息
   */
  getDeviceInfo() {
    return {
      userAgent: navigator.userAgent,
      platform: navigator.platform,
      language: navigator.language,
      cookieEnabled: navigator.cookieEnabled,
      onLine: navigator.onLine,
      hardwareConcurrency: navigator.hardwareConcurrency,
      maxTouchPoints: navigator.maxTouchPoints,
      screen: {
        width: screen.width,
        height: screen.height,
        colorDepth: screen.colorDepth,
        pixelDepth: screen.pixelDepth
      },
      viewport: {
        width: window.innerWidth,
        height: window.innerHeight
      }
    }
  }

  /**
   * 检查浏览器支持的功能
   */
  getFeatureSupport() {
    return {
      serviceWorker: 'serviceWorker' in navigator,
      pushManager: 'PushManager' in window,
      notification: 'Notification' in window,
      backgroundSync: 'serviceWorker' in navigator && 'sync' in window.ServiceWorkerRegistration.prototype,
      webShare: 'share' in navigator,
      webShareTarget: 'serviceWorker' in navigator,
      badging: 'setAppBadge' in navigator,
      periodicBackgroundSync: 'serviceWorker' in navigator && 'periodicSync' in window.ServiceWorkerRegistration.prototype,
      webLocks: 'locks' in navigator,
      wakeLock: 'wakeLock' in navigator
    }
  }
}

// 创建全局实例
export const pwaManager = new PWAManager()

// 导出工具函数
export const isPWAInstalled = () => pwaManager.isInstalled()
export const isPWAInstallable = () => pwaManager.isInstallable()
export const isPWAStandalone = () => pwaManager.isStandalone()
export const showPWAInstallPrompt = () => pwaManager.showInstallPrompt()
export const updatePWA = () => pwaManager.updateServiceWorker()
export const getPWAStatus = () => pwaManager.getStatus()

export default pwaManager
