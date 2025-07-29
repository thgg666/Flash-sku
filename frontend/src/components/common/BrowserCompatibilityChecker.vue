<template>
  <div class="browser-compatibility-checker">
    <el-card class="compatibility-card">
      <template #header>
        <div class="card-header">
          <h3>浏览器兼容性检测</h3>
          <el-button @click="runTest" :loading="testing" type="primary" size="small">
            重新检测
          </el-button>
        </div>
      </template>

      <!-- 浏览器信息 -->
      <div class="browser-info">
        <h4>浏览器信息</h4>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="浏览器">
            {{ browserInfo.name }} {{ browserInfo.version }}
          </el-descriptions-item>
          <el-descriptions-item label="引擎">
            {{ browserInfo.engine }}
          </el-descriptions-item>
          <el-descriptions-item label="平台">
            {{ browserInfo.platform }}
          </el-descriptions-item>
          <el-descriptions-item label="设备类型">
            {{ browserInfo.mobile ? '移动设备' : '桌面设备' }}
          </el-descriptions-item>
          <el-descriptions-item label="兼容性" :span="2">
            <el-tag :type="browserInfo.supported ? 'success' : 'danger'">
              {{ browserInfo.supported ? '完全支持' : '部分支持' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- 功能支持检测 -->
      <div class="features-check">
        <h4>功能支持检测</h4>
        
        <!-- 核心功能 -->
        <div class="feature-category">
          <h5>核心功能</h5>
          <el-row :gutter="16">
            <el-col :span="12">
              <FeatureItem 
                name="WebSocket" 
                :supported="browserInfo.features.webSocket"
                description="实时通信功能"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="本地存储" 
                :supported="browserInfo.features.localStorage"
                description="数据持久化"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="Fetch API" 
                :supported="browserInfo.features.js.fetch"
                description="网络请求"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="Service Worker" 
                :supported="browserInfo.features.serviceWorker"
                description="离线缓存"
              />
            </el-col>
          </el-row>
        </div>

        <!-- CSS功能 -->
        <div class="feature-category">
          <h5>CSS功能</h5>
          <el-row :gutter="16">
            <el-col :span="12">
              <FeatureItem 
                name="Flexbox" 
                :supported="browserInfo.features.css.flexbox"
                description="弹性布局"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="Grid" 
                :supported="browserInfo.features.css.grid"
                description="网格布局"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="CSS变量" 
                :supported="browserInfo.features.css.customProperties"
                description="自定义属性"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="CSS动画" 
                :supported="browserInfo.features.css.animations"
                description="动画效果"
              />
            </el-col>
          </el-row>
        </div>

        <!-- JavaScript功能 -->
        <div class="feature-category">
          <h5>JavaScript功能</h5>
          <el-row :gutter="16">
            <el-col :span="12">
              <FeatureItem 
                name="ES6语法" 
                :supported="browserInfo.features.js.es6"
                description="现代JavaScript"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="Async/Await" 
                :supported="browserInfo.features.js.asyncAwait"
                description="异步编程"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="ES模块" 
                :supported="browserInfo.features.js.modules"
                description="模块化"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="Promise" 
                :supported="browserInfo.features.js.promises"
                description="异步处理"
              />
            </el-col>
          </el-row>
        </div>

        <!-- 高级功能 -->
        <div class="feature-category">
          <h5>高级功能</h5>
          <el-row :gutter="16">
            <el-col :span="12">
              <FeatureItem 
                name="推送通知" 
                :supported="browserInfo.features.pushNotifications"
                description="消息推送"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="地理位置" 
                :supported="browserInfo.features.geolocation"
                description="位置服务"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="剪贴板" 
                :supported="browserInfo.features.clipboard"
                description="复制粘贴"
              />
            </el-col>
            <el-col :span="12">
              <FeatureItem 
                name="分享API" 
                :supported="browserInfo.features.share"
                description="原生分享"
              />
            </el-col>
          </el-row>
        </div>
      </div>

      <!-- 建议 -->
      <div v-if="!browserInfo.supported" class="recommendations">
        <h4>兼容性建议</h4>
        <el-alert
          title="浏览器兼容性提醒"
          type="warning"
          :description="getRecommendationText()"
          show-icon
          :closable="false"
        />
        
        <div class="browser-links">
          <h5>推荐浏览器下载</h5>
          <el-row :gutter="16">
            <el-col :span="6">
              <el-button @click="openBrowserDownload('chrome')" class="browser-btn">
                Chrome
              </el-button>
            </el-col>
            <el-col :span="6">
              <el-button @click="openBrowserDownload('firefox')" class="browser-btn">
                Firefox
              </el-button>
            </el-col>
            <el-col :span="6">
              <el-button @click="openBrowserDownload('edge')" class="browser-btn">
                Edge
              </el-button>
            </el-col>
            <el-col :span="6">
              <el-button @click="openBrowserDownload('safari')" class="browser-btn">
                Safari
              </el-button>
            </el-col>
          </el-row>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineComponent } from 'vue'
import { Check, Close } from '@element-plus/icons-vue'
import { detectBrowser, getUnsupportedFeatures, type BrowserInfo } from '@/utils/browserCompatibility'

// 功能项组件
const FeatureItem = defineComponent({
  name: 'FeatureItem',
  props: {
    name: String,
    supported: Boolean,
    description: String
  },
  components: {
    Check,
    Close
  },
  template: `
    <div class="feature-item">
      <div class="feature-name">
        <el-icon :class="supported ? 'success' : 'error'">
          <Check v-if="supported" />
          <Close v-else />
        </el-icon>
        {{ name }}
      </div>
      <div class="feature-desc">{{ description }}</div>
    </div>
  `
})

// 状态
const testing = ref(false)
const browserInfo = ref<BrowserInfo>({
  name: 'Unknown',
  version: 'Unknown',
  engine: 'Unknown',
  platform: 'Unknown',
  mobile: false,
  supported: false,
  features: {} as any
})

// 运行检测
const runTest = async () => {
  testing.value = true
  try {
    // 模拟检测过程
    await new Promise(resolve => setTimeout(resolve, 1000))
    browserInfo.value = detectBrowser()
  } finally {
    testing.value = false
  }
}

// 获取建议文本
const getRecommendationText = () => {
  const unsupported = getUnsupportedFeatures(browserInfo.value.features)
  return `您的浏览器不支持以下功能：${unsupported.join('、')}。为了获得最佳体验，建议升级到最新版本的现代浏览器。`
}

// 打开浏览器下载页面
const openBrowserDownload = (browser: string) => {
  const urls = {
    chrome: 'https://www.google.com/chrome/',
    firefox: 'https://www.mozilla.org/firefox/',
    edge: 'https://www.microsoft.com/edge',
    safari: 'https://www.apple.com/safari/'
  }
  
  window.open(urls[browser as keyof typeof urls], '_blank')
}

// 组件挂载时运行检测
onMounted(() => {
  runTest()
})
</script>

<style scoped lang="scss">
@import "@/styles/variables.scss";

.browser-compatibility-checker {
  .compatibility-card {
    max-width: 800px;
    margin: 0 auto;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    h3 {
      margin: 0;
      color: $text-color-primary;
    }
  }

  .browser-info {
    margin-bottom: 24px;

    h4 {
      margin-bottom: 16px;
      color: $text-color-primary;
    }
  }

  .features-check {
    .feature-category {
      margin-bottom: 24px;

      h5 {
        margin-bottom: 12px;
        color: $text-color-regular;
        font-size: 14px;
        font-weight: 600;
      }
    }

    .feature-item {
      padding: 8px 0;
      border-bottom: 1px solid $border-color-lighter;

      .feature-name {
        display: flex;
        align-items: center;
        font-weight: 500;
        margin-bottom: 4px;

        .el-icon {
          margin-right: 8px;
          
          &.success {
            color: $success-color;
          }
          
          &.error {
            color: $danger-color;
          }
        }
      }

      .feature-desc {
        font-size: 12px;
        color: $text-color-secondary;
        margin-left: 24px;
      }
    }
  }

  .recommendations {
    margin-top: 24px;

    h4 {
      margin-bottom: 16px;
      color: $text-color-primary;
    }

    .browser-links {
      margin-top: 16px;

      h5 {
        margin-bottom: 12px;
        color: $text-color-regular;
        font-size: 14px;
      }

      .browser-btn {
        width: 100%;
      }
    }
  }
}
</style>
