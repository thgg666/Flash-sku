/**
 * Vite资源优化插件
 * 用于在构建时优化图片、字体等资源
 */

import type { Plugin } from 'vite'
import { writeFileSync } from 'fs'
import { resolve, dirname, extname, basename } from 'path'

interface ResourceOptimizationOptions {
  // 图片优化配置
  images?: {
    quality?: number
    formats?: ('webp' | 'avif' | 'jpeg' | 'png')[]
    sizes?: number[]
    progressive?: boolean
  }
  // 字体优化配置
  fonts?: {
    preload?: string[]
    subset?: boolean
    formats?: ('woff2' | 'woff' | 'ttf')[]
  }
  // CSS优化配置
  css?: {
    purge?: boolean
    minify?: boolean
    critical?: boolean
  }
  // 输出配置
  output?: {
    dir?: string
    publicPath?: string
  }
}

const defaultOptions: Required<ResourceOptimizationOptions> = {
  images: {
    quality: 80,
    formats: ['webp', 'avif', 'jpeg'],
    sizes: [320, 640, 768, 1024, 1280, 1920],
    progressive: true
  },
  fonts: {
    preload: [],
    subset: false,
    formats: ['woff2', 'woff']
  },
  css: {
    purge: true,
    minify: true,
    critical: true
  },
  output: {
    dir: 'dist',
    publicPath: '/'
  }
}

export function resourceOptimization(options: ResourceOptimizationOptions = {}): Plugin {
  const config = { ...defaultOptions, ...options }

  const plugin: Plugin = {
    name: 'resource-optimization',

    configResolved(resolvedConfig) {
      // 在配置解析后设置输出目录
      config.output.dir = resolvedConfig.build.outDir
    },

    generateBundle(_outputOptions, bundle) {
      // 在生成bundle时处理资源优化
      processAssets(bundle, config)
    },

    writeBundle(options, bundle) {
      const outputDir = options.dir || 'dist'

      // 生成资源清单
      generateAssetManifest(bundle, outputDir)

      // 生成预加载提示
      const preloadHints = generatePreloadHints(bundle)
      if (preloadHints.length > 0) {
        const hintsPath = resolve(outputDir, 'preload-hints.json')
        writeFileSync(hintsPath, JSON.stringify(preloadHints, null, 2))
        console.log('Preload hints generated:', hintsPath)
      }

      // 生成Service Worker缓存配置
      const swConfig = generateSWCacheConfig(bundle)
      const swConfigPath = resolve(outputDir, 'sw-cache-config.json')
      writeFileSync(swConfigPath, JSON.stringify(swConfig, null, 2))
      console.log('Service Worker cache config generated:', swConfigPath)
    }
  }

  // 处理资源文件
  function processAssets(bundle: any, config: Required<ResourceOptimizationOptions>) {
      Object.keys(bundle).forEach(fileName => {
        const asset = bundle[fileName]

        if (asset.type === 'asset') {
          const ext = extname(fileName).toLowerCase()

          // 处理图片文件
          if (['.jpg', '.jpeg', '.png', '.gif', '.svg'].includes(ext)) {
            processImage(asset, fileName, config.images)
          }

          // 处理字体文件
          if (['.woff', '.woff2', '.ttf', '.otf'].includes(ext)) {
            processFont(asset, fileName, config.fonts)
          }

          // 处理CSS文件
          if (ext === '.css') {
            processCSS(asset, fileName, config.css)
          }
        }
      })
    }

  return plugin
}

// 处理图片
function processImage(_asset: any, fileName: string, imageConfig: Required<ResourceOptimizationOptions>['images']) {
      const ext = extname(fileName).toLowerCase()
      const baseName = basename(fileName, ext)
      const dir = dirname(fileName)
      
      // 生成不同格式的图片
      imageConfig.formats?.forEach(format => {
        if (format !== ext.slice(1)) {
          const newFileName = `${dir}/${baseName}.${format}`

          // 这里应该调用实际的图片转换库
          // 例如 sharp, imagemin 等
          console.log(`Generating ${format} version: ${newFileName}`)

          // 模拟生成新格式的图片
          // 在实际实现中，这里应该是真正的图片转换逻辑
        }
      })

      // 生成不同尺寸的图片
      imageConfig.sizes?.forEach(size => {
        const newFileName = `${dir}/${baseName}-${size}w${ext}`
        console.log(`Generating ${size}px version: ${newFileName}`)

        // 模拟生成不同尺寸的图片
        // 在实际实现中，这里应该是真正的图片缩放逻辑
      })
    }

// 处理字体
function processFont(_asset: any, fileName: string, fontConfig: Required<ResourceOptimizationOptions>['fonts']) {
      const ext = extname(fileName).toLowerCase()
      
      // 检查是否需要预加载
      if (fontConfig.preload?.some(pattern => fileName.includes(pattern))) {
        console.log(`Font marked for preload: ${fileName}`)

        // 在HTML中添加preload链接
        // 这里应该与HTML插件集成
      }
      
      // 字体子集化
      if (fontConfig.subset && ['.ttf', '.otf'].includes(ext)) {
        console.log(`Subsetting font: ${fileName}`)

        // 这里应该调用字体子集化工具
        // 例如 fonttools, pyftsubset 等
      }
    }

// 处理CSS
function processCSS(asset: any, fileName: string, cssConfig: Required<ResourceOptimizationOptions>['css']) {
      let cssContent = asset.source.toString()
      
      // CSS压缩
      if (cssConfig.minify) {
        // 这里应该调用CSS压缩工具
        // 例如 cssnano, clean-css 等
        console.log(`Minifying CSS: ${fileName}`)
      }
      
      // CSS清理
      if (cssConfig.purge) {
        // 这里应该调用CSS清理工具
        // 例如 purgecss 等
        console.log(`Purging unused CSS: ${fileName}`)
      }
      
      // 提取关键CSS
      if (cssConfig.critical) {
        console.log(`Extracting critical CSS from: ${fileName}`)
        
        // 这里应该调用关键CSS提取工具
        // 例如 critical, penthouse 等
      }
      
      // 更新asset内容
      asset.source = cssContent
    }

// 生成资源清单
function generateAssetManifest(bundle: any, outputDir: string) {
      const manifest: Record<string, any> = {}
      
      Object.keys(bundle).forEach(fileName => {
        const ext = extname(fileName).toLowerCase()
        
        if (['.jpg', '.jpeg', '.png', '.webp', '.avif'].includes(ext)) {
          const baseName = basename(fileName, ext)
          
          if (!manifest[baseName]) {
            manifest[baseName] = {
              formats: {},
              sizes: {}
            }
          }
          
          // 记录不同格式
          manifest[baseName].formats[ext.slice(1)] = fileName
          
          // 记录不同尺寸
          const sizeMatch = fileName.match(/-(\d+)w/)
          if (sizeMatch) {
            const size = parseInt(sizeMatch[1])
            manifest[baseName].sizes[size] = fileName
          }
        }
      })
      
      // 写入清单文件
      const manifestPath = resolve(outputDir, 'asset-manifest.json')
      writeFileSync(manifestPath, JSON.stringify(manifest, null, 2))
      console.log('Asset manifest generated:', manifestPath)
    }

// 生成预加载提示
function generatePreloadHints(bundle: any) {
      const preloadHints: string[] = []
      
      Object.keys(bundle).forEach(fileName => {
        const ext = extname(fileName).toLowerCase()
        
        // 关键CSS文件
        if (fileName.includes('critical') && ext === '.css') {
          preloadHints.push(`<link rel="preload" href="/${fileName}" as="style">`)
        }
        
        // 关键字体文件
        if (['.woff2', '.woff'].includes(ext) && fileName.includes('critical')) {
          preloadHints.push(`<link rel="preload" href="/${fileName}" as="font" type="font/${ext.slice(1)}" crossorigin>`)
        }
        
        // 关键JavaScript文件
        if (fileName.includes('critical') && ext === '.js') {
          preloadHints.push(`<link rel="preload" href="/${fileName}" as="script">`)
        }
      })
      
      return preloadHints
    }

// 生成Service Worker缓存配置
function generateSWCacheConfig(bundle: any) {
      const cacheConfig = {
        static: [] as string[],
        images: [] as string[],
        fonts: [] as string[]
      }
      
      Object.keys(bundle).forEach(fileName => {
        const ext = extname(fileName).toLowerCase()
        
        if (['.js', '.css'].includes(ext)) {
          cacheConfig.static.push(fileName)
        } else if (['.jpg', '.jpeg', '.png', '.webp', '.avif', '.svg'].includes(ext)) {
          cacheConfig.images.push(fileName)
        } else if (['.woff', '.woff2', '.ttf', '.otf'].includes(ext)) {
          cacheConfig.fonts.push(fileName)
        }
      })
      
      return cacheConfig
    }



export { type ResourceOptimizationOptions }
