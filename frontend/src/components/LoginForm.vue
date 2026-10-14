<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()

const model = defineModel({
  default: { email: '', password: '' },
})

const emit = defineEmits(['submit'])

const handleSubmit = () => {
  emit('submit')
}
</script>

<template>
  <div :class="cn('flex flex-col gap-6', props.class)">
    <Card class="border-0 bg-white text-[#0B2347] shadow-none">
      <CardHeader>
        <CardTitle class="text-2xl font-bold text-[#0B2347]">Chào mừng trở lại</CardTitle>
        <CardDescription> Đăng nhập để tiếp tục mua sắm tại KIEN STORE </CardDescription>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleSubmit">
          <FieldGroup>
            <Field>
              <FieldLabel for="email"> Email </FieldLabel>
              <Input
                v-model="model.email"
                id="email"
                type="email"
                placeholder="m@example.com"
                required
              />
            </Field>
            <Field>
              <div class="flex items-center">
                <FieldLabel for="password"> Password </FieldLabel>
                <a href="#" class="ml-auto inline-block text-sm underline-offset-4 hover:underline">
                  Forgot your password?
                </a>
              </div>
              <Input v-model="model.password" id="password" type="password" required />
            </Field>
            <Field>
              <Button type="submit" class="h-11 w-full rounded-xl bg-[#175CD3] text-white font-semibold hover:bg-[#0B2347]">Đăng nhập</Button>
              <Button variant="outline" type="button"> Login with Google </Button>
              <FieldDescription class="text-center">
                Don't have an account?
                <RouterLink to="/signup"> Sign up </RouterLink>
              </FieldDescription>
            </Field>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
