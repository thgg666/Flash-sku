<template>
  <div class="countdown-timer" :class="{ expired: isExpired, urgent: isUrgent }">
    <div v-if="!isExpired" class="countdown-display">
      <div v-if="days > 0" class="time-unit">
        <span class="number">{{ days }}</span>
        <span class="label">天</span>
      </div>
      <div class="time-unit">
        <span class="number">{{ hours.toString().padStart(2, '0') }}</span>
        <span class="label">时</span>
      </div>
      <div class="separator">:</div>
      <div class="time-unit">
        <span class="number">{{ minutes.toString().padStart(2, '0') }}</span>
        <span class="label">分</span>
      </div>
      <div class="separator">:</div>
      <div class="time-unit">
        <span class="number">{{ seconds.toString().padStart(2, '0') }}</span>
        <span class="label">秒</span>
      </div>
    </div>
    <div v-else class="expired-text">
      {{ expiredText }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getCountdown } from '@/utils'

interface Props {
  endTime: string | Date
  expiredText?: string
  urgentThreshold?: number // 紧急状态阈值（分钟）
}

interface Emits {
  (e: 'expired'): void
  (e: 'urgent'): void
  (e: 'tick', countdown: { days: number; hours: number; minutes: number; seconds: number }): void
}

const props = withDefaults(defineProps<Props>(), {
  expiredText: '已结束',
  urgentThreshold: 10,
})

const emit = defineEmits<Emits>()

// 状态
const days = ref(0)
const hours = ref(0)
const minutes = ref(0)
const seconds = ref(0)
const isExpired = ref(false)
const timer = ref<number | null>(null)

// 计算属性
const isUrgent = computed(() => {
  if (isExpired.value) return false
  const totalMinutes = days.value * 24 * 60 + hours.value * 60 + minutes.value
  return totalMinutes <= props.urgentThreshold
})

// 更新倒计时
const updateCountdown = () => {
  const countdown = getCountdown(props.endTime)
  
  if (countdown.isExpired) {
    if (!isExpired.value) {
      isExpired.value = true
      emit('expired')
    }
    return
  }

  const wasUrgent = isUrgent.value
  
  days.value = countdown.days
  hours.value = countdown.hours
  minutes.value = countdown.minutes
  seconds.value = countdown.seconds
  
  // 发出tick事件
  emit('tick', { days: days.value, hours: hours.value, minutes: minutes.value, seconds: seconds.value })
  
  // 检查是否进入紧急状态
  if (!wasUrgent && isUrgent.value) {
    emit('urgent')
  }
}

// 开始倒计时
const startCountdown = () => {
  updateCountdown()
  
  if (!isExpired.value) {
    timer.value = setInterval(updateCountdown, 1000)
  }
}

// 停止倒计时
const stopCountdown = () => {
  if (timer.value) {
    clearInterval(timer.value)
    timer.value = null
  }
}

// 重置倒计时
const resetCountdown = () => {
  stopCountdown()
  isExpired.value = false
  startCountdown()
}

// 暴露方法给父组件
defineExpose({
  resetCountdown,
  stopCountdown,
})

// 组件挂载时开始倒计时
onMounted(() => {
  startCountdown()
})

// 组件卸载时清理定时器
onUnmounted(() => {
  stopCountdown()
})
</script>

<style scoped lang="scss">
.countdown-timer {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-family: 'Courier New', monospace;
  
  &.urgent {
    animation: pulse 1s infinite;
  }
  
  &.expired {
    .expired-text {
      color: var(--el-text-color-placeholder);
      font-size: 16px;
    }
  }

  .countdown-display {
    display: flex;
    align-items: center;
    gap: 4px;

    .time-unit {
      display: flex;
      flex-direction: column;
      align-items: center;
      min-width: 40px;
      padding: 8px 6px;
      background: var(--el-text-color-primary);
      color: white;
      border-radius: 6px;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

      .number {
        font-size: 18px;
        font-weight: 700;
        line-height: 1;
      }

      .label {
        font-size: 10px;
        margin-top: 2px;
        opacity: 0.8;
      }
    }

    .separator {
      font-size: 18px;
      font-weight: 700;
      color: var(--el-text-color-primary);
      margin: 0 2px;
    }
  }

  &.urgent {
    .countdown-display .time-unit {
      background: var(--el-color-danger);
      animation: shake 0.5s infinite;
    }
    
    .separator {
      color: var(--el-color-danger);
    }
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
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

// 大屏幕样式
@media (min-width: 768px) {
  .countdown-timer {
    .countdown-display {
      gap: 8px;

      .time-unit {
        min-width: 50px;
        padding: 12px 8px;

        .number {
          font-size: 24px;
        }

        .label {
          font-size: 12px;
        }
      }

      .separator {
        font-size: 24px;
        margin: 0 4px;
      }
    }
  }
}

// 小屏幕样式
@media (max-width: 480px) {
  .countdown-timer {
    .countdown-display {
      gap: 2px;

      .time-unit {
        min-width: 32px;
        padding: 6px 4px;

        .number {
          font-size: 14px;
        }

        .label {
          font-size: 8px;
        }
      }

      .separator {
        font-size: 14px;
        margin: 0 1px;
      }
    }
  }
}
</style>
