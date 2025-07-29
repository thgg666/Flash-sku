import type { Directive, DirectiveBinding } from 'vue'

/**
 * 防抖指令配置
 */
interface DebounceOptions {
  delay?: number
  immediate?: boolean
  maxWait?: number
}

/**
 * 防抖指令绑定值类型
 */
type DebounceValue = {
  handler: Function
  options?: DebounceOptions
} | Function

/**
 * 防抖函数
 */
function debounce(
  func: Function,
  delay: number = 300,
  options: { immediate?: boolean; maxWait?: number } = {}
) {
  let timeoutId: number | null = null
  let maxTimeoutId: number | null = null
  let lastCallTime = 0
  
  const { immediate = false, maxWait } = options

  return function (this: any, ...args: any[]) {
    const context = this
    const now = Date.now()
    
    // 清除之前的定时器
    if (timeoutId) {
      clearTimeout(timeoutId)
    }
    
    // 如果设置了最大等待时间
    if (maxWait && !maxTimeoutId) {
      maxTimeoutId = setTimeout(() => {
        func.apply(context, args)
        maxTimeoutId = null
        lastCallTime = now
      }, maxWait)
    }
    
    // 立即执行模式
    if (immediate && now - lastCallTime > delay) {
      func.apply(context, args)
      lastCallTime = now
      return
    }
    
    // 延迟执行
    timeoutId = setTimeout(() => {
      func.apply(context, args)
      timeoutId = null
      lastCallTime = now
      
      if (maxTimeoutId) {
        clearTimeout(maxTimeoutId)
        maxTimeoutId = null
      }
    }, delay)
  }
}

/**
 * 防抖指令
 * 
 * 使用方法：
 * v-debounce="handler"
 * v-debounce="{ handler, options: { delay: 500 } }"
 * v-debounce:click="handler"
 * v-debounce:click.immediate="handler"
 */
export const vDebounce: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding<DebounceValue>) {
    const { value, arg = 'click', modifiers } = binding
    
    let handler: Function
    let options: DebounceOptions = {}
    
    // 解析绑定值
    if (typeof value === 'function') {
      handler = value
    } else if (value && typeof value.handler === 'function') {
      handler = value.handler
      options = value.options || {}
    } else {
      console.warn('v-debounce: 绑定值必须是函数或包含handler的对象')
      return
    }
    
    // 解析修饰符
    if (modifiers.immediate) {
      options.immediate = true
    }
    
    // 创建防抖函数
    const debouncedHandler = debounce(handler, options.delay, {
      immediate: options.immediate,
      maxWait: options.maxWait
    })
    
    // 绑定事件
    el.addEventListener(arg, debouncedHandler)
    
    // 保存到元素上，用于后续更新和卸载
    ;(el as any)._debounceHandler = debouncedHandler
    ;(el as any)._debounceEvent = arg
  },
  
  updated(el: HTMLElement, binding: DirectiveBinding<DebounceValue>) {
    // 如果绑定值改变，重新创建防抖函数
    const { value, arg = 'click', modifiers } = binding
    
    let handler: Function
    let options: DebounceOptions = {}
    
    if (typeof value === 'function') {
      handler = value
    } else if (value && typeof value.handler === 'function') {
      handler = value.handler
      options = value.options || {}
    } else {
      return
    }
    
    if (modifiers.immediate) {
      options.immediate = true
    }
    
    // 移除旧的事件监听器
    const oldHandler = (el as any)._debounceHandler
    const oldEvent = (el as any)._debounceEvent
    if (oldHandler && oldEvent) {
      el.removeEventListener(oldEvent, oldHandler)
    }
    
    // 创建新的防抖函数并绑定
    const debouncedHandler = debounce(handler, options.delay, {
      immediate: options.immediate,
      maxWait: options.maxWait
    })
    
    el.addEventListener(arg, debouncedHandler)
    ;(el as any)._debounceHandler = debouncedHandler
    ;(el as any)._debounceEvent = arg
  },
  
  unmounted(el: HTMLElement) {
    // 清理事件监听器
    const handler = (el as any)._debounceHandler
    const event = (el as any)._debounceEvent
    if (handler && event) {
      el.removeEventListener(event, handler)
    }
    delete (el as any)._debounceHandler
    delete (el as any)._debounceEvent
  }
}

/**
 * 节流指令
 */
function throttle(func: Function, delay: number = 300) {
  let lastCallTime = 0
  
  return function (this: any, ...args: any[]) {
    const now = Date.now()
    
    if (now - lastCallTime >= delay) {
      func.apply(this, args)
      lastCallTime = now
    }
  }
}

/**
 * 节流指令
 * 
 * 使用方法：
 * v-throttle="handler"
 * v-throttle:scroll="handler"
 */
export const vThrottle: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    const { value, arg = 'click' } = binding
    
    if (typeof value !== 'function') {
      console.warn('v-throttle: 绑定值必须是函数')
      return
    }
    
    const throttledHandler = throttle(value, 300)
    el.addEventListener(arg, throttledHandler)
    ;(el as any)._throttleHandler = throttledHandler
    ;(el as any)._throttleEvent = arg
  },
  
  updated(el: HTMLElement, binding: DirectiveBinding) {
    const { value, arg = 'click' } = binding
    
    if (typeof value !== 'function') {
      return
    }
    
    const oldHandler = (el as any)._throttleHandler
    const oldEvent = (el as any)._throttleEvent
    if (oldHandler && oldEvent) {
      el.removeEventListener(oldEvent, oldHandler)
    }
    
    const throttledHandler = throttle(value, 300)
    el.addEventListener(arg, throttledHandler)
    ;(el as any)._throttleHandler = throttledHandler
    ;(el as any)._throttleEvent = arg
  },
  
  unmounted(el: HTMLElement) {
    const handler = (el as any)._throttleHandler
    const event = (el as any)._throttleEvent
    if (handler && event) {
      el.removeEventListener(event, handler)
    }
    delete (el as any)._throttleHandler
    delete (el as any)._throttleEvent
  }
}

/**
 * 防重复点击指令 - 专门用于防止重复提交
 */
export const vPreventRepeat: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    const { value, arg = 'click' } = binding
    const delay = typeof binding.value === 'number' ? binding.value : 1000
    
    let isProcessing = false
    let originalHandler: Function | null = null
    
    // 如果绑定值是函数，使用它作为处理器
    if (typeof value === 'function') {
      originalHandler = value
    }
    
    const preventRepeatHandler = async function (this: any, event: Event) {
      if (isProcessing) {
        event.preventDefault()
        event.stopPropagation()
        return false
      }
      
      isProcessing = true
      el.classList.add('processing')
      el.setAttribute('disabled', 'true')
      
      try {
        if (originalHandler) {
          await originalHandler.call(this, event)
        }
      } finally {
        setTimeout(() => {
          isProcessing = false
          el.classList.remove('processing')
          el.removeAttribute('disabled')
        }, delay)
      }
    }
    
    el.addEventListener(arg, preventRepeatHandler)
    ;(el as any)._preventRepeatHandler = preventRepeatHandler
    ;(el as any)._preventRepeatEvent = arg
  },
  
  unmounted(el: HTMLElement) {
    const handler = (el as any)._preventRepeatHandler
    const event = (el as any)._preventRepeatEvent
    if (handler && event) {
      el.removeEventListener(event, handler)
    }
    delete (el as any)._preventRepeatHandler
    delete (el as any)._preventRepeatEvent
  }
}

// 导出所有指令
export default {
  debounce: vDebounce,
  throttle: vThrottle,
  preventRepeat: vPreventRepeat
}
