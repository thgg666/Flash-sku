<template>
  <div class="forgot-password-view">
    <div class="forgot-password-container">
      <div class="forgot-password-header">
        <h2>忘记密码</h2>
        <p>请输入您的邮箱地址，我们将发送重置密码的链接</p>
      </div>

      <el-form
        ref="forgotPasswordFormRef"
        :model="forgotPasswordForm"
        :rules="forgotPasswordRules"
        label-width="0"
        size="large"
        @submit.prevent="handleSubmit"
      >
        <!-- 邮箱 -->
        <el-form-item prop="email">
          <el-input
            v-model="forgotPasswordForm.email"
            type="email"
            placeholder="请输入您的邮箱地址"
            :disabled="loading"
            @keyup.enter="handleSubmit"
          >
            <template #prefix>
              <el-icon><Message /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 验证码 -->
        <el-form-item prop="captcha">
          <CaptchaInput
            ref="captchaRef"
            v-model="forgotPasswordForm.captcha"
            :disabled="loading"
            @captcha-loaded="handleCaptchaLoaded"
            @submit="handleSubmit"
          />
        </el-form-item>

        <!-- 提交按钮 -->
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleSubmit"
            class="submit-button"
          >
            {{ loading ? '发送中...' : '发送重置链接' }}
          </el-button>
        </el-form-item>

        <!-- 返回登录 -->
        <div class="back-to-login">
          想起密码了？
          <router-link to="/auth/login" class="link">返回登录</router-link>
        </div>
      </el-form>

      <!-- 成功提示 -->
      <div v-if="emailSent" class="success-message">
        <el-result
          icon="success"
          title="邮件发送成功"
          sub-title="我们已向您的邮箱发送了重置密码的链接，请查收邮件并按照指示操作。"
        >
          <template #extra>
            <el-button type="primary" @click="resendEmail" :loading="resendLoading">
              重新发送
            </el-button>
            <el-button @click="goToLogin">返回登录</el-button>
          </template>
        </el-result>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Message } from '@element-plus/icons-vue'
import { authApi } from '@/api/auth'
import { isValidEmail } from '@/utils'
import CaptchaInput from '@/components/auth/CaptchaInput.vue'

// 路由
const router = useRouter()

// 表单引用
const forgotPasswordFormRef = ref<FormInstance>()
const captchaRef = ref()

// 状态
const loading = ref(false)
const resendLoading = ref(false)
const emailSent = ref(false)
const captchaKey = ref('')

// 表单数据
const forgotPasswordForm = reactive({
  email: '',
  captcha: '',
})

// 表单验证规则
const forgotPasswordRules: FormRules = {
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
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
  ],
}

// 处理验证码加载
const handleCaptchaLoaded = (key: string) => {
  captchaKey.value = key
}

// 处理提交
const handleSubmit = async () => {
  if (!forgotPasswordFormRef.value) return

  try {
    const valid = await forgotPasswordFormRef.value.validate()
    if (!valid) return
  } catch (error) {
    return
  }

  loading.value = true
  try {
    await authApi.forgotPassword({
      email: forgotPasswordForm.email,
      captcha: forgotPasswordForm.captcha,
    })

    emailSent.value = true
    ElMessage.success('重置密码邮件发送成功')
  } catch (error: any) {
    ElMessage.error(error.message || '发送失败，请稍后重试')
    
    // 刷新验证码
    captchaRef.value?.refreshCaptcha()
    forgotPasswordForm.captcha = ''
  } finally {
    loading.value = false
  }
}

// 重新发送邮件
const resendEmail = async () => {
  resendLoading.value = true
  try {
    await authApi.forgotPassword({
      email: forgotPasswordForm.email,
      captcha: forgotPasswordForm.captcha,
    })
    ElMessage.success('邮件重新发送成功')
  } catch (error: any) {
    ElMessage.error(error.message || '发送失败，请稍后重试')
  } finally {
    resendLoading.value = false
  }
}

// 返回登录
const goToLogin = () => {
  router.push('/auth/login')
}
</script>

<style scoped lang="scss">
@import "@/styles/variables.scss";

.forgot-password-view {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;

  .forgot-password-container {
    width: 100%;
    max-width: 400px;
    background: white;
    border-radius: 12px;
    padding: 40px 32px;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);

    .forgot-password-header {
      text-align: center;
      margin-bottom: 32px;

      h2 {
        color: $text-color-primary;
        margin-bottom: 8px;
        font-weight: 600;
      }

      p {
        color: $text-color-regular;
        font-size: 14px;
        line-height: 1.5;
      }
    }

    .submit-button {
      width: 100%;
      height: 48px;
      font-size: 16px;
      font-weight: 500;
    }

    .back-to-login {
      text-align: center;
      margin-top: 24px;
      color: $text-color-regular;
      font-size: 14px;

      .link {
        color: $primary-color;
        text-decoration: none;
        font-weight: 500;

        &:hover {
          text-decoration: underline;
        }
      }
    }

    .success-message {
      margin-top: 24px;

      :deep(.el-result) {
        padding: 20px 0;

        .el-result__title {
          margin-top: 16px;
        }

        .el-result__subtitle {
          margin-top: 8px;
          line-height: 1.5;
        }

        .el-result__extra {
          margin-top: 24px;

          .el-button {
            margin: 0 8px;
          }
        }
      }
    }
  }
}

// 移动端适配
@media (max-width: 768px) {
  .forgot-password-view {
    padding: 16px;

    .forgot-password-container {
      padding: 32px 24px;
    }
  }
}
</style>
