<template>
  <div class="pwa-test">
    <div class="test-header">
      <h1>PWA 功能测试</h1>
      <p>测试 Progressive Web App 的各项功能</p>
    </div>

    <div class="test-sections">
      <!-- 基础功能测试 -->
      <el-card class="test-card">
        <template #header>
          <span>基础功能测试</span>
        </template>

        <div class="test-items">
          <div class="test-item">
            <div class="test-info">
              <strong>Service Worker 状态</strong>
              <p>检查 Service Worker 是否正常运行</p>
            </div>
            <div class="test-result">
              <el-tag :type="swStatus.registered ? 'success' : 'danger'">
                {{ swStatus.registered ? '已注册' : '未注册' }}
              </el-tag>
            </div>
          </div>

          <div class="test-item">
            <div class="test-info">
              <strong>离线功能</strong>
              <p>测试应用的离线访问能力</p>
            </div>
            <div class="test-result">
              <el-button @click="testOfflineMode" :loading="testing.offline">
                测试离线模式
              </el-button>
            </div>
          </div>

          <div class="test-item">
            <div class="test-info">
              <strong>缓存功能</strong>
              <p>测试资源缓存和数据缓存</p>
            </div>
            <div class="test-result">
              <el-button @click="testCaching" :loading="testing.cache">
                测试缓存
              </el-button>
            </div>
          </div>

          <div class="test-item">
            <div class="test-info">
              <strong>安装提示</strong>
              <p>测试 PWA 安装提示功能</p>
            </div>
            <div class="test-result">
              <el-button @click="testInstallPrompt" :loading="testing.install">
                测试安装提示
              </el-button>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 离线数据测试 -->
      <el-card class="test-card">
        <template #header>
          <span>离线数据测试</span>
        </template>

        <div class="offline-test">
          <div class="test-item">
            <div class="test-info">
              <strong>创建离线订单</strong>
              <p>模拟在离线状态下创建订单</p>
            </div>
            <div class="test-actions">
              <el-button @click="createTestOrder" :loading="testing.order">
                创建测试订单
              </el-button>
            </div>
          </div>

          <div class="test-item">
            <div class="test-info">
              <strong>数据同步</strong>
              <p>测试离线数据的同步功能</p>
            </div>
            <div class="test-actions">
              <el-button @click="testDataSync" :loading="testing.sync">
                测试数据同步
              </el-button>
            </div>
          </div>

          <div class="offline-orders" v-if="offlineOrders.length > 0">
            <h4>离线订单列表</h4>
            <div class="order-list">
              <div 
                class="order-item" 
                v-for="order in offlineOrders" 
                :key="order.id"
              >
                <div class="order-info">
                  <div class="order-id">订单ID: {{ order.id }}</div>
                  <div class="order-details">
                    商品: {{ order.productId }} | 数量: {{ order.quantity }} | 
                    金额: ¥{{ order.totalPrice }}
                  </div>
                  <div class="order-status">
                    状态: {{ order.status }} | 
                    同步: {{ order.synced ? '已同步' : '待同步' }}
                  </div>
                </div>
                <div class="order-actions">
                  <el-button 
                    size="small" 
                    type="primary" 
                    @click="syncOrder(order.id)"
                    :disabled="order.synced"
                  >
                    同步
                  </el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 通知测试 -->
      <el-card class="test-card">
        <template #header>
          <span>通知功能测试</span>
        </template>

        <div class="notification-test">
          <div class="test-item">
            <div class="test-info">
              <strong>通知权限</strong>
              <p>检查和请求通知权限</p>
            </div>
            <div class="test-result">
              <el-tag :type="getNotificationPermissionType()">
                {{ notificationPermission }}
              </el-tag>
              <el-button 
                v-if="notificationPermission === 'default'"
                @click="requestNotificationPermission"
                size="small"
              >
                请求权限
              </el-button>
            </div>
          </div>

          <div class="test-item">
            <div class="test-info">
              <strong>发送测试通知</strong>
              <p>发送一个测试通知</p>
            </div>
            <div class="test-actions">
              <el-button 
                @click="sendTestNotification" 
                :loading="testing.notification"
                :disabled="notificationPermission !== 'granted'"
              >
                发送通知
              </el-button>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 测试结果 -->
      <el-card class="test-card" v-if="testResults.length > 0">
        <template #header>
          <span>测试结果</span>
        </template>

        <div class="test-results">
          <div 
            class="result-item" 
            v-for="(result, index) in testResults" 
            :key="index"
            :class="{ 'success': result.success, 'error': !result.success }"
          >
            <div class="result-icon">
              <el-icon v-if="result.success" color="#52c41a">
                <Check />
              </el-icon>
              <el-icon v-else color="#ff4d4f">
                <Close />
              </el-icon>
            </div>
            <div class="result-content">
              <div class="result-title">{{ result.title }}</div>
              <div class="result-message">{{ result.message }}</div>
              <div class="result-time">{{ formatTime(result.timestamp) }}</div>
            </div>
          </div>
        </div>

        <div class="results-actions">
          <el-button @click="clearResults">清空结果</el-button>
          <el-button type="primary" @click="exportResults">导出结果</el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElCard, ElButton, ElTag, ElIcon, ElMessage } from 'element-plus'
import { Check, Close } from '@element-plus/icons-vue'
import { pwaManager } from '@/utils/pwa'
import { useOfflineSync } from '@/composables/useOfflineSync'

// 响应式数据
const swStatus = ref({ registered: false, active: false })
const notificationPermission = ref(Notification.permission)
const offlineOrders = ref<any[]>([])
const testResults = ref<any[]>([])

const testing = ref({
  offline: false,
  cache: false,
  install: false,
  order: false,
  sync: false,
  notification: false
})

// 使用离线同步功能
const { 
  createOfflineOrder, 
  syncPendingData, 
  getUserOrderHistory,
  isOnline 
} = useOfflineSync()

// 方法
const addTestResult = (title: string, message: string, success: boolean) => {
  testResults.value.unshift({
    title,
    message,
    success,
    timestamp: Date.now()
  })
}

const testOfflineMode = async () => {
  testing.value.offline = true
  
  try {
    // 模拟离线状态
    const originalOnLine = navigator.onLine
    Object.defineProperty(navigator, 'onLine', {
      writable: true,
      value: false
    })
    
    // 测试离线页面访问
    const response = await fetch('/offline.html')
    const success = response.ok
    
    addTestResult(
      '离线模式测试',
      success ? '离线页面可以正常访问' : '离线页面访问失败',
      success
    )
    
    // 恢复在线状态
    Object.defineProperty(navigator, 'onLine', {
      writable: true,
      value: originalOnLine
    })
    
  } catch (error) {
    addTestResult('离线模式测试', `测试失败: ${error}`, false)
  } finally {
    testing.value.offline = false
  }
}

const testCaching = async () => {
  testing.value.cache = true
  
  try {
    const cacheInfo = await pwaManager.getCacheInfo()
    const cacheCount = Object.keys(cacheInfo).length
    
    addTestResult(
      '缓存功能测试',
      `发现 ${cacheCount} 个缓存，缓存功能正常`,
      cacheCount > 0
    )
  } catch (error) {
    addTestResult('缓存功能测试', `测试失败: ${error}`, false)
  } finally {
    testing.value.cache = false
  }
}

const testInstallPrompt = async () => {
  testing.value.install = true
  
  try {
    const status = pwaManager.getStatus()
    
    if (status.isInstallable) {
      addTestResult('安装提示测试', '应用可以安装，安装提示功能正常', true)
    } else if (status.isInstalled) {
      addTestResult('安装提示测试', '应用已安装', true)
    } else {
      addTestResult('安装提示测试', '应用不可安装或已在独立模式运行', false)
    }
  } catch (error) {
    addTestResult('安装提示测试', `测试失败: ${error}`, false)
  } finally {
    testing.value.install = false
  }
}

const createTestOrder = async () => {
  testing.value.order = true
  
  try {
    const testOrderData = {
      activityId: 'test_activity_' + Date.now(),
      productId: 'test_product_' + Date.now(),
      quantity: 1,
      totalPrice: 99.99
    }
    
    const order = await createOfflineOrder(testOrderData)
    
    addTestResult(
      '创建离线订单',
      `成功创建离线订单: ${order.id}`,
      true
    )
    
    await loadOfflineOrders()
  } catch (error) {
    addTestResult('创建离线订单', `创建失败: ${error}`, false)
  } finally {
    testing.value.order = false
  }
}

const testDataSync = async () => {
  testing.value.sync = true
  
  try {
    await syncPendingData()
    
    addTestResult(
      '数据同步测试',
      '数据同步完成',
      true
    )
    
    await loadOfflineOrders()
  } catch (error) {
    addTestResult('数据同步测试', `同步失败: ${error}`, false)
  } finally {
    testing.value.sync = false
  }
}

const syncOrder = async (orderId: string) => {
  try {
    await syncPendingData()
    await loadOfflineOrders()
    ElMessage.success('订单同步成功')
  } catch (error) {
    ElMessage.error('订单同步失败')
  }
}

const requestNotificationPermission = async () => {
  try {
    const permission = await Notification.requestPermission()
    notificationPermission.value = permission
    
    addTestResult(
      '通知权限请求',
      `权限状态: ${permission}`,
      permission === 'granted'
    )
  } catch (error) {
    addTestResult('通知权限请求', `请求失败: ${error}`, false)
  }
}

const sendTestNotification = async () => {
  testing.value.notification = true
  
  try {
    if (notificationPermission.value !== 'granted') {
      throw new Error('没有通知权限')
    }
    
    const notification = new Notification('Flash Sku 测试通知', {
      body: '这是一个测试通知，PWA 通知功能正常工作！',
      icon: '/icons/icon-192x192.png',
      badge: '/icons/badge-72x72.png'
    })
    
    notification.onclick = () => {
      notification.close()
    }
    
    addTestResult(
      '发送测试通知',
      '测试通知发送成功',
      true
    )
  } catch (error) {
    addTestResult('发送测试通知', `发送失败: ${error}`, false)
  } finally {
    testing.value.notification = false
  }
}

const loadOfflineOrders = async () => {
  try {
    const orders = await getUserOrderHistory()
    offlineOrders.value = orders
  } catch (error) {
    console.error('Failed to load offline orders:', error)
  }
}

const getNotificationPermissionType = () => {
  switch (notificationPermission.value) {
    case 'granted': return 'success'
    case 'denied': return 'danger'
    default: return 'warning'
  }
}

const formatTime = (timestamp: number) => {
  return new Date(timestamp).toLocaleTimeString()
}

const clearResults = () => {
  testResults.value = []
}

const exportResults = () => {
  const data = {
    timestamp: new Date().toISOString(),
    results: testResults.value,
    environment: {
      userAgent: navigator.userAgent,
      online: navigator.onLine,
      serviceWorker: 'serviceWorker' in navigator,
      notifications: 'Notification' in window
    }
  }
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `pwa-test-results-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
}

// 生命周期
onMounted(async () => {
  // 检查 Service Worker 状态
  if ('serviceWorker' in navigator) {
    const registration = await navigator.serviceWorker.getRegistration()
    swStatus.value = {
      registered: !!registration,
      active: !!registration?.active
    }
  }
  
  // 加载离线订单
  await loadOfflineOrders()
})
</script>

<style scoped>
.pwa-test {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

.test-header {
  margin-bottom: 24px;
  text-align: center;
}

.test-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 600;
}

.test-header p {
  margin: 0;
  color: #666;
  font-size: 16px;
}

.test-sections {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.test-card {
  width: 100%;
}

.test-items {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.test-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  background: #fafafa;
}

.test-info {
  flex: 1;
}

.test-info strong {
  display: block;
  margin-bottom: 4px;
  font-size: 16px;
}

.test-info p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.test-result,
.test-actions {
  flex-shrink: 0;
  margin-left: 16px;
}

.offline-orders {
  margin-top: 20px;
}

.offline-orders h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
}

.order-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.order-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  background: white;
}

.order-info {
  flex: 1;
}

.order-id {
  font-weight: 500;
  margin-bottom: 4px;
}

.order-details,
.order-status {
  font-size: 12px;
  color: #666;
  margin-bottom: 2px;
}

.order-actions {
  flex-shrink: 0;
  margin-left: 12px;
}

.test-results {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}

.result-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border-radius: 6px;
  border: 1px solid #e8e8e8;
}

.result-item.success {
  background: #f6ffed;
  border-color: #b7eb8f;
}

.result-item.error {
  background: #fff2f0;
  border-color: #ffccc7;
}

.result-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.result-content {
  flex: 1;
}

.result-title {
  font-weight: 500;
  margin-bottom: 4px;
}

.result-message {
  color: #666;
  font-size: 14px;
  margin-bottom: 4px;
}

.result-time {
  color: #999;
  font-size: 12px;
}

.results-actions {
  margin-top: 16px;
  text-align: center;
  display: flex;
  gap: 12px;
  justify-content: center;
}

@media (max-width: 768px) {
  .pwa-test {
    padding: 16px;
  }
  
  .test-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .test-result,
  .test-actions {
    margin-left: 0;
    width: 100%;
  }
  
  .order-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .order-actions {
    margin-left: 0;
    width: 100%;
  }
  
  .results-actions {
    flex-direction: column;
  }
}
</style>
