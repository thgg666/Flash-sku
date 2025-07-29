<template>
  <div class="seckill-result">
    <!-- 成功结果 -->
    <el-dialog
      v-model="successVisible"
      title="抢购成功"
      width="400px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      center
    >
      <div class="result-content success">
        <div class="result-icon">
          <el-icon class="success-icon"><SuccessFilled /></el-icon>
        </div>
        <h3>恭喜您抢购成功！</h3>
        <div class="result-info">
          <p>订单号：{{ result?.order_id }}</p>
          <p>商品：{{ activity?.product.name }}</p>
          <p>数量：{{ result?.quantity || 1 }} 件</p>
          <p>金额：¥{{ formatPrice(result?.amount || activity?.seckill_price || '0') }}</p>
        </div>
        <div class="countdown-tip">
          <el-icon><Clock /></el-icon>
          <span>请在 <strong>30分钟</strong> 内完成支付</span>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="handleContinueShopping">继续购物</el-button>
          <el-button type="primary" @click="handleGoToPay">立即支付</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 失败结果 -->
    <el-dialog
      v-model="failureVisible"
      title="抢购失败"
      width="400px"
      center
    >
      <div class="result-content failure">
        <div class="result-icon">
          <el-icon class="failure-icon"><CircleCloseFilled /></el-icon>
        </div>
        <h3>很遗憾，抢购失败</h3>
        <div class="failure-reason">
          <p>{{ getFailureMessage() }}</p>
        </div>
        <div class="suggestion">
          <p>{{ getFailureSuggestion() }}</p>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="handleCloseFailure">关闭</el-button>
          <el-button type="primary" @click="handleRetry" :disabled="!canRetry">
            {{ retryButtonText }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 限流结果 -->
    <el-dialog
      v-model="rateLimitVisible"
      title="请求过于频繁"
      width="400px"
      center
    >
      <div class="result-content rate-limit">
        <div class="result-icon">
          <el-icon class="warning-icon"><WarningFilled /></el-icon>
        </div>
        <h3>请求过于频繁</h3>
        <div class="rate-limit-info">
          <p>为了保证系统稳定，请稍后再试</p>
          <div class="countdown">
            <span>请等待 </span>
            <span class="countdown-number">{{ rateLimitCountdown }}</span>
            <span> 秒后重试</span>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="handleCloseRateLimit">关闭</el-button>
          <el-button 
            type="primary" 
            @click="handleRetry" 
            :disabled="rateLimitCountdown > 0"
          >
            重新抢购
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 排队中状态 -->
    <el-dialog
      v-model="queueVisible"
      title="排队中"
      width="400px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      center
    >
      <div class="result-content queue">
        <div class="result-icon">
          <el-icon class="queue-icon rotating"><Loading /></el-icon>
        </div>
        <h3>正在排队抢购</h3>
        <div class="queue-info">
          <p>当前排队人数较多，请耐心等待</p>
          <div class="queue-position" v-if="queuePosition">
            <span>您前面还有 </span>
            <span class="position-number">{{ queuePosition }}</span>
            <span> 人</span>
          </div>
          <div class="estimated-time" v-if="estimatedWaitTime">
            <span>预计等待时间：{{ estimatedWaitTime }}秒</span>
          </div>
        </div>
        <div class="queue-progress">
          <el-progress 
            :percentage="queueProgress" 
            :show-text="false"
            :stroke-width="8"
            color="#409eff"
          />
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="handleCancelQueue">取消排队</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  SuccessFilled, 
  CircleCloseFilled, 
  WarningFilled, 
  Loading, 
  Clock 
} from '@element-plus/icons-vue'
import { formatPrice } from '@/utils'
import type { SeckillActivity, SeckillResult } from '@/types'

interface Props {
  activity?: SeckillActivity | null
  result?: SeckillResult | null
  visible?: boolean
  type?: 'success' | 'failure' | 'rate_limit' | 'queue'
}

interface Emits {
  (e: 'close'): void
  (e: 'retry'): void
  (e: 'pay', orderId: string): void
  (e: 'continue-shopping'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 路由
const router = useRouter()

// 状态
const successVisible = ref(false)
const failureVisible = ref(false)
const rateLimitVisible = ref(false)
const queueVisible = ref(false)
const rateLimitCountdown = ref(0)
const queuePosition = ref(0)
const estimatedWaitTime = ref(0)
const queueProgress = ref(0)

// 定时器
const rateLimitTimer = ref<number | null>(null)
const queueTimer = ref<number | null>(null)

// 计算属性
const canRetry = computed(() => {
  if (props.type === 'rate_limit') {
    return rateLimitCountdown.value <= 0
  }
  return props.type === 'failure'
})

const retryButtonText = computed(() => {
  if (props.type === 'rate_limit' && rateLimitCountdown.value > 0) {
    return `${rateLimitCountdown.value}s后重试`
  }
  return '重新抢购'
})

// 监听props变化
watch(() => props.visible, (visible) => {
  if (visible && props.type) {
    showResult(props.type)
  } else {
    hideAllDialogs()
  }
})

watch(() => props.type, (type) => {
  if (props.visible && type) {
    showResult(type)
  }
})

// 显示结果
const showResult = (type: string) => {
  hideAllDialogs()
  
  switch (type) {
    case 'success':
      successVisible.value = true
      break
    case 'failure':
      failureVisible.value = true
      break
    case 'rate_limit':
      rateLimitVisible.value = true
      startRateLimitCountdown()
      break
    case 'queue':
      queueVisible.value = true
      startQueueSimulation()
      break
  }
}

// 隐藏所有对话框
const hideAllDialogs = () => {
  successVisible.value = false
  failureVisible.value = false
  rateLimitVisible.value = false
  queueVisible.value = false
  
  // 清理定时器
  if (rateLimitTimer.value) {
    clearInterval(rateLimitTimer.value)
    rateLimitTimer.value = null
  }
  if (queueTimer.value) {
    clearInterval(queueTimer.value)
    queueTimer.value = null
  }
}

// 获取失败消息
const getFailureMessage = () => {
  const code = props.result?.code
  const message = props.result?.message
  
  if (message) return message
  
  switch (code) {
    case 'SOLD_OUT':
      return '商品已售罄'
    case 'LIMIT_EXCEEDED':
      return '超出限购数量'
    case 'ACTIVITY_ENDED':
      return '活动已结束'
    case 'INSUFFICIENT_STOCK':
      return '库存不足'
    default:
      return '抢购失败，请稍后重试'
  }
}

// 获取失败建议
const getFailureSuggestion = () => {
  const code = props.result?.code
  
  switch (code) {
    case 'SOLD_OUT':
      return '关注我们的其他活动，不要错过下次机会'
    case 'LIMIT_EXCEEDED':
      return '每人限购数量有限，感谢您的理解'
    case 'ACTIVITY_ENDED':
      return '活动已结束，敬请关注下次活动'
    case 'INSUFFICIENT_STOCK':
      return '手速要快一点哦，下次加油'
    default:
      return '可以尝试重新抢购，或者关注其他活动'
  }
}

// 开始限流倒计时
const startRateLimitCountdown = () => {
  rateLimitCountdown.value = 30 // 30秒倒计时
  
  rateLimitTimer.value = setInterval(() => {
    rateLimitCountdown.value--
    if (rateLimitCountdown.value <= 0) {
      clearInterval(rateLimitTimer.value!)
      rateLimitTimer.value = null
    }
  }, 1000)
}

// 开始排队模拟
const startQueueSimulation = () => {
  queuePosition.value = Math.floor(Math.random() * 100) + 50
  estimatedWaitTime.value = queuePosition.value * 2
  queueProgress.value = 0
  
  queueTimer.value = setInterval(() => {
    if (queuePosition.value > 0) {
      queuePosition.value--
      estimatedWaitTime.value = queuePosition.value * 2
      queueProgress.value = Math.min(100, (100 - queuePosition.value) * 2)
    } else {
      clearInterval(queueTimer.value!)
      queueTimer.value = null
      // 模拟排队完成，显示结果
      queueVisible.value = false
      // 这里可以触发实际的抢购结果
    }
  }, 1000)
}

// 事件处理
const handleGoToPay = () => {
  if (props.result?.order_id) {
    emit('pay', props.result.order_id)
    router.push(`/user/orders/${props.result.order_id}/pay`)
  }
  hideAllDialogs()
  emit('close')
}

const handleContinueShopping = () => {
  emit('continue-shopping')
  hideAllDialogs()
  emit('close')
}

const handleCloseFailure = () => {
  hideAllDialogs()
  emit('close')
}

const handleCloseRateLimit = () => {
  hideAllDialogs()
  emit('close')
}

const handleRetry = () => {
  if (canRetry.value) {
    emit('retry')
    hideAllDialogs()
    emit('close')
  }
}

const handleCancelQueue = () => {
  hideAllDialogs()
  emit('close')
  ElMessage.info('已取消排队')
}

// 组件卸载时清理定时器
onUnmounted(() => {
  hideAllDialogs()
})
</script>

<style scoped lang="scss">
.seckill-result {
  .result-content {
    text-align: center;
    padding: 20px 0;

    .result-icon {
      margin-bottom: 16px;

      .success-icon {
        font-size: 64px;
        color: var(--el-color-success);
      }

      .failure-icon {
        font-size: 64px;
        color: var(--el-color-danger);
      }

      .warning-icon {
        font-size: 64px;
        color: var(--el-color-warning);
      }

      .queue-icon {
        font-size: 64px;
        color: var(--el-color-primary);

        &.rotating {
          animation: rotating 2s linear infinite;
        }
      }
    }

    h3 {
      margin: 0 0 16px 0;
      font-size: 20px;
      font-weight: 600;
    }

    .result-info {
      background: var(--el-bg-color-page);
      border-radius: 8px;
      padding: 16px;
      margin: 16px 0;
      text-align: left;

      p {
        margin: 8px 0;
        display: flex;
        justify-content: space-between;
        font-size: 14px;
      }
    }

    .countdown-tip {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      color: var(--el-color-warning);
      font-size: 14px;
      margin-top: 16px;

      strong {
        color: var(--el-color-danger);
      }
    }

    .failure-reason {
      margin: 16px 0;
      
      p {
        font-size: 16px;
        color: var(--el-color-danger);
        margin: 0;
      }
    }

    .suggestion {
      margin: 16px 0;
      
      p {
        font-size: 14px;
        color: var(--el-text-color-regular);
        margin: 0;
      }
    }

    .rate-limit-info {
      margin: 16px 0;

      .countdown {
        margin-top: 12px;
        font-size: 16px;

        .countdown-number {
          color: var(--el-color-primary);
          font-weight: 600;
          font-size: 20px;
        }
      }
    }

    .queue-info {
      margin: 16px 0;

      .queue-position {
        margin: 12px 0;
        font-size: 16px;

        .position-number {
          color: var(--el-color-primary);
          font-weight: 600;
          font-size: 20px;
        }
      }

      .estimated-time {
        font-size: 14px;
        color: var(--el-text-color-regular);
      }
    }

    .queue-progress {
      margin: 20px 0;
    }
  }

  .dialog-footer {
    display: flex;
    justify-content: center;
    gap: 12px;
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
</style>
