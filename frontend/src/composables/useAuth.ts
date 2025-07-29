import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

/**
 * 认证相关的组合式函数
 */
export function useAuth() {
  const router = useRouter()
  const authStore = useAuthStore()

  // 计算属性
  const isAuthenticated = computed(() => authStore.isAuthenticated)
  const user = computed(() => authStore.user)
  const userName = computed(() => authStore.userName)
  const userEmail = computed(() => authStore.userEmail)
  const isEmailVerified = computed(() => authStore.isEmailVerified)

  /**
   * 检查是否已登录，未登录则跳转到登录页
   */
  const requireAuth = (redirectPath?: string) => {
    if (!isAuthenticated.value) {
      ElMessage.warning('请先登录')
      router.push({
        name: 'login',
        query: { redirect: redirectPath || router.currentRoute.value.fullPath }
      })
      return false
    }
    return true
  }

  /**
   * 检查邮箱是否已验证
   */
  const requireEmailVerified = () => {
    if (!isEmailVerified.value) {
      ElMessageBox.confirm(
        '您的邮箱尚未验证，请先完成邮箱验证',
        '邮箱验证',
        {
          confirmButtonText: '去验证',
          cancelButtonText: '稍后验证',
          type: 'warning',
        }
      ).then(() => {
        router.push('/auth/verify-email')
      }).catch(() => {
        // 用户选择稍后验证
      })
      return false
    }
    return true
  }

  /**
   * 安全登出
   */
  const logout = async () => {
    try {
      await ElMessageBox.confirm(
        '确定要退出登录吗？',
        '退出登录',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        }
      )

      await authStore.logout()
      router.push('/')
    } catch (error) {
      // 用户取消或登出失败
    }
  }

  /**
   * 检查用户权限
   */
  const hasPermission = (permission: string) => {
    // 这里可以根据实际需求实现权限检查逻辑
    // 目前简单返回是否已登录
    return isAuthenticated.value
  }

  /**
   * 检查用户角色
   */
  const hasRole = (role: string) => {
    // 这里可以根据实际需求实现角色检查逻辑
    // 目前简单返回是否已登录
    return isAuthenticated.value
  }

  /**
   * 获取用户头像URL
   */
  const getUserAvatar = () => {
    // 这里可以返回用户头像URL
    // 如果没有头像，返回默认头像
    return user.value?.avatar || ''
  }

  /**
   * 获取用户显示名称
   */
  const getDisplayName = () => {
    if (user.value?.first_name && user.value?.last_name) {
      return `${user.value.first_name} ${user.value.last_name}`
    }
    return userName.value
  }

  /**
   * 检查是否为新用户（需要完善信息）
   */
  const isNewUser = computed(() => {
    return isAuthenticated.value && 
           (!user.value?.first_name || !user.value?.last_name)
  })

  /**
   * 引导新用户完善信息
   */
  const guideNewUser = () => {
    if (isNewUser.value) {
      ElMessageBox.confirm(
        '为了更好的使用体验，建议您完善个人信息',
        '完善信息',
        {
          confirmButtonText: '去完善',
          cancelButtonText: '稍后',
          type: 'info',
        }
      ).then(() => {
        router.push('/user/profile')
      }).catch(() => {
        // 用户选择稍后
      })
    }
  }

  return {
    // 状态
    isAuthenticated,
    user,
    userName,
    userEmail,
    isEmailVerified,
    isNewUser,

    // 方法
    requireAuth,
    requireEmailVerified,
    logout,
    hasPermission,
    hasRole,
    getUserAvatar,
    getDisplayName,
    guideNewUser,
  }
}
