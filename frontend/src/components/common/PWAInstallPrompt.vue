<template>
  <Teleport to="body">
    <Transition name="slide-up">
      <div 
        v-if="showPrompt" 
        class="pwa-install-prompt"
        :class="{ 'mobile': isMobile }"
      >
        <div class="prompt-content">
          <div class="prompt-icon">
            <el-icon :size="24">
              <Download />
            </el-icon>
          </div>
          
          <div class="prompt-text">
            <h3 class="prompt-title">安装 Flash Sku</h3>
            <p class="prompt-description">
              {{ isMobile ? '添加到主屏幕，享受更好的购物体验' : '安装应用到桌面，快速访问秒杀活动' }}
            </p>
          </div>
          
          <div class="prompt-actions">
            <el-button 
              type="primary" 
              size="small"
              @click="handleInstall"
              :loading="installing"
            >
              {{ installing ? '安装中...' : '安装' }}
            </el-button>
            
            <el-button 
              size="small" 
              @click="handleDismiss"
            >
              稍后
            </el-button>
            
            <el-button 
              size="small" 
              text 
              @click="handleNeverShow"
            >
              不再提示
            </el-button>
          </div>
          
          <button 
            class="prompt-close" 
            @click="handleDismiss"
            aria-label="关闭"
          >
            <el-icon :size="16">
              <Close />
            </el-icon>
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElButton, ElIcon, ElMessage } from 'element-plus'
import { Download, Close } from '@element-plus/icons-vue'

// 响应式数据
const showPrompt = ref(false)
const installing = ref(false)
const isMobile = ref(false)
const deferredPrompt = ref<any>(null)

// 检测设备类型
const checkDevice = () => {
  isMobile.value = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
}

// 检查PWA安装条件
const checkInstallability = () => {
  // 检查是否已经安装
  if (window.matchMedia && window.matchMedia('(display-mode: standalone)').matches) {
    return false
  }
  
  // 检查是否在支持的浏览器中
  if (!('serviceWorker' in navigator)) {
    return false
  }
  
  // 检查用户是否已经拒绝过
  const neverShow = localStorage.getItem('pwa-install-never-show')
  if (neverShow === 'true') {
    return false
  }
  
  // 检查是否在合适的时间显示（用户访问过几次后）
  const visitCount = parseInt(localStorage.getItem('visit-count') || '0')
  if (visitCount < 3) {
    localStorage.setItem('visit-count', (visitCount + 1).toString())
    return false
  }
  
  return true
}

// 处理安装事件
const handleInstall = async () => {
  if (!deferredPrompt.value) {
    // 如果没有原生安装提示，显示手动安装指导
    showManualInstallGuide()
    return
  }
  
  installing.value = true
  
  try {
    // 显示安装提示
    deferredPrompt.value.prompt()
    
    // 等待用户响应
    const { outcome } = await deferredPrompt.value.userChoice
    
    if (outcome === 'accepted') {
      ElMessage.success('应用安装成功！')
      showPrompt.value = false
      
      // 记录安装成功
      localStorage.setItem('pwa-installed', 'true')
    } else {
      ElMessage.info('安装已取消')
    }
    
    // 清除延迟的提示
    deferredPrompt.value = null
  } catch (error) {
    console.error('PWA installation failed:', error)
    ElMessage.error('安装失败，请稍后重试')
  } finally {
    installing.value = false
  }
}

// 显示手动安装指导
const showManualInstallGuide = () => {
  const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent)
  const isAndroid = /Android/.test(navigator.userAgent)
  
  let message = ''
  
  if (isIOS) {
    message = '在 Safari 中点击分享按钮，然后选择"添加到主屏幕"'
  } else if (isAndroid) {
    message = '在浏览器菜单中选择"添加到主屏幕"或"安装应用"'
  } else {
    message = '在浏览器地址栏中点击安装图标，或在菜单中选择"安装应用"'
  }
  
  ElMessage({
    message,
    type: 'info',
    duration: 5000,
    showClose: true
  })
}

// 处理稍后提示
const handleDismiss = () => {
  showPrompt.value = false
  
  // 设置下次显示时间（24小时后）
  const nextShow = Date.now() + 24 * 60 * 60 * 1000
  localStorage.setItem('pwa-install-next-show', nextShow.toString())
}

// 处理不再提示
const handleNeverShow = () => {
  showPrompt.value = false
  localStorage.setItem('pwa-install-never-show', 'true')
  ElMessage.info('已设置不再提示安装')
}

// 监听beforeinstallprompt事件
const handleBeforeInstallPrompt = (e: Event) => {
  // 阻止默认的安装提示
  e.preventDefault()
  
  // 保存事件以便稍后使用
  deferredPrompt.value = e
  
  // 检查是否应该显示自定义提示
  if (checkInstallability()) {
    // 延迟显示，让用户先体验应用
    setTimeout(() => {
      showPrompt.value = true
    }, 3000)
  }
}

// 监听应用安装事件
const handleAppInstalled = () => {
  console.log('PWA was installed')
  showPrompt.value = false
  localStorage.setItem('pwa-installed', 'true')
  ElMessage.success('应用已成功安装到设备！')
}

// 检查是否应该显示提示
const checkShouldShow = () => {
  if (!checkInstallability()) {
    return
  }
  
  // 检查是否到了下次显示时间
  const nextShow = localStorage.getItem('pwa-install-next-show')
  if (nextShow && Date.now() < parseInt(nextShow)) {
    return
  }
  
  // 对于移动设备，在用户滚动到页面底部时显示
  if (isMobile.value) {
    const handleScroll = () => {
      const scrollTop = window.pageYOffset || document.documentElement.scrollTop
      const windowHeight = window.innerHeight
      const documentHeight = document.documentElement.scrollHeight
      
      if (scrollTop + windowHeight >= documentHeight - 100) {
        showPrompt.value = true
        window.removeEventListener('scroll', handleScroll)
      }
    }
    
    window.addEventListener('scroll', handleScroll)
  } else {
    // 桌面设备延迟显示
    setTimeout(() => {
      showPrompt.value = true
    }, 5000)
  }
}

// 生命周期
onMounted(() => {
  checkDevice()
  
  // 添加事件监听器
  window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
  window.addEventListener('appinstalled', handleAppInstalled)
  
  // 检查是否应该显示提示
  checkShouldShow()
})

onUnmounted(() => {
  // 移除事件监听器
  window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
  window.removeEventListener('appinstalled', handleAppInstalled)
})
</script>

<style scoped>
.pwa-install-prompt {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  max-width: 400px;
  width: calc(100% - 40px);
}

.pwa-install-prompt.mobile {
  bottom: 0;
  left: 0;
  transform: none;
  max-width: none;
  width: 100%;
  border-radius: 16px 16px 0 0;
}

.prompt-content {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  border: 1px solid rgba(0, 0, 0, 0.08);
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.mobile .prompt-content {
  border-radius: 16px 16px 0 0;
  padding: 24px 20px;
}

.prompt-icon {
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.prompt-text {
  flex: 1;
  min-width: 0;
}

.prompt-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 4px 0;
}

.prompt-description {
  font-size: 14px;
  color: #6b7280;
  margin: 0 0 16px 0;
  line-height: 1.4;
}

.prompt-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.mobile .prompt-actions {
  width: 100%;
  justify-content: space-between;
}

.prompt-close {
  position: absolute;
  top: 12px;
  right: 12px;
  background: none;
  border: none;
  cursor: pointer;
  color: #9ca3af;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.prompt-close:hover {
  background: #f3f4f6;
  color: #6b7280;
}

/* 动画 */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from {
  transform: translateX(-50%) translateY(100%);
  opacity: 0;
}

.slide-up-leave-to {
  transform: translateX(-50%) translateY(100%);
  opacity: 0;
}

.mobile .slide-up-enter-from,
.mobile .slide-up-leave-to {
  transform: translateY(100%);
}

/* 响应式 */
@media (max-width: 480px) {
  .prompt-content {
    flex-direction: column;
    text-align: center;
  }
  
  .prompt-actions {
    justify-content: center;
  }
  
  .prompt-close {
    top: 16px;
    right: 16px;
  }
}
</style>
