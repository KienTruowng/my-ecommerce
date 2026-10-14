<script setup lang="ts">
import { api } from '@/api/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { toTypedSchema } from '@vee-validate/zod'
import { useForm } from 'vee-validate'
import { toast } from 'vue-sonner'
import * as z from 'zod'

// Schema Validation by zod
const formSchema = toTypedSchema(
  z
    .object({
      oldPassword: z.string().min(1, 'Vui lòng nhập mật khẩu hiện tại'),
      newPassword: z.string().min(8, 'Mật khẩu mới phải có ít nhất 8 ký tự'),
      confirmNewPassword: z.string().min(1, 'Vui lòng xác nhận mật khẩu mới'),
    })
    .refine((data) => data.oldPassword !== data.newPassword, {
      error: 'Mật khẩu mới không được trùng với mật khẩu hiện tại',
      path: ['newPassword'],
    })
    .refine((data) => data.newPassword === data.confirmNewPassword, {
      error: 'Mật khẩu xác nhận không khớp',
      path: ['confirmNewPassword'],
    }),
)

const { defineField, handleSubmit, errors, isSubmitting } = useForm({
  validationSchema: formSchema,
  initialValues: {
    oldPassword: '',
    newPassword: '',
    confirmNewPassword: '',
  },
})

// Define fields
const [oldPassword, oldPasswordProps] = defineField('oldPassword')
const [newPassword, newPasswordProps] = defineField('newPassword')
const [confirmNewPassword, confirmNewPasswordProps] = defineField('confirmNewPassword')

const onSubmit = handleSubmit(async (values) => {
  try {
    await api.POST('/api/auth/change-password', {
      body: {
        old_password: values.oldPassword,
        new_password: values.newPassword,
      },
    })
    toast.success('Cập nhật mật khẩu thành công')
  } catch (error) {
    toast.error('Cập nhật mật khẩu mới thất bại')
  }
})
</script>

<template>
  <h2 class="uppercase font-bold">Thay đổi mật khẩu</h2>

  <form class="space-y-4 max-w-md" @submit.prevent="onSubmit">
    <div class="space-y-2">
      <Label for="old_password">Mật khẩu hiện tại</Label>
      <Input id="old_password" type="password" v-model="oldPassword" v-bind="oldPasswordProps" />
      <p v-if="errors.oldPassword" class="text-sm text-red-500 mt-1">{{ errors.oldPassword }}</p>
    </div>

    <div class="space-y-2">
      <Label for="new_password">Mật khẩu mới</Label>
      <Input id="new_password" type="password" v-model="newPassword" v-bind="newPasswordProps" />
      <p v-if="errors.newPassword" class="text-sm text-red-500 mt-1">{{ errors.newPassword }}</p>
    </div>

    <div class="space-y-2">
      <Label for="confirm_new_password">Xác nhận mật khẩu</Label>
      <Input
        id="confirm_new_password"
        type="password"
        v-model="confirmNewPassword"
        v-bind="confirmNewPasswordProps"
      />
      <p v-if="errors.confirmNewPassword" class="text-sm text-red-500 mt-1">
        {{ errors.confirmNewPassword }}
      </p>
    </div>

    <Button type="submit" class="mt-4" :disabled="isSubmitting">
      <Spinner v-if="isSubmitting" class="animate-spin" />
      <span v-if="isSubmitting">Đang xử lý...</span>
      <span v-else>Xác nhận thay đổi</span>
    </Button>
  </form>
</template>
