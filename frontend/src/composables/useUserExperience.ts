import { ref, nextTick, readonly } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'

/**
 * 用户体验增强配置
 */
interface UserExperienceOptions {
  /** 是否启用触觉反馈 */
  enableHaptic?: boolean
  /** 是否启用声音反馈 */
  enableSound?: boolean
  /** 是否启用动画 */
  enableAnimation?: boolean
  /** 是否启用通知 */
  enableNotification?: boolean
  /** 动画持续时间 */
  animationDuration?: number
}

/**
 * 动画状态
 */
interface AnimationState {
  showSuccess: boolean
  showFail: boolean
  showLoading: boolean
  showCountdown: boolean
  showParticles: boolean
}

/**
 * 用户体验增强组合式函数
 */
export function useUserExperience(options: UserExperienceOptions = {}) {
  const {
    enableHaptic = true,
    enableSound = true,
    enableAnimation = true,
    enableNotification = true,
    animationDuration = 3000
  } = options

  // 响应式状态
  const animationState = ref<AnimationState>({
    showSuccess: false,
    showFail: false,
    showLoading: false,
    showCountdown: false,
    showParticles: false
  })

  const loadingProgress = ref(0)
  const countdownValue = ref(0)
  const isExperienceActive = ref(false)

  // 音频对象
  let successAudio: HTMLAudioElement | null = null
  let failAudio: HTMLAudioElement | null = null
  let clickAudio: HTMLAudioElement | null = null

  /**
   * 初始化音频
   */
  const initAudio = () => {
    if (!enableSound) return

    try {
      // 创建音频对象（使用Web Audio API或简单的Audio对象）
      successAudio = new Audio()
      successAudio.preload = 'auto'
      // 这里可以设置成功音效的URL
      // successAudio.src = '/sounds/success.mp3'

      failAudio = new Audio()
      failAudio.preload = 'auto'
      // failAudio.src = '/sounds/fail.mp3'

      clickAudio = new Audio()
      clickAudio.preload = 'auto'
      // clickAudio.src = '/sounds/click.mp3'
    } catch (error) {
      console.warn('音频初始化失败:', error)
    }
  }

  /**
   * 播放音效
   */
  const playSound = (type: 'success' | 'fail' | 'click') => {
    if (!enableSound) return

    try {
      let audio: HTMLAudioElement | null = null
      switch (type) {
        case 'success':
          audio = successAudio
          break
        case 'fail':
          audio = failAudio
          break
        case 'click':
          audio = clickAudio
          break
      }

      if (audio) {
        audio.currentTime = 0
        audio.play().catch(console.warn)
      }
    } catch (error) {
      console.warn('音效播放失败:', error)
    }
  }

  /**
   * 触觉反馈
   */
  const hapticFeedback = (type: 'light' | 'medium' | 'heavy' = 'medium') => {
    if (!enableHaptic) return

    try {
      // 检查是否支持触觉反馈
      if ('vibrate' in navigator) {
        const patterns = {
          light: [50],
          medium: [100],
          heavy: [200]
        }
        navigator.vibrate(patterns[type])
      }

      // 检查是否支持Haptic Feedback API（iOS Safari）
      if ('hapticFeedback' in window) {
        const feedbackTypes = {
          light: 'impactLight',
          medium: 'impactMedium',
          heavy: 'impactHeavy'
        }
        ;(window as any).hapticFeedback(feedbackTypes[type])
      }
    } catch (error) {
      console.warn('触觉反馈失败:', error)
    }
  }

  /**
   * 显示成功体验
   */
  const showSuccessExperience = (message = '操作成功！') => {
    if (!enableAnimation) {
      ElMessage.success(message)
      return
    }

    isExperienceActive.value = true
    animationState.value.showSuccess = true
    animationState.value.showParticles = true

    // 播放成功音效
    playSound('success')

    // 触觉反馈
    hapticFeedback('medium')

    // 显示通知
    if (enableNotification) {
      ElNotification({
        title: '成功',
        message,
        type: 'success',
        duration: 3000
      })
    }

    // 自动隐藏
    setTimeout(() => {
      hideAllAnimations()
    }, animationDuration)
  }

  /**
   * 显示失败体验
   */
  const showFailExperience = (message = '操作失败') => {
    if (!enableAnimation) {
      ElMessage.error(message)
      return
    }

    isExperienceActive.value = true
    animationState.value.showFail = true

    // 播放失败音效
    playSound('fail')

    // 触觉反馈
    hapticFeedback('heavy')

    // 显示通知
    if (enableNotification) {
      ElMessage.error(message)
    }

    // 自动隐藏
    setTimeout(() => {
      hideAllAnimations()
    }, animationDuration)
  }

  /**
   * 显示加载体验
   */
  const showLoadingExperience = (title = '处理中...', message = '请稍候') => {
    if (!enableAnimation) {
      return
    }

    isExperienceActive.value = true
    animationState.value.showLoading = true
    loadingProgress.value = 0

    // 模拟进度
    const progressInterval = setInterval(() => {
      if (loadingProgress.value < 90) {
        loadingProgress.value += Math.random() * 10
      } else {
        clearInterval(progressInterval)
      }
    }, 200)

    return () => {
      clearInterval(progressInterval)
      loadingProgress.value = 100
      setTimeout(() => {
        animationState.value.showLoading = false
      }, 500)
    }
  }

  /**
   * 显示倒计时体验
   */
  const showCountdownExperience = (
    seconds: number,
    message = '倒计时中...'
  ): Promise<void> => {
    return new Promise((resolve) => {
      if (!enableAnimation) {
        resolve()
        return
      }

      isExperienceActive.value = true
      animationState.value.showCountdown = true
      countdownValue.value = seconds

      const countdownInterval = setInterval(() => {
        countdownValue.value--
        
        // 播放滴答声
        if (countdownValue.value > 0) {
          playSound('click')
          hapticFeedback('light')
        }

        if (countdownValue.value <= 0) {
          clearInterval(countdownInterval)
          animationState.value.showCountdown = false
          isExperienceActive.value = false
          resolve()
        }
      }, 1000)
    })
  }

  /**
   * 隐藏所有动画
   */
  const hideAllAnimations = () => {
    animationState.value = {
      showSuccess: false,
      showFail: false,
      showLoading: false,
      showCountdown: false,
      showParticles: false
    }
    isExperienceActive.value = false
    loadingProgress.value = 0
    countdownValue.value = 0
  }

  /**
   * 按钮点击体验增强
   */
  const enhanceButtonClick = (element: HTMLElement) => {
    if (!enableAnimation) return

    // 添加点击动画类
    element.classList.add('button-clicked')
    
    // 播放点击音效
    playSound('click')
    
    // 轻微触觉反馈
    hapticFeedback('light')

    // 移除动画类
    setTimeout(() => {
      element.classList.remove('button-clicked')
    }, 200)
  }

  /**
   * 页面震动效果
   */
  const shakeScreen = (duration = 500) => {
    if (!enableAnimation) return

    document.body.classList.add('screen-shake')
    hapticFeedback('heavy')

    setTimeout(() => {
      document.body.classList.remove('screen-shake')
    }, duration)
  }

  /**
   * 页面闪烁效果
   */
  const flashScreen = (color = '#ffffff', duration = 200) => {
    if (!enableAnimation) return

    const flash = document.createElement('div')
    flash.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100vw;
      height: 100vh;
      background: ${color};
      opacity: 0.8;
      z-index: 99999;
      pointer-events: none;
      animation: flash-fade ${duration}ms ease-out;
    `

    document.body.appendChild(flash)

    setTimeout(() => {
      document.body.removeChild(flash)
    }, duration)
  }

  // 初始化
  nextTick(() => {
    initAudio()
  })

  return {
    // 状态
    animationState: readonly(animationState),
    loadingProgress: readonly(loadingProgress),
    countdownValue: readonly(countdownValue),
    isExperienceActive: readonly(isExperienceActive),

    // 方法
    showSuccessExperience,
    showFailExperience,
    showLoadingExperience,
    showCountdownExperience,
    hideAllAnimations,
    enhanceButtonClick,
    shakeScreen,
    flashScreen,
    playSound,
    hapticFeedback
  }
}
