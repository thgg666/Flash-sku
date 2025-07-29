<template>
  <div class="verify-email-view">
    <div class="verify-container">
      <div class="verify-header">
        <el-icon class="header-icon" :class="{ success: verifySuccess }">
          <Message v-if="!verifySuccess" />
          <Check v-else />
        </el-icon>
        <h2>{{ verifySuccess ? '验证成功' : '邮箱验证' }}</h2>
        <p v-if="!verifySuccess">请输入发送到您邮箱的验证码</p>
        <p v-else>您的邮箱已成功验证</p>
      </div>

      <!-- 验证成功状态 -->
      <div v-if="verifySuccess" class="success-content">
        <el-result
          icon="success"
          title="邮箱验证成功"
          sub-title="您现在可以正常使用所有功能了"
        >
          <template #extra>
            <el-button type="primary" @click="goToLogin">
              去登录
            </el-button>
            <el-button @click="goToHome">
              返回首页
            </el-button>
          </template>
        </el-result>
      </div>

      <!-- 验证表单 -->
      <div v-else class="verify-form">
        <el-form
          ref="verifyFormRef"
          :model="verifyForm"
          :rules="verifyRules"
          label-width="0"
          size="large"
          @submit.prevent="handleVerify"
        >
          <!-- 邮箱地址 -->
          <el-form-item prop="email">
            <el-input
              v-model="verifyForm.email"
              placeholder="请输入邮箱地址"
              :disabled="loading"
            >
              <template #prefix>
                <el-icon><Message /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <!-- 验证码 -->
          <el-form-item prop="code">
            <el-input
              v-model="verifyForm.code"
              placeholder="请输入6位验证码"
              :disabled="loading"
              maxlength="6"
              @keyup.enter="handleVerify"
            >
              <template #prefix>
                <el-icon><Key /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <!-- 重新发送验证码 -->
          <div class="resend-section">
            <span class="resend-text">
              没有收到验证码？
            </span>
            <el-button
              type="text"
              :disabled="loading || countdown > 0"
              @click="handleResendCode"
              class="resend-button"
            >
              {{ countdown > 0 ? `${countdown}秒后重新发送` : '重新发送' }}
            </el-button>
          </div>

          <!-- 验证按钮 -->
          <el-form-item>
            <el-button
              type="primary"
              size="large"
              :loading="loading"
              @click="handleVerify"
              class="verify-button"
            >
              {{ loading ? '验证中...' : '验证邮箱' }}
            </el-button>
          </el-form-item>

          <!-- 返回登录 -->
          <div class="back-link">
            <router-link to="/auth/login" class="link">
              返回登录
            </router-link>
          </div>
        </el-form>
      </div>

      <!-- 验证提示 -->
      <div v-if="!verifySuccess" class="verify-tips">
        <el-alert
          title="验证提示"
          type="info"
          :closable="false"
          show-icon
        >
          <ul>
            <li>验证码有效期为10分钟</li>
            <li>请检查邮箱的垃圾邮件文件夹</li>
            <li>如果长时间未收到邮件，请联系客服</li>
          </ul>
        </el-alert>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Message, Check, Key } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { isValidEmail } from '@/utils'

// 路由和store
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// 表单引用
const verifyFormRef = ref<FormInstance>()

// 状态
const loading = ref(false)
const verifySuccess = ref(false)
const countdown = ref(0)
const countdownTimer = ref<number | null>(null)

// 表单数据
const verifyForm = reactive({
  email: '',
  code: '',
})

// 表单验证规则
const verifyRules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { validator: (_rule, value, callback) => {
      if (!isValidEmail(value)) {
        callback(new Error('请输入有效的邮箱地址'))
      } else {
        callback()
      }
    }, trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '验证码只能包含数字', trigger: 'blur' },
  ],
}

// 处理邮箱验证
const handleVerify = async () => {
  if (!verifyFormRef.value) return

  try {
    const valid = await verifyFormRef.value.validate()
    if (!valid) return
  } catch (error) {
    return
  }

  loading.value = true
  try {
    const result = await authStore.verifyEmail({
      email: verifyForm.email,
      code: verifyForm.code,
    })

    if (result.success) {
      verifySuccess.value = true
      ElMessage.success('邮箱验证成功！')
    }
  } catch (error) {
    // 错误已在store中处理
  } finally {
    loading.value = false
  }
}

// 重新发送验证码
const handleResendCode = async () => {
  if (!verifyForm.email) {
    ElMessage.warning('请先输入邮箱地址')
    return
  }

  if (!isValidEmail(verifyForm.email)) {
    ElMessage.warning('请输入有效的邮箱地址')
    return
  }

  loading.value = true
  try {
    const result = await authStore.sendEmailVerification(verifyForm.email)
    
    if (result.success) {
      // 开始倒计时
      startCountdown()
    }
  } catch (error) {
    // 错误已在store中处理
  } finally {
    loading.value = false
  }
}

// 开始倒计时
const startCountdown = () => {
  countdown.value = 60
  countdownTimer.value = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countdownTimer.value!)
      countdownTimer.value = null
    }
  }, 1000)
}

// 跳转到登录页
const goToLogin = () => {
  router.push('/auth/login')
}

// 跳转到首页
const goToHome = () => {
  router.push('/')
}

// 组件挂载时处理URL参数
onMounted(() => {
  // 从URL参数中获取邮箱地址
  const email = route.query.email as string
  if (email && isValidEmail(email)) {
    verifyForm.email = email
  }

  // 从URL参数中获取验证码（如果是通过邮件链接访问）
  const code = route.query.code as string
  if (code && /^\d{6}$/.test(code)) {
    verifyForm.code = code
    // 自动验证
    setTimeout(() => {
      handleVerify()
    }, 500)
  }
})

// 组件卸载时清理定时器
onUnmounted(() => {
  if (countdownTimer.value) {
    clearInterval(countdownTimer.value)
  }
})
</script>

<style scoped lang="scss">
.verify-email-view {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;

  .verify-container {
    width: 100%;
    max-width: 480px;
    background: white;
    border-radius: 12px;
    padding: 40px 32px;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);

    .verify-header {
      text-align: center;
      margin-bottom: 32px;

      .header-icon {
        font-size: 48px;
        color: var(--el-color-primary);
        margin-bottom: 16px;

        &.success {
          color: var(--el-color-success);
        }
      }

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

    .success-content {
      .el-result {
        padding: 0;
      }
    }

    .verify-form {
      .el-form-item {
        margin-bottom: 20px;
      }

      .resend-section {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 24px;
        font-size: 14px;

        .resend-text {
          color: var(--el-text-color-regular);
        }

        .resend-button {
          padding: 0;
          font-size: 14px;
        }
      }

      .verify-button {
        width: 100%;
        height: 48px;
        font-size: 16px;
        font-weight: 600;
      }

      .back-link {
        text-align: center;
        margin-top: 24px;

        .link {
          color: var(--el-color-primary);
          text-decoration: none;
          font-size: 14px;

          &:hover {
            text-decoration: underline;
          }
        }
      }
    }

    .verify-tips {
      margin-top: 24px;

      .el-alert {
        ul {
          margin: 8px 0 0 0;
          padding-left: 16px;

          li {
            margin-bottom: 4px;
            font-size: 13px;
            color: var(--el-text-color-regular);

            &:last-child {
              margin-bottom: 0;
            }
          }
        }
      }
    }
  }
}

@media (max-width: 480px) {
  .verify-email-view {
    padding: 12px;

    .verify-container {
      padding: 24px 20px;
    }
  }
}
</style>
