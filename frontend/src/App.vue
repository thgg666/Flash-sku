<script setup lang="ts">
import { onMounted } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import PWAInstallPrompt from '@/components/common/PWAInstallPrompt.vue'
import OfflineIndicator from '@/components/common/OfflineIndicator.vue'
import MemoryMonitor from '@/components/common/MemoryMonitor.vue'
import { useAuth } from '@/composables/useAuth'
import { pwaManager } from '@/utils/pwa'
import { offlineStorage } from '@/utils/offlineStorage'

// 环境检测
const isDevelopment = import.meta.env.DEV

// 认证相关
const { guideNewUser } = useAuth()

// 组件挂载时检查新用户和初始化PWA
onMounted(() => {
  // 延迟执行，确保页面加载完成
  setTimeout(() => {
    guideNewUser()
  }, 1000)

  // 初始化PWA功能
  initPWA()
})

// 初始化PWA功能
const initPWA = () => {
  // 监听PWA事件
  pwaManager.on('updateAvailable', () => {
    console.log('PWA update available')
    // 可以在这里显示更新提示
  })

  pwaManager.on('appInstalled', () => {
    console.log('PWA installed successfully')
  })

  pwaManager.on('networkStatusChanged', (status: { online: boolean }) => {
    console.log('Network status changed:', status.online ? 'online' : 'offline')
  })
}
</script>

<template>
  <div id="app">
    <router-view />
  </div>
</template>

<style>
/* 全局样式重置 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  margin: 0;
  padding: 0;
  width: 100%;
  height: 100%;
  overflow-x: hidden;
}

#app {
  margin: 0;
  padding: 0;
  min-height: 100vh;
  width: 100%;
  position: relative;
}
</style>
