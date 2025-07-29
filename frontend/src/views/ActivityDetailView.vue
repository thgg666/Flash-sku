<template>
  <div class="activity-detail-view">
    <div class="container">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-section">
        <el-skeleton :rows="8" animated />
      </div>

      <!-- 活动详情 -->
      <div v-else-if="activity" class="activity-detail">
        <!-- 面包屑导航 -->
        <el-breadcrumb class="breadcrumb" separator="/">
          <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item :to="{ path: '/activities' }">秒杀活动</el-breadcrumb-item>
          <el-breadcrumb-item>{{ activity.name }}</el-breadcrumb-item>
        </el-breadcrumb>

        <el-row :gutter="32">
          <!-- 左侧：商品图片 -->
          <el-col :xs="24" :md="12">
            <div class="product-gallery">
              <div class="main-image">
                <el-image
                  :src="activity.product.image_url || defaultImage"
                  :alt="activity.product.name"
                  fit="cover"
                  class="product-image"
                >
                  <template #error>
                    <div class="image-error">
                      <el-icon><Picture /></el-icon>
                      <span>暂无图片</span>
                    </div>
                  </template>
                </el-image>

                <!-- 活动状态标签 -->
                <div class="status-overlay">
                  <el-tag
                    :type="getStatusType(activity.status)"
                    size="large"
                    effect="dark"
                  >
                    {{ getStatusText(activity.status) }}
                  </el-tag>
                </div>
              </div>
            </div>
          </el-col>

          <!-- 右侧：商品信息 -->
          <el-col :xs="24" :md="12">
            <div class="product-info">
              <!-- 商品标题 -->
              <h1 class="product-title">{{ activity.product.name }}</h1>
              
              <!-- 活动名称 -->
              <h2 class="activity-title">{{ activity.name }}</h2>

              <!-- 价格信息 -->
              <div class="price-section">
                <div class="seckill-price">
                  <span class="label">秒杀价</span>
                  <span class="currency">¥</span>
                  <span class="price">{{ formatPrice(activity.seckill_price) }}</span>
                </div>
                <div class="original-price">
                  <span class="label">原价</span>
                  <span class="price">¥{{ formatPrice(activity.original_price) }}</span>
                </div>
                <div class="discount-badge">
                  {{ getDiscountText() }}
                </div>
              </div>

              <!-- 库存信息 -->
              <div class="stock-section">
                <StockDisplay
                  :stock-info="stockInfo"
                  :total-stock="activity.total_stock"
                  :activity-id="activity.id"
                  :is-real-time="activity.status === 'active'"
                  :update-interval="3000"
                  :enable-web-socket="activity.status === 'active'"
                  @refresh="handleStockRefresh"
                  @stock-change="handleStockChange"
                />
              </div>

              <!-- 时间信息 -->
              <div class="time-section">
                <div class="time-item">
                  <span class="label">开始时间</span>
                  <span class="time">{{ formatDateTime(activity.start_time) }}</span>
                </div>
                <div class="time-item">
                  <span class="label">结束时间</span>
                  <span class="time">{{ formatDateTime(activity.end_time) }}</span>
                </div>
                
                <!-- 倒计时 -->
                <div v-if="activity.status === 'pending'" class="countdown-section">
                  <div class="countdown-label">距离开始还有</div>
                  <CountdownTimer :end-time="activity.start_time" @expired="handleCountdownExpired" />
                </div>
                <div v-else-if="activity.status === 'active'" class="countdown-section">
                  <div class="countdown-label">距离结束还有</div>
                  <CountdownTimer :end-time="activity.end_time" @expired="handleCountdownExpired" />
                </div>
              </div>

              <!-- 购买限制 -->
              <div class="limit-section">
                <div class="limit-item">
                  <span class="label">限购数量</span>
                  <span class="value">每人限购 {{ activity.max_per_user }} 件</span>
                </div>
                <div v-if="userParticipation" class="participation-info">
                  <span class="label">已购买</span>
                  <span class="value">{{ userParticipation.participation_count }} 件</span>
                </div>
              </div>

              <!-- 操作按钮 -->
              <div class="action-section">
                <SeckillButton
                  :activity="activity"
                  :stock-info="stockInfo"
                  :user-participation="userParticipation"
                  @participate="handleParticipate"
                  @result="handleSeckillResult"
                  size="large"
                />
              </div>
            </div>
          </el-col>
        </el-row>

        <!-- 详细信息标签页 -->
        <div class="detail-tabs">
          <el-tabs v-model="activeTab" type="border-card">
            <el-tab-pane label="商品详情" name="product">
              <div class="product-description">
                <h3>商品描述</h3>
                <p>{{ activity.product.description || '暂无商品描述' }}</p>
              </div>
            </el-tab-pane>
            
            <el-tab-pane label="活动规则" name="rules">
              <div class="activity-rules">
                <h3>活动规则</h3>
                <ul>
                  <li>每个用户限购 {{ activity.max_per_user }} 件</li>
                  <li>活动时间：{{ formatDateTime(activity.start_time) }} 至 {{ formatDateTime(activity.end_time) }}</li>
                  <li>库存有限，先到先得</li>
                  <li>支付时间限制：下单后30分钟内完成支付</li>
                  <li>活动商品不支持退换货</li>
                </ul>
              </div>
            </el-tab-pane>
            
            <el-tab-pane label="购买记录" name="records">
              <div class="purchase-records">
                <h3>最新购买记录</h3>
                <el-empty description="暂无购买记录" :image-size="80" />
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>

      <!-- 错误状态 -->
      <div v-else class="error-section">
        <el-result
          icon="error"
          title="活动不存在"
          sub-title="您访问的活动可能已被删除或不存在"
        >
          <template #extra>
            <el-button type="primary" @click="goToActivities">
              返回活动列表
            </el-button>
          </template>
        </el-result>
      </div>
    </div>

    <!-- 秒杀结果弹窗 -->
    <SeckillResult
      :activity="activity"
      :result="seckillResult"
      :visible="resultVisible"
      :type="resultType"
      @close="handleResultClose"
      @retry="handleRetry"
      @pay="handleGoToPay"
      @continue-shopping="handleContinueShopping"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Picture } from '@element-plus/icons-vue'
import { useActivityStore } from '@/stores/activity'
import { useAuth } from '@/composables/useAuth'
import { useRealTimeActivity } from '@/composables/useRealTimeActivity'
import { formatPrice, formatDateTime } from '@/utils'
import CountdownTimer from '@/components/activity/CountdownTimer.vue'
import SeckillButton from '@/components/activity/SeckillButton.vue'
import StockDisplay from '@/components/activity/StockDisplay.vue'
import SeckillResult from '@/components/activity/SeckillResult.vue'
import type { SeckillActivity, StockInfo, SeckillResult as SeckillResultType } from '@/types'

// 路由和store
const route = useRoute()
const router = useRouter()
const activityStore = useActivityStore()
const { isAuthenticated } = useAuth()

// 实时活动状态 (在获取到activityId后初始化)
let realTimeActivity: ReturnType<typeof useRealTimeActivity> | null = null

// 状态
const loading = ref(true)
const activeTab = ref('product')
const defaultImage = '/src/assets/default-product.png'
const refreshTimer = ref<number | null>(null)

// 秒杀结果相关状态
const resultVisible = ref(false)
const resultType = ref<'success' | 'failure' | 'rate_limit' | 'queue'>('success')
const seckillResult = ref<SeckillResultType | null>(null)

// 计算属性
const activity = computed(() => activityStore.currentActivity)
const stockInfo = computed(() => {
  if (!activity.value) return null
  return activityStore.getActivityStockInfo(activity.value.id)
})

const userParticipation = ref<any>(null)

// 获取状态文本和类型
const getStatusText = (status: string) => activityStore.getActivityStatusText(status)
const getStatusType = (status: string) => activityStore.getActivityStatusType(status)

// 获取库存状态类型
const getStockType = (status: string) => {
  const typeMap = {
    normal: 'success',
    low_stock: 'warning',
    out_of_stock: 'danger',
  }
  return typeMap[status as keyof typeof typeMap] || 'info'
}

// 获取已售数量和百分比
const getSoldCount = () => {
  if (!activity.value || !stockInfo.value) return 0
  return activity.value.total_stock - stockInfo.value.available_stock
}

const getSoldPercentage = () => {
  if (!activity.value) return 0
  const sold = getSoldCount()
  return Math.round((sold / activity.value.total_stock) * 100)
}

// 获取进度条颜色
const getProgressColor = () => {
  const percentage = getSoldPercentage()
  if (percentage >= 90) return '#f56c6c'
  if (percentage >= 70) return '#e6a23c'
  return '#67c23a'
}

// 获取折扣文本
const getDiscountText = () => {
  if (!activity.value) return ''
  const original = parseFloat(activity.value.original_price)
  const seckill = parseFloat(activity.value.seckill_price)
  const discount = Math.round((1 - seckill / original) * 100)
  return `${discount}折`
}

// 处理倒计时过期
const handleCountdownExpired = () => {
  // 重新获取活动信息
  fetchActivityData()
}

// 处理参与秒杀
const handleParticipate = async (activityId: number) => {
  try {
    // 先显示排队状态
    seckillResult.value = null
    resultType.value = 'queue'
    resultVisible.value = true

    const result = await activityStore.participateInSeckill(activityId)

    // 关闭排队弹窗
    resultVisible.value = false

    // 根据结果显示相应弹窗
    seckillResult.value = result

    if (result.code === 'SUCCESS') {
      resultType.value = 'success'
      // 刷新数据
      await fetchActivityData()
    } else if (result.code === 'RATE_LIMIT') {
      resultType.value = 'rate_limit'
    } else {
      resultType.value = 'failure'
    }

    resultVisible.value = true
  } catch (error: any) {
    resultVisible.value = false
    ElMessage.error(error.message || '抢购失败')
  }
}

// 处理秒杀结果
const handleSeckillResult = (result: { type: string; data?: any }) => {
  seckillResult.value = result.data
  resultType.value = result.type as any
  resultVisible.value = true
}

// 处理库存刷新
const handleStockRefresh = async () => {
  if (activity.value) {
    await activityStore.fetchStockInfo(activity.value.id)
  }
}

// 处理库存变化
const handleStockChange = (change: { old: number; new: number; diff: number }) => {
  console.log('库存变化:', change)
  // 这里可以添加库存变化的处理逻辑，比如显示通知
  if (change.diff < 0) {
    ElMessage.info(`有 ${Math.abs(change.diff)} 件商品被抢购`)
  }
}

// 处理结果弹窗关闭
const handleResultClose = () => {
  resultVisible.value = false
  seckillResult.value = null
}

// 处理重试
const handleRetry = () => {
  if (activity.value) {
    handleParticipate(activity.value.id)
  }
}

// 处理去支付
const handleGoToPay = (orderId: string) => {
  router.push(`/user/orders/${orderId}/pay`)
}

// 处理继续购物
const handleContinueShopping = () => {
  router.push('/activities')
}

// 跳转到活动列表
const goToActivities = () => {
  router.push('/activities')
}

// 获取活动数据
const fetchActivityData = async () => {
  const activityId = parseInt(route.params.id as string)
  if (!activityId) {
    loading.value = false
    return
  }

  try {
    await activityStore.fetchActivity(activityId)

    // 初始化实时活动状态监听
    if (!realTimeActivity) {
      realTimeActivity = useRealTimeActivity(activityId)
    }

    // 如果用户已登录，获取用户参与记录
    if (isAuthenticated.value) {
      userParticipation.value = await activityStore.fetchUserParticipation(activityId)
    }
  } catch (error) {
    // 错误已在store中处理
  } finally {
    loading.value = false
  }
}

// 定时刷新库存信息
const startRefreshTimer = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
  }
  
  refreshTimer.value = setInterval(async () => {
    if (activity.value && activity.value.status === 'active') {
      await activityStore.fetchStockInfo(activity.value.id)
    }
  }, 5000) // 每5秒刷新一次
}

// 组件挂载时获取数据
onMounted(() => {
  fetchActivityData().then(() => {
    if (activity.value && activity.value.status === 'active') {
      startRefreshTimer()
    }
  })
})

// 组件卸载时清理定时器
onUnmounted(() => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
  }
})
</script>

<style scoped lang="scss">
.activity-detail-view {
  min-height: 100vh;
  background: var(--el-bg-color-page);
  padding: 24px 0;

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 24px;
  }

  .breadcrumb {
    margin-bottom: 24px;
  }

  .loading-section {
    background: white;
    border-radius: 12px;
    padding: 32px;
  }

  .activity-detail {
    .product-gallery {
      .main-image {
        position: relative;
        border-radius: 12px;
        overflow: hidden;
        background: white;
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);

        .product-image {
          width: 100%;
          height: 400px;
        }

        .image-error {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 400px;
          color: var(--el-text-color-placeholder);
          background: var(--el-bg-color-page);

          .el-icon {
            font-size: 48px;
            margin-bottom: 12px;
          }
        }

        .status-overlay {
          position: absolute;
          top: 16px;
          left: 16px;
        }
      }
    }

    .product-info {
      background: white;
      border-radius: 12px;
      padding: 32px;
      box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);

      .product-title {
        margin: 0 0 12px 0;
        font-size: 24px;
        font-weight: 600;
        color: var(--el-text-color-primary);
        line-height: 1.4;
      }

      .activity-title {
        margin: 0 0 24px 0;
        font-size: 18px;
        font-weight: 500;
        color: var(--el-text-color-regular);
      }

      .price-section {
        display: flex;
        align-items: center;
        gap: 16px;
        margin-bottom: 24px;
        padding: 20px;
        background: var(--el-color-danger-light-9);
        border-radius: 8px;

        .seckill-price {
          display: flex;
          align-items: baseline;
          gap: 4px;

          .label {
            font-size: 14px;
            color: var(--el-text-color-regular);
          }

          .currency {
            font-size: 16px;
            color: var(--el-color-danger);
            font-weight: 600;
          }

          .price {
            font-size: 32px;
            color: var(--el-color-danger);
            font-weight: 700;
          }
        }

        .original-price {
          display: flex;
          align-items: baseline;
          gap: 4px;

          .label {
            font-size: 12px;
            color: var(--el-text-color-placeholder);
          }

          .price {
            font-size: 16px;
            color: var(--el-text-color-placeholder);
            text-decoration: line-through;
          }
        }

        .discount-badge {
          margin-left: auto;
          background: var(--el-color-danger);
          color: white;
          padding: 6px 12px;
          border-radius: 16px;
          font-size: 14px;
          font-weight: 600;
        }
      }

      .stock-section,
      .time-section,
      .limit-section {
        margin-bottom: 24px;
        padding-bottom: 24px;
        border-bottom: 1px solid var(--el-border-color-lighter);

        &:last-child {
          border-bottom: none;
          margin-bottom: 0;
          padding-bottom: 0;
        }
      }

      .stock-section {
        .stock-info {
          display: flex;
          align-items: center;
          gap: 12px;
          margin-bottom: 16px;

          .label {
            font-size: 14px;
            color: var(--el-text-color-regular);
            font-weight: 500;
          }
        }

        .progress-info {
          .progress-text {
            display: flex;
            justify-content: space-between;
            font-size: 12px;
            color: var(--el-text-color-regular);
            margin-bottom: 8px;
          }
        }
      }

      .time-section {
        .time-item {
          display: flex;
          justify-content: space-between;
          margin-bottom: 8px;

          .label {
            font-size: 14px;
            color: var(--el-text-color-regular);
          }

          .time {
            font-size: 14px;
            color: var(--el-text-color-primary);
          }
        }

        .countdown-section {
          margin-top: 16px;
          text-align: center;

          .countdown-label {
            font-size: 14px;
            color: var(--el-text-color-regular);
            margin-bottom: 8px;
          }
        }
      }

      .limit-section {
        .limit-item,
        .participation-info {
          display: flex;
          justify-content: space-between;
          margin-bottom: 8px;

          .label {
            font-size: 14px;
            color: var(--el-text-color-regular);
          }

          .value {
            font-size: 14px;
            color: var(--el-text-color-primary);
            font-weight: 500;
          }
        }
      }

      .action-section {
        margin-top: 32px;
      }
    }

    .detail-tabs {
      margin-top: 32px;

      .product-description,
      .activity-rules,
      .purchase-records {
        padding: 24px;

        h3 {
          margin: 0 0 16px 0;
          font-size: 18px;
          font-weight: 600;
          color: var(--el-text-color-primary);
        }

        p {
          margin: 0;
          line-height: 1.6;
          color: var(--el-text-color-regular);
        }

        ul {
          margin: 0;
          padding-left: 20px;

          li {
            margin-bottom: 8px;
            line-height: 1.6;
            color: var(--el-text-color-regular);
          }
        }
      }
    }
  }

  .error-section {
    background: white;
    border-radius: 12px;
    padding: 48px 24px;
    text-align: center;
  }
}

@media (max-width: 768px) {
  .activity-detail-view {
    padding: 16px 0;

    .container {
      padding: 0 16px;
    }

    .activity-detail {
      .product-info {
        padding: 20px;
        margin-top: 16px;

        .price-section {
          flex-direction: column;
          align-items: flex-start;
          gap: 12px;

          .discount-badge {
            margin-left: 0;
            align-self: flex-end;
          }
        }
      }
    }
  }
}
</style>
