/**
 * 运行时性能优化工具
 * 提供内存管理、渲染优化、性能监控等功能
 */

interface PerformanceMetrics {
  memoryUsage: number
  renderTime: number
  componentCount: number
  eventListeners: number
  domNodes: number
}

interface OptimizationConfig {
  enableMemoryCleanup: boolean
  enableRenderOptimization: boolean
  enableEventOptimization: boolean
  memoryThreshold: number
  renderThreshold: number
}

class PerformanceOptimizer {
  private config: OptimizationConfig
  private metrics: PerformanceMetrics
  private observers: Map<string, any> = new Map()
  private cleanupTasks: Set<() => void> = new Set()
  private renderQueue: Set<() => void> = new Set()
  private isOptimizing = false

  constructor(config: Partial<OptimizationConfig> = {}) {
    this.config = {
      enableMemoryCleanup: true,
      enableRenderOptimization: true,
      enableEventOptimization: true,
      memoryThreshold: 50 * 1024 * 1024, // 50MB
      renderThreshold: 16, // 16ms (60fps)
      ...config
    }

    this.metrics = {
      memoryUsage: 0,
      renderTime: 0,
      componentCount: 0,
      eventListeners: 0,
      domNodes: 0
    }

    this.init()
  }

  private init() {
    if (this.config.enableMemoryCleanup) {
      this.setupMemoryMonitoring()
    }

    if (this.config.enableRenderOptimization) {
      this.setupRenderOptimization()
    }

    if (this.config.enableEventOptimization) {
      this.setupEventOptimization()
    }

    // 定期性能检查
    setInterval(() => {
      this.performanceCheck()
    }, 5000)
  }

  /**
   * 内存监控和清理
   */
  private setupMemoryMonitoring() {
    // 监控内存使用
    if ('memory' in performance) {
      const checkMemory = () => {
        const memory = (performance as any).memory
        this.metrics.memoryUsage = memory.usedJSHeapSize

        if (memory.usedJSHeapSize > this.config.memoryThreshold) {
          this.triggerMemoryCleanup()
        }
      }

      setInterval(checkMemory, 10000) // 每10秒检查一次
    }

    // 页面卸载时清理
    window.addEventListener('beforeunload', () => {
      this.cleanup()
    })
  }

  /**
   * 渲染性能优化
   */
  private setupRenderOptimization() {
    // 使用 requestIdleCallback 优化渲染
    const scheduleRender = (callback: () => void) => {
      if ('requestIdleCallback' in window) {
        requestIdleCallback(callback, { timeout: 1000 })
      } else {
        setTimeout(callback, 0)
      }
    }

    // 批量处理渲染任务
    const flushRenderQueue = () => {
      if (this.renderQueue.size === 0) return

      const startTime = performance.now()
      
      for (const task of this.renderQueue) {
        try {
          task()
        } catch (error) {
          console.error('Render task error:', error)
        }
      }
      
      this.renderQueue.clear()
      this.metrics.renderTime = performance.now() - startTime
    }

    // 定期刷新渲染队列
    setInterval(() => {
      if (this.renderQueue.size > 0) {
        scheduleRender(flushRenderQueue)
      }
    }, 16) // 60fps
  }

  /**
   * 事件优化
   */
  private setupEventOptimization() {
    // 防抖和节流工具
    const debounceMap = new Map<string, number>()
    const throttleMap = new Map<string, number>()

    // 重写 addEventListener 来跟踪事件监听器
    const originalAddEventListener = EventTarget.prototype.addEventListener
    const originalRemoveEventListener = EventTarget.prototype.removeEventListener
    const self = this

    EventTarget.prototype.addEventListener = function(type, listener, options) {
      self.metrics.eventListeners++
      return originalAddEventListener.call(this, type, listener, options)
    }

    EventTarget.prototype.removeEventListener = function(type, listener, options) {
      self.metrics.eventListeners--
      return originalRemoveEventListener.call(this, type, listener, options)
    }
  }

  /**
   * 添加渲染任务到队列
   */
  addRenderTask(task: () => void) {
    this.renderQueue.add(task)
  }

  /**
   * 添加清理任务
   */
  addCleanupTask(task: () => void) {
    this.cleanupTasks.add(task)
  }

  /**
   * 移除清理任务
   */
  removeCleanupTask(task: () => void) {
    this.cleanupTasks.delete(task)
  }

  /**
   * 触发内存清理
   */
  private triggerMemoryCleanup() {
    if (this.isOptimizing) return

    this.isOptimizing = true

    try {
      // 执行清理任务
      for (const task of this.cleanupTasks) {
        try {
          task()
        } catch (error) {
          console.error('Cleanup task error:', error)
        }
      }

      // 清理观察者
      for (const [key, observer] of this.observers) {
        if (observer && typeof observer.disconnect === 'function') {
          observer.disconnect()
        }
      }

      // 强制垃圾回收（如果可用）
      if ('gc' in window) {
        (window as any).gc()
      }

      console.log('Memory cleanup completed')
    } finally {
      this.isOptimizing = false
    }
  }

  /**
   * 性能检查
   */
  private performanceCheck() {
    // 更新DOM节点数量
    this.metrics.domNodes = document.querySelectorAll('*').length

    // 检查是否需要优化
    if (this.metrics.renderTime > this.config.renderThreshold) {
      console.warn('Render time exceeds threshold:', this.metrics.renderTime)
    }

    if (this.metrics.domNodes > 5000) {
      console.warn('Too many DOM nodes:', this.metrics.domNodes)
    }
  }

  /**
   * 创建防抖函数
   */
  debounce<T extends (...args: any[]) => any>(
    func: T,
    delay: number,
    key?: string
  ): T {
    const debounceKey = key || func.toString()
    let timeoutId: number

    return ((...args: Parameters<T>) => {
      clearTimeout(timeoutId)
      timeoutId = window.setTimeout(() => func.apply(this, args), delay)
    }) as T
  }

  /**
   * 创建节流函数
   */
  throttle<T extends (...args: any[]) => any>(
    func: T,
    delay: number,
    key?: string
  ): T {
    const throttleKey = key || func.toString()
    let lastCall = 0

    return ((...args: Parameters<T>) => {
      const now = Date.now()
      if (now - lastCall >= delay) {
        lastCall = now
        return func.apply(this, args)
      }
    }) as T
  }

  /**
   * 虚拟滚动优化
   */
  createVirtualScroller(container: HTMLElement, itemHeight: number, buffer = 5) {
    let startIndex = 0
    let endIndex = 0
    const visibleCount = Math.ceil(container.clientHeight / itemHeight)

    const updateVisibleItems = this.throttle(() => {
      const scrollTop = container.scrollTop
      startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - buffer)
      endIndex = Math.min(
        startIndex + visibleCount + buffer * 2,
        container.children.length
      )

      // 隐藏不可见的元素
      for (let i = 0; i < container.children.length; i++) {
        const child = container.children[i] as HTMLElement
        if (i < startIndex || i > endIndex) {
          child.style.display = 'none'
        } else {
          child.style.display = ''
        }
      }
    }, 16)

    container.addEventListener('scroll', updateVisibleItems)
    
    // 返回清理函数
    return () => {
      container.removeEventListener('scroll', updateVisibleItems)
    }
  }

  /**
   * 图片懒加载优化
   */
  createImageLazyLoader(options: IntersectionObserverInit = {}) {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const img = entry.target as HTMLImageElement
          const src = img.dataset.src
          
          if (src) {
            img.src = src
            img.removeAttribute('data-src')
            observer.unobserve(img)
          }
        }
      })
    }, {
      rootMargin: '50px',
      threshold: 0.1,
      ...options
    })

    this.observers.set('imageLoader', observer)

    return {
      observe: (img: HTMLImageElement) => observer.observe(img),
      unobserve: (img: HTMLImageElement) => observer.unobserve(img),
      disconnect: () => observer.disconnect()
    }
  }

  /**
   * 获取性能指标
   */
  getMetrics(): PerformanceMetrics {
    return { ...this.metrics }
  }

  /**
   * 清理所有资源
   */
  cleanup() {
    // 执行所有清理任务
    for (const task of this.cleanupTasks) {
      try {
        task()
      } catch (error) {
        console.error('Cleanup task error:', error)
      }
    }

    // 断开所有观察者
    for (const [key, observer] of this.observers) {
      if (observer && typeof observer.disconnect === 'function') {
        observer.disconnect()
      }
    }

    // 清空队列
    this.renderQueue.clear()
    this.cleanupTasks.clear()
    this.observers.clear()
  }
}

// 创建全局实例
export const performanceOptimizer = new PerformanceOptimizer()

export default performanceOptimizer
