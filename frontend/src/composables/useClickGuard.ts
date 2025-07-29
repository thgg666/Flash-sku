import { ref, computed, onUnmounted, readonly } from 'vue'
import { ElMessage } from 'element-plus'

/**
 * 防重复点击配置
 */
interface UseClickGuardOptions {
  /** 冷却时间（毫秒） */
  cooldown?: number
  /** 最大并发数 */
  maxConcurrent?: number
  /** 是否显示提示信息 */
  showMessage?: boolean
  /** 自定义提示信息 */
  message?: string
  /** 防护键名 */
  key?: string
  /** 自动重置时间（毫秒） */
  autoReset?: number
}

/**
 * 防重复点击组合式函数
 */
export function useClickGuard(options: UseClickGuardOptions = {}) {
  const {
    cooldown = 1000,
    maxConcurrent = 1,
    showMessage = true,
    message = '操作过于频繁，请稍后再试',
    key = 'default',
    autoReset = 0
  } = options

  // 响应式状态
  const isProcessing = ref(false)
  const lastClickTime = ref(0)
  const concurrentCount = ref(0)
  const remainingCooldown = ref(0)

  // 定时器
  let cooldownTimer: number | null = null
  let autoResetTimer: number | null = null

  // 计算属性
  const canExecute = computed(() => {
    const now = Date.now()
    const timeSinceLastClick = now - lastClickTime.value
    
    return !isProcessing.value && 
           timeSinceLastClick >= cooldown && 
           concurrentCount.value < maxConcurrent
  })

  const isInCooldown = computed(() => {
    return remainingCooldown.value > 0
  })

  /**
   * 更新冷却时间
   */
  const updateCooldown = () => {
    const now = Date.now()
    const elapsed = now - lastClickTime.value
    remainingCooldown.value = Math.max(0, cooldown - elapsed)

    if (remainingCooldown.value > 0) {
      cooldownTimer = setTimeout(() => {
        updateCooldown()
      }, 100)
    } else if (cooldownTimer) {
      clearTimeout(cooldownTimer)
      cooldownTimer = null
    }
  }

  /**
   * 执行操作
   */
  const execute = async <T>(fn: () => Promise<T> | T): Promise<T | undefined> => {
    if (!canExecute.value) {
      if (showMessage) {
        if (isInCooldown.value) {
          ElMessage.warning(`${message}（${Math.ceil(remainingCooldown.value / 1000)}秒后可重试）`)
        } else {
          ElMessage.warning(message)
        }
      }
      return undefined
    }

    // 开始执行
    isProcessing.value = true
    concurrentCount.value++
    lastClickTime.value = Date.now()
    
    // 开始冷却倒计时
    updateCooldown()

    try {
      const result = await fn()
      return result
    } catch (error) {
      throw error
    } finally {
      // 结束执行
      concurrentCount.value = Math.max(0, concurrentCount.value - 1)
      isProcessing.value = concurrentCount.value > 0

      // 设置自动重置
      if (autoReset > 0) {
        if (autoResetTimer) {
          clearTimeout(autoResetTimer)
        }
        autoResetTimer = setTimeout(() => {
          reset()
        }, autoReset)
      }
    }
  }

  /**
   * 重置状态
   */
  const reset = () => {
    isProcessing.value = false
    concurrentCount.value = 0
    remainingCooldown.value = 0
    
    if (cooldownTimer) {
      clearTimeout(cooldownTimer)
      cooldownTimer = null
    }
    
    if (autoResetTimer) {
      clearTimeout(autoResetTimer)
      autoResetTimer = null
    }
  }

  /**
   * 强制设置冷却时间
   */
  const setCooldown = (time: number) => {
    lastClickTime.value = Date.now() - cooldown + time
    updateCooldown()
  }

  // 组件卸载时清理
  onUnmounted(() => {
    reset()
  })

  return {
    // 状态
    isProcessing: readonly(isProcessing),
    canExecute,
    isInCooldown,
    remainingCooldown: readonly(remainingCooldown),
    concurrentCount: readonly(concurrentCount),
    
    // 方法
    execute,
    reset,
    setCooldown
  }
}

/**
 * 秒杀专用防重复点击
 */
export function useSeckillClickGuard(activityId: number) {
  return useClickGuard({
    cooldown: 2000, // 秒杀冷却2秒
    maxConcurrent: 1,
    showMessage: true,
    message: '秒杀请求过于频繁，请稍后再试',
    key: `seckill_${activityId}`,
    autoReset: 10000 // 10秒后自动重置
  })
}

/**
 * 表单提交防重复点击
 */
export function useFormSubmitGuard() {
  return useClickGuard({
    cooldown: 3000, // 表单提交冷却3秒
    maxConcurrent: 1,
    showMessage: true,
    message: '表单提交中，请勿重复提交',
    key: 'form_submit'
  })
}

/**
 * API请求防重复点击
 */
export function useApiRequestGuard(apiKey: string) {
  return useClickGuard({
    cooldown: 1000, // API请求冷却1秒
    maxConcurrent: 3, // 允许3个并发请求
    showMessage: true,
    message: 'API请求过于频繁，请稍后再试',
    key: `api_${apiKey}`
  })
}
