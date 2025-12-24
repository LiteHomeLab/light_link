import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'

export const useUserStore = defineStore('user', () => {
  // State
  const token = ref<string>(localStorage.getItem('token') || '')
  const username = ref<string>(localStorage.getItem('username') || '')
  const role = ref<string>(localStorage.getItem('role') || '')

  // Getters
  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => role.value === 'admin')

  // Actions
  async function login(user: string, password: string) {
    try {
      const data = await authApi.login({ username: user, password })
      token.value = data.token
      username.value = user
      role.value = data.role

      localStorage.setItem('token', data.token)
      localStorage.setItem('username', user)
      localStorage.setItem('role', data.role)

      return true
    } catch (error: any) {
      console.error('登录失败:', error)
      throw error
    }
  }

  function logout() {
    token.value = ''
    username.value = ''
    role.value = ''

    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
  }

  // 初始化时从 localStorage 恢复状态
  function init() {
    const savedToken = localStorage.getItem('token')
    const savedUsername = localStorage.getItem('username')
    const savedRole = localStorage.getItem('role')

    if (savedToken) {
      token.value = savedToken
    }
    if (savedUsername) {
      username.value = savedUsername
    }
    if (savedRole) {
      role.value = savedRole
    }
  }

  return {
    token,
    username,
    role,
    isLoggedIn,
    isAdmin,
    login,
    logout,
    init
  }
})
