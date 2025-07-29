<template>
  <div class="activity-card interactive-element" :class="{ disabled: isDisabled }" @click="handleCardClick">
    <div class="card-image">
      <el-image
        :src="activity.product.image_url || defaultImage"
        :alt="activity.product.name"
        fit="cover"
        class="product-image"
      >
        <template #error>
          <div class="image-error">
            <el-icon><Picture /></el-icon>
            <span>暂无图片</span>
          </div>
        </template>
      </el-image>
      
      <!-- 活动状态标签 -->
      <div class="status-badge">
        <el-tag
          :type="getStatusType(activity.status)"
          size="small"
          effect="dark"
        >
          {{ getStatusText(activity.status) }}
        </el-tag>
      </div>

      <!-- 库存状态 -->
      <div v-if="stockInfo" class="stock-badge">
        <el-tag
          :type="getStockType(stockInfo.status)"
          size="small"
          effect="plain"
        >
          剩余 {{ stockInfo.available_stock }}
        </el-tag>
      </div>
    </div>

    <div class="card-content">
      <!-- 商品名称 -->
      <h3 class="product-name" :title="activity.product.name">
        {{ activity.product.name }}
      </h3>

      <!-- 活动名称 -->
      <p class="activity-name" :title="activity.name">
        {{ activity.name }}
      </p>

      <!-- 价格信息 -->
      <div class="price-section">
        <div class="seckill-price">
          <span class="currency">¥</span>
          <span class="price">{{ formatPrice(activity.seckill_price) }}</span>
        </div>
        <div class="original-price">
          原价 ¥{{ formatPrice(activity.original_price) }}
        </div>
        <div class="discount">
          {{ getDiscountText() }}
        </div>
      </div>

      <!-- 库存显示 -->
      <div class="stock-section">
        <StockDisplay
          :stock-info="stockInfo || null"
          :total-stock="activity.total_stock"
          :activity-id="activity.id"
          :is-real-time="activity.status === 'active'"
          :show-animation="false"
          :enable-web-socket="false"
          @refresh="handleStockRefresh"
        />
      </div>

      <!-- 时间信息 -->
      <div class="time-section">
        <div v-if="activity.status === 'pending'" class="countdown">
          <el-icon><Timer /></el-icon>
          <span>{{ getTimeText() }}</span>
        </div>
        <div v-else-if="activity.status === 'active'" class="countdown active">
          <el-icon><Timer /></el-icon>
          <span>{{ getTimeText() }}</span>
        </div>
        <div v-else class="time-info">
          <span>{{ getTimeText() }}</span>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="action-section">
        <el-button
          v-if="activity.status === 'active'"
          type="primary"
          size="large"
          :disabled="!canParticipate"
          :loading="participating"
          @click="handleParticipate"
          class="participate-btn desktop-hover-lift"
        >
          {{ getButtonText() }}
        </el-button>
        <el-button
          v-else-if="activity.status === 'pending'"
          size="large"
          disabled
          class="participate-btn"
        >
          即将开始
        </el-button>
        <el-button
          v-else
          size="large"
          disabled
          class="participate-btn"
        >
          已结束
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture, Timer } from '@element-plus/icons-vue'
import { formatPrice, formatDateTime, getCountdown } from '@/utils'
import { useActivityStore } from '@/stores/activity'
import { useAuth } from '@/composables/useAuth'
import StockDisplay from './StockDisplay.vue'
import type { SeckillActivity, StockInfo } from '@/types'

interface Props {
  activity: SeckillActivity
  stockInfo?: StockInfo | null
}

interface Emits {
  (e: 'participate', activityId: number): void
  (e: 'click', activity: SeckillActivity): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// Store和认证
const activityStore = useActivityStore()
const { isAuthenticated, requireAuth } = useAuth()

// 状态
const participating = ref(false)
const defaultImage = '/src/assets/default-product.png'

// 计算属性
const isDisabled = computed(() => 
  props.activity.status === 'ended' || props.activity.status === 'cancelled'
)

const canParticipate = computed(() => {
  if (!isAuthenticated.value) return false
  if (props.activity.status !== 'active') return false
  if (!props.stockInfo) return false
  return props.stockInfo.available_stock > 0
})

// 获取状态文本
const getStatusText = (status: string) => {
  return activityStore.getActivityStatusText(status)
}

// 获取状态类型
const getStatusType = (status: string) => {
  return activityStore.getActivityStatusType(status)
}

// 获取库存状态类型
const getStockType = (status: string) => {
  const typeMap = {
    normal: 'success',
    low_stock: 'warning',
    out_of_stock: 'danger',
  }
  return typeMap[status as keyof typeof typeMap] || 'info'
}

// 获取已售数量
const getSoldCount = () => {
  if (!props.stockInfo) return 0
  return props.activity.total_stock - props.stockInfo.available_stock
}

// 获取已售百分比
const getSoldPercentage = () => {
  const sold = getSoldCount()
  return Math.round((sold / props.activity.total_stock) * 100)
}

// 获取进度条颜色
const getProgressColor = () => {
  const percentage = getSoldPercentage()
  if (percentage >= 90) return '#f56c6c'
  if (percentage >= 70) return '#e6a23c'
  return '#67c23a'
}

// 获取折扣文本
const getDiscountText = () => {
  const original = parseFloat(props.activity.original_price)
  const seckill = parseFloat(props.activity.seckill_price)
  const discount = Math.round((1 - seckill / original) * 100)
  return `${discount}折`
}

// 获取时间文本
const getTimeText = () => {
  const now = new Date()
  
  if (props.activity.status === 'pending') {
    const startTime = new Date(props.activity.start_time)
    if (startTime > now) {
      const countdown = getCountdown(props.activity.start_time)
      if (countdown.isExpired) return '即将开始'
      return `${countdown.days}天${countdown.hours}时${countdown.minutes}分后开始`
    }
    return '即将开始'
  }
  
  if (props.activity.status === 'active') {
    const endTime = new Date(props.activity.end_time)
    if (endTime > now) {
      const countdown = getCountdown(props.activity.end_time)
      if (countdown.isExpired) return '已结束'
      return `${countdown.hours}:${countdown.minutes.toString().padStart(2, '0')}:${countdown.seconds.toString().padStart(2, '0')}`
    }
    return '已结束'
  }
  
  return formatDateTime(props.activity.end_time, 'MM-DD HH:mm')
}

// 获取按钮文本
const getButtonText = () => {
  if (!isAuthenticated.value) return '登录后参与'
  if (!props.stockInfo) return '加载中...'
  if (props.stockInfo.available_stock <= 0) return '已抢完'
  return '立即抢购'
}

// 处理参与秒杀
const handleParticipate = async () => {
  if (!requireAuth()) return

  if (!canParticipate.value) {
    ElMessage.warning('当前无法参与此活动')
    return
  }

  participating.value = true
  try {
    emit('participate', props.activity.id)
  } catch (error) {
    // 错误处理由父组件处理
  } finally {
    participating.value = false
  }
}

// 处理库存刷新
const handleStockRefresh = () => {
  // 这里可以触发库存刷新
  console.log('刷新库存:', props.activity.id)
}

// 处理卡片点击
const handleCardClick = () => {
  emit('click', props.activity)
}
</script>

<style scoped lang="scss">
.activity-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;

  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  }

  &.disabled {
    opacity: 0.6;
    cursor: not-allowed;

    &:hover {
      transform: none;
      box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
    }
  }

  .card-image {
    position: relative;
    height: 200px;
    overflow: hidden;

    .product-image {
      width: 100%;
      height: 100%;
    }

    .image-error {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 100%;
      color: var(--el-text-color-placeholder);
      background: var(--el-bg-color-page);

      .el-icon {
        font-size: 32px;
        margin-bottom: 8px;
      }

      span {
        font-size: 12px;
      }
    }

    .status-badge {
      position: absolute;
      top: 12px;
      left: 12px;
    }

    .stock-badge {
      position: absolute;
      top: 12px;
      right: 12px;
    }
  }

  .card-content {
    padding: 16px;

    .product-name {
      margin: 0 0 8px 0;
      font-size: 16px;
      font-weight: 600;
      color: var(--el-text-color-primary);
      line-height: 1.4;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .activity-name {
      margin: 0 0 12px 0;
      font-size: 14px;
      color: var(--el-text-color-regular);
      line-height: 1.4;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .price-section {
      display: flex;
      align-items: baseline;
      gap: 8px;
      margin-bottom: 12px;

      .seckill-price {
        display: flex;
        align-items: baseline;
        color: var(--el-color-danger);
        font-weight: 600;

        .currency {
          font-size: 14px;
        }

        .price {
          font-size: 20px;
        }
      }

      .original-price {
        font-size: 12px;
        color: var(--el-text-color-placeholder);
        text-decoration: line-through;
      }

      .discount {
        font-size: 12px;
        color: var(--el-color-danger);
        background: var(--el-color-danger-light-9);
        padding: 2px 6px;
        border-radius: 4px;
        margin-left: auto;
      }
    }

    .progress-section {
      margin-bottom: 12px;

      .progress-info {
        display: flex;
        justify-content: space-between;
        font-size: 12px;
        color: var(--el-text-color-regular);
        margin-bottom: 6px;
      }
    }

    .time-section {
      margin-bottom: 16px;

      .countdown {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 12px;
        color: var(--el-text-color-regular);

        &.active {
          color: var(--el-color-danger);
          font-weight: 600;
        }

        .el-icon {
          font-size: 14px;
        }
      }

      .time-info {
        font-size: 12px;
        color: var(--el-text-color-placeholder);
      }
    }

    .action-section {
      .participate-btn {
        width: 100%;
        height: 40px;
        font-weight: 600;
      }
    }
  }
}

@media (max-width: 768px) {
  .activity-card {
    // 移动端触摸优化
    &:active {
      transform: scale(0.98);
      transition: transform 0.1s ease;
    }

    .card-image {
      height: 160px;
    }

    .card-content {
      padding: 12px;

      .product-name {
        font-size: 14px;
      }

      .activity-name {
        font-size: 12px;
      }

      .price-section {
        margin-bottom: 8px;

        .seckill-price .price {
          font-size: 18px;
        }

        .original-price {
          font-size: 11px;
        }

        .discount {
          font-size: 10px;
          padding: 1px 4px;
        }
      }

      .time-section {
        margin-bottom: 12px;

        .countdown,
        .time-info {
          font-size: 11px;
        }
      }

      .action-section {
        .participate-btn {
          height: 36px;
          font-size: 14px;
        }
      }
    }
  }
}

// 小屏幕移动端优化
@media (max-width: 480px) {
  .activity-card {
    .card-image {
      height: 140px;
    }

    .card-content {
      padding: 10px;

      .product-name {
        font-size: 13px;
        margin-bottom: 6px;
      }

      .activity-name {
        font-size: 11px;
        margin-bottom: 8px;
      }

      .price-section {
        .seckill-price .price {
          font-size: 16px;
        }
      }
    }
  }
}
</style>
