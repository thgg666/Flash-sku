/**
 * Vue 性能优化 Composable
 * 提供组件级别的性能优化功能
 */

import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { performanceOptimizer } from '@/utils/performanceOptimizer'

export function usePerformanceOptimization() {
  const isOptimizing = ref(false)
  const metrics = ref(performanceOptimizer.getMetrics())
  const cleanupTasks: (() => void)[] = []

  /**
   * 组件性能监控
   */
  const useComponentPerformance = (componentName: string) => {
    const renderTime = ref(0)
    const mountTime = ref(0)
    const updateCount = ref(0)

    onMounted(() => {
      const startTime = performance.now()
      
      nextTick(() => {
        mountTime.value = performance.now() - startTime
        console.log(`Component ${componentName} mounted in ${mountTime.value.toFixed(2)}ms`)
      })
    })

    // 监控更新性能
    const trackUpdate = () => {
      const startTime = performance.now()
      updateCount.value++
      
      nextTick(() => {
        renderTime.value = performance.now() - startTime
        if (renderTime.value > 16) {
          console.warn(`Component ${componentName} render time: ${renderTime.value.toFixed(2)}ms`)
        }
      })
    }

    return {
      renderTime,
      mountTime,
      updateCount,
      trackUpdate
    }
  }

  /**
   * 内存泄漏防护
   */
  const useMemoryGuard = () => {
    const addCleanupTask = (task: () => void) => {
      cleanupTasks.push(task)
      performanceOptimizer.addCleanupTask(task)
    }

    const removeCleanupTask = (task: () => void) => {
      const index = cleanupTasks.indexOf(task)
      if (index > -1) {
        cleanupTasks.splice(index, 1)
        performanceOptimizer.removeCleanupTask(task)
      }
    }

    // 自动清理定时器
    const createSafeInterval = (callback: () => void, delay: number) => {
      const intervalId = setInterval(callback, delay)
      const cleanup = () => clearInterval(intervalId)
      addCleanupTask(cleanup)
      return cleanup
    }

    const createSafeTimeout = (callback: () => void, delay: number) => {
      const timeoutId = setTimeout(callback, delay)
      const cleanup = () => clearTimeout(timeoutId)
      addCleanupTask(cleanup)
      return cleanup
    }

    // 自动清理事件监听器
    const createSafeEventListener = (
      target: EventTarget,
      event: string,
      handler: EventListener,
      options?: AddEventListenerOptions
    ) => {
      target.addEventListener(event, handler, options)
      const cleanup = () => target.removeEventListener(event, handler, options)
      addCleanupTask(cleanup)
      return cleanup
    }

    return {
      addCleanupTask,
      removeCleanupTask,
      createSafeInterval,
      createSafeTimeout,
      createSafeEventListener
    }
  }

  /**
   * 渲染优化
   */
  const useRenderOptimization = () => {
    // 批量DOM更新
    const batchDOMUpdates = (updates: (() => void)[]) => {
      performanceOptimizer.addRenderTask(() => {
        updates.forEach(update => {
          try {
            update()
          } catch (error) {
            console.error('DOM update error:', error)
          }
        })
      })
    }

    // 防抖更新
    const createDebouncedUpdate = (
      updateFn: () => void,
      delay = 100,
      key?: string
    ) => {
      return performanceOptimizer.debounce(updateFn, delay, key)
    }

    // 节流更新
    const createThrottledUpdate = (
      updateFn: () => void,
      delay = 16,
      key?: string
    ) => {
      return performanceOptimizer.throttle(updateFn, delay, key)
    }

    return {
      batchDOMUpdates,
      createDebouncedUpdate,
      createThrottledUpdate
    }
  }

  /**
   * 虚拟滚动
   */
  const useVirtualScroll = (
    containerRef: Ref<HTMLElement | null>,
    itemHeight: number,
    buffer = 5
  ) => {
    const visibleItems = ref<any[]>([])
    const startIndex = ref(0)
    const endIndex = ref(0)
    let cleanup: (() => void) | null = null

    const setupVirtualScroll = (items: any[]) => {
      if (!containerRef.value) return

      cleanup = performanceOptimizer.createVirtualScroller(
        containerRef.value,
        itemHeight,
        buffer
      )

      const updateVisibleItems = () => {
        if (!containerRef.value) return

        const scrollTop = containerRef.value.scrollTop
        const visibleCount = Math.ceil(containerRef.value.clientHeight / itemHeight)
        
        startIndex.value = Math.max(0, Math.floor(scrollTop / itemHeight) - buffer)
        endIndex.value = Math.min(
          startIndex.value + visibleCount + buffer * 2,
          items.length
        )

        visibleItems.value = items.slice(startIndex.value, endIndex.value)
      }

      updateVisibleItems()
      
      // 监听滚动
      const throttledUpdate = performanceOptimizer.throttle(updateVisibleItems, 16)
      containerRef.value.addEventListener('scroll', throttledUpdate)
    }

    onUnmounted(() => {
      cleanup?.()
    })

    return {
      visibleItems,
      startIndex,
      endIndex,
      setupVirtualScroll
    }
  }

  /**
   * 图片懒加载
   */
  const useLazyImages = (options?: IntersectionObserverInit) => {
    const lazyLoader = performanceOptimizer.createImageLazyLoader(options)
    const { addCleanupTask } = useMemoryGuard()

    addCleanupTask(() => lazyLoader.disconnect())

    const observeImage = (img: HTMLImageElement) => {
      lazyLoader.observe(img)
    }

    const unobserveImage = (img: HTMLImageElement) => {
      lazyLoader.unobserve(img)
    }

    return {
      observeImage,
      unobserveImage
    }
  }

  /**
   * 响应式数据优化
   */
  const useReactiveOptimization = () => {
    // 浅层响应式（用于大型对象）
    const createShallowReactive = <T extends object>(obj: T) => {
      return shallowReactive(obj)
    }

    // 只读响应式（用于不需要修改的数据）
    const createReadonlyReactive = <T>(obj: T) => {
      return readonly(ref(obj))
    }

    // 计算属性缓存优化
    const createMemoizedComputed = <T>(
      getter: () => T,
      deps: any[] = []
    ) => {
      return computed(() => {
        // 依赖项变化时才重新计算
        deps.forEach(dep => dep.value) // 触发依赖收集
        return getter()
      })
    }

    return {
      createShallowReactive,
      createReadonlyReactive,
      createMemoizedComputed
    }
  }

  /**
   * 更新性能指标
   */
  const updateMetrics = () => {
    metrics.value = performanceOptimizer.getMetrics()
  }

  // 定期更新指标
  onMounted(() => {
    const interval = setInterval(updateMetrics, 5000)
    
    onUnmounted(() => {
      clearInterval(interval)
      
      // 清理所有任务
      cleanupTasks.forEach(task => {
        performanceOptimizer.removeCleanupTask(task)
      })
    })
  })

  return {
    // 状态
    isOptimizing,
    metrics,
    
    // 功能
    useComponentPerformance,
    useMemoryGuard,
    useRenderOptimization,
    useVirtualScroll,
    useLazyImages,
    useReactiveOptimization,
    updateMetrics
  }
}

// 导入必要的 Vue 函数
import { Ref, shallowReactive, readonly, computed } from 'vue'

export default usePerformanceOptimization
