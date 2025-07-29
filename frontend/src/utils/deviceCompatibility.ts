/**
 * 设备兼容性检测和测试工具
 */

// 设备类型定义
export interface DeviceInfo {
  type: 'mobile' | 'tablet' | 'desktop'
  width: number
  height: number
  pixelRatio: number
  orientation: 'portrait' | 'landscape'
  touchSupport: boolean
  userAgent: string
  browser: string
  os: string
}

// 断点定义
export const BREAKPOINTS = {
  xs: 480,
  sm: 768,
  md: 992,
  lg: 1200,
  xl: 1920
} as const

// 常见设备分辨率
export const COMMON_DEVICES = {
  mobile: [
    { name: 'iPhone SE', width: 375, height: 667, pixelRatio: 2 },
    { name: 'iPhone 12', width: 390, height: 844, pixelRatio: 3 },
    { name: 'iPhone 12 Pro Max', width: 428, height: 926, pixelRatio: 3 },
    { name: 'Samsung Galaxy S21', width: 384, height: 854, pixelRatio: 2.75 },
    { name: 'Google Pixel 5', width: 393, height: 851, pixelRatio: 2.75 }
  ],
  tablet: [
    { name: 'iPad', width: 768, height: 1024, pixelRatio: 2 },
    { name: 'iPad Pro 11"', width: 834, height: 1194, pixelRatio: 2 },
    { name: 'iPad Pro 12.9"', width: 1024, height: 1366, pixelRatio: 2 },
    { name: 'Samsung Galaxy Tab S7', width: 753, height: 1037, pixelRatio: 2.4 },
    { name: 'Surface Pro 7', width: 912, height: 1368, pixelRatio: 2 }
  ],
  desktop: [
    { name: '13" Laptop', width: 1280, height: 800, pixelRatio: 1 },
    { name: '15" Laptop', width: 1366, height: 768, pixelRatio: 1 },
    { name: '24" Monitor', width: 1920, height: 1080, pixelRatio: 1 },
    { name: '27" Monitor', width: 2560, height: 1440, pixelRatio: 1 },
    { name: '4K Monitor', width: 3840, height: 2160, pixelRatio: 1 }
  ]
} as const

/**
 * 获取当前设备信息
 */
export function getCurrentDeviceInfo(): DeviceInfo {
  const width = window.innerWidth
  const height = window.innerHeight
  const pixelRatio = window.devicePixelRatio || 1
  const userAgent = navigator.userAgent
  
  // 检测设备类型
  let type: DeviceInfo['type'] = 'desktop'
  if (width < BREAKPOINTS.sm) {
    type = 'mobile'
  } else if (width < BREAKPOINTS.lg) {
    type = 'tablet'
  }
  
  // 检测方向
  const orientation: DeviceInfo['orientation'] = width > height ? 'landscape' : 'portrait'
  
  // 检测触摸支持
  const touchSupport = 'ontouchstart' in window || navigator.maxTouchPoints > 0
  
  // 检测浏览器
  const browser = getBrowserName(userAgent)
  
  // 检测操作系统
  const os = getOperatingSystem(userAgent)
  
  return {
    type,
    width,
    height,
    pixelRatio,
    orientation,
    touchSupport,
    userAgent,
    browser,
    os
  }
}

/**
 * 获取浏览器名称
 */
function getBrowserName(userAgent: string): string {
  if (userAgent.includes('Chrome')) return 'Chrome'
  if (userAgent.includes('Firefox')) return 'Firefox'
  if (userAgent.includes('Safari') && !userAgent.includes('Chrome')) return 'Safari'
  if (userAgent.includes('Edge')) return 'Edge'
  if (userAgent.includes('Opera')) return 'Opera'
  return 'Unknown'
}

/**
 * 获取操作系统
 */
function getOperatingSystem(userAgent: string): string {
  if (userAgent.includes('Windows')) return 'Windows'
  if (userAgent.includes('Mac')) return 'macOS'
  if (userAgent.includes('Linux')) return 'Linux'
  if (userAgent.includes('Android')) return 'Android'
  if (userAgent.includes('iOS')) return 'iOS'
  return 'Unknown'
}

/**
 * 检查是否为移动设备
 */
export function isMobileDevice(): boolean {
  return getCurrentDeviceInfo().type === 'mobile'
}

/**
 * 检查是否为平板设备
 */
export function isTabletDevice(): boolean {
  return getCurrentDeviceInfo().type === 'tablet'
}

/**
 * 检查是否为桌面设备
 */
export function isDesktopDevice(): boolean {
  return getCurrentDeviceInfo().type === 'desktop'
}

/**
 * 检查是否支持触摸
 */
export function isTouchDevice(): boolean {
  return getCurrentDeviceInfo().touchSupport
}

/**
 * 获取当前断点
 */
export function getCurrentBreakpoint(): keyof typeof BREAKPOINTS {
  const width = window.innerWidth
  
  if (width < BREAKPOINTS.xs) return 'xs'
  if (width < BREAKPOINTS.sm) return 'sm'
  if (width < BREAKPOINTS.md) return 'md'
  if (width < BREAKPOINTS.lg) return 'lg'
  return 'xl'
}

/**
 * 检查是否匹配指定断点
 */
export function matchesBreakpoint(breakpoint: keyof typeof BREAKPOINTS, direction: 'up' | 'down' = 'up'): boolean {
  const width = window.innerWidth
  const breakpointValue = BREAKPOINTS[breakpoint]
  
  return direction === 'up' ? width >= breakpointValue : width < breakpointValue
}

/**
 * 设备兼容性测试类
 */
export class DeviceCompatibilityTester {
  private testResults: Map<string, boolean> = new Map()
  
  /**
   * 运行所有兼容性测试
   */
  async runAllTests(): Promise<Map<string, boolean>> {
    this.testResults.clear()
    
    // 基础功能测试
    this.testBasicFeatures()
    
    // CSS功能测试
    this.testCSSFeatures()
    
    // JavaScript功能测试
    this.testJavaScriptFeatures()
    
    // 性能测试
    await this.testPerformance()
    
    return this.testResults
  }
  
  /**
   * 测试基础功能
   */
  private testBasicFeatures(): void {
    // 测试视口大小
    this.testResults.set('viewport', window.innerWidth > 0 && window.innerHeight > 0)
    
    // 测试设备像素比
    this.testResults.set('pixelRatio', window.devicePixelRatio >= 1)
    
    // 测试触摸支持
    this.testResults.set('touchSupport', 'ontouchstart' in window)
    
    // 测试本地存储
    this.testResults.set('localStorage', this.testLocalStorage())
    
    // 测试会话存储
    this.testResults.set('sessionStorage', this.testSessionStorage())
  }
  
  /**
   * 测试CSS功能
   */
  private testCSSFeatures(): void {
    const testElement = document.createElement('div')
    document.body.appendChild(testElement)
    
    try {
      // 测试Flexbox
      testElement.style.display = 'flex'
      this.testResults.set('flexbox', testElement.style.display === 'flex')
      
      // 测试Grid
      testElement.style.display = 'grid'
      this.testResults.set('grid', testElement.style.display === 'grid')
      
      // 测试CSS变量
      testElement.style.setProperty('--test-var', 'test')
      this.testResults.set('cssVariables', testElement.style.getPropertyValue('--test-var') === 'test')
      
      // 测试Transform
      testElement.style.transform = 'translateX(10px)'
      this.testResults.set('transform', testElement.style.transform.includes('translateX'))
      
      // 测试Transition
      testElement.style.transition = 'all 0.3s ease'
      this.testResults.set('transition', testElement.style.transition.includes('0.3s'))
      
    } finally {
      document.body.removeChild(testElement)
    }
  }
  
  /**
   * 测试JavaScript功能
   */
  private testJavaScriptFeatures(): void {
    // 测试ES6功能
    this.testResults.set('es6Arrow', this.testES6Arrow())
    this.testResults.set('es6Promise', typeof Promise !== 'undefined')
    this.testResults.set('es6Map', typeof Map !== 'undefined')
    this.testResults.set('es6Set', typeof Set !== 'undefined')
    
    // 测试Web APIs
    this.testResults.set('fetch', typeof fetch !== 'undefined')
    this.testResults.set('webSocket', typeof WebSocket !== 'undefined')
    this.testResults.set('geolocation', 'geolocation' in navigator)
    this.testResults.set('clipboard', 'clipboard' in navigator)
  }
  
  /**
   * 测试性能
   */
  private async testPerformance(): Promise<void> {
    // 测试渲染性能
    const startTime = performance.now()
    
    // 创建测试元素
    const testContainer = document.createElement('div')
    testContainer.style.position = 'absolute'
    testContainer.style.top = '-9999px'
    testContainer.style.left = '-9999px'
    
    for (let i = 0; i < 100; i++) {
      const element = document.createElement('div')
      element.textContent = `Test element ${i}`
      element.style.padding = '10px'
      element.style.margin = '5px'
      element.style.backgroundColor = '#f0f0f0'
      testContainer.appendChild(element)
    }
    
    document.body.appendChild(testContainer)
    
    // 强制重排
    testContainer.offsetHeight
    
    const endTime = performance.now()
    const renderTime = endTime - startTime
    
    // 清理
    document.body.removeChild(testContainer)
    
    // 性能测试结果
    this.testResults.set('renderPerformance', renderTime < 100) // 100ms内完成渲染
    this.testResults.set('memoryUsage', this.testMemoryUsage())
  }
  
  /**
   * 测试本地存储
   */
  private testLocalStorage(): boolean {
    try {
      const testKey = '__test_localStorage__'
      localStorage.setItem(testKey, 'test')
      const result = localStorage.getItem(testKey) === 'test'
      localStorage.removeItem(testKey)
      return result
    } catch {
      return false
    }
  }
  
  /**
   * 测试会话存储
   */
  private testSessionStorage(): boolean {
    try {
      const testKey = '__test_sessionStorage__'
      sessionStorage.setItem(testKey, 'test')
      const result = sessionStorage.getItem(testKey) === 'test'
      sessionStorage.removeItem(testKey)
      return result
    } catch {
      return false
    }
  }
  
  /**
   * 测试ES6箭头函数
   */
  private testES6Arrow(): boolean {
    try {
      eval('(() => {})')
      return true
    } catch {
      return false
    }
  }
  
  /**
   * 测试内存使用
   */
  private testMemoryUsage(): boolean {
    if ('memory' in performance) {
      const memory = (performance as any).memory
      return memory.usedJSHeapSize < memory.jsHeapSizeLimit * 0.8
    }
    return true // 无法检测时假设正常
  }
  
  /**
   * 获取测试报告
   */
  getTestReport(): string {
    const deviceInfo = getCurrentDeviceInfo()
    const passedTests = Array.from(this.testResults.entries()).filter(([, passed]) => passed)
    const failedTests = Array.from(this.testResults.entries()).filter(([, passed]) => !passed)
    
    return `
设备兼容性测试报告
==================

设备信息:
- 类型: ${deviceInfo.type}
- 分辨率: ${deviceInfo.width}x${deviceInfo.height}
- 像素比: ${deviceInfo.pixelRatio}
- 方向: ${deviceInfo.orientation}
- 触摸支持: ${deviceInfo.touchSupport ? '是' : '否'}
- 浏览器: ${deviceInfo.browser}
- 操作系统: ${deviceInfo.os}

测试结果:
- 总测试数: ${this.testResults.size}
- 通过测试: ${passedTests.length}
- 失败测试: ${failedTests.length}
- 通过率: ${Math.round((passedTests.length / this.testResults.size) * 100)}%

通过的测试:
${passedTests.map(([test]) => `✓ ${test}`).join('\n')}

失败的测试:
${failedTests.map(([test]) => `✗ ${test}`).join('\n')}
    `.trim()
  }
}
