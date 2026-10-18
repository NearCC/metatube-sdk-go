import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api'

export const useUserStore = defineStore('user', () => {
  const user = ref<{ id: number; username: string; role: string } | null>(null)
  const isLoggedIn = ref(false)

  async function checkLogin() {
    if (!api.isLoggedIn()) {
      isLoggedIn.value = false
      user.value = null
      return false
    }
    try {
      const res = await api.getMe()
      user.value = res.user
      isLoggedIn.value = true
      return true
    } catch {
      isLoggedIn.value = false
      user.value = null
      return false
    }
  }

  async function logout() {
    try {
      await api.logout()
    } catch {
    }
    isLoggedIn.value = false
    user.value = null
  }

  return {
    user,
    isLoggedIn,
    checkLogin,
    logout,
  }
})