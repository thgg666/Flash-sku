import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { authApi } from '@/api/auth'
import { storage } from '@/utils'
import type { User, UserProfile, LoginRequest, RegisterRequest, AuthResponse } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const user = ref<User | null>(null)
  const userProfile = ref<UserProfile | null>(null)
  const isLoggedIn = ref(false)
  const loading = ref(false)
  const token = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)

  // 计算属性
  const isAuthenticated = computed(() => isLoggedIn.value && !!token.value)
  const userName = computed(() => user.value?.username || '')
  const userEmail = computed(() => user.value?.email || '')
  const isEmailVerified = computed(() => user.value?.is_active || false)

  // 初始化认证状态
  const initAuth = () => {
    const savedToken = storage.get('access_token')
    const savedRefreshToken = storage.get('refresh_token')
    const savedUser = storage.get('user')

    if (savedToken && savedUser) {
      token.value = savedToken
      refreshToken.value = savedRefreshToken
      user.value = savedUser
      isLoggedIn.value = true
    }
  }

  // 用户注册
  const register = async (data: RegisterRequest) => {
    loading.value = true
    try {
      const response = await authApi.register(data)
      ElMessage.success(response.message || '注册成功，请查收邮箱验证邮件')
      return { success: true, message: response.message }
    } catch (error: any) {
      const message = error.message || '注册失败'
      ElMessage.error(message)
      return { success: false, message }
    } finally {
      loading.value = false
    }
  }

  // 用户登录
  const login = async (data: LoginRequest) => {
    loading.value = true
    try {
      const response: AuthResponse = await authApi.login(data)
      
      // 保存认证信息
      token.value = response.access_token
      refreshToken.value = response.refresh_token
      user.value = response.user
      isLoggedIn.value = true

      // 持久化存储
      storage.set('access_token', response.access_token)
      storage.set('refresh_token', response.refresh_token)
      storage.set('user', response.user)

      ElMessage.success('登录成功')
      return { success: true, user: response.user }
    } catch (error: any) {
      const message = error.message || '登录失败'
      ElMessage.error(message)
      return { success: false, message }
    } finally {
      loading.value = false
    }
  }

  // 用户登出
  const logout = async () => {
    loading.value = true
    try {
      await authApi.logout()
    } catch (error) {
      console.warn('登出API调用失败:', error)
    } finally {
      // 清除本地状态
      clearAuth()
      ElMessage.success('已退出登录')
      loading.value = false
    }
  }

  // 清除认证状态
  const clearAuth = () => {
    user.value = null
    userProfile.value = null
    token.value = null
    refreshToken.value = null
    isLoggedIn.value = false
    
    // 清除存储
    storage.remove('access_token')
    storage.remove('refresh_token')
    storage.remove('user')
  }

  // 获取当前用户信息
  const fetchCurrentUser = async () => {
    if (!isAuthenticated.value) return

    loading.value = true
    try {
      const userData = await authApi.getCurrentUser()
      user.value = userData
      storage.set('user', userData)
      return userData
    } catch (error: any) {
      console.error('获取用户信息失败:', error)
      if (error.response?.status === 401) {
        clearAuth()
      }
    } finally {
      loading.value = false
    }
  }

  // 获取用户资料
  const fetchUserProfile = async () => {
    if (!isAuthenticated.value) return

    loading.value = true
    try {
      const profile = await authApi.getUserProfile()
      userProfile.value = profile
      return profile
    } catch (error: any) {
      console.error('获取用户资料失败:', error)
    } finally {
      loading.value = false
    }
  }

  // 更新用户信息
  const updateUser = async (data: Partial<User>) => {
    loading.value = true
    try {
      const updatedUser = await authApi.updateUser(data)
      user.value = updatedUser
      storage.set('user', updatedUser)
      ElMessage.success('用户信息更新成功')
      return { success: true, user: updatedUser }
    } catch (error: any) {
      const message = error.message || '更新失败'
      ElMessage.error(message)
      return { success: false, message }
    } finally {
      loading.value = false
    }
  }

  // 修改密码
  const changePassword = async (data: {
    old_password: string
    new_password: string
    new_password_confirm: string
  }) => {
    loading.value = true
    try {
      const response = await authApi.changePassword(data)
      ElMessage.success(response.message || '密码修改成功')
      return { success: true, message: response.message }
    } catch (error: any) {
      const message = error.message || '密码修改失败'
      ElMessage.error(message)
      return { success: false, message }
    } finally {
      loading.value = false
    }
  }

  // 发送邮箱验证
  const sendEmailVerification = async (email: string) => {
    loading.value = true
    try {
      const response = await authApi.sendEmailVerification(email)
      ElMessage.success(response.message || '验证邮件已发送')
      return { success: true, message: response.message }
    } catch (error: any) {
      const message = error.message || '发送失败'
      ElMessage.error(message)
      return { success: false, message }
    } finally {
      loading.value = false
    }
  }

  // 验证邮箱
  const verifyEmail = async (data: { email: string; code: string }) => {
    loading.value = true
    try {
      const response = await authApi.verifyEmail(data)
      ElMessage.success(response.message || '邮箱验证成功')
      
      // 更新用户状态
      if (user.value) {
        user.value.is_active = true
        storage.set('user', user.value)
      }
      
      return { success: true, message: response.message }
    } catch (error: any) {
      const message = error.message || '验证失败'
      ElMessage.error(message)
      return { success: false, message }
    } finally {
      loading.value = false
    }
  }

  // 检查用户名可用性
  const checkUsername = async (username: string) => {
    try {
      const response = await authApi.checkUsername(username)
      return response.available
    } catch (error) {
      return false
    }
  }

  // 检查邮箱可用性
  const checkEmail = async (email: string) => {
    try {
      const response = await authApi.checkEmail(email)
      return response.available
    } catch (error) {
      return false
    }
  }

  return {
    // 状态
    user,
    userProfile,
    isLoggedIn,
    loading,
    token,
    refreshToken,
    
    // 计算属性
    isAuthenticated,
    userName,
    userEmail,
    isEmailVerified,
    
    // 方法
    initAuth,
    register,
    login,
    logout,
    clearAuth,
    fetchCurrentUser,
    fetchUserProfile,
    updateUser,
    changePassword,
    sendEmailVerification,
    verifyEmail,
    checkUsername,
    checkEmail,
  }
})
