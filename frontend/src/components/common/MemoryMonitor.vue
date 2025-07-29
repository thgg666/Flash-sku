<template>
  <div class="memory-monitor" v-if="showMonitor">
    <div class="monitor-header">
      <span class="monitor-title">内存监控</span>
      <div class="monitor-controls">
        <el-button size="small" @click="toggleAutoRefresh">
          {{ autoRefresh ? '停止' : '开始' }}
        </el-button>
        <el-button size="small" @click="refreshMetrics">
          刷新
        </el-button>
        <el-button size="small" @click="triggerCleanup">
          清理
        </el-button>
        <el-button size="small" text @click="showMonitor = false">
          ×
        </el-button>
      </div>
    </div>

    <div class="monitor-content">
      <!-- 内存使用情况 -->
      <div class="metric-section">
        <h4>内存使用</h4>
        <div class="metric-item">
          <span class="metric-label">已用内存:</span>
          <span class="metric-value">{{ formatBytes(memoryInfo.usedJSHeapSize) }}</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">总内存:</span>
          <span class="metric-value">{{ formatBytes(memoryInfo.totalJSHeapSize) }}</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">内存限制:</span>
          <span class="metric-value">{{ formatBytes(memoryInfo.jsHeapSizeLimit) }}</span>
        </div>
        <div class="memory-usage-bar">
          <div 
            class="memory-usage-fill"
            :style="{ 
              width: memoryUsagePercent + '%',
              backgroundColor: getUsageColor(memoryUsagePercent)
            }"
          ></div>
        </div>
      </div>

      <!-- 性能指标 -->
      <div class="metric-section">
        <h4>性能指标</h4>
        <div class="metric-item">
          <span class="metric-label">DOM节点:</span>
          <span class="metric-value">{{ performanceMetrics.domNodes }}</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">事件监听器:</span>
          <span class="metric-value">{{ performanceMetrics.eventListeners }}</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">渲染时间:</span>
          <span class="metric-value">{{ performanceMetrics.renderTime.toFixed(2) }}ms</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">组件数量:</span>
          <span class="metric-value">{{ performanceMetrics.componentCount }}</span>
        </div>
      </div>

      <!-- 内存历史图表 -->
      <div class="metric-section" v-if="showChart">
        <h4>内存使用历史</h4>
        <div class="memory-chart">
          <canvas ref="chartRef" width="300" height="100"></canvas>
        </div>
      </div>

      <!-- 警告信息 -->
      <div class="warnings" v-if="warnings.length > 0">
        <h4>性能警告</h4>
        <div 
          v-for="(warning, index) in warnings" 
          :key="index"
          class="warning-item"
          :class="warning.level"
        >
          <span class="warning-icon">⚠️</span>
          <span class="warning-message">{{ warning.message }}</span>
          <span class="warning-time">{{ formatTime(warning.timestamp) }}</span>
        </div>
      </div>
    </div>
  </div>

  <!-- 浮动按钮 -->
  <div 
    v-else 
    class="monitor-toggle"
    @click="showMonitor = true"
    :class="{ 'warning': hasWarnings }"
  >
    📊
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { ElButton } from 'element-plus'
import { usePerformanceOptimization } from '@/composables/usePerformanceOptimization'
import { performanceOptimizer } from '@/utils/performanceOptimizer'

interface MemoryInfo {
  usedJSHeapSize: number
  totalJSHeapSize: number
  jsHeapSizeLimit: number
}

interface Warning {
  level: 'info' | 'warning' | 'error'
  message: string
  timestamp: number
}

// 响应式数据
const showMonitor = ref(false)
const showChart = ref(true)
const autoRefresh = ref(true)
const chartRef = ref<HTMLCanvasElement>()

const memoryInfo = ref<MemoryInfo>({
  usedJSHeapSize: 0,
  totalJSHeapSize: 0,
  jsHeapSizeLimit: 0
})

const warnings = ref<Warning[]>([])
const memoryHistory = ref<number[]>([])
const maxHistoryLength = 50

// 使用性能优化
const { metrics: performanceMetrics, useMemoryGuard } = usePerformanceOptimization()
const { createSafeInterval } = useMemoryGuard()

// 计算属性
const memoryUsagePercent = computed(() => {
  if (memoryInfo.value.totalJSHeapSize === 0) return 0
  return Math.round((memoryInfo.value.usedJSHeapSize / memoryInfo.value.totalJSHeapSize) * 100)
})

const hasWarnings = computed(() => 
  warnings.value.some(w => w.level === 'warning' || w.level === 'error')
)

// 方法
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatTime = (timestamp: number): string => {
  return new Date(timestamp).toLocaleTimeString()
}

const getUsageColor = (percent: number): string => {
  if (percent < 50) return '#52c41a'
  if (percent < 80) return '#fa8c16'
  return '#ff4d4f'
}

const refreshMetrics = () => {
  // 获取内存信息
  if ('memory' in performance) {
    const memory = (performance as any).memory
    memoryInfo.value = {
      usedJSHeapSize: memory.usedJSHeapSize,
      totalJSHeapSize: memory.totalJSHeapSize,
      jsHeapSizeLimit: memory.jsHeapSizeLimit
    }

    // 添加到历史记录
    memoryHistory.value.push(memory.usedJSHeapSize)
    if (memoryHistory.value.length > maxHistoryLength) {
      memoryHistory.value.shift()
    }

    // 检查警告
    checkWarnings()
    
    // 更新图表
    if (showChart.value) {
      nextTick(() => {
        drawChart()
      })
    }
  }
}

const checkWarnings = () => {
  const now = Date.now()
  const memUsage = memoryInfo.value.usedJSHeapSize
  const memLimit = memoryInfo.value.jsHeapSizeLimit
  const domNodes = performanceMetrics.value.domNodes
  const renderTime = performanceMetrics.value.renderTime

  // 内存使用警告
  if (memUsage > memLimit * 0.8) {
    addWarning('error', '内存使用率超过80%，建议清理内存', now)
  } else if (memUsage > memLimit * 0.6) {
    addWarning('warning', '内存使用率超过60%', now)
  }

  // DOM节点警告
  if (domNodes > 5000) {
    addWarning('warning', `DOM节点过多 (${domNodes})，可能影响性能`, now)
  }

  // 渲染时间警告
  if (renderTime > 50) {
    addWarning('warning', `渲染时间过长 (${renderTime.toFixed(2)}ms)`, now)
  }
}

const addWarning = (level: Warning['level'], message: string, timestamp: number) => {
  // 避免重复警告
  const exists = warnings.value.some(w => 
    w.message === message && timestamp - w.timestamp < 5000
  )
  
  if (!exists) {
    warnings.value.unshift({ level, message, timestamp })
    
    // 限制警告数量
    if (warnings.value.length > 10) {
      warnings.value.pop()
    }
  }
}

const drawChart = () => {
  if (!chartRef.value || memoryHistory.value.length === 0) return

  const canvas = chartRef.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const width = canvas.width
  const height = canvas.height
  const data = memoryHistory.value

  // 清空画布
  ctx.clearRect(0, 0, width, height)

  // 计算缩放
  const maxValue = Math.max(...data)
  const minValue = Math.min(...data)
  const range = maxValue - minValue || 1

  // 绘制网格
  ctx.strokeStyle = '#f0f0f0'
  ctx.lineWidth = 1
  for (let i = 0; i <= 4; i++) {
    const y = (height / 4) * i
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(width, y)
    ctx.stroke()
  }

  // 绘制数据线
  ctx.strokeStyle = '#1890ff'
  ctx.lineWidth = 2
  ctx.beginPath()

  data.forEach((value, index) => {
    const x = (width / (data.length - 1)) * index
    const y = height - ((value - minValue) / range) * height
    
    if (index === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  })

  ctx.stroke()

  // 绘制当前值点
  if (data.length > 0) {
    const lastValue = data[data.length - 1]
    const x = width
    const y = height - ((lastValue - minValue) / range) * height
    
    ctx.fillStyle = '#1890ff'
    ctx.beginPath()
    ctx.arc(x, y, 3, 0, 2 * Math.PI)
    ctx.fill()
  }
}

const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
}

const triggerCleanup = () => {
  performanceOptimizer.cleanup()
  
  // 强制垃圾回收（如果可用）
  if ('gc' in window) {
    (window as any).gc()
  }
  
  addWarning('info', '手动清理完成', Date.now())
  
  // 刷新指标
  setTimeout(() => {
    refreshMetrics()
  }, 1000)
}

// 生命周期
onMounted(() => {
  // 初始刷新
  refreshMetrics()

  // 自动刷新
  createSafeInterval(() => {
    if (autoRefresh.value) {
      refreshMetrics()
    }
  }, 2000)
})
</script>

<style scoped>
.memory-monitor {
  position: fixed;
  top: 20px;
  right: 20px;
  width: 350px;
  max-height: 80vh;
  background: white;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 9999;
  overflow: hidden;
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f5f5f5;
  border-bottom: 1px solid #e8e8e8;
}

.monitor-title {
  font-weight: 600;
  font-size: 14px;
}

.monitor-controls {
  display: flex;
  gap: 8px;
}

.monitor-content {
  padding: 16px;
  max-height: calc(80vh - 60px);
  overflow-y: auto;
}

.metric-section {
  margin-bottom: 20px;
}

.metric-section h4 {
  margin: 0 0 12px 0;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.metric-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
  font-size: 12px;
}

.metric-label {
  color: #666;
}

.metric-value {
  font-weight: 500;
  color: #333;
}

.memory-usage-bar {
  width: 100%;
  height: 6px;
  background: #f0f0f0;
  border-radius: 3px;
  margin-top: 8px;
  overflow: hidden;
}

.memory-usage-fill {
  height: 100%;
  transition: width 0.3s ease;
}

.memory-chart {
  margin-top: 8px;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  overflow: hidden;
}

.warnings {
  border-top: 1px solid #e8e8e8;
  padding-top: 12px;
}

.warning-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  margin-bottom: 4px;
  border-radius: 4px;
  font-size: 12px;
}

.warning-item.warning {
  background: #fff7e6;
  border: 1px solid #ffd591;
}

.warning-item.error {
  background: #fff2f0;
  border: 1px solid #ffccc7;
}

.warning-item.info {
  background: #f6ffed;
  border: 1px solid #b7eb8f;
}

.warning-icon {
  flex-shrink: 0;
}

.warning-message {
  flex: 1;
  color: #333;
}

.warning-time {
  flex-shrink: 0;
  color: #999;
  font-size: 11px;
}

.monitor-toggle {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 50px;
  height: 50px;
  background: #1890ff;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  z-index: 9998;
  font-size: 20px;
  transition: all 0.3s ease;
}

.monitor-toggle:hover {
  transform: scale(1.1);
}

.monitor-toggle.warning {
  background: #fa8c16;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  }
  50% {
    box-shadow: 0 2px 8px rgba(250, 140, 22, 0.5);
  }
  100% {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  }
}

/* 响应式 */
@media (max-width: 768px) {
  .memory-monitor {
    width: calc(100vw - 40px);
    right: 20px;
  }
}
</style>
