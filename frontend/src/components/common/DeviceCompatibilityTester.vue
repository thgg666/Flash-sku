<template>
  <div class="device-compatibility-tester">
    <!-- 测试控制面板 -->
    <div class="test-panel">
      <h3>设备兼容性测试</h3>
      
      <!-- 当前设备信息 -->
      <div class="device-info">
        <h4>当前设备信息</h4>
        <div class="info-grid">
          <div class="info-item">
            <span class="label">设备类型:</span>
            <span class="value">{{ deviceInfo.type }}</span>
          </div>
          <div class="info-item">
            <span class="label">分辨率:</span>
            <span class="value">{{ deviceInfo.width }}x{{ deviceInfo.height }}</span>
          </div>
          <div class="info-item">
            <span class="label">像素比:</span>
            <span class="value">{{ deviceInfo.pixelRatio }}</span>
          </div>
          <div class="info-item">
            <span class="label">方向:</span>
            <span class="value">{{ deviceInfo.orientation }}</span>
          </div>
          <div class="info-item">
            <span class="label">触摸支持:</span>
            <span class="value">{{ deviceInfo.touchSupport ? '是' : '否' }}</span>
          </div>
          <div class="info-item">
            <span class="label">浏览器:</span>
            <span class="value">{{ deviceInfo.browser }}</span>
          </div>
        </div>
      </div>
      
      <!-- 断点测试 -->
      <div class="breakpoint-test">
        <h4>断点测试</h4>
        <div class="breakpoint-indicators">
          <div
            v-for="(value, name) in BREAKPOINTS"
            :key="name"
            class="breakpoint-indicator"
            :class="{ active: currentBreakpoint === name }"
          >
            {{ name }}: {{ value }}px
          </div>
        </div>
        <div class="current-breakpoint">
          当前断点: <strong>{{ currentBreakpoint }}</strong>
        </div>
      </div>
      
      <!-- 兼容性测试 -->
      <div class="compatibility-test">
        <h4>兼容性测试</h4>
        <el-button
          type="primary"
          :loading="testing"
          @click="runCompatibilityTest"
        >
          {{ testing ? '测试中...' : '运行测试' }}
        </el-button>
        
        <div v-if="testResults.size > 0" class="test-results">
          <div class="test-summary">
            <span class="passed">通过: {{ passedCount }}</span>
            <span class="failed">失败: {{ failedCount }}</span>
            <span class="total">总计: {{ testResults.size }}</span>
            <span class="rate">通过率: {{ passRate }}%</span>
          </div>
          
          <div class="test-details">
            <div
              v-for="[test, passed] in testResults"
              :key="test"
              class="test-item"
              :class="{ passed, failed: !passed }"
            >
              <el-icon>
                <Check v-if="passed" />
                <Close v-else />
              </el-icon>
              <span>{{ test }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 设备模拟器 -->
      <div class="device-simulator">
        <h4>设备模拟器</h4>
        <el-select v-model="selectedDevice" placeholder="选择设备" @change="simulateDevice">
          <el-option-group
            v-for="(devices, category) in COMMON_DEVICES"
            :key="category"
            :label="category"
          >
            <el-option
              v-for="device in devices"
              :key="device.name"
              :label="device.name"
              :value="device"
            />
          </el-option-group>
        </el-select>
        
        <div v-if="simulatedDevice" class="simulated-device-info">
          <p>模拟设备: {{ simulatedDevice.name }}</p>
          <p>分辨率: {{ simulatedDevice.width }}x{{ simulatedDevice.height }}</p>
          <p>像素比: {{ simulatedDevice.pixelRatio }}</p>
          <el-button size="small" @click="resetSimulation">重置</el-button>
        </div>
      </div>
      
      <!-- 测试报告 -->
      <div v-if="testReport" class="test-report">
        <h4>测试报告</h4>
        <el-button size="small" @click="copyReport">复制报告</el-button>
        <el-button size="small" @click="downloadReport">下载报告</el-button>
        <pre class="report-content">{{ testReport }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, Close } from '@element-plus/icons-vue'
import {
  getCurrentDeviceInfo,
  getCurrentBreakpoint,
  DeviceCompatibilityTester,
  BREAKPOINTS,
  COMMON_DEVICES,
  type DeviceInfo
} from '@/utils/deviceCompatibility'

// 响应式数据
const deviceInfo = ref<DeviceInfo>(getCurrentDeviceInfo())
const currentBreakpoint = ref(getCurrentBreakpoint())
const testing = ref(false)
const testResults = ref<Map<string, boolean>>(new Map())
const testReport = ref('')
const selectedDevice = ref()
const simulatedDevice = ref()

// 兼容性测试器
const tester = new DeviceCompatibilityTester()

// 计算属性
const passedCount = computed(() => {
  return Array.from(testResults.value.values()).filter(Boolean).length
})

const failedCount = computed(() => {
  return Array.from(testResults.value.values()).filter(v => !v).length
})

const passRate = computed(() => {
  if (testResults.value.size === 0) return 0
  return Math.round((passedCount.value / testResults.value.size) * 100)
})

// 运行兼容性测试
const runCompatibilityTest = async () => {
  testing.value = true
  try {
    const results = await tester.runAllTests()
    testResults.value = results
    testReport.value = tester.getTestReport()
    
    ElMessage.success(`测试完成！通过率: ${passRate.value}%`)
  } catch (error) {
    ElMessage.error('测试失败: ' + error)
  } finally {
    testing.value = false
  }
}

// 模拟设备
const simulateDevice = (device: any) => {
  if (!device) return
  
  simulatedDevice.value = device
  
  // 模拟视口大小（仅用于测试，实际不会改变窗口大小）
  const viewport = document.querySelector('meta[name="viewport"]')
  if (viewport) {
    viewport.setAttribute('content', `width=${device.width}, initial-scale=1.0`)
  }
  
  // 添加设备类名到body
  document.body.classList.add(`simulated-${device.name.toLowerCase().replace(/\s+/g, '-')}`)
  
  ElMessage.info(`已模拟设备: ${device.name}`)
}

// 重置模拟
const resetSimulation = () => {
  simulatedDevice.value = null
  selectedDevice.value = null
  
  // 重置视口
  const viewport = document.querySelector('meta[name="viewport"]')
  if (viewport) {
    viewport.setAttribute('content', 'width=device-width, initial-scale=1.0')
  }
  
  // 移除设备类名
  document.body.className = document.body.className.replace(/simulated-[\w-]+/g, '')
  
  ElMessage.info('已重置设备模拟')
}

// 复制报告
const copyReport = async () => {
  try {
    await navigator.clipboard.writeText(testReport.value)
    ElMessage.success('报告已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

// 下载报告
const downloadReport = () => {
  const blob = new Blob([testReport.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `device-compatibility-report-${new Date().toISOString().slice(0, 10)}.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  ElMessage.success('报告已下载')
}

// 更新设备信息
const updateDeviceInfo = () => {
  deviceInfo.value = getCurrentDeviceInfo()
  currentBreakpoint.value = getCurrentBreakpoint()
}

// 生命周期
onMounted(() => {
  window.addEventListener('resize', updateDeviceInfo)
  window.addEventListener('orientationchange', updateDeviceInfo)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateDeviceInfo)
  window.removeEventListener('orientationchange', updateDeviceInfo)
  resetSimulation()
})
</script>

<style scoped lang="scss">
.device-compatibility-tester {
  position: fixed;
  top: 20px;
  right: 20px;
  width: 400px;
  max-height: 80vh;
  background: white;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  z-index: 9999;
  overflow-y: auto;

  .test-panel {
    padding: 20px;

    h3 {
      margin: 0 0 20px 0;
      font-size: 18px;
      color: var(--el-text-color-primary);
    }

    h4 {
      margin: 20px 0 10px 0;
      font-size: 14px;
      color: var(--el-text-color-regular);
      border-bottom: 1px solid var(--el-border-color-lighter);
      padding-bottom: 5px;
    }

    .device-info {
      .info-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 8px;
        font-size: 12px;

        .info-item {
          display: flex;
          justify-content: space-between;

          .label {
            color: var(--el-text-color-regular);
          }

          .value {
            font-weight: 500;
            color: var(--el-text-color-primary);
          }
        }
      }
    }

    .breakpoint-test {
      .breakpoint-indicators {
        display: flex;
        flex-wrap: wrap;
        gap: 4px;
        margin-bottom: 10px;

        .breakpoint-indicator {
          padding: 2px 6px;
          font-size: 10px;
          background: var(--el-bg-color-page);
          border-radius: 4px;
          border: 1px solid var(--el-border-color-lighter);

          &.active {
            background: var(--el-color-primary);
            color: white;
            border-color: var(--el-color-primary);
          }
        }
      }

      .current-breakpoint {
        font-size: 12px;
        color: var(--el-text-color-regular);
      }
    }

    .compatibility-test {
      .test-results {
        margin-top: 15px;

        .test-summary {
          display: flex;
          gap: 10px;
          margin-bottom: 10px;
          font-size: 12px;

          .passed {
            color: var(--el-color-success);
          }

          .failed {
            color: var(--el-color-danger);
          }

          .total {
            color: var(--el-text-color-regular);
          }

          .rate {
            font-weight: 600;
            color: var(--el-color-primary);
          }
        }

        .test-details {
          max-height: 200px;
          overflow-y: auto;

          .test-item {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 4px 0;
            font-size: 12px;

            &.passed {
              color: var(--el-color-success);
            }

            &.failed {
              color: var(--el-color-danger);
            }

            .el-icon {
              font-size: 14px;
            }
          }
        }
      }
    }

    .device-simulator {
      .simulated-device-info {
        margin-top: 10px;
        padding: 10px;
        background: var(--el-bg-color-page);
        border-radius: 4px;
        font-size: 12px;

        p {
          margin: 0 0 5px 0;
        }
      }
    }

    .test-report {
      .report-content {
        max-height: 300px;
        overflow-y: auto;
        font-size: 10px;
        background: var(--el-bg-color-page);
        padding: 10px;
        border-radius: 4px;
        margin-top: 10px;
      }
    }
  }
}

// 开发环境显示
@media (max-width: 768px) {
  .device-compatibility-tester {
    position: relative;
    top: auto;
    right: auto;
    width: 100%;
    max-height: none;
    margin: 20px 0;
  }
}
</style>
