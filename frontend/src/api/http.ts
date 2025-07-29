import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { storage } from '@/utils'
import type { ApiResponse } from '@/types'

// 创建axios实例
const createHttpInstance = (baseURL: string): AxiosInstance => {
  const instance = axios.create({
    baseURL,
    timeout: 10000,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  // 请求拦截器
  instance.interceptors.request.use(
    (config) => {
      // 添加认证token
      const token = storage.get('access_token')
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }

      // 添加请求时间戳，防止缓存
      if (config.method === 'get') {
        config.params = {
          ...config.params,
          _t: Date.now(),
        }
      }

      return config
    },
    (error) => {
      console.error('请求拦截器错误:', error)
      return Promise.reject(error)
    }
  )

  // 响应拦截器
  instance.interceptors.response.use(
    (response: AxiosResponse<ApiResponse>) => {
      const { data } = response

      // 如果是文件下载等特殊响应，直接返回
      if (response.config.responseType === 'blob') {
        return response
      }

      // 检查业务状态码
      if (data.code === 'SUCCESS' || data.code === '200') {
        return data.data !== undefined ? data.data : data
      }

      // 处理业务错误
      const message = data.message || '请求失败'
      ElMessage.error(message)
      return Promise.reject(new Error(message))
    },
    async (error) => {
      const { response, config } = error

      // 网络错误
      if (!response) {
        ElMessage.error('网络连接失败，请检查网络设置')
        return Promise.reject(error)
      }

      const { status, data } = response

      switch (status) {
        case 401:
          // Token过期或无效
          const refreshToken = storage.get('refresh_token')
          if (refreshToken && !config._retry) {
            config._retry = true
            try {
              // 尝试刷新token
              const refreshResponse = await axios.post('/api/auth/refresh/', {
                refresh: refreshToken,
              })
              const newToken = refreshResponse.data.access
              storage.set('access_token', newToken)
              
              // 重新发送原请求
              config.headers.Authorization = `Bearer ${newToken}`
              return instance(config)
            } catch (refreshError) {
              // 刷新失败，清除token并跳转登录
              storage.remove('access_token')
              storage.remove('refresh_token')
              storage.remove('user')
              
              ElMessageBox.confirm(
                '登录状态已过期，请重新登录',
                '提示',
                {
                  confirmButtonText: '重新登录',
                  cancelButtonText: '取消',
                  type: 'warning',
                }
              ).then(() => {
                window.location.href = '/login'
              })
            }
          } else {
            ElMessage.error('登录状态已过期，请重新登录')
            // 跳转到登录页
            setTimeout(() => {
              window.location.href = '/login'
            }, 1000)
          }
          break

        case 403:
          ElMessage.error('没有权限访问该资源')
          break

        case 404:
          ElMessage.error('请求的资源不存在')
          break

        case 429:
          ElMessage.error('请求过于频繁，请稍后再试')
          break

        case 500:
          ElMessage.error('服务器内部错误')
          break

        default:
          const message = data?.message || `请求失败 (${status})`
          ElMessage.error(message)
      }

      return Promise.reject(error)
    }
  )

  return instance
}

// 创建Django API实例
export const djangoApi = createHttpInstance(
  import.meta.env.VITE_DJANGO_API_BASE_URL || 'http://localhost:8000/api'
)

// 创建Go API实例
export const goApi = createHttpInstance(
  import.meta.env.VITE_GO_API_BASE_URL || 'http://localhost:8080'
)

// 通用请求方法
export class HttpClient {
  private instance: AxiosInstance

  constructor(instance: AxiosInstance) {
    this.instance = instance
  }

  async get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.get<T>(url, config)
    return response.data
  }

  async post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.post<T>(url, data, config)
    return response.data
  }

  async put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.put<T>(url, data, config)
    return response.data
  }

  async patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.patch<T>(url, data, config)
    return response.data
  }

  async delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.delete<T>(url, config)
    return response.data
  }

  // 文件上传
  async upload<T = any>(url: string, file: File, onProgress?: (progress: number) => void): Promise<T> {
    const formData = new FormData()
    formData.append('file', file)

    const config: AxiosRequestConfig = {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }

    if (onProgress) {
      config.onUploadProgress = (progressEvent) => {
        if (progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        }
      }
    }

    const response = await this.instance.post<T>(url, formData, config)
    return response.data
  }

  // 文件下载
  async download(url: string, filename?: string): Promise<void> {
    const response = await this.instance.get(url, {
      responseType: 'blob',
    })

    const blob = new Blob([response.data])
    const downloadUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = filename || 'download'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(downloadUrl)
  }
}

// 导出HTTP客户端实例
export const djangoClient = new HttpClient(djangoApi)
export const goClient = new HttpClient(goApi)

// 默认导出Django客户端
export default djangoClient
