<template>
  <div class="order-status-polling">
    <!-- 轮询状态指示器 -->
    <div v-if="isPolling" class="polling-indicator">
      <el-card class="status-card" shadow="hover">
        <div class="status-header">
          <el-icon class="spinning">
            <Loading />
          </el-icon>
          <span class="status-title">正在查询订单状态...</span>
        </div>
        
        <!-- 进度条 -->
        <div class="progress-section">
          <el-progress
            :percentage="progress"
            :stroke-width="6"
            :show-text="false"
            status="success"
            class="progress-bar"
          />
          <div class="progress-info">
            <span class="attempts">第 {{ pollingState.attempts }} 次查询</span>
            <span class="max-attempts">/ {{ maxAttempts }}</span>
          </div>
        </div>
        
        <!-- 当前状态 -->
        <div v-if="orderStatus" class="current-status">
          <el-tag :type="getStatusTagType(orderStatus)" size="large">
            {{ getStatusText(orderStatus) }}
          </el-tag>
        </div>
        
        <!-- 剩余时间 -->
        <div v-if="remainingTime > 0" class="remaining-time">
          <el-icon><Clock /></el-icon>
          <span>支付剩余时间: {{ formatTime(remainingTime) }}</span>
        </div>
        
        <!-- 取消按钮 -->
        <div class="actions">
          <el-button
            type="info"
            size="small"
            plain
            @click="handleStopPolling"
          >
            停止查询
          </el-button>
        </div>
      </el-card>
    </div>
    
    <!-- 订单信息显示 -->
    <div v-if="currentOrder && !isPolling" class="order-result">
      <el-card class="result-card" shadow="hover">
        <div class="result-header">
          <el-icon :class="getStatusIconClass(orderStatus)">
            <component :is="getStatusIcon(orderStatus)" />
          </el-icon>
          <span class="result-title">{{ getResultTitle(orderStatus) }}</span>
        </div>
        
        <div class="order-info">
          <div class="info-row">
            <span class="label">订单号:</span>
            <span class="value">{{ currentOrder.id }}</span>
          </div>
          <div class="info-row">
            <span class="label">商品:</span>
            <span class="value">{{ currentOrder.product_name }}</span>
          </div>
          <div class="info-row">
            <span class="label">数量:</span>
            <span class="value">{{ currentOrder.quantity }}</span>
          </div>
          <div class="info-row">
            <span class="label">金额:</span>
            <span class="value price">¥{{ currentOrder.total_amount }}</span>
          </div>
          <div class="info-row">
            <span class="label">状态:</span>
            <el-tag :type="getStatusTagType(orderStatus)">
              {{ getStatusText(orderStatus) }}
            </el-tag>
          </div>
          <div v-if="currentOrder.payment_deadline && orderStatus === 'pending_payment'" class="info-row">
            <span class="label">支付截止:</span>
            <span class="value">{{ formatDateTime(currentOrder.payment_deadline) }}</span>
          </div>
        </div>
        
        <!-- 操作按钮 -->
        <div class="order-actions">
          <el-button
            v-if="currentOrder.can_pay"
            type="primary"
            @click="handlePayment"
          >
            立即支付
          </el-button>
          <el-button
            v-if="currentOrder.can_cancel"
            type="danger"
            plain
            @click="handleCancelOrder"
          >
            取消订单
          </el-button>
          <el-button
            type="info"
            plain
            @click="handleViewDetail"
          >
            查看详情
          </el-button>
        </div>
      </el-card>
    </div>
    
    <!-- 错误信息 -->
    <div v-if="pollingState.error" class="error-info">
      <el-alert
        :title="pollingState.error"
        type="error"
        :closable="false"
        show-icon
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  ElCard,
  ElIcon,
  ElProgress,
  ElTag,
  ElButton,
  ElAlert
} from 'element-plus'
import {
  Loading,
  Clock,
  SuccessFilled,
  CircleCloseFilled,
  WarningFilled,
  InfoFilled
} from '@element-plus/icons-vue'
import { useOrderPolling, OrderStatus } from '@/composables/useOrderPolling'

// 组件属性
interface Props {
  orderId?: string
  maxAttempts?: number
  interval?: number
  autoStart?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxAttempts: 30,
  interval: 2000,
  autoStart: false
})

// 组件事件
interface Emits {
  success: [order: any]
  failed: [error: any]
  statusChange: [status: string, order: any]
  payment: [order: any]
  cancel: [order: any]
  viewDetail: [order: any]
}

const emit = defineEmits<Emits>()

// 使用订单轮询
const {
  pollingState,
  currentOrder,
  orderStatus,
  remainingTime,
  isPolling,
  progress,
  startPolling,
  stopPolling,
  reset
} = useOrderPolling({
  interval: props.interval,
  maxAttempts: props.maxAttempts,
  onSuccess: (order) => emit('success', order),
  onError: (error) => emit('failed', error),
  onStatusChange: (status, order) => emit('statusChange', status, order)
})

// 计算属性
const progressPercentage = computed(() => {
  if (props.maxAttempts === 0) return 0
  return Math.min((pollingState.value.attempts / props.maxAttempts) * 100, 100)
})

// 方法
const getStatusText = (status: string): string => {
  const statusTexts: Record<string, string> = {
    [OrderStatus.PENDING_PAYMENT]: '待支付',
    [OrderStatus.PAID]: '已支付',
    [OrderStatus.CANCELLED]: '已取消',
    [OrderStatus.EXPIRED]: '已过期',
    [OrderStatus.PROCESSING]: '处理中',
    [OrderStatus.COMPLETED]: '已完成'
  }
  return statusTexts[status] || '未知状态'
}

const getStatusTagType = (status: string): string => {
  const tagTypes: Record<string, string> = {
    [OrderStatus.PENDING_PAYMENT]: 'warning',
    [OrderStatus.PAID]: 'success',
    [OrderStatus.CANCELLED]: 'info',
    [OrderStatus.EXPIRED]: 'danger',
    [OrderStatus.PROCESSING]: 'primary',
    [OrderStatus.COMPLETED]: 'success'
  }
  return tagTypes[status] || 'info'
}

const getStatusIcon = (status: string) => {
  const icons: Record<string, any> = {
    [OrderStatus.PENDING_PAYMENT]: WarningFilled,
    [OrderStatus.PAID]: SuccessFilled,
    [OrderStatus.CANCELLED]: CircleCloseFilled,
    [OrderStatus.EXPIRED]: CircleCloseFilled,
    [OrderStatus.PROCESSING]: Loading,
    [OrderStatus.COMPLETED]: SuccessFilled
  }
  return icons[status] || InfoFilled
}

const getStatusIconClass = (status: string): string => {
  const classes: Record<string, string> = {
    [OrderStatus.PENDING_PAYMENT]: 'status-icon warning',
    [OrderStatus.PAID]: 'status-icon success',
    [OrderStatus.CANCELLED]: 'status-icon error',
    [OrderStatus.EXPIRED]: 'status-icon error',
    [OrderStatus.PROCESSING]: 'status-icon processing',
    [OrderStatus.COMPLETED]: 'status-icon success'
  }
  return classes[status] || 'status-icon'
}

const getResultTitle = (status: string): string => {
  const titles: Record<string, string> = {
    [OrderStatus.PENDING_PAYMENT]: '订单创建成功，请及时支付',
    [OrderStatus.PAID]: '支付成功！',
    [OrderStatus.CANCELLED]: '订单已取消',
    [OrderStatus.EXPIRED]: '订单已过期',
    [OrderStatus.PROCESSING]: '订单处理中',
    [OrderStatus.COMPLETED]: '订单已完成'
  }
  return titles[status] || '订单状态更新'
}

const formatTime = (seconds: number): string => {
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`
}

const formatDateTime = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN')
}

// 事件处理
const handleStopPolling = () => {
  stopPolling('用户手动停止')
}

const handlePayment = () => {
  if (currentOrder.value) {
    emit('payment', currentOrder.value)
  }
}

const handleCancelOrder = () => {
  if (currentOrder.value) {
    emit('cancel', currentOrder.value)
  }
}

const handleViewDetail = () => {
  if (currentOrder.value) {
    emit('viewDetail', currentOrder.value)
  }
}

// 暴露方法给父组件
defineExpose({
  startPolling,
  stopPolling,
  reset,
  isPolling,
  currentOrder,
  orderStatus
})
</script>

<style scoped lang="scss">
.order-status-polling {
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
}

.polling-indicator {
  .status-card {
    text-align: center;
    
    .status-header {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      margin-bottom: 16px;
      
      .spinning {
        animation: spin 1s linear infinite;
        color: var(--el-color-primary);
      }
      
      .status-title {
        font-weight: 600;
        color: var(--el-text-color-primary);
      }
    }
    
    .progress-section {
      margin-bottom: 16px;
      
      .progress-bar {
        margin-bottom: 8px;
      }
      
      .progress-info {
        font-size: 12px;
        color: var(--el-text-color-secondary);
        
        .attempts {
          font-weight: 600;
        }
      }
    }
    
    .current-status {
      margin-bottom: 16px;
    }
    
    .remaining-time {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 4px;
      margin-bottom: 16px;
      font-size: 14px;
      color: var(--el-color-warning);
    }
    
    .actions {
      margin-top: 16px;
    }
  }
}

.order-result {
  .result-card {
    .result-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 20px;
      padding-bottom: 12px;
      border-bottom: 1px solid var(--el-border-color-light);
      
      .status-icon {
        font-size: 20px;
        
        &.success {
          color: var(--el-color-success);
        }
        
        &.warning {
          color: var(--el-color-warning);
        }
        
        &.error {
          color: var(--el-color-error);
        }
        
        &.processing {
          color: var(--el-color-primary);
          animation: spin 1s linear infinite;
        }
      }
      
      .result-title {
        font-size: 16px;
        font-weight: 600;
        color: var(--el-text-color-primary);
      }
    }
    
    .order-info {
      margin-bottom: 20px;
      
      .info-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 8px 0;
        border-bottom: 1px solid var(--el-border-color-lighter);
        
        &:last-child {
          border-bottom: none;
        }
        
        .label {
          font-weight: 500;
          color: var(--el-text-color-regular);
        }
        
        .value {
          color: var(--el-text-color-primary);
          
          &.price {
            font-weight: 600;
            color: var(--el-color-danger);
          }
        }
      }
    }
    
    .order-actions {
      display: flex;
      gap: 8px;
      justify-content: center;
    }
  }
}

.error-info {
  margin-top: 16px;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

// 响应式设计
@media (max-width: 768px) {
  .order-status-polling {
    max-width: 100%;
    padding: 0 16px;
  }
  
  .order-actions {
    flex-direction: column;
    
    .el-button {
      width: 100%;
    }
  }
}
</style>
