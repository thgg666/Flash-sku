import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useWebSocketMessage } from '@/composables/useWebSocket'
import { useAuthStore } from '@/stores/auth'
import { notificationManager, notify } from '@/utils/notification'
import { formatDateTime } from '@/utils'

// 系统通知消息类型
interface SystemNotificationMessage {
  type: 'system_announcement' | 'maintenance_notice' | 'activity_reminder' | 'promotion_alert'
  title: string
  content: string
  level: 'info' | 'warning' | 'error' | 'success'
  target_users?: number[] // 目标用户ID列表，空表示全体用户
  start_time?: string
  end_time?: string
  action_url?: string
  action_text?: string
  persistent?: boolean
  sound?: boolean
  vibrate?: boolean
  id: string
  timestamp: string
}

// 活动提醒消息
interface ActivityReminderMessage {
  activity_id: number
  activity_name: string
  reminder_type: 'starting_soon' | 'ending_soon' | 'low_stock' | 'sold_out'
  message: string
  time_remaining?: number
  stock_remaining?: number
  timestamp: string
}

/**
 * 系统通知管理
 */
export function useNotificationSystem() {
  const authStore = useAuthStore()
  
  // WebSocket消息监听
  const { 
    data: systemNotification, 
    sendMessage: sendSystemMessage, 
    isConnected: isSystemConnected 
  } = useWebSocketMessage<SystemNotificationMessage>('system_notification')
  
  const { 
    data: activityReminder, 
    sendMessage: sendReminderMessage, 
    isConnected: isReminderConnected 
  } = useWebSocketMessage<ActivityReminderMessage>('activity_reminder')
  
  // 状态
  const isSubscribed = ref(false)
  const notificationHistory = ref<SystemNotificationMessage[]>([])
  const reminderHistory = ref<ActivityReminderMessage[]>([])
  const settings = ref({
    soundEnabled: true,
    vibrationEnabled: true,
    browserNotificationEnabled: false,
    activityRemindersEnabled: true,
    systemAnnouncementsEnabled: true,
    maintenanceNoticesEnabled: true,
  })
  
  // 计算属性
  const isConnected = computed(() => isSystemConnected.value && isReminderConnected.value)
  
  const activeNotifications = computed(() => {
    const now = new Date()
    return notificationHistory.value.filter(notification => {
      if (!notification.end_time) return true
      return new Date(notification.end_time) > now
    })
  })
  
  const recentReminders = computed(() => {
    return reminderHistory.value.slice(-10).reverse()
  })
  
  // 订阅系统通知
  const subscribe = () => {
    if (!isConnected.value) {
      console.warn('WebSocket未连接，无法订阅系统通知')
      return
    }
    
    // 订阅系统通知
    sendSystemMessage({
      action: 'subscribe_system_notifications',
      user_id: authStore.user?.id
    })
    
    // 订阅活动提醒
    sendReminderMessage({
      action: 'subscribe_activity_reminders',
      user_id: authStore.user?.id
    })
    
    isSubscribed.value = true
    console.log('已订阅系统通知和活动提醒')
  }
  
  // 取消订阅
  const unsubscribe = () => {
    if (!isConnected.value) return
    
    sendSystemMessage({
      action: 'unsubscribe_system_notifications',
      user_id: authStore.user?.id
    })
    
    sendReminderMessage({
      action: 'unsubscribe_activity_reminders',
      user_id: authStore.user?.id
    })
    
    isSubscribed.value = false
    console.log('已取消订阅系统通知和活动提醒')
  }
  
  // 处理系统通知
  const handleSystemNotification = (notification: SystemNotificationMessage) => {
    // 检查是否为目标用户
    if (notification.target_users && notification.target_users.length > 0) {
      if (!authStore.user?.id || !notification.target_users.includes(authStore.user.id)) {
        return
      }
    }
    
    // 检查用户设置
    if (!shouldShowNotification(notification.type)) {
      return
    }
    
    // 添加到历史记录
    notificationHistory.value.push(notification)
    
    // 只保留最近50条记录
    if (notificationHistory.value.length > 50) {
      notificationHistory.value = notificationHistory.value.slice(-50)
    }
    
    // 显示通知
    showSystemNotification(notification)
  }
  
  // 处理活动提醒
  const handleActivityReminder = (reminder: ActivityReminderMessage) => {
    if (!settings.value.activityRemindersEnabled) {
      return
    }
    
    // 添加到历史记录
    reminderHistory.value.push(reminder)
    
    // 只保留最近20条记录
    if (reminderHistory.value.length > 20) {
      reminderHistory.value = reminderHistory.value.slice(-20)
    }
    
    // 显示提醒
    showActivityReminder(reminder)
  }
  
  // 显示系统通知
  const showSystemNotification = (notification: SystemNotificationMessage) => {
    const config = {
      id: notification.id,
      persistent: notification.persistent,
      sound: notification.sound && settings.value.soundEnabled,
      vibrate: notification.vibrate && settings.value.vibrationEnabled,
      action: notification.action_url ? {
        text: notification.action_text || '查看详情',
        handler: () => {
          window.location.href = notification.action_url!
        }
      } : undefined
    }
    
    switch (notification.level) {
      case 'success':
        notify.success(notification.title, notification.content, config)
        break
      case 'warning':
        notify.warning(notification.title, notification.content, config)
        break
      case 'error':
        notify.error(notification.title, notification.content, config)
        break
      default:
        notify.info(notification.title, notification.content, config)
    }
    
    // 浏览器原生通知
    if (settings.value.browserNotificationEnabled) {
      notificationManager.showBrowserNotification({
        title: notification.title,
        body: notification.content,
        tag: notification.id,
        onClick: config.action?.handler
      })
    }
  }
  
  // 显示活动提醒
  const showActivityReminder = (reminder: ActivityReminderMessage) => {
    const { activity_name, reminder_type, message, activity_id } = reminder
    
    switch (reminder_type) {
      case 'starting_soon':
        notify.activityStarted(activity_name, activity_id)
        break
      case 'ending_soon':
        if (reminder.time_remaining) {
          notify.activityEndingSoon(activity_name, activity_id, reminder.time_remaining)
        }
        break
      case 'low_stock':
        if (reminder.stock_remaining) {
          notify.lowStock(activity_name, activity_id, reminder.stock_remaining)
        }
        break
      case 'sold_out':
        notify.warning('商品售罄', `${activity_name} 已售罄`, {
          action: {
            text: '查看其他活动',
            handler: () => {
              window.location.href = '/activities'
            }
          }
        })
        break
    }
  }
  
  // 检查是否应该显示通知
  const shouldShowNotification = (type: string): boolean => {
    switch (type) {
      case 'system_announcement':
        return settings.value.systemAnnouncementsEnabled
      case 'maintenance_notice':
        return settings.value.maintenanceNoticesEnabled
      case 'activity_reminder':
        return settings.value.activityRemindersEnabled
      default:
        return true
    }
  }
  
  // 更新设置
  const updateSettings = (newSettings: Partial<typeof settings.value>) => {
    Object.assign(settings.value, newSettings)
    
    // 保存到localStorage
    localStorage.setItem('notification_settings', JSON.stringify(settings.value))
    
    // 更新通知管理器设置
    notificationManager.setSoundEnabled(settings.value.soundEnabled)
    notificationManager.setVibrationEnabled(settings.value.vibrationEnabled)
  }
  
  // 加载设置
  const loadSettings = () => {
    try {
      const saved = localStorage.getItem('notification_settings')
      if (saved) {
        const parsedSettings = JSON.parse(saved)
        Object.assign(settings.value, parsedSettings)
      }
    } catch (error) {
      console.warn('加载通知设置失败:', error)
    }
    
    // 同步到通知管理器
    notificationManager.setSoundEnabled(settings.value.soundEnabled)
    notificationManager.setVibrationEnabled(settings.value.vibrationEnabled)
  }
  
  // 请求浏览器通知权限
  const requestBrowserPermission = async () => {
    if ('Notification' in window) {
      const permission = await Notification.requestPermission()
      settings.value.browserNotificationEnabled = permission === 'granted'
      updateSettings({ browserNotificationEnabled: settings.value.browserNotificationEnabled })
      return permission === 'granted'
    }
    return false
  }
  
  // 清除历史记录
  const clearHistory = () => {
    notificationHistory.value = []
    reminderHistory.value = []
  }
  
  // 监听WebSocket消息
  watch(systemNotification, (notification) => {
    if (notification) {
      handleSystemNotification(notification)
    }
  })
  
  watch(activityReminder, (reminder) => {
    if (reminder) {
      handleActivityReminder(reminder)
    }
  })
  
  // 监听连接状态
  watch(isConnected, (connected) => {
    if (connected && !isSubscribed.value && authStore.isAuthenticated) {
      setTimeout(() => {
        subscribe()
      }, 1000)
    }
  })
  
  // 监听认证状态
  watch(() => authStore.isAuthenticated, (authenticated) => {
    if (authenticated && isConnected.value && !isSubscribed.value) {
      subscribe()
    } else if (!authenticated && isSubscribed.value) {
      unsubscribe()
    }
  })
  
  // 组件挂载时初始化
  onMounted(() => {
    loadSettings()
    
    if (isConnected.value && authStore.isAuthenticated) {
      subscribe()
    }
  })
  
  // 组件卸载时清理
  onUnmounted(() => {
    if (isSubscribed.value) {
      unsubscribe()
    }
  })
  
  return {
    // 状态
    isSubscribed: computed(() => isSubscribed.value),
    isConnected,
    settings: computed(() => settings.value),
    notificationHistory: computed(() => notificationHistory.value),
    reminderHistory: computed(() => reminderHistory.value),
    activeNotifications,
    recentReminders,
    
    // 方法
    subscribe,
    unsubscribe,
    updateSettings,
    requestBrowserPermission,
    clearHistory,
  }
}
