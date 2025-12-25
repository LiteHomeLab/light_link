<template>
  <div class="service-detail-view" v-loading="loading">
    <el-page-header :title="serviceName" @back="goBack" class="header">
      <template #extra>
        <el-button @click="loadData" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </template>
    </el-page-header>

    <el-card class="status-card" v-if="service">
      <div class="status-info">
        <el-tag
          :type="serviceStatus?.online ? 'success' : 'danger'"
          size="large"
        >
          {{ serviceStatus?.online ? '在线' : '离线' }}
        </el-tag>
        <div class="status-meta">
          <span>版本: {{ service.version }}</span>
          <span>作者: {{ service.author || '未知' }}</span>
          <span>注册时间: {{ formatDate(service.registered_at) }}</span>
        </div>
      </div>
      <p v-if="service.description" class="description">{{ service.description }}</p>
      <div v-if="service.tags && service.tags.length" class="tags">
        <el-tag v-for="tag in service.tags" :key="tag" size="small">
          {{ tag }}
        </el-tag>
      </div>
    </el-card>

    <el-card class="methods-card">
      <template #header>
        <h3>方法列表</h3>
      </template>

      <el-empty v-if="!methods.length" description="暂无方法" />

      <div v-else class="methods-list">
        <div
          v-for="method in methods"
          :key="method.name"
          class="method-item"
        >
          <div class="method-header">
            <h4>{{ method.name }}</h4>
            <el-button
              type="primary"
              size="small"
              @click="goToDebug(serviceName, method.name)"
            >
              调试
            </el-button>
          </div>

          <p v-if="method.description" class="method-description">
            {{ method.description }}
          </p>

          <div class="method-meta">
            <el-tag size="small" type="info">
              返回: {{ method.return_info?.type || 'void' }}
            </el-tag>
          </div>

          <div v-if="method.parameters && method.parameters.length" class="parameters">
            <h5>参数:</h5>
            <el-table :data="method.parameters" size="small">
              <el-table-column prop="name" label="名称" width="150" />
              <el-table-column prop="type" label="类型" width="150" />
              <el-table-column prop="description" label="描述" />
              <el-table-column label="必填" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.required ? 'danger' : 'info'" size="small">
                    {{ row.required ? '是' : '否' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <div v-if="method.examples && method.examples.length" class="examples">
            <h5>示例:</h5>
            <div
              v-for="(example, index) in method.examples"
              :key="index"
              class="example-item"
            >
              <div class="example-header">
                <el-tag size="small">{{ example.name || `示例 ${index + 1}` }}</el-tag>
              </div>
              <div class="example-content">
                <div v-if="example.input" class="example-block">
                  <span class="label">输入:</span>
                  <pre>{{ formatJSON(example.input) }}</pre>
                </div>
                <div v-if="example.output" class="example-block">
                  <span class="label">输出:</span>
                  <pre>{{ formatJSON(example.output) }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 实例列表卡片 -->
    <el-card class="instances-card">
      <template #header>
        <div class="card-header">
          <h3>实例列表</h3>
          <el-button @click="loadInstances" :loading="instancesLoading" size="small">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-empty v-if="!serviceInstances.length" description="暂无实例" />

      <el-collapse v-else v-model="activeInstances" accordion>
        <el-collapse-item
          v-for="instance in serviceInstances"
          :key="instance.instance_key"
          :name="instance.instance_key"
        >
          <template #title>
            <div class="instance-title">
              <el-badge
                :value="instance.online ? '在线' : '离线'"
                :type="instance.online ? 'success' : 'info'"
              >
                <span class="instance-name">
                  {{ instance.language }} - {{ instance.host_ip }}
                </span>
              </el-badge>
              <el-tag size="small" type="info" class="version-tag">
                {{ instance.version }}
              </el-tag>
            </div>
          </template>

          <div class="instance-detail">
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="实例 Key">
                <code>{{ instance.instance_key }}</code>
              </el-descriptions-item>
              <el-descriptions-item label="语言">
                <el-tag :type="getLanguageTagType(instance.language)">
                  {{ instance.language }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="主机 IP">
                {{ instance.host_ip }}
              </el-descriptions-item>
              <el-descriptions-item label="主机 MAC">
                {{ instance.host_mac }}
              </el-descriptions-item>
              <el-descriptions-item label="工作目录" :span="2">
                <code>{{ instance.working_dir }}</code>
              </el-descriptions-item>
              <el-descriptions-item label="首次发现">
                {{ formatDateTime(instance.first_seen) }}
              </el-descriptions-item>
              <el-descriptions-item label="最后心跳">
                {{ formatDateTime(instance.last_heartbeat) }}
              </el-descriptions-item>
            </el-descriptions>

            <!-- 控制按钮区域 (仅管理员可见) -->
            <div class="instance-controls" v-if="isAdmin">
              <el-button
                type="warning"
                size="small"
                @click.stop="handleStop(instance)"
                :disabled="!instance.online"
              >
                <el-icon><VideoPause /></el-icon>
                停止
              </el-button>
              <el-button
                type="primary"
                size="small"
                @click.stop="handleRestart(instance)"
                :disabled="!instance.online"
              >
                <el-icon><RefreshRight /></el-icon>
                重启
              </el-button>
              <el-button
                type="danger"
                size="small"
                @click.stop="handleDelete(instance)"
                :disabled="instance.online"
              >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Refresh, VideoPause, RefreshRight, Delete } from '@element-plus/icons-vue'
import { servicesApi, type ServiceMetadata, type MethodMetadata, type ServiceStatus, type Instance } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useServicesStore, useInstancesStore, useUserStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const servicesStore = useServicesStore()
const instancesStore = useInstancesStore()
const userStore = useUserStore()

const serviceName = computed(() => route.params.name as string)
const service = ref<ServiceMetadata | null>(null)
const methods = ref<MethodMetadata[]>([])
const serviceStatus = ref<ServiceStatus | null>(null)
const loading = ref(false)
const instancesLoading = ref(false)
const activeInstances = ref<string[]>([])

// 当前服务的实例列表
const serviceInstances = computed(() => {
  return instancesStore.getInstancesByService(serviceName.value)
})

// 是否管理员
const isAdmin = computed(() => {
  return userStore.role === 'admin'
})

function goBack() {
  router.push('/services')
}

function goToDebug(serviceName: string, methodName: string) {
  router.push(`/services/${serviceName}/debug/${methodName}`)
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

function formatJSON(obj: any) {
  return JSON.stringify(obj, null, 2)
}

// 加载实例数据
async function loadInstances() {
  instancesLoading.value = true
  try {
    await instancesStore.loadServiceInstances(serviceName.value)
  } catch (error: any) {
    ElMessage.error('加载实例列表失败')
  } finally {
    instancesLoading.value = false
  }
}

// 停止实例
async function handleStop(instance: Instance) {
  try {
    await ElMessageBox.confirm(
      `确定要停止实例 ${instance.instance_key} 吗?`,
      '确认操作',
      { type: 'warning' }
    )
    await instancesStore.stopInstance(instance.instance_key)
    await loadInstances()
  } catch {
    // 用户取消
  }
}

// 重启实例
async function handleRestart(instance: Instance) {
  try {
    await ElMessageBox.confirm(
      `确定要重启实例 ${instance.instance_key} 吗?`,
      '确认操作',
      { type: 'warning' }
    )
    await instancesStore.restartInstance(instance.instance_key)
    await loadInstances()
  } catch {
    // 用户取消
  }
}

// 删除实例
async function handleDelete(instance: Instance) {
  try {
    await ElMessageBox.confirm(
      `确定要删除离线实例 ${instance.instance_key} 吗? 此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await instancesStore.deleteInstance(instance.instance_key)
  } catch {
    // 用户取消
  }
}

// 语言标签颜色
function getLanguageTagType(language: string) {
  const types: Record<string, string> = {
    go: 'success',
    python: 'warning',
    csharp: 'danger',
    javascript: 'primary'
  }
  return types[language] || 'info'
}

// 日期时间格式化
function formatDateTime(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

async function loadData() {
  loading.value = true
  try {
    // Get service details
    const services = await servicesApi.list()
    service.value = services.find(s => s.name === serviceName.value) || null

    if (!service.value) {
      ElMessage.error('服务不存在')
      router.push('/services')
      return
    }

    // Get methods for this service
    methods.value = await servicesApi.getMethods(serviceName.value)

    // Get service status from store
    await servicesStore.loadStatus()
    serviceStatus.value = servicesStore.getServiceStatus(serviceName.value) || null
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
  loadInstances()
})
</script>

<style scoped>
.service-detail-view {
  padding: 20px;
}

.header {
  margin-bottom: 20px;
}

.status-card {
  margin-bottom: 20px;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 20px;
}

.status-meta {
  display: flex;
  gap: 20px;
  color: #666;
}

.description {
  margin: 10px 0;
  color: #333;
}

.tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.methods-card h3 {
  margin: 0;
}

.methods-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.method-item {
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 16px;
}

.method-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.method-header h4 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.method-description {
  color: #666;
  margin-bottom: 12px;
}

.method-meta {
  margin-bottom: 12px;
}

.parameters,
.examples {
  margin-top: 12px;
}

.parameters h5,
.examples h5 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #333;
}

.example-item {
  background-color: #f5f5f5;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 8px;
}

.example-header {
  margin-bottom: 8px;
}

.example-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.example-block {
  display: flex;
  flex-direction: column;
}

.example-block .label {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.example-block pre {
  background-color: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 8px;
  margin: 0;
  font-size: 12px;
  overflow-x: auto;
}

/* 实例卡片样式 */
.instances-card {
  margin-top: 20px;
}

.instances-card h3 {
  margin: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.instance-title {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.instance-name {
  font-weight: 500;
}

.version-tag {
  margin-left: auto;
}

.instance-detail {
  padding: 10px 0;
}

.instance-controls {
  display: flex;
  gap: 10px;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #e0e0e0;
}

.instance-controls code {
  background-color: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
}
</style>
