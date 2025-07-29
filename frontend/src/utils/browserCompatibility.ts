/**
 * 浏览器兼容性检测工具
 */

// 浏览器信息接口
export interface BrowserInfo {
  name: string
  version: string
  engine: string
  platform: string
  mobile: boolean
  supported: boolean
  features: BrowserFeatures
}

// 浏览器功能支持检测
export interface BrowserFeatures {
  webSocket: boolean
  localStorage: boolean
  sessionStorage: boolean
  indexedDB: boolean
  webWorkers: boolean
  serviceWorker: boolean
  pushNotifications: boolean
  geolocation: boolean
  camera: boolean
  microphone: boolean
  fullscreen: boolean
  clipboard: boolean
  share: boolean
  css: {
    flexbox: boolean
    grid: boolean
    customProperties: boolean
    transforms: boolean
    transitions: boolean
    animations: boolean
  }
  js: {
    es6: boolean
    modules: boolean
    asyncAwait: boolean
    fetch: boolean
    promises: boolean
    arrow: boolean
    destructuring: boolean
    spread: boolean
  }
}

/**
 * 检测浏览器信息
 */
export function detectBrowser(): BrowserInfo {
  const userAgent = navigator.userAgent
  const platform = navigator.platform
  
  // 检测浏览器名称和版本
  let name = 'Unknown'
  let version = 'Unknown'
  let engine = 'Unknown'
  
  // Chrome
  if (/Chrome\/(\d+)/.test(userAgent)) {
    name = 'Chrome'
    version = RegExp.$1
    engine = 'Blink'
  }
  // Firefox
  else if (/Firefox\/(\d+)/.test(userAgent)) {
    name = 'Firefox'
    version = RegExp.$1
    engine = 'Gecko'
  }
  // Safari
  else if (/Safari\/(\d+)/.test(userAgent) && !/Chrome/.test(userAgent)) {
    name = 'Safari'
    const match = userAgent.match(/Version\/(\d+)/)
    version = match ? match[1] : 'Unknown'
    engine = 'WebKit'
  }
  // Edge
  else if (/Edg\/(\d+)/.test(userAgent)) {
    name = 'Edge'
    version = RegExp.$1
    engine = 'Blink'
  }
  // Internet Explorer
  else if (/MSIE (\d+)/.test(userAgent) || /Trident.*rv:(\d+)/.test(userAgent)) {
    name = 'Internet Explorer'
    version = RegExp.$1
    engine = 'Trident'
  }
  
  // 检测移动设备
  const mobile = /Mobile|Android|iPhone|iPad/.test(userAgent)
  
  // 检测功能支持
  const features = detectFeatures()
  
  // 判断是否支持
  const supported = isBrowserSupported(name, parseInt(version), features)
  
  return {
    name,
    version,
    engine,
    platform,
    mobile,
    supported,
    features
  }
}

/**
 * 检测浏览器功能支持
 */
export function detectFeatures(): BrowserFeatures {
  return {
    webSocket: 'WebSocket' in window,
    localStorage: 'localStorage' in window,
    sessionStorage: 'sessionStorage' in window,
    indexedDB: 'indexedDB' in window,
    webWorkers: 'Worker' in window,
    serviceWorker: 'serviceWorker' in navigator,
    pushNotifications: 'PushManager' in window,
    geolocation: 'geolocation' in navigator,
    camera: 'mediaDevices' in navigator && 'getUserMedia' in navigator.mediaDevices,
    microphone: 'mediaDevices' in navigator && 'getUserMedia' in navigator.mediaDevices,
    fullscreen: 'requestFullscreen' in document.documentElement,
    clipboard: 'clipboard' in navigator,
    share: 'share' in navigator,
    css: {
      flexbox: CSS.supports('display', 'flex'),
      grid: CSS.supports('display', 'grid'),
      customProperties: CSS.supports('--test', 'red'),
      transforms: CSS.supports('transform', 'translateX(1px)'),
      transitions: CSS.supports('transition', 'all 1s'),
      animations: CSS.supports('animation', 'test 1s')
    },
    js: {
      es6: checkES6Support(),
      modules: 'noModule' in HTMLScriptElement.prototype,
      asyncAwait: checkAsyncAwaitSupport(),
      fetch: 'fetch' in window,
      promises: 'Promise' in window,
      arrow: checkArrowFunctionSupport(),
      destructuring: checkDestructuringSupport(),
      spread: checkSpreadSupport()
    }
  }
}

/**
 * 检查ES6支持
 */
function checkES6Support(): boolean {
  try {
    // 检查let/const
    eval('let test = 1; const test2 = 2;')
    return true
  } catch {
    return false
  }
}

/**
 * 检查async/await支持
 */
function checkAsyncAwaitSupport(): boolean {
  try {
    eval('(async function() { await Promise.resolve(); })')
    return true
  } catch {
    return false
  }
}

/**
 * 检查箭头函数支持
 */
function checkArrowFunctionSupport(): boolean {
  try {
    eval('(() => {})')
    return true
  } catch {
    return false
  }
}

/**
 * 检查解构赋值支持
 */
function checkDestructuringSupport(): boolean {
  try {
    eval('const {a} = {a: 1}; const [b] = [1];')
    return true
  } catch {
    return false
  }
}

/**
 * 检查展开运算符支持
 */
function checkSpreadSupport(): boolean {
  try {
    eval('const a = [...[1, 2, 3]]; const b = {...{x: 1}};')
    return true
  } catch {
    return false
  }
}

/**
 * 判断浏览器是否支持
 */
export function isBrowserSupported(name: string, version: number, features: BrowserFeatures): boolean {
  // 最低版本要求
  const minVersions: Record<string, number> = {
    'Chrome': 70,
    'Firefox': 65,
    'Safari': 12,
    'Edge': 79,
    'Internet Explorer': 0 // 不支持IE
  }
  
  // 检查版本
  const minVersion = minVersions[name]
  if (minVersion === undefined || version < minVersion) {
    return false
  }
  
  // 检查关键功能
  const requiredFeatures = [
    features.webSocket,
    features.localStorage,
    features.fetch,
    features.css.flexbox,
    features.js.es6,
    features.js.promises
  ]
  
  return requiredFeatures.every(feature => feature)
}

/**
 * 获取不支持的功能列表
 */
export function getUnsupportedFeatures(features: BrowserFeatures): string[] {
  const unsupported: string[] = []
  
  if (!features.webSocket) unsupported.push('WebSocket')
  if (!features.localStorage) unsupported.push('本地存储')
  if (!features.fetch) unsupported.push('Fetch API')
  if (!features.css.flexbox) unsupported.push('CSS Flexbox')
  if (!features.css.grid) unsupported.push('CSS Grid')
  if (!features.js.es6) unsupported.push('ES6语法')
  if (!features.js.asyncAwait) unsupported.push('Async/Await')
  if (!features.js.modules) unsupported.push('ES模块')
  
  return unsupported
}

/**
 * 显示浏览器兼容性警告
 */
export function showCompatibilityWarning(browserInfo: BrowserInfo): void {
  if (!browserInfo.supported) {
    const unsupportedFeatures = getUnsupportedFeatures(browserInfo.features)
    
    console.warn('浏览器兼容性警告:', {
      browser: `${browserInfo.name} ${browserInfo.version}`,
      unsupportedFeatures,
      recommendation: '建议升级到最新版本的Chrome、Firefox、Safari或Edge浏览器'
    })
    
    // 可以在这里显示用户友好的警告消息
    if (typeof window !== 'undefined' && document.body) {
      showUserWarning(browserInfo, unsupportedFeatures)
    }
  }
}

/**
 * 显示用户警告
 */
function showUserWarning(browserInfo: BrowserInfo, unsupportedFeatures: string[]): void {
  // 创建警告横幅
  const banner = document.createElement('div')
  banner.style.cssText = `
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    background: #ff6b6b;
    color: white;
    padding: 12px 20px;
    text-align: center;
    z-index: 10000;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    font-size: 14px;
    line-height: 1.4;
  `
  
  banner.innerHTML = `
    <strong>浏览器兼容性提醒</strong><br>
    您的浏览器 ${browserInfo.name} ${browserInfo.version} 可能无法完全支持本网站的所有功能。<br>
    建议升级到最新版本的Chrome、Firefox、Safari或Edge浏览器以获得最佳体验。
    <button onclick="this.parentElement.remove()" style="
      background: rgba(255,255,255,0.2);
      border: 1px solid rgba(255,255,255,0.3);
      color: white;
      padding: 4px 8px;
      margin-left: 10px;
      border-radius: 4px;
      cursor: pointer;
    ">关闭</button>
  `
  
  document.body.insertBefore(banner, document.body.firstChild)
  
  // 5秒后自动隐藏
  setTimeout(() => {
    if (banner.parentElement) {
      banner.remove()
    }
  }, 5000)
}

/**
 * 初始化浏览器兼容性检测
 */
export function initBrowserCompatibility(): BrowserInfo {
  const browserInfo = detectBrowser()
  
  // 在开发环境下输出详细信息
  if (import.meta.env.DEV) {
    console.log('浏览器信息:', browserInfo)
  }
  
  // 显示兼容性警告
  showCompatibilityWarning(browserInfo)
  
  return browserInfo
}
