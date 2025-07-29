<template>
  <div class="seckill-button-wrapper">
    <el-button
      :type="buttonType"
      :size="size"
      :loading="loading"
      :disabled="isDisabled"
      @click="handleClick"
      class="seckill-button"
      :class="buttonClass"
    >
      <template #loading>
        <div class="loading-content">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>{{ loadingText }}</span>
        </div>
      </template>
      
      <div v-if="!loading" class="button-content">
        <el-icon v-if="buttonIcon"><component :is="buttonIcon" /></el-icon>
        <span>{{ buttonText }}</span>
      </div>
    </el-button>

    <!-- 提示信息 -->
    <div v-if="tipText" class="tip-text" :class="tipType">
      <el-icon><InfoFilled /></el-icon>
      <span>{{ tipText }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Loading, 
  ShoppingCart, 
  Lock, 
  Clock, 
  Check, 
  Close,
  InfoFilled 
} from '@element-plus/icons-vue'
import { useAuth } from '@/composables/useAuth'
import type { SeckillActivity, StockInfo } from '@/types'

interface Props {
  activity: SeckillActivity
  stockInfo?: StockInfo | null
  userParticipation?: any
  size?: 'small' | 'default' | 'large'
  block?: boolean
}

interface Emits {
  (e: 'participate', activityId: number): void
  (e: 'result', result: { type: string; data?: any }): void
}

const props = withDefaults(defineProps<Props>(), {
  size: 'default',
  block: false,
})

const emit = defineEmits<Emits>()

// 认证相关
const { isAuthenticated, requireAuth } = useAuth()

// 状态
const loading = ref(false)
const clickCount = ref(0)
const lastClickTime = ref(0)

// 计算属性
const buttonState = computed(() => {
  // 未登录
  if (!isAuthenticated.value) {
    return 'login'
  }

  // 活动状态检查
  if (props.activity.status === 'pending') {
    return 'pending'
  }
  
  if (props.activity.status === 'ended' || props.activity.status === 'cancelled') {
    return 'ended'
  }

  // 库存检查
  if (!props.stockInfo) {
    return 'loading'
  }

  if (props.stockInfo.available_stock <= 0) {
    return 'sold_out'
  }

  // 用户限购检查
  if (props.userParticipation) {
    const { participation_count, max_allowed } = props.userParticipation
    if (participation_count >= max_allowed) {
      return 'limit_reached'
    }
  }

  // 可以参与
  return 'available'
})

const buttonType = computed(() => {
  switch (buttonState.value) {
    case 'available':
      return 'danger'
    case 'login':
      return 'primary'
    case 'loading':
      return 'info'
    default:
      return 'info'
  }
})

const buttonText = computed(() => {
  switch (buttonState.value) {
    case 'login':
      return '登录后参与'
    case 'pending':
      return '即将开始'
    case 'ended':
      return '活动已结束'
    case 'loading':
      return '加载中...'
    case 'sold_out':
      return '已抢完'
    case 'limit_reached':
      return '已达限购'
    case 'available':
      return '立即抢购'
    default:
      return '暂不可用'
  }
})

const buttonIcon = computed(() => {
  switch (buttonState.value) {
    case 'available':
      return ShoppingCart
    case 'login':
      return Lock
    case 'pending':
      return Clock
    case 'ended':
      return Close
    case 'sold_out':
      return Close
    case 'limit_reached':
      return Check
    default:
      return null
  }
})

const buttonClass = computed(() => {
  const classes = []
  
  if (props.block) {
    classes.push('block')
  }
  
  if (buttonState.value === 'available') {
    classes.push('seckill-active')
  }
  
  return classes
})

const isDisabled = computed(() => {
  return buttonState.value !== 'available' && buttonState.value !== 'login'
})

const loadingText = computed(() => {
  if (clickCount.value > 5) {
    return '排队中...'
  }
  return '抢购中...'
})

const tipText = computed(() => {
  switch (buttonState.value) {
    case 'pending':
      return '活动尚未开始，请耐心等待'
    case 'sold_out':
      return '商品已售罄，下次要快一点哦'
    case 'limit_reached':
      return `您已购买 ${props.userParticipation?.participation_count} 件，已达限购数量`
    case 'ended':
      return '活动已结束，敬请关注下次活动'
    default:
      return ''
  }
})

const tipType = computed(() => {
  switch (buttonState.value) {
    case 'pending':
      return 'warning'
    case 'sold_out':
    case 'ended':
      return 'error'
    case 'limit_reached':
      return 'success'
    default:
      return 'info'
  }
})

// 处理点击
const handleClick = async () => {
  // 防重复点击
  const now = Date.now()
  if (now - lastClickTime.value < 1000) {
    ElMessage.warning('请不要重复点击')
    return
  }
  lastClickTime.value = now
  clickCount.value++

  // 未登录处理
  if (buttonState.value === 'login') {
    requireAuth()
    return
  }

  // 只有可用状态才能参与
  if (buttonState.value !== 'available') {
    return
  }

  loading.value = true
  try {
    emit('participate', props.activity.id)
  } catch (error) {
    // 错误处理由父组件处理
  } finally {
    loading.value = false
    // 重置点击计数
    setTimeout(() => {
      clickCount.value = 0
    }, 5000)
  }
}
</script>

<style scoped lang="scss">
.seckill-button-wrapper {
  .seckill-button {
    position: relative;
    font-weight: 600;
    transition: all 0.3s ease;

    &.block {
      width: 100%;
    }

    &.seckill-active {
      background: linear-gradient(135deg, #ff4757, #ff6b7a);
      border-color: #ff4757;
      box-shadow: 0 4px 12px rgba(255, 71, 87, 0.3);

      &:hover {
        background: linear-gradient(135deg, #ff3742, #ff5a6d);
        box-shadow: 0 6px 16px rgba(255, 71, 87, 0.4);
        transform: translateY(-2px);
      }

      &:active {
        transform: translateY(0);
      }
    }

    .button-content {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 6px;
    }

    .loading-content {
      display: flex;
      align-items: center;
      gap: 6px;

      .is-loading {
        animation: rotating 2s linear infinite;
      }
    }
  }

  .tip-text {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-top: 8px;
    font-size: 12px;
    line-height: 1.4;

    &.warning {
      color: var(--el-color-warning);
    }

    &.error {
      color: var(--el-color-danger);
    }

    &.success {
      color: var(--el-color-success);
    }

    &.info {
      color: var(--el-text-color-regular);
    }

    .el-icon {
      font-size: 14px;
    }
  }
}

@keyframes rotating {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

// 大尺寸按钮特殊样式
.seckill-button.el-button--large.seckill-active {
  height: 50px;
  font-size: 18px;
  
  &:hover {
    animation: pulse 0.6s ease-in-out;
  }
}

@keyframes pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
  100% {
    transform: scale(1);
  }
}

// 移动端优化
@media (max-width: 768px) {
  .seckill-button-wrapper {
    .seckill-button {
      &.seckill-active:hover {
        transform: none;
      }
    }
  }
}
</style>
