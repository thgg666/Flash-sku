/**
 * 前端性能监控和优化工具
 */

// 性能指标接口
export interface PerformanceMetrics {
  // 页面加载性能
  loadTime: number
  domContentLoaded: number
  firstPaint: number
  firstContentfulPaint: number
  largestContentfulPaint: number
  firstInputDelay: number
  cumulativeLayoutShift: number
  
  // 资源加载性能
  resourceLoadTime: number
  jsLoadTime: number
  cssLoadTime: number
  imageLoadTime: number
  
  // 运行时性能
  memoryUsage: number
  jsHeapSize: number
  
  // 网络性能
  connectionType: string
  effectiveType: string
  downlink: number
  rtt: number
}

// 性能监控器
export class PerformanceMonitor {
  private metrics: Partial<PerformanceMetrics> = {}
  private observers: PerformanceObserver[] = []

  constructor() {
    this.initObservers()
  }

  /**
   * 初始化性能观察器
   */
  private initObservers() {
    // 观察导航时间
    if ('PerformanceObserver' in window) {
      // 观察页面加载性能
      const navObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'navigation') {
            const navEntry = entry as PerformanceNavigationTiming
            this.metrics.loadTime = navEntry.loadEventEnd - navEntry.startTime
            this.metrics.domContentLoaded = navEntry.domContentLoadedEventEnd - navEntry.startTime
          }
        }
      })
      navObserver.observe({ entryTypes: ['navigation'] })
      this.observers.push(navObserver)

      // 观察绘制性能
      const paintObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.name === 'first-paint') {
            this.metrics.firstPaint = entry.startTime
          } else if (entry.name === 'first-contentful-paint') {
            this.metrics.firstContentfulPaint = entry.startTime
          }
        }
      })
      paintObserver.observe({ entryTypes: ['paint'] })
      this.observers.push(paintObserver)

      // 观察最大内容绘制
      const lcpObserver = new PerformanceObserver((list) => {
        const entries = list.getEntries()
        const lastEntry = entries[entries.length - 1]
        this.metrics.largestContentfulPaint = lastEntry.startTime
      })
      lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] })
      this.observers.push(lcpObserver)

      // 观察累积布局偏移
      const clsObserver = new PerformanceObserver((list) => {
        let clsValue = 0
        for (const entry of list.getEntries()) {
          if (!(entry as any).hadRecentInput) {
            clsValue += (entry as any).value
          }
        }
        this.metrics.cumulativeLayoutShift = clsValue
      })
      clsObserver.observe({ entryTypes: ['layout-shift'] })
      this.observers.push(clsObserver)

      // 观察资源加载
      const resourceObserver = new PerformanceObserver((list) => {
        let jsTime = 0, cssTime = 0, imageTime = 0
        
        for (const entry of list.getEntries()) {
          const resource = entry as PerformanceResourceTiming
          const loadTime = resource.responseEnd - resource.startTime
          
          if (resource.name.includes('.js')) {
            jsTime += loadTime
          } else if (resource.name.includes('.css')) {
            cssTime += loadTime
          } else if (resource.name.match(/\.(jpg|jpeg|png|gif|webp|svg)$/)) {
            imageTime += loadTime
          }
        }
        
        this.metrics.jsLoadTime = jsTime
        this.metrics.cssLoadTime = cssTime
        this.metrics.imageLoadTime = imageTime
        this.metrics.resourceLoadTime = jsTime + cssTime + imageTime
      })
      resourceObserver.observe({ entryTypes: ['resource'] })
      this.observers.push(resourceObserver)
    }

    // 监控内存使用
    this.monitorMemoryUsage()
    
    // 监控网络信息
    this.monitorNetworkInfo()
  }

  /**
   * 监控内存使用情况
   */
  private monitorMemoryUsage() {
    if ('memory' in performance) {
      const memory = (performance as any).memory
      this.metrics.memoryUsage = memory.usedJSHeapSize
      this.metrics.jsHeapSize = memory.totalJSHeapSize
    }
  }

  /**
   * 监控网络信息
   */
  private monitorNetworkInfo() {
    if ('connection' in navigator) {
      const connection = (navigator as any).connection
      this.metrics.connectionType = connection.type || 'unknown'
      this.metrics.effectiveType = connection.effectiveType || 'unknown'
      this.metrics.downlink = connection.downlink || 0
      this.metrics.rtt = connection.rtt || 0
    }
  }

  /**
   * 获取当前性能指标
   */
  getMetrics(): Partial<PerformanceMetrics> {
    this.monitorMemoryUsage()
    this.monitorNetworkInfo()
    return { ...this.metrics }
  }

  /**
   * 销毁监控器
   */
  destroy() {
    this.observers.forEach(observer => observer.disconnect())
    this.observers = []
  }
}

/**
 * 代码分割工具
 */
export class CodeSplittingHelper {
  /**
   * 动态导入组件
   */
  static async importComponent(componentPath: string) {
    try {
      const startTime = performance.now()
      const component = await import(componentPath)
      const loadTime = performance.now() - startTime
      
      console.log(`Component ${componentPath} loaded in ${loadTime.toFixed(2)}ms`)
      return component
    } catch (error) {
      console.error(`Failed to load component ${componentPath}:`, error)
      throw error
    }
  }

  /**
   * 预加载组件
   */
  static preloadComponent(componentPath: string) {
    const link = document.createElement('link')
    link.rel = 'modulepreload'
    link.href = componentPath
    document.head.appendChild(link)
  }

  /**
   * 预加载路由
   */
  static preloadRoute(routePath: string) {
    // 这里可以根据路由配置预加载对应的组件
    console.log(`Preloading route: ${routePath}`)
  }
}

/**
 * 懒加载工具
 */
export class LazyLoadHelper {
  private observer: IntersectionObserver | null = null

  constructor() {
    this.initIntersectionObserver()
  }

  /**
   * 初始化交叉观察器
   */
  private initIntersectionObserver() {
    if ('IntersectionObserver' in window) {
      this.observer = new IntersectionObserver(
        (entries) => {
          entries.forEach(entry => {
            if (entry.isIntersecting) {
              const element = entry.target as HTMLElement
              this.loadElement(element)
              this.observer?.unobserve(element)
            }
          })
        },
        {
          rootMargin: '50px 0px',
          threshold: 0.1
        }
      )
    }
  }

  /**
   * 观察元素
   */
  observe(element: HTMLElement) {
    if (this.observer) {
      this.observer.observe(element)
    }
  }

  /**
   * 加载元素
   */
  private loadElement(element: HTMLElement) {
    if (element.tagName === 'IMG') {
      const img = element as HTMLImageElement
      const dataSrc = img.dataset.src
      if (dataSrc) {
        img.src = dataSrc
        img.classList.add('loaded')
      }
    } else if (element.dataset.component) {
      // 懒加载组件
      this.loadComponent(element, element.dataset.component)
    }
  }

  /**
   * 懒加载组件
   */
  private async loadComponent(element: HTMLElement, componentPath: string) {
    try {
      await CodeSplittingHelper.importComponent(componentPath)
      // 这里可以动态渲染组件
      element.classList.add('component-loaded')
    } catch (error) {
      console.error('Failed to lazy load component:', error)
    }
  }

  /**
   * 销毁观察器
   */
  destroy() {
    if (this.observer) {
      this.observer.disconnect()
      this.observer = null
    }
  }
}

/**
 * 资源预加载管理器
 */
export class ResourcePreloader {
  private preloadedResources = new Set<string>()

  /**
   * 预加载图片
   */
  preloadImage(src: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.preloadedResources.has(src)) {
        resolve()
        return
      }

      const img = new Image()
      img.onload = () => {
        this.preloadedResources.add(src)
        resolve()
      }
      img.onerror = reject
      img.src = src
    })
  }

  /**
   * 预加载多个图片
   */
  async preloadImages(srcs: string[]): Promise<void> {
    const promises = srcs.map(src => this.preloadImage(src))
    await Promise.all(promises)
  }

  /**
   * 预加载CSS
   */
  preloadCSS(href: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.preloadedResources.has(href)) {
        resolve()
        return
      }

      const link = document.createElement('link')
      link.rel = 'preload'
      link.as = 'style'
      link.href = href
      link.onload = () => {
        this.preloadedResources.add(href)
        resolve()
      }
      link.onerror = reject
      document.head.appendChild(link)
    })
  }

  /**
   * 预加载JavaScript
   */
  preloadJS(src: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.preloadedResources.has(src)) {
        resolve()
        return
      }

      const link = document.createElement('link')
      link.rel = 'modulepreload'
      link.href = src
      link.onload = () => {
        this.preloadedResources.add(src)
        resolve()
      }
      link.onerror = reject
      document.head.appendChild(link)
    })
  }
}

/**
 * 性能优化建议生成器
 */
export class PerformanceOptimizer {
  /**
   * 分析性能指标并生成建议
   */
  static analyzeAndSuggest(metrics: Partial<PerformanceMetrics>): string[] {
    const suggestions: string[] = []

    // 检查页面加载时间
    if (metrics.loadTime && metrics.loadTime > 3000) {
      suggestions.push('页面加载时间过长，建议优化资源大小和数量')
    }

    // 检查首次内容绘制
    if (metrics.firstContentfulPaint && metrics.firstContentfulPaint > 1500) {
      suggestions.push('首次内容绘制时间过长，建议优化关键渲染路径')
    }

    // 检查最大内容绘制
    if (metrics.largestContentfulPaint && metrics.largestContentfulPaint > 2500) {
      suggestions.push('最大内容绘制时间过长，建议优化主要内容的加载')
    }

    // 检查累积布局偏移
    if (metrics.cumulativeLayoutShift && metrics.cumulativeLayoutShift > 0.1) {
      suggestions.push('累积布局偏移过大，建议为图片和广告预留空间')
    }

    // 检查JavaScript加载时间
    if (metrics.jsLoadTime && metrics.jsLoadTime > 1000) {
      suggestions.push('JavaScript加载时间过长，建议进行代码分割和压缩')
    }

    // 检查内存使用
    if (metrics.memoryUsage && metrics.memoryUsage > 50 * 1024 * 1024) { // 50MB
      suggestions.push('内存使用过高，建议检查内存泄漏和优化数据结构')
    }

    // 检查网络条件
    if (metrics.effectiveType === '2g' || metrics.effectiveType === 'slow-2g') {
      suggestions.push('检测到慢速网络，建议启用更激进的优化策略')
    }

    return suggestions
  }
}

// 全局性能监控实例
let globalPerformanceMonitor: PerformanceMonitor | null = null

/**
 * 初始化性能监控
 */
export function initPerformanceMonitoring() {
  if (!globalPerformanceMonitor) {
    globalPerformanceMonitor = new PerformanceMonitor()
    
    // 在页面加载完成后记录性能指标
    window.addEventListener('load', () => {
      setTimeout(() => {
        const metrics = globalPerformanceMonitor?.getMetrics()
        if (metrics && import.meta.env.DEV) {
          console.log('Performance Metrics:', metrics)
          const suggestions = PerformanceOptimizer.analyzeAndSuggest(metrics)
          if (suggestions.length > 0) {
            console.log('Performance Suggestions:', suggestions)
          }
        }
      }, 1000)
    })
  }
  
  return globalPerformanceMonitor
}

/**
 * 获取性能指标
 */
export function getPerformanceMetrics(): Partial<PerformanceMetrics> | null {
  return globalPerformanceMonitor?.getMetrics() || null
}

/**
 * 销毁性能监控
 */
export function destroyPerformanceMonitoring() {
  if (globalPerformanceMonitor) {
    globalPerformanceMonitor.destroy()
    globalPerformanceMonitor = null
  }
}
