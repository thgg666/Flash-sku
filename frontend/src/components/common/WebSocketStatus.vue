<template>
  <div class="websocket-status" :class="statusClass">
    <el-tooltip :content="tooltipText" placement="bottom">
      <div class="status-indicator" @click="handleClick">
        <div class="status-dot" :class="dotClass"></div>
        <span v-if="showText" class="status-text">{{ statusText }}</span>
      </div>
    </el-tooltip>
    
    <!-- 详细状态弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="连接状态"
      width="400px"
      center
    >
      <div class="status-detail">
        <div class="detail-item">
          <span class="label">连接状态：</span>
          <el-tag :type="tagType" size="small">{{ statusText }}</el-tag>
        </div>
        <div class="detail-item">
          <span class="label">服务器地址：</span>
          <span class="value">{{ serverUrl }}</span>
        </div>
        <div class="detail-item">
          <span class="label">连接时间：</span>
          <span class="value">{{ connectionTime || '未连接' }}</span>
        </div>
        <div class="detail-item">
          <span class="label">重连次数：</span>
          <span class="value">{{ reconnectCount }}</span>
        </div>
        <div v-if="lastError" class="detail-item">
          <span class="label">最后错误：</span>
          <span class="value error">{{ lastError }}</span>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="detailVisible = false">关闭</el-button>
          <el-button 
            v-if="!isConnected" 
            type="primary" 
            @click="handleReconnect"
            :loading="isConnecting"
          >
            重新连接
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useWebSocket } from '@/composables/useWebSocket'
import { WebSocketState } from '@/utils/websocket'
import { formatDateTime } from '@/utils'

interface Props {
  showText?: boolean
  clickable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showText: false,
  clickable: true,
})

// WebSocket状态
const { state, isConnected, isConnecting, isReconnecting, hasError, reconnect } = useWebSocket()

// 状态
const detailVisible = ref(false)
const connectionTime = ref<string>('')
const reconnectCount = ref(0)
const lastError = ref<string>('')

// 计算属性
const statusClass = computed(() => {
  return `status-${state.value}`
})

const dotClass = computed(() => {
  switch (state.value) {
    case WebSocketState.CONNECTED:
      return 'connected'
    case WebSocketState.CONNECTING:
    case WebSocketState.RECONNECTING:
      return 'connecting'
    case WebSocketState.ERROR:
      return 'error'
    default:
      return 'disconnected'
  }
})

const statusText = computed(() => {
  switch (state.value) {
    case WebSocketState.CONNECTED:
      return '已连接'
    case WebSocketState.CONNECTING:
      return '连接中'
    case WebSocketState.RECONNECTING:
      return '重连中'
    case WebSocketState.ERROR:
      return '连接错误'
    default:
      return '未连接'
  }
})

const tagType = computed(() => {
  switch (state.value) {
    case WebSocketState.CONNECTED:
      return 'success'
    case WebSocketState.CONNECTING:
    case WebSocketState.RECONNECTING:
      return 'warning'
    case WebSocketState.ERROR:
      return 'danger'
    default:
      return 'info'
  }
})

const tooltipText = computed(() => {
  const baseText = `实时连接状态: ${statusText.value}`
  
  if (state.value === WebSocketState.CONNECTED) {
    return `${baseText}\n点击查看详细信息`
  } else if (state.value === WebSocketState.ERROR) {
    return `${baseText}\n点击重新连接`
  } else if (state.value === WebSocketState.RECONNECTING) {
    return `${baseText}\n正在尝试重新连接...`
  }
  
  return baseText
})

const serverUrl = computed(() => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = import.meta.env.VITE_WS_HOST || window.location.host
  const path = import.meta.env.VITE_WS_PATH || '/ws/'
  return `${protocol}//${host}${path}`
})

// 监听状态变化
watch(state, (newState, oldState) => {
  if (newState === WebSocketState.CONNECTED && oldState !== WebSocketState.CONNECTED) {
    connectionTime.value = formatDateTime(new Date())
    ElMessage.success('实时连接已建立')
  } else if (newState === WebSocketState.RECONNECTING) {
    reconnectCount.value++
  } else if (newState === WebSocketState.ERROR) {
    lastError.value = '连接失败或异常断开'
  }
})

// 处理点击事件
const handleClick = () => {
  if (!props.clickable) return
  
  if (state.value === WebSocketState.ERROR || state.value === WebSocketState.DISCONNECTED) {
    handleReconnect()
  } else {
    detailVisible.value = true
  }
}

// 处理重连
const handleReconnect = async () => {
  try {
    await reconnect()
  } catch (error) {
    ElMessage.error('重连失败，请稍后再试')
  }
}
</script>

<style scoped lang="scss">
.websocket-status {
  display: inline-flex;
  align-items: center;
  
  .status-indicator {
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    padding: 4px 8px;
    border-radius: 4px;
    transition: background-color 0.3s ease;
    
    &:hover {
      background-color: var(--el-bg-color-page);
    }
    
    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      transition: all 0.3s ease;
      
      &.connected {
        background-color: var(--el-color-success);
        box-shadow: 0 0 6px rgba(103, 194, 58, 0.6);
      }
      
      &.connecting {
        background-color: var(--el-color-warning);
        animation: pulse 1.5s infinite;
      }
      
      &.error {
        background-color: var(--el-color-danger);
        animation: shake 0.5s infinite;
      }
      
      &.disconnected {
        background-color: var(--el-color-info);
      }
    }
    
    .status-text {
      font-size: 12px;
      color: var(--el-text-color-regular);
      font-weight: 500;
    }
  }
  
  &.status-connected .status-indicator {
    .status-text {
      color: var(--el-color-success);
    }
  }
  
  &.status-error .status-indicator {
    .status-text {
      color: var(--el-color-danger);
    }
  }
  
  &.status-connecting .status-indicator,
  &.status-reconnecting .status-indicator {
    .status-text {
      color: var(--el-color-warning);
    }
  }
}

.status-detail {
  .detail-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    
    &:last-child {
      margin-bottom: 0;
    }
    
    .label {
      font-size: 14px;
      color: var(--el-text-color-regular);
      font-weight: 500;
    }
    
    .value {
      font-size: 14px;
      color: var(--el-text-color-primary);
      
      &.error {
        color: var(--el-color-danger);
      }
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: center;
  gap: 12px;
}

// 动画
@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.2);
  }
}

@keyframes shake {
  0%, 100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-2px);
  }
  75% {
    transform: translateX(2px);
  }
}
</style>
