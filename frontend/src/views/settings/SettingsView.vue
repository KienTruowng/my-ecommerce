<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Lock, User } from 'lucide-vue-next'
</script>

<template>
  <div class="flex flex-col min-h-full md:flex-row">
    <!-- Settings button -->
    <div class="w-full md:w-64 flex flex-col gap-2">
      <Button variant="outline" class="w-full" size="icon" as-child>
        <RouterLink :to="{ name: 'SettingsProfile' }">
          <User />
          Hồ sơ của tôi
        </RouterLink>
      </Button>
      <Button variant="outline" class="w-full" as-child>
        <RouterLink :to="{ name: 'ChangePassword' }">
          <Lock />
          Đổi mật khẩu
        </RouterLink>
      </Button>
    </div>

    <Separator class="my-4 md:my-0 md:mx-4" orientation="vertical" />
    <!-- Settings views -->
    <div class="flex-1 px-4 py-2 space-y-2">
      <RouterView />
    </div>
  </div>
</template>

<style scoped></style>
