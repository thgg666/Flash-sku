<template>
  <div class="app-layout" :class="{ 'keyboard-navigation': isKeyboardNavigation }">
    <!-- 滚动进度指示器 -->
    <div v-if="isDesktop" class="scroll-indicator" :style="{ width: scrollProgress + '%' }"></div>

    <!-- 头部导航 -->
    <AppHeader />

    <!-- 主要内容区域 -->
    <main id="main-content" class="main-content smooth-scroll" tabindex="-1">
      <router-view />
    </main>

    <!-- 返回顶部按钮 -->
    <button
      v-if="isDesktop && showBackToTop"
      class="back-to-top desktop-hover-lift"
      @click="scrollToTop"
      :title="'返回顶部 (快捷键: Ctrl+Home)'"
    >
      <el-icon><ArrowUp /></el-icon>
    </button>

    <!-- 右键菜单 -->
    <div
      v-if="isContextMenuVisible"
      class="context-menu"
      :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
    >
      <div
        v-for="(item, index) in contextMenuItems"
        :key="index"
        class="menu-item"
        :class="{ disabled: item.disabled, divider: item.divider }"
        @click="!item.disabled && executeContextAction(item.action)"
      >
        <el-icon v-if="item.icon" class="menu-icon">
          <component :is="item.icon" />
        </el-icon>
        {{ item.label }}
      </div>
    </div>

    <!-- 页脚 -->
    <footer class="app-footer">
      <div class="container">
        <div class="footer-content">
          <div class="footer-info">
            <h3>Flash Sku</h3>
            <p>高性能秒杀系统</p>
          </div>
          <div class="footer-links">
            <div class="link-group">
              <h4>产品</h4>
              <ul>
                <li><a href="#" class="focusable">秒杀活动</a></li>
                <li><a href="#" class="focusable">商品管理</a></li>
                <li><a href="#" class="focusable">订单管理</a></li>
              </ul>
            </div>
            <div class="link-group">
              <h4>帮助</h4>
              <ul>
                <li><a href="#" class="focusable">使用指南</a></li>
                <li><a href="#" class="focusable">常见问题</a></li>
                <li><a href="#" class="focusable">联系客服</a></li>
              </ul>
            </div>
            <div class="link-group">
              <h4>关于</h4>
              <ul>
                <li><a href="#" class="focusable">关于我们</a></li>
                <li><a href="#" class="focusable">隐私政策</a></li>
                <li><a href="#" class="focusable">服务条款</a></li>
              </ul>
            </div>
          </div>
        </div>
        <div class="footer-bottom">
          <p>&copy; 2025 Flash Sku. All rights reserved.</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ArrowUp, Refresh, DocumentCopy, Share } from '@element-plus/icons-vue'
import AppHeader from './AppHeader.vue'
import { useDesktopInteraction, useScrollEnhancement, useContextMenu, useKeyboardShortcuts } from '@/composables/useDesktopInteraction'

// 桌面端交互增强
const { isDesktop, isKeyboardNavigation } = useDesktopInteraction()

// 滚动增强
const { scrollToTop } = useScrollEnhancement()

// 右键菜单
const contextMenu = useContextMenu()
const isContextMenuVisible = contextMenu.isVisible
const contextMenuPosition = contextMenu.position
const contextMenuItems = contextMenu.menuItems
const executeContextAction = contextMenu.executeAction

// 键盘快捷键
const { registerShortcut } = useKeyboardShortcuts()

// 滚动相关状态
const scrollProgress = ref(0)
const showBackToTop = ref(false)

// 计算滚动进度
const updateScrollProgress = () => {
  const scrollTop = window.pageYOffset || document.documentElement.scrollTop
  const scrollHeight = document.documentElement.scrollHeight - window.innerHeight
  scrollProgress.value = (scrollTop / scrollHeight) * 100
  showBackToTop.value = scrollTop > 300
}

// 处理右键菜单
const handleContextMenu = (event: MouseEvent) => {
  if (!isDesktop.value) return

  const menuItems = [
    {
      label: '刷新页面',
      icon: Refresh,
      action: () => window.location.reload()
    },
    {
      label: '复制链接',
      icon: DocumentCopy,
      action: () => navigator.clipboard.writeText(window.location.href)
    },
    {
      label: '分享页面',
      icon: Share,
      action: () => {
        if (navigator.share) {
          navigator.share({
            title: document.title,
            url: window.location.href
          })
        }
      },
      disabled: !navigator.share
    }
  ]

  contextMenu.showMenu(event, menuItems)
}

onMounted(() => {
  // 监听滚动事件
  window.addEventListener('scroll', updateScrollProgress, { passive: true })

  // 监听右键菜单
  document.addEventListener('contextmenu', handleContextMenu)

  // 注册键盘快捷键
  registerShortcut('ctrl+home', scrollToTop)
  registerShortcut('home', scrollToTop)
  registerShortcut('f5', () => window.location.reload())

  // 初始化滚动进度
  updateScrollProgress()
})

onUnmounted(() => {
  window.removeEventListener('scroll', updateScrollProgress)
  document.removeEventListener('contextmenu', handleContextMenu)
})
</script>

<style scoped lang="scss">
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;

  .main-content {
    flex: 1;
    min-height: calc(100vh - 64px - 200px); // 减去头部和页脚高度
  }

  .app-footer {
    background: var(--el-bg-color-page);
    border-top: 1px solid var(--el-border-color-lighter);
    padding: 40px 0 20px;
    margin-top: auto;

    .container {
      max-width: 1200px;
      margin: 0 auto;
      padding: 0 24px;
    }

    .footer-content {
      display: grid;
      grid-template-columns: 1fr 2fr;
      gap: 40px;
      margin-bottom: 30px;

      .footer-info {
        h3 {
          margin: 0 0 12px 0;
          color: var(--el-text-color-primary);
          font-size: 24px;
          font-weight: 600;
          background: linear-gradient(135deg, var(--el-color-primary), #667eea);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
        }

        p {
          margin: 0;
          color: var(--el-text-color-regular);
          font-size: 14px;
        }
      }

      .footer-links {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 30px;

        .link-group {
          h4 {
            margin: 0 0 16px 0;
            color: var(--el-text-color-primary);
            font-size: 16px;
            font-weight: 600;
          }

          ul {
            list-style: none;
            padding: 0;
            margin: 0;

            li {
              margin-bottom: 8px;

              a {
                color: var(--el-text-color-regular);
                text-decoration: none;
                font-size: 14px;
                transition: color 0.3s ease;

                &:hover {
                  color: var(--el-color-primary);
                }
              }
            }
          }
        }
      }
    }

    .footer-bottom {
      padding-top: 20px;
      border-top: 1px solid var(--el-border-color-lighter);
      text-align: center;

      p {
        margin: 0;
        color: var(--el-text-color-placeholder);
        font-size: 12px;
      }
    }
  }
}

@media (max-width: 768px) {
  .app-layout {
    .app-footer {
      padding: 30px 0 15px;

      .container {
        padding: 0 16px;
      }

      .footer-content {
        grid-template-columns: 1fr;
        gap: 30px;

        .footer-links {
          grid-template-columns: 1fr;
          gap: 20px;
        }
      }
    }
  }
}
</style>
