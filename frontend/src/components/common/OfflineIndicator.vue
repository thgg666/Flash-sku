<template>
  <Teleport to="body">
    <Transition name="slide-down">
      <div 
        v-if="showIndicator" 
        class="offline-indicator"
        :class="{ 
          'offline': !isOnline,
          'syncing': isSyncing,
          'has-pending': hasPendingSync
        }"
      >
        <div class="indicator-content">
          <div class="indicator-icon">
            <el-icon v-if="!isOnline" :size="16">
              <Close />
            </el-icon>
            <el-icon v-else-if="isSyncing" :size="16" class="spinning">
              <Loading />
            </el-icon>
            <el-icon v-else-if="hasPendingSync" :size="16">
              <Clock />
            </el-icon>
            <el-icon v-else :size="16">
              <Connection />
            </el-icon>
          </div>
          
          <div class="indicator-text">
            <span v-if="!isOnline">离线模式</span>
            <span v-else-if="isSyncing">正在同步...</span>
            <span v-else-if="hasPendingSync">有 {{ pendingCount }} 项待同步</span>
            <span v-else>已连接</span>
          </div>
          
          <div class="indicator-actions" v-if="isOnline && hasPendingSync">
            <el-button 
              size="small" 
              text 
              @click="handleSync"
              :loading="isSyncing"
            >
              立即同步
            </el-button>
          </div>
          
          <button 
            class="indicator-close" 
            @click="hideIndicator"
            v-if="isOnline && !isSyncing"
          >
            <el-icon :size="12">
              <Close />
            </el-icon>
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElButton, ElIcon, ElMessage } from 'element-plus'
import { Connection, Loading, Clock, Close } from '@element-plus/icons-vue'
import { offlineSync, offlineStorage } from '@/utils/offlineStorage'

// 响应式数据
const isOnline = ref(navigator.onLine)
const isSyncing = ref(false)
const pendingCount = ref(0)
const showIndicator = ref(false)
const autoHideTimer = ref<number | null>(null)

// 计算属性
const hasPendingSync = computed(() => pendingCount.value > 0)

// 检查待同步项目数量
const checkPendingSync = async () => {
  try {
    const [unsyncedOrders, syncQueue] = await Promise.all([
      offlineStorage.getUnsyncedOrders(),
      offlineStorage.getSyncQueue()
    ])
    
    pendingCount.value = unsyncedOrders.length + syncQueue.length
  } catch (error) {
    console.error('Failed to check pending sync:', error)
  }
}

// 更新在线状态
const updateOnlineStatus = (online: boolean) => {
  isOnline.value = online
  
  if (online) {
    showIndicator.value = true
    checkPendingSync()
    
    // 自动开始同步
    if (hasPendingSync.value) {
      handleSync()
    } else {
      // 如果没有待同步项，3秒后自动隐藏
      autoHideAfterDelay(3000)
    }
  } else {
    showIndicator.value = true
    clearAutoHideTimer()
  }
}

// 处理同步
const handleSync = async () => {
  if (isSyncing.value || !isOnline.value) {
    return
  }

  isSyncing.value = true
  clearAutoHideTimer()

  try {
    await offlineSync.startSync()
    await checkPendingSync()
    
    ElMessage.success('数据同步完成')
    
    // 同步完成后延迟隐藏
    autoHideAfterDelay(2000)
  } catch (error) {
    console.error('Sync failed:', error)
    ElMessage.error('同步失败，请稍后重试')
  } finally {
    isSyncing.value = false
  }
}

// 隐藏指示器
const hideIndicator = () => {
  showIndicator.value = false
  clearAutoHideTimer()
}

// 延迟自动隐藏
const autoHideAfterDelay = (delay: number) => {
  clearAutoHideTimer()
  autoHideTimer.value = window.setTimeout(() => {
    if (isOnline.value && !isSyncing.value && !hasPendingSync.value) {
      hideIndicator()
    }
  }, delay)
}

// 清除自动隐藏定时器
const clearAutoHideTimer = () => {
  if (autoHideTimer.value) {
    clearTimeout(autoHideTimer.value)
    autoHideTimer.value = null
  }
}

// 网络状态监听器
const handleOnline = () => updateOnlineStatus(true)
const handleOffline = () => updateOnlineStatus(false)

// 生命周期
onMounted(() => {
  // 添加网络状态监听器
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
  
  // 初始检查
  checkPendingSync()
  
  // 如果离线，显示指示器
  if (!isOnline.value) {
    showIndicator.value = true
  }
  
  // 定期检查待同步项目
  const checkInterval = setInterval(checkPendingSync, 30000) // 每30秒检查一次
  
  onUnmounted(() => {
    clearInterval(checkInterval)
  })
})

onUnmounted(() => {
  // 移除事件监听器
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
  
  // 清除定时器
  clearAutoHideTimer()
})
</script>

<style scoped>
.offline-indicator {
  position: fixed;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9998;
  max-width: 400px;
  width: calc(100% - 40px);
  margin-top: 20px;
}

.indicator-content {
  background: white;
  border-radius: 8px;
  padding: 12px 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border: 1px solid #e1e5e9;
  display: flex;
  align-items: center;
  gap: 12px;
  position: relative;
}

.offline .indicator-content {
  background: #fff2f0;
  border-color: #ffccc7;
}

.syncing .indicator-content {
  background: #f6ffed;
  border-color: #b7eb8f;
}

.has-pending .indicator-content {
  background: #fff7e6;
  border-color: #ffd591;
}

.indicator-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
}

.offline .indicator-icon {
  background: #ff4d4f;
  color: white;
}

.syncing .indicator-icon {
  background: #52c41a;
  color: white;
}

.has-pending .indicator-icon {
  background: #fa8c16;
  color: white;
}

.indicator-icon.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.indicator-text {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: #262626;
}

.offline .indicator-text {
  color: #a8071a;
}

.syncing .indicator-text {
  color: #389e0d;
}

.has-pending .indicator-text {
  color: #d46b08;
}

.indicator-actions {
  flex-shrink: 0;
}

.indicator-close {
  position: absolute;
  top: 8px;
  right: 8px;
  background: none;
  border: none;
  cursor: pointer;
  color: #8c8c8c;
  padding: 2px;
  border-radius: 2px;
  transition: all 0.2s;
}

.indicator-close:hover {
  background: #f5f5f5;
  color: #595959;
}

/* 动画 */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
}

.slide-down-enter-from {
  transform: translateX(-50%) translateY(-100%);
  opacity: 0;
}

.slide-down-leave-to {
  transform: translateX(-50%) translateY(-100%);
  opacity: 0;
}

/* 响应式 */
@media (max-width: 480px) {
  .offline-indicator {
    width: calc(100% - 20px);
    margin-top: 10px;
  }
  
  .indicator-content {
    padding: 10px 12px;
    font-size: 13px;
  }
  
  .indicator-text {
    font-size: 13px;
  }
}
</style>
