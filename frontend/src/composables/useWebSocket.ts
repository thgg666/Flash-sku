import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { 
  WebSocketManager, 
  WebSocketState, 
  createWebSocketManager,
  type WebSocketMessage,
  type WebSocketConfig 
} from '@/utils/websocket'

// WebSocket连接配置
const getWebSocketConfig = (): WebSocketConfig => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = import.meta.env.VITE_WS_HOST || window.location.host
  const path = import.meta.env.VITE_WS_PATH || '/ws/'
  
  return {
    url: `${protocol}//${host}${path}`,
    reconnectInterval: 3000,
    maxReconnectAttempts: 5,
    heartbeatInterval: 30000,
    debug: import.meta.env.DEV
  }
}

// 全局WebSocket管理器实例
let globalWebSocketManager: WebSocketManager | null = null

/**
 * WebSocket组合式函数
 */
export function useWebSocket() {
  const authStore = useAuthStore()
  
  // 状态
  const state = ref<WebSocketState>(WebSocketState.DISCONNECTED)
  const isConnected = computed(() => state.value === WebSocketState.CONNECTED)
  const isConnecting = computed(() => state.value === WebSocketState.CONNECTING)
  const isReconnecting = computed(() => state.value === WebSocketState.RECONNECTING)
  const hasError = computed(() => state.value === WebSocketState.ERROR)
  
  // 消息监听器
  const messageListeners = new Map<string, Set<(data: any) => void>>()
  
  // 获取或创建WebSocket管理器
  const getManager = (): WebSocketManager => {
    if (!globalWebSocketManager) {
      const config = getWebSocketConfig()
      globalWebSocketManager = createWebSocketManager(config)
      
      // 注册全局事件监听器
      globalWebSocketManager.on({
        onStateChange: (newState) => {
          state.value = newState
        },
        onMessage: (message) => {
          handleMessage(message)
        },
        onError: () => {
          ElMessage.error('WebSocket连接异常')
        },
        onClose: (event) => {
          if (event.code !== 1000) { // 非正常关闭
            console.warn('WebSocket异常关闭:', event.code, event.reason)
          }
        }
      })
    }
    
    return globalWebSocketManager
  }
  
  // 连接WebSocket
  const connect = async (): Promise<void> => {
    if (!authStore.isAuthenticated) {
      console.warn('用户未登录，无法建立WebSocket连接')
      return
    }
    
    try {
      const manager = getManager()
      await manager.connect()
      
      // 发送认证消息
      manager.send({
        type: 'auth',
        data: {
          token: authStore.token,
          user_id: authStore.user?.id
        }
      })
      
      console.log('WebSocket连接成功')
    } catch (error) {
      console.error('WebSocket连接失败:', error)
      ElMessage.error('实时连接失败，部分功能可能受影响')
    }
  }
  
  // 断开连接
  const disconnect = (): void => {
    if (globalWebSocketManager) {
      globalWebSocketManager.disconnect()
    }
  }
  
  // 发送消息
  const send = (type: string, data: any): boolean => {
    const manager = getManager()
    return manager.send({ type, data })
  }
  
  // 订阅消息
  const subscribe = (messageType: string, callback: (data: any) => void): void => {
    if (!messageListeners.has(messageType)) {
      messageListeners.set(messageType, new Set())
    }
    messageListeners.get(messageType)!.add(callback)
  }
  
  // 取消订阅
  const unsubscribe = (messageType: string, callback?: (data: any) => void): void => {
    const listeners = messageListeners.get(messageType)
    if (listeners) {
      if (callback) {
        listeners.delete(callback)
      } else {
        listeners.clear()
      }
    }
  }
  
  // 处理接收到的消息
  const handleMessage = (message: WebSocketMessage): void => {
    const listeners = messageListeners.get(message.type)
    if (listeners) {
      listeners.forEach(callback => {
        try {
          callback(message.data)
        } catch (error) {
          console.error('消息处理错误:', error)
        }
      })
    }
  }
  
  // 重连
  const reconnect = async (): Promise<void> => {
    disconnect()
    await new Promise(resolve => setTimeout(resolve, 1000))
    await connect()
  }
  
  return {
    // 状态
    state: computed(() => state.value),
    isConnected,
    isConnecting,
    isReconnecting,
    hasError,
    
    // 方法
    connect,
    disconnect,
    reconnect,
    send,
    subscribe,
    unsubscribe,
  }
}

/**
 * 特定消息类型的WebSocket Hook
 */
export function useWebSocketMessage<T = any>(messageType: string) {
  const { subscribe, unsubscribe, send, isConnected } = useWebSocket()
  
  const data = ref<T | null>(null)
  const lastMessage = ref<WebSocketMessage | null>(null)
  
  // 消息处理函数
  const handleMessage = (messageData: T) => {
    data.value = messageData
    lastMessage.value = {
      type: messageType,
      data: messageData,
      timestamp: Date.now()
    }
  }
  
  // 发送特定类型的消息
  const sendMessage = (messageData: any): boolean => {
    return send(messageType, messageData)
  }
  
  // 组件挂载时订阅
  onMounted(() => {
    subscribe(messageType, handleMessage)
  })
  
  // 组件卸载时取消订阅
  onUnmounted(() => {
    unsubscribe(messageType, handleMessage)
  })
  
  return {
    data: computed(() => data.value),
    lastMessage: computed(() => lastMessage.value),
    sendMessage,
    isConnected,
  }
}

/**
 * 自动连接的WebSocket Hook
 */
export function useAutoWebSocket() {
  const webSocket = useWebSocket()
  const authStore = useAuthStore()
  
  // 监听认证状态变化
  const handleAuthChange = () => {
    if (authStore.isAuthenticated) {
      webSocket.connect()
    } else {
      webSocket.disconnect()
    }
  }
  
  // 组件挂载时自动连接
  onMounted(() => {
    if (authStore.isAuthenticated) {
      webSocket.connect()
    }
    
    // 监听认证状态变化
    authStore.$subscribe(() => {
      handleAuthChange()
    })
  })
  
  // 组件卸载时断开连接
  onUnmounted(() => {
    // 注意：这里不直接断开连接，因为可能有其他组件在使用
    // 实际的断开连接应该在用户登出时进行
  })
  
  return webSocket
}

// 清理全局WebSocket连接
export function cleanupWebSocket(): void {
  if (globalWebSocketManager) {
    globalWebSocketManager.destroy()
    globalWebSocketManager = null
  }
}
