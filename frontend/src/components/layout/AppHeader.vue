<template>
  <header class="app-header">
    <div class="container">
      <div class="header-content">
        <!-- Logo和标题 -->
        <div class="logo-section">
          <router-link to="/" class="logo-link">
            <el-icon class="logo-icon"><ShoppingBag /></el-icon>
            <span class="logo-text">Flash Sku</span>
          </router-link>
        </div>

        <!-- 导航菜单 -->
        <nav class="nav-menu desktop-only" role="navigation" aria-label="主导航">
          <el-menu
            mode="horizontal"
            :default-active="activeMenu"
            class="header-menu"
            @select="handleMenuSelect"
          >
            <el-menu-item index="/">
              <el-icon><House /></el-icon>
              <span>首页</span>
            </el-menu-item>
            <el-menu-item index="/activities">
              <el-icon><Timer /></el-icon>
              <span>秒杀活动</span>
            </el-menu-item>
          </el-menu>
        </nav>

        <!-- 移动端菜单按钮 -->
        <div class="mobile-menu-btn mobile-only">
          <el-button circle @click="toggleMobileMenu">
            <el-icon><Menu /></el-icon>
          </el-button>
        </div>

        <!-- 用户操作区 -->
        <div class="user-actions">
          <!-- WebSocket状态指示器 -->
          <div v-if="isAuthenticated" class="websocket-status">
            <WebSocketStatus :show-text="false" />
          </div>

          <!-- 用户反馈通知 -->
          <div v-if="isAuthenticated" class="feedback-notification">
            <FeedbackNotification />
          </div>

          <!-- 未登录状态 -->
          <div v-if="!isAuthenticated" class="auth-buttons">
            <el-button @click="goToLogin" class="desktop-hover-lift">登录</el-button>
            <el-button type="primary" @click="goToRegister" class="desktop-hover-lift">注册</el-button>
          </div>

          <!-- 已登录状态 -->
          <div v-else class="user-info">
            <!-- 邮箱验证提醒 -->
            <el-tooltip
              v-if="!isEmailVerified"
              content="您的邮箱尚未验证，点击去验证"
              placement="bottom"
            >
              <el-button
                type="warning"
                size="small"
                circle
                @click="goToVerifyEmail"
                class="desktop-hover-scale"
              >
                <el-icon><Warning /></el-icon>
              </el-button>
            </el-tooltip>

            <!-- 用户下拉菜单 -->
            <el-dropdown @command="handleUserCommand" trigger="click">
              <div class="user-dropdown-trigger interactive-element">
                <el-avatar :size="32" :src="getUserAvatar()">
                  <el-icon><User /></el-icon>
                </el-avatar>
                <span class="username desktop-only">{{ getDisplayName() }}</span>
                <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">
                    <el-icon><User /></el-icon>
                    个人中心
                  </el-dropdown-item>
                  <el-dropdown-item command="orders">
                    <el-icon><ShoppingBag /></el-icon>
                    我的订单
                  </el-dropdown-item>
                  <el-dropdown-item command="settings">
                    <el-icon><Setting /></el-icon>
                    账户设置
                  </el-dropdown-item>
                  <el-dropdown-item divided command="logout">
                    <el-icon><SwitchButton /></el-icon>
                    退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  House,
  Timer,
  ShoppingBag,
  User,
  Setting,
  SwitchButton,
  ArrowDown,
  Warning,
  Menu
} from '@element-plus/icons-vue'
import { useAuth } from '@/composables/useAuth'
import { useAutoWebSocket } from '@/composables/useWebSocket'
import { useRealTimeFeedback } from '@/composables/useRealTimeFeedback'
import { useNotificationSystem } from '@/composables/useNotificationSystem'
import WebSocketStatus from '@/components/common/WebSocketStatus.vue'
import FeedbackNotification from '@/components/common/FeedbackNotification.vue'

// 路由
const router = useRouter()
const route = useRoute()

// 认证相关
const {
  isAuthenticated,
  isEmailVerified,
  getUserAvatar,
  getDisplayName,
  logout
} = useAuth()

// WebSocket连接
useAutoWebSocket()

// 实时用户反馈
useRealTimeFeedback()

// 消息通知系统
useNotificationSystem()

// 状态
const mobileMenuVisible = ref(false)

// 计算当前激活的菜单
const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/activities')) return '/activities'
  return '/'
})

// 处理菜单选择
const handleMenuSelect = (index: string) => {
  if (index !== route.path) {
    router.push(index)
  }
}

// 切换移动端菜单
const toggleMobileMenu = () => {
  mobileMenuVisible.value = !mobileMenuVisible.value
}

// 处理用户下拉菜单命令
const handleUserCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/user/profile')
      break
    case 'orders':
      router.push('/user/orders')
      break
    case 'settings':
      router.push('/user/settings')
      break
    case 'logout':
      logout()
      break
  }
}

// 跳转到登录页
const goToLogin = () => {
  router.push('/auth/login')
}

// 跳转到注册页
const goToRegister = () => {
  router.push('/auth/register')
}

// 跳转到邮箱验证页
const goToVerifyEmail = () => {
  router.push('/auth/verify-email')
}
</script>

<style scoped lang="scss">
.app-header {
  background: white;
  border-bottom: 1px solid var(--el-border-color-lighter);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  position: sticky;
  top: 0;
  z-index: 1000;

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 24px;
  }

  .header-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 64px;

    .logo-section {
      .logo-link {
        display: flex;
        align-items: center;
        text-decoration: none;
        color: var(--el-text-color-primary);
        font-weight: 600;
        font-size: 20px;

        .logo-icon {
          font-size: 28px;
          color: var(--el-color-primary);
          margin-right: 8px;
        }

        .logo-text {
          background: linear-gradient(135deg, var(--el-color-primary), #667eea);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
        }

        &:hover {
          .logo-icon {
            transform: scale(1.1);
            transition: transform 0.3s ease;
          }
        }
      }
    }

    .nav-menu {
      flex: 1;
      display: flex;
      justify-content: center;

      .header-menu {
        border-bottom: none;

        .el-menu-item {
          border-bottom: 2px solid transparent;
          
          &:hover {
            background: var(--el-color-primary-light-9);
            border-bottom-color: var(--el-color-primary-light-5);
          }

          &.is-active {
            border-bottom-color: var(--el-color-primary);
            background: var(--el-color-primary-light-9);
          }
        }
      }
    }

    .user-actions {
      display: flex;
      align-items: center;
      gap: 16px;

      .websocket-status {
        display: flex;
        align-items: center;
      }

      .auth-buttons {
        display: flex;
        gap: 12px;
      }

      .user-info {
        display: flex;
        align-items: center;
        gap: 12px;

        .user-dropdown-trigger {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 8px 12px;
          border-radius: 6px;
          cursor: pointer;
          transition: background-color 0.3s ease;

          &:hover {
            background: var(--el-bg-color-page);
          }

          .username {
            font-size: 14px;
            color: var(--el-text-color-primary);
            font-weight: 500;
          }

          .dropdown-icon {
            font-size: 12px;
            color: var(--el-text-color-placeholder);
            transition: transform 0.3s ease;
          }

          &:hover .dropdown-icon {
            transform: rotate(180deg);
          }
        }
      }
    }
  }
}

@media (max-width: 768px) {
  .app-header {
    .container {
      padding: 0 16px;
    }

    .header-content {
      .nav-menu {
        display: none;
      }

      .user-info {
        .username {
          display: none;
        }
      }
    }
  }
}
</style>
