<script setup lang="ts">
import { api } from '@/api/api'
import SignupForm from '@/components/SignupForm.vue'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'

const signupData = ref({
  name: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const router = useRouter()

const handleSignup = async () => {
  // 1. Validate phía Frontend: Kiểm tra mật khẩu khớp nhau
  if (signupData.value.password !== signupData.value.confirmPassword) {
    toast.error('Mật khẩu nhập lại không khớp!')
    return // Dừng hàm ngay
  }

  try {
    const { confirmPassword, ...apiPayload } = signupData.value

    // 2. Gọi API với openapi-fetch
    const { data, error } = await api.POST('/api/auth/register', {
      body: apiPayload,
    })

    // 3. Xử lý lỗi Logic (HTTP 400, 409...) ví dụ: Email đã tồn tại, validate xịt
    if (error) {
      console.error('Lỗi đăng ký từ API:', error)
      const errorMessage =
        typeof error === 'string' ? error : 'Đăng ký thất bại, vui lòng kiểm tra lại dữ liệu'
      toast.error(errorMessage)
      return // Dừng hàm ngay
    }

    // 4. Xử lý thành công (HTTP 201)
    toast.success('Đăng ký tài khoản thành công!')
    router.push('/login')
  } catch (err) {
    // 5. Bắt các lỗi ngoại lệ (Đứt mạng, sập server...)
    console.error('Lỗi hệ thống (Exception):', err)
    toast.error('Lỗi kết nối mạng, vui lòng thử lại!')
  }
}
</script>

<template>
  <div class="w-full max-w-sm">
    <SignupForm v-model="signupData" @submit="handleSignup" />
  </div>
</template>
