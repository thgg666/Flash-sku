// 用户相关类型
export interface User {
  id: number
  username: string
  email: string
  first_name?: string
  last_name?: string
  avatar?: string
  is_active: boolean
  date_joined: string
}

export interface UserProfile {
  id: number
  user: number
  phone?: string
  avatar?: string
  birth_date?: string
  gender?: 'M' | 'F' | 'O'
  address?: string
  created_at: string
  updated_at: string
}

// 商品相关类型
export interface Product {
  id: number
  name: string
  description?: string
  image_url?: string
  category_id?: number
  created_at: string
}

export interface Category {
  id: number
  name: string
  description?: string
  parent_id?: number
  created_at: string
}

// 秒杀活动相关类型
export interface SeckillActivity {
  id: number
  product_id: number
  product: Product
  name: string
  start_time: string
  end_time: string
  original_price: string
  seckill_price: string
  total_stock: number
  available_stock: number
  max_per_user: number
  status: 'pending' | 'active' | 'ended' | 'cancelled'
  created_at: string
}

// 订单相关类型
export interface Order {
  id: number
  user_id: number
  activity_id: number
  product_name: string
  seckill_price: string
  status: 'pending_payment' | 'paid' | 'cancelled' | 'expired'
  payment_deadline?: string
  created_at: string
}

// API响应类型
export interface ApiResponse<T = any> {
  code: string
  message: string
  data?: T
}

export interface PaginatedResponse<T> {
  count: number
  next?: string
  previous?: string
  results: T[]
}

// 认证相关类型
export interface LoginRequest {
  username: string
  password: string
  captcha?: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  password_confirm: string
  captcha: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user: User
}

// 秒杀相关类型
export interface SeckillRequest {
  activity_id: number
  user_id: number
}

export interface StockInfo {
  activity_id: number
  available_stock: number
  total_stock: number
  status: 'normal' | 'low_stock' | 'out_of_stock'
  activity_status: 'pending' | 'active' | 'ended'
  last_updated?: string
}

export interface SeckillResult {
  code: 'SUCCESS' | 'FAILURE' | 'RATE_LIMIT' | 'SOLD_OUT' | 'LIMIT_EXCEEDED' | 'ACTIVITY_ENDED' | 'INSUFFICIENT_STOCK'
  message: string
  order_id?: string
  quantity?: number
  amount?: string
  data?: any
}

// 表单验证类型
export interface FormRules {
  [key: string]: Array<{
    required?: boolean
    message: string
    trigger?: string | string[]
    min?: number
    max?: number
    pattern?: RegExp
    validator?: (rule: any, value: any, callback: any) => void
  }>
}

// 路由元信息类型
export interface RouteMeta extends Record<PropertyKey, unknown> {
  title?: string
  requiresAuth?: boolean
  roles?: string[]
  icon?: string
  hidden?: boolean
}
