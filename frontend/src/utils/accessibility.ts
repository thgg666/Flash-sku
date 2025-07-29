/**
 * 可访问性工具函数
 * Accessibility utilities for better user experience
 */

/**
 * 键盘导航管理器
 */
export class KeyboardNavigationManager {
  private focusableElements: HTMLElement[] = []
  private currentIndex = -1
  private container: HTMLElement | null = null

  constructor(container?: HTMLElement) {
    this.container = container || document.body
    this.updateFocusableElements()
  }

  /**
   * 更新可聚焦元素列表
   */
  updateFocusableElements() {
    if (!this.container) return

    const selector = [
      'a[href]',
      'button:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      'textarea:not([disabled])',
      '[tabindex]:not([tabindex="-1"])',
      '[contenteditable="true"]'
    ].join(', ')

    this.focusableElements = Array.from(
      this.container.querySelectorAll(selector)
    ).filter(el => {
      const element = el as HTMLElement
      return element.offsetWidth > 0 && 
             element.offsetHeight > 0 && 
             !element.hidden &&
             getComputedStyle(element).visibility !== 'hidden'
    }) as HTMLElement[]
  }

  /**
   * 聚焦到下一个元素
   */
  focusNext() {
    this.updateFocusableElements()
    if (this.focusableElements.length === 0) return

    this.currentIndex = (this.currentIndex + 1) % this.focusableElements.length
    this.focusableElements[this.currentIndex].focus()
  }

  /**
   * 聚焦到上一个元素
   */
  focusPrevious() {
    this.updateFocusableElements()
    if (this.focusableElements.length === 0) return

    this.currentIndex = this.currentIndex <= 0 
      ? this.focusableElements.length - 1 
      : this.currentIndex - 1
    this.focusableElements[this.currentIndex].focus()
  }

  /**
   * 聚焦到第一个元素
   */
  focusFirst() {
    this.updateFocusableElements()
    if (this.focusableElements.length === 0) return

    this.currentIndex = 0
    this.focusableElements[this.currentIndex].focus()
  }

  /**
   * 聚焦到最后一个元素
   */
  focusLast() {
    this.updateFocusableElements()
    if (this.focusableElements.length === 0) return

    this.currentIndex = this.focusableElements.length - 1
    this.focusableElements[this.currentIndex].focus()
  }

  /**
   * 获取当前聚焦元素的索引
   */
  getCurrentIndex(): number {
    const activeElement = document.activeElement as HTMLElement
    return this.focusableElements.indexOf(activeElement)
  }
}

/**
 * 屏幕阅读器公告
 */
export class ScreenReaderAnnouncer {
  private announcer: HTMLElement

  constructor() {
    this.announcer = this.createAnnouncer()
  }

  private createAnnouncer(): HTMLElement {
    const announcer = document.createElement('div')
    announcer.setAttribute('aria-live', 'polite')
    announcer.setAttribute('aria-atomic', 'true')
    announcer.style.cssText = `
      position: absolute;
      left: -10000px;
      width: 1px;
      height: 1px;
      overflow: hidden;
    `
    document.body.appendChild(announcer)
    return announcer
  }

  /**
   * 发布公告
   */
  announce(message: string, priority: 'polite' | 'assertive' = 'polite') {
    this.announcer.setAttribute('aria-live', priority)
    this.announcer.textContent = message

    // 清除消息以便下次公告
    setTimeout(() => {
      this.announcer.textContent = ''
    }, 1000)
  }

  /**
   * 销毁公告器
   */
  destroy() {
    if (this.announcer.parentNode) {
      this.announcer.parentNode.removeChild(this.announcer)
    }
  }
}

/**
 * 焦点陷阱管理器
 */
export class FocusTrap {
  private element: HTMLElement
  private previousActiveElement: HTMLElement | null = null
  private keyboardManager: KeyboardNavigationManager

  constructor(element: HTMLElement) {
    this.element = element
    this.keyboardManager = new KeyboardNavigationManager(element)
  }

  /**
   * 激活焦点陷阱
   */
  activate() {
    this.previousActiveElement = document.activeElement as HTMLElement
    this.keyboardManager.focusFirst()
    document.addEventListener('keydown', this.handleKeyDown)
  }

  /**
   * 停用焦点陷阱
   */
  deactivate() {
    document.removeEventListener('keydown', this.handleKeyDown)
    if (this.previousActiveElement) {
      this.previousActiveElement.focus()
    }
  }

  private handleKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Tab') {
      event.preventDefault()
      if (event.shiftKey) {
        this.keyboardManager.focusPrevious()
      } else {
        this.keyboardManager.focusNext()
      }
    } else if (event.key === 'Escape') {
      this.deactivate()
    }
  }
}

/**
 * 高对比度模式检测
 */
export function detectHighContrastMode(): boolean {
  // 创建测试元素
  const testElement = document.createElement('div')
  testElement.style.cssText = `
    position: absolute;
    left: -9999px;
    background-color: rgb(31, 31, 31);
    color: rgb(255, 255, 255);
  `
  document.body.appendChild(testElement)

  // 检测是否在高对比度模式下
  const computedStyle = getComputedStyle(testElement)
  const isHighContrast = computedStyle.backgroundColor === computedStyle.color

  // 清理测试元素
  document.body.removeChild(testElement)

  return isHighContrast
}

/**
 * 减少动画偏好检测
 */
export function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/**
 * 颜色对比度计算
 */
export function calculateContrastRatio(color1: string, color2: string): number {
  const getLuminance = (color: string): number => {
    // 简化的亮度计算
    const rgb = color.match(/\d+/g)
    if (!rgb) return 0

    const [r, g, b] = rgb.map(Number)
    const [rs, gs, bs] = [r, g, b].map(c => {
      c = c / 255
      return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
    })

    return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs
  }

  const lum1 = getLuminance(color1)
  const lum2 = getLuminance(color2)
  const brightest = Math.max(lum1, lum2)
  const darkest = Math.min(lum1, lum2)

  return (brightest + 0.05) / (darkest + 0.05)
}

/**
 * ARIA属性管理器
 */
export class AriaManager {
  /**
   * 设置ARIA标签
   */
  static setLabel(element: HTMLElement, label: string) {
    element.setAttribute('aria-label', label)
  }

  /**
   * 设置ARIA描述
   */
  static setDescription(element: HTMLElement, description: string) {
    const descId = `desc-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    const descElement = document.createElement('div')
    descElement.id = descId
    descElement.textContent = description
    descElement.style.cssText = `
      position: absolute;
      left: -10000px;
      width: 1px;
      height: 1px;
      overflow: hidden;
    `
    document.body.appendChild(descElement)
    element.setAttribute('aria-describedby', descId)
  }

  /**
   * 设置ARIA状态
   */
  static setState(element: HTMLElement, state: string, value: string | boolean) {
    element.setAttribute(`aria-${state}`, value.toString())
  }

  /**
   * 设置ARIA属性
   */
  static setProperty(element: HTMLElement, property: string, value: string | boolean) {
    element.setAttribute(`aria-${property}`, value.toString())
  }
}

/**
 * 键盘快捷键管理器
 */
export class KeyboardShortcutManager {
  private shortcuts: Map<string, () => void> = new Map()

  constructor() {
    document.addEventListener('keydown', this.handleKeyDown)
  }

  /**
   * 注册快捷键
   */
  register(combination: string, callback: () => void) {
    this.shortcuts.set(combination.toLowerCase(), callback)
  }

  /**
   * 注销快捷键
   */
  unregister(combination: string) {
    this.shortcuts.delete(combination.toLowerCase())
  }

  private handleKeyDown = (event: KeyboardEvent) => {
    const combination = this.getCombination(event)
    const callback = this.shortcuts.get(combination)
    
    if (callback) {
      event.preventDefault()
      callback()
    }
  }

  private getCombination(event: KeyboardEvent): string {
    const parts: string[] = []
    
    if (event.ctrlKey) parts.push('ctrl')
    if (event.altKey) parts.push('alt')
    if (event.shiftKey) parts.push('shift')
    if (event.metaKey) parts.push('meta')
    
    parts.push(event.key.toLowerCase())
    
    return parts.join('+')
  }

  /**
   * 销毁管理器
   */
  destroy() {
    document.removeEventListener('keydown', this.handleKeyDown)
    this.shortcuts.clear()
  }
}

/**
 * 初始化可访问性功能
 */
export function initAccessibility() {
  // 检测高对比度模式
  if (detectHighContrastMode()) {
    document.body.classList.add('high-contrast')
  }

  // 检测减少动画偏好
  if (prefersReducedMotion()) {
    document.body.classList.add('reduce-motion')
  }

  // 添加跳转到主内容的链接
  addSkipToMainLink()

  // 初始化全局键盘导航
  initGlobalKeyboardNavigation()
}

/**
 * 添加跳转到主内容的链接
 */
function addSkipToMainLink() {
  const skipLink = document.createElement('a')
  skipLink.href = '#main-content'
  skipLink.textContent = '跳转到主内容'
  skipLink.className = 'skip-to-main'
  skipLink.style.cssText = `
    position: absolute;
    left: -9999px;
    z-index: 999999;
    padding: 8px 16px;
    background: #000;
    color: #fff;
    text-decoration: none;
    border-radius: 4px;
  `

  skipLink.addEventListener('focus', () => {
    skipLink.style.left = '10px'
    skipLink.style.top = '10px'
  })

  skipLink.addEventListener('blur', () => {
    skipLink.style.left = '-9999px'
  })

  document.body.insertBefore(skipLink, document.body.firstChild)
}

/**
 * 初始化全局键盘导航
 */
function initGlobalKeyboardNavigation() {
  const shortcutManager = new KeyboardShortcutManager()

  // Alt + M: 跳转到主菜单
  shortcutManager.register('alt+m', () => {
    const mainMenu = document.querySelector('[role="navigation"]') as HTMLElement
    if (mainMenu) {
      mainMenu.focus()
    }
  })

  // Alt + S: 跳转到搜索
  shortcutManager.register('alt+s', () => {
    const searchInput = document.querySelector('input[type="search"]') as HTMLElement
    if (searchInput) {
      searchInput.focus()
    }
  })

  // Alt + C: 跳转到主内容
  shortcutManager.register('alt+c', () => {
    const mainContent = document.querySelector('#main-content') as HTMLElement
    if (mainContent) {
      mainContent.focus()
    }
  })
}
