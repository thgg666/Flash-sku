/**
 * 路由预加载组合式函数
 * 用于智能预加载路由组件，提升用户体验
 */

import { ref, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'

// 预加载策略
export type PreloadStrategy = 'hover' | 'visible' | 'idle' | 'immediate'

// 预加载配置
export interface PreloadConfig {
  strategy: PreloadStrategy
  delay?: number
  priority?: 'high' | 'low'
  prefetch?: boolean
}

// 预加载状态
export interface PreloadState {
  loading: boolean
  loaded: boolean
  error: boolean
}

/**
 * 路由预加载组合式函数
 */
export function useRoutePreload() {
  const router = useRouter()
  const preloadedRoutes = new Set<string>()
  const preloadStates = ref<Record<string, PreloadState>>({})

  /**
   * 预加载路由
   */
  const preloadRoute = async (routeName: string, config: PreloadConfig = { strategy: 'hover' }) => {
    // 如果已经预加载过，直接返回
    if (preloadedRoutes.has(routeName)) {
      return
    }

    // 设置加载状态
    preloadStates.value[routeName] = {
      loading: true,
      loaded: false,
      error: false
    }

    try {
      // 查找路由配置
      const route = router.getRoutes().find(r => r.name === routeName)
      if (!route) {
        throw new Error(`Route ${routeName} not found`)
      }

      // 根据策略执行预加载
      await executePreloadStrategy(route, config)

      // 标记为已预加载
      preloadedRoutes.add(routeName)
      preloadStates.value[routeName] = {
        loading: false,
        loaded: true,
        error: false
      }

      console.log(`Route ${routeName} preloaded successfully`)
    } catch (error) {
      console.error(`Failed to preload route ${routeName}:`, error)
      preloadStates.value[routeName] = {
        loading: false,
        loaded: false,
        error: true
      }
    }
  }

  /**
   * 执行预加载策略
   */
  const executePreloadStrategy = async (route: any, config: PreloadConfig) => {
    const { strategy, delay = 0 } = config

    // 延迟执行
    if (delay > 0) {
      await new Promise(resolve => setTimeout(resolve, delay))
    }

    switch (strategy) {
      case 'immediate':
        await loadRouteComponent(route)
        break
      case 'idle':
        await loadOnIdle(route)
        break
      case 'visible':
        // visible策略需要在组件中配合使用
        await loadRouteComponent(route)
        break
      case 'hover':
        // hover策略需要在组件中配合使用
        await loadRouteComponent(route)
        break
      default:
        await loadRouteComponent(route)
    }
  }

  /**
   * 加载路由组件
   */
  const loadRouteComponent = async (route: any) => {
    if (typeof route.component === 'function') {
      await route.component()
    } else if (route.components) {
      // 加载所有命名视图组件
      const promises = Object.values(route.components).map((component: any) => {
        if (typeof component === 'function') {
          return component()
        }
        return Promise.resolve()
      })
      await Promise.all(promises)
    }
  }

  /**
   * 在浏览器空闲时加载
   */
  const loadOnIdle = async (route: any) => {
    return new Promise<void>((resolve) => {
      if ('requestIdleCallback' in window) {
        requestIdleCallback(() => {
          loadRouteComponent(route).then(resolve)
        })
      } else {
        // 降级到setTimeout
        setTimeout(() => {
          loadRouteComponent(route).then(resolve)
        }, 100)
      }
    })
  }

  /**
   * 预加载多个路由
   */
  const preloadRoutes = async (routeNames: string[], config?: PreloadConfig) => {
    const promises = routeNames.map(name => preloadRoute(name, config))
    await Promise.all(promises)
  }

  /**
   * 智能预加载 - 根据用户行为预测可能访问的路由
   */
  const smartPreload = () => {
    // 预加载当前路由的子路由
    const currentRoute = router.currentRoute.value
    const childRoutes = router.getRoutes().filter(route => 
      route.path.startsWith(currentRoute.path) && route.path !== currentRoute.path
    )

    childRoutes.forEach(route => {
      if (route.name) {
        preloadRoute(route.name.toString(), { strategy: 'idle', delay: 1000 })
      }
    })

    // 预加载常用路由
    const commonRoutes = ['home', 'activities', 'about']
    commonRoutes.forEach(routeName => {
      if (routeName !== currentRoute.name) {
        preloadRoute(routeName, { strategy: 'idle', delay: 2000 })
      }
    })
  }

  /**
   * 获取预加载状态
   */
  const getPreloadState = (routeName: string): PreloadState => {
    return preloadStates.value[routeName] || {
      loading: false,
      loaded: false,
      error: false
    }
  }

  /**
   * 清除预加载缓存
   */
  const clearPreloadCache = () => {
    preloadedRoutes.clear()
    preloadStates.value = {}
  }

  return {
    preloadRoute,
    preloadRoutes,
    smartPreload,
    getPreloadState,
    clearPreloadCache,
    preloadStates: preloadStates.value
  }
}

/**
 * 链接预加载指令
 */
export function createPreloadDirective() {
  const { preloadRoute } = useRoutePreload()

  return {
    mounted(el: HTMLElement, binding: any) {
      const routeName = binding.value?.route || binding.value
      const strategy = binding.value?.strategy || 'hover'
      const delay = binding.value?.delay || 0

      if (!routeName) return

      const handlePreload = () => {
        preloadRoute(routeName, { strategy, delay })
      }

      switch (strategy) {
        case 'hover':
          el.addEventListener('mouseenter', handlePreload, { once: true })
          break
        case 'visible':
          // 使用IntersectionObserver
          if ('IntersectionObserver' in window) {
            const observer = new IntersectionObserver(
              (entries) => {
                entries.forEach(entry => {
                  if (entry.isIntersecting) {
                    handlePreload()
                    observer.unobserve(el)
                  }
                })
              },
              { threshold: 0.1 }
            )
            observer.observe(el)
            
            // 保存observer以便清理
            ;(el as any)._preloadObserver = observer
          }
          break
        case 'immediate':
          handlePreload()
          break
        case 'idle':
          if ('requestIdleCallback' in window) {
            requestIdleCallback(handlePreload)
          } else {
            setTimeout(handlePreload, 100)
          }
          break
      }
    },

    unmounted(el: HTMLElement) {
      // 清理observer
      const observer = (el as any)._preloadObserver
      if (observer) {
        observer.disconnect()
        delete (el as any)._preloadObserver
      }
    }
  }
}

/**
 * 预加载链接组件
 */
export const PreloadLink = {
  name: 'PreloadLink',
  props: {
    to: {
      type: [String, Object],
      required: true
    },
    strategy: {
      type: String as () => PreloadStrategy,
      default: 'hover'
    },
    delay: {
      type: Number,
      default: 0
    },
    tag: {
      type: String,
      default: 'router-link'
    }
  },
  setup(props: any, { slots }: any) {
    const { preloadRoute } = useRoutePreload()
    const router = useRouter()

    const handlePreload = () => {
      const routeName = typeof props.to === 'string' ? props.to : props.to.name
      if (routeName) {
        preloadRoute(routeName, {
          strategy: props.strategy,
          delay: props.delay
        })
      }
    }

    onMounted(() => {
      if (props.strategy === 'immediate') {
        handlePreload()
      }
    })

    return () => {
      const linkProps = {
        to: props.to,
        onMouseenter: props.strategy === 'hover' ? handlePreload : undefined
      }

      if (props.tag === 'router-link') {
        return h('router-link', linkProps, slots.default?.())
      } else {
        return h(props.tag, {
          ...linkProps,
          onClick: () => {
            // 手动导航
            router.push(props.to)
          }
        }, slots.default?.())
      }
    }
  }
}

/**
 * 路由预加载管理器
 */
export class RoutePreloadManager {
  private preloadQueue: Array<{ routeName: string; config: PreloadConfig }> = []
  private isProcessing = false

  /**
   * 添加到预加载队列
   */
  addToQueue(routeName: string, config: PreloadConfig) {
    this.preloadQueue.push({ routeName, config })
    this.processQueue()
  }

  /**
   * 处理预加载队列
   */
  private async processQueue() {
    if (this.isProcessing || this.preloadQueue.length === 0) {
      return
    }

    this.isProcessing = true
    const { preloadRoute } = useRoutePreload()

    while (this.preloadQueue.length > 0) {
      const item = this.preloadQueue.shift()
      if (item) {
        try {
          await preloadRoute(item.routeName, item.config)
        } catch (error) {
          console.error('Failed to preload route from queue:', error)
        }
      }
    }

    this.isProcessing = false
  }

  /**
   * 清空队列
   */
  clearQueue() {
    this.preloadQueue = []
  }
}

// 全局预加载管理器实例
export const globalPreloadManager = new RoutePreloadManager()
