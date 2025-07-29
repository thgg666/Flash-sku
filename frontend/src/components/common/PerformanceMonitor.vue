<template>
  <div class="performance-monitor">
    <el-card class="monitor-card">
      <template #header>
        <div class="card-header">
          <h3>性能监控面板</h3>
          <div class="header-actions">
            <el-button @click="refreshMetrics" :loading="refreshing" size="small" type="primary">
              刷新数据
            </el-button>
            <el-button @click="exportReport" size="small">
              导出报告
            </el-button>
          </div>
        </div>
      </template>

      <!-- 核心性能指标 -->
      <div class="metrics-section">
        <h4>核心性能指标</h4>
        <el-row :gutter="16">
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value" :class="getMetricClass('loadTime')">
                {{ formatTime(metrics.loadTime) }}
              </div>
              <div class="metric-label">页面加载时间</div>
              <div class="metric-target">目标: &lt; 3s</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value" :class="getMetricClass('firstContentfulPaint')">
                {{ formatTime(metrics.firstContentfulPaint) }}
              </div>
              <div class="metric-label">首次内容绘制</div>
              <div class="metric-target">目标: &lt; 1.5s</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value" :class="getMetricClass('largestContentfulPaint')">
                {{ formatTime(metrics.largestContentfulPaint) }}
              </div>
              <div class="metric-label">最大内容绘制</div>
              <div class="metric-target">目标: &lt; 2.5s</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value" :class="getMetricClass('cumulativeLayoutShift')">
                {{ formatCLS(metrics.cumulativeLayoutShift) }}
              </div>
              <div class="metric-label">累积布局偏移</div>
              <div class="metric-target">目标: &lt; 0.1</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 资源加载性能 -->
      <div class="metrics-section">
        <h4>资源加载性能</h4>
        <el-row :gutter="16">
          <el-col :span="8">
            <div class="metric-card">
              <div class="metric-value">{{ formatTime(metrics.jsLoadTime) }}</div>
              <div class="metric-label">JavaScript加载</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="metric-card">
              <div class="metric-value">{{ formatTime(metrics.cssLoadTime) }}</div>
              <div class="metric-label">CSS加载</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="metric-card">
              <div class="metric-value">{{ formatTime(metrics.imageLoadTime) }}</div>
              <div class="metric-label">图片加载</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 内存使用情况 -->
      <div class="metrics-section">
        <h4>内存使用情况</h4>
        <el-row :gutter="16">
          <el-col :span="12">
            <div class="metric-card">
              <div class="metric-value">{{ formatMemory(metrics.memoryUsage) }}</div>
              <div class="metric-label">已用内存</div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="metric-card">
              <div class="metric-value">{{ formatMemory(metrics.jsHeapSize) }}</div>
              <div class="metric-label">JS堆大小</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 网络信息 -->
      <div class="metrics-section">
        <h4>网络信息</h4>
        <el-row :gutter="16">
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value">{{ metrics.connectionType || 'unknown' }}</div>
              <div class="metric-label">连接类型</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value">{{ metrics.effectiveType || 'unknown' }}</div>
              <div class="metric-label">有效类型</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value">{{ formatSpeed(metrics.downlink) }}</div>
              <div class="metric-label">下行速度</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value">{{ formatTime(metrics.rtt) }}</div>
              <div class="metric-label">往返时间</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 性能建议 -->
      <div class="metrics-section" v-if="suggestions.length > 0">
        <h4>性能优化建议</h4>
        <el-alert
          v-for="(suggestion, index) in suggestions"
          :key="index"
          :title="suggestion"
          type="warning"
          :closable="false"
          style="margin-bottom: 8px;"
        />
      </div>

      <!-- 性能趋势图 -->
      <div class="metrics-section">
        <h4>性能趋势</h4>
        <div class="chart-container">
          <canvas ref="chartCanvas" width="800" height="300"></canvas>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { getPerformanceMetrics, PerformanceOptimizer, type PerformanceMetrics } from '@/utils/performance'

// 状态
const refreshing = ref(false)
const chartCanvas = ref<HTMLCanvasElement>()
const metrics = reactive<Partial<PerformanceMetrics>>({})
const suggestions = ref<string[]>([])
const performanceHistory = ref<Array<{ timestamp: number; metrics: Partial<PerformanceMetrics> }>>([])

// 定时器
let refreshTimer: number | null = null

// 初始化
onMounted(() => {
  refreshMetrics()
  startAutoRefresh()
  nextTick(() => {
    initChart()
  })
})

onUnmounted(() => {
  stopAutoRefresh()
})

// 刷新性能指标
const refreshMetrics = async () => {
  refreshing.value = true
  try {
    const currentMetrics = getPerformanceMetrics()
    if (currentMetrics) {
      Object.assign(metrics, currentMetrics)
      
      // 生成优化建议
      suggestions.value = PerformanceOptimizer.analyzeAndSuggest(currentMetrics)
      
      // 记录历史数据
      performanceHistory.value.push({
        timestamp: Date.now(),
        metrics: { ...currentMetrics }
      })
      
      // 保持最近20条记录
      if (performanceHistory.value.length > 20) {
        performanceHistory.value.shift()
      }
      
      // 更新图表
      updateChart()
    }
  } catch (error) {
    console.error('Failed to refresh metrics:', error)
    ElMessage.error('刷新性能数据失败')
  } finally {
    refreshing.value = false
  }
}

// 开始自动刷新
const startAutoRefresh = () => {
  refreshTimer = window.setInterval(refreshMetrics, 5000) // 每5秒刷新一次
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 导出报告
const exportReport = () => {
  const report = {
    timestamp: new Date().toISOString(),
    metrics,
    suggestions: suggestions.value,
    history: performanceHistory.value
  }
  
  const blob = new Blob([JSON.stringify(report, null, 2)], {
    type: 'application/json'
  })
  
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `performance-report-${Date.now()}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  ElMessage.success('性能报告导出成功')
}

// 格式化时间
const formatTime = (time?: number) => {
  if (time === undefined || time === null) return 'N/A'
  if (time < 1000) return `${Math.round(time)}ms`
  return `${(time / 1000).toFixed(2)}s`
}

// 格式化内存
const formatMemory = (bytes?: number) => {
  if (bytes === undefined || bytes === null) return 'N/A'
  const mb = bytes / (1024 * 1024)
  return `${mb.toFixed(2)}MB`
}

// 格式化速度
const formatSpeed = (mbps?: number) => {
  if (mbps === undefined || mbps === null) return 'N/A'
  return `${mbps.toFixed(2)}Mbps`
}

// 格式化CLS
const formatCLS = (cls?: number) => {
  if (cls === undefined || cls === null) return 'N/A'
  return cls.toFixed(3)
}

// 获取指标样式类
const getMetricClass = (metricName: string) => {
  const value = metrics[metricName as keyof PerformanceMetrics]
  if (value === undefined || value === null) return ''
  
  switch (metricName) {
    case 'loadTime':
      return value < 3000 ? 'good' : value < 5000 ? 'needs-improvement' : 'poor'
    case 'firstContentfulPaint':
      return value < 1500 ? 'good' : value < 2500 ? 'needs-improvement' : 'poor'
    case 'largestContentfulPaint':
      return value < 2500 ? 'good' : value < 4000 ? 'needs-improvement' : 'poor'
    case 'cumulativeLayoutShift':
      return value < 0.1 ? 'good' : value < 0.25 ? 'needs-improvement' : 'poor'
    default:
      return ''
  }
}

// 初始化图表
const initChart = () => {
  if (!chartCanvas.value) return
  
  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return
  
  // 简单的图表绘制
  drawChart(ctx)
}

// 更新图表
const updateChart = () => {
  if (!chartCanvas.value) return
  
  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return
  
  drawChart(ctx)
}

// 绘制图表
const drawChart = (ctx: CanvasRenderingContext2D) => {
  const canvas = ctx.canvas
  const width = canvas.width
  const height = canvas.height
  
  // 清空画布
  ctx.clearRect(0, 0, width, height)
  
  if (performanceHistory.value.length < 2) return
  
  // 绘制加载时间趋势
  ctx.strokeStyle = '#409EFF'
  ctx.lineWidth = 2
  ctx.beginPath()
  
  const maxTime = Math.max(...performanceHistory.value.map(h => h.metrics.loadTime || 0))
  const minTime = Math.min(...performanceHistory.value.map(h => h.metrics.loadTime || 0))
  const timeRange = maxTime - minTime || 1
  
  performanceHistory.value.forEach((history, index) => {
    const x = (index / (performanceHistory.value.length - 1)) * width
    const y = height - ((history.metrics.loadTime || 0) - minTime) / timeRange * height
    
    if (index === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  })
  
  ctx.stroke()
  
  // 绘制标签
  ctx.fillStyle = '#666'
  ctx.font = '12px Arial'
  ctx.fillText('加载时间趋势', 10, 20)
  ctx.fillText(`${formatTime(minTime)} - ${formatTime(maxTime)}`, 10, height - 10)
}
</script>

<style scoped lang="scss">
@import "@/styles/variables.scss";

.performance-monitor {
  .monitor-card {
    max-width: 1200px;
    margin: 0 auto;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    h3 {
      margin: 0;
      color: $text-color-primary;
    }

    .header-actions {
      .el-button {
        margin-left: 8px;
      }
    }
  }

  .metrics-section {
    margin-bottom: 32px;

    h4 {
      color: $text-color-primary;
      margin-bottom: 16px;
      font-size: 16px;
      font-weight: 600;
    }
  }

  .metric-card {
    text-align: center;
    padding: 20px;
    background: $bg-color-page;
    border-radius: 8px;
    border: 1px solid $border-color-lighter;

    .metric-value {
      font-size: 24px;
      font-weight: 600;
      margin-bottom: 8px;

      &.good {
        color: $success-color;
      }

      &.needs-improvement {
        color: $warning-color;
      }

      &.poor {
        color: $danger-color;
      }
    }

    .metric-label {
      font-size: 14px;
      color: $text-color-regular;
      margin-bottom: 4px;
    }

    .metric-target {
      font-size: 12px;
      color: $text-color-secondary;
    }
  }

  .chart-container {
    background: $bg-color-page;
    border-radius: 8px;
    padding: 16px;
    border: 1px solid $border-color-lighter;

    canvas {
      width: 100%;
      height: 300px;
    }
  }
}

// 移动端适配
@media (max-width: 768px) {
  .performance-monitor {
    .card-header {
      flex-direction: column;
      gap: 12px;
      align-items: stretch;

      .header-actions {
        display: flex;
        gap: 8px;

        .el-button {
          flex: 1;
          margin-left: 0;
        }
      }
    }

    .metric-card {
      padding: 16px;

      .metric-value {
        font-size: 20px;
      }
    }
  }
}
</style>
