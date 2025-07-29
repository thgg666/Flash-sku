<template>
  <div class="feedback-notification">
    <!-- 通知图标 -->
    <el-badge :value="unreadCount" :hidden="!hasUnreadFeedback" :max="99">
      <el-button 
        circle 
        :type="hasUnreadFeedback ? 'primary' : 'default'"
        @click="togglePanel"
      >
        <el-icon><Bell /></el-icon>
      </el-button>
    </el-badge>

    <!-- 通知面板 -->
    <el-drawer
      v-model="panelVisible"
      title="消息通知"
      direction="rtl"
      size="400px"
    >
      <template #header>
        <div class="panel-header">
          <span>消息通知</span>
          <div class="header-actions">
            <el-button 
              v-if="hasUnreadFeedback"
              size="small" 
              type="primary" 
              text
              @click="markAllAsRead"
            >
              全部已读
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              text
              @click="clearAllHistory"
            >
              清空
            </el-button>
          </div>
        </div>
      </template>

      <div class="notification-content">
        <!-- 空状态 -->
        <div v-if="recentFeedback.length === 0" class="empty-state">
          <el-empty 
            description="暂无消息通知" 
            :image-size="80"
          />
        </div>

        <!-- 消息列表 -->
        <div v-else class="message-list">
          <div 
            v-for="(message, index) in recentFeedback" 
            :key="index"
            class="message-item"
            :class="getMessageClass(message)"
            @click="handleMessageClick(message)"
          >
            <div class="message-icon">
              <el-icon><component :is="getMessageIcon(message.type)" /></el-icon>
            </div>
            
            <div class="message-content">
              <div class="message-title">{{ getMessageTitle(message) }}</div>
              <div class="message-text">{{ getMessageText(message) }}</div>
              <div class="message-time">{{ formatMessageTime(message.timestamp) }}</div>
            </div>
            
            <div class="message-actions">
              <el-button 
                size="small" 
                circle 
                @click.stop="removeMessage(index)"
              >
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="panel-footer">
          <el-button @click="panelVisible = false">关闭</el-button>
          <el-button type="primary" @click="goToMessageCenter">
            消息中心
          </el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Bell, 
  Close, 
  ShoppingCart, 
  Clock, 
  InfoFilled, 
  Warning 
} from '@element-plus/icons-vue'
import { useRealTimeFeedback } from '@/composables/useRealTimeFeedback'
import { formatRelativeTime } from '@/utils'

// 路由
const router = useRouter()

// 实时反馈
const {
  recentFeedback,
  unreadCount,
  hasUnreadFeedback,
  markAsRead,
  clearHistory
} = useRealTimeFeedback()

// 状态
const panelVisible = ref(false)

// 切换面板显示
const togglePanel = () => {
  panelVisible.value = !panelVisible.value
}

// 获取消息图标
const getMessageIcon = (type: string) => {
  switch (type) {
    case 'seckill_result':
      return ShoppingCart
    case 'queue_update':
      return Clock
    case 'activity_reminder':
      return Warning
    default:
      return InfoFilled
  }
}

// 获取消息样式类
const getMessageClass = (message: any) => {
  const classes = ['message-item']
  
  switch (message.type) {
    case 'seckill_result':
      if (message.data.result.code === 'SUCCESS') {
        classes.push('success')
      } else {
        classes.push('error')
      }
      break
    case 'queue_update':
      classes.push('info')
      break
    case 'system_message':
      classes.push(message.data.level)
      break
    case 'activity_reminder':
      classes.push('warning')
      break
    default:
      classes.push('info')
  }
  
  return classes
}

// 获取消息标题
const getMessageTitle = (message: any) => {
  switch (message.type) {
    case 'seckill_result':
      return message.data.result.code === 'SUCCESS' ? '抢购成功' : '抢购失败'
    case 'queue_update':
      return '排队状态更新'
    case 'system_message':
      return '系统消息'
    case 'activity_reminder':
      return '活动提醒'
    default:
      return '通知'
  }
}

// 获取消息内容
const getMessageText = (message: any) => {
  switch (message.type) {
    case 'seckill_result':
      if (message.data.result.code === 'SUCCESS') {
        return `订单号：${message.data.result.order_id}`
      } else {
        return message.data.result.message || '抢购失败'
      }
    case 'queue_update':
      return `排队位置：${message.data.queue_position}，预计等待：${message.data.estimated_wait_time}秒`
    case 'system_message':
      return message.data.message
    case 'activity_reminder':
      return `${message.data.activity_name}: ${message.data.message}`
    default:
      return JSON.stringify(message.data)
  }
}

// 格式化消息时间
const formatMessageTime = (timestamp: string) => {
  return formatRelativeTime(timestamp)
}

// 处理消息点击
const handleMessageClick = (message: any) => {
  switch (message.type) {
    case 'seckill_result':
      if (message.data.result.code === 'SUCCESS' && message.data.result.order_id) {
        router.push(`/user/orders/${message.data.result.order_id}`)
      } else {
        router.push(`/activity/${message.data.activity_id}`)
      }
      break
    case 'queue_update':
    case 'activity_reminder':
      router.push(`/activity/${message.data.activity_id}`)
      break
    case 'system_message':
      if (message.data.action) {
        router.push(message.data.action.url)
      }
      break
  }
  
  panelVisible.value = false
}

// 移除单条消息
const removeMessage = (index: number) => {
  // 这里可以实现移除单条消息的逻辑
  ElMessage.info('消息已移除')
}

// 标记所有消息为已读
const markAllAsRead = () => {
  markAsRead()
  ElMessage.success('所有消息已标记为已读')
}

// 清空所有历史记录
const clearAllHistory = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有消息记录吗？此操作不可恢复。',
      '清空消息',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    clearHistory()
    ElMessage.success('消息记录已清空')
  } catch (error) {
    // 用户取消
  }
}

// 跳转到消息中心
const goToMessageCenter = () => {
  router.push('/user/messages')
  panelVisible.value = false
}
</script>

<style scoped lang="scss">
.feedback-notification {
  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;

    .header-actions {
      display: flex;
      gap: 8px;
    }
  }

  .notification-content {
    height: calc(100vh - 120px);
    overflow-y: auto;

    .empty-state {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 200px;
    }

    .message-list {
      .message-item {
        display: flex;
        align-items: flex-start;
        gap: 12px;
        padding: 16px;
        border-bottom: 1px solid var(--el-border-color-lighter);
        cursor: pointer;
        transition: background-color 0.3s ease;

        &:hover {
          background-color: var(--el-bg-color-page);
        }

        &:last-child {
          border-bottom: none;
        }

        .message-icon {
          flex-shrink: 0;
          width: 32px;
          height: 32px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 16px;
          
          &.success {
            background-color: var(--el-color-success-light-9);
            color: var(--el-color-success);
          }
          
          &.error {
            background-color: var(--el-color-danger-light-9);
            color: var(--el-color-danger);
          }
          
          &.warning {
            background-color: var(--el-color-warning-light-9);
            color: var(--el-color-warning);
          }
          
          &.info {
            background-color: var(--el-color-info-light-9);
            color: var(--el-color-info);
          }
        }

        .message-content {
          flex: 1;
          min-width: 0;

          .message-title {
            font-size: 14px;
            font-weight: 600;
            color: var(--el-text-color-primary);
            margin-bottom: 4px;
          }

          .message-text {
            font-size: 13px;
            color: var(--el-text-color-regular);
            line-height: 1.4;
            margin-bottom: 4px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }

          .message-time {
            font-size: 12px;
            color: var(--el-text-color-placeholder);
          }
        }

        .message-actions {
          flex-shrink: 0;
          opacity: 0;
          transition: opacity 0.3s ease;

          .el-button {
            width: 24px;
            height: 24px;
          }
        }

        &:hover .message-actions {
          opacity: 1;
        }

        // 消息类型样式
        &.success {
          border-left: 3px solid var(--el-color-success);
        }

        &.error {
          border-left: 3px solid var(--el-color-danger);
        }

        &.warning {
          border-left: 3px solid var(--el-color-warning);
        }

        &.info {
          border-left: 3px solid var(--el-color-info);
        }
      }
    }
  }

  .panel-footer {
    display: flex;
    justify-content: space-between;
    gap: 12px;
  }
}
</style>
