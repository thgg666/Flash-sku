<template>
  <div class="seckill-animations">
    <!-- 成功动画 -->
    <transition name="success-animation" appear>
      <div v-if="showSuccess" class="success-overlay">
        <div class="success-content">
          <div class="success-icon">
            <el-icon class="check-icon">
              <SuccessFilled />
            </el-icon>
          </div>
          <div class="success-text">
            <h3>抢购成功！</h3>
            <p>{{ successMessage }}</p>
          </div>
          <div class="success-effects">
            <div class="firework firework-1"></div>
            <div class="firework firework-2"></div>
            <div class="firework firework-3"></div>
          </div>
        </div>
      </div>
    </transition>

    <!-- 失败动画 -->
    <transition name="fail-animation" appear>
      <div v-if="showFail" class="fail-overlay">
        <div class="fail-content">
          <div class="fail-icon">
            <el-icon class="close-icon">
              <CircleCloseFilled />
            </el-icon>
          </div>
          <div class="fail-text">
            <h3>抢购失败</h3>
            <p>{{ failMessage }}</p>
          </div>
          <div class="fail-effects">
            <div class="shake-effect"></div>
          </div>
        </div>
      </div>
    </transition>

    <!-- 加载动画 -->
    <transition name="loading-animation" appear>
      <div v-if="showLoading" class="loading-overlay">
        <div class="loading-content">
          <div class="loading-spinner">
            <div class="spinner-ring"></div>
            <div class="spinner-ring"></div>
            <div class="spinner-ring"></div>
          </div>
          <div class="loading-text">
            <h3>{{ loadingTitle }}</h3>
            <p>{{ loadingMessage }}</p>
          </div>
          <div class="loading-progress">
            <div class="progress-bar" :style="{ width: `${progress}%` }"></div>
          </div>
        </div>
      </div>
    </transition>

    <!-- 倒计时动画 -->
    <transition name="countdown-animation" appear>
      <div v-if="showCountdown" class="countdown-overlay">
        <div class="countdown-content">
          <div class="countdown-circle">
            <svg class="countdown-svg" viewBox="0 0 100 100">
              <circle
                class="countdown-bg"
                cx="50"
                cy="50"
                r="45"
                fill="none"
                stroke="#f0f0f0"
                stroke-width="8"
              />
              <circle
                class="countdown-progress"
                cx="50"
                cy="50"
                r="45"
                fill="none"
                stroke="#409eff"
                stroke-width="8"
                stroke-linecap="round"
                :stroke-dasharray="circumference"
                :stroke-dashoffset="strokeDashoffset"
                transform="rotate(-90 50 50)"
              />
            </svg>
            <div class="countdown-number">{{ countdownValue }}</div>
          </div>
          <div class="countdown-text">
            <p>{{ countdownMessage }}</p>
          </div>
        </div>
      </div>
    </transition>

    <!-- 粒子效果 -->
    <div v-if="showParticles" class="particles-container">
      <div
        v-for="particle in particles"
        :key="particle.id"
        class="particle"
        :style="particle.style"
      ></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElIcon } from 'element-plus'
import { SuccessFilled, CircleCloseFilled } from '@element-plus/icons-vue'

// 组件属性
interface Props {
  showSuccess?: boolean
  showFail?: boolean
  showLoading?: boolean
  showCountdown?: boolean
  showParticles?: boolean
  successMessage?: string
  failMessage?: string
  loadingTitle?: string
  loadingMessage?: string
  countdownMessage?: string
  countdownValue?: number
  countdownMax?: number
  progress?: number
  duration?: number
}

const props = withDefaults(defineProps<Props>(), {
  showSuccess: false,
  showFail: false,
  showLoading: false,
  showCountdown: false,
  showParticles: false,
  successMessage: '恭喜您抢购成功！',
  failMessage: '很遗憾，抢购失败了',
  loadingTitle: '正在抢购中...',
  loadingMessage: '请稍候，正在为您处理',
  countdownMessage: '秒杀即将开始',
  countdownValue: 3,
  countdownMax: 3,
  progress: 0,
  duration: 3000
})

// 组件事件
interface Emits {
  animationEnd: [type: string]
  countdownEnd: []
}

const emit = defineEmits<Emits>()

// 响应式数据
const particles = ref<Array<{
  id: number
  style: Record<string, string>
}>>([])

// 计算属性
const circumference = computed(() => 2 * Math.PI * 45)
const strokeDashoffset = computed(() => {
  const progress = props.countdownMax > 0 ? props.countdownValue / props.countdownMax : 0
  return circumference.value * (1 - progress)
})

// 粒子效果
const createParticles = () => {
  particles.value = []
  for (let i = 0; i < 20; i++) {
    particles.value.push({
      id: i,
      style: {
        left: `${Math.random() * 100}%`,
        top: `${Math.random() * 100}%`,
        animationDelay: `${Math.random() * 2}s`,
        animationDuration: `${2 + Math.random() * 3}s`
      }
    })
  }
}

// 监听动画状态变化
watch(() => props.showSuccess, (newVal) => {
  if (newVal) {
    createParticles()
    setTimeout(() => {
      emit('animationEnd', 'success')
    }, props.duration)
  }
})

watch(() => props.showFail, (newVal) => {
  if (newVal) {
    setTimeout(() => {
      emit('animationEnd', 'fail')
    }, props.duration)
  }
})

watch(() => props.showLoading, (newVal) => {
  if (newVal) {
    setTimeout(() => {
      emit('animationEnd', 'loading')
    }, props.duration)
  }
})

watch(() => props.countdownValue, (newVal) => {
  if (newVal <= 0) {
    emit('countdownEnd')
  }
})

onMounted(() => {
  if (props.showParticles) {
    createParticles()
  }
})

onUnmounted(() => {
  particles.value = []
})
</script>

<style scoped lang="scss">
.seckill-animations {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  pointer-events: none;
  z-index: 9999;
}

// 成功动画
.success-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: auto;
}

.success-content {
  text-align: center;
  color: white;
  position: relative;
  
  .success-icon {
    margin-bottom: 20px;
    
    .check-icon {
      font-size: 80px;
      color: #67c23a;
      animation: success-bounce 0.6s ease-out;
    }
  }
  
  .success-text {
    h3 {
      font-size: 28px;
      margin-bottom: 10px;
      animation: success-fade-in 0.8s ease-out 0.3s both;
    }
    
    p {
      font-size: 16px;
      opacity: 0.9;
      animation: success-fade-in 0.8s ease-out 0.5s both;
    }
  }
}

// 失败动画
.fail-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: auto;
}

.fail-content {
  text-align: center;
  color: white;
  
  .fail-icon {
    margin-bottom: 20px;
    
    .close-icon {
      font-size: 80px;
      color: #f56c6c;
      animation: fail-shake 0.6s ease-out;
    }
  }
  
  .fail-text {
    h3 {
      font-size: 28px;
      margin-bottom: 10px;
      animation: fail-fade-in 0.8s ease-out 0.3s both;
    }
    
    p {
      font-size: 16px;
      opacity: 0.9;
      animation: fail-fade-in 0.8s ease-out 0.5s both;
    }
  }
}

// 加载动画
.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: auto;
}

.loading-content {
  text-align: center;
  color: white;
  
  .loading-spinner {
    position: relative;
    width: 80px;
    height: 80px;
    margin: 0 auto 20px;
    
    .spinner-ring {
      position: absolute;
      width: 100%;
      height: 100%;
      border: 4px solid transparent;
      border-top: 4px solid #409eff;
      border-radius: 50%;
      animation: loading-spin 1s linear infinite;
      
      &:nth-child(2) {
        animation-delay: 0.3s;
        border-top-color: #67c23a;
      }
      
      &:nth-child(3) {
        animation-delay: 0.6s;
        border-top-color: #e6a23c;
      }
    }
  }
  
  .loading-text {
    h3 {
      font-size: 24px;
      margin-bottom: 10px;
    }
    
    p {
      font-size: 14px;
      opacity: 0.8;
    }
  }
  
  .loading-progress {
    width: 200px;
    height: 4px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 2px;
    margin: 20px auto 0;
    overflow: hidden;
    
    .progress-bar {
      height: 100%;
      background: linear-gradient(90deg, #409eff, #67c23a);
      border-radius: 2px;
      transition: width 0.3s ease;
    }
  }
}

// 倒计时动画
.countdown-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: auto;
}

.countdown-content {
  text-align: center;
  color: white;
  
  .countdown-circle {
    position: relative;
    width: 120px;
    height: 120px;
    margin: 0 auto 20px;
    
    .countdown-svg {
      width: 100%;
      height: 100%;
      transform: rotate(-90deg);
    }
    
    .countdown-progress {
      transition: stroke-dashoffset 1s ease;
    }
    
    .countdown-number {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      font-size: 36px;
      font-weight: bold;
      color: #409eff;
    }
  }
  
  .countdown-text {
    p {
      font-size: 18px;
      margin: 0;
    }
  }
}

// 粒子效果
.particles-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.particle {
  position: absolute;
  width: 6px;
  height: 6px;
  background: #ffd700;
  border-radius: 50%;
  animation: particle-float 3s ease-out infinite;
}

// 烟花效果
.firework {
  position: absolute;
  width: 4px;
  height: 4px;
  background: #ffd700;
  border-radius: 50%;
  
  &.firework-1 {
    top: 20%;
    left: 20%;
    animation: firework-explode 1s ease-out 0.5s;
  }
  
  &.firework-2 {
    top: 30%;
    right: 20%;
    animation: firework-explode 1s ease-out 0.8s;
  }
  
  &.firework-3 {
    bottom: 30%;
    left: 50%;
    animation: firework-explode 1s ease-out 1.1s;
  }
}

// 动画定义
@keyframes success-bounce {
  0% { transform: scale(0); }
  50% { transform: scale(1.2); }
  100% { transform: scale(1); }
}

@keyframes success-fade-in {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes fail-shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-10px); }
  75% { transform: translateX(10px); }
}

@keyframes fail-fade-in {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes loading-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes particle-float {
  0% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
  100% {
    opacity: 0;
    transform: translateY(-100px) scale(0);
  }
}

@keyframes firework-explode {
  0% {
    opacity: 1;
    transform: scale(0);
  }
  50% {
    opacity: 1;
    transform: scale(1);
  }
  100% {
    opacity: 0;
    transform: scale(2);
  }
}

// 过渡动画
.success-animation-enter-active,
.fail-animation-enter-active,
.loading-animation-enter-active,
.countdown-animation-enter-active {
  transition: all 0.3s ease;
}

.success-animation-leave-active,
.fail-animation-leave-active,
.loading-animation-leave-active,
.countdown-animation-leave-active {
  transition: all 0.3s ease;
}

.success-animation-enter-from,
.fail-animation-enter-from,
.loading-animation-enter-from,
.countdown-animation-enter-from {
  opacity: 0;
  transform: scale(0.8);
}

.success-animation-leave-to,
.fail-animation-leave-to,
.loading-animation-leave-to,
.countdown-animation-leave-to {
  opacity: 0;
  transform: scale(1.2);
}
</style>
