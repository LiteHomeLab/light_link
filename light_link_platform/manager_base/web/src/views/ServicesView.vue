<template>
  <div class="services-view">
    <el-page-header title="服务列表" class="header">
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
        placeholder="搜索服务..."
        style="width: 300px"
        :prefix-icon="Search"
        clearable
      />

      <el-radio-group v-model="filter" style="margin-left: 20px">
        <el-radio-button label="all">全部</el-radio-button>
        <el-radio-button label="online">在线</el-radio-button>
        <el-radio-button label="offline">离线</el-radio-button>
      </el-radio-group>

      <div class="stats" style="margin-left: auto">
        <el-tag>总计: {{ services.length }}</el-tag>
        <el-tag type="success" style="margin-left: 10px">在线: {{ onlineCount }}</el-tag>
        <el-tag type="danger" style="margin-left: 10px">离线: {{ offlineCount }}</el-tag>
      </div>
    </div>

    <el-row :gutter="20" class="service-list" v-loading="loading">
      <el-col
        v-for="service in filteredServices"
        :key="service.name"
        :xs="24"
        :sm="12"
        :md="8"
        :lg="6"
      >
        <el-card class="service-card" @click="viewService(service.name)" shadow="hover">
          <template #header>
            <div class="card-header">
              <el-badge
                :value="status(service.name)?.online ? '在线' : '离线'"
                :type="status(service.name)?.online ? 'success' : 'danger'"
              >
                <h3>{{ service.name }}</h3>
              </el-badge>
            </div>
          </template>

          <div class="service-info">
            <div class="info-row">
              <el-icon><Document /></el-icon>
              <span>版本: {{ service.version }}</span>
            </div>

            <div class="info-row" v-if="service.description">
              <el-icon><InfoFilled /></el-icon>
              <span>{{ service.description }}</span>
            </div>

            <div class="info-row">
              <el-icon><User /></el-icon>
              <span>{{ service.author || '未知' }}</span>
            </div>

            <div class="info-row" v-if="service.tags && service.tags.length">
              <el-icon><PriceTag /></el-icon>
              <el-tag
                v-for="tag in service.tags"
                :key="tag"
                size="small"
                style="margin-right: 5px"
              >
                {{ tag }}
              </el-tag>
            </div>

            <div class="info-row">
              <el-icon><Calendar /></el-icon>
              <span>注册: {{ formatDate(service.registered_at) }}</span>
            </div>
          </div>

          <template #footer>
            <div class="card-footer">
              <el-button size="small" @click.stop="viewService(service.name)">
                查看详情
              </el-button>
            </div>
          </template>
        </el-card>
      </el-col>
    </el-row>

    <el-empty
      v-if="!loading && filteredServices.length === 0"
      description="暂无服务"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  Refresh,
  Search,
  Document,
  InfoFilled,
  User,
  PriceTag,
  Calendar
} from '@element-plus/icons-vue'
import { useServicesStore } from '@/stores'
import { storeToRefs } from 'pinia'

const router = useRouter()
const servicesStore = useServicesStore()
const { services, servicesStatus, loading } = storeToRefs(servicesStore)

const search = ref('')
const filter = ref('all')

// 计算属性
const filteredServices = computed(() => {
  let result = services.value

  if (search.value) {
    const keyword = search.value.toLowerCase()
    result = result.filter(s =>
      s.name.toLowerCase().includes(keyword) ||
      s.description?.toLowerCase().includes(keyword) ||
      s.tags?.some(t => t.toLowerCase().includes(keyword))
    )
  }

  if (filter.value === 'online') {
    result = result.filter(s => servicesStatus.value.get(s.name)?.online)
  } else if (filter.value === 'offline') {
    result = result.filter(s => {
      const status = servicesStatus.value.get(s.name)
      return status && !status.online
    })
  }

  return result
})

const onlineCount = computed(() => {
  return services.value.filter(s => servicesStatus.value.get(s.name)?.online).length
})

const offlineCount = computed(() => {
  return services.value.filter(s => {
    const status = servicesStatus.value.get(s.name)
    return status && !status.online
  }).length
})

// 方法
const status = (name: string) => {
  return servicesStatus.value.get(name)
}

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

const viewService = (name: string) => {
  router.push(`/services/${name}`)
}

const loadData = () => {
  servicesStore.loadServices()
  servicesStore.loadStatus()
}

// 生命周期
onMounted(() => {
  servicesStore.init()
})
</script>

<style scoped>
.services-view {
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

.service-list {
  min-height: 300px;
}

.service-card {
  cursor: pointer;
  transition: all 0.3s;
  margin-bottom: 20px;
}

.service-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
}

.service-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 5px;
  color: #666;
  font-size: 14px;
}

.info-row .el-icon {
  font-size: 16px;
}

.card-footer {
  display: flex;
  justify-content: flex-end;
}
</style>
