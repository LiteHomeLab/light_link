import type { ServiceEvent } from '@/api'

// WebSocket 消息类型
export interface WSMessage {
  channel: string
  event: ServiceEvent | any
}

// 订阅消息类型
export interface SubscribeMessage {
  action: string
  channels: string[]
}

// WebSocket 事件处理器类型
type WSEventHandler = (event: any) => void

// WebSocket 客户端类
export class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private channels: Set<string> = new Set()
  private handlers: Map<string, WSEventHandler[]> = new Map()
  private reconnectTimer: number | null = null
  private reconnectDelay: number = 3000
  private manualClose: boolean = false

  constructor(url: string) {
    this.url = url
  }

  // 连接 WebSocket
  connect(): void {
    const token = localStorage.getItem('token')
    const wsUrl = token ? `${this.url}?token=${token}` : this.url

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('[WebSocket] 已连接')
      this.manualClose = false

      // 重新订阅之前的频道
      if (this.channels.size > 0) {
        this.sendSubscription()
      }
    }

    this.ws.onmessage = (event) => {
      try {
        const message: WSMessage = JSON.parse(event.data)
        const handlers = this.handlers.get(message.channel) || []
        handlers.forEach((handler) => {
          try {
            handler(message.event)
          } catch (err) {
            console.error('[WebSocket] 事件处理器错误:', err)
          }
        })
      } catch (e) {
        console.error('[WebSocket] 解析消息错误:', e)
      }
    }

    this.ws.onclose = () => {
      console.log('[WebSocket] 连接关闭')
      if (!this.manualClose) {
        this.reconnect()
      }
    }

    this.ws.onerror = (error) => {
      console.error('[WebSocket] 错误:', error)
    }
  }

  // 断开连接
  disconnect(): void {
    this.manualClose = true
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  // 重连
  private reconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }
    this.reconnectTimer = window.setTimeout(() => {
      console.log('[WebSocket] 重连中...')
      this.connect()
    }, this.reconnectDelay)
  }

  // 订阅频道
  subscribe(channels: string[]): void {
    channels.forEach((ch) => this.channels.add(ch))
    this.sendSubscription()
  }

  // 取消订阅
  unsubscribe(channels: string[]): void {
    channels.forEach((ch) => this.channels.delete(ch))
    this.sendSubscription()
  }

  // 发送订阅消息
  private sendSubscription(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const message: SubscribeMessage = {
        action: 'subscribe',
        channels: Array.from(this.channels)
      }
      this.ws.send(JSON.stringify(message))
    }
  }

  // 监听频道事件
  on(channel: string, handler: WSEventHandler): void {
    if (!this.handlers.has(channel)) {
      this.handlers.set(channel, [])
    }
    this.handlers.get(channel)!.push(handler)
  }

  // 取消监听频道事件
  off(channel: string, handler: WSEventHandler): void {
    const handlers = this.handlers.get(channel)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  // 获取连接状态
  get readyState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED
  }

  // 是否已连接
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

// 创建全局 WebSocket 实例
export const ws = new WebSocketClient(`ws://${location.host}/api/ws`)

// 自动连接
if (import.meta.env.DEV) {
  // 开发环境自动连接
  // 生产环境在用户登录后连接
}
