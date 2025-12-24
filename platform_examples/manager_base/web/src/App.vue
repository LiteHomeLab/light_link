<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useServicesStore } from '@/stores/services'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const servicesStore = useServicesStore()
const router = useRouter()

onMounted(() => {
  // Initialize services if logged in
  if (userStore.isLoggedIn) {
    servicesStore.init()
  }
})

// Watch for route changes to initialize services when needed
router.afterEach((to) => {
  if (to.meta.requiresAuth && userStore.isLoggedIn) {
    servicesStore.init()
  }
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
    Ubuntu, Cantarell, 'Helvetica Neue', sans-serif;
}

#app {
  min-height: 100vh;
  background-color: #f5f5f5;
}
</style>
