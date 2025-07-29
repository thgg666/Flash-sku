/**
 * 点击防护工具类
 * 用于防止重复点击、频繁操作等
 */

/**
 * 点击防护配置
 */
interface ClickGuardOptions {
  /** 冷却时间（毫秒） */
  cooldown?: number
  /** 最大并发数 */
  maxConcurrent?: number
  /** 是否显示提示信息 */
  showMessage?: boolean
  /** 自定义提示信息 */
  message?: string
  /** 是否使用全局防护 */
  global?: boolean
  /** 防护键名 */
  key?: string
}

/**
 * 防护状态
 */
interface GuardState {
  lastClickTime: number
  concurrentCount: number
  isProcessing: boolean
}

/**
 * 点击防护管理器
 */
class ClickGuardManager {
  private guards = new Map<string, GuardState>()
  private globalGuard: GuardState = {
    lastClickTime: 0,
    concurrentCount: 0,
    isProcessing: false
  }

  /**
   * 检查是否可以执行操作
   */
  canExecute(key: string, options: ClickGuardOptions = {}): boolean {
    const {
      cooldown = 1000,
      maxConcurrent = 1,
      global = false
    } = options

    const guard = global ? this.globalGuard : this.getGuard(key)
    const now = Date.now()

    // 检查冷却时间
    if (now - guard.lastClickTime < cooldown) {
      return false
    }

    // 检查并发数
    if (guard.concurrentCount >= maxConcurrent) {
      return false
    }

    return true
  }

  /**
   * 开始执行操作
   */
  startExecution(key: string, options: ClickGuardOptions = {}): void {
    const { global = false } = options
    const guard = global ? this.globalGuard : this.getGuard(key)
    
    guard.lastClickTime = Date.now()
    guard.concurrentCount++
    guard.isProcessing = true
  }

  /**
   * 结束执行操作
   */
  endExecution(key: string, options: ClickGuardOptions = {}): void {
    const { global = false } = options
    const guard = global ? this.globalGuard : this.getGuard(key)
    
    guard.concurrentCount = Math.max(0, guard.concurrentCount - 1)
    guard.isProcessing = guard.concurrentCount > 0
  }

  /**
   * 获取防护状态
   */
  private getGuard(key: string): GuardState {
    if (!this.guards.has(key)) {
      this.guards.set(key, {
        lastClickTime: 0,
        concurrentCount: 0,
        isProcessing: false
      })
    }
    return this.guards.get(key)!
  }

  /**
   * 清除防护状态
   */
  clear(key?: string): void {
    if (key) {
      this.guards.delete(key)
    } else {
      this.guards.clear()
      this.globalGuard = {
        lastClickTime: 0,
        concurrentCount: 0,
        isProcessing: false
      }
    }
  }

  /**
   * 获取防护信息
   */
  getGuardInfo(key: string, global = false): GuardState {
    const guard = global ? this.globalGuard : this.getGuard(key)
    return { ...guard }
  }
}

// 全局防护管理器实例
const guardManager = new ClickGuardManager()

/**
 * 防重复点击装饰器
 */
export function preventRepeatClick(options: ClickGuardOptions = {}) {
  return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    const originalMethod = descriptor.value
    const key = options.key || `${target.constructor.name}.${propertyKey}`

    descriptor.value = async function (...args: any[]) {
      if (!guardManager.canExecute(key, options)) {
        if (options.showMessage !== false) {
          const message = options.message || '操作过于频繁，请稍后再试'
          // 这里可以集成消息提示组件
          console.warn(message)
        }
        return
      }

      guardManager.startExecution(key, options)

      try {
        const result = await originalMethod.apply(this, args)
        return result
      } finally {
        guardManager.endExecution(key, options)
      }
    }

    return descriptor
  }
}

/**
 * 防重复点击函数包装器
 */
export function withClickGuard<T extends (...args: any[]) => any>(
  fn: T,
  options: ClickGuardOptions = {}
): T {
  const key = options.key || fn.name || 'anonymous'

  return (async (...args: any[]) => {
    if (!guardManager.canExecute(key, options)) {
      if (options.showMessage !== false) {
        const message = options.message || '操作过于频繁，请稍后再试'
        console.warn(message)
      }
      return
    }

    guardManager.startExecution(key, options)

    try {
      const result = await fn(...args)
      return result
    } finally {
      guardManager.endExecution(key, options)
    }
  }) as T
}

/**
 * 秒杀专用防护
 */
export function createSeckillGuard(activityId: number, options: ClickGuardOptions = {}) {
  const key = `seckill_${activityId}`
  
  return {
    /**
     * 检查是否可以参与秒杀
     */
    canParticipate(): boolean {
      return guardManager.canExecute(key, {
        cooldown: 2000, // 秒杀冷却时间2秒
        maxConcurrent: 1,
        ...options
      })
    },

    /**
     * 开始秒杀
     */
    startSeckill(): void {
      guardManager.startExecution(key, options)
    },

    /**
     * 结束秒杀
     */
    endSeckill(): void {
      guardManager.endExecution(key, options)
    },

    /**
     * 获取秒杀状态
     */
    getStatus(): GuardState {
      return guardManager.getGuardInfo(key)
    },

    /**
     * 清除秒杀防护
     */
    clear(): void {
      guardManager.clear(key)
    }
  }
}

/**
 * 全局防护函数
 */
export const clickGuard = {
  /**
   * 检查是否可以执行
   */
  canExecute: (key: string, options?: ClickGuardOptions) => 
    guardManager.canExecute(key, options),

  /**
   * 开始执行
   */
  start: (key: string, options?: ClickGuardOptions) => 
    guardManager.startExecution(key, options),

  /**
   * 结束执行
   */
  end: (key: string, options?: ClickGuardOptions) => 
    guardManager.endExecution(key, options),

  /**
   * 清除防护
   */
  clear: (key?: string) => 
    guardManager.clear(key),

  /**
   * 获取状态
   */
  getStatus: (key: string, global?: boolean) => 
    guardManager.getGuardInfo(key, global),

  /**
   * 包装函数
   */
  wrap: withClickGuard,

  /**
   * 创建秒杀防护
   */
  createSeckillGuard
}

export default clickGuard
