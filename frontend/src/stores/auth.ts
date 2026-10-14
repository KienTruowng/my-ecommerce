import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { jwtDecode } from 'jwt-decode'
import { useUserStore } from './user'

export const useAuthStore = defineStore('auth', () => {
  // STATE
  const token = ref<string | null>(localStorage.getItem('access_token'))
  const tokenExp = ref<number | null>(null)

  if (token.value) {
    try {
      const decoded = jwtDecode<{ exp: number }>(token.value)
      tokenExp.value = decoded.exp
    } catch (error) {
      console.error(error)
      localStorage.removeItem('access_token')
      token.value = null
    }
  }

  // GETTERS
  const isAuthenticated = computed(() => {
    if (!token.value || !tokenExp.value) return false
    const currentTime = Math.floor(Date.now() / 1000)
    return tokenExp.value > currentTime
  })

  // ACTIONS
  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('access_token', newToken)

    try {
      const decoded = jwtDecode<{ exp: number }>(newToken)
      tokenExp.value = decoded.exp
    } catch (error) {
      console.error('Invalid token received format, error: ', error)
    }
  }

  function logout() {
    token.value = null
    tokenExp.value = null
    localStorage.removeItem('access_token')

    const userStore = useUserStore()
    userStore.clearUserData()
  }

  return {
    token,
    tokenExp,
    isAuthenticated,
    setToken,
    logout,
  }
})
