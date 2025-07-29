import { ref, onMounted, onUnmounted } from 'vue'

/**
 * 桌面端交互增强组合式函数
 * 提供鼠标悬停、键盘导航、焦点管理等桌面端专用交互功能
 */
export function useDesktopInteraction() {
  const isDesktop = ref(false)
  const isKeyboardNavigation = ref(false)

  // 检测是否为桌面端
  const checkIsDesktop = () => {
    isDesktop.value = window.innerWidth >= 992 && !('ontouchstart' in window)
  }

  // 键盘导航检测
  const handleKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Tab') {
      isKeyboardNavigation.value = true
      document.body.classList.add('keyboard-navigation')
    }
  }

  const handleMouseDown = () => {
    isKeyboardNavigation.value = false
    document.body.classList.remove('keyboard-navigation')
  }

  // 窗口大小变化监听
  const handleResize = () => {
    checkIsDesktop()
  }

  onMounted(() => {
    checkIsDesktop()
    window.addEventListener('resize', handleResize)
    document.addEventListener('keydown', handleKeyDown)
    document.addEventListener('mousedown', handleMouseDown)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
    document.removeEventListener('keydown', handleKeyDown)
    document.removeEventListener('mousedown', handleMouseDown)
  })

  return {
    isDesktop,
    isKeyboardNavigation
  }
}

/**
 * 悬停效果组合式函数
 */
export function useHoverEffect() {
  const isHovered = ref(false)

  const handleMouseEnter = () => {
    isHovered.value = true
  }

  const handleMouseLeave = () => {
    isHovered.value = false
  }

  return {
    isHovered,
    handleMouseEnter,
    handleMouseLeave
  }
}

/**
 * 焦点管理组合式函数
 */
export function useFocusManagement() {
  const isFocused = ref(false)
  const focusableElements = ref<HTMLElement[]>([])

  // 获取可聚焦元素
  const getFocusableElements = (container: HTMLElement) => {
    const selectors = [
      'button:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      'textarea:not([disabled])',
      'a[href]',
      '[tabindex]:not([tabindex="-1"])'
    ].join(', ')

    return Array.from(container.querySelectorAll(selectors)) as HTMLElement[]
  }

  // 设置焦点陷阱
  const trapFocus = (container: HTMLElement) => {
    focusableElements.value = getFocusableElements(container)
    
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Tab') return

      const firstElement = focusableElements.value[0]
      const lastElement = focusableElements.value[focusableElements.value.length - 1]

      if (event.shiftKey) {
        if (document.activeElement === firstElement) {
          event.preventDefault()
          lastElement?.focus()
        }
      } else {
        if (document.activeElement === lastElement) {
          event.preventDefault()
          firstElement?.focus()
        }
      }
    }

    container.addEventListener('keydown', handleKeyDown)
    
    return () => {
      container.removeEventListener('keydown', handleKeyDown)
    }
  }

  // 焦点状态管理
  const handleFocus = () => {
    isFocused.value = true
  }

  const handleBlur = () => {
    isFocused.value = false
  }

  return {
    isFocused,
    focusableElements,
    getFocusableElements,
    trapFocus,
    handleFocus,
    handleBlur
  }
}

/**
 * 键盘快捷键组合式函数
 */
export function useKeyboardShortcuts() {
  const shortcuts = ref<Map<string, () => void>>(new Map())

  // 注册快捷键
  const registerShortcut = (key: string, callback: () => void) => {
    shortcuts.value.set(key, callback)
  }

  // 注销快捷键
  const unregisterShortcut = (key: string) => {
    shortcuts.value.delete(key)
  }

  // 处理键盘事件
  const handleKeyDown = (event: KeyboardEvent) => {
    const key = getKeyString(event)
    const callback = shortcuts.value.get(key)
    
    if (callback) {
      event.preventDefault()
      callback()
    }
  }

  // 获取按键字符串
  const getKeyString = (event: KeyboardEvent) => {
    const parts = []
    
    if (event.ctrlKey) parts.push('ctrl')
    if (event.altKey) parts.push('alt')
    if (event.shiftKey) parts.push('shift')
    if (event.metaKey) parts.push('meta')
    
    parts.push(event.key.toLowerCase())
    
    return parts.join('+')
  }

  onMounted(() => {
    document.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeyDown)
  })

  return {
    registerShortcut,
    unregisterShortcut
  }
}

/**
 * 右键菜单组合式函数
 */
export function useContextMenu() {
  const isVisible = ref(false)
  const position = ref({ x: 0, y: 0 })
  const menuItems = ref<Array<{
    label: string
    action: () => void
    disabled?: boolean
    divider?: boolean
    icon?: any
  }>>([])

  // 显示右键菜单
  const showMenu = (event: MouseEvent, items: typeof menuItems.value) => {
    event.preventDefault()
    
    position.value = {
      x: event.clientX,
      y: event.clientY
    }
    
    menuItems.value = items
    isVisible.value = true
    
    // 点击其他地方关闭菜单
    const handleClickOutside = () => {
      isVisible.value = false
      document.removeEventListener('click', handleClickOutside)
    }
    
    setTimeout(() => {
      document.addEventListener('click', handleClickOutside)
    }, 0)
  }

  // 隐藏菜单
  const hideMenu = () => {
    isVisible.value = false
  }

  // 执行菜单项操作
  const executeAction = (action: () => void) => {
    action()
    hideMenu()
  }

  return {
    isVisible,
    position,
    menuItems,
    showMenu,
    hideMenu,
    executeAction
  }
}

/**
 * 拖拽功能组合式函数
 */
export function useDragAndDrop() {
  const isDragging = ref(false)
  const dragData = ref<any>(null)

  // 开始拖拽
  const startDrag = (event: DragEvent, data: any) => {
    isDragging.value = true
    dragData.value = data
    
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move'
      event.dataTransfer.setData('text/plain', JSON.stringify(data))
    }
  }

  // 拖拽结束
  const endDrag = () => {
    isDragging.value = false
    dragData.value = null
  }

  // 处理放置
  const handleDrop = (event: DragEvent, callback: (data: any) => void) => {
    event.preventDefault()
    
    try {
      const data = JSON.parse(event.dataTransfer?.getData('text/plain') || '{}')
      callback(data)
    } catch (error) {
      console.error('拖拽数据解析失败:', error)
    }
    
    endDrag()
  }

  // 允许放置
  const allowDrop = (event: DragEvent) => {
    event.preventDefault()
  }

  return {
    isDragging,
    dragData,
    startDrag,
    endDrag,
    handleDrop,
    allowDrop
  }
}

/**
 * 滚动增强组合式函数
 */
export function useScrollEnhancement() {
  const isScrolling = ref(false)
  const scrollDirection = ref<'up' | 'down' | null>(null)
  const lastScrollY = ref(0)

  // 平滑滚动到元素
  const scrollToElement = (element: HTMLElement, options?: ScrollIntoViewOptions) => {
    element.scrollIntoView({
      behavior: 'smooth',
      block: 'center',
      ...options
    })
  }

  // 平滑滚动到顶部
  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    })
  }

  // 监听滚动
  const handleScroll = () => {
    const currentScrollY = window.scrollY
    
    isScrolling.value = true
    scrollDirection.value = currentScrollY > lastScrollY.value ? 'down' : 'up'
    lastScrollY.value = currentScrollY
    
    // 防抖处理
    clearTimeout(scrollTimeout)
    scrollTimeout = setTimeout(() => {
      isScrolling.value = false
    }, 150)
  }

  let scrollTimeout: number

  onMounted(() => {
    window.addEventListener('scroll', handleScroll, { passive: true })
  })

  onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll)
    clearTimeout(scrollTimeout)
  })

  return {
    isScrolling,
    scrollDirection,
    scrollToElement,
    scrollToTop
  }
}
