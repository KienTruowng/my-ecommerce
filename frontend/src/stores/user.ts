import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'
import type { components } from '@/schemas/types'
import { api } from '@/api/api'

type UserProfile =
  components['schemas']['github_com_AiEmStylix_ecommerce_backend_internal_user_dto.UserProfileResponse']

export const useUserStore = defineStore('user', () => {
  const profile = ref<null | UserProfile>(null)
  const isLoading = ref(false)

  const isAdmin = computed(() => profile.value?.role === 'admin')

  async function fetchProfile() {
    const authStore = useAuthStore()

    if (!authStore.isAuthenticated) {
      return
    }

    isLoading.value = true
    try {
      const { data, error } = await api.GET('/api/users/profile')

      if (error) {
        profile.value = null
        console.error('Lỗi từ API trả về:', error)
        return
      }
      console.log(profile.value)

      profile.value = data ?? null
    } catch (error) {
      profile.value = null
      console.error('Lỗi Exception khi fetch profile:', error)
    } finally {
      isLoading.value = false
    }
  }

  function clearUserData() {
    profile.value = null
  }

  return { profile, isLoading, isAdmin, fetchProfile, clearUserData }
})
