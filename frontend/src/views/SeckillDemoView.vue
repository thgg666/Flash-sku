<template>
  <div class="seckill-demo-view">
    <div class="container">
      <div class="demo-header">
        <h1>秒杀组件演示</h1>
        <p>展示不同状态下的秒杀按钮和倒计时组件</p>
      </div>

      <!-- 倒计时组件演示 -->
      <div class="demo-section">
        <h2>倒计时组件</h2>
        <el-row :gutter="24">
          <el-col :xs="24" :sm="12" :md="8">
            <el-card>
              <template #header>
                <span>正常倒计时</span>
              </template>
              <div class="demo-item">
                <CountdownTimer 
                  :end-time="normalEndTime"
                  @expired="handleExpired"
                  @urgent="handleUrgent"
                  @tick="handleTick"
                />
              </div>
            </el-card>
          </el-col>
          
          <el-col :xs="24" :sm="12" :md="8">
            <el-card>
              <template #header>
                <span>紧急倒计时（5分钟内）</span>
              </template>
              <div class="demo-item">
                <CountdownTimer 
                  :end-time="urgentEndTime"
                  :urgent-threshold="5"
                  @expired="handleExpired"
                  @urgent="handleUrgent"
                />
              </div>
            </el-card>
          </el-col>
          
          <el-col :xs="24" :sm="12" :md="8">
            <el-card>
              <template #header>
                <span>已过期倒计时</span>
              </template>
              <div class="demo-item">
                <CountdownTimer 
                  :end-time="expiredEndTime"
                  expired-text="活动已结束"
                />
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 秒杀按钮演示 -->
      <div class="demo-section">
        <h2>秒杀按钮状态</h2>
        <el-row :gutter="24">
          <!-- 未登录状态 -->
          <el-col :xs="24" :sm="12" :md="8">
            <el-card>
              <template #header>
                <span>未登录状态</span>
              </template>
              <div class="demo-item">
                <SeckillButton
                  :activity="demoActivity"
                  :stock-info="normalStock"
                  @participate="handleParticipate"
                  size="large"
                  block
                />
              </div>
            </el-card>
          </el-col>

          <!-- 即将开始 -->
          <el-col :xs="24" :sm="12" :md="8">
            <el-card>
              <template #header>
                <span>即将开始</span>
              </template>
              <div class="demo-item">
                <SeckillButton
                  :activity="pendingActivity"
                  :stock-info="normalStock"
                  @participate="handleParticipate"
                  size="large"
                  block
                />
              </div>
            </el-card>
          </el-col>

          <!-- 正常可购买 -->
          <el-col :xs="24" :sm="12" :md="8">
            <el-card>
              <template #header>
                <span>正常可购买</span>
              </template>
              <div class="demo-item">
                <SeckillButton
                  :activity="activeActivity"
                  :stock-info="normalStock"
                  @participate="handleParticipate"
                  size="large"
                  block
                />
              </div>
            </el-card>
          </el-col>

          <!-- 库存不足 -->
          <el-col :xs="24" :sm="12" :md="8">
            <el-card>
              <template #header>
                <span>库存不足</span>
              </template>
              <div class="demo-item">
                <SeckillButton
                  :activity="activeActivity"
                  :stock-info="lowStock"
                  @participate="handleParticipate"
                  size="large"
                  block
                />
              </div>
            </el-card>
          </el-col>

          <!-- 已售罄 -->
          <el-col :xs="24" :sm="12" :md="8">
            <el-card>
              <template #header>
                <span>已售罄</span>
              </template>
              <div class="demo-item">
                <SeckillButton
                  :activity="activeActivity"
                  :stock-info="soldOutStock"
                  @participate="handleParticipate"
                  size="large"
                  block
                />
              </div>
            </el-card>
          </el-col>

          <!-- 已结束 -->
          <el-col :xs="24" :sm="12" :md="8">
            <el-card>
              <template #header>
                <span>活动已结束</span>
              </template>
              <div class="demo-item">
                <SeckillButton
                  :activity="endedActivity"
                  :stock-info="normalStock"
                  @participate="handleParticipate"
                  size="large"
                  block
                />
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 综合演示 -->
      <div class="demo-section">
        <h2>综合演示</h2>
        <el-card>
          <div class="comprehensive-demo">
            <div class="demo-product">
              <el-image
                src="/src/assets/demo-product.jpg"
                alt="演示商品"
                fit="cover"
                class="product-image"
              >
                <template #error>
                  <div class="image-error">
                    <el-icon><Picture /></el-icon>
                  </div>
                </template>
              </el-image>
            </div>
            
            <div class="demo-info">
              <h3>iPhone 15 Pro Max 256GB</h3>
              <p>秒杀专场 - 限时抢购</p>
              
              <div class="price-info">
                <span class="seckill-price">¥8999</span>
                <span class="original-price">¥9999</span>
                <span class="discount">9折</span>
              </div>
              
              <div class="countdown-section">
                <span class="countdown-label">距离结束还有</span>
                <CountdownTimer :end-time="normalEndTime" />
              </div>
              
              <div class="action-section">
                <SeckillButton
                  :activity="activeActivity"
                  :stock-info="normalStock"
                  @participate="handleParticipate"
                  size="large"
                  block
                />
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture } from '@element-plus/icons-vue'
import CountdownTimer from '@/components/activity/CountdownTimer.vue'
import SeckillButton from '@/components/activity/SeckillButton.vue'
import type { SeckillActivity, StockInfo } from '@/types'

// 演示数据
const now = new Date()

// 倒计时时间
const normalEndTime = ref(new Date(now.getTime() + 2 * 60 * 60 * 1000)) // 2小时后
const urgentEndTime = ref(new Date(now.getTime() + 3 * 60 * 1000)) // 3分钟后
const expiredEndTime = ref(new Date(now.getTime() - 60 * 1000)) // 1分钟前

// 演示活动数据
const demoActivity: SeckillActivity = {
  id: 1,
  product_id: 1,
  name: 'iPhone 15 Pro Max 秒杀',
  status: 'active',
  start_time: new Date(now.getTime() - 60 * 60 * 1000).toISOString(),
  end_time: normalEndTime.value.toISOString(),
  seckill_price: '8999.00',
  original_price: '9999.00',
  total_stock: 100,
  available_stock: 50,
  max_per_user: 1,
  product: {
    id: 1,
    name: 'iPhone 15 Pro Max 256GB',
    description: '最新款iPhone，性能强劲',
    image_url: '/src/assets/demo-product.jpg',
    created_at: now.toISOString(),
  },
  created_at: now.toISOString(),
}

const pendingActivity: SeckillActivity = {
  ...demoActivity,
  status: 'pending',
  start_time: new Date(now.getTime() + 60 * 60 * 1000).toISOString(),
}

const activeActivity: SeckillActivity = {
  ...demoActivity,
  status: 'active',
}

const endedActivity: SeckillActivity = {
  ...demoActivity,
  status: 'ended',
  end_time: new Date(now.getTime() - 60 * 60 * 1000).toISOString(),
}

// 库存信息
const normalStock: StockInfo = {
  activity_id: 1,
  available_stock: 50,
  total_stock: 100,
  status: 'normal',
  activity_status: 'active',
  last_updated: now.toISOString(),
}

const lowStock: StockInfo = {
  ...normalStock,
  available_stock: 5,
  status: 'low_stock',
}

const soldOutStock: StockInfo = {
  ...normalStock,
  available_stock: 0,
  status: 'out_of_stock',
}

// 事件处理
const handleExpired = () => {
  ElMessage.info('倒计时已结束')
}

const handleUrgent = () => {
  ElMessage.warning('时间紧急！')
}

const handleTick = (countdown: any) => {
  console.log('倒计时更新:', countdown)
}

const handleParticipate = (activityId: number) => {
  ElMessage.success(`参与活动 ${activityId}`)
}
</script>

<style scoped lang="scss">
.seckill-demo-view {
  min-height: 100vh;
  background: var(--el-bg-color-page);
  padding: 24px 0;

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 24px;
  }

  .demo-header {
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

  .demo-section {
    margin-bottom: 48px;

    h2 {
      margin: 0 0 24px 0;
      color: var(--el-text-color-primary);
      font-size: 24px;
      font-weight: 600;
    }

    .demo-item {
      padding: 20px;
      text-align: center;
    }
  }

  .comprehensive-demo {
    display: flex;
    gap: 32px;
    align-items: center;

    .demo-product {
      flex-shrink: 0;

      .product-image {
        width: 200px;
        height: 200px;
        border-radius: 8px;
      }

      .image-error {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 200px;
        height: 200px;
        background: var(--el-bg-color-page);
        border-radius: 8px;

        .el-icon {
          font-size: 48px;
          color: var(--el-text-color-placeholder);
        }
      }
    }

    .demo-info {
      flex: 1;

      h3 {
        margin: 0 0 8px 0;
        font-size: 24px;
        font-weight: 600;
        color: var(--el-text-color-primary);
      }

      p {
        margin: 0 0 16px 0;
        color: var(--el-text-color-regular);
      }

      .price-info {
        display: flex;
        align-items: baseline;
        gap: 12px;
        margin-bottom: 16px;

        .seckill-price {
          font-size: 28px;
          font-weight: 700;
          color: var(--el-color-danger);
        }

        .original-price {
          font-size: 16px;
          color: var(--el-text-color-placeholder);
          text-decoration: line-through;
        }

        .discount {
          background: var(--el-color-danger);
          color: white;
          padding: 4px 8px;
          border-radius: 4px;
          font-size: 12px;
        }
      }

      .countdown-section {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 24px;

        .countdown-label {
          font-size: 14px;
          color: var(--el-text-color-regular);
        }
      }

      .action-section {
        max-width: 300px;
      }
    }
  }
}

@media (max-width: 768px) {
  .seckill-demo-view {
    padding: 16px 0;

    .container {
      padding: 0 16px;
    }

    .comprehensive-demo {
      flex-direction: column;
      text-align: center;

      .demo-info {
        .action-section {
          max-width: none;
        }
      }
    }
  }
}
</style>
