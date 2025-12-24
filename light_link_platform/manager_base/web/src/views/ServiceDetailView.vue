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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import { servicesApi, type ServiceMetadata, type MethodMetadata, type ServiceStatus } from '@/api'
import { ElMessage } from 'element-plus'
import { useServicesStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const servicesStore = useServicesStore()

const serviceName = computed(() => route.params.name as string)
const service = ref<ServiceMetadata | null>(null)
const methods = ref<MethodMetadata[]>([])
const serviceStatus = ref<ServiceStatus | null>(null)
const loading = ref(false)

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
</style>
