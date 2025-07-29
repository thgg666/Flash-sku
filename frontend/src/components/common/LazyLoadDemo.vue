<template>
  <div class="lazy-load-demo">
    <el-card class="demo-card">
      <template #header>
        <h3>懒加载演示</h3>
      </template>

      <!-- 图片懒加载演示 -->
      <div class="demo-section">
        <h4>图片懒加载</h4>
        <p>滚动到图片位置时才开始加载，提升页面初始加载速度</p>
        
        <div class="image-grid">
          <div 
            v-for="(image, index) in demoImages" 
            :key="index"
            class="image-item"
          >
            <img
              v-lazy-load="{
                src: image.src,
                loading: image.loading,
                error: image.error,
                delay: index * 100
              }"
              :alt="image.alt"
              class="demo-image"
              @lazy-loaded="onImageLoaded"
              @lazy-error="onImageError"
            />
            <div class="image-info">
              <span>{{ image.title }}</span>
              <el-tag v-if="loadedImages.includes(index)" type="success" size="small">
                已加载
              </el-tag>
            </div>
          </div>
        </div>
      </div>

      <!-- 组件懒加载演示 -->
      <div class="demo-section">
        <h4>组件懒加载</h4>
        <p>点击按钮动态加载组件，减少初始包体积</p>
        
        <div class="component-demo">
          <el-button @click="loadHeavyComponent" :loading="componentLoading" type="primary">
            加载重型组件
          </el-button>
          
          <div v-if="heavyComponentLoaded" class="component-container">
            <Suspense>
              <template #default>
                <component :is="HeavyComponent" />
              </template>
              <template #fallback>
                <div class="loading-placeholder">
                  <el-skeleton :rows="3" animated />
                </div>
              </template>
            </Suspense>
          </div>
        </div>
      </div>

      <!-- 路由预加载演示 -->
      <div class="demo-section">
        <h4>路由预加载</h4>
        <p>鼠标悬停时预加载路由组件，提升导航体验</p>
        
        <div class="route-demo">
          <PreloadLink 
            to="/activities" 
            strategy="hover"
            class="preload-link"
          >
            <el-button type="primary" plain>
              悬停预加载活动页面
            </el-button>
          </PreloadLink>
          
          <PreloadLink 
            to="/about" 
            strategy="visible"
            class="preload-link"
          >
            <el-button type="success" plain>
              可见时预加载关于页面
            </el-button>
          </PreloadLink>
          
          <div class="preload-status">
            <h5>预加载状态</h5>
            <div v-for="(state, route) in preloadStates" :key="route" class="status-item">
              <span>{{ route }}:</span>
              <el-tag 
                :type="state.loaded ? 'success' : state.loading ? 'warning' : 'info'"
                size="small"
              >
                {{ state.loaded ? '已加载' : state.loading ? '加载中' : '未加载' }}
              </el-tag>
            </div>
          </div>
        </div>
      </div>

      <!-- 资源预加载演示 -->
      <div class="demo-section">
        <h4>资源预加载</h4>
        <p>提前加载可能需要的资源</p>
        
        <div class="resource-demo">
          <el-button @click="preloadResources" :loading="preloading" type="primary">
            预加载资源
          </el-button>
          
          <div v-if="preloadedResources.length > 0" class="preloaded-list">
            <h5>已预加载的资源</h5>
            <ul>
              <li v-for="resource in preloadedResources" :key="resource">
                {{ resource }}
              </li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 性能统计 -->
      <div class="demo-section">
        <h4>性能统计</h4>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-value">{{ loadedImages.length }}</div>
            <div class="stat-label">已加载图片</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ heavyComponentLoaded ? 1 : 0 }}</div>
            <div class="stat-label">已加载组件</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ Object.keys(preloadStates).length }}</div>
            <div class="stat-label">预加载路由</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ preloadedResources.length }}</div>
            <div class="stat-label">预加载资源</div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, defineAsyncComponent } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoutePreload } from '@/composables/useRoutePreload'
import { PreloadLink } from '@/composables/useRoutePreload'
import { ResourcePreloader } from '@/utils/performance'

// 路由预加载
const { preloadStates } = useRoutePreload()

// 状态
const loadedImages = ref<number[]>([])
const componentLoading = ref(false)
const heavyComponentLoaded = ref(false)
const preloading = ref(false)
const preloadedResources = ref<string[]>([])

// 重型组件懒加载
const HeavyComponent = defineAsyncComponent({
  loader: () => import('@/components/common/PerformanceMonitor.vue'),
  loadingComponent: {
    template: '<div class="loading">加载中...</div>'
  },
  errorComponent: {
    template: '<div class="error">加载失败</div>'
  },
  delay: 200,
  timeout: 30000
})

// 演示图片数据
const demoImages = [
  {
    src: 'https://picsum.photos/300/200?random=1',
    loading: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjBmMGYwIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkxvYWRpbmcuLi48L3RleHQ+PC9zdmc+',
    error: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjVmNWY1Ii8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iI2NjYyIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkVycm9yPC90ZXh0Pjwvc3ZnPg==',
    alt: '演示图片1',
    title: '随机图片 1'
  },
  {
    src: 'https://picsum.photos/300/200?random=2',
    loading: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjBmMGYwIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkxvYWRpbmcuLi48L3RleHQ+PC9zdmc+',
    error: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjVmNWY1Ii8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iI2NjYyIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkVycm9yPC90ZXh0Pjwvc3ZnPg==',
    alt: '演示图片2',
    title: '随机图片 2'
  },
  {
    src: 'https://picsum.photos/300/200?random=3',
    loading: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjBmMGYwIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkxvYWRpbmcuLi48L3RleHQ+PC9zdmc+',
    error: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjVmNWY1Ii8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iI2NjYyIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkVycm9yPC90ZXh0Pjwvc3ZnPg==',
    alt: '演示图片3',
    title: '随机图片 3'
  },
  {
    src: 'https://picsum.photos/300/200?random=4',
    loading: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjBmMGYwIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkxvYWRpbmcuLi48L3RleHQ+PC9zdmc+',
    error: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjVmNWY1Ii8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iI2NjYyIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkVycm9yPC90ZXh0Pjwvc3ZnPg==',
    alt: '演示图片4',
    title: '随机图片 4'
  }
]

// 图片加载成功回调
const onImageLoaded = (event: CustomEvent) => {
  const img = event.target as HTMLImageElement
  const index = demoImages.findIndex(item => item.src === event.detail.src)
  if (index !== -1 && !loadedImages.value.includes(index)) {
    loadedImages.value.push(index)
    ElMessage.success(`图片 ${index + 1} 加载成功`)
  }
}

// 图片加载失败回调
const onImageError = (event: CustomEvent) => {
  ElMessage.error(`图片加载失败: ${event.detail.src}`)
}

// 加载重型组件
const loadHeavyComponent = async () => {
  componentLoading.value = true
  try {
    // 模拟加载时间
    await new Promise(resolve => setTimeout(resolve, 1000))
    heavyComponentLoaded.value = true
    ElMessage.success('重型组件加载成功')
  } catch (error) {
    ElMessage.error('组件加载失败')
  } finally {
    componentLoading.value = false
  }
}

// 预加载资源
const preloadResources = async () => {
  preloading.value = true
  try {
    const preloader = new ResourcePreloader()
    
    // 预加载图片
    const imagesToPreload = [
      'https://picsum.photos/300/200?random=5',
      'https://picsum.photos/300/200?random=6'
    ]
    
    await preloader.preloadImages(imagesToPreload)
    preloadedResources.value.push(...imagesToPreload)
    
    ElMessage.success('资源预加载完成')
  } catch (error) {
    ElMessage.error('资源预加载失败')
  } finally {
    preloading.value = false
  }
}
</script>

<style scoped lang="scss">
@import "@/styles/variables.scss";

.lazy-load-demo {
  .demo-card {
    max-width: 1000px;
    margin: 0 auto;
  }

  .demo-section {
    margin-bottom: 32px;
    padding-bottom: 24px;
    border-bottom: 1px solid $border-color-lighter;

    &:last-child {
      border-bottom: none;
    }

    h4 {
      color: $text-color-primary;
      margin-bottom: 8px;
    }

    p {
      color: $text-color-regular;
      margin-bottom: 16px;
    }
  }

  .image-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 16px;

    .image-item {
      .demo-image {
        width: 100%;
        height: 200px;
        object-fit: cover;
        border-radius: 8px;
        transition: opacity 0.3s ease;

        &.lazy-loading {
          opacity: 0.6;
        }

        &.lazy-loaded {
          opacity: 1;
        }

        &.lazy-error {
          opacity: 0.5;
        }
      }

      .image-info {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 8px;
        font-size: 14px;
        color: $text-color-regular;
      }
    }
  }

  .component-demo {
    .component-container {
      margin-top: 16px;
      padding: 16px;
      border: 1px solid $border-color-lighter;
      border-radius: 8px;
      background: $bg-color-page;

      .loading-placeholder {
        padding: 20px;
      }
    }
  }

  .route-demo {
    .preload-link {
      margin-right: 12px;
      margin-bottom: 12px;
      display: inline-block;
    }

    .preload-status {
      margin-top: 16px;
      padding: 16px;
      background: $bg-color-page;
      border-radius: 8px;

      h5 {
        margin-bottom: 12px;
        color: $text-color-primary;
      }

      .status-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 4px 0;
        border-bottom: 1px solid $border-color-lighter;

        &:last-child {
          border-bottom: none;
        }
      }
    }
  }

  .resource-demo {
    .preloaded-list {
      margin-top: 16px;
      padding: 16px;
      background: $bg-color-page;
      border-radius: 8px;

      h5 {
        margin-bottom: 12px;
        color: $text-color-primary;
      }

      ul {
        margin: 0;
        padding-left: 20px;

        li {
          margin-bottom: 4px;
          color: $text-color-regular;
          font-size: 14px;
        }
      }
    }
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 16px;

    .stat-item {
      text-align: center;
      padding: 16px;
      background: $bg-color-page;
      border-radius: 8px;
      border: 1px solid $border-color-lighter;

      .stat-value {
        font-size: 24px;
        font-weight: 600;
        color: $primary-color;
        margin-bottom: 4px;
      }

      .stat-label {
        font-size: 14px;
        color: $text-color-regular;
      }
    }
  }
}

// 移动端适配
@media (max-width: 768px) {
  .lazy-load-demo {
    .image-grid {
      grid-template-columns: 1fr;
    }

    .stats-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }
}
</style>
