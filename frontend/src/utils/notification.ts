import { ElNotification } from 'element-plus'
import type { NotificationOptions } from 'element-plus'

// 通知类型
export type NotificationType = 'success' | 'warning' | 'info' | 'error'

// 通知配置
export interface NotificationConfig extends Partial<NotificationOptions> {
  id?: string
  persistent?: boolean // 是否持久显示
  sound?: boolean // 是否播放声音
  vibrate?: boolean // 是否震动（移动端）
  action?: {
    text: string
    handler: () => void
  }
}

// 通知管理器
class NotificationManager {
  private activeNotifications = new Map<string, any>()
  private soundEnabled = true
  private vibrationEnabled = true
  private notificationPermission: NotificationPermission = 'default'

  constructor() {
    this.initBrowserNotification()
  }

  // 初始化浏览器通知权限
  private async initBrowserNotification() {
    if ('Notification' in window) {
      this.notificationPermission = Notification.permission
      
      if (this.notificationPermission === 'default') {
        try {
          this.notificationPermission = await Notification.requestPermission()
        } catch (error) {
          console.warn('无法请求通知权限:', error)
        }
      }
    }
  }

  // 显示应用内通知
  showNotification(config: NotificationConfig) {
    const {
      id = `notification_${Date.now()}`,
      title = '通知',
      message = '',
      type = 'info',
      duration = 4500,
      persistent = false,
      sound = false,
      vibrate = false,
      action,
      ...options
    } = config

    // 如果是持久通知，设置duration为0
    const finalDuration = persistent ? 0 : duration

    // 创建通知
    const notification = ElNotification({
      title,
      message,
      type,
      duration: finalDuration,
      showClose: true,
      onClick: action?.handler,
      onClose: () => {
        this.activeNotifications.delete(id)
      },
      ...options
    })

    // 保存通知引用
    this.activeNotifications.set(id, notification)

    // 播放声音
    if (sound && this.soundEnabled) {
      this.playNotificationSound(type as NotificationType)
    }

    // 震动
    if (vibrate && this.vibrationEnabled) {
      this.vibrate()
    }

    return id
  }

  // 显示浏览器原生通知
  showBrowserNotification(config: {
    title: string
    body?: string
    icon?: string
    tag?: string
    onClick?: () => void
  }) {
    if (this.notificationPermission !== 'granted') {
      console.warn('浏览器通知权限未授予')
      return
    }

    const { title, body, icon, tag, onClick } = config

    const notification = new Notification(title, {
      body,
      icon: icon || '/favicon.ico',
      tag,
      requireInteraction: false,
    })

    if (onClick) {
      notification.onclick = onClick
    }

    // 自动关闭
    setTimeout(() => {
      notification.close()
    }, 5000)

    return notification
  }

  // 显示成功通知
  success(title: string, message?: string, config?: Partial<NotificationConfig>) {
    return this.showNotification({
      title,
      message,
      type: 'success',
      sound: true,
      ...config
    })
  }

  // 显示警告通知
  warning(title: string, message?: string, config?: Partial<NotificationConfig>) {
    return this.showNotification({
      title,
      message,
      type: 'warning',
      sound: true,
      ...config
    })
  }

  // 显示信息通知
  info(title: string, message?: string, config?: Partial<NotificationConfig>) {
    return this.showNotification({
      title,
      message,
      type: 'info',
      ...config
    })
  }

  // 显示错误通知
  error(title: string, message?: string, config?: Partial<NotificationConfig>) {
    return this.showNotification({
      title,
      message,
      type: 'error',
      persistent: true,
      sound: true,
      vibrate: true,
      ...config
    })
  }

  // 显示活动开始通知
  activityStarted(activityName: string, activityId: number) {
    return this.showNotification({
      id: `activity_start_${activityId}`,
      title: '活动开始',
      message: `${activityName} 已开始，快来抢购吧！`,
      type: 'success',
      duration: 8000,
      sound: true,
      vibrate: true,
      action: {
        text: '立即查看',
        handler: () => {
          window.location.href = `/activity/${activityId}`
        }
      }
    })
  }

  // 显示活动即将结束通知
  activityEndingSoon(activityName: string, activityId: number, remainingTime: number) {
    const minutes = Math.floor(remainingTime / 60)
    return this.showNotification({
      id: `activity_ending_${activityId}`,
      title: '活动即将结束',
      message: `${activityName} 还有 ${minutes} 分钟结束，抓紧时间！`,
      type: 'warning',
      duration: 10000,
      sound: true,
      action: {
        text: '立即查看',
        handler: () => {
          window.location.href = `/activity/${activityId}`
        }
      }
    })
  }

  // 显示库存告急通知
  lowStock(activityName: string, activityId: number, remainingStock: number) {
    return this.showNotification({
      id: `low_stock_${activityId}`,
      title: '库存告急',
      message: `${activityName} 仅剩 ${remainingStock} 件，手慢无！`,
      type: 'warning',
      duration: 6000,
      sound: true,
      vibrate: true,
      action: {
        text: '立即抢购',
        handler: () => {
          window.location.href = `/activity/${activityId}`
        }
      }
    })
  }

  // 显示系统维护通知
  systemMaintenance(message: string, startTime?: string) {
    return this.showNotification({
      id: 'system_maintenance',
      title: '系统维护通知',
      message: startTime ? `${message}，维护时间：${startTime}` : message,
      type: 'warning',
      persistent: true,
      sound: true
    })
  }

  // 关闭特定通知
  close(id: string) {
    const notification = this.activeNotifications.get(id)
    if (notification) {
      notification.close()
      this.activeNotifications.delete(id)
    }
  }

  // 关闭所有通知
  closeAll() {
    this.activeNotifications.forEach(notification => {
      notification.close()
    })
    this.activeNotifications.clear()
  }

  // 播放通知声音
  private playNotificationSound(type: NotificationType) {
    try {
      const audio = new Audio()
      
      switch (type) {
        case 'success':
          audio.src = '/sounds/success.mp3'
          break
        case 'warning':
          audio.src = '/sounds/warning.mp3'
          break
        case 'error':
          audio.src = '/sounds/error.mp3'
          break
        default:
          audio.src = '/sounds/notification.mp3'
      }
      
      audio.volume = 0.3
      audio.play().catch(() => {
        // 忽略播放失败（可能是用户未交互）
      })
    } catch (error) {
      // 忽略音频播放错误
    }
  }

  // 震动
  private vibrate() {
    if ('vibrate' in navigator) {
      navigator.vibrate([200, 100, 200])
    }
  }

  // 设置声音开关
  setSoundEnabled(enabled: boolean) {
    this.soundEnabled = enabled
    localStorage.setItem('notification_sound', enabled.toString())
  }

  // 设置震动开关
  setVibrationEnabled(enabled: boolean) {
    this.vibrationEnabled = enabled
    localStorage.setItem('notification_vibration', enabled.toString())
  }

  // 获取设置
  getSettings() {
    return {
      soundEnabled: this.soundEnabled,
      vibrationEnabled: this.vibrationEnabled,
      browserPermission: this.notificationPermission
    }
  }

  // 初始化设置
  initSettings() {
    const soundSetting = localStorage.getItem('notification_sound')
    if (soundSetting !== null) {
      this.soundEnabled = soundSetting === 'true'
    }

    const vibrationSetting = localStorage.getItem('notification_vibration')
    if (vibrationSetting !== null) {
      this.vibrationEnabled = vibrationSetting === 'true'
    }
  }
}

// 创建全局通知管理器实例
export const notificationManager = new NotificationManager()

// 初始化设置
notificationManager.initSettings()

// 导出便捷方法
export const notify = {
  success: (title: string, message?: string, config?: Partial<NotificationConfig>) =>
    notificationManager.success(title, message, config),
  
  warning: (title: string, message?: string, config?: Partial<NotificationConfig>) =>
    notificationManager.warning(title, message, config),
  
  info: (title: string, message?: string, config?: Partial<NotificationConfig>) =>
    notificationManager.info(title, message, config),
  
  error: (title: string, message?: string, config?: Partial<NotificationConfig>) =>
    notificationManager.error(title, message, config),
  
  activityStarted: (activityName: string, activityId: number) =>
    notificationManager.activityStarted(activityName, activityId),
  
  activityEndingSoon: (activityName: string, activityId: number, remainingTime: number) =>
    notificationManager.activityEndingSoon(activityName, activityId, remainingTime),
  
  lowStock: (activityName: string, activityId: number, remainingStock: number) =>
    notificationManager.lowStock(activityName, activityId, remainingStock),
  
  systemMaintenance: (message: string, startTime?: string) =>
    notificationManager.systemMaintenance(message, startTime),
}
