import axios from 'axios'
import type { AxiosInstance } from 'axios'

// API 客户端实例
const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器 - 添加 token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    // 直接返回 data，axios 拦截器已经处理
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token 过期或无效，跳转登录
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      localStorage.removeItem('role')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// ============ 类型定义 ============

export interface ServiceMetadata {
  id?: number
  name: string
  version: string
  description?: string
  author?: string
  tags?: string[]
  registered_at: string
  updated_at?: string
  methods?: MethodMetadata[]
}

export interface MethodMetadata {
  id?: number
  service_id?: number
  name: string
  description?: string
  parameters?: ParameterMetadata[]
  return_info?: ReturnMetadata
  examples?: ExampleMetadata[]
  tags?: string[]
  deprecated?: boolean
  created_at?: string
}

export interface ParameterMetadata {
  name: string
  type: string // string, number, boolean, array, object
  required: boolean
  description?: string
  default?: any
}

export interface ReturnMetadata {
  type: string
  description?: string
}

export interface ExampleMetadata {
  name?: string
  input?: Record<string, any>
  output?: Record<string, any>
  description?: string
}

export interface ServiceStatus {
  id?: number
  service_id?: number
  service_name: string
  online: boolean
  last_seen: string
  version: string
  updated_at: string
}

export interface ServiceEvent {
  id?: number
  type: string // online, offline, registered, updated
  service_name?: string
  service?: string
  message?: string
  metadata?: Record<string, any>
  created_at?: string
  timestamp?: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  role: string
}

export interface CallRequest {
  service: string
  method: string
  params: Record<string, any>
}

export interface CallResult {
  success: boolean
  data?: Record<string, any>
  result?: Record<string, any>
  error?: string
  duration?: number
  durationMs?: number
}

// 实例相关类型定义
export interface Instance {
  id: number
  service_name: string
  instance_key: string
  language: string          // 语言类型: go, python, csharp, javascript
  host_ip: string           // 部署服务器 IP
  host_mac: string          // 部署服务器 MAC
  working_dir: string       // 工作目录
  version: string
  first_seen: string        // ISO 8601 格式时间戳
  last_heartbeat: string    // ISO 8601 格式时间戳
  online: boolean
  created_at: string
  updated_at: string
}

export interface InstanceControlResponse {
  status: 'stopping' | 'restarting' | 'deleted'
  instance: string
}

export interface InstanceListParams {
  service?: string  // 按服务名称筛选
}

// ============ API 方法 ============

// 认证相关
export const authApi = {
  login: (data: LoginRequest) => api.post<any, LoginResponse>('/auth/login', data) as unknown as Promise<LoginResponse>
}

// 服务相关
export const servicesApi = {
  list: () => api.get<ServiceMetadata[]>('/services') as unknown as Promise<ServiceMetadata[]>,
  get: (name: string) => api.get<ServiceMetadata>(`/services/${name}`) as unknown as Promise<ServiceMetadata>,
  getMethods: (name: string) => api.get<MethodMetadata[]>(`/services/${name}/methods`) as unknown as Promise<MethodMetadata[]>,
  getMethod: (service: string, method: string) =>
    api.get<MethodMetadata>(`/services/${service}/methods/${method}`) as unknown as Promise<MethodMetadata>,
  delete: (name: string) => api.delete<{message: string}>(`/services/${name}`) as unknown as Promise<{message: string}>,
  getOpenAPI: (name: string, format: 'json' | 'yaml' = 'json') =>
    api.get(`/services/${name}/openapi?format=${format}`, {
      responseType: format === 'yaml' ? 'text' : 'json'
    })
}

// 状态相关
export const statusApi = {
  list: () => api.get<ServiceStatus[]>('/status') as unknown as Promise<ServiceStatus[]>,
  get: (name: string) => api.get<ServiceStatus>(`/status/${name}`) as unknown as Promise<ServiceStatus>
}

// 事件相关
export const eventsApi = {
  list: (limit = 100, offset = 0) =>
    api.get<ServiceEvent[]>(`/events?limit=${limit}&offset=${offset}`) as unknown as Promise<ServiceEvent[]>
}

// 调用相关
export const callApi = {
  call: (data: CallRequest) => api.post<CallResult>('/call', data) as unknown as Promise<CallResult>
}

// 实例相关 API
export const instancesApi = {
  // 获取实例列表
  list: (params?: InstanceListParams) =>
    api.get<Instance[]>('/instances', { params }) as unknown as Promise<Instance[]>,

  // 获取单个实例详情
  get: (key: string) =>
    api.get<Instance>(`/instances/${encodeURIComponent(key)}`) as unknown as Promise<Instance>,

  // 停止实例 (管理员权限)
  stop: (key: string) =>
    api.post<InstanceControlResponse>(`/instances/${encodeURIComponent(key)}/stop`) as unknown as Promise<InstanceControlResponse>,

  // 重启实例 (管理员权限)
  restart: (key: string) =>
    api.post<InstanceControlResponse>(`/instances/${encodeURIComponent(key)}/restart`) as unknown as Promise<InstanceControlResponse>,

  // 删除离线实例 (管理员权限)
  delete: (key: string) =>
    api.delete<InstanceControlResponse>(`/instances/${encodeURIComponent(key)}`) as unknown as Promise<InstanceControlResponse>
}

// WebSocket client
export { ws as defaultWebSocket } from '@/utils/websocket'

export default api
