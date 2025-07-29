import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { activityApi, seckillApi } from '@/api/seckill'
import type { SeckillActivity, StockInfo } from '@/types'

export const useActivityStore = defineStore('activity', () => {
  // 状态
  const activities = ref<SeckillActivity[]>([])
  const currentActivity = ref<SeckillActivity | null>(null)
  const stockInfos = ref<Record<number, StockInfo>>({})
  const loading = ref(false)
  const pagination = ref({
    page: 1,
    pageSize: 12,
    total: 0,
  })
  const filters = ref({
    status: '',
    search: '',
  })

  // 计算属性
  const activeActivities = computed(() => 
    activities.value.filter(activity => activity.status === 'active')
  )
  
  const upcomingActivities = computed(() => 
    activities.value.filter(activity => activity.status === 'pending')
  )
  
  const endedActivities = computed(() => 
    activities.value.filter(activity => activity.status === 'ended')
  )

  // 获取活动列表
  const fetchActivities = async (params?: {
    page?: number
    pageSize?: number
    status?: string
    search?: string
  }) => {
    loading.value = true
    try {
      const queryParams: {
        page?: number
        page_size?: number
        status?: 'pending' | 'active' | 'ended' | 'cancelled'
        search?: string
      } = {
        page: params?.page || pagination.value.page,
        page_size: params?.pageSize || pagination.value.pageSize,
        status: (params?.status || filters.value.status) as 'pending' | 'active' | 'ended' | 'cancelled' | undefined,
        search: params?.search || filters.value.search,
      }

      // 移除空值参数
      Object.keys(queryParams).forEach(key => {
        if (!queryParams[key as keyof typeof queryParams]) {
          delete queryParams[key as keyof typeof queryParams]
        }
      })

      const response = await activityApi.getActivities(queryParams)
      
      activities.value = response.results
      pagination.value.total = response.count
      pagination.value.page = queryParams.page || 1

      // 批量获取库存信息
      if (response.results.length > 0) {
        await fetchBatchStockInfo(response.results.map(activity => activity.id))
      }

      return response
    } catch (error: any) {
      ElMessage.error(error.message || '获取活动列表失败')
      throw error
    } finally {
      loading.value = false
    }
  }

  // 获取活动详情
  const fetchActivity = async (id: number) => {
    loading.value = true
    try {
      const activity = await activityApi.getActivity(id)
      currentActivity.value = activity
      
      // 获取库存信息
      await fetchStockInfo(id)
      
      return activity
    } catch (error: any) {
      ElMessage.error(error.message || '获取活动详情失败')
      throw error
    } finally {
      loading.value = false
    }
  }

  // 获取库存信息
  const fetchStockInfo = async (activityId: number) => {
    try {
      const stockInfo = await seckillApi.getStock(activityId)
      stockInfos.value[activityId] = stockInfo
      return stockInfo
    } catch (error: any) {
      console.error('获取库存信息失败:', error)
      // 不显示错误消息，因为这可能是正常的（活动未开始等）
    }
  }

  // 批量获取库存信息
  const fetchBatchStockInfo = async (activityIds: number[]) => {
    try {
      const batchStockInfo = await seckillApi.getBatchStock(activityIds)
      Object.assign(stockInfos.value, batchStockInfo)
      return batchStockInfo
    } catch (error: any) {
      console.error('批量获取库存信息失败:', error)
    }
  }

  // 获取热门活动
  const fetchHotActivities = async (limit = 10) => {
    try {
      const hotActivities = await activityApi.getHotActivities(limit)
      return hotActivities
    } catch (error: any) {
      ElMessage.error(error.message || '获取热门活动失败')
      throw error
    }
  }

  // 获取即将开始的活动
  const fetchUpcomingActivities = async (limit = 10) => {
    try {
      const upcoming = await activityApi.getUpcomingActivities(limit)
      return upcoming
    } catch (error: any) {
      ElMessage.error(error.message || '获取即将开始的活动失败')
      throw error
    }
  }

  // 获取正在进行的活动
  const fetchActiveActivities = async (limit = 10) => {
    try {
      const active = await activityApi.getActiveActivities(limit)
      return active
    } catch (error: any) {
      ElMessage.error(error.message || '获取正在进行的活动失败')
      throw error
    }
  }

  // 搜索活动
  const searchActivities = async (query: string, params?: {
    page?: number
    pageSize?: number
  }) => {
    loading.value = true
    try {
      const response = await activityApi.searchActivities(query, params)
      activities.value = response.results
      pagination.value.total = response.count
      pagination.value.page = params?.page || 1
      
      // 批量获取库存信息
      if (response.results.length > 0) {
        await fetchBatchStockInfo(response.results.map(activity => activity.id))
      }
      
      return response
    } catch (error: any) {
      ElMessage.error(error.message || '搜索活动失败')
      throw error
    } finally {
      loading.value = false
    }
  }

  // 参与秒杀
  const participateInSeckill = async (activityId: number) => {
    try {
      const result = await seckillApi.participate(activityId)
      
      // 更新库存信息
      await fetchStockInfo(activityId)
      
      return result
    } catch (error: any) {
      throw error
    }
  }

  // 获取用户参与记录
  const fetchUserParticipation = async (activityId: number) => {
    try {
      const participation = await seckillApi.getUserParticipation(activityId)
      return participation
    } catch (error: any) {
      console.error('获取用户参与记录失败:', error)
      return null
    }
  }

  // 设置筛选条件
  const setFilters = (newFilters: Partial<typeof filters.value>) => {
    Object.assign(filters.value, newFilters)
  }

  // 重置筛选条件
  const resetFilters = () => {
    filters.value = {
      status: '',
      search: '',
    }
    pagination.value.page = 1
  }

  // 设置分页
  const setPagination = (newPagination: Partial<typeof pagination.value>) => {
    Object.assign(pagination.value, newPagination)
  }

  // 获取活动的库存信息
  const getActivityStockInfo = (activityId: number) => {
    return stockInfos.value[activityId] || null
  }

  // 获取活动状态文本
  const getActivityStatusText = (status: string) => {
    const statusMap = {
      pending: '即将开始',
      active: '进行中',
      ended: '已结束',
      cancelled: '已取消',
    }
    return statusMap[status as keyof typeof statusMap] || '未知'
  }

  // 获取活动状态类型
  const getActivityStatusType = (status: string) => {
    const typeMap = {
      pending: 'warning',
      active: 'success',
      ended: 'info',
      cancelled: 'danger',
    }
    return typeMap[status as keyof typeof typeMap] || 'info'
  }

  return {
    // 状态
    activities,
    currentActivity,
    stockInfos,
    loading,
    pagination,
    filters,
    
    // 计算属性
    activeActivities,
    upcomingActivities,
    endedActivities,
    
    // 方法
    fetchActivities,
    fetchActivity,
    fetchStockInfo,
    fetchBatchStockInfo,
    fetchHotActivities,
    fetchUpcomingActivities,
    fetchActiveActivities,
    searchActivities,
    participateInSeckill,
    fetchUserParticipation,
    setFilters,
    resetFilters,
    setPagination,
    getActivityStockInfo,
    getActivityStatusText,
    getActivityStatusType,
  }
})
