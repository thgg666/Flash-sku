<template>
  <div class="activity-list-view">
    <div class="container">
      <!-- 页面头部 -->
      <div class="page-header">
        <h1>秒杀活动</h1>
        <p>精选商品，限时抢购</p>
      </div>

      <!-- 筛选和搜索 -->
      <div class="filter-section">
        <div class="filter-bar">
          <!-- 状态筛选 -->
          <div class="status-filters">
            <el-button
              :type="filters.status === '' ? 'primary' : ''"
              @click="handleStatusFilter('')"
            >
              全部
            </el-button>
            <el-button
              :type="filters.status === 'active' ? 'primary' : ''"
              @click="handleStatusFilter('active')"
            >
              进行中
            </el-button>
            <el-button
              :type="filters.status === 'pending' ? 'primary' : ''"
              @click="handleStatusFilter('pending')"
            >
              即将开始
            </el-button>
            <el-button
              :type="filters.status === 'ended' ? 'primary' : ''"
              @click="handleStatusFilter('ended')"
            >
              已结束
            </el-button>
          </div>

          <!-- 搜索框 -->
          <div class="search-box">
            <el-input
              v-model="searchQuery"
              placeholder="搜索商品或活动名称"
              :prefix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearchClear"
            >
              <template #append>
                <el-button :icon="Search" @click="handleSearch" />
              </template>
            </el-input>
          </div>
        </div>

        <!-- 排序和视图切换 -->
        <div class="toolbar">
          <div class="sort-options">
            <el-select v-model="sortBy" placeholder="排序方式" @change="handleSort">
              <el-option label="默认排序" value="default" />
              <el-option label="价格从低到高" value="price_asc" />
              <el-option label="价格从高到低" value="price_desc" />
              <el-option label="最新发布" value="newest" />
              <el-option label="即将结束" value="ending_soon" />
            </el-select>
          </div>

          <div class="view-toggle">
            <el-button-group>
              <el-button
                :type="viewMode === 'grid' ? 'primary' : ''"
                :icon="Grid"
                @click="viewMode = 'grid'"
              />
              <el-button
                :type="viewMode === 'list' ? 'primary' : ''"
                :icon="List"
                @click="viewMode = 'list'"
              />
            </el-button-group>
          </div>
        </div>
      </div>

      <!-- 活动统计 -->
      <div class="stats-section">
        <el-row :gutter="16">
          <el-col :xs="12" :sm="6">
            <div class="stat-card">
              <div class="stat-number">{{ activeCount }}</div>
              <div class="stat-label">进行中</div>
            </div>
          </el-col>
          <el-col :xs="12" :sm="6">
            <div class="stat-card">
              <div class="stat-number">{{ upcomingCount }}</div>
              <div class="stat-label">即将开始</div>
            </div>
          </el-col>
          <el-col :xs="12" :sm="6">
            <div class="stat-card">
              <div class="stat-number">{{ totalCount }}</div>
              <div class="stat-label">总活动数</div>
            </div>
          </el-col>
          <el-col :xs="12" :sm="6">
            <div class="stat-card">
              <div class="stat-number">{{ formatNumber(totalParticipants) }}</div>
              <div class="stat-label">参与人次</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 活动列表 -->
      <div class="activities-section">
        <!-- 加载状态 -->
        <div v-if="loading && activities.length === 0" class="loading-section">
          <el-skeleton :rows="6" animated />
        </div>

        <!-- 空状态 -->
        <div v-else-if="!loading && activities.length === 0" class="empty-section">
          <el-empty
            description="暂无活动数据"
            :image-size="120"
          >
            <el-button type="primary" @click="handleRefresh">
              刷新页面
            </el-button>
          </el-empty>
        </div>

        <!-- 活动网格 -->
        <div v-else class="activities-grid" :class="{ 'list-view': viewMode === 'list' }">
          <ActivityCard
            v-for="activity in activities"
            :key="activity.id"
            :activity="activity"
            :stock-info="getActivityStockInfo(activity.id)"
            @participate="handleParticipate"
            @click="handleActivityClick"
          />
        </div>

        <!-- 分页 -->
        <div v-if="activities.length > 0" class="pagination-section">
          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :total="pagination.total"
            :page-sizes="[12, 24, 48, 96]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Grid, List } from '@element-plus/icons-vue'
import { useActivityStore } from '@/stores/activity'
import { useAuth } from '@/composables/useAuth'
import { useRealTimeBatchActivity } from '@/composables/useRealTimeActivity'
import { formatNumber } from '@/utils'
import ActivityCard from '@/components/activity/ActivityCard.vue'
import type { SeckillActivity } from '@/types'

// 路由和store
const router = useRouter()
const activityStore = useActivityStore()
const { requireAuth } = useAuth()

// 状态
const searchQuery = ref('')
const sortBy = ref('default')
const viewMode = ref<'grid' | 'list'>('grid')

// 计算属性
const {
  activities,
  loading,
  pagination,
  filters,
  activeActivities,
  upcomingActivities
} = activityStore

// 实时活动状态更新 (在activities定义后初始化)
const realTimeBatchActivity = useRealTimeBatchActivity([])

// 监听activities变化，更新实时订阅
watch(activities, (newActivities) => {
  if (newActivities.length > 0) {
    const activityIds = newActivities.map((a: SeckillActivity) => a.id)
    realTimeBatchActivity.subscribeBatch(activityIds)
  }
}, { immediate: true })

const activeCount = computed(() => activeActivities.length)
const upcomingCount = computed(() => upcomingActivities.length)
const totalCount = computed(() => pagination.total)
const totalParticipants = computed(() => {
  // 这里可以从API获取总参与人次
  return 12580
})

// 获取活动库存信息
const getActivityStockInfo = (activityId: number) => {
  return activityStore.getActivityStockInfo(activityId)
}

// 处理状态筛选
const handleStatusFilter = (status: string) => {
  activityStore.setFilters({ status })
  activityStore.setPagination({ page: 1 })
  fetchActivities()
}

// 处理搜索
const handleSearch = () => {
  if (searchQuery.value.trim()) {
    activityStore.searchActivities(searchQuery.value.trim(), {
      page: 1,
      pageSize: pagination.pageSize
    })
  } else {
    handleSearchClear()
  }
}

// 清除搜索
const handleSearchClear = () => {
  searchQuery.value = ''
  activityStore.setFilters({ search: '' })
  activityStore.setPagination({ page: 1 })
  fetchActivities()
}

// 处理排序
const handleSort = () => {
  // 这里可以根据sortBy的值来排序
  // 目前先简单刷新数据
  fetchActivities()
}

// 处理分页大小变化
const handleSizeChange = (size: number) => {
  activityStore.setPagination({ pageSize: size, page: 1 })
  fetchActivities()
}

// 处理页码变化
const handleCurrentChange = (page: number) => {
  activityStore.setPagination({ page })
  fetchActivities()
}

// 处理参与秒杀
const handleParticipate = async (activityId: number) => {
  if (!requireAuth()) return

  try {
    const result = await activityStore.participateInSeckill(activityId)
    
    if (result.code === 'SUCCESS') {
      ElMessage.success('抢购成功！')
      // 可以跳转到订单页面
      if (result.order_id) {
        router.push(`/user/orders/${result.order_id}`)
      }
    } else {
      ElMessage.warning(result.message || '抢购失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '抢购失败')
  }
}

// 处理活动点击
const handleActivityClick = (activity: SeckillActivity) => {
  router.push(`/activity/${activity.id}`)
}

// 刷新页面
const handleRefresh = () => {
  fetchActivities()
}

// 获取活动列表
const fetchActivities = () => {
  activityStore.fetchActivities({
    page: pagination.page,
    pageSize: pagination.pageSize,
    status: filters.status,
    search: filters.search,
  })
}

// 监听筛选条件变化
watch(() => filters.status, () => {
  fetchActivities()
})

// 组件挂载时获取数据
onMounted(() => {
  fetchActivities()
})
</script>

<style scoped lang="scss">
.activity-list-view {
  min-height: 100vh;
  background: var(--el-bg-color-page);
  padding: 24px 0;

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 24px;
  }

  .page-header {
    text-align: center;
    margin-bottom: 32px;

    h1 {
      margin: 0 0 8px 0;
      color: var(--el-text-color-primary);
      font-size: 32px;
      font-weight: 600;
    }

    p {
      margin: 0;
      color: var(--el-text-color-regular);
      font-size: 16px;
    }
  }

  .filter-section {
    background: white;
    border-radius: 12px;
    padding: 24px;
    margin-bottom: 24px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);

    .filter-bar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      gap: 16px;

      .status-filters {
        display: flex;
        gap: 8px;
      }

      .search-box {
        width: 300px;
      }
    }

    .toolbar {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .sort-options {
        .el-select {
          width: 150px;
        }
      }
    }
  }

  .stats-section {
    margin-bottom: 24px;

    .stat-card {
      background: white;
      border-radius: 8px;
      padding: 20px;
      text-align: center;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

      .stat-number {
        font-size: 24px;
        font-weight: 600;
        color: var(--el-color-primary);
        margin-bottom: 4px;
      }

      .stat-label {
        font-size: 14px;
        color: var(--el-text-color-regular);
      }
    }
  }

  .activities-section {
    .loading-section {
      background: white;
      border-radius: 12px;
      padding: 24px;
    }

    .empty-section {
      background: white;
      border-radius: 12px;
      padding: 48px 24px;
      text-align: center;
    }

    .activities-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: 24px;
      margin-bottom: 32px;

      &.list-view {
        grid-template-columns: 1fr;
      }
    }

    .pagination-section {
      display: flex;
      justify-content: center;
      padding: 24px 0;
    }
  }
}

@media (max-width: 768px) {
  .activity-list-view {
    padding: 16px 0;

    .container {
      padding: 0 16px;
    }

    .filter-section {
      padding: 16px;

      .filter-bar {
        flex-direction: column;
        align-items: stretch;
        gap: 12px;

        .status-filters {
          justify-content: center;
          flex-wrap: wrap;
        }

        .search-box {
          width: 100%;
        }
      }

      .toolbar {
        flex-direction: column;
        gap: 12px;
        align-items: stretch;
      }
    }

    .activities-grid {
      grid-template-columns: 1fr;
      gap: 16px;
    }
  }
}
</style>
