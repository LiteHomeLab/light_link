<template>
  <div class="debug-view" v-loading="loading">
    <el-page-header
      :title="`${serviceName} :: ${methodName}`"
      @back="goBack"
      class="header"
    >
      <template #extra>
        <el-button @click="goToDetail" type="primary" plain>
          <el-icon><Document /></el-icon>
          方法详情
        </el-button>
      </template>
    </el-page-header>

    <el-card class="debug-card">
      <template #header>
        <h3>调试接口</h3>
      </template>

      <el-form label-width="100px">
        <el-form-item label="服务名称">
          <el-input :value="serviceName" disabled />
        </el-form-item>

        <el-form-item label="方法名称">
          <el-input :value="methodName" disabled />
        </el-form-item>

        <el-form-item label="请求参数 (JSON)">
          <el-input
            v-model="paramsInput"
            type="textarea"
            :rows="10"
            placeholder='{"key": "value"}'
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            @click="handleCall"
            :loading="calling"
            :disabled="!serviceOnline"
          >
            <el-icon><Position /></el-icon>
            {{ serviceOnline ? '调用' : '服务离线' }}
          </el-button>
          <el-button @click="formatInput">
            <el-icon><MagicStick /></el-icon>
            格式化输入
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-if="result" class="result-card">
      <template #header>
        <div class="result-header">
          <h3>调用结果</h3>
          <div class="result-meta">
            <el-tag :type="result.success ? 'success' : 'danger'">
              {{ result.success ? '成功' : '失败' }}
            </el-tag>
            <span v-if="result.duration" class="duration">
              耗时: {{ result.duration }}ms
            </span>
          </div>
        </div>
      </template>

      <div class="result-content">
        <div v-if="result.error" class="error-section">
          <h4>错误信息:</h4>
          <pre class="error">{{ result.error }}</pre>
        </div>

        <div v-if="result.data !== undefined && result.data !== null" class="data-section">
          <h4>返回数据:</h4>
          <pre>{{ formatJSON(result.data) }}</pre>
        </div>
      </div>
    </el-card>

    <el-card class="examples-card" v-if="method && method.examples && method.examples.length">
      <template #header>
        <h3>示例</h3>
      </template>

      <div class="examples-list">
        <div
          v-for="(example, index) in method.examples"
          :key="index"
          class="example-item"
        >
          <div class="example-header">
            <el-tag>{{ example.name || `示例 ${index + 1}` }}</el-tag>
            <el-button
              size="small"
              @click="useExample(example)"
              type="primary"
              plain
            >
              使用此示例
            </el-button>
          </div>
          <div v-if="example.description" class="example-description">
            {{ example.description }}
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Position,
  MagicStick,
  Document
} from '@element-plus/icons-vue'
import { servicesApi, callApi, type MethodMetadata, type CallResult } from '@/api'
import { ElMessage } from 'element-plus'
import { useServicesStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const servicesStore = useServicesStore()

const serviceName = computed(() => route.params.name as string)
const methodName = computed(() => route.params.method as string)

const method = ref<MethodMetadata | null>(null)
const paramsInput = ref('{\n  \n}')
const result = ref<CallResult | null>(null)
const loading = ref(false)
const calling = ref(false)

const serviceOnline = computed(() => {
  const status = servicesStore.servicesStatus.get(serviceName.value)
  return status?.online || false
})

function goBack() {
  router.push(`/services/${serviceName.value}`)
}

function goToDetail() {
  router.push(`/services/${serviceName.value}`)
}

function formatJSON(obj: any) {
  return JSON.stringify(obj, null, 2)
}

function formatInput() {
  try {
    const parsed = JSON.parse(paramsInput.value)
    paramsInput.value = formatJSON(parsed)
  } catch (e) {
    ElMessage.error('JSON 格式错误，无法格式化')
  }
}

function useExample(example: any) {
  if (example.input !== undefined && example.input !== null) {
    paramsInput.value = formatJSON(example.input)
  } else {
    paramsInput.value = '{\n  \n}'
  }
}

async function handleCall() {
  let params: any
  try {
    params = JSON.parse(paramsInput.value)
  } catch (e: any) {
    ElMessage.error('JSON 格式错误: ' + e.message)
    return
  }

  calling.value = true
  result.value = null

  try {
    const response = await callApi.call({
      service: serviceName.value,
      method: methodName.value,
      params
    })
    result.value = response
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '调用失败')
  } finally {
    calling.value = false
  }
}

async function loadMethod() {
  loading.value = true
  try {
    const methods = await servicesApi.getMethods(serviceName.value)
    method.value = methods.find(m => m.name === methodName.value) || null

    if (!method.value) {
      ElMessage.error('方法不存在')
      router.push(`/services/${serviceName.value}`)
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadMethod()
})
</script>

<style scoped>
.debug-view {
  padding: 20px;
}

.header {
  margin-bottom: 20px;
}

.debug-card,
.result-card,
.examples-card {
  margin-bottom: 20px;
}

.debug-card h3,
.result-card h3,
.examples-card h3 {
  margin: 0;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.result-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.duration {
  font-size: 14px;
  color: #666;
}

.result-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.error-section h4,
.data-section h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #333;
}

.error {
  background-color: #fee;
  border: 1px solid #fcc;
  border-radius: 4px;
  padding: 12px;
  color: #c00;
  margin: 0;
}

.data-section pre {
  background-color: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
  margin: 0;
  overflow-x: auto;
}

.examples-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.example-item {
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
}

.example-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.example-description {
  color: #666;
  font-size: 14px;
}
</style>
