import { ref, onMounted, onUnmounted } from 'vue'

// 触摸事件类型
interface TouchPoint {
  x: number
  y: number
  timestamp: number
}

interface SwipeDirection {
  direction: 'left' | 'right' | 'up' | 'down'
  distance: number
  duration: number
  velocity: number
}

interface PinchGesture {
  scale: number
  center: TouchPoint
}

// 手势配置
interface GestureConfig {
  swipeThreshold?: number // 滑动阈值
  swipeTimeout?: number // 滑动超时时间
  tapTimeout?: number // 点击超时时间
  doubleTapTimeout?: number // 双击超时时间
  longPressTimeout?: number // 长按超时时间
  pinchThreshold?: number // 缩放阈值
}

/**
 * 移动端触摸手势处理
 */
export function useTouch(element: HTMLElement | null, config: GestureConfig = {}) {
  const {
    swipeThreshold = 50,
    swipeTimeout = 300,
    tapTimeout = 200,
    doubleTapTimeout = 300,
    longPressTimeout = 500,
    pinchThreshold = 0.1
  } = config

  // 状态
  const isTouch = ref(false)
  const touchStart = ref<TouchPoint | null>(null)
  const touchEnd = ref<TouchPoint | null>(null)
  const lastTap = ref<TouchPoint | null>(null)
  const longPressTimer = ref<number | null>(null)
  const initialDistance = ref(0)
  const currentScale = ref(1)

  // 事件回调
  const callbacks = {
    onTap: null as ((point: TouchPoint) => void) | null,
    onDoubleTap: null as ((point: TouchPoint) => void) | null,
    onLongPress: null as ((point: TouchPoint) => void) | null,
    onSwipe: null as ((swipe: SwipeDirection) => void) | null,
    onPinch: null as ((gesture: PinchGesture) => void) | null,
    onTouchStart: null as ((point: TouchPoint) => void) | null,
    onTouchMove: null as ((point: TouchPoint) => void) | null,
    onTouchEnd: null as ((point: TouchPoint) => void) | null,
  }

  // 获取触摸点坐标
  const getTouchPoint = (event: TouchEvent): TouchPoint => {
    const touch = event.touches[0] || event.changedTouches[0]
    return {
      x: touch.clientX,
      y: touch.clientY,
      timestamp: Date.now()
    }
  }

  // 计算两点距离
  const getDistance = (point1: TouchPoint, point2: TouchPoint): number => {
    const dx = point2.x - point1.x
    const dy = point2.y - point1.y
    return Math.sqrt(dx * dx + dy * dy)
  }

  // 计算两个触摸点之间的距离（用于缩放）
  const getTouchDistance = (event: TouchEvent): number => {
    if (event.touches.length < 2) return 0
    const touch1 = event.touches[0]
    const touch2 = event.touches[1]
    const dx = touch2.clientX - touch1.clientX
    const dy = touch2.clientY - touch1.clientY
    return Math.sqrt(dx * dx + dy * dy)
  }

  // 获取两个触摸点的中心点
  const getTouchCenter = (event: TouchEvent): TouchPoint => {
    if (event.touches.length < 2) return getTouchPoint(event)
    const touch1 = event.touches[0]
    const touch2 = event.touches[1]
    return {
      x: (touch1.clientX + touch2.clientX) / 2,
      y: (touch1.clientY + touch2.clientY) / 2,
      timestamp: Date.now()
    }
  }

  // 判断滑动方向
  const getSwipeDirection = (start: TouchPoint, end: TouchPoint): SwipeDirection | null => {
    const dx = end.x - start.x
    const dy = end.y - start.y
    const distance = getDistance(start, end)
    const duration = end.timestamp - start.timestamp
    
    if (distance < swipeThreshold || duration > swipeTimeout) {
      return null
    }

    const velocity = distance / duration
    const absDx = Math.abs(dx)
    const absDy = Math.abs(dy)

    let direction: 'left' | 'right' | 'up' | 'down'
    
    if (absDx > absDy) {
      direction = dx > 0 ? 'right' : 'left'
    } else {
      direction = dy > 0 ? 'down' : 'up'
    }

    return {
      direction,
      distance,
      duration,
      velocity
    }
  }

  // 处理触摸开始
  const handleTouchStart = (event: TouchEvent) => {
    event.preventDefault()
    isTouch.value = true
    
    const point = getTouchPoint(event)
    touchStart.value = point
    
    // 多点触摸（缩放）
    if (event.touches.length === 2) {
      initialDistance.value = getTouchDistance(event)
    } else {
      // 长按检测
      longPressTimer.value = window.setTimeout(() => {
        if (touchStart.value && callbacks.onLongPress) {
          callbacks.onLongPress(touchStart.value)
        }
      }, longPressTimeout)
    }
    
    callbacks.onTouchStart?.(point)
  }

  // 处理触摸移动
  const handleTouchMove = (event: TouchEvent) => {
    event.preventDefault()
    
    const point = getTouchPoint(event)
    
    // 多点触摸（缩放）
    if (event.touches.length === 2 && initialDistance.value > 0) {
      const currentDistance = getTouchDistance(event)
      const scale = currentDistance / initialDistance.value
      
      if (Math.abs(scale - currentScale.value) > pinchThreshold) {
        currentScale.value = scale
        const center = getTouchCenter(event)
        
        if (callbacks.onPinch) {
          callbacks.onPinch({ scale, center })
        }
      }
    }
    
    // 清除长按定时器
    if (longPressTimer.value) {
      clearTimeout(longPressTimer.value)
      longPressTimer.value = null
    }
    
    callbacks.onTouchMove?.(point)
  }

  // 处理触摸结束
  const handleTouchEnd = (event: TouchEvent) => {
    event.preventDefault()
    isTouch.value = false
    
    const point = getTouchPoint(event)
    touchEnd.value = point
    
    // 清除长按定时器
    if (longPressTimer.value) {
      clearTimeout(longPressTimer.value)
      longPressTimer.value = null
    }
    
    // 重置缩放状态
    if (event.touches.length === 0) {
      initialDistance.value = 0
      currentScale.value = 1
    }
    
    // 处理滑动
    if (touchStart.value && touchEnd.value) {
      const swipe = getSwipeDirection(touchStart.value, touchEnd.value)
      if (swipe && callbacks.onSwipe) {
        callbacks.onSwipe(swipe)
        return
      }
    }
    
    // 处理点击和双击
    if (touchStart.value && touchEnd.value) {
      const distance = getDistance(touchStart.value, touchEnd.value)
      const duration = touchEnd.value.timestamp - touchStart.value.timestamp
      
      // 判断是否为点击
      if (distance < swipeThreshold && duration < tapTimeout) {
        // 检查双击
        if (lastTap.value) {
          const timeSinceLastTap = touchEnd.value.timestamp - lastTap.value.timestamp
          const distanceFromLastTap = getDistance(lastTap.value, touchEnd.value)
          
          if (timeSinceLastTap < doubleTapTimeout && distanceFromLastTap < swipeThreshold) {
            // 双击
            if (callbacks.onDoubleTap) {
              callbacks.onDoubleTap(touchEnd.value)
            }
            lastTap.value = null
            return
          }
        }
        
        // 单击
        lastTap.value = touchEnd.value
        setTimeout(() => {
          if (lastTap.value === touchEnd.value && callbacks.onTap && touchEnd.value) {
            callbacks.onTap(touchEnd.value)
          }
        }, doubleTapTimeout)
      }
    }
    
    callbacks.onTouchEnd?.(point)
  }

  // 绑定事件
  const bindEvents = () => {
    if (!element) return
    
    element.addEventListener('touchstart', handleTouchStart, { passive: false })
    element.addEventListener('touchmove', handleTouchMove, { passive: false })
    element.addEventListener('touchend', handleTouchEnd, { passive: false })
    element.addEventListener('touchcancel', handleTouchEnd, { passive: false })
  }

  // 解绑事件
  const unbindEvents = () => {
    if (!element) return
    
    element.removeEventListener('touchstart', handleTouchStart)
    element.removeEventListener('touchmove', handleTouchMove)
    element.removeEventListener('touchend', handleTouchEnd)
    element.removeEventListener('touchcancel', handleTouchEnd)
  }

  // 设置回调函数
  const onTap = (callback: (point: TouchPoint) => void) => {
    callbacks.onTap = callback
  }

  const onDoubleTap = (callback: (point: TouchPoint) => void) => {
    callbacks.onDoubleTap = callback
  }

  const onLongPress = (callback: (point: TouchPoint) => void) => {
    callbacks.onLongPress = callback
  }

  const onSwipe = (callback: (swipe: SwipeDirection) => void) => {
    callbacks.onSwipe = callback
  }

  const onPinch = (callback: (gesture: PinchGesture) => void) => {
    callbacks.onPinch = callback
  }

  const onTouchStart = (callback: (point: TouchPoint) => void) => {
    callbacks.onTouchStart = callback
  }

  const onTouchMove = (callback: (point: TouchPoint) => void) => {
    callbacks.onTouchMove = callback
  }

  const onTouchEnd = (callback: (point: TouchPoint) => void) => {
    callbacks.onTouchEnd = callback
  }

  // 组件挂载时绑定事件
  onMounted(() => {
    bindEvents()
  })

  // 组件卸载时解绑事件
  onUnmounted(() => {
    unbindEvents()
    if (longPressTimer.value) {
      clearTimeout(longPressTimer.value)
    }
  })

  return {
    // 状态
    isTouch,
    
    // 方法
    onTap,
    onDoubleTap,
    onLongPress,
    onSwipe,
    onPinch,
    onTouchStart,
    onTouchMove,
    onTouchEnd,
    bindEvents,
    unbindEvents,
  }
}

/**
 * 移动端设备检测
 */
export function useMobileDetection() {
  const isMobile = ref(false)
  const isTablet = ref(false)
  const isDesktop = ref(false)
  const isTouchDevice = ref(false)
  const orientation = ref<'portrait' | 'landscape'>('portrait')

  const checkDevice = () => {
    const userAgent = navigator.userAgent.toLowerCase()
    const width = window.innerWidth
    const height = window.innerHeight
    
    // 检测触摸设备
    isTouchDevice.value = 'ontouchstart' in window || navigator.maxTouchPoints > 0
    
    // 检测设备类型
    isMobile.value = width <= 768 || /mobile|android|iphone|ipod|blackberry|iemobile|opera mini/.test(userAgent)
    isTablet.value = (width > 768 && width <= 1024) || /ipad|tablet/.test(userAgent)
    isDesktop.value = width > 1024 && !isTouchDevice.value
    
    // 检测屏幕方向
    orientation.value = width > height ? 'landscape' : 'portrait'
  }

  // 监听窗口大小变化
  const handleResize = () => {
    checkDevice()
  }

  // 监听屏幕方向变化
  const handleOrientationChange = () => {
    setTimeout(checkDevice, 100) // 延迟检测，等待屏幕旋转完成
  }

  onMounted(() => {
    checkDevice()
    window.addEventListener('resize', handleResize)
    window.addEventListener('orientationchange', handleOrientationChange)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
    window.removeEventListener('orientationchange', handleOrientationChange)
  })

  return {
    isMobile,
    isTablet,
    isDesktop,
    isTouchDevice,
    orientation,
  }
}
