import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import { VitePWA } from 'vite-plugin-pwa'
import { visualizer } from 'rollup-plugin-visualizer'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { resourceOptimization } from './vite-plugins/resource-optimization'

// https://vite.dev/config/
export default defineConfig(({ command, mode }) => {
  // 加载环境变量
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [
      vue(),
      vueDevTools(),
      // Element Plus 按需加载
      AutoImport({
        resolvers: [ElementPlusResolver()],
        dts: true
      }),
      Components({
        resolvers: [ElementPlusResolver()],
        dts: true
      }),
      // PWA 插件
      VitePWA({
        registerType: 'prompt',
        includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'masked-icon.svg'],
        manifest: {
          name: 'Flash Sku - 闪购秒杀',
          short_name: 'Flash Sku',
          description: '高性能秒杀抢购平台，支持大规模并发的闪购活动',
          theme_color: '#667eea',
          background_color: '#ffffff',
          display: 'standalone',
          scope: '/',
          start_url: '/',
          icons: [
            {
              src: 'icons/icon-192x192.png',
              sizes: '192x192',
              type: 'image/png'
            },
            {
              src: 'icons/icon-512x512.png',
              sizes: '512x512',
              type: 'image/png'
            },
            {
              src: 'icons/icon-512x512.png',
              sizes: '512x512',
              type: 'image/png',
              purpose: 'any maskable'
            }
          ]
        },
        workbox: {
          globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
          runtimeCaching: [
            {
              urlPattern: /^https:\/\/api\.flash-sku\.com\/api\/products/,
              handler: 'CacheFirst',
              options: {
                cacheName: 'api-products',
                expiration: {
                  maxEntries: 100,
                  maxAgeSeconds: 60 * 60 * 24 // 24 hours
                }
              }
            },
            {
              urlPattern: /^https:\/\/api\.flash-sku\.com\/api\/activities/,
              handler: 'NetworkFirst',
              options: {
                cacheName: 'api-activities',
                expiration: {
                  maxEntries: 50,
                  maxAgeSeconds: 60 * 5 // 5 minutes
                }
              }
            },
            {
              urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp|avif)$/,
              handler: 'CacheFirst',
              options: {
                cacheName: 'images',
                expiration: {
                  maxEntries: 200,
                  maxAgeSeconds: 60 * 60 * 24 * 7 // 7 days
                }
              }
            }
          ]
        },
        devOptions: {
          enabled: mode === 'development',
          type: 'module'
        }
      }),
      // 资源优化插件
      resourceOptimization({
        images: {
          quality: 80,
          formats: ['webp', 'avif', 'jpeg'],
          sizes: [320, 640, 768, 1024, 1280, 1920],
          progressive: true
        },
        fonts: {
          preload: ['inter', 'roboto'],
          subset: true,
          formats: ['woff2', 'woff']
        },
        css: {
          purge: command === 'build',
          minify: command === 'build',
          critical: command === 'build'
        }
      }),
      // Bundle分析器 (仅在构建时启用)
      command === 'build' && visualizer({
        filename: 'dist/stats.html',
        open: false,
        gzipSize: true,
        brotliSize: true
      })
    ].filter(Boolean),
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      },
    },
    server: {
      port: parseInt(env.VITE_PORT) || 3000,
      host: true,
      proxy: {
        // 代理Django API
        '/api': {
          target: env.VITE_DJANGO_API_BASE_URL || 'http://localhost:8000',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, '/api'),
        },
        // 代理Go API
        '/seckill': {
          target: env.VITE_GO_API_BASE_URL || 'http://localhost:8080',
          changeOrigin: true,
        },
        // 代理WebSocket
        '/ws': {
          target: env.VITE_GO_API_BASE_URL || 'http://localhost:8080',
          changeOrigin: true,
          ws: true,
        },
      },
    },
    build: {
      // 浏览器兼容性配置
      target: ['es2015', 'chrome63', 'firefox67', 'safari12', 'edge79'],
      outDir: 'dist',
      assetsDir: 'assets',
      sourcemap: mode === 'development',
      // 压缩配置
      minify: 'terser',
      terserOptions: {
        compress: {
          drop_console: mode === 'production',
          drop_debugger: mode === 'production',
          pure_funcs: mode === 'production' ? ['console.log', 'console.info'] : []
        },
        mangle: {
          safari10: true
        }
      },
      // 构建优化
      chunkSizeWarningLimit: 1000,
      assetsInlineLimit: 4096,
      rollupOptions: {
        output: {
          manualChunks: (id) => {
            // Vue 核心库
            if (id.includes('vue') && !id.includes('node_modules')) {
              return 'vue-app'
            }
            if (id.includes('vue') || id.includes('pinia') || id.includes('vue-router')) {
              return 'vue-vendor'
            }
            // Element Plus
            if (id.includes('element-plus')) {
              return 'element-plus'
            }
            // Element Plus 图标
            if (id.includes('@element-plus/icons')) {
              return 'element-icons'
            }
            // 工具库
            if (id.includes('axios') || id.includes('dayjs') || id.includes('lodash')) {
              return 'utils'
            }
            // PWA 相关
            if (id.includes('workbox') || id.includes('sw-')) {
              return 'pwa'
            }
            // 大型第三方库单独分包
            if (id.includes('echarts')) return 'echarts'
            if (id.includes('monaco-editor')) return 'monaco'

            // 其他第三方库
            if (id.includes('node_modules')) {
              return 'vendor'
            }
          },
          // 文件名配置
          chunkFileNames: 'js/[name]-[hash].js',
          entryFileNames: 'js/[name]-[hash].js',
          assetFileNames: (assetInfo) => {
            const fileName = assetInfo.names?.[0] || assetInfo.originalFileNames?.[0] || 'unknown'
            const info = fileName.split('.') || []
            let extType = info[info.length - 1]
            if (/\.(mp4|webm|ogg|mp3|wav|flac|aac)(\?.*)?$/i.test(fileName)) {
              extType = 'media'
            } else if (/\.(png|jpe?g|gif|svg)(\?.*)?$/i.test(fileName)) {
              extType = 'img'
            } else if (/\.(woff2?|eot|ttf|otf)(\?.*)?$/i.test(fileName)) {
              extType = 'fonts'
            }
            return `${extType}/[name]-[hash].[ext]`
          }
        },
        // 外部化依赖（CDN）
        external: command === 'build' ? [] : [],
      }
    },
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: `@import "@/styles/variables.scss";`,
        },
      },
    },
    define: {
      __VUE_PROD_DEVTOOLS__: mode === 'development',
    },
    // 优化依赖预构建
    optimizeDeps: {
      include: [
        'vue',
        'vue-router',
        'pinia',
        'element-plus',
        'axios',
        'dayjs',
        'lodash-es'
      ],
      // 排除不需要预构建的依赖
      exclude: ['@vueuse/core']
    },
    // 实验性功能
    experimental: {
      // 启用渲染优化
      renderBuiltUrl(filename, { hostType }) {
        if (hostType === 'js') {
          return { js: `/${filename}` }
        } else {
          return { relative: true }
        }
      }
    },
  }
})
