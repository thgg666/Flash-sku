<template>
  <div class="login-view">
    <div class="login-container">
      <div class="login-header">
        <h2>用户登录</h2>
        <p>欢迎回到Flash Sku</p>
      </div>

      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        label-width="0"
        size="large"
        @submit.prevent="handleLogin"
      >
        <!-- 用户名/邮箱 -->
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名或邮箱"
            :disabled="loading"
            @keyup.enter="handleLogin"
          >
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 密码 -->
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            :disabled="loading"
            show-password
            @keyup.enter="handleLogin"
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 验证码 (登录失败次数过多时显示) -->
        <el-form-item v-if="showCaptcha" prop="captcha">
          <CaptchaInput
            ref="captchaRef"
            v-model="loginForm.captcha"
            :disabled="loading"
            @captcha-loaded="handleCaptchaLoaded"
            @submit="handleLogin"
          />
        </el-form-item>

        <!-- 记住登录和忘记密码 -->
        <div class="login-options">
          <el-checkbox v-model="loginForm.remember" :disabled="loading">
            记住登录
          </el-checkbox>
          <router-link to="/auth/forgot-password" class="forgot-link">
            忘记密码？
          </router-link>
        </div>

        <!-- 登录按钮 -->
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleLogin"
            class="login-button"
          >
            {{ loading ? '登录中...' : '立即登录' }}
          </el-button>
        </el-form-item>

        <!-- 注册链接 -->
        <div class="register-link">
          还没有账户？
          <router-link to="/auth/register" class="link">立即注册</router-link>
        </div>
      </el-form>

      <!-- 第三方登录 (可选) -->
      <div class="social-login">
        <el-divider>
          <span class="divider-text">其他登录方式</span>
        </el-divider>
        <div class="social-buttons">
          <el-button circle @click="handleSocialLogin('github')">
            <el-icon><svg viewBox="0 0 1024 1024" width="16" height="16">
              <path d="M512 12.64c-282.752 0-512 229.216-512 512 0 226.208 146.688 418.144 350.08 485.824 25.6 4.736 35.008-11.104 35.008-24.64 0-12.192-0.48-52.544-0.704-95.328-142.464 30.976-172.512-60.416-172.512-60.416-23.296-59.168-56.832-74.912-56.832-74.912-46.464-31.776 3.52-31.136 3.52-31.136 51.392 3.616 78.464 52.768 78.464 52.768 45.664 78.272 119.776 55.648 148.992 42.56 4.576-33.088 17.856-55.68 32.512-68.48-113.728-12.928-233.216-56.864-233.216-253.024 0-55.904 19.936-101.568 52.672-137.408-5.312-12.896-22.848-64.96 4.96-135.488 0 0 42.88-13.76 140.8 52.48 40.832-11.36 84.64-17.024 128.16-17.248 43.488 0.192 87.328 5.888 128.256 17.248 97.728-66.24 140.64-52.48 140.64-52.48 27.872 70.528 10.336 122.592 5.024 135.488 32.832 35.84 52.608 81.504 52.608 137.408 0 196.64-119.776 239.936-233.856 252.64 18.368 15.904 34.72 47.04 34.72 94.816 0 68.512-0.608 123.648-0.608 140.512 0 13.632 9.216 29.6 35.168 24.576C877.472 942.08 1024 750.208 1024 524.64c0-282.784-229.248-512-512-512z" fill="currentColor"/>
            </svg></el-icon>
          </el-button>
          <el-button circle @click="handleSocialLogin('google')">
            <el-icon><svg viewBox="0 0 1024 1024" width="16" height="16">
              <path d="M881 442.4H519.7v148.5h206.4c-8.9 48-35.9 88.6-76.6 115.8-34.4 23-78.3 36.6-129.9 36.6-99.9 0-184.4-67.5-214.6-158.2-7.6-23-12-47.6-12-72.9s4.4-49.9 12-72.9c30.3-90.6 114.8-158.1 214.6-158.1 56.3 0 106.8 19.4 146.6 57.4l110-110.1c-66.5-62-153.2-100-256.6-100-149.9 0-279.6 86.8-342.7 213.1C59.2 295.6 51.4 357.9 51.4 512s7.8 216.4 40.8 299.9C155.4 937.2 285.1 1024 435 1024c117.8 0 218.2-39 291.2-104 67.5-60.2 105.8-149.5 105.8-265.2 0-27.3-2.4-53.8-6.9-79.5z" fill="currentColor"/>
            </svg></el-icon>
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { isValidEmail } from '@/utils'
import CaptchaInput from '@/components/auth/CaptchaInput.vue'
import type { LoginRequest } from '@/types'

// 路由和store
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// 表单引用
const loginFormRef = ref<FormInstance>()
const captchaRef = ref()

// 状态
const loading = ref(false)
const loginFailCount = ref(0)
const captchaKey = ref('')

// 表单数据
const loginForm = reactive<LoginRequest & { remember: boolean }>({
  username: '',
  password: '',
  captcha: '',
  remember: false,
})

// 计算属性
const showCaptcha = computed(() => loginFailCount.value >= 3)

// 表单验证规则
const loginRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名或邮箱', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6个字符', trigger: 'blur' },
  ],
  captcha: [
    { 
      required: showCaptcha.value, 
      message: '请输入验证码', 
      trigger: 'blur' 
    },
  ],
}

// 处理验证码加载
const handleCaptchaLoaded = (key: string) => {
  captchaKey.value = key
}

// 处理登录
const handleLogin = async () => {
  if (!loginFormRef.value) return

  try {
    const valid = await loginFormRef.value.validate()
    if (!valid) return
  } catch (error) {
    return
  }

  loading.value = true
  try {
    const result = await authStore.login({
      username: loginForm.username,
      password: loginForm.password,
      captcha: loginForm.captcha,
    })

    if (result.success) {
      // 登录成功，重置失败计数
      loginFailCount.value = 0
      
      // 如果选择记住登录，设置更长的token过期时间
      if (loginForm.remember) {
        // 这里可以调用API设置更长的token过期时间
        console.log('用户选择记住登录')
      }

      ElMessage.success('登录成功')
      
      // 跳转到原来要访问的页面或首页
      const redirect = route.query.redirect as string
      router.push(redirect || '/')
    } else {
      // 登录失败，增加失败计数
      loginFailCount.value++
      
      // 如果显示验证码，刷新验证码
      if (showCaptcha.value) {
        captchaRef.value?.refreshCaptcha()
        loginForm.captcha = ''
      }
    }
  } catch (error) {
    loginFailCount.value++
    
    // 刷新验证码
    if (showCaptcha.value) {
      captchaRef.value?.refreshCaptcha()
      loginForm.captcha = ''
    }
  } finally {
    loading.value = false
  }
}

// 处理第三方登录
const handleSocialLogin = (provider: string) => {
  ElMessage.info(`${provider} 登录功能开发中...`)
  // 这里可以实现第三方登录逻辑
}
</script>

<style scoped lang="scss">
.login-view {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;

  .login-container {
    width: 100%;
    max-width: 400px;
    background: white;
    border-radius: 12px;
    padding: 40px 32px;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);

    .login-header {
      text-align: center;
      margin-bottom: 32px;

      h2 {
        margin: 0 0 8px 0;
        color: var(--el-text-color-primary);
        font-size: 28px;
        font-weight: 600;
      }

      p {
        margin: 0;
        color: var(--el-text-color-regular);
        font-size: 14px;
      }
    }

    .el-form-item {
      margin-bottom: 20px;
    }

    .login-options {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 24px;

      .forgot-link {
        color: var(--el-color-primary);
        text-decoration: none;
        font-size: 14px;

        &:hover {
          text-decoration: underline;
        }
      }
    }

    .login-button {
      width: 100%;
      height: 48px;
      font-size: 16px;
      font-weight: 600;
    }

    .register-link {
      text-align: center;
      margin-top: 24px;
      color: var(--el-text-color-regular);
      font-size: 14px;

      .link {
        color: var(--el-color-primary);
        text-decoration: none;
        font-weight: 500;

        &:hover {
          text-decoration: underline;
        }
      }
    }

    .social-login {
      margin-top: 32px;

      .divider-text {
        color: var(--el-text-color-placeholder);
        font-size: 12px;
      }

      .social-buttons {
        display: flex;
        justify-content: center;
        gap: 16px;
        margin-top: 16px;

        .el-button {
          width: 40px;
          height: 40px;
          border-color: var(--el-border-color);
          color: var(--el-text-color-regular);

          &:hover {
            border-color: var(--el-color-primary);
            color: var(--el-color-primary);
          }
        }
      }
    }
  }
}

@media (max-width: 480px) {
  .login-view {
    padding: 12px;

    .login-container {
      padding: 24px 20px;
    }
  }
}
</style>
