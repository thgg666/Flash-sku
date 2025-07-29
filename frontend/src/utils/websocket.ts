import { ElMessage } from 'element-plus'

// WebSocket连接状态
export enum WebSocketState {
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected',
  RECONNECTING = 'reconnecting',
  ERROR = 'error'
}

// WebSocket消息类型
export interface WebSocketMessage {
  type: string
  data: any
  timestamp: number
}

// WebSocket事件类型
export interface WebSocketEvents {
  onOpen?: () => void
  onClose?: (event: CloseEvent) => void
  onError?: (error: Event) => void
  onMessage?: (message: WebSocketMessage) => void
  onStateChange?: (state: WebSocketState) => void
}

// WebSocket配置
export interface WebSocketConfig {
  url: string
  protocols?: string[]
  reconnectInterval?: number
  maxReconnectAttempts?: number
  heartbeatInterval?: number
  heartbeatMessage?: string
  debug?: boolean
}

export class WebSocketManager {
  private ws: WebSocket | null = null
  private config: Required<WebSocketConfig>
  private events: WebSocketEvents = {}
  private state: WebSocketState = WebSocketState.DISCONNECTED
  private reconnectAttempts = 0
  private reconnectTimer: number | null = null
  private heartbeatTimer: number | null = null
  private messageQueue: WebSocketMessage[] = []
  private isManualClose = false

  constructor(config: WebSocketConfig) {
    this.config = {
      protocols: [],
      reconnectInterval: 3000,
      maxReconnectAttempts: 5,
      heartbeatInterval: 30000,
      heartbeatMessage: JSON.stringify({ type: 'ping' }),
      debug: false,
      ...config
    }
  }

  // 连接WebSocket
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      this.isManualClose = false
      this.setState(WebSocketState.CONNECTING)
      
      try {
        this.ws = new WebSocket(this.config.url, this.config.protocols)
        
        this.ws.onopen = (event) => {
          this.handleOpen(event)
          resolve()
        }
        
        this.ws.onclose = (event) => {
          this.handleClose(event)
          if (this.state === WebSocketState.CONNECTING) {
            reject(new Error('WebSocket连接失败'))
          }
        }
        
        this.ws.onerror = (error) => {
          this.handleError(error)
          reject(error)
        }
        
        this.ws.onmessage = (event) => {
          this.handleMessage(event)
        }
      } catch (error) {
        this.setState(WebSocketState.ERROR)
        reject(error)
      }
    })
  }

  // 断开连接
  disconnect(): void {
    this.isManualClose = true
    this.clearTimers()
    
    if (this.ws) {
      this.ws.close(1000, 'Manual disconnect')
      this.ws = null
    }
    
    this.setState(WebSocketState.DISCONNECTED)
  }

  // 发送消息
  send(message: Omit<WebSocketMessage, 'timestamp'>): boolean {
    const fullMessage: WebSocketMessage = {
      ...message,
      timestamp: Date.now()
    }

    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify(fullMessage))
        this.log('发送消息:', fullMessage)
        return true
      } catch (error) {
        this.log('发送消息失败:', error)
        return false
      }
    } else {
      // 连接未建立时，将消息加入队列
      this.messageQueue.push(fullMessage)
      this.log('消息已加入队列:', fullMessage)
      return false
    }
  }

  // 注册事件监听器
  on(events: WebSocketEvents): void {
    this.events = { ...this.events, ...events }
  }

  // 移除事件监听器
  off(eventType?: keyof WebSocketEvents): void {
    if (eventType) {
      delete this.events[eventType]
    } else {
      this.events = {}
    }
  }

  // 获取当前状态
  getState(): WebSocketState {
    return this.state
  }

  // 获取连接状态
  isConnected(): boolean {
    return this.state === WebSocketState.CONNECTED
  }

  // 处理连接打开
  private handleOpen(_event: Event): void {
    this.setState(WebSocketState.CONNECTED)
    this.reconnectAttempts = 0
    this.startHeartbeat()
    this.processMessageQueue()
    
    this.log('WebSocket连接已建立')
    this.events.onOpen?.()
  }

  // 处理连接关闭
  private handleClose(event: CloseEvent): void {
    this.clearTimers()
    
    if (!this.isManualClose && this.shouldReconnect()) {
      this.setState(WebSocketState.RECONNECTING)
      this.scheduleReconnect()
    } else {
      this.setState(WebSocketState.DISCONNECTED)
    }
    
    this.log('WebSocket连接已关闭:', event.code, event.reason)
    this.events.onClose?.(event)
  }

  // 处理连接错误
  private handleError(error: Event): void {
    this.setState(WebSocketState.ERROR)
    this.log('WebSocket连接错误:', error)
    this.events.onError?.(error)
  }

  // 处理接收消息
  private handleMessage(event: MessageEvent): void {
    try {
      const message: WebSocketMessage = JSON.parse(event.data)
      this.log('接收消息:', message)
      
      // 处理心跳响应
      if (message.type === 'pong') {
        this.log('收到心跳响应')
        return
      }
      
      this.events.onMessage?.(message)
    } catch (error) {
      this.log('解析消息失败:', error)
    }
  }

  // 设置状态
  private setState(state: WebSocketState): void {
    if (this.state !== state) {
      this.state = state
      this.log('状态变更:', state)
      this.events.onStateChange?.(state)
    }
  }

  // 判断是否应该重连
  private shouldReconnect(): boolean {
    return this.reconnectAttempts < this.config.maxReconnectAttempts
  }

  // 安排重连
  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }
    
    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++
      this.log(`尝试重连 (${this.reconnectAttempts}/${this.config.maxReconnectAttempts})`)
      
      this.connect().catch(() => {
        if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
          this.setState(WebSocketState.ERROR)
          ElMessage.error('WebSocket连接失败，请刷新页面重试')
        }
      })
    }, this.config.reconnectInterval)
  }

  // 开始心跳
  private startHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
    }
    
    this.heartbeatTimer = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(this.config.heartbeatMessage)
        this.log('发送心跳')
      }
    }, this.config.heartbeatInterval)
  }

  // 处理消息队列
  private processMessageQueue(): void {
    while (this.messageQueue.length > 0) {
      const message = this.messageQueue.shift()
      if (message) {
        this.send(message)
      }
    }
  }

  // 清理定时器
  private clearTimers(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  // 日志输出
  private log(...args: any[]): void {
    if (this.config.debug) {
      console.log('[WebSocket]', ...args)
    }
  }

  // 销毁实例
  destroy(): void {
    this.disconnect()
    this.off()
    this.messageQueue = []
  }
}

// 创建WebSocket管理器实例
export function createWebSocketManager(config: WebSocketConfig): WebSocketManager {
  return new WebSocketManager(config)
}
