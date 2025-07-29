/**
 * 懒加载指令
 * 用于图片、组件等资源的懒加载
 */

import type { Directive, DirectiveBinding } from 'vue'

// 懒加载配置接口
interface LazyLoadOptions {
  loading?: string // 加载中的占位图
  error?: string   // 加载失败的占位图
  threshold?: number // 触发加载的阈值
  rootMargin?: string // 根边距
  delay?: number   // 延迟加载时间
}

// 默认配置
const defaultOptions: LazyLoadOptions = {
  loading: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjBmMGYwIi8+CiAgPHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkxvYWRpbmcuLi48L3RleHQ+Cjwvc3ZnPg==',
  error: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjVmNWY1Ii8+CiAgPHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxNCIgZmlsbD0iI2NjYyIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkVycm9yPC90ZXh0Pgo8L3N2Zz4=',
  threshold: 0.1,
  rootMargin: '50px',
  delay: 0
}

// 观察器实例
let observer: IntersectionObserver | null = null

// 待加载元素的映射
const elementMap = new WeakMap<Element, LazyLoadOptions & { src: string }>()

/**
 * 创建交叉观察器
 */
function createObserver(): IntersectionObserver {
  return new IntersectionObserver(
    (entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const element = entry.target
          const config = elementMap.get(element)
          
          if (config) {
            loadElement(element as HTMLElement, config)
            observer?.unobserve(element)
            elementMap.delete(element)
          }
        }
      })
    },
    {
      rootMargin: defaultOptions.rootMargin,
      threshold: defaultOptions.threshold
    }
  )
}

/**
 * 加载元素
 */
function loadElement(element: HTMLElement, config: LazyLoadOptions & { src: string }) {
  const { src, loading, error, delay } = config

  // 延迟加载
  const loadFn = () => {
    if (element.tagName === 'IMG') {
      loadImage(element as HTMLImageElement, src, error || defaultOptions.error!)
    } else {
      // 对于其他元素，设置背景图片或其他属性
      loadOtherElement(element, src, error || defaultOptions.error!)
    }
  }

  if (delay && delay > 0) {
    setTimeout(loadFn, delay)
  } else {
    loadFn()
  }
}

/**
 * 加载图片
 */
function loadImage(img: HTMLImageElement, src: string, errorSrc: string) {
  // 添加加载状态类
  img.classList.add('lazy-loading')
  
  const image = new Image()
  
  image.onload = () => {
    img.src = src
    img.classList.remove('lazy-loading')
    img.classList.add('lazy-loaded')
    
    // 触发自定义事件
    img.dispatchEvent(new CustomEvent('lazy-loaded', { detail: { src } }))
  }
  
  image.onerror = () => {
    img.src = errorSrc
    img.classList.remove('lazy-loading')
    img.classList.add('lazy-error')
    
    // 触发自定义事件
    img.dispatchEvent(new CustomEvent('lazy-error', { detail: { src } }))
  }
  
  image.src = src
}

/**
 * 加载其他元素
 */
function loadOtherElement(element: HTMLElement, src: string, errorSrc: string) {
  element.classList.add('lazy-loading')
  
  // 对于背景图片
  if (element.dataset.background === 'true') {
    const image = new Image()
    
    image.onload = () => {
      element.style.backgroundImage = `url(${src})`
      element.classList.remove('lazy-loading')
      element.classList.add('lazy-loaded')
      element.dispatchEvent(new CustomEvent('lazy-loaded', { detail: { src } }))
    }
    
    image.onerror = () => {
      element.style.backgroundImage = `url(${errorSrc})`
      element.classList.remove('lazy-loading')
      element.classList.add('lazy-error')
      element.dispatchEvent(new CustomEvent('lazy-error', { detail: { src } }))
    }
    
    image.src = src
  } else {
    // 对于其他类型的元素，直接设置属性
    element.setAttribute('src', src)
    element.classList.remove('lazy-loading')
    element.classList.add('lazy-loaded')
  }
}

/**
 * 懒加载指令
 */
export const vLazyLoad: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    // 检查浏览器支持
    if (!('IntersectionObserver' in window)) {
      // 不支持的浏览器直接加载
      const src = binding.value?.src || binding.value
      if (typeof src === 'string') {
        if (el.tagName === 'IMG') {
          (el as HTMLImageElement).src = src
        } else {
          el.style.backgroundImage = `url(${src})`
        }
      }
      return
    }

    // 创建观察器
    if (!observer) {
      observer = createObserver()
    }

    // 解析配置
    let config: LazyLoadOptions & { src: string }
    
    if (typeof binding.value === 'string') {
      config = { ...defaultOptions, src: binding.value }
    } else {
      config = { ...defaultOptions, ...binding.value }
    }

    // 设置加载中的占位图
    if (el.tagName === 'IMG') {
      const img = el as HTMLImageElement
      if (!img.src && config.loading) {
        img.src = config.loading
      }
    } else if (el.dataset.background === 'true' && config.loading) {
      el.style.backgroundImage = `url(${config.loading})`
    }

    // 添加初始样式类
    el.classList.add('lazy-element')

    // 保存配置
    elementMap.set(el, config)

    // 开始观察
    observer.observe(el)
  },

  updated(el: HTMLElement, binding: DirectiveBinding) {
    // 如果src发生变化，重新设置
    const oldConfig = elementMap.get(el)
    const newSrc = binding.value?.src || binding.value

    if (oldConfig && oldConfig.src !== newSrc) {
      // 停止观察旧元素
      observer?.unobserve(el)
      elementMap.delete(el)

      // 重新挂载
      this.mounted?.(el, binding, null as any, null as any)
    }
  },

  unmounted(el: HTMLElement) {
    // 停止观察
    observer?.unobserve(el)
    elementMap.delete(el)
  }
}

/**
 * 预加载图片
 */
export function preloadImage(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve()
    image.onerror = reject
    image.src = src
  })
}

/**
 * 预加载多个图片
 */
export function preloadImages(srcs: string[]): Promise<void[]> {
  return Promise.all(srcs.map(src => preloadImage(src)))
}

/**
 * 懒加载组件工厂
 */
export function createLazyComponent(importFn: () => Promise<any>, options?: {
  loading?: any
  error?: any
  delay?: number
  timeout?: number
}) {
  return {
    component: importFn,
    loading: options?.loading,
    error: options?.error,
    delay: options?.delay || 0,
    timeout: options?.timeout || 30000
  }
}

/**
 * 图片懒加载组件
 */
export const LazyImage = {
  name: 'LazyImage',
  props: {
    src: {
      type: String,
      required: true
    },
    alt: {
      type: String,
      default: ''
    },
    loading: {
      type: String,
      default: defaultOptions.loading
    },
    error: {
      type: String,
      default: defaultOptions.error
    },
    threshold: {
      type: Number,
      default: defaultOptions.threshold
    },
    rootMargin: {
      type: String,
      default: defaultOptions.rootMargin
    },
    delay: {
      type: Number,
      default: defaultOptions.delay
    }
  },
  template: `
    <img
      v-lazy-load="{
        src,
        loading,
        error,
        threshold,
        rootMargin,
        delay
      }"
      :alt="alt"
      class="lazy-image"
      @lazy-loaded="$emit('loaded', $event)"
      @lazy-error="$emit('error', $event)"
    />
  `,
  emits: ['loaded', 'error']
}

/**
 * 销毁懒加载观察器
 */
export function destroyLazyLoad() {
  if (observer) {
    observer.disconnect()
    observer = null
  }
  elementMap.clear?.()
}
