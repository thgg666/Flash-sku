<template>
  <div class="password-strength">
    <el-input
      v-model="password"
      type="password"
      :placeholder="placeholder"
      :disabled="disabled"
      show-password
      @input="handleInput"
      @blur="handleBlur"
    >
      <template #prefix>
        <el-icon><Lock /></el-icon>
      </template>
    </el-input>
    
    <!-- 密码强度指示器 -->
    <div v-if="showStrength && password" class="strength-indicator">
      <div class="strength-bar">
        <div 
          class="strength-fill" 
          :class="strengthClass"
          :style="{ width: strengthPercentage + '%' }"
        ></div>
      </div>
      <div class="strength-text">
        <span :class="strengthClass">{{ strengthText }}</span>
      </div>
    </div>

    <!-- 密码要求提示 -->
    <div v-if="showRequirements && password" class="password-requirements">
      <div class="requirement-item" :class="{ valid: hasMinLength }">
        <el-icon><Check v-if="hasMinLength" /><Close v-else /></el-icon>
        <span>至少8个字符</span>
      </div>
      <div class="requirement-item" :class="{ valid: hasLowerCase }">
        <el-icon><Check v-if="hasLowerCase" /><Close v-else /></el-icon>
        <span>包含小写字母</span>
      </div>
      <div class="requirement-item" :class="{ valid: hasUpperCase }">
        <el-icon><Check v-if="hasUpperCase" /><Close v-else /></el-icon>
        <span>包含大写字母</span>
      </div>
      <div class="requirement-item" :class="{ valid: hasNumber }">
        <el-icon><Check v-if="hasNumber" /><Close v-else /></el-icon>
        <span>包含数字</span>
      </div>
      <div class="requirement-item" :class="{ valid: hasSpecialChar }">
        <el-icon><Check v-if="hasSpecialChar" /><Close v-else /></el-icon>
        <span>包含特殊字符</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Lock, Check, Close } from '@element-plus/icons-vue'

interface Props {
  modelValue?: string
  placeholder?: string
  disabled?: boolean
  showStrength?: boolean
  showRequirements?: boolean
  minLength?: number
}

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'strength-change', strength: number): void
  (e: 'valid-change', isValid: boolean): void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '请输入密码',
  disabled: false,
  showStrength: true,
  showRequirements: true,
  minLength: 8,
})

const emit = defineEmits<Emits>()

// 状态
const password = ref(props.modelValue)

// 密码强度检查
const hasMinLength = computed(() => password.value.length >= props.minLength)
const hasLowerCase = computed(() => /[a-z]/.test(password.value))
const hasUpperCase = computed(() => /[A-Z]/.test(password.value))
const hasNumber = computed(() => /\d/.test(password.value))
const hasSpecialChar = computed(() => /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password.value))

// 计算密码强度 (0-100)
const strengthScore = computed(() => {
  let score = 0
  
  if (hasMinLength.value) score += 20
  if (hasLowerCase.value) score += 20
  if (hasUpperCase.value) score += 20
  if (hasNumber.value) score += 20
  if (hasSpecialChar.value) score += 20
  
  return score
})

// 强度等级
const strengthLevel = computed(() => {
  if (strengthScore.value >= 80) return 'strong'
  if (strengthScore.value >= 60) return 'medium'
  if (strengthScore.value >= 40) return 'weak'
  return 'very-weak'
})

// 强度文本
const strengthText = computed(() => {
  switch (strengthLevel.value) {
    case 'strong': return '强'
    case 'medium': return '中等'
    case 'weak': return '弱'
    case 'very-weak': return '很弱'
    default: return ''
  }
})

// 强度样式类
const strengthClass = computed(() => `strength-${strengthLevel.value}`)

// 强度百分比
const strengthPercentage = computed(() => strengthScore.value)

// 密码是否有效
const isValid = computed(() => {
  return hasMinLength.value && 
         hasLowerCase.value && 
         hasUpperCase.value && 
         hasNumber.value
})

// 处理输入
const handleInput = (value: string) => {
  password.value = value
  emit('update:modelValue', value)
}

// 处理失焦
const handleBlur = () => {
  // 可以在这里添加额外的验证逻辑
}

// 监听密码变化，发出强度和有效性事件
watch(strengthScore, (newScore) => {
  emit('strength-change', newScore)
})

watch(isValid, (newValid) => {
  emit('valid-change', newValid)
})

// 监听props变化
watch(() => props.modelValue, (newValue) => {
  password.value = newValue
})
</script>

<style scoped lang="scss">
.password-strength {
  .strength-indicator {
    margin-top: 8px;
    
    .strength-bar {
      height: 4px;
      background: var(--el-border-color-lighter);
      border-radius: 2px;
      overflow: hidden;
      
      .strength-fill {
        height: 100%;
        transition: all 0.3s ease;
        border-radius: 2px;
        
        &.strength-very-weak {
          background: var(--el-color-danger);
        }
        
        &.strength-weak {
          background: var(--el-color-warning);
        }
        
        &.strength-medium {
          background: var(--el-color-primary);
        }
        
        &.strength-strong {
          background: var(--el-color-success);
        }
      }
    }
    
    .strength-text {
      margin-top: 4px;
      font-size: 12px;
      
      .strength-very-weak {
        color: var(--el-color-danger);
      }
      
      .strength-weak {
        color: var(--el-color-warning);
      }
      
      .strength-medium {
        color: var(--el-color-primary);
      }
      
      .strength-strong {
        color: var(--el-color-success);
      }
    }
  }
  
  .password-requirements {
    margin-top: 12px;
    padding: 12px;
    background: var(--el-bg-color-page);
    border-radius: var(--el-border-radius-base);
    border: 1px solid var(--el-border-color-lighter);
    
    .requirement-item {
      display: flex;
      align-items: center;
      margin-bottom: 6px;
      font-size: 12px;
      color: var(--el-text-color-regular);
      
      &:last-child {
        margin-bottom: 0;
      }
      
      .el-icon {
        margin-right: 6px;
        font-size: 14px;
        color: var(--el-color-danger);
      }
      
      &.valid {
        color: var(--el-color-success);
        
        .el-icon {
          color: var(--el-color-success);
        }
      }
    }
  }
}
</style>
