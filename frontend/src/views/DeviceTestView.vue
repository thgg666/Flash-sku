<template>
  <div class="device-test-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>设备兼容性测试页面</h1>
      <p>此页面用于测试不同设备和屏幕尺寸下的界面显示效果</p>
    </div>

    <!-- 测试组件展示 -->
    <div class="test-sections">
      <!-- 导航测试 -->
      <section class="test-section">
        <h2>导航组件测试</h2>
        <div class="test-content">
          <AppHeader />
        </div>
      </section>

      <!-- 活动卡片测试 -->
      <section class="test-section">
        <h2>活动卡片测试</h2>
        <div class="test-content">
          <div class="activities-grid desktop-grid-3">
            <ActivityCard
              v-for="activity in mockActivities"
              :key="activity.id"
              :activity="activity"
              :stock-info="mockStockInfo"
              @participate="handleParticipate"
              @click="handleCardClick"
            />
          </div>
        </div>
      </section>

      <!-- 表单测试 -->
      <section class="test-section">
        <h2>表单组件测试</h2>
        <div class="test-content">
          <div class="form-container">
            <el-form :model="testForm" label-width="120px">
              <el-form-item label="用户名">
                <el-input v-model="testForm.username" placeholder="请输入用户名" />
              </el-form-item>
              <el-form-item label="邮箱">
                <el-input v-model="testForm.email" type="email" placeholder="请输入邮箱" />
              </el-form-item>
              <el-form-item label="密码">
                <el-input v-model="testForm.password" type="password" placeholder="请输入密码" />
              </el-form-item>
              <el-form-item label="性别">
                <el-radio-group v-model="testForm.gender">
                  <el-radio value="male">男</el-radio>
                  <el-radio value="female">女</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="兴趣爱好">
                <el-checkbox-group v-model="testForm.interests">
                  <el-checkbox value="reading">阅读</el-checkbox>
                  <el-checkbox value="sports">运动</el-checkbox>
                  <el-checkbox value="music">音乐</el-checkbox>
                  <el-checkbox value="travel">旅行</el-checkbox>
                </el-checkbox-group>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" class="desktop-hover-lift">提交</el-button>
                <el-button class="desktop-hover-lift">重置</el-button>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </section>

      <!-- 按钮测试 -->
      <section class="test-section">
        <h2>按钮组件测试</h2>
        <div class="test-content">
          <div class="button-groups">
            <div class="button-group">
              <h3>基础按钮</h3>
              <el-button class="desktop-hover-lift">默认按钮</el-button>
              <el-button type="primary" class="desktop-hover-lift">主要按钮</el-button>
              <el-button type="success" class="desktop-hover-lift">成功按钮</el-button>
              <el-button type="warning" class="desktop-hover-lift">警告按钮</el-button>
              <el-button type="danger" class="desktop-hover-lift">危险按钮</el-button>
            </div>
            
            <div class="button-group">
              <h3>不同尺寸</h3>
              <el-button size="large" class="desktop-hover-lift">大型按钮</el-button>
              <el-button class="desktop-hover-lift">默认按钮</el-button>
              <el-button size="small" class="desktop-hover-lift">小型按钮</el-button>
            </div>
            
            <div class="button-group">
              <h3>图标按钮</h3>
              <el-button type="primary" :icon="Search" class="desktop-hover-lift">搜索</el-button>
              <el-button type="success" :icon="Check" class="desktop-hover-lift">确认</el-button>
              <el-button type="danger" :icon="Delete" class="desktop-hover-lift">删除</el-button>
            </div>
          </div>
        </div>
      </section>

      <!-- 表格测试 -->
      <section class="test-section">
        <h2>表格组件测试</h2>
        <div class="test-content">
          <el-table :data="mockTableData" style="width: 100%">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="name" label="商品名称" />
            <el-table-column prop="price" label="价格" width="120" />
            <el-table-column prop="stock" label="库存" width="100" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default>
                <el-button size="small" type="primary" class="desktop-hover-lift">编辑</el-button>
                <el-button size="small" type="danger" class="desktop-hover-lift">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </section>

      <!-- 对话框测试 -->
      <section class="test-section">
        <h2>对话框测试</h2>
        <div class="test-content">
          <div class="dialog-buttons">
            <el-button @click="showDialog = true" class="desktop-hover-lift">打开对话框</el-button>
            <el-button @click="showDrawer = true" class="desktop-hover-lift">打开抽屉</el-button>
          </div>
          
          <el-dialog v-model="showDialog" title="测试对话框" width="500px">
            <p>这是一个测试对话框，用于验证在不同设备上的显示效果。</p>
            <p>对话框应该在所有设备上都能正常显示和交互。</p>
            <template #footer>
              <el-button @click="showDialog = false">取消</el-button>
              <el-button type="primary" @click="showDialog = false">确认</el-button>
            </template>
          </el-dialog>
          
          <el-drawer v-model="showDrawer" title="测试抽屉" direction="rtl">
            <p>这是一个测试抽屉，用于验证在不同设备上的显示效果。</p>
            <p>抽屉应该在移动端和桌面端都有良好的体验。</p>
          </el-drawer>
        </div>
      </section>

      <!-- 响应式网格测试 -->
      <section class="test-section">
        <h2>响应式网格测试</h2>
        <div class="test-content">
          <div class="responsive-grid">
            <div class="grid-item">网格项目 1</div>
            <div class="grid-item">网格项目 2</div>
            <div class="grid-item">网格项目 3</div>
            <div class="grid-item">网格项目 4</div>
            <div class="grid-item">网格项目 5</div>
            <div class="grid-item">网格项目 6</div>
          </div>
        </div>
      </section>
    </div>

    <!-- 设备兼容性测试器 -->
    <DeviceCompatibilityTester v-if="showTester" />
    
    <!-- 测试控制按钮 -->
    <div class="test-controls">
      <el-button
        type="primary"
        circle
        :icon="Setting"
        @click="showTester = !showTester"
        class="test-toggle-btn"
        title="切换测试器"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Search, Check, Delete, Setting } from '@element-plus/icons-vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import ActivityCard from '@/components/activity/ActivityCard.vue'
import DeviceCompatibilityTester from '@/components/common/DeviceCompatibilityTester.vue'
import type { SeckillActivity, StockInfo } from '@/types'

// 响应式数据
const showDialog = ref(false)
const showDrawer = ref(false)
const showTester = ref(false)

// 测试表单数据
const testForm = ref({
  username: '',
  email: '',
  password: '',
  gender: '',
  interests: []
})

// 模拟活动数据
const mockActivities: SeckillActivity[] = [
  {
    id: 1,
    name: '限时秒杀活动',
    product: {
      id: 1,
      name: 'iPhone 15 Pro Max',
      image_url: '/images/iphone15.jpg'
    },
    seckill_price: '8999.00',
    original_price: '9999.00',
    total_stock: 100,
    start_time: new Date(Date.now() + 3600000).toISOString(),
    end_time: new Date(Date.now() + 7200000).toISOString(),
    status: 'active'
  },
  {
    id: 2,
    name: '新品首发',
    product: {
      id: 2,
      name: 'MacBook Pro M3',
      image_url: '/images/macbook.jpg'
    },
    seckill_price: '12999.00',
    original_price: '14999.00',
    total_stock: 50,
    start_time: new Date(Date.now() - 3600000).toISOString(),
    end_time: new Date(Date.now() + 3600000).toISOString(),
    status: 'pending'
  },
  {
    id: 3,
    name: '清仓特价',
    product: {
      id: 3,
      name: 'iPad Air',
      image_url: '/images/ipad.jpg'
    },
    seckill_price: '3999.00',
    original_price: '4999.00',
    total_stock: 200,
    start_time: new Date(Date.now() - 7200000).toISOString(),
    end_time: new Date(Date.now() - 3600000).toISOString(),
    status: 'ended'
  }
]

// 模拟库存信息
const mockStockInfo: StockInfo = {
  activity_id: '1',
  available_stock: 85,
  status: 'normal',
  last_updated: new Date().toISOString()
}

// 模拟表格数据
const mockTableData = [
  { id: 1, name: 'iPhone 15', price: '¥8999', stock: 100, status: '有库存' },
  { id: 2, name: 'MacBook Pro', price: '¥12999', stock: 50, status: '有库存' },
  { id: 3, name: 'iPad Air', price: '¥3999', stock: 0, status: '缺货' },
  { id: 4, name: 'Apple Watch', price: '¥2999', stock: 200, status: '有库存' },
  { id: 5, name: 'AirPods Pro', price: '¥1999', stock: 150, status: '有库存' }
]

// 事件处理
const handleParticipate = (activityId: number) => {
  console.log('参与活动:', activityId)
}

const handleCardClick = (activity: SeckillActivity) => {
  console.log('点击卡片:', activity)
}

const getStatusType = (status: string) => {
  return status === '有库存' ? 'success' : 'danger'
}
</script>

<style scoped lang="scss">
.device-test-view {
  min-height: 100vh;
  background: var(--el-bg-color-page);
  padding: 20px;

  .page-header {
    text-align: center;
    margin-bottom: 40px;

    h1 {
      font-size: 32px;
      margin-bottom: 16px;
      color: var(--el-text-color-primary);
    }

    p {
      font-size: 16px;
      color: var(--el-text-color-regular);
    }
  }

  .test-sections {
    max-width: 1200px;
    margin: 0 auto;

    .test-section {
      background: white;
      border-radius: 12px;
      padding: 24px;
      margin-bottom: 24px;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

      h2 {
        margin: 0 0 20px 0;
        font-size: 20px;
        color: var(--el-text-color-primary);
        border-bottom: 2px solid var(--el-color-primary);
        padding-bottom: 8px;
      }

      .test-content {
        .form-container {
          max-width: 600px;
        }

        .button-groups {
          .button-group {
            margin-bottom: 20px;

            h3 {
              margin: 0 0 12px 0;
              font-size: 16px;
              color: var(--el-text-color-regular);
            }

            .el-button {
              margin-right: 12px;
              margin-bottom: 8px;
            }
          }
        }

        .dialog-buttons {
          .el-button {
            margin-right: 12px;
          }
        }

        .responsive-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
          gap: 16px;

          .grid-item {
            background: var(--el-color-primary-light-9);
            padding: 20px;
            border-radius: 8px;
            text-align: center;
            color: var(--el-color-primary);
            font-weight: 500;
          }
        }
      }
    }
  }

  .test-controls {
    position: fixed;
    bottom: 20px;
    left: 20px;
    z-index: 1000;

    .test-toggle-btn {
      width: 56px;
      height: 56px;
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
    }
  }
}

// 移动端优化
@media (max-width: 768px) {
  .device-test-view {
    padding: 16px;

    .page-header {
      h1 {
        font-size: 24px;
      }

      p {
        font-size: 14px;
      }
    }

    .test-sections {
      .test-section {
        padding: 16px;
        margin-bottom: 16px;

        h2 {
          font-size: 18px;
        }
      }
    }
  }
}
</style>
