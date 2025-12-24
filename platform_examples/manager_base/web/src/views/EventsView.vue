<template>
  <div class="events-view">
    <el-page-header title="事件列表" class="header">
      <template #extra>
        <el-button @click="loadData" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </template>
    </el-page-header>

    <div class="toolbar">
      <el-input
        v-model="search"
        placeholder="搜索事件..."
        style="width: 300px"
        :prefix-icon="Search"
        clearable
      />

      <el-select
        v-model="filterType"
        placeholder="事件类型"
        style="width: 150px; margin-left: 10px"
        clearable
      >
        <el-option label="全部" value="" />
        <el-option label="注册" value="registered" />
        <el-option label="更新" value="updated" />
        <el-option label="上线" value="online" />
        <el-option label="下线" value="offline" />
      </el-select>

      <el-select
        v-model="filterService"
        placeholder="服务名称"
        style="width: 200px; margin-left: 10px"
        clearable
        filterable
      >
        <el-option label="全部" value="" />
        <el-option
          v-for="service in services"
          :key="service.name"
          :label="service.name"
          :value="service.name"
        />
      </el-select>
    </div>

    <el-table :data="filteredEvents" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="type" label="类型" width="120">
        <template #default="{ row }">
          <el-tag
            :type="getEventTypeColor(row.type)"
            size="small"
          >
            {{ getEventTypeName(row.type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="服务名称" width="200">
        <template #default="{ row }">
          {{ row.service_name || row.service || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="message" label="消息" show-overflow-tooltip />
      <el-table-column label="时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at || row.timestamp || '') }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button
            size="small"
            @click="viewDetail(row)"
          >
            详情
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && filteredEvents.length === 0" description="暂无事件" />

    <!-- Detail Dialog -->
    <el-dialog
      v-model="detailDialogVisible"
      title="事件详情"
      width="600px"
    >
      <div v-if="selectedEvent" class="event-detail">
        <div class="detail-row">
          <span class="label">ID:</span>
          <span>{{ selectedEvent.id }}</span>
        </div>
        <div class="detail-row">
          <span class="label">类型:</span>
          <el-tag :type="getEventTypeColor(selectedEvent.type)">
            {{ getEventTypeName(selectedEvent.type) }}
          </el-tag>
        </div>
        <div class="detail-row">
          <span class="label">服务名称:</span>
          <span>{{ selectedEvent.service_name || selectedEvent.service }}</span>
        </div>
        <div class="detail-row">
          <span class="label">消息:</span>
          <span>{{ selectedEvent.message }}</span>
        </div>
        <div class="detail-row">
          <span class="label">时间:</span>
          <span>{{ formatDate(selectedEvent.created_at || selectedEvent.timestamp || '') }}</span>
        </div>
        <div v-if="selectedEvent.metadata" class="detail-row full">
          <span class="label">元数据:</span>
          <pre>{{ formatJSON(selectedEvent.metadata) }}</pre>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Refresh, Search } from '@element-plus/icons-vue'
import { eventsApi, type ServiceEvent } from '@/api'
import { ws } from '@/utils/websocket'
import { useServicesStore } from '@/stores'
import { storeToRefs } from 'pinia'

const servicesStore = useServicesStore()
const { services } = storeToRefs(servicesStore)

const events = ref<ServiceEvent[]>([])
const search = ref('')
const filterType = ref('')
const filterService = ref('')
const loading = ref(false)
const detailDialogVisible = ref(false)
const selectedEvent = ref<ServiceEvent | null>(null)

const filteredEvents = computed(() => {
  let result = events.value

  if (search.value) {
    const keyword = search.value.toLowerCase()
    result = result.filter(e =>
      e.service_name?.toLowerCase().includes(keyword) ||
      e.service?.toLowerCase().includes(keyword) ||
      e.message?.toLowerCase().includes(keyword)
    )
  }

  if (filterType.value) {
    result = result.filter(e => e.type === filterType.value)
  }

  if (filterService.value) {
    result = result.filter(e => (e.service_name || e.service) === filterService.value)
  }

  return result
})

function getEventTypeName(type: string): string {
  const map: Record<string, string> = {
    registered: '注册',
    updated: '更新',
    online: '上线',
    offline: '下线'
  }
  return map[type] || type
}

function getEventTypeColor(type: string): string {
  const map: Record<string, string> = {
    registered: 'success',
    updated: 'primary',
    online: 'success',
    offline: 'danger'
  }
  return map[type] || 'info'
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

function formatJSON(obj: any) {
  return JSON.stringify(obj, null, 2)
}

function viewDetail(event: ServiceEvent) {
  selectedEvent.value = event
  detailDialogVisible.value = true
}

async function loadData() {
  loading.value = true
  try {
    events.value = await eventsApi.list()
  } catch (error: any) {
    console.error('加载事件失败:', error)
  } finally {
    loading.value = false
  }
}

function setupWebSocket() {
  ws.on('events', (event: ServiceEvent) => {
    // Add new event at the beginning
    events.value.unshift(event)
  })
  ws.subscribe(['events'])
}

function cleanupWebSocket() {
  ws.unsubscribe(['events'])
}

onMounted(() => {
  servicesStore.init()
  loadData()
  setupWebSocket()
})

onUnmounted(() => {
  cleanupWebSocket()
})
</script>

<style scoped>
.events-view {
  padding: 20px;
}

.header {
  margin-bottom: 20px;
}

.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.event-detail {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-row {
  display: flex;
  gap: 8px;
}

.detail-row .label {
  font-weight: 500;
  min-width: 80px;
  color: #666;
}

.detail-row.full {
  flex-direction: column;
}

.detail-row.full pre {
  background-color: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 8px;
  margin: 0;
  overflow-x: auto;
  font-size: 12px;
}
</style>
