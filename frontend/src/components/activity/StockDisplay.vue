<template>
  <div class="stock-display" :class="stockClass">
    <!-- 库存数量显示 -->
    <div class="stock-info">
      <div class="stock-number">
        <span class="label">剩余库存</span>
        <div class="number-container">
          <transition name="number-change" mode="out-in">
            <span :key="displayStock" class="number">{{ displayStock }}</span>
          </transition>
          <span class="unit">件</span>
        </div>
      </div>
      
      <!-- 库存状态指示器 -->
      <div class="stock-status">
        <el-tag
          :type="statusType"
          :effect="statusEffect"
          size="small"
        >
          {{ statusText }}
        </el-tag>
      </div>
    </div>

    <!-- 库存进度条 -->
    <div class="stock-progress">
      <div class="progress-info">
        <span class="sold">已售 {{ soldCount }}</span>
        <span class="total">总量 {{ totalStock }}</span>
      </div>
      <el-progress
        :percentage="soldPercentage"
        :stroke-width="8"
        :show-text="false"
        :color="progressColor"
        class="progress-bar"
      />
      <div class="progress-labels">
        <span class="start-label">0</span>
        <span class="end-label">{{ totalStock }}</span>
      </div>
    </div>

    <!-- 实时更新指示器 -->
    <div v-if="isRealTime" class="realtime-indicator">
      <div class="indicator-dot" :class="{
        active: isUpdating || (realTimeStock?.hasRecentChange.value),
        websocket: realTimeStock?.isSubscribed.value
      }"></div>
      <span class="indicator-text">
        {{ realTimeStock?.isSubscribed.value ? 'WebSocket实时' : '定时更新' }}
      </span>
      <span class="last-update">{{ lastUpdateText }}</span>
    </div>

    <!-- 库存变化动画 -->
    <transition name="stock-change">
      <div v-if="showChangeIndicator" class="change-indicator" :class="changeType">
        <el-icon><component :is="changeIcon" /></el-icon>
        <span>{{ changeText }}</span>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ArrowUp, ArrowDown, Refresh } from '@element-plus/icons-vue'
import { formatRelativeTime } from '@/utils'
import { useRealTimeStock } from '@/composables/useRealTimeStock'
import type { StockInfo } from '@/types'

interface Props {
  stockInfo: StockInfo | null
  totalStock: number
  activityId?: number
  isRealTime?: boolean
  updateInterval?: number // 更新间隔（毫秒）
  showAnimation?: boolean
  enableWebSocket?: boolean // 是否启用WebSocket实时更新
}

interface Emits {
  (e: 'refresh'): void
  (e: 'stock-change', change: { old: number; new: number; diff: number }): void
}

const props = withDefaults(defineProps<Props>(), {
  isRealTime: true,
  updateInterval: 5000,
  showAnimation: true,
  enableWebSocket: true,
})

const emit = defineEmits<Emits>()

// 状态
const isUpdating = ref(false)
const showChangeIndicator = ref(false)
const changeType = ref<'increase' | 'decrease' | 'refresh'>('refresh')
const changeText = ref('')
const previousStock = ref<number | null>(null)
const updateTimer = ref<number | null>(null)

// 实时库存更新 (仅在启用WebSocket且有activityId时使用)
const realTimeStock = props.enableWebSocket && props.activityId
  ? useRealTimeStock(props.activityId)
  : null

// 计算属性
const displayStock = computed(() => {
  // 优先使用实时数据
  if (realTimeStock?.stockInfo.value) {
    return realTimeStock.stockInfo.value.available_stock
  }
  return props.stockInfo?.available_stock ?? 0
})

const soldCount = computed(() => {
  return props.totalStock - displayStock.value
})

const soldPercentage = computed(() => {
  if (props.totalStock === 0) return 0
  return Math.round((soldCount.value / props.totalStock) * 100)
})

const stockClass = computed(() => {
  const classes = []
  
  if (props.stockInfo?.status) {
    classes.push(`status-${props.stockInfo.status}`)
  }
  
  if (displayStock.value === 0) {
    classes.push('sold-out')
  } else if (displayStock.value <= props.totalStock * 0.1) {
    classes.push('low-stock')
  }
  
  return classes
})

const statusType = computed(() => {
  if (!props.stockInfo) return 'info'
  
  switch (props.stockInfo.status) {
    case 'normal':
      return 'success'
    case 'low_stock':
      return 'warning'
    case 'out_of_stock':
      return 'danger'
    default:
      return 'info'
  }
})

const statusEffect = computed(() => {
  return displayStock.value <= props.totalStock * 0.2 ? 'dark' : 'plain'
})

const statusText = computed(() => {
  if (!props.stockInfo) return '加载中'
  
  switch (props.stockInfo.status) {
    case 'normal':
      return '库存充足'
    case 'low_stock':
      return '库存紧张'
    case 'out_of_stock':
      return '已售罄'
    default:
      return '未知状态'
  }
})

const progressColor = computed(() => {
  const percentage = soldPercentage.value
  
  if (percentage >= 95) return '#f56c6c'
  if (percentage >= 80) return '#e6a23c'
  if (percentage >= 60) return '#409eff'
  return '#67c23a'
})

const changeIcon = computed(() => {
  switch (changeType.value) {
    case 'increase':
      return ArrowUp
    case 'decrease':
      return ArrowDown
    default:
      return Refresh
  }
})

const lastUpdateText = computed(() => {
  if (!props.stockInfo?.last_updated) return ''
  return formatRelativeTime(props.stockInfo.last_updated)
})

// 监听库存变化
watch(() => props.stockInfo?.available_stock, (newStock, oldStock) => {
  if (oldStock !== undefined && newStock !== undefined && newStock !== oldStock) {
    handleStockChange(oldStock, newStock)
  }
  
  if (newStock !== undefined) {
    previousStock.value = newStock
  }
}, { immediate: true })

// 处理库存变化
const handleStockChange = (oldStock: number, newStock: number) => {
  const diff = newStock - oldStock
  
  if (diff !== 0 && props.showAnimation) {
    // 显示变化指示器
    if (diff > 0) {
      changeType.value = 'increase'
      changeText.value = `+${diff}`
    } else {
      changeType.value = 'decrease'
      changeText.value = `${diff}`
    }
    
    showChangeIndicator.value = true
    
    // 3秒后隐藏指示器
    setTimeout(() => {
      showChangeIndicator.value = false
    }, 3000)
  }
  
  // 发出变化事件
  emit('stock-change', { old: oldStock, new: newStock, diff })
}

// 开始实时更新
const startRealTimeUpdate = () => {
  if (!props.isRealTime) return
  
  updateTimer.value = setInterval(() => {
    isUpdating.value = true
    emit('refresh')
    
    // 更新指示器动画
    setTimeout(() => {
      isUpdating.value = false
    }, 1000)
  }, props.updateInterval)
}

// 停止实时更新
const stopRealTimeUpdate = () => {
  if (updateTimer.value) {
    clearInterval(updateTimer.value)
    updateTimer.value = null
  }
}

// 手动刷新
const refresh = () => {
  isUpdating.value = true
  emit('refresh')
  
  setTimeout(() => {
    isUpdating.value = false
  }, 1000)
}

// 暴露方法
defineExpose({
  refresh,
  startRealTimeUpdate,
  stopRealTimeUpdate,
})

// 组件挂载时开始实时更新
onMounted(() => {
  if (props.isRealTime) {
    startRealTimeUpdate()
  }
})

// 组件卸载时清理定时器
onUnmounted(() => {
  stopRealTimeUpdate()
})
</script>

<style scoped lang="scss">
.stock-display {
  position: relative;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;

  &.low-stock {
    border-left: 4px solid var(--el-color-warning);
  }

  &.sold-out {
    border-left: 4px solid var(--el-color-danger);
    opacity: 0.8;
  }

  .stock-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .stock-number {
      .label {
        display: block;
        font-size: 12px;
        color: var(--el-text-color-regular);
        margin-bottom: 4px;
      }

      .number-container {
        display: flex;
        align-items: baseline;
        gap: 4px;

        .number {
          font-size: 28px;
          font-weight: 700;
          color: var(--el-text-color-primary);
          font-family: 'Courier New', monospace;
        }

        .unit {
          font-size: 14px;
          color: var(--el-text-color-regular);
        }
      }
    }
  }

  .stock-progress {
    margin-bottom: 16px;

    .progress-info {
      display: flex;
      justify-content: space-between;
      font-size: 12px;
      color: var(--el-text-color-regular);
      margin-bottom: 8px;
    }

    .progress-bar {
      margin-bottom: 4px;
    }

    .progress-labels {
      display: flex;
      justify-content: space-between;
      font-size: 10px;
      color: var(--el-text-color-placeholder);
    }
  }

  .realtime-indicator {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--el-text-color-regular);

    .indicator-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--el-color-success);
      transition: all 0.3s ease;

      &.active {
        background: var(--el-color-primary);
        animation: pulse 1s infinite;
      }
    }

    .indicator-text {
      font-weight: 500;
    }

    .last-update {
      margin-left: auto;
      color: var(--el-text-color-placeholder);
    }
  }

  .change-indicator {
    position: absolute;
    top: 10px;
    right: 10px;
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 12px;
    border-radius: 16px;
    font-size: 12px;
    font-weight: 600;
    z-index: 10;

    &.increase {
      background: var(--el-color-success-light-9);
      color: var(--el-color-success);
    }

    &.decrease {
      background: var(--el-color-danger-light-9);
      color: var(--el-color-danger);
    }

    &.refresh {
      background: var(--el-color-info-light-9);
      color: var(--el-color-info);
    }
  }
}

// 动画
.number-change-enter-active,
.number-change-leave-active {
  transition: all 0.3s ease;
}

.number-change-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.number-change-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

.stock-change-enter-active,
.stock-change-leave-active {
  transition: all 0.3s ease;
}

.stock-change-enter-from,
.stock-change-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

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

// 状态特定样式
.status-normal {
  .stock-number .number {
    color: var(--el-color-success);
  }
}

.status-low_stock {
  .stock-number .number {
    color: var(--el-color-warning);
  }
}

.status-out_of_stock {
  .stock-number .number {
    color: var(--el-color-danger);
  }
}

// 响应式设计
@media (max-width: 768px) {
  .stock-display {
    padding: 16px;

    .stock-info {
      .stock-number {
        .number-container .number {
          font-size: 24px;
        }
      }
    }

    .change-indicator {
      top: 8px;
      right: 8px;
      padding: 4px 8px;
      font-size: 11px;
    }
  }
}
</style>
