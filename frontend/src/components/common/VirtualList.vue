<template>
  <div 
    ref="containerRef" 
    class="virtual-list"
    :style="{ height: containerHeight + 'px' }"
    @scroll="handleScroll"
  >
    <!-- 占位符，用于撑开滚动条 -->
    <div 
      class="virtual-list-spacer"
      :style="{ height: totalHeight + 'px' }"
    ></div>
    
    <!-- 可见项容器 -->
    <div 
      class="virtual-list-items"
      :style="{ 
        transform: `translateY(${offsetY}px)`,
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0
      }"
    >
      <div
        v-for="(item, index) in visibleItems"
        :key="getItemKey(item, startIndex + index)"
        class="virtual-list-item"
        :style="{ height: itemHeight + 'px' }"
      >
        <slot 
          :item="item" 
          :index="startIndex + index"
          :isVisible="true"
        >
          {{ item }}
        </slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { usePerformanceOptimization } from '@/composables/usePerformanceOptimization'

interface Props {
  items: any[]
  itemHeight: number
  containerHeight: number
  buffer?: number
  keyField?: string
}

const props = withDefaults(defineProps<Props>(), {
  buffer: 5,
  keyField: 'id'
})

const emit = defineEmits<{
  scroll: [{ scrollTop: number; startIndex: number; endIndex: number }]
  itemVisible: [{ item: any; index: number }]
}>()

// 性能优化
const { useRenderOptimization, useMemoryGuard } = usePerformanceOptimization()
const { createThrottledUpdate } = useRenderOptimization()
const { createSafeEventListener } = useMemoryGuard()

// 响应式数据
const containerRef = ref<HTMLElement>()
const scrollTop = ref(0)
const startIndex = ref(0)
const endIndex = ref(0)

// 计算属性
const totalHeight = computed(() => props.items.length * props.itemHeight)

const visibleCount = computed(() => 
  Math.ceil(props.containerHeight / props.itemHeight)
)

const visibleItems = computed(() => {
  const start = Math.max(0, startIndex.value - props.buffer)
  const end = Math.min(
    props.items.length,
    endIndex.value + props.buffer
  )
  return props.items.slice(start, end)
})

const offsetY = computed(() => 
  Math.max(0, startIndex.value - props.buffer) * props.itemHeight
)

// 方法
const getItemKey = (item: any, index: number): string | number => {
  if (props.keyField && item && typeof item === 'object') {
    return item[props.keyField] || index
  }
  return index
}

const updateVisibleRange = () => {
  const newStartIndex = Math.floor(scrollTop.value / props.itemHeight)
  const newEndIndex = Math.min(
    props.items.length - 1,
    newStartIndex + visibleCount.value
  )

  if (newStartIndex !== startIndex.value || newEndIndex !== endIndex.value) {
    startIndex.value = newStartIndex
    endIndex.value = newEndIndex

    // 发送滚动事件
    emit('scroll', {
      scrollTop: scrollTop.value,
      startIndex: startIndex.value,
      endIndex: endIndex.value
    })

    // 发送可见项事件
    visibleItems.value.forEach((item, index) => {
      emit('itemVisible', {
        item,
        index: startIndex.value + index
      })
    })
  }
}

// 节流的滚动处理
const throttledUpdateVisibleRange = createThrottledUpdate(updateVisibleRange, 16)

const handleScroll = (event: Event) => {
  const target = event.target as HTMLElement
  scrollTop.value = target.scrollTop
  throttledUpdateVisibleRange()
}

// 滚动到指定索引
const scrollToIndex = (index: number, behavior: ScrollBehavior = 'smooth') => {
  if (!containerRef.value) return

  const targetScrollTop = index * props.itemHeight
  containerRef.value.scrollTo({
    top: targetScrollTop,
    behavior
  })
}

// 滚动到指定项
const scrollToItem = (item: any, behavior: ScrollBehavior = 'smooth') => {
  const index = props.items.findIndex(i => 
    props.keyField ? i[props.keyField] === item[props.keyField] : i === item
  )
  
  if (index !== -1) {
    scrollToIndex(index, behavior)
  }
}

// 获取可见范围
const getVisibleRange = () => ({
  startIndex: startIndex.value,
  endIndex: endIndex.value,
  visibleItems: visibleItems.value
})

// 刷新列表
const refresh = () => {
  nextTick(() => {
    updateVisibleRange()
  })
}

// 监听数据变化
watch(() => props.items, () => {
  // 重置滚动位置和可见范围
  scrollTop.value = 0
  startIndex.value = 0
  endIndex.value = Math.min(props.items.length - 1, visibleCount.value)
  
  nextTick(() => {
    if (containerRef.value) {
      containerRef.value.scrollTop = 0
    }
    updateVisibleRange()
  })
}, { deep: false })

// 监听容器高度变化
watch(() => props.containerHeight, () => {
  nextTick(() => {
    updateVisibleRange()
  })
})

// 生命周期
onMounted(() => {
  if (containerRef.value) {
    // 初始化可见范围
    endIndex.value = Math.min(props.items.length - 1, visibleCount.value)
    updateVisibleRange()

    // 监听窗口大小变化
    const handleResize = createThrottledUpdate(() => {
      refresh()
    }, 100)

    createSafeEventListener(window, 'resize', handleResize)
  }
})

// 暴露方法给父组件
defineExpose({
  scrollToIndex,
  scrollToItem,
  getVisibleRange,
  refresh,
  containerRef
})
</script>

<style scoped>
.virtual-list {
  position: relative;
  overflow-y: auto;
  overflow-x: hidden;
}

.virtual-list-spacer {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  pointer-events: none;
}

.virtual-list-items {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
}

.virtual-list-item {
  position: relative;
  overflow: hidden;
}

/* 滚动条样式 */
.virtual-list::-webkit-scrollbar {
  width: 6px;
}

.virtual-list::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.virtual-list::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.virtual-list::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 性能优化：减少重绘 */
.virtual-list-items {
  will-change: transform;
}

.virtual-list-item {
  contain: layout style paint;
}
</style>
