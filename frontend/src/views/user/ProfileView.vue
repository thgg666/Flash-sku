<template>
  <div class="profile-view">
    <div class="container">
      <div class="profile-header">
        <h1>个人中心</h1>
        <p>管理您的账户信息和设置</p>
      </div>

      <el-row :gutter="24">
        <!-- 侧边栏 -->
        <el-col :xs="24" :sm="6" :md="6">
          <div class="sidebar">
            <div class="user-info">
              <el-avatar :size="80" :src="userAvatar">
                <el-icon><User /></el-icon>
              </el-avatar>
              <h3>{{ authStore.userName }}</h3>
              <p>{{ authStore.userEmail }}</p>
              <el-tag v-if="authStore.isEmailVerified" type="success" size="small">
                已验证
              </el-tag>
              <el-tag v-else type="warning" size="small">
                未验证
              </el-tag>
            </div>

            <el-menu
              :default-active="activeTab"
              class="profile-menu"
              @select="handleMenuSelect"
            >
              <el-menu-item index="profile">
                <el-icon><User /></el-icon>
                <span>个人资料</span>
              </el-menu-item>
              <el-menu-item index="security">
                <el-icon><Lock /></el-icon>
                <span>安全设置</span>
              </el-menu-item>
              <el-menu-item index="orders">
                <el-icon><ShoppingBag /></el-icon>
                <span>我的订单</span>
              </el-menu-item>
              <el-menu-item index="settings">
                <el-icon><Setting /></el-icon>
                <span>账户设置</span>
              </el-menu-item>
            </el-menu>
          </div>
        </el-col>

        <!-- 主内容区 -->
        <el-col :xs="24" :sm="18" :md="18">
          <div class="main-content">
            <!-- 个人资料 -->
            <div v-if="activeTab === 'profile'" class="content-section">
              <div class="section-header">
                <h2>个人资料</h2>
                <p>更新您的个人信息</p>
              </div>

              <el-form
                ref="profileFormRef"
                :model="profileForm"
                :rules="profileRules"
                label-width="100px"
                size="large"
              >
                <el-form-item label="用户名" prop="username">
                  <el-input
                    v-model="profileForm.username"
                    :disabled="true"
                    placeholder="用户名不可修改"
                  />
                </el-form-item>

                <el-form-item label="邮箱" prop="email">
                  <el-input
                    v-model="profileForm.email"
                    :disabled="true"
                    placeholder="邮箱不可修改"
                  >
                    <template #append>
                      <el-button
                        v-if="!authStore.isEmailVerified"
                        type="primary"
                        @click="sendVerificationEmail"
                        :loading="emailLoading"
                      >
                        验证邮箱
                      </el-button>
                    </template>
                  </el-input>
                </el-form-item>

                <el-form-item label="姓名" prop="first_name">
                  <el-row :gutter="12">
                    <el-col :span="12">
                      <el-input
                        v-model="profileForm.first_name"
                        placeholder="名"
                      />
                    </el-col>
                    <el-col :span="12">
                      <el-input
                        v-model="profileForm.last_name"
                        placeholder="姓"
                      />
                    </el-col>
                  </el-row>
                </el-form-item>

                <el-form-item>
                  <el-button
                    type="primary"
                    :loading="profileLoading"
                    @click="updateProfile"
                  >
                    保存修改
                  </el-button>
                  <el-button @click="resetProfile">
                    重置
                  </el-button>
                </el-form-item>
              </el-form>
            </div>

            <!-- 安全设置 -->
            <div v-else-if="activeTab === 'security'" class="content-section">
              <div class="section-header">
                <h2>安全设置</h2>
                <p>管理您的账户安全</p>
              </div>

              <el-form
                ref="passwordFormRef"
                :model="passwordForm"
                :rules="passwordRules"
                label-width="120px"
                size="large"
              >
                <el-form-item label="当前密码" prop="old_password">
                  <el-input
                    v-model="passwordForm.old_password"
                    type="password"
                    placeholder="请输入当前密码"
                    show-password
                  />
                </el-form-item>

                <el-form-item label="新密码" prop="new_password">
                  <PasswordStrength
                    v-model="passwordForm.new_password"
                    placeholder="请输入新密码"
                    @valid-change="handlePasswordValidChange"
                  />
                </el-form-item>

                <el-form-item label="确认新密码" prop="new_password_confirm">
                  <el-input
                    v-model="passwordForm.new_password_confirm"
                    type="password"
                    placeholder="请再次输入新密码"
                    show-password
                  />
                </el-form-item>

                <el-form-item>
                  <el-button
                    type="primary"
                    :loading="passwordLoading"
                    :disabled="!passwordValid"
                    @click="changePassword"
                  >
                    修改密码
                  </el-button>
                  <el-button @click="resetPasswordForm">
                    重置
                  </el-button>
                </el-form-item>
              </el-form>
            </div>

            <!-- 我的订单 -->
            <div v-else-if="activeTab === 'orders'" class="content-section">
              <div class="section-header">
                <h2>我的订单</h2>
                <p>查看您的订单历史</p>
              </div>

              <el-empty
                description="暂无订单数据"
                :image-size="120"
              >
                <el-button type="primary" @click="goToActivities">
                  去参加秒杀
                </el-button>
              </el-empty>
            </div>

            <!-- 账户设置 -->
            <div v-else-if="activeTab === 'settings'" class="content-section">
              <div class="section-header">
                <h2>账户设置</h2>
                <p>管理您的账户偏好</p>
              </div>

              <el-card>
                <div class="setting-item">
                  <div class="setting-info">
                    <h4>注销账户</h4>
                    <p>永久删除您的账户和所有数据</p>
                  </div>
                  <el-button type="danger" @click="handleDeleteAccount">
                    注销账户
                  </el-button>
                </div>
              </el-card>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock, ShoppingBag, Setting } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import PasswordStrength from '@/components/auth/PasswordStrength.vue'
import type { User as UserType } from '@/types'

// 路由和store
const router = useRouter()
const authStore = useAuthStore()

// 表单引用
const profileFormRef = ref<FormInstance>()
const passwordFormRef = ref<FormInstance>()

// 状态
const activeTab = ref('profile')
const profileLoading = ref(false)
const passwordLoading = ref(false)
const emailLoading = ref(false)
const passwordValid = ref(false)

// 计算属性
const userAvatar = computed(() => {
  // 这里可以返回用户头像URL
  return ''
})

// 个人资料表单
const profileForm = reactive<Partial<UserType>>({
  username: '',
  email: '',
  first_name: '',
  last_name: '',
})

// 密码修改表单
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  new_password_confirm: '',
})

// 表单验证规则
const profileRules: FormRules = {
  first_name: [
    { max: 30, message: '名字长度不能超过30个字符', trigger: 'blur' },
  ],
  last_name: [
    { max: 30, message: '姓氏长度不能超过30个字符', trigger: 'blur' },
  ],
}

const passwordRules: FormRules = {
  old_password: [
    { required: true, message: '请输入当前密码', trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码长度至少8个字符', trigger: 'blur' },
  ],
  new_password_confirm: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: (rule, value, callback) => {
      if (value !== passwordForm.new_password) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }, trigger: 'blur' },
  ],
}

// 处理菜单选择
const handleMenuSelect = (index: string) => {
  activeTab.value = index
}

// 处理密码有效性变化
const handlePasswordValidChange = (isValid: boolean) => {
  passwordValid.value = isValid
}

// 更新个人资料
const updateProfile = async () => {
  if (!profileFormRef.value) return

  try {
    const valid = await profileFormRef.value.validate()
    if (!valid) return
  } catch (error) {
    return
  }

  profileLoading.value = true
  try {
    const result = await authStore.updateUser({
      first_name: profileForm.first_name,
      last_name: profileForm.last_name,
    })

    if (result.success) {
      ElMessage.success('个人资料更新成功')
    }
  } catch (error) {
    // 错误已在store中处理
  } finally {
    profileLoading.value = false
  }
}

// 重置个人资料表单
const resetProfile = () => {
  if (authStore.user) {
    Object.assign(profileForm, {
      username: authStore.user.username,
      email: authStore.user.email,
      first_name: authStore.user.first_name || '',
      last_name: authStore.user.last_name || '',
    })
  }
}

// 修改密码
const changePassword = async () => {
  if (!passwordFormRef.value) return

  try {
    const valid = await passwordFormRef.value.validate()
    if (!valid) return
  } catch (error) {
    return
  }

  passwordLoading.value = true
  try {
    const result = await authStore.changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
      new_password_confirm: passwordForm.new_password_confirm,
    })

    if (result.success) {
      resetPasswordForm()
    }
  } catch (error) {
    // 错误已在store中处理
  } finally {
    passwordLoading.value = false
  }
}

// 重置密码表单
const resetPasswordForm = () => {
  Object.assign(passwordForm, {
    old_password: '',
    new_password: '',
    new_password_confirm: '',
  })
  passwordFormRef.value?.clearValidate()
}

// 发送验证邮件
const sendVerificationEmail = async () => {
  if (!authStore.userEmail) return

  emailLoading.value = true
  try {
    await authStore.sendEmailVerification(authStore.userEmail)
  } catch (error) {
    // 错误已在store中处理
  } finally {
    emailLoading.value = false
  }
}

// 跳转到活动页面
const goToActivities = () => {
  router.push('/activities')
}

// 处理账户注销
const handleDeleteAccount = async () => {
  try {
    await ElMessageBox.confirm(
      '此操作将永久删除您的账户和所有数据，且无法恢复。确定要继续吗？',
      '注销账户',
      {
        confirmButtonText: '确定注销',
        cancelButtonText: '取消',
        type: 'error',
      }
    )

    // 这里可以调用注销账户的API
    ElMessage.info('账户注销功能开发中...')
  } catch (error) {
    // 用户取消
  }
}

// 组件挂载时初始化数据
onMounted(() => {
  resetProfile()
  
  // 如果用户信息不完整，获取最新信息
  if (authStore.isAuthenticated && !authStore.user?.first_name) {
    authStore.fetchCurrentUser()
  }
})
</script>

<style scoped lang="scss">
.profile-view {
  min-height: 100vh;
  background: var(--el-bg-color-page);
  padding: 24px 0;

  .container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 24px;
  }

  .profile-header {
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

  .sidebar {
    background: white;
    border-radius: 8px;
    padding: 24px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);

    .user-info {
      text-align: center;
      margin-bottom: 24px;
      padding-bottom: 24px;
      border-bottom: 1px solid var(--el-border-color-lighter);

      .el-avatar {
        margin-bottom: 16px;
      }

      h3 {
        margin: 0 0 8px 0;
        color: var(--el-text-color-primary);
        font-size: 18px;
        font-weight: 600;
      }

      p {
        margin: 0 0 12px 0;
        color: var(--el-text-color-regular);
        font-size: 14px;
      }
    }

    .profile-menu {
      border: none;

      .el-menu-item {
        border-radius: 6px;
        margin-bottom: 4px;

        &:hover {
          background: var(--el-color-primary-light-9);
        }

        &.is-active {
          background: var(--el-color-primary);
          color: white;

          .el-icon {
            color: white;
          }
        }
      }
    }
  }

  .main-content {
    background: white;
    border-radius: 8px;
    padding: 32px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);

    .content-section {
      .section-header {
        margin-bottom: 32px;
        padding-bottom: 16px;
        border-bottom: 1px solid var(--el-border-color-lighter);

        h2 {
          margin: 0 0 8px 0;
          color: var(--el-text-color-primary);
          font-size: 24px;
          font-weight: 600;
        }

        p {
          margin: 0;
          color: var(--el-text-color-regular);
          font-size: 14px;
        }
      }

      .setting-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 16px 0;

        .setting-info {
          h4 {
            margin: 0 0 4px 0;
            color: var(--el-text-color-primary);
            font-size: 16px;
            font-weight: 500;
          }

          p {
            margin: 0;
            color: var(--el-text-color-regular);
            font-size: 14px;
          }
        }
      }
    }
  }
}

@media (max-width: 768px) {
  .profile-view {
    .sidebar {
      margin-bottom: 24px;
    }

    .main-content {
      padding: 24px 16px;
    }
  }
}
</style>
