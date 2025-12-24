import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { servicesApi, statusApi, type ServiceMetadata, type ServiceStatus } from '@/api'
import { ws } from '@/utils/websocket'

export const useServicesStore = defineStore('services', () => {
  // State
  const services = ref<ServiceMetadata[]>([])
  const servicesStatus = ref<Map<string, ServiceStatus>>(new Map())
  const loading = ref(false)

  // Getters
  const onlineServices = computed(() => {
    return services.value.filter(s => {
      const status = servicesStatus.value.get(s.name)
      return status?.online
    })
  })

  const offlineServices = computed(() => {
    return services.value.filter(s => {
      const status = servicesStatus.value.get(s.name)
      return status && !status.online
    })
  })

  // Actions
  async function loadServices() {
    loading.value = true
    try {
      services.value = await servicesApi.list()
    } catch (error: any) {
      console.error('加载服务列表失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  async function loadStatus() {
    try {
      const statuses = await statusApi.list()
      const statusMap = new Map<string, ServiceStatus>()
      statuses.forEach(s => {
        statusMap.set(s.service_name, s)
      })
      servicesStatus.value = statusMap
    } catch (error: any) {
      console.error('加载服务状态失败:', error)
      throw error
    }
  }

  function setupWebSocket() {
    // 监听状态更新
    ws.on('status', (event: any) => {
      if (event.service && event.service_name) {
        servicesStatus.value.set(event.service_name, event)
      }
    })

    // 监听服务事件
    ws.on('events', (event: any) => {
      if (event.type === 'online' || event.type === 'offline') {
        // 重新加载状态
        loadStatus()
      }
      if (event.type === 'registered' || event.type === 'updated') {
        // 重新加载服务列表
        loadServices()
      }
    })

    // 订阅频道
    ws.subscribe(['status', 'events'])
  }

  function getService(name: string) {
    return services.value.find(s => s.name === name)
  }

  function getServiceStatus(name: string) {
    return servicesStatus.value.get(name)
  }

  // 初始化
  function init() {
    loadServices()
    loadStatus()
    setupWebSocket()
  }

  return {
    services,
    servicesStatus,
    loading,
    onlineServices,
    offlineServices,
    loadServices,
    loadStatus,
    setupWebSocket,
    getService,
    getServiceStatus,
    init
  }
})
