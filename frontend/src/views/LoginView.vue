<script setup lang="ts">
import { api } from '@/api/api'
import LoginForm from '@/components/LoginForm.vue'
import { useAuthStore } from '@/stores/auth'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'

const loginData = ref({
  email: '',
  password: '',
})

const router = useRouter()
const auth = useAuthStore()

const handleLogin = async () => {
  try {
    const { data, error } = await api.POST('/api/auth/login', {
      body: loginData.value,
    })

    if (error) {
      console.error('Lỗi từ API:', error)
      const errorMessage = typeof error === 'string' ? error : 'Sai email hoặc mật khẩu'
      toast.error(errorMessage)
      return // Dừng hàm ngay lập tức
    }

    if (data && data.token) {
      auth.setToken(data.token) // data đã được Type-Safe tự động
      router.push('/')
      toast.success('Đăng nhập thành công')
    }
  } catch (err) {
    console.error('Lỗi hệ thống (Exception):', err)
    toast.error('Lỗi kết nối mạng, vui lòng thử lại!')
  }
}
</script>

<template>
  <div class="w-full max-w-sm">
    <LoginForm v-model="loginData" @submit="handleLogin" />
  </div>
</template>
