<template>
  <div class="seckill-button-container">
    <!-- 主要秒杀按钮 -->
    <el-button
      :type="buttonType"
      :loading="isProcessing"
      :disabled="!canParticipate || countdown > 0"
      :size="size"
      :class="[
        'seckill-button',
        {
          'seckill-button--success': status === SeckillStatus.SUCCESS,
          'seckill-button--failed': status === SeckillStatus.FAILED,
          'seckill-button--processing': isProcessing
        }
      ]"
      @click="handleClick"
    >
      <!-- 按钮图标 -->
      <el-icon v-if="showIcon && !isProcessing" class="button-icon">
        <component :is="buttonIcon" />
      </el-icon>
      
      <!-- 按钮文本 -->
      <span class="button-text">
        {{ displayText }}
      </span>
      
      <!-- 倒计时显示 -->
      <span v-if="countdown > 0" class="countdown">
        ({{ countdown }}s)
      </span>
    </el-button>
    
    <!-- 取消按钮 -->
    <el-button
      v-if="showCancelButton && isProcessing"
      type="info"
      :size="size"
      plain
      class="cancel-button"
      @click="handleCancel"
    >
      取消
    </el-button>
    
    <!-- 进度条 -->
    <div v-if="showProgress && isProcessing" class="progress-container">
      <el-progress
        :percentage="progressPercentage"
        :stroke-width="4"
        :show-text="false"
        status="success"
        class="progress-bar"
      />
      <div class="progress-text">
        {{ progressText }}
      </div>
    </div>
    
    <!-- 结果提示 -->
    <transition name="fade">
      <div
        v-if="result && showResult"
        :class="[
          'result-tip',
          {
            'result-tip--success': result.success,
            'result-tip--error': !result.success
          }
        ]"
      >
        <el-icon class="result-icon">
          <component :is="result.success ? 'SuccessFilled' : 'CircleCloseFilled'" />
        </el-icon>
        <span class="result-text">{{ result.message }}</span>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElButton, ElIcon, ElProgress } from 'element-plus'
import { 
  Lightning, 
  SuccessFilled, 
  CircleCloseFilled, 
  Loading,
  Warning 
} from '@element-plus/icons-vue'
import { useSeckill, SeckillStatus } from '@/composables/useSeckill'
import type { SeckillActivity } from '@/types'

// 组件属性
interface Props {
  activity: SeckillActivity
  size?: 'large' | 'default' | 'small'
  showIcon?: boolean
  showCancelButton?: boolean
  showProgress?: boolean
  showResult?: boolean
  customText?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'default',
  showIcon: true,
  showCancelButton: true,
  showProgress: true,
  showResult: true,
  disabled: false
})

// 组件事件
interface Emits {
  success: [result: any]
  failed: [error: any]
  click: []
  cancel: []
}

const emit = defineEmits<Emits>()

// 使用秒杀组合式函数
const {
  status,
  loading,
  result,
  error,
  countdown,
  canParticipate,
  isProcessing,
  buttonText,
  buttonType,
  participate,
  cancel,
  reset
} = useSeckill(props.activity)

// 本地状态
const progressPercentage = ref(0)
const progressText = ref('')

// 计算属性
const displayText = computed(() => {
  if (props.customText) {
    return props.customText
  }
  return buttonText.value
})

const buttonIcon = computed(() => {
  switch (status.value) {
    case SeckillStatus.SUCCESS:
      return SuccessFilled
    case SeckillStatus.FAILED:
      return Warning
    case SeckillStatus.REQUESTING:
    case SeckillStatus.PREPARING:
      return Loading
    default:
      return Lightning
  }
})

const actualCanParticipate = computed(() => {
  return canParticipate.value && !props.disabled
})

// 监听处理状态变化
watch(isProcessing, (newVal) => {
  if (newVal) {
    startProgress()
  } else {
    stopProgress()
  }
})

watch(result, (newVal) => {
  if (newVal) {
    if (newVal.success) {
      emit('success', newVal)
    } else {
      emit('failed', newVal)
    }
  }
})

// 进度模拟
const startProgress = () => {
  progressPercentage.value = 0
  progressText.value = '正在提交请求...'
  
  const interval = setInterval(() => {
    if (progressPercentage.value < 90) {
      progressPercentage.value += Math.random() * 10
      
      if (progressPercentage.value < 30) {
        progressText.value = '正在提交请求...'
      } else if (progressPercentage.value < 60) {
        progressText.value = '服务器处理中...'
      } else {
        progressText.value = '即将完成...'
      }
    } else {
      clearInterval(interval)
      if (isProcessing.value) {
        progressText.value = '等待响应...'
      }
    }
  }, 100)
}

const stopProgress = () => {
  progressPercentage.value = 100
  progressText.value = status.value === SeckillStatus.SUCCESS ? '处理完成' : '处理结束'
}

// 事件处理
const handleClick = async () => {
  if (!actualCanParticipate.value) {
    return
  }
  
  emit('click')
  
  try {
    await participate(props.activity.id)
  } catch (err) {
    console.error('秒杀失败:', err)
  }
}

const handleCancel = () => {
  emit('cancel')
  cancel()
}

// 暴露方法给父组件
defineExpose({
  participate,
  cancel,
  reset,
  status,
  result
})
</script>

<style scoped lang="scss">
.seckill-button-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.seckill-button {
  position: relative;
  min-width: 120px;
  font-weight: 600;
  transition: all 0.3s ease;
  
  .button-icon {
    margin-right: 4px;
  }
  
  .button-text {
    flex: 1;
  }
  
  .countdown {
    margin-left: 4px;
    font-size: 0.9em;
    opacity: 0.8;
  }
  
  &--success {
    animation: success-pulse 0.6s ease-out;
  }
  
  &--failed {
    animation: shake 0.5s ease-out;
  }
  
  &--processing {
    .button-text {
      animation: processing-dots 1.5s infinite;
    }
  }
}

.cancel-button {
  margin-left: 8px;
}

.progress-container {
  width: 100%;
  max-width: 200px;
  
  .progress-bar {
    margin-bottom: 4px;
  }
  
  .progress-text {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    text-align: center;
  }
}

.result-tip {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  
  &--success {
    color: var(--el-color-success);
    background-color: var(--el-color-success-light-9);
  }
  
  &--error {
    color: var(--el-color-error);
    background-color: var(--el-color-error-light-9);
  }
  
  .result-icon {
    font-size: 14px;
  }
}

// 动画
@keyframes success-pulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.05); }
  100% { transform: scale(1); }
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-4px); }
  75% { transform: translateX(4px); }
}

@keyframes processing-dots {
  0%, 20% { content: ''; }
  40% { content: '.'; }
  60% { content: '..'; }
  80%, 100% { content: '...'; }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

// 响应式设计
@media (max-width: 768px) {
  .seckill-button {
    min-width: 100px;
    font-size: 14px;
  }
  
  .progress-container {
    max-width: 150px;
  }
}
</style>
