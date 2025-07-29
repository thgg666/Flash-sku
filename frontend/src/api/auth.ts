import { djangoClient } from './http'
import type { 
  LoginRequest, 
  RegisterRequest, 
  AuthResponse, 
  User,
  UserProfile 
} from '@/types'

/**
 * 认证相关API
 */
export const authApi = {
  /**
   * 用户登录
   */
  login(data: LoginRequest): Promise<AuthResponse> {
    return djangoClient.post('/auth/login/', data)
  },

  /**
   * 用户注册
   */
  register(data: RegisterRequest): Promise<{ message: string }> {
    return djangoClient.post('/auth/register/', data)
  },

  /**
   * 刷新Token
   */
  refreshToken(refreshToken: string): Promise<{ access: string }> {
    return djangoClient.post('/auth/refresh/', { refresh: refreshToken })
  },

  /**
   * 用户登出
   */
  logout(): Promise<{ message: string }> {
    return djangoClient.post('/auth/logout/')
  },

  /**
   * 获取当前用户信息
   */
  getCurrentUser(): Promise<User> {
    return djangoClient.get('/auth/user/')
  },

  /**
   * 更新用户信息
   */
  updateUser(data: Partial<User>): Promise<User> {
    return djangoClient.patch('/auth/user/', data)
  },

  /**
   * 修改密码
   */
  changePassword(data: {
    old_password: string
    new_password: string
    new_password_confirm: string
  }): Promise<{ message: string }> {
    return djangoClient.post('/auth/change-password/', data)
  },

  /**
   * 发送邮箱验证码
   */
  sendEmailVerification(email: string): Promise<{ message: string }> {
    return djangoClient.post('/auth/send-verification/', { email })
  },

  /**
   * 验证邮箱
   */
  verifyEmail(data: {
    email: string
    code: string
  }): Promise<{ message: string }> {
    return djangoClient.post('/auth/verify-email/', data)
  },

  /**
   * 发送密码重置邮件
   */
  sendPasswordReset(email: string): Promise<{ message: string }> {
    return djangoClient.post('/auth/password-reset/', { email })
  },

  /**
   * 重置密码
   */
  resetPassword(data: {
    email: string
    code: string
    new_password: string
    new_password_confirm: string
  }): Promise<{ message: string }> {
    return djangoClient.post('/auth/password-reset-confirm/', data)
  },

  /**
   * 获取图片验证码
   */
  getCaptcha(): Promise<{ image: string; key: string }> {
    return djangoClient.get('/auth/captcha/')
  },

  /**
   * 验证图片验证码
   */
  verifyCaptcha(data: {
    key: string
    code: string
  }): Promise<{ valid: boolean }> {
    return djangoClient.post('/auth/captcha/verify/', data)
  },

  /**
   * 获取用户资料
   */
  getUserProfile(): Promise<UserProfile> {
    return djangoClient.get('/auth/profile/')
  },

  /**
   * 更新用户资料
   */
  updateUserProfile(data: Partial<UserProfile>): Promise<UserProfile> {
    return djangoClient.patch('/auth/profile/', data)
  },

  /**
   * 上传头像
   */
  uploadAvatar(file: File): Promise<{ avatar_url: string }> {
    return djangoClient.upload('/auth/avatar/', file)
  },

  /**
   * 检查用户名是否可用
   */
  checkUsername(username: string): Promise<{ available: boolean }> {
    return djangoClient.get(`/auth/check-username/?username=${username}`)
  },

  /**
   * 检查邮箱是否可用
   */
  checkEmail(email: string): Promise<{ available: boolean }> {
    return djangoClient.get(`/auth/check-email/?email=${email}`)
  },

  /**
   * 获取用户订单列表
   */
  getUserOrders(params?: {
    page?: number
    page_size?: number
    status?: string
  }): Promise<{
    count: number
    next?: string
    previous?: string
    results: any[]
  }> {
    return djangoClient.get('/auth/orders/', { params })
  },

  /**
   * 删除账户
   */
  deleteAccount(password: string): Promise<{ message: string }> {
    return djangoClient.post('/auth/delete-account/', { password })
  },

  /**
   * 忘记密码
   */
  forgotPassword(data: { email: string; captcha: string }): Promise<{ message: string }> {
    return djangoClient.post('/auth/forgot-password/', data)
  },

  /**
   * 验证重置密码token
   */
  verifyResetToken(token: string): Promise<{ valid: boolean }> {
    return djangoClient.post('/auth/verify-reset-token/', { token })
  },
}
