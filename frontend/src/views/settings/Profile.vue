<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Button from '@/components/ui/button/Button.vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from 'vue-sonner'
import { Spinner } from '@/components/ui/spinner'
import { useUserStore } from '@/stores/user'
import { toTypedSchema } from '@vee-validate/zod'
import * as z from 'zod'
import { useForm } from 'vee-validate'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Pencil } from 'lucide-vue-next'
import { api } from '@/api/api' // Import api client từ openapi-fetch
import { storeToRefs } from 'pinia'

const userStore = useUserStore()
const isUploadingAvatar = ref(false)

const formSchema = toTypedSchema(
  z.object({
    name: z.string().min(1, 'Ít nhất phải có 1 ký tự ở tên'),
    email: z.email('Phải là email hợp lệ'),
    city: z.string().optional(),
    district: z.string().optional(),
    latitude: z.number().nullable().optional(),
    longitude: z.number().nullable().optional(),
  }),
)

const { defineField, handleSubmit, errors, isSubmitting, setValues } = useForm({
  validationSchema: formSchema,
  initialValues: {
    name: '',
    email: '',
    city: '',
    district: '',
    latitude: null as number | null,
    longitude: null as number | null,
  },
})

const [name, nameProps] = defineField('name')
const [email, emailProps] = defineField('email')
const [city, cityProps] = defineField('city')
const [district, districtProps] = defineField('district')
const [latitude, latitudeProps] = defineField('latitude')
const [longitude, longitudeProps] = defineField('longitude')

// --- HANDLE UPDATE PROFILE ---
const onSubmit = handleSubmit(async (values) => {
  try {
    const { data, error } = await api.PATCH('/api/users/profile', {
      body: values,
    })

    if (error) {
      console.error('Lỗi từ API:', error)
      toast.error(typeof error === 'string' ? error : 'Lỗi khi cập nhật thông tin')
      return
    }

    // Nếu có data trả về từ API, dùng luôn data đó để update store (an toàn nhất)
    if (data) {
      userStore.profile = { ...userStore.profile, ...data }
    }
    toast.success('Cập nhật hồ sơ thành công!')
  } catch (err) {
    console.error('Lỗi Exception:', err)
    toast.error('Lỗi kết nối mạng')
  }
})
// --- HANDLE UPLOAD AVATAR ---
const handleUploadAvatar = async (event: Event) => {
  // Upload file to browser
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const allowedTypes = ['image/jpeg', 'image/png', 'image/webp']

  if (!allowedTypes.includes(file.type)) {
    toast.error('Phải là ảnh')
    return
  }

  try {
    // Send request to endpoint
    const { data, error } = await api.POST('/api/users/avatar/upload-url', {
      body: {
        filename: file.name,
        filetype: file.type,
      },
    })

    if (error || !data) {
      toast.error('Đã có lỗi xảy ra khi tiến hành upload avatar, vui lòng thử  lại sau')
      return
    }

    const uploadUrl = data?.upload_url

    if (!uploadUrl) {
      toast.error('Lỗi: Không lấy được đường dẫn tải lên từ server.')
      return
    }

    // Now upload image to bucket
    const uploadResponse = await fetch(uploadUrl, {
      method: 'PUT',
      body: file,
      headers: {
        'Content-Type': file.type,
      },
    })

    if (!uploadResponse.ok) {
      toast.error('Đã có lỗi xảy ra khi thực hiện upload avatar, vui lòng thử lại sau')
      return
    }

    const objectKeys = data?.object_key

    if (!objectKeys) {
      toast.error('Lỗi không có đường dẫn')
      return
    }
    const { data: confirmData, error: confirmError } = await api.POST('/api/users/avatar/confirm', {
      body: {
        object_key: objectKeys,
      },
    })

    if (confirmError || !confirmData) {
      toast.error('Đã có lỗi xảy ra khi thực hiện upload avatar, vui lòng thử lại sau')
      return
    }

    // Update new avatar url to pinia store
    if (userStore.profile) {
      userStore.profile.avatar_url = confirmData.avatar_url
      toast.success('Cập nhật avatar mới thành công')
    }
  } catch (error) {
    console.error(error)
    toast.error('Đã có lỗi xảy ra, vui lòng thử lại sau')
  }
}

onMounted(async () => {
  await userStore.fetchProfile()

  if (userStore.profile) {
    setValues({
      name: userStore.profile.name || '',
      email: userStore.profile.email || '',
      city: (userStore.profile as any).city || '',
      district: (userStore.profile as any).district || '',
      latitude: (userStore.profile as any).latitude ?? null,
      longitude: (userStore.profile as any).longitude ?? null,
    })
  }
})
</script>

<template>
  <h2 class="uppercase font-bold">Chỉnh sửa hồ sơ cá nhân</h2>

  <div v-if="userStore.isLoading">Đang tải...</div>

  <form v-else class="space-y-4 max-w-md" @submit.prevent="onSubmit">
    <div class="relative inline-block">
      <Label>
        <Avatar class="size-40">
          <AvatarImage :src="userStore.profile?.avatar_url || ''" />
          <AvatarFallback>{{ userStore.profile?.name }}</AvatarFallback>
        </Avatar>
        <Input type="file" accept="image/png, image/jpeg" hidden @change="handleUploadAvatar" />
      </Label>
    </div>

    <div class="space-y-2">
      <Label for="name">Tên hiển thị</Label>
      <Input id="name" type="text" v-model="name" v-bind="nameProps" />
      <p v-if="errors.name" class="text-sm text-red-500">{{ errors.name }}</p>
    </div>

    <div class="space-y-2">
      <Label for="email">Email</Label>
      <Input
        id="email"
        type="email"
        v-model="email"
        v-bind="emailProps"
        disabled
        class="bg-muted"
      />
      <p v-if="errors.email" class="text-sm text-red-500">{{ errors.email }}</p>
    </div>

    <!-- Address section -->
    <div class="border-t pt-4 mt-4">
      <h3 class="text-sm font-semibold mb-3">📍 Địa chỉ giao dịch</h3>
      <p class="text-xs text-muted-foreground mb-3">
        Địa chỉ của bạn được ẩn với người khác — chỉ thành phố được hiển thị công khai.
        Vĩ độ/kinh độ dùng để tính khoảng cách và gợi ý địa điểm gặp mặt.
      </p>

      <div class="grid gap-3 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="city">Thành phố</Label>
          <Input id="city" type="text" v-model="city" v-bind="cityProps" placeholder="VD: Hồ Chí Minh" />
        </div>
        <div class="space-y-2">
          <Label for="district">Quận/Huyện</Label>
          <Input id="district" type="text" v-model="district" v-bind="districtProps" placeholder="VD: Quận 1" />
        </div>
        <div class="space-y-2">
          <Label for="latitude">Vĩ độ (Latitude)</Label>
          <Input id="latitude" type="number" step="0.0000001" v-model.number="latitude" v-bind="latitudeProps" placeholder="VD: 10.762622" />
        </div>
        <div class="space-y-2">
          <Label for="longitude">Kinh độ (Longitude)</Label>
          <Input id="longitude" type="number" step="0.0000001" v-model.number="longitude" v-bind="longitudeProps" placeholder="VD: 106.660172" />
        </div>
      </div>
      <p class="text-[10px] text-muted-foreground mt-1">
        💡 Bạn có thể tìm tọa độ của mình tại <a href="https://www.latlong.net" target="_blank" class="text-emerald-600 hover:underline">latlong.net</a>
      </p>
    </div>

    <Button type="submit" class="my-2" :disabled="isSubmitting || isUploadingAvatar">
      <Spinner v-if="isSubmitting" class="animate-spin mr-2" />
      <span v-if="isSubmitting">Đang lưu...</span>
      <span v-else>Lưu thay đổi</span>
    </Button>
  </form>
</template>
