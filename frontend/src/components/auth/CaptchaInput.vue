<template>
  <div class="captcha-input">
    <el-row :gutter="12">
      <el-col :span="14">
        <el-input
          v-model="captchaCode"
          placeholder="请输入验证码"
          :disabled="loading"
          @input="handleInput"
          @keyup.enter="$emit('submit')"
        >
          <template #prefix>
            <el-icon><Picture /></el-icon>
          </template>
        </el-input>
      </el-col>
      <el-col :span="10">
        <div 
          class="captcha-image" 
          :class="{ loading: loading }"
          @click="refreshCaptcha"
          :title="loading ? '加载中...' : '点击刷新验证码'"
        >
          <el-image
            v-if="captchaImage && !loading"
            :src="captchaImage"
            fit="contain"
            alt="验证码"
          />
          <div v-else class="captcha-placeholder">
            <el-icon v-if="loading" class="is-loading"><Loading /></el-icon>
            <span v-else>点击获取</span>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture, Loading } from '@element-plus/icons-vue'
import { authApi } from '@/api/auth'

interface Props {
  modelValue?: string
  disabled?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'submit'): void
  (e: 'captcha-loaded', key: string): void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  disabled: false,
})

const emit = defineEmits<Emits>()

// 状态
const captchaCode = ref(props.modelValue)
const captchaImage = ref('')
const captchaKey = ref('')
const loading = ref(false)

// 处理输入
const handleInput = (value: string) => {
  captchaCode.value = value
  emit('update:modelValue', value)
}

// 获取验证码
const getCaptcha = async () => {
  loading.value = true
  try {
    const response = await authApi.getCaptcha()
    captchaImage.value = `data:image/png;base64,${response.image}`
    captchaKey.value = response.key
    emit('captcha-loaded', response.key)
  } catch (error: any) {
    ElMessage.error(error.message || '获取验证码失败')
  } finally {
    loading.value = false
  }
}

// 刷新验证码
const refreshCaptcha = () => {
  if (!loading.value) {
    captchaCode.value = ''
    emit('update:modelValue', '')
    getCaptcha()
  }
}

// 验证验证码
const verifyCaptcha = async () => {
  if (!captchaCode.value || !captchaKey.value) {
    return false
  }

  try {
    const response = await authApi.verifyCaptcha({
      key: captchaKey.value,
      code: captchaCode.value,
    })
    return response.valid
  } catch (error) {
    return false
  }
}

// 暴露方法给父组件
defineExpose({
  refreshCaptcha,
  verifyCaptcha,
  getCaptchaKey: () => captchaKey.value,
})

// 组件挂载时获取验证码
onMounted(() => {
  getCaptcha()
})
</script>

<style scoped lang="scss">
.captcha-input {
  .captcha-image {
    height: 40px;
    border: 1px solid var(--el-border-color);
    border-radius: var(--el-border-radius-base);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--el-bg-color);
    transition: all 0.3s;
    overflow: hidden;

    &:hover {
      border-color: var(--el-color-primary);
    }

    &.loading {
      cursor: not-allowed;
      opacity: 0.6;
    }

    .el-image {
      width: 100%;
      height: 100%;
    }

    .captcha-placeholder {
      display: flex;
      align-items: center;
      justify-content: center;
      color: var(--el-text-color-placeholder);
      font-size: 12px;
      height: 100%;

      .el-icon {
        margin-right: 4px;
      }

      .is-loading {
        animation: rotating 2s linear infinite;
      }
    }
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
</style>
