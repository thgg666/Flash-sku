<template>
  <div class="api-test">
    <div class="test-header">
      <h1>API集成测试</h1>
      <p>测试前后端API接口的连通性和正确性</p>
    </div>

    <div class="test-controls">
      <el-button 
        type="primary" 
        @click="runAllTests"
        :loading="isRunning"
        :disabled="isRunning"
      >
        运行所有测试
      </el-button>
      <el-button @click="clearResults" :disabled="isRunning">
        清空结果
      </el-button>
      <el-button @click="exportReport" :disabled="results.length === 0">
        导出报告
      </el-button>
    </div>

    <!-- 测试统计 -->
    <el-card class="stats-card" v-if="stats.total > 0">
      <template #header>
        <span>测试统计</span>
      </template>
      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-value">{{ stats.total }}</div>
          <div class="stat-label">总测试数</div>
        </div>
        <div class="stat-item success">
          <div class="stat-value">{{ stats.success }}</div>
          <div class="stat-label">成功</div>
        </div>
        <div class="stat-item error">
          <div class="stat-value">{{ stats.error }}</div>
          <div class="stat-label">失败</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ stats.successRate.toFixed(1) }}%</div>
          <div class="stat-label">成功率</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ stats.avgResponseTime }}ms</div>
          <div class="stat-label">平均响应时间</div>
        </div>
      </div>
    </el-card>

    <!-- 测试套件 -->
    <div class="test-suites">
      <el-card 
        v-for="(suite, index) in testSuites" 
        :key="index"
        class="suite-card"
      >
        <template #header>
          <div class="suite-header">
            <span>{{ suite.name }}</span>
            <el-button 
              size="small" 
              @click="runTestSuite(suite)"
              :loading="runningSuite === suite.name"
              :disabled="isRunning"
            >
              运行测试套件
            </el-button>
          </div>
        </template>

        <div class="test-cases">
          <div 
            v-for="(testCase, testIndex) in suite.tests" 
            :key="testIndex"
            class="test-case"
          >
            <div class="test-case-header">
              <div class="test-info">
                <span class="test-name">{{ testCase.name }}</span>
                <span class="test-endpoint">{{ testCase.method }} {{ testCase.endpoint }}</span>
              </div>
              <div class="test-actions">
                <el-button 
                  size="small" 
                  @click="runSingleTest(testCase)"
                  :loading="runningTest === testCase.name"
                  :disabled="isRunning"
                >
                  运行
                </el-button>
              </div>
            </div>

            <!-- 测试结果 -->
            <div 
              v-if="getTestResult(testCase.name)"
              class="test-result"
              :class="getTestResult(testCase.name)?.status"
            >
              <div class="result-header">
                <span class="result-status">
                  <el-icon v-if="getTestResult(testCase.name)?.status === 'success'" color="#52c41a">
                    <Check />
                  </el-icon>
                  <el-icon v-else color="#ff4d4f">
                    <Close />
                  </el-icon>
                  {{ getTestResult(testCase.name)?.status === 'success' ? '成功' : '失败' }}
                </span>
                <span class="result-time">
                  {{ getTestResult(testCase.name)?.responseTime.toFixed(2) }}ms
                </span>
                <span class="result-status-code" v-if="getTestResult(testCase.name)?.statusCode">
                  {{ getTestResult(testCase.name)?.statusCode }}
                </span>
              </div>
              
              <div 
                v-if="getTestResult(testCase.name)?.error" 
                class="result-error"
              >
                {{ getTestResult(testCase.name)?.error }}
              </div>
              
              <div 
                v-if="getTestResult(testCase.name)?.data && showDetails"
                class="result-data"
              >
                <pre>{{ JSON.stringify(getTestResult(testCase.name)?.data, null, 2) }}</pre>
              </div>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 详细结果 -->
    <el-card class="results-card" v-if="results.length > 0">
      <template #header>
        <div class="results-header">
          <span>详细结果</span>
          <el-switch 
            v-model="showDetails" 
            active-text="显示响应数据"
            inactive-text="隐藏响应数据"
          />
        </div>
      </template>

      <div class="results-list">
        <div 
          v-for="(result, index) in results" 
          :key="index"
          class="result-item"
          :class="result.status"
        >
          <div class="result-summary">
            <div class="result-info">
              <span class="result-name">{{ result.name }}</span>
              <span class="result-endpoint">{{ result.method }} {{ result.endpoint }}</span>
            </div>
            <div class="result-metrics">
              <span class="result-status-badge" :class="result.status">
                {{ result.status === 'success' ? '成功' : '失败' }}
              </span>
              <span class="result-time">{{ result.responseTime.toFixed(2) }}ms</span>
              <span class="result-code" v-if="result.statusCode">{{ result.statusCode }}</span>
            </div>
          </div>
          
          <div v-if="result.error" class="result-error">
            {{ result.error }}
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElCard, ElButton, ElIcon, ElSwitch, ElMessage } from 'element-plus'
import { Check, Close } from '@element-plus/icons-vue'
import { apiTester, testSuites } from '@/utils/apiTester'

// 响应式数据
const isRunning = ref(false)
const runningSuite = ref<string | null>(null)
const runningTest = ref<string | null>(null)
const results = ref<any[]>([])
const showDetails = ref(false)

// 计算属性
const stats = computed(() => apiTester.getStats())

// 方法
const runAllTests = async () => {
  isRunning.value = true
  results.value = []
  
  try {
    for (const suite of testSuites) {
      runningSuite.value = suite.name
      await apiTester.runTestSuite(suite)
    }
    
    results.value = apiTester.getResults()
    ElMessage.success('所有测试完成')
  } catch (error) {
    ElMessage.error('测试过程中出现错误')
    console.error('Test error:', error)
  } finally {
    isRunning.value = false
    runningSuite.value = null
  }
}

const runTestSuite = async (suite: any) => {
  isRunning.value = true
  runningSuite.value = suite.name
  
  try {
    await apiTester.runTestSuite(suite)
    results.value = apiTester.getResults()
    ElMessage.success(`测试套件 "${suite.name}" 完成`)
  } catch (error) {
    ElMessage.error('测试套件运行失败')
    console.error('Test suite error:', error)
  } finally {
    isRunning.value = false
    runningSuite.value = null
  }
}

const runSingleTest = async (testCase: any) => {
  runningTest.value = testCase.name
  
  try {
    await apiTester.runTest(testCase)
    results.value = apiTester.getResults()
    ElMessage.success(`测试 "${testCase.name}" 完成`)
  } catch (error) {
    ElMessage.error('单个测试运行失败')
    console.error('Single test error:', error)
  } finally {
    runningTest.value = null
  }
}

const getTestResult = (testName: string) => {
  return results.value.find(r => r.name === testName)
}

const clearResults = () => {
  apiTester.clearResults()
  results.value = []
  ElMessage.info('测试结果已清空')
}

const exportReport = () => {
  const report = apiTester.exportReport()
  const blob = new Blob([report], { type: 'text/markdown' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `api-test-report-${Date.now()}.md`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('测试报告已导出')
}

// 生命周期
onMounted(() => {
  results.value = apiTester.getResults()
})
</script>

<style scoped>
.api-test {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.test-header {
  margin-bottom: 24px;
  text-align: center;
}

.test-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 600;
}

.test-header p {
  margin: 0;
  color: #666;
  font-size: 16px;
}

.test-controls {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  justify-content: center;
}

.stats-card {
  margin-bottom: 24px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 16px;
}

.stat-item {
  text-align: center;
  padding: 16px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
}

.stat-item.success {
  background: #f6ffed;
  border-color: #b7eb8f;
}

.stat-item.error {
  background: #fff2f0;
  border-color: #ffccc7;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

.test-suites {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-bottom: 24px;
}

.suite-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.test-cases {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.test-case {
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  overflow: hidden;
}

.test-case-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #fafafa;
}

.test-info {
  flex: 1;
}

.test-name {
  display: block;
  font-weight: 500;
  margin-bottom: 4px;
}

.test-endpoint {
  font-size: 12px;
  color: #666;
  font-family: monospace;
}

.test-result {
  padding: 12px 16px;
  border-top: 1px solid #f0f0f0;
}

.test-result.success {
  background: #f6ffed;
}

.test-result.error {
  background: #fff2f0;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.result-status {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
}

.result-time {
  font-size: 12px;
  color: #666;
}

.result-status-code {
  font-size: 12px;
  padding: 2px 6px;
  background: #f0f0f0;
  border-radius: 3px;
  font-family: monospace;
}

.result-error {
  color: #ff4d4f;
  font-size: 12px;
  margin-bottom: 8px;
}

.result-data {
  font-size: 11px;
  background: #f8f8f8;
  padding: 8px;
  border-radius: 4px;
  overflow-x: auto;
}

.result-data pre {
  margin: 0;
  white-space: pre-wrap;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.results-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.result-item {
  padding: 12px;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
}

.result-item.success {
  border-color: #b7eb8f;
  background: #f6ffed;
}

.result-item.error {
  border-color: #ffccc7;
  background: #fff2f0;
}

.result-summary {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.result-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.result-endpoint {
  font-size: 12px;
  color: #666;
  font-family: monospace;
}

.result-metrics {
  display: flex;
  align-items: center;
  gap: 8px;
}

.result-status-badge {
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.result-status-badge.success {
  background: #52c41a;
  color: white;
}

.result-status-badge.error {
  background: #ff4d4f;
  color: white;
}

@media (max-width: 768px) {
  .api-test {
    padding: 16px;
  }
  
  .test-controls {
    flex-direction: column;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .test-case-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .result-summary {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}
</style>
