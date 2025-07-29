<template>
  <div class="pwa-settings">
    <div class="settings-header">
      <h1>PWA 设置</h1>
      <p>管理应用的离线功能和缓存设置</p>
    </div>

    <div class="settings-content">
      <!-- PWA 状态 -->
      <el-card class="status-card">
        <template #header>
          <div class="card-header">
            <span>应用状态</span>
            <el-tag :type="pwaStatus.isInstalled ? 'success' : 'info'">
              {{ pwaStatus.isInstalled ? '已安装' : '未安装' }}
            </el-tag>
          </div>
        </template>

        <div class="status-grid">
          <div class="status-item">
            <div class="status-icon">
              <el-icon :size="24" :color="pwaStatus.isStandalone ? '#52c41a' : '#d9d9d9'">
                <Monitor />
              </el-icon>
            </div>
            <div class="status-info">
              <div class="status-title">独立模式</div>
              <div class="status-desc">
                {{ pwaStatus.isStandalone ? '正在独立模式运行' : '在浏览器中运行' }}
              </div>
            </div>
          </div>

          <div class="status-item">
            <div class="status-icon">
              <el-icon :size="24" :color="pwaStatus.hasServiceWorker ? '#52c41a' : '#d9d9d9'">
                <Setting />
              </el-icon>
            </div>
            <div class="status-info">
              <div class="status-title">Service Worker</div>
              <div class="status-desc">
                {{ pwaStatus.hasServiceWorker ? '已启用' : '不支持' }}
              </div>
            </div>
          </div>

          <div class="status-item">
            <div class="status-icon">
              <el-icon :size="24" :color="pwaStatus.isOnline ? '#52c41a' : '#ff4d4f'">
                <Connection />
              </el-icon>
            </div>
            <div class="status-info">
              <div class="status-title">网络状态</div>
              <div class="status-desc">
                {{ pwaStatus.isOnline ? '在线' : '离线' }}
              </div>
            </div>
          </div>

          <div class="status-item">
            <div class="status-icon">
              <el-icon :size="24" :color="pwaStatus.isInstallable ? '#1890ff' : '#d9d9d9'">
                <Download />
              </el-icon>
            </div>
            <div class="status-info">
              <div class="status-title">可安装性</div>
              <div class="status-desc">
                {{ pwaStatus.isInstallable ? '可以安装' : '不可安装' }}
              </div>
            </div>
          </div>
        </div>

        <div class="status-actions" v-if="pwaStatus.isInstallable">
          <el-button type="primary" @click="handleInstall" :loading="installing">
            安装应用
          </el-button>
        </div>
      </el-card>

      <!-- 缓存管理 -->
      <el-card class="cache-card">
        <template #header>
          <div class="card-header">
            <span>缓存管理</span>
            <el-button size="small" @click="refreshCacheInfo">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </template>

        <div class="cache-info">
          <div class="cache-usage">
            <div class="usage-title">存储使用情况</div>
            <div class="usage-bar">
              <el-progress 
                :percentage="storageUsagePercent" 
                :color="getUsageColor(storageUsagePercent)"
                :show-text="false"
              />
            </div>
            <div class="usage-text">
              {{ formatBytes(storageUsage.used) }} / {{ formatBytes(storageUsage.quota) }}
            </div>
          </div>

          <div class="cache-list">
            <div class="cache-item" v-for="(cache, name) in cacheInfo" :key="name">
              <div class="cache-name">{{ name }}</div>
              <div class="cache-count">{{ cache.count }} 项</div>
              <el-button 
                size="small" 
                type="danger" 
                text 
                @click="clearCache(name)"
                :loading="clearingCache === name"
              >
                清空
              </el-button>
            </div>
          </div>
        </div>

        <div class="cache-actions">
          <el-button @click="clearAllCache" :loading="clearingAllCache">
            清空所有缓存
          </el-button>
          <el-button type="primary" @click="updateApp" :loading="updating">
            检查更新
          </el-button>
        </div>
      </el-card>

      <!-- 离线设置 -->
      <el-card class="offline-card">
        <template #header>
          <span>离线设置</span>
        </template>

        <div class="offline-settings">
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">离线数据同步</div>
              <div class="setting-desc">当网络恢复时自动同步离线期间的操作</div>
            </div>
            <el-switch v-model="offlineSync" @change="updateOfflineSync" />
          </div>

          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">后台同步</div>
              <div class="setting-desc">在后台自动同步数据，即使应用未打开</div>
            </div>
            <el-switch v-model="backgroundSync" @change="updateBackgroundSync" />
          </div>

          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">推送通知</div>
              <div class="setting-desc">接收秒杀活动和订单状态的推送通知</div>
            </div>
            <el-switch v-model="pushNotifications" @change="updatePushNotifications" />
          </div>
        </div>

        <div class="sync-status" v-if="pendingSync > 0">
          <el-alert
            :title="`有 ${pendingSync} 项数据待同步`"
            type="warning"
            :closable="false"
          >
            <template #default>
              <el-button size="small" type="primary" @click="syncNow" :loading="syncing">
                立即同步
              </el-button>
            </template>
          </el-alert>
        </div>
      </el-card>

      <!-- 设备信息 -->
      <el-card class="device-card">
        <template #header>
          <span>设备信息</span>
        </template>

        <div class="device-info">
          <div class="info-item">
            <span class="info-label">用户代理:</span>
            <span class="info-value">{{ deviceInfo.userAgent }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">平台:</span>
            <span class="info-value">{{ deviceInfo.platform }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">语言:</span>
            <span class="info-value">{{ deviceInfo.language }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">屏幕分辨率:</span>
            <span class="info-value">{{ deviceInfo.screen.width }} × {{ deviceInfo.screen.height }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">视口大小:</span>
            <span class="info-value">{{ deviceInfo.viewport.width }} × {{ deviceInfo.viewport.height }}</span>
          </div>
        </div>
      </el-card>

      <!-- 功能支持 -->
      <el-card class="features-card">
        <template #header>
          <span>功能支持</span>
        </template>

        <div class="features-grid">
          <div 
            class="feature-item" 
            v-for="(supported, feature) in featureSupport" 
            :key="feature"
          >
            <el-icon :size="20" :color="supported ? '#52c41a' : '#ff4d4f'">
              <Check v-if="supported" />
              <Close v-else />
            </el-icon>
            <span class="feature-name">{{ getFeatureName(feature) }}</span>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { 
  ElCard, ElTag, ElIcon, ElButton, ElProgress, ElSwitch, ElAlert, ElMessage 
} from 'element-plus'
import {
  Monitor, Setting, Connection, Download, Refresh, Check, Close
} from '@element-plus/icons-vue'
import { pwaManager, getPWAStatus } from '@/utils/pwa'
import { offlineStorage, offlineSync as offlineSyncManager } from '@/utils/offlineStorage'

// 响应式数据
const pwaStatus = ref(getPWAStatus())
const installing = ref(false)
const cacheInfo = ref<any>({})
const storageUsage = ref({ used: 0, quota: 0 })
const clearingCache = ref<string | null>(null)
const clearingAllCache = ref(false)
const updating = ref(false)
const offlineSync = ref(true)
const backgroundSync = ref(false)
const pushNotifications = ref(false)
const pendingSync = ref(0)
const syncing = ref(false)
const deviceInfo = ref(pwaManager.getDeviceInfo())
const featureSupport = ref(pwaManager.getFeatureSupport())

// 计算属性
const storageUsagePercent = computed(() => {
  if (storageUsage.value.quota === 0) return 0
  return Math.round((storageUsage.value.used / storageUsage.value.quota) * 100)
})

// 方法
const handleInstall = async () => {
  installing.value = true
  try {
    const success = await pwaManager.showInstallPrompt()
    if (success) {
      ElMessage.success('应用安装成功！')
      pwaStatus.value = getPWAStatus()
    }
  } catch (error) {
    ElMessage.error('安装失败，请稍后重试')
  } finally {
    installing.value = false
  }
}

const refreshCacheInfo = async () => {
  try {
    const [cache, usage] = await Promise.all([
      pwaManager.getCacheInfo(),
      offlineStorage.getStorageUsage()
    ])
    cacheInfo.value = cache
    storageUsage.value = usage
  } catch (error) {
    console.error('Failed to refresh cache info:', error)
  }
}

const clearCache = async (cacheName: string) => {
  clearingCache.value = cacheName
  try {
    await pwaManager.clearCache(cacheName)
    await refreshCacheInfo()
    ElMessage.success('缓存已清空')
  } catch (error) {
    ElMessage.error('清空缓存失败')
  } finally {
    clearingCache.value = null
  }
}

const clearAllCache = async () => {
  clearingAllCache.value = true
  try {
    await pwaManager.clearCache()
    await refreshCacheInfo()
    ElMessage.success('所有缓存已清空')
  } catch (error) {
    ElMessage.error('清空缓存失败')
  } finally {
    clearingAllCache.value = false
  }
}

const updateApp = async () => {
  updating.value = true
  try {
    await pwaManager.updateServiceWorker()
    ElMessage.success('应用已更新')
  } catch (error) {
    ElMessage.info('当前已是最新版本')
  } finally {
    updating.value = false
  }
}

const updateOfflineSync = (value: string | number | boolean) => {
  localStorage.setItem('offline-sync-enabled', value.toString())
}

const updateBackgroundSync = (value: string | number | boolean) => {
  localStorage.setItem('background-sync-enabled', value.toString())
}

const updatePushNotifications = async (value: string | number | boolean) => {
  if (Boolean(value) && 'Notification' in window) {
    const permission = await Notification.requestPermission()
    if (permission !== 'granted') {
      pushNotifications.value = false
      ElMessage.warning('需要授权通知权限')
      return
    }
  }
  localStorage.setItem('push-notifications-enabled', Boolean(value).toString())
}

const syncNow = async () => {
  syncing.value = true
  try {
    await offlineSyncManager.startSync()
    await checkPendingSync()
    ElMessage.success('同步完成')
  } catch (error) {
    ElMessage.error('同步失败')
  } finally {
    syncing.value = false
  }
}

const checkPendingSync = async () => {
  try {
    const [orders, queue] = await Promise.all([
      offlineStorage.getUnsyncedOrders(),
      offlineStorage.getSyncQueue()
    ])
    pendingSync.value = orders.length + queue.length
  } catch (error) {
    console.error('Failed to check pending sync:', error)
  }
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getUsageColor = (percent: number): string => {
  if (percent < 50) return '#52c41a'
  if (percent < 80) return '#fa8c16'
  return '#ff4d4f'
}

const getFeatureName = (feature: string): string => {
  const names: Record<string, string> = {
    serviceWorker: 'Service Worker',
    pushManager: '推送管理',
    notification: '通知',
    backgroundSync: '后台同步',
    webShare: 'Web 分享',
    webShareTarget: '分享目标',
    badging: '应用徽章',
    periodicBackgroundSync: '定期后台同步',
    webLocks: 'Web 锁',
    wakeLock: '唤醒锁'
  }
  return names[feature] || feature
}

// 生命周期
onMounted(() => {
  refreshCacheInfo()
  checkPendingSync()
  
  // 加载设置
  offlineSync.value = localStorage.getItem('offline-sync-enabled') !== 'false'
  backgroundSync.value = localStorage.getItem('background-sync-enabled') === 'true'
  pushNotifications.value = localStorage.getItem('push-notifications-enabled') === 'true'
})
</script>

<style scoped>
.pwa-settings {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.settings-header {
  margin-bottom: 24px;
}

.settings-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.settings-header p {
  margin: 0;
  color: #666;
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
}

.status-icon {
  flex-shrink: 0;
}

.status-info {
  flex: 1;
}

.status-title {
  font-weight: 500;
  margin-bottom: 4px;
}

.status-desc {
  font-size: 12px;
  color: #666;
}

.status-actions {
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.cache-info {
  margin-bottom: 20px;
}

.cache-usage {
  margin-bottom: 20px;
}

.usage-title {
  font-weight: 500;
  margin-bottom: 8px;
}

.usage-bar {
  margin-bottom: 8px;
}

.usage-text {
  font-size: 12px;
  color: #666;
}

.cache-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cache-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: #fafafa;
  border-radius: 4px;
}

.cache-name {
  font-weight: 500;
}

.cache-count {
  font-size: 12px;
  color: #666;
}

.cache-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.offline-settings {
  margin-bottom: 20px;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-info {
  flex: 1;
}

.setting-title {
  font-weight: 500;
  margin-bottom: 4px;
}

.setting-desc {
  font-size: 12px;
  color: #666;
}

.sync-status {
  margin-top: 16px;
}

.device-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.info-label {
  font-weight: 500;
  min-width: 100px;
}

.info-value {
  color: #666;
  word-break: break-all;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border: 1px solid #f0f0f0;
  border-radius: 4px;
}

.feature-name {
  font-size: 14px;
}

@media (max-width: 768px) {
  .pwa-settings {
    padding: 16px;
  }
  
  .status-grid {
    grid-template-columns: 1fr;
  }
  
  .cache-actions {
    flex-direction: column;
  }
  
  .features-grid {
    grid-template-columns: 1fr;
  }
}
</style>
