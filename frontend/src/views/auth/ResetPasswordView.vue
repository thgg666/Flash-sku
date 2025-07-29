<template>
  <div class="reset-password-view">
    <div class="reset-password-container">
      <div class="reset-password-header">
        <h2>重置密码</h2>
        <p>请设置您的新密码</p>
      </div>

      <el-form
        ref="resetPasswordFormRef"
        :model="resetPasswordForm"
        :rules="resetPasswordRules"
        label-width="0"
        size="large"
        @submit.prevent="handleSubmit"
      >
        <!-- 新密码 -->
        <el-form-item prop="password">
          <PasswordStrength
            v-model="resetPasswordForm.password"
            placeholder="请输入新密码"
            :disabled="loading"
            :show-strength="true"
            :show-requirements="true"
            @strength-change="handlePasswordStrengthChange"
          />
        </el-form-item>

        <!-- 确认密码 -->
        <el-form-item prop="confirmPassword">
          <el-input
            v-model="resetPasswordForm.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            :disabled="loading"
            show-password
            @keyup.enter="handleSubmit"
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 提交按钮 -->
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            :disabled="!isPasswordStrong"
            @click="handleSubmit"
            class="submit-button"
          >
            {{ loading ? '重置中...' : '重置密码' }}
          </el-button>
        </el-form-item>

        <!-- 返回登录 -->
        <div class="back-to-login">
          <router-link to="/auth/login" class="link">返回登录</router-link>
        </div>
      </el-form>

      <!-- 成功提示 -->
      <div v-if="resetSuccess" class="success-message">
        <el-result
          icon="success"
          title="密码重置成功"
          sub-title="您的密码已成功重置，请使用新密码登录。"
        >
          <template #extra>
            <el-button type="primary" @click="goToLogin">立即登录</el-button>
          </template>
        </el-result>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Lock } from '@element-plus/icons-vue'
import { authApi } from '@/api/auth'
import PasswordStrength from '@/components/auth/PasswordStrength.vue'

// 路由
const router = useRouter()
const route = useRoute()

// 表单引用
const resetPasswordFormRef = ref<FormInstance>()

// 状态
const loading = ref(false)
const resetSuccess = ref(false)
const passwordStrength = ref(0)

// 获取URL参数中的token
const resetToken = computed(() => route.query.token as string)

// 表单数据
const resetPasswordForm = reactive({
  password: '',
  confirmPassword: '',
})

// 计算属性
const isPasswordStrong = computed(() => passwordStrength.value >= 3)

// 表单验证规则
const resetPasswordRules: FormRules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码长度至少8个字符', trigger: 'blur' },
    { 
      validator: (rule, value, callback) => {
        if (passwordStrength.value < 3) {
          callback(new Error('密码强度不够，请设置更强的密码'))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    },
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { 
      validator: (rule, value, callback) => {
        if (value !== resetPasswordForm.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    },
  ],
}

// 处理密码强度变化
const handlePasswordStrengthChange = (strength: number) => {
  passwordStrength.value = strength
}

// 处理提交
const handleSubmit = async () => {
  if (!resetToken.value) {
    ElMessage.error('重置链接无效或已过期')
    router.push('/auth/forgot-password')
    return
  }

  if (!resetPasswordFormRef.value) return

  try {
    const valid = await resetPasswordFormRef.value.validate()
    if (!valid) return
  } catch (error) {
    return
  }

  loading.value = true
  try {
    await authApi.resetPassword({
      token: resetToken.value,
      password: resetPasswordForm.password,
      confirmPassword: resetPasswordForm.confirmPassword,
    })

    resetSuccess.value = true
    ElMessage.success('密码重置成功')
  } catch (error: any) {
    ElMessage.error(error.message || '重置失败，请稍后重试')
    
    // 如果token无效，跳转到忘记密码页面
    if (error.response?.status === 400) {
      setTimeout(() => {
        router.push('/auth/forgot-password')
      }, 2000)
    }
  } finally {
    loading.value = false
  }
}

// 返回登录
const goToLogin = () => {
  router.push('/auth/login')
}

// 检查token有效性
const checkTokenValidity = async () => {
  if (!resetToken.value) {
    ElMessage.error('重置链接无效')
    router.push('/auth/forgot-password')
    return
  }

  try {
    await authApi.verifyResetToken(resetToken.value)
  } catch (error) {
    ElMessage.error('重置链接无效或已过期')
    router.push('/auth/forgot-password')
  }
}

// 组件挂载时检查token
checkTokenValidity()
</script>

<style scoped lang="scss">
@import "@/styles/variables.scss";

.reset-password-view {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;

  .reset-password-container {
    width: 100%;
    max-width: 400px;
    background: white;
    border-radius: 12px;
    padding: 40px 32px;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);

    .reset-password-header {
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
      }
    }

    .submit-button {
      width: 100%;
      height: 48px;
      font-size: 16px;
      font-weight: 500;

      &:disabled {
        opacity: 0.6;
        cursor: not-allowed;
      }
    }

    .back-to-login {
      text-align: center;
      margin-top: 24px;

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
        }
      }
    }
  }
}

// 移动端适配
@media (max-width: 768px) {
  .reset-password-view {
    padding: 16px;

    .reset-password-container {
      padding: 32px 24px;
    }
  }
}
</style>
