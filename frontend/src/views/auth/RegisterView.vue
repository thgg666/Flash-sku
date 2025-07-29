<template>
  <div class="register-view">
    <div class="register-container">
      <div class="register-header">
        <h2>用户注册</h2>
        <p>创建您的Flash Sku账户</p>
      </div>

      <el-form
        ref="registerFormRef"
        :model="registerForm"
        :rules="registerRules"
        label-width="0"
        size="large"
        @submit.prevent="handleRegister"
      >
        <!-- 用户名 -->
        <el-form-item prop="username">
          <el-input
            v-model="registerForm.username"
            placeholder="请输入用户名"
            :disabled="loading"
            @blur="checkUsernameAvailable"
          >
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
            <template #suffix>
              <el-icon v-if="usernameChecking" class="is-loading"><Loading /></el-icon>
              <el-icon v-else-if="usernameAvailable === true" class="success-icon"><Check /></el-icon>
              <el-icon v-else-if="usernameAvailable === false" class="error-icon"><Close /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 邮箱 -->
        <el-form-item prop="email">
          <el-input
            v-model="registerForm.email"
            type="email"
            placeholder="请输入邮箱地址"
            :disabled="loading"
            @blur="checkEmailAvailable"
          >
            <template #prefix>
              <el-icon><Message /></el-icon>
            </template>
            <template #suffix>
              <el-icon v-if="emailChecking" class="is-loading"><Loading /></el-icon>
              <el-icon v-else-if="emailAvailable === true" class="success-icon"><Check /></el-icon>
              <el-icon v-else-if="emailAvailable === false" class="error-icon"><Close /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 密码 -->
        <el-form-item prop="password">
          <PasswordStrength
            v-model="registerForm.password"
            placeholder="请输入密码"
            :disabled="loading"
            @valid-change="handlePasswordValidChange"
          />
        </el-form-item>

        <!-- 确认密码 -->
        <el-form-item prop="password_confirm">
          <el-input
            v-model="registerForm.password_confirm"
            type="password"
            placeholder="请再次输入密码"
            :disabled="loading"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 验证码 -->
        <el-form-item prop="captcha">
          <CaptchaInput
            ref="captchaRef"
            v-model="registerForm.captcha"
            :disabled="loading"
            @captcha-loaded="handleCaptchaLoaded"
            @submit="handleRegister"
          />
        </el-form-item>

        <!-- 用户协议 -->
        <el-form-item prop="agreement">
          <el-checkbox v-model="registerForm.agreement" :disabled="loading">
            我已阅读并同意
            <el-link type="primary" @click="showUserAgreement">《用户协议》</el-link>
            和
            <el-link type="primary" @click="showPrivacyPolicy">《隐私政策》</el-link>
          </el-checkbox>
        </el-form-item>

        <!-- 注册按钮 -->
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            :disabled="!canSubmit"
            @click="handleRegister"
            class="register-button"
          >
            {{ loading ? '注册中...' : '立即注册' }}
          </el-button>
        </el-form-item>

        <!-- 登录链接 -->
        <div class="login-link">
          已有账户？
          <router-link to="/auth/login" class="link">立即登录</router-link>
        </div>
      </el-form>
    </div>

    <!-- 用户协议对话框 -->
    <el-dialog v-model="agreementVisible" title="用户协议" width="60%">
      <div class="agreement-content">
        <p>欢迎使用Flash Sku秒杀系统！</p>
        <p>在使用我们的服务之前，请仔细阅读以下用户协议...</p>
        <!-- 这里可以添加完整的用户协议内容 -->
      </div>
      <template #footer>
        <el-button @click="agreementVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 隐私政策对话框 -->
    <el-dialog v-model="privacyVisible" title="隐私政策" width="60%">
      <div class="privacy-content">
        <p>我们重视您的隐私保护...</p>
        <!-- 这里可以添加完整的隐私政策内容 -->
      </div>
      <template #footer>
        <el-button @click="privacyVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { User, Message, Lock, Check, Close, Loading } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { isValidEmail } from '@/utils'
import PasswordStrength from '@/components/auth/PasswordStrength.vue'
import CaptchaInput from '@/components/auth/CaptchaInput.vue'
import type { RegisterRequest } from '@/types'

// 路由和store
const router = useRouter()
const authStore = useAuthStore()

// 表单引用
const registerFormRef = ref<FormInstance>()
const captchaRef = ref()

// 状态
const loading = ref(false)
const passwordValid = ref(false)
const usernameChecking = ref(false)
const emailChecking = ref(false)
const usernameAvailable = ref<boolean | null>(null)
const emailAvailable = ref<boolean | null>(null)
const captchaKey = ref('')
const agreementVisible = ref(false)
const privacyVisible = ref(false)

// 表单数据
const registerForm = reactive<RegisterRequest & { agreement: boolean }>({
  username: '',
  email: '',
  password: '',
  password_confirm: '',
  captcha: '',
  agreement: false,
})

// 表单验证规则
const registerRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度为3-20个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { validator: (rule, value, callback) => {
      if (!isValidEmail(value)) {
        callback(new Error('请输入有效的邮箱地址'))
      } else {
        callback()
      }
    }, trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码长度至少8个字符', trigger: 'blur' },
  ],
  password_confirm: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    { validator: (rule, value, callback) => {
      if (value !== registerForm.password) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }, trigger: 'blur' },
  ],
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { min: 4, max: 6, message: '验证码长度为4-6位', trigger: 'blur' },
  ],
  agreement: [
    { validator: (rule, value, callback) => {
      if (!value) {
        callback(new Error('请阅读并同意用户协议'))
      } else {
        callback()
      }
    }, trigger: 'change' },
  ],
}

// 计算属性
const canSubmit = computed(() => {
  return registerForm.username &&
         registerForm.email &&
         registerForm.password &&
         registerForm.password_confirm &&
         registerForm.captcha &&
         registerForm.agreement &&
         passwordValid.value &&
         usernameAvailable.value === true &&
         emailAvailable.value === true &&
         !loading.value
})

// 检查用户名可用性
const checkUsernameAvailable = async () => {
  if (!registerForm.username || registerForm.username.length < 3) {
    usernameAvailable.value = null
    return
  }

  usernameChecking.value = true
  try {
    const available = await authStore.checkUsername(registerForm.username)
    usernameAvailable.value = available
    if (!available) {
      ElMessage.warning('用户名已被使用')
    }
  } catch (error) {
    usernameAvailable.value = null
  } finally {
    usernameChecking.value = false
  }
}

// 检查邮箱可用性
const checkEmailAvailable = async () => {
  if (!registerForm.email || !isValidEmail(registerForm.email)) {
    emailAvailable.value = null
    return
  }

  emailChecking.value = true
  try {
    const available = await authStore.checkEmail(registerForm.email)
    emailAvailable.value = available
    if (!available) {
      ElMessage.warning('邮箱已被注册')
    }
  } catch (error) {
    emailAvailable.value = null
  } finally {
    emailChecking.value = false
  }
}

// 处理密码有效性变化
const handlePasswordValidChange = (isValid: boolean) => {
  passwordValid.value = isValid
}

// 处理验证码加载
const handleCaptchaLoaded = (key: string) => {
  captchaKey.value = key
}

// 处理注册
const handleRegister = async () => {
  if (!registerFormRef.value) return

  try {
    const valid = await registerFormRef.value.validate()
    if (!valid) return
  } catch (error) {
    return
  }

  loading.value = true
  try {
    const result = await authStore.register({
      username: registerForm.username,
      email: registerForm.email,
      password: registerForm.password,
      password_confirm: registerForm.password_confirm,
      captcha: registerForm.captcha,
    })

    if (result.success) {
      await ElMessageBox.confirm(
        '注册成功！我们已向您的邮箱发送了验证邮件，请查收并点击邮件中的链接完成邮箱验证。',
        '注册成功',
        {
          confirmButtonText: '去登录',
          cancelButtonText: '稍后验证',
          type: 'success',
        }
      )
      router.push('/auth/login')
    } else {
      // 注册失败，刷新验证码
      captchaRef.value?.refreshCaptcha()
    }
  } catch (error) {
    // 用户取消或其他错误
    captchaRef.value?.refreshCaptcha()
  } finally {
    loading.value = false
  }
}

// 显示用户协议
const showUserAgreement = () => {
  agreementVisible.value = true
}

// 显示隐私政策
const showPrivacyPolicy = () => {
  privacyVisible.value = true
}
</script>

<style scoped lang="scss">
.register-view {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;

  .register-container {
    width: 100%;
    max-width: 400px;
    background: white;
    border-radius: 12px;
    padding: 40px 32px;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);

    .register-header {
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

      .success-icon {
        color: var(--el-color-success);
      }

      .error-icon {
        color: var(--el-color-danger);
      }

      .is-loading {
        animation: rotating 2s linear infinite;
      }
    }

    .register-button {
      width: 100%;
      height: 48px;
      font-size: 16px;
      font-weight: 600;
    }

    .login-link {
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
  }
}

.agreement-content,
.privacy-content {
  max-height: 400px;
  overflow-y: auto;
  line-height: 1.6;
  color: var(--el-text-color-regular);

  p {
    margin-bottom: 16px;
  }
}

@keyframes rotating {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

@media (max-width: 480px) {
  .register-view {
    padding: 12px;

    .register-container {
      padding: 24px 20px;
    }
  }
}
</style>
