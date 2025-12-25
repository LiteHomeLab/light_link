import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { instancesApi, type Instance } from '@/api'
import { ElMessage } from 'element-plus'

export const useInstancesStore = defineStore('instances', () => {
  // State
  const instances = ref<Instance[]>([])
  const loading = ref(false)

  // Getters
  // 按服务名称分组实例
  const instancesByService = computed(() => {
    const grouped = new Map<string, Instance[]>()
    instances.value.forEach(inst => {
      if (!grouped.has(inst.service_name)) {
        grouped.set(inst.service_name, [])
      }
      grouped.get(inst.service_name)!.push(inst)
    })
    return grouped
  })

  // 获取指定服务的实例
  function getInstancesByService(serviceName: string) {
    return instancesByService.value.get(serviceName) || []
  }

  // 获取指定服务的在线实例数量
  function getOnlineCount(serviceName: string) {
    return getInstancesByService(serviceName)
      .filter(inst => inst.online).length
  }

  // 获取指定服务的离线实例数量
  function getOfflineCount(serviceName: string) {
    return getInstancesByService(serviceName)
      .filter(inst => !inst.online).length
  }

  // 获取指定服务的实例总数
  function getTotalCount(serviceName: string) {
    return getInstancesByService(serviceName).length
  }

  // Actions
  // 加载所有实例
  async function loadInstances() {
    loading.value = true
    try {
      instances.value = await instancesApi.list()
    } catch (error: any) {
      ElMessage.error('加载实例列表失败')
      console.error(error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 加载指定服务的实例
  async function loadServiceInstances(serviceName: string) {
    loading.value = true
    try {
      const serviceInstances = await instancesApi.list({ service: serviceName })

      // 移除该服务的旧实例，添加新实例
      instances.value = instances.value.filter(
        inst => inst.service_name !== serviceName
      )
      instances.value.push(...serviceInstances)
    } catch (error: any) {
      ElMessage.error(`加载服务 ${serviceName} 的实例失败`)
      console.error(error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 停止实例
  async function stopInstance(instanceKey: string) {
    try {
      const result = await instancesApi.stop(instanceKey)
      ElMessage.success(`实例正在停止...`)

      // 更新本地状态
      const inst = instances.value.find(i => i.instance_key === instanceKey)
      if (inst) {
        inst.online = false
      }

      return result
    } catch (error: any) {
      const msg = error.response?.data?.error || '停止实例失败'
      ElMessage.error(msg)
      throw error
    }
  }

  // 重启实例
  async function restartInstance(instanceKey: string) {
    try {
      const result = await instancesApi.restart(instanceKey)
      ElMessage.success(`实例正在重启...`)
      return result
    } catch (error: any) {
      const msg = error.response?.data?.error || '重启实例失败'
      ElMessage.error(msg)
      throw error
    }
  }

  // 删除离线实例
  async function deleteInstance(instanceKey: string) {
    try {
      const result = await instancesApi.delete(instanceKey)
      ElMessage.success(`实例已删除`)

      // 从本地状态移除
      const index = instances.value.findIndex(i => i.instance_key === instanceKey)
      if (index > -1) {
        instances.value.splice(index, 1)
      }

      return result
    } catch (error: any) {
      const msg = error.response?.data?.error || '删除实例失败'
      ElMessage.error(msg)
      throw error
    }
  }

  return {
    instances,
    loading,
    instancesByService,
    getInstancesByService,
    getOnlineCount,
    getOfflineCount,
    getTotalCount,
    loadInstances,
    loadServiceInstances,
    stopInstance,
    restartInstance,
    deleteInstance
  }
})
