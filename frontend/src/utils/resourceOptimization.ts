/**
 * 资源优化工具
 * 图片压缩、字体优化、CDN配置等
 */

// 图片格式支持检测
export interface ImageFormatSupport {
  webp: boolean
  avif: boolean
  jpeg2000: boolean
  jpegXR: boolean
}

// 图片优化配置
export interface ImageOptimizationConfig {
  quality: number
  maxWidth: number
  maxHeight: number
  format: 'webp' | 'jpeg' | 'png' | 'auto'
  progressive: boolean
}

/**
 * 检测浏览器支持的图片格式
 */
export function detectImageFormatSupport(): Promise<ImageFormatSupport> {
  return new Promise((resolve) => {
    const support: ImageFormatSupport = {
      webp: false,
      avif: false,
      jpeg2000: false,
      jpegXR: false
    }

    let testsCompleted = 0
    const totalTests = 4

    const checkComplete = () => {
      testsCompleted++
      if (testsCompleted === totalTests) {
        resolve(support)
      }
    }

    // 检测WebP支持
    const webpImg = new Image()
    webpImg.onload = webpImg.onerror = () => {
      support.webp = webpImg.height === 2
      checkComplete()
    }
    webpImg.src = 'data:image/webp;base64,UklGRjoAAABXRUJQVlA4IC4AAACyAgCdASoCAAIALmk0mk0iIiIiIgBoSygABc6WWgAA/veff/0PP8bA//LwYAAA'

    // 检测AVIF支持
    const avifImg = new Image()
    avifImg.onload = avifImg.onerror = () => {
      support.avif = avifImg.height === 2
      checkComplete()
    }
    avifImg.src = 'data:image/avif;base64,AAAAIGZ0eXBhdmlmAAAAAGF2aWZtaWYxbWlhZk1BMUIAAADybWV0YQAAAAAAAAAoaGRscgAAAAAAAAAAcGljdAAAAAAAAAAAAAAAAGxpYmF2aWYAAAAADnBpdG0AAAAAAAEAAAAeaWxvYwAAAABEAAABAAEAAAABAAABGgAAAB0AAAAoaWluZgAAAAAAAQAAABppbmZlAgAAAAABAABhdjAxQ29sb3IAAAAAamlwcnAAAABLaXBjbwAAABRpc3BlAAAAAAAAAAIAAAACAAAAEHBpeGkAAAAAAwgICAAAAAxhdjFDgQ0MAAAAABNjb2xybmNseAACAAIAAYAAAAAXaXBtYQAAAAAAAAABAAEEAQKDBAAAACVtZGF0EgAKCBgABogQEAwgMg8f8D///8WfhwB8+ErK42A='

    // 检测JPEG 2000支持
    const jp2Img = new Image()
    jp2Img.onload = jp2Img.onerror = () => {
      support.jpeg2000 = jp2Img.height === 2
      checkComplete()
    }
    jp2Img.src = 'data:image/jp2;base64,/0//UQAyAAAAAAABAAAAAgAAAAAAAAAAAAAABAAAAAQAAAAAAAAAAAAEBwEBBwEBBwEBBwEB/1IADAAAAAEAAAQEAAH/2AMQAAAAAQAB/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/2wBDAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwA/8H//2Q=='

    // 检测JPEG XR支持
    const jxrImg = new Image()
    jxrImg.onload = jxrImg.onerror = () => {
      support.jpegXR = jxrImg.height === 2
      checkComplete()
    }
    jxrImg.src = 'data:image/vnd.ms-photo;base64,SUm8AQgAAAAFAAG8AQAQAAAASgAAAIC8BAABAAAAAQAAAIG8BAABAAAAAQAAAIG8BAABAAAAAgAAAMC8BAABAAAAWgAAAMG8BAABAAAAHwAAAAAAAAAkw91vA07+S7GFPXd2jckQAA=='
  })
}

/**
 * 图片优化器
 */
export class ImageOptimizer {
  private formatSupport: ImageFormatSupport | null = null

  constructor() {
    this.initFormatSupport()
  }

  /**
   * 初始化格式支持检测
   */
  private async initFormatSupport() {
    this.formatSupport = await detectImageFormatSupport()
  }

  /**
   * 获取最佳图片格式
   */
  getBestFormat(originalFormat: string): string {
    if (!this.formatSupport) return originalFormat

    // 优先级：AVIF > WebP > 原格式
    if (this.formatSupport.avif && originalFormat !== 'gif') {
      return 'avif'
    }
    if (this.formatSupport.webp && originalFormat !== 'gif') {
      return 'webp'
    }
    return originalFormat
  }

  /**
   * 生成响应式图片URL
   */
  generateResponsiveImageUrl(
    baseUrl: string, 
    width: number, 
    config: Partial<ImageOptimizationConfig> = {}
  ): string {
    const defaultConfig: ImageOptimizationConfig = {
      quality: 80,
      maxWidth: 1920,
      maxHeight: 1080,
      format: 'auto',
      progressive: true
    }

    const finalConfig = { ...defaultConfig, ...config }
    const format = finalConfig.format === 'auto' 
      ? this.getBestFormat('jpeg') 
      : finalConfig.format

    // 这里可以根据实际的CDN服务调整URL格式
    // 示例使用通用的查询参数格式
    const params = new URLSearchParams({
      w: width.toString(),
      q: finalConfig.quality.toString(),
      f: format,
      ...(finalConfig.progressive && { progressive: 'true' })
    })

    return `${baseUrl}?${params.toString()}`
  }

  /**
   * 生成srcset属性
   */
  generateSrcSet(baseUrl: string, config?: Partial<ImageOptimizationConfig>): string {
    const breakpoints = [320, 640, 768, 1024, 1280, 1920]
    
    return breakpoints
      .map(width => `${this.generateResponsiveImageUrl(baseUrl, width, config)} ${width}w`)
      .join(', ')
  }

  /**
   * 压缩图片文件
   */
  async compressImage(
    file: File, 
    config: Partial<ImageOptimizationConfig> = {}
  ): Promise<Blob> {
    const defaultConfig: ImageOptimizationConfig = {
      quality: 0.8,
      maxWidth: 1920,
      maxHeight: 1080,
      format: 'auto',
      progressive: true
    }

    const finalConfig = { ...defaultConfig, ...config }

    return new Promise((resolve, reject) => {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      const img = new Image()

      img.onload = () => {
        // 计算新尺寸
        let { width, height } = img
        const aspectRatio = width / height

        if (width > finalConfig.maxWidth) {
          width = finalConfig.maxWidth
          height = width / aspectRatio
        }

        if (height > finalConfig.maxHeight) {
          height = finalConfig.maxHeight
          width = height * aspectRatio
        }

        // 设置canvas尺寸
        canvas.width = width
        canvas.height = height

        // 绘制图片
        ctx?.drawImage(img, 0, 0, width, height)

        // 转换为blob
        canvas.toBlob(
          (blob) => {
            if (blob) {
              resolve(blob)
            } else {
              reject(new Error('Failed to compress image'))
            }
          },
          finalConfig.format === 'auto' ? 'image/jpeg' : `image/${finalConfig.format}`,
          finalConfig.quality
        )
      }

      img.onerror = () => reject(new Error('Failed to load image'))
      img.src = URL.createObjectURL(file)
    })
  }
}

/**
 * 字体优化器
 */
export class FontOptimizer {
  private loadedFonts = new Set<string>()

  /**
   * 预加载字体
   */
  async preloadFont(fontUrl: string, fontFamily: string): Promise<void> {
    if (this.loadedFonts.has(fontUrl)) {
      return
    }

    return new Promise((resolve, reject) => {
      const link = document.createElement('link')
      link.rel = 'preload'
      link.as = 'font'
      link.type = 'font/woff2'
      link.crossOrigin = 'anonymous'
      link.href = fontUrl

      link.onload = () => {
        this.loadedFonts.add(fontUrl)
        resolve()
      }

      link.onerror = () => reject(new Error(`Failed to preload font: ${fontUrl}`))

      document.head.appendChild(link)
    })
  }

  /**
   * 动态加载字体
   */
  async loadFont(fontFamily: string, fontUrl: string, fontWeight = '400'): Promise<void> {
    if ('FontFace' in window) {
      const fontFace = new FontFace(fontFamily, `url(${fontUrl})`, {
        weight: fontWeight
      })

      try {
        const loadedFont = await fontFace.load()
        document.fonts.add(loadedFont)
        this.loadedFonts.add(fontUrl)
      } catch (error) {
        console.error('Failed to load font:', error)
        throw error
      }
    } else {
      // 降级方案
      await this.preloadFont(fontUrl, fontFamily)
    }
  }

  /**
   * 检测字体加载状态
   */
  isFontLoaded(fontFamily: string): boolean {
    if ('fonts' in document) {
      return document.fonts.check(`1em ${fontFamily}`)
    }
    return false
  }

  /**
   * 等待字体加载完成
   */
  async waitForFontLoad(fontFamily: string, timeout = 3000): Promise<boolean> {
    const startTime = Date.now()

    while (Date.now() - startTime < timeout) {
      if (this.isFontLoaded(fontFamily)) {
        return true
      }
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    return false
  }
}

/**
 * CDN资源管理器
 */
export class CDNManager {
  private cdnBaseUrl: string
  private fallbackUrls: string[] = []

  constructor(cdnBaseUrl: string, fallbackUrls: string[] = []) {
    this.cdnBaseUrl = cdnBaseUrl
    this.fallbackUrls = fallbackUrls
  }

  /**
   * 获取CDN资源URL
   */
  getResourceUrl(path: string): string {
    return `${this.cdnBaseUrl}/${path.replace(/^\//, '')}`
  }

  /**
   * 获取带降级的资源URL
   */
  async getResourceUrlWithFallback(path: string): Promise<string> {
    const primaryUrl = this.getResourceUrl(path)

    // 检测主CDN是否可用
    if (await this.testResourceAvailability(primaryUrl)) {
      return primaryUrl
    }

    // 尝试降级URL
    for (const fallbackBase of this.fallbackUrls) {
      const fallbackUrl = `${fallbackBase}/${path.replace(/^\//, '')}`
      if (await this.testResourceAvailability(fallbackUrl)) {
        return fallbackUrl
      }
    }

    // 如果都不可用，返回原始路径
    return path
  }

  /**
   * 测试资源可用性
   */
  private async testResourceAvailability(url: string): Promise<boolean> {
    try {
      const response = await fetch(url, { method: 'HEAD', mode: 'no-cors' })
      return response.ok
    } catch {
      return false
    }
  }

  /**
   * 预加载关键资源
   */
  preloadCriticalResources(resources: string[]): void {
    resources.forEach(resource => {
      const link = document.createElement('link')
      link.rel = 'preload'
      link.href = this.getResourceUrl(resource)
      
      // 根据文件扩展名设置as属性
      const ext = resource.split('.').pop()?.toLowerCase()
      switch (ext) {
        case 'css':
          link.as = 'style'
          break
        case 'js':
          link.as = 'script'
          break
        case 'woff':
        case 'woff2':
          link.as = 'font'
          link.crossOrigin = 'anonymous'
          break
        case 'jpg':
        case 'jpeg':
        case 'png':
        case 'webp':
        case 'avif':
          link.as = 'image'
          break
        default:
          link.as = 'fetch'
          link.crossOrigin = 'anonymous'
      }

      document.head.appendChild(link)
    })
  }
}

/**
 * 资源优化管理器
 */
export class ResourceOptimizationManager {
  private imageOptimizer: ImageOptimizer
  private fontOptimizer: FontOptimizer
  private cdnManager: CDNManager | null = null

  constructor(cdnConfig?: { baseUrl: string; fallbackUrls?: string[] }) {
    this.imageOptimizer = new ImageOptimizer()
    this.fontOptimizer = new FontOptimizer()
    
    if (cdnConfig) {
      this.cdnManager = new CDNManager(cdnConfig.baseUrl, cdnConfig.fallbackUrls)
    }
  }

  /**
   * 获取图片优化器
   */
  getImageOptimizer(): ImageOptimizer {
    return this.imageOptimizer
  }

  /**
   * 获取字体优化器
   */
  getFontOptimizer(): FontOptimizer {
    return this.fontOptimizer
  }

  /**
   * 获取CDN管理器
   */
  getCDNManager(): CDNManager | null {
    return this.cdnManager
  }

  /**
   * 初始化资源优化
   */
  async initialize(): Promise<void> {
    // 预加载关键字体
    const criticalFonts = [
      { family: 'Inter', url: '/fonts/inter-regular.woff2' },
      { family: 'Inter', url: '/fonts/inter-medium.woff2', weight: '500' }
    ]

    for (const font of criticalFonts) {
      try {
        await this.fontOptimizer.loadFont(font.family, font.url, font.weight)
      } catch (error) {
        console.warn('Failed to load critical font:', font, error)
      }
    }

    // 预加载关键资源
    if (this.cdnManager) {
      this.cdnManager.preloadCriticalResources([
        'css/critical.css',
        'js/critical.js'
      ])
    }
  }
}

// 全局资源优化管理器实例
let globalResourceManager: ResourceOptimizationManager | null = null

/**
 * 初始化资源优化
 */
export function initResourceOptimization(cdnConfig?: { baseUrl: string; fallbackUrls?: string[] }) {
  if (!globalResourceManager) {
    globalResourceManager = new ResourceOptimizationManager(cdnConfig)
    globalResourceManager.initialize()
  }
  return globalResourceManager
}

/**
 * 获取全局资源管理器
 */
export function getResourceManager(): ResourceOptimizationManager | null {
  return globalResourceManager
}
