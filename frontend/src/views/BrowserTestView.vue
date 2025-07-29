<template>
  <div class="browser-test-view">
    <div class="container">
      <div class="page-header">
        <h1>浏览器兼容性测试</h1>
        <p>检测当前浏览器对Flash Sku应用的支持情况</p>
      </div>

      <!-- 浏览器兼容性检测器 -->
      <BrowserCompatibilityChecker />

      <!-- 可访问性测试器 -->
      <AccessibilityTester />

      <!-- 性能监控面板 -->
      <PerformanceMonitor />

      <!-- 功能测试区域 -->
      <div class="test-sections">
        <!-- WebSocket测试 -->
        <el-card class="test-card">
          <template #header>
            <h3>WebSocket连接测试</h3>
          </template>
          <div class="test-content">
            <p>测试WebSocket实时通信功能</p>
            <el-button @click="testWebSocket" :loading="wsLoading" type="primary">
              测试WebSocket连接
            </el-button>
            <div v-if="wsResult" class="test-result">
              <el-tag :type="wsResult.success ? 'success' : 'danger'">
                {{ wsResult.message }}
              </el-tag>
            </div>
          </div>
        </el-card>

        <!-- 本地存储测试 -->
        <el-card class="test-card">
          <template #header>
            <h3>本地存储测试</h3>
          </template>
          <div class="test-content">
            <p>测试localStorage和sessionStorage功能</p>
            <el-button @click="testStorage" type="primary">
              测试本地存储
            </el-button>
            <div v-if="storageResult" class="test-result">
              <el-tag :type="storageResult.success ? 'success' : 'danger'">
                {{ storageResult.message }}
              </el-tag>
            </div>
          </div>
        </el-card>

        <!-- Fetch API测试 -->
        <el-card class="test-card">
          <template #header>
            <h3>Fetch API测试</h3>
          </template>
          <div class="test-content">
            <p>测试现代网络请求API</p>
            <el-button @click="testFetch" :loading="fetchLoading" type="primary">
              测试Fetch API
            </el-button>
            <div v-if="fetchResult" class="test-result">
              <el-tag :type="fetchResult.success ? 'success' : 'danger'">
                {{ fetchResult.message }}
              </el-tag>
            </div>
          </div>
        </el-card>

        <!-- CSS功能测试 -->
        <el-card class="test-card">
          <template #header>
            <h3>CSS功能测试</h3>
          </template>
          <div class="test-content">
            <p>测试现代CSS特性支持</p>
            <div class="css-tests">
              <!-- Flexbox测试 -->
              <div class="css-test-item">
                <span>Flexbox布局:</span>
                <div class="flexbox-test">
                  <div class="flex-item">1</div>
                  <div class="flex-item">2</div>
                  <div class="flex-item">3</div>
                </div>
              </div>
              
              <!-- Grid测试 -->
              <div class="css-test-item">
                <span>Grid布局:</span>
                <div class="grid-test">
                  <div class="grid-item">A</div>
                  <div class="grid-item">B</div>
                  <div class="grid-item">C</div>
                  <div class="grid-item">D</div>
                </div>
              </div>
              
              <!-- CSS变量测试 -->
              <div class="css-test-item">
                <span>CSS变量:</span>
                <div class="css-var-test">CSS变量颜色</div>
              </div>
              
              <!-- CSS动画测试 -->
              <div class="css-test-item">
                <span>CSS动画:</span>
                <div class="css-animation-test">动画效果</div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- JavaScript功能测试 -->
        <el-card class="test-card">
          <template #header>
            <h3>JavaScript功能测试</h3>
          </template>
          <div class="test-content">
            <p>测试现代JavaScript特性</p>
            <el-button @click="testJavaScript" type="primary">
              测试JavaScript功能
            </el-button>
            <div v-if="jsResult" class="test-result">
              <div v-for="(result, feature) in jsResult" :key="feature" class="js-test-result">
                <span>{{ feature }}:</span>
                <el-tag :type="result ? 'success' : 'danger'" size="small">
                  {{ result ? '支持' : '不支持' }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 性能测试 -->
        <el-card class="test-card">
          <template #header>
            <h3>性能测试</h3>
          </template>
          <div class="test-content">
            <p>测试浏览器性能指标</p>
            <el-button @click="testPerformance" :loading="perfLoading" type="primary">
              运行性能测试
            </el-button>
            <div v-if="perfResult" class="test-result">
              <div class="perf-metrics">
                <div class="metric-item">
                  <span>页面加载时间:</span>
                  <strong>{{ perfResult.loadTime }}ms</strong>
                </div>
                <div class="metric-item">
                  <span>DOM解析时间:</span>
                  <strong>{{ perfResult.domTime }}ms</strong>
                </div>
                <div class="metric-item">
                  <span>资源加载时间:</span>
                  <strong>{{ perfResult.resourceTime }}ms</strong>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 测试报告 -->
      <div class="test-report">
        <el-card>
          <template #header>
            <h3>兼容性测试报告</h3>
          </template>
          <div class="report-content">
            <el-button @click="generateReport" type="success">
              生成测试报告
            </el-button>
            <el-button @click="downloadReport" :disabled="!reportData" type="primary">
              下载报告
            </el-button>
            
            <div v-if="reportData" class="report-summary">
              <h4>测试摘要</h4>
              <p>浏览器: {{ reportData.browser }}</p>
              <p>兼容性评分: {{ reportData.score }}/100</p>
              <p>测试时间: {{ reportData.timestamp }}</p>
            </div>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import BrowserCompatibilityChecker from '@/components/common/BrowserCompatibilityChecker.vue'
import AccessibilityTester from '@/components/common/AccessibilityTester.vue'
import PerformanceMonitor from '@/components/common/PerformanceMonitor.vue'
import { detectBrowser } from '@/utils/browserCompatibility'

// 测试状态
const wsLoading = ref(false)
const fetchLoading = ref(false)
const perfLoading = ref(false)

// 测试结果
const wsResult = ref<{success: boolean, message: string} | null>(null)
const storageResult = ref<{success: boolean, message: string} | null>(null)
const fetchResult = ref<{success: boolean, message: string} | null>(null)
const jsResult = ref<Record<string, boolean> | null>(null)
const perfResult = ref<{loadTime: number, domTime: number, resourceTime: number} | null>(null)
const reportData = ref<any>(null)

// WebSocket测试
const testWebSocket = async () => {
  wsLoading.value = true
  try {
    if (!window.WebSocket) {
      wsResult.value = { success: false, message: '浏览器不支持WebSocket' }
      return
    }

    // 尝试连接WebSocket (这里使用一个测试地址)
    const ws = new WebSocket('wss://echo.websocket.org/')
    
    await new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        ws.close()
        reject(new Error('连接超时'))
      }, 5000)

      ws.onopen = () => {
        clearTimeout(timeout)
        ws.send('test message')
      }

      ws.onmessage = (event) => {
        if (event.data === 'test message') {
          ws.close()
          resolve(true)
        }
      }

      ws.onerror = () => {
        clearTimeout(timeout)
        reject(new Error('连接失败'))
      }
    })

    wsResult.value = { success: true, message: 'WebSocket连接测试成功' }
  } catch (error: any) {
    wsResult.value = { success: false, message: `WebSocket测试失败: ${error.message}` }
  } finally {
    wsLoading.value = false
  }
}

// 本地存储测试
const testStorage = () => {
  try {
    // 测试localStorage
    const testKey = 'browser_test_' + Date.now()
    const testValue = 'test_value'
    
    localStorage.setItem(testKey, testValue)
    const retrieved = localStorage.getItem(testKey)
    localStorage.removeItem(testKey)
    
    if (retrieved !== testValue) {
      throw new Error('localStorage读写失败')
    }
    
    // 测试sessionStorage
    sessionStorage.setItem(testKey, testValue)
    const sessionRetrieved = sessionStorage.getItem(testKey)
    sessionStorage.removeItem(testKey)
    
    if (sessionRetrieved !== testValue) {
      throw new Error('sessionStorage读写失败')
    }
    
    storageResult.value = { success: true, message: '本地存储测试成功' }
  } catch (error: any) {
    storageResult.value = { success: false, message: `本地存储测试失败: ${error.message}` }
  }
}

// Fetch API测试
const testFetch = async () => {
  fetchLoading.value = true
  try {
    if (!window.fetch) {
      fetchResult.value = { success: false, message: '浏览器不支持Fetch API' }
      return
    }

    const response = await fetch('https://httpbin.org/json', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const data = await response.json()
    if (data && typeof data === 'object') {
      fetchResult.value = { success: true, message: 'Fetch API测试成功' }
    } else {
      throw new Error('响应数据格式错误')
    }
  } catch (error: any) {
    fetchResult.value = { success: false, message: `Fetch API测试失败: ${error.message}` }
  } finally {
    fetchLoading.value = false
  }
}

// JavaScript功能测试
const testJavaScript = () => {
  const results: Record<string, boolean> = {}
  
  // 测试ES6箭头函数
  try {
    eval('(() => {})')
    results['箭头函数'] = true
  } catch {
    results['箭头函数'] = false
  }
  
  // 测试解构赋值
  try {
    eval('const {a} = {a: 1}')
    results['解构赋值'] = true
  } catch {
    results['解构赋值'] = false
  }
  
  // 测试模板字符串
  try {
    eval('`template ${1} string`')
    results['模板字符串'] = true
  } catch {
    results['模板字符串'] = false
  }
  
  // 测试Promise
  results['Promise'] = 'Promise' in window
  
  // 测试async/await
  try {
    eval('(async function() { await Promise.resolve() })')
    results['Async/Await'] = true
  } catch {
    results['Async/Await'] = false
  }
  
  jsResult.value = results
}

// 性能测试
const testPerformance = async () => {
  perfLoading.value = true
  try {
    if (!window.performance) {
      ElMessage.warning('浏览器不支持Performance API')
      return
    }

    const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
    
    perfResult.value = {
      loadTime: Math.round(navigation.loadEventEnd - navigation.navigationStart),
      domTime: Math.round(navigation.domContentLoadedEventEnd - navigation.navigationStart),
      resourceTime: Math.round(navigation.loadEventEnd - navigation.domContentLoadedEventEnd)
    }
  } catch (error) {
    ElMessage.error('性能测试失败')
  } finally {
    perfLoading.value = false
  }
}

// 生成测试报告
const generateReport = () => {
  const browserInfo = detectBrowser()
  
  // 计算兼容性评分
  let score = 0
  const features = browserInfo.features
  
  if (features.webSocket) score += 15
  if (features.localStorage) score += 15
  if (features.js.fetch) score += 15
  if (features.css.flexbox) score += 10
  if (features.css.grid) score += 10
  if (features.js.es6) score += 15
  if (features.js.asyncAwait) score += 10
  if (features.js.promises) score += 10
  
  reportData.value = {
    browser: `${browserInfo.name} ${browserInfo.version}`,
    score,
    timestamp: new Date().toLocaleString(),
    details: {
      browserInfo,
      testResults: {
        webSocket: wsResult.value,
        storage: storageResult.value,
        fetch: fetchResult.value,
        javascript: jsResult.value,
        performance: perfResult.value
      }
    }
  }
  
  ElMessage.success('测试报告生成成功')
}

// 下载报告
const downloadReport = () => {
  if (!reportData.value) return
  
  const blob = new Blob([JSON.stringify(reportData.value, null, 2)], {
    type: 'application/json'
  })
  
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `browser-compatibility-report-${Date.now()}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  ElMessage.success('报告下载成功')
}
</script>

<style scoped lang="scss">
@import "@/styles/variables.scss";

.browser-test-view {
  min-height: 100vh;
  background: $bg-color-page;
  padding: 20px;

  .container {
    max-width: 1200px;
    margin: 0 auto;
  }

  .page-header {
    text-align: center;
    margin-bottom: 32px;

    h1 {
      color: $text-color-primary;
      margin-bottom: 8px;
    }

    p {
      color: $text-color-regular;
    }
  }

  .test-sections {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 24px;
    margin-bottom: 32px;

    .test-card {
      .test-content {
        p {
          margin-bottom: 16px;
          color: $text-color-regular;
        }

        .test-result {
          margin-top: 16px;
        }

        .css-tests {
          .css-test-item {
            margin-bottom: 16px;
            
            > span {
              display: inline-block;
              width: 100px;
              font-weight: 500;
            }
          }

          .flexbox-test {
            display: flex;
            gap: 8px;
            margin-top: 8px;

            .flex-item {
              flex: 1;
              padding: 8px;
              background: $primary-color;
              color: white;
              text-align: center;
              border-radius: 4px;
            }
          }

          .grid-test {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 8px;
            margin-top: 8px;

            .grid-item {
              padding: 8px;
              background: $success-color;
              color: white;
              text-align: center;
              border-radius: 4px;
            }
          }

          .css-var-test {
            --test-color: #{$warning-color};
            color: var(--test-color);
            font-weight: 500;
            margin-top: 8px;
          }

          .css-animation-test {
            margin-top: 8px;
            padding: 8px 16px;
            background: $info-color;
            color: white;
            border-radius: 4px;
            animation: pulse 2s infinite;
          }

          @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.7; }
          }
        }

        .js-test-result {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 4px 0;
          border-bottom: 1px solid $border-color-lighter;

          &:last-child {
            border-bottom: none;
          }
        }

        .perf-metrics {
          .metric-item {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid $border-color-lighter;

            &:last-child {
              border-bottom: none;
            }
          }
        }
      }
    }
  }

  .test-report {
    .report-content {
      .el-button {
        margin-right: 12px;
        margin-bottom: 16px;
      }

      .report-summary {
        margin-top: 16px;
        padding: 16px;
        background: $bg-color;
        border-radius: 8px;

        h4 {
          margin-bottom: 12px;
          color: $text-color-primary;
        }

        p {
          margin: 4px 0;
          color: $text-color-regular;
        }
      }
    }
  }
}

// 移动端适配
@media (max-width: 768px) {
  .browser-test-view {
    padding: 16px;

    .test-sections {
      grid-template-columns: 1fr;
    }
  }
}
</style>
