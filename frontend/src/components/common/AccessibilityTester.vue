<template>
  <div class="accessibility-tester">
    <el-card class="tester-card">
      <template #header>
        <div class="card-header">
          <h3>可访问性测试工具</h3>
          <el-button @click="runAllTests" :loading="testing" type="primary" size="small">
            运行所有测试
          </el-button>
        </div>
      </template>

      <!-- 键盘导航测试 -->
      <div class="test-section">
        <h4>键盘导航测试</h4>
        <p>使用Tab键在以下元素间导航，测试焦点顺序是否正确：</p>
        <div class="keyboard-test-area">
          <el-button>按钮 1</el-button>
          <el-input v-model="testInput" placeholder="输入框" style="width: 200px; margin: 0 8px;" />
          <el-select v-model="testSelect" placeholder="选择框" style="width: 150px; margin: 0 8px;">
            <el-option label="选项1" value="1" />
            <el-option label="选项2" value="2" />
          </el-select>
          <el-button>按钮 2</el-button>
        </div>
        <div class="test-result">
          <el-button @click="testKeyboardNavigation" type="success" size="small">
            测试键盘导航
          </el-button>
          <span v-if="keyboardResult" :class="keyboardResult.success ? 'success' : 'error'">
            {{ keyboardResult.message }}
          </span>
        </div>
      </div>

      <!-- 屏幕阅读器测试 -->
      <div class="test-section">
        <h4>屏幕阅读器测试</h4>
        <p>测试ARIA标签和屏幕阅读器公告功能：</p>
        <div class="screen-reader-test-area">
          <button 
            aria-label="这是一个带有ARIA标签的按钮"
            @click="announceMessage"
          >
            点击测试屏幕阅读器公告
          </button>
          <div 
            role="status" 
            aria-live="polite" 
            aria-atomic="true"
            class="sr-announcement"
          >
            {{ announcement }}
          </div>
        </div>
      </div>

      <!-- 颜色对比度测试 -->
      <div class="test-section">
        <h4>颜色对比度测试</h4>
        <p>测试不同颜色组合的对比度是否符合WCAG标准：</p>
        <div class="contrast-test-area">
          <div class="contrast-sample" style="background: #ffffff; color: #000000;">
            黑色文字 / 白色背景 (对比度: 21:1)
          </div>
          <div class="contrast-sample" style="background: #007bff; color: #ffffff;">
            白色文字 / 蓝色背景 (对比度: 5.9:1)
          </div>
          <div class="contrast-sample" style="background: #28a745; color: #ffffff;">
            白色文字 / 绿色背景 (对比度: 4.1:1)
          </div>
          <div class="contrast-sample warning" style="background: #ffc107; color: #212529;">
            深色文字 / 黄色背景 (对比度: 2.8:1) ⚠️
          </div>
        </div>
      </div>

      <!-- 焦点管理测试 -->
      <div class="test-section">
        <h4>焦点管理测试</h4>
        <p>测试模态框和焦点陷阱功能：</p>
        <div class="focus-test-area">
          <el-button @click="showModal = true" type="primary">
            打开模态框测试焦点陷阱
          </el-button>
        </div>
      </div>

      <!-- 快捷键测试 -->
      <div class="test-section">
        <h4>快捷键测试</h4>
        <p>测试以下快捷键是否正常工作：</p>
        <div class="shortcut-test-area">
          <ul>
            <li><kbd>Alt + M</kbd> - 跳转到主菜单</li>
            <li><kbd>Alt + S</kbd> - 跳转到搜索</li>
            <li><kbd>Alt + C</kbd> - 跳转到主内容</li>
            <li><kbd>Esc</kbd> - 关闭模态框</li>
          </ul>
          <div class="shortcut-status">
            <span>按下快捷键测试: {{ lastShortcut || '无' }}</span>
          </div>
        </div>
      </div>

      <!-- 测试结果汇总 -->
      <div class="test-summary">
        <h4>测试结果汇总</h4>
        <div class="summary-grid">
          <div class="summary-item">
            <span class="label">键盘导航:</span>
            <span :class="getStatusClass(testResults.keyboard)">
              {{ getStatusText(testResults.keyboard) }}
            </span>
          </div>
          <div class="summary-item">
            <span class="label">屏幕阅读器:</span>
            <span :class="getStatusClass(testResults.screenReader)">
              {{ getStatusText(testResults.screenReader) }}
            </span>
          </div>
          <div class="summary-item">
            <span class="label">颜色对比度:</span>
            <span :class="getStatusClass(testResults.contrast)">
              {{ getStatusText(testResults.contrast) }}
            </span>
          </div>
          <div class="summary-item">
            <span class="label">焦点管理:</span>
            <span :class="getStatusClass(testResults.focus)">
              {{ getStatusText(testResults.focus) }}
            </span>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 模态框 -->
    <el-dialog
      v-model="showModal"
      title="焦点陷阱测试"
      width="400px"
      :before-close="handleModalClose"
    >
      <p>这是一个测试焦点陷阱的模态框。</p>
      <p>使用Tab键在以下元素间导航，焦点应该被限制在模态框内：</p>
      <div class="modal-test-elements">
        <el-input v-model="modalInput" placeholder="模态框输入框" />
        <el-button style="margin: 8px 0;">模态框按钮 1</el-button>
        <el-button>模态框按钮 2</el-button>
      </div>
      <template #footer>
        <el-button @click="showModal = false">取消</el-button>
        <el-button type="primary" @click="showModal = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  ScreenReaderAnnouncer, 
  KeyboardNavigationManager,
  FocusTrap,
  calculateContrastRatio 
} from '@/utils/accessibility'

// 状态
const testing = ref(false)
const showModal = ref(false)
const testInput = ref('')
const testSelect = ref('')
const modalInput = ref('')
const announcement = ref('')
const lastShortcut = ref('')

// 测试结果
const keyboardResult = ref<{success: boolean, message: string} | null>(null)
const testResults = reactive({
  keyboard: null as boolean | null,
  screenReader: null as boolean | null,
  contrast: null as boolean | null,
  focus: null as boolean | null
})

// 工具实例
let screenReaderAnnouncer: ScreenReaderAnnouncer
let keyboardManager: KeyboardNavigationManager
let focusTrap: FocusTrap | null = null

// 初始化
onMounted(() => {
  screenReaderAnnouncer = new ScreenReaderAnnouncer()
  keyboardManager = new KeyboardNavigationManager()
  
  // 监听快捷键
  document.addEventListener('keydown', handleShortcut)
})

onUnmounted(() => {
  screenReaderAnnouncer?.destroy()
  document.removeEventListener('keydown', handleShortcut)
})

// 运行所有测试
const runAllTests = async () => {
  testing.value = true
  try {
    await testKeyboardNavigation()
    await testScreenReader()
    await testColorContrast()
    await testFocusManagement()
    
    ElMessage.success('所有可访问性测试完成')
  } catch (error) {
    ElMessage.error('测试过程中出现错误')
  } finally {
    testing.value = false
  }
}

// 测试键盘导航
const testKeyboardNavigation = async () => {
  try {
    keyboardManager.updateFocusableElements()
    const focusableCount = keyboardManager['focusableElements'].length
    
    if (focusableCount > 0) {
      keyboardResult.value = { 
        success: true, 
        message: `找到 ${focusableCount} 个可聚焦元素` 
      }
      testResults.keyboard = true
    } else {
      keyboardResult.value = { 
        success: false, 
        message: '未找到可聚焦元素' 
      }
      testResults.keyboard = false
    }
  } catch (error) {
    keyboardResult.value = { 
      success: false, 
      message: '键盘导航测试失败' 
    }
    testResults.keyboard = false
  }
}

// 测试屏幕阅读器
const testScreenReader = async () => {
  try {
    screenReaderAnnouncer.announce('屏幕阅读器测试消息')
    testResults.screenReader = true
  } catch (error) {
    testResults.screenReader = false
  }
}

// 测试颜色对比度
const testColorContrast = async () => {
  try {
    const samples = [
      { bg: 'rgb(255, 255, 255)', fg: 'rgb(0, 0, 0)', min: 7 },
      { bg: 'rgb(0, 123, 255)', fg: 'rgb(255, 255, 255)', min: 4.5 },
      { bg: 'rgb(40, 167, 69)', fg: 'rgb(255, 255, 255)', min: 4.5 },
      { bg: 'rgb(255, 193, 7)', fg: 'rgb(33, 37, 41)', min: 4.5 }
    ]
    
    let allPassed = true
    samples.forEach(sample => {
      const ratio = calculateContrastRatio(sample.fg, sample.bg)
      if (ratio < sample.min) {
        allPassed = false
      }
    })
    
    testResults.contrast = allPassed
  } catch (error) {
    testResults.contrast = false
  }
}

// 测试焦点管理
const testFocusManagement = async () => {
  try {
    // 简单测试：检查是否能创建焦点陷阱
    const testElement = document.createElement('div')
    const trap = new FocusTrap(testElement)
    trap.activate()
    trap.deactivate()
    
    testResults.focus = true
  } catch (error) {
    testResults.focus = false
  }
}

// 屏幕阅读器公告
const announceMessage = () => {
  const message = '这是一个屏幕阅读器测试消息'
  announcement.value = message
  screenReaderAnnouncer.announce(message)
  
  setTimeout(() => {
    announcement.value = ''
  }, 3000)
}

// 处理快捷键
const handleShortcut = (event: KeyboardEvent) => {
  const combination = []
  if (event.ctrlKey) combination.push('Ctrl')
  if (event.altKey) combination.push('Alt')
  if (event.shiftKey) combination.push('Shift')
  combination.push(event.key)
  
  lastShortcut.value = combination.join(' + ')
}

// 处理模态框关闭
const handleModalClose = () => {
  if (focusTrap) {
    focusTrap.deactivate()
    focusTrap = null
  }
  showModal.value = false
}

// 获取状态类名
const getStatusClass = (status: boolean | null) => {
  if (status === null) return 'pending'
  return status ? 'success' : 'error'
}

// 获取状态文本
const getStatusText = (status: boolean | null) => {
  if (status === null) return '未测试'
  return status ? '通过' : '失败'
}
</script>

<style scoped lang="scss">
@import "@/styles/variables.scss";

.accessibility-tester {
  .tester-card {
    max-width: 800px;
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
  }

  .test-section {
    margin-bottom: 32px;
    padding-bottom: 24px;
    border-bottom: 1px solid $border-color-lighter;

    &:last-child {
      border-bottom: none;
    }

    h4 {
      color: $text-color-primary;
      margin-bottom: 8px;
    }

    p {
      color: $text-color-regular;
      margin-bottom: 16px;
    }
  }

  .keyboard-test-area {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
    flex-wrap: wrap;
  }

  .screen-reader-test-area {
    .sr-announcement {
      margin-top: 16px;
      padding: 8px 12px;
      background: $bg-color-page;
      border-radius: 4px;
      min-height: 40px;
      color: $text-color-regular;
    }
  }

  .contrast-test-area {
    .contrast-sample {
      padding: 12px 16px;
      margin-bottom: 8px;
      border-radius: 4px;
      font-weight: 500;

      &.warning {
        position: relative;
        
        &::after {
          content: ' (不符合WCAG AA标准)';
          font-size: 12px;
          opacity: 0.8;
        }
      }
    }
  }

  .shortcut-test-area {
    ul {
      margin-bottom: 16px;
      
      li {
        margin-bottom: 8px;
        
        kbd {
          background: $bg-color-page;
          padding: 2px 6px;
          border-radius: 3px;
          font-family: monospace;
          font-size: 12px;
          border: 1px solid $border-color;
        }
      }
    }

    .shortcut-status {
      padding: 8px 12px;
      background: $bg-color-page;
      border-radius: 4px;
      font-family: monospace;
    }
  }

  .test-result {
    margin-top: 16px;
    
    .success {
      color: $success-color;
      margin-left: 12px;
    }
    
    .error {
      color: $danger-color;
      margin-left: 12px;
    }
  }

  .test-summary {
    margin-top: 32px;
    padding-top: 24px;
    border-top: 2px solid $border-color;

    h4 {
      margin-bottom: 16px;
      color: $text-color-primary;
    }

    .summary-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 16px;

      .summary-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px 16px;
        background: $bg-color-page;
        border-radius: 6px;

        .label {
          font-weight: 500;
          color: $text-color-regular;
        }

        .success {
          color: $success-color;
          font-weight: 500;
        }

        .error {
          color: $danger-color;
          font-weight: 500;
        }

        .pending {
          color: $text-color-placeholder;
          font-weight: 500;
        }
      }
    }
  }

  .modal-test-elements {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin: 16px 0;
  }
}
</style>
