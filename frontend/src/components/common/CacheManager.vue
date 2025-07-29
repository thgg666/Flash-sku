<template>
  <div class="cache-manager">
    <el-card class="manager-card">
      <template #header>
        <div class="card-header">
          <h3>缓存管理</h3>
          <div class="header-actions">
            <el-button @click="refreshCacheInfo" :loading="loading" size="small" type="primary">
              刷新
            </el-button>
            <el-button @click="clearAllCache" :loading="clearing" size="small" type="danger">
              清空所有缓存
            </el-button>
          </div>
        </div>
      </template>

      <!-- Service Worker状态 -->
      <div class="sw-status-section">
        <h4>Service Worker状态</h4>
        <div class="status-info">
          <el-tag :type="getStatusType(swStatus)" size="large">
            {{ getStatusText(swStatus) }}
          </el-tag>
          <div class="status-actions">
            <el-button
              v-if="swStatus === 'unsupported'"
              size="small"
              disabled
            >
              不支持Service Worker
            </el-button>
            <el-button 
              v-else-if="swStatus === 'installing'" 
              size="small" 
              loading
            >
              安装中...
            </el-button>
            <el-button 
              v-else-if="swStatus === 'installed'" 
              @click="checkForUpdate" 
              size="small"
            >
              检查更新
            </el-button>
            <el-button 
              v-else-if="swStatus === 'updating'" 
              size="small" 
              loading
            >
              更新中...
            </el-button>
            <el-button 
              v-else-if="swStatus === 'updated'" 
              @click="activateUpdate" 
              size="small" 
              type="success"
            >
              激活更新
            </el-button>
          </div>
        </div>
      </div>

      <!-- 网络状态 -->
      <div class="network-status-section">
        <h4>网络状态</h4>
        <div class="network-info">
          <el-tag :type="isOnline ? 'success' : 'danger'" size="large">
            {{ isOnline ? '在线' : '离线' }}
          </el-tag>
          <span class="network-type">
            连接类型: {{ connectionType }}
          </span>
        </div>
      </div>

      <!-- 缓存信息 -->
      <div class="cache-info-section">
        <h4>缓存详情</h4>
        <div v-if="Object.keys(cacheInfo).length === 0" class="no-cache">
          <el-empty description="暂无缓存数据" />
        </div>
        <div v-else class="cache-list">
          <div 
            v-for="(info, cacheName) in cacheInfo" 
            :key="cacheName" 
            class="cache-item"
          >
            <div class="cache-header">
              <div class="cache-name">
                <el-icon><FolderOpened /></el-icon>
                <span>{{ cacheName }}</span>
              </div>
              <div class="cache-actions">
                <el-tag size="small">{{ info.count }} 项</el-tag>
                <el-button 
                  @click="viewCacheDetails(cacheName, info)" 
                  size="small" 
                  type="primary" 
                  text
                >
                  查看详情
                </el-button>
                <el-button 
                  @click="clearSpecificCache(cacheName)" 
                  size="small" 
                  type="danger" 
                  text
                >
                  清空
                </el-button>
              </div>
            </div>
            <div class="cache-size">
              <span>估计大小: {{ formatCacheSize(info.count) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 缓存策略配置 -->
      <div class="cache-strategy-section">
        <h4>缓存策略</h4>
        <div class="strategy-list">
          <div class="strategy-item">
            <span class="strategy-name">静态资源:</span>
            <el-tag size="small">缓存优先</el-tag>
            <span class="strategy-desc">CSS、JS文件优先从缓存加载</span>
          </div>
          <div class="strategy-item">
            <span class="strategy-name">API请求:</span>
            <el-tag size="small" type="warning">网络优先</el-tag>
            <span class="strategy-desc">优先从网络获取最新数据</span>
          </div>
          <div class="strategy-item">
            <span class="strategy-name">图片资源:</span>
            <el-tag size="small">缓存优先</el-tag>
            <span class="strategy-desc">图片文件优先从缓存加载</span>
          </div>
          <div class="strategy-item">
            <span class="strategy-name">字体文件:</span>
            <el-tag size="small">缓存优先</el-tag>
            <span class="strategy-desc">字体文件长期缓存</span>
          </div>
        </div>
      </div>

      <!-- 预缓存管理 -->
      <div class="precache-section">
        <h4>预缓存管理</h4>
        <div class="precache-actions">
          <el-input
            v-model="precacheUrl"
            placeholder="输入要预缓存的URL"
            style="width: 300px; margin-right: 12px;"
          />
          <el-button @click="addToPrecache" :loading="precaching" type="primary">
            添加到预缓存
          </el-button>
        </div>
        <div v-if="precachedUrls.length > 0" class="precached-list">
          <h5>已预缓存的URL:</h5>
          <ul>
            <li v-for="url in precachedUrls" :key="url">
              {{ url }}
            </li>
          </ul>
        </div>
      </div>
    </el-card>

    <!-- 缓存详情对话框 -->
    <el-dialog
      v-model="showCacheDetails"
      title="缓存详情"
      width="60%"
      :before-close="closeCacheDetails"
    >
      <div v-if="selectedCacheInfo" class="cache-details">
        <h4>{{ selectedCacheName }}</h4>
        <p>缓存项数量: {{ selectedCacheInfo.count }}</p>
        <div class="url-list">
          <h5>缓存的URL列表:</h5>
          <el-scrollbar height="300px">
            <ul>
              <li v-for="url in selectedCacheInfo.urls" :key="url" class="url-item">
                {{ url }}
              </li>
            </ul>
          </el-scrollbar>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { FolderOpened } from '@element-plus/icons-vue'
import { 
  getServiceWorkerManager, 
  type ServiceWorkerStatus, 
  type CacheInfo 
} from '@/utils/serviceWorker'

// 状态
const loading = ref(false)
const clearing = ref(false)
const precaching = ref(false)
const swStatus = ref<ServiceWorkerStatus>('unsupported')
const isOnline = ref(navigator.onLine)
const connectionType = ref('unknown')
const cacheInfo = reactive<CacheInfo>({})
const precacheUrl = ref('')
const precachedUrls = ref<string[]>([])

// 缓存详情对话框
const showCacheDetails = ref(false)
const selectedCacheName = ref('')
const selectedCacheInfo = ref<CacheInfo[string] | null>(null)

// Service Worker管理器
const swManager = getServiceWorkerManager()

// 生命周期
onMounted(() => {
  initCacheManager()
  detectConnectionType()
  setupNetworkListeners()
})

onUnmounted(() => {
  removeNetworkListeners()
})

// 初始化缓存管理器
const initCacheManager = async () => {
  if (swManager) {
    swStatus.value = swManager.getStatus()
    
    // 设置事件监听器
    swManager.on('statusChange', (status) => {
      swStatus.value = status
    })
    
    swManager.on('updateAvailable', () => {
      ElMessage.info('发现新版本，正在下载...')
    })
    
    swManager.on('updateReady', () => {
      ElMessage.success('新版本已准备就绪，点击激活更新')
    })
    
    swManager.on('online', () => {
      isOnline.value = true
      ElMessage.success('网络连接已恢复')
    })
    
    swManager.on('offline', () => {
      isOnline.value = false
      ElMessage.warning('网络连接已断开，将使用缓存数据')
    })
    
    // 获取缓存信息
    await refreshCacheInfo()
  }
}

// 刷新缓存信息
const refreshCacheInfo = async () => {
  if (!swManager) return
  
  loading.value = true
  try {
    const info = await swManager.getCacheInfo()
    if (info) {
      Object.assign(cacheInfo, info)
    }
  } catch (error) {
    console.error('Failed to get cache info:', error)
    ElMessage.error('获取缓存信息失败')
  } finally {
    loading.value = false
  }
}

// 清空所有缓存
const clearAllCache = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有缓存吗？这将删除所有离线数据。',
      '确认清空',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    clearing.value = true
    if (swManager) {
      await swManager.clearCache()
      Object.keys(cacheInfo).forEach(key => delete cacheInfo[key])
      ElMessage.success('所有缓存已清空')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to clear cache:', error)
      ElMessage.error('清空缓存失败')
    }
  } finally {
    clearing.value = false
  }
}

// 清空特定缓存
const clearSpecificCache = async (cacheName: string) => {
  try {
    await ElMessageBox.confirm(
      `确定要清空缓存 "${cacheName}" 吗？`,
      '确认清空',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    if (swManager) {
      await swManager.clearCache(cacheName)
      delete cacheInfo[cacheName]
      ElMessage.success(`缓存 "${cacheName}" 已清空`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to clear specific cache:', error)
      ElMessage.error('清空缓存失败')
    }
  }
}

// 检查更新
const checkForUpdate = async () => {
  if (!swManager) return
  
  try {
    const hasUpdate = await swManager.checkForUpdate()
    if (!hasUpdate) {
      ElMessage.info('当前已是最新版本')
    }
  } catch (error) {
    console.error('Failed to check for update:', error)
    ElMessage.error('检查更新失败')
  }
}

// 激活更新
const activateUpdate = async () => {
  if (!swManager) return
  
  try {
    await swManager.skipWaiting()
    ElMessage.success('更新已激活，页面将重新加载')
  } catch (error) {
    console.error('Failed to activate update:', error)
    ElMessage.error('激活更新失败')
  }
}

// 添加到预缓存
const addToPrecache = async () => {
  if (!precacheUrl.value.trim()) {
    ElMessage.warning('请输入有效的URL')
    return
  }
  
  if (!swManager) {
    ElMessage.error('Service Worker未就绪')
    return
  }
  
  precaching.value = true
  try {
    await swManager.precacheUrls([precacheUrl.value])
    precachedUrls.value.push(precacheUrl.value)
    precacheUrl.value = ''
    ElMessage.success('URL已添加到预缓存')
  } catch (error) {
    console.error('Failed to precache URL:', error)
    ElMessage.error('预缓存失败')
  } finally {
    precaching.value = false
  }
}

// 查看缓存详情
const viewCacheDetails = (cacheName: string | number, info: CacheInfo[string]) => {
  selectedCacheName.value = String(cacheName)
  selectedCacheInfo.value = info
  showCacheDetails.value = true
}

// 关闭缓存详情
const closeCacheDetails = () => {
  showCacheDetails.value = false
  selectedCacheName.value = ''
  selectedCacheInfo.value = null
}

// 获取状态类型
const getStatusType = (status: ServiceWorkerStatus) => {
  switch (status) {
    case 'installed':
    case 'updated':
      return 'success'
    case 'installing':
    case 'updating':
      return 'warning'
    case 'error':
      return 'danger'
    case 'unsupported':
      return 'info'
    default:
      return 'info'
  }
}

// 获取状态文本
const getStatusText = (status: ServiceWorkerStatus) => {
  switch (status) {
    case 'unsupported':
      return '不支持'
    case 'installing':
      return '安装中'
    case 'installed':
      return '已安装'
    case 'updating':
      return '更新中'
    case 'updated':
      return '有更新'
    case 'error':
      return '错误'
    default:
      return '未知'
  }
}

// 格式化缓存大小
const formatCacheSize = (count: number) => {
  // 估算每个缓存项平均大小为50KB
  const estimatedSize = count * 50 * 1024
  if (estimatedSize < 1024 * 1024) {
    return `${(estimatedSize / 1024).toFixed(1)} KB`
  } else {
    return `${(estimatedSize / (1024 * 1024)).toFixed(1)} MB`
  }
}

// 检测连接类型
const detectConnectionType = () => {
  if ('connection' in navigator) {
    const connection = (navigator as any).connection
    connectionType.value = connection.effectiveType || connection.type || 'unknown'
  }
}

// 设置网络监听器
const setupNetworkListeners = () => {
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
}

// 移除网络监听器
const removeNetworkListeners = () => {
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
}

// 处理上线事件
const handleOnline = () => {
  isOnline.value = true
}

// 处理离线事件
const handleOffline = () => {
  isOnline.value = false
}


</script>

<style scoped lang="scss">
@import "@/styles/variables.scss";

.cache-manager {
  .manager-card {
    max-width: 1000px;
    margin: 0 auto;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    h3 {
      margin: 0;
      color: $text-color-primary;
    }

    .header-actions {
      .el-button {
        margin-left: 8px;
      }
    }
  }

  .sw-status-section,
  .network-status-section,
  .cache-info-section,
  .cache-strategy-section,
  .precache-section {
    margin-bottom: 32px;
    padding-bottom: 24px;
    border-bottom: 1px solid $border-color-lighter;

    &:last-child {
      border-bottom: none;
    }

    h4 {
      color: $text-color-primary;
      margin-bottom: 16px;
      font-size: 16px;
      font-weight: 600;
    }
  }

  .status-info {
    display: flex;
    align-items: center;
    gap: 16px;

    .status-actions {
      .el-button {
        margin-left: 8px;
      }
    }
  }

  .network-info {
    display: flex;
    align-items: center;
    gap: 16px;

    .network-type {
      color: $text-color-regular;
      font-size: 14px;
    }
  }

  .cache-list {
    .cache-item {
      padding: 16px;
      background: $bg-color-page;
      border-radius: 8px;
      margin-bottom: 12px;
      border: 1px solid $border-color-lighter;

      .cache-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        .cache-name {
          display: flex;
          align-items: center;
          gap: 8px;
          font-weight: 500;
          color: $text-color-primary;
        }

        .cache-actions {
          display: flex;
          align-items: center;
          gap: 8px;
        }
      }

      .cache-size {
        color: $text-color-secondary;
        font-size: 12px;
      }
    }
  }

  .strategy-list {
    .strategy-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 8px 0;
      border-bottom: 1px solid $border-color-lighter;

      &:last-child {
        border-bottom: none;
      }

      .strategy-name {
        font-weight: 500;
        color: $text-color-primary;
        min-width: 80px;
      }

      .strategy-desc {
        color: $text-color-regular;
        font-size: 14px;
      }
    }
  }

  .precache-actions {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
  }

  .precached-list {
    h5 {
      margin-bottom: 8px;
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

  .cache-details {
    h4 {
      margin-bottom: 16px;
      color: $text-color-primary;
    }

    .url-list {
      h5 {
        margin-bottom: 12px;
        color: $text-color-primary;
      }

      ul {
        margin: 0;
        padding: 0;
        list-style: none;

        .url-item {
          padding: 8px 12px;
          background: $bg-color-page;
          border-radius: 4px;
          margin-bottom: 4px;
          font-family: monospace;
          font-size: 12px;
          color: $text-color-regular;
          word-break: break-all;
        }
      }
    }
  }
}

// 移动端适配
@media (max-width: 768px) {
  .cache-manager {
    .card-header {
      flex-direction: column;
      gap: 12px;
      align-items: stretch;

      .header-actions {
        display: flex;
        gap: 8px;

        .el-button {
          flex: 1;
          margin-left: 0;
        }
      }
    }

    .status-info,
    .network-info {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;
    }

    .cache-item .cache-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;
    }

    .strategy-item {
      flex-direction: column;
      align-items: flex-start;
      gap: 4px;
    }

    .precache-actions {
      flex-direction: column;
      gap: 8px;

      .el-input {
        width: 100% !important;
      }
    }
  }
}
</style>
