<template>
  <picture class="optimized-image">
    <!-- AVIF格式 -->
    <source 
      v-if="avifSupported && avifSrcSet"
      :srcset="avifSrcSet"
      :sizes="sizes"
      type="image/avif"
    />
    
    <!-- WebP格式 -->
    <source 
      v-if="webpSupported && webpSrcSet"
      :srcset="webpSrcSet"
      :sizes="sizes"
      type="image/webp"
    />
    
    <!-- 原始格式 -->
    <img
      ref="imgRef"
      :src="optimizedSrc"
      :srcset="originalSrcSet"
      :sizes="sizes"
      :alt="alt"
      :loading="loading"
      :decoding="decoding"
      :class="imageClasses"
      :style="imageStyles"
      @load="handleLoad"
      @error="handleError"
      @loadstart="handleLoadStart"
    />
  </picture>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ImageOptimizer, detectImageFormatSupport, type ImageOptimizationConfig } from '@/utils/resourceOptimization'

// Props定义
interface Props {
  src: string
  alt: string
  width?: number
  height?: number
  sizes?: string
  quality?: number
  format?: 'auto' | 'webp' | 'jpeg' | 'png'
  loading?: 'lazy' | 'eager'
  decoding?: 'async' | 'sync' | 'auto'
  responsive?: boolean
  breakpoints?: number[]
  placeholder?: string
  errorImage?: string
  aspectRatio?: string
  objectFit?: 'cover' | 'contain' | 'fill' | 'scale-down' | 'none'
  blur?: boolean
  progressive?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: 'lazy',
  decoding: 'async',
  responsive: true,
  quality: 80,
  format: 'auto',
  breakpoints: () => [320, 640, 768, 1024, 1280, 1920],
  objectFit: 'cover',
  progressive: true
})

// Emits定义
const emit = defineEmits<{
  load: [event: Event]
  error: [event: Event]
  loadstart: [event: Event]
}>()

// 响应式数据
const imgRef = ref<HTMLImageElement>()
const imageOptimizer = new ImageOptimizer()
const isLoading = ref(true)
const hasError = ref(false)
const formatSupport = ref({
  webp: false,
  avif: false,
  jpeg2000: false,
  jpegXR: false
})

// 计算属性
const avifSupported = computed(() => formatSupport.value.avif)
const webpSupported = computed(() => formatSupport.value.webp)

const optimizationConfig = computed((): Partial<ImageOptimizationConfig> => ({
  quality: props.quality,
  format: props.format,
  progressive: props.progressive
}))

const optimizedSrc = computed(() => {
  if (!props.responsive || !props.width) {
    return props.src
  }
  return imageOptimizer.generateResponsiveImageUrl(props.src, props.width, optimizationConfig.value)
})

const originalSrcSet = computed(() => {
  if (!props.responsive) return undefined
  return imageOptimizer.generateSrcSet(props.src, optimizationConfig.value)
})

const webpSrcSet = computed(() => {
  if (!props.responsive || !webpSupported.value) return undefined
  return imageOptimizer.generateSrcSet(props.src, {
    ...optimizationConfig.value,
    format: 'webp'
  })
})

const avifSrcSet = computed(() => {
  if (!props.responsive || !avifSupported.value) return undefined
  return imageOptimizer.generateSrcSet(props.src, {
    ...optimizationConfig.value,
    format: 'avif'
  })
})

const imageClasses = computed(() => ({
  'optimized-image__img': true,
  'optimized-image__img--loading': isLoading.value,
  'optimized-image__img--error': hasError.value,
  'optimized-image__img--blur': props.blur && isLoading.value
}))

const imageStyles = computed(() => ({
  aspectRatio: props.aspectRatio,
  objectFit: props.objectFit,
  width: props.width ? `${props.width}px` : undefined,
  height: props.height ? `${props.height}px` : undefined
}))

// 方法
const handleLoad = (event: Event) => {
  isLoading.value = false
  hasError.value = false
  emit('load', event)
}

const handleError = (event: Event) => {
  isLoading.value = false
  hasError.value = true
  
  // 如果有错误图片，设置为错误图片
  if (props.errorImage && imgRef.value) {
    imgRef.value.src = props.errorImage
  }
  
  emit('error', event)
}

const handleLoadStart = (event: Event) => {
  isLoading.value = true
  emit('loadstart', event)
}

// 生命周期
onMounted(async () => {
  // 检测格式支持
  formatSupport.value = await detectImageFormatSupport()
  
  // 如果有占位图，先显示占位图
  if (props.placeholder && imgRef.value) {
    imgRef.value.src = props.placeholder
  }
})

// 监听src变化
watch(() => props.src, () => {
  isLoading.value = true
  hasError.value = false
})
</script>

<style scoped lang="scss">
@import "@/styles/variables.scss";

.optimized-image {
  display: inline-block;
  position: relative;
  overflow: hidden;

  &__img {
    display: block;
    width: 100%;
    height: auto;
    transition: opacity 0.3s ease, filter 0.3s ease;

    &--loading {
      opacity: 0.7;
    }

    &--error {
      opacity: 0.5;
      filter: grayscale(100%);
    }

    &--blur {
      filter: blur(5px);
    }
  }

  // 响应式图片容器
  &--responsive {
    width: 100%;
    height: auto;
  }

  // 固定宽高比容器
  &--aspect-ratio {
    position: relative;
    
    &::before {
      content: '';
      display: block;
      padding-top: var(--aspect-ratio, 56.25%); // 默认16:9
    }

    .optimized-image__img {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  // 加载状态指示器
  &--loading::after {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 20px;
    height: 20px;
    margin: -10px 0 0 -10px;
    border: 2px solid $border-color-lighter;
    border-top-color: $primary-color;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
}

// 不同尺寸的预设样式
.optimized-image {
  &--small {
    max-width: 200px;
  }

  &--medium {
    max-width: 400px;
  }

  &--large {
    max-width: 800px;
  }

  &--full {
    width: 100%;
  }
}

// 圆角样式
.optimized-image {
  &--rounded {
    border-radius: 8px;
    overflow: hidden;
  }

  &--circle {
    border-radius: 50%;
    overflow: hidden;
    aspect-ratio: 1;

    .optimized-image__img {
      object-fit: cover;
    }
  }
}

// 阴影效果
.optimized-image {
  &--shadow {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  &--shadow-lg {
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  }
}

// 悬停效果
.optimized-image {
  &--hover-zoom {
    overflow: hidden;

    .optimized-image__img {
      transition: transform 0.3s ease;
    }

    &:hover .optimized-image__img {
      transform: scale(1.05);
    }
  }

  &--hover-brightness {
    .optimized-image__img {
      transition: filter 0.3s ease;
    }

    &:hover .optimized-image__img {
      filter: brightness(1.1);
    }
  }
}

// 移动端优化
@media (max-width: 768px) {
  .optimized-image {
    &--mobile-full {
      width: 100%;
    }

    &--mobile-square {
      aspect-ratio: 1;
    }
  }
}

// 高分辨率屏幕优化
@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
  .optimized-image__img {
    image-rendering: -webkit-optimize-contrast;
    image-rendering: crisp-edges;
  }
}

// 减少动画偏好
@media (prefers-reduced-motion: reduce) {
  .optimized-image__img {
    transition: none !important;
  }

  .optimized-image--loading::after {
    animation: none !important;
  }
}

// 高对比度模式
@media (prefers-contrast: high) {
  .optimized-image__img {
    &--error {
      filter: contrast(1.5) grayscale(100%);
    }
  }
}

// 暗色主题适配
@media (prefers-color-scheme: dark) {
  .optimized-image {
    &--loading::after {
      border-color: rgba(255, 255, 255, 0.3);
      border-top-color: #fff;
    }
  }
}
</style>
