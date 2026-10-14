<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Button } from '@/components/ui/button'
import InputGroup from '../ui/input-group/InputGroup.vue'
import { InputGroupAddon, InputGroupButton, InputGroupInput } from '../ui/input-group'
import { Moon, Plus, SearchIcon, Sun } from 'lucide-vue-next'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { storeToRefs } from 'pinia'

const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()

const { profile } = storeToRefs(userStore)

const avatarUrl = computed(() => {
  return profile.value?.avatar_url || ''
})

const avatarFallback = computed(() => {
  const name = (userStore.profile as any)?.name || (userStore.profile as any)?.email || 'U'
  return String(name).trim().charAt(0).toUpperCase()
})

const searchId = ref('')
const isDark = ref(false)

const applyTheme = (value: boolean) => {
  isDark.value = value
  document.documentElement.classList.toggle('dark', value)
  localStorage.setItem('theme', value ? 'dark' : 'light')
}

const toggleTheme = () => {
  applyTheme(!isDark.value)
}

const handleSearch = () => {
  if (!searchId.value.trim()) return
  router.push(`/items/${searchId.value.trim()}`)
}

const handleLogout = async () => {
  try {
    await authStore.logout()
    router.push('/login')
  } catch (error) {
    console.error('Lỗi khi đăng xuất', error)
  }
}

onMounted(async () => {
  if (!userStore.profile) {
    await userStore.fetchProfile()
  }

  const stored = localStorage.getItem('theme')
  if (stored) {
    applyTheme(stored === 'dark')
  } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    applyTheme(true)
  }
})
</script>

<template>
  <header
    class="flex flex-wrap md:flex-nowrap items-center justify-between px-4 md:px-6 py-4 gap-y-3 border-b bg-background/90 backdrop-blur"
  >
    <div class="flex items-center gap-4 md:gap-6">
      <RouterLink to="/" class="text-xl font-semibold tracking-tight">CampusTrade</RouterLink>

      <nav class="hidden md:flex items-center gap-2 text-sm font-medium">
        <Button variant="ghost" as-child>
          <RouterLink to="/">Overview</RouterLink>
        </Button>
        <Button variant="ghost" as-child>
          <RouterLink to="/inventory">Inventory</RouterLink>
        </Button>
        <Button variant="ghost" as-child>
          <RouterLink to="/offers">Offers</RouterLink>
        </Button>
        <Button variant="ghost" as-child>
          <RouterLink to="/transactions">Transactions</RouterLink>
        </Button>
        <Button variant="ghost" as-child>
          <RouterLink to="/appointments">Appointments</RouterLink>
        </Button>
        <Button variant="ghost" as-child>
          <RouterLink to="/ai-chat">AI Chat</RouterLink>
        </Button>
        <Button variant="ghost" as-child>
          <RouterLink to="/recommendations">Recommendations</RouterLink>
        </Button>
        <Button variant="ghost" as-child>
          <RouterLink to="/categories">Categories</RouterLink>
        </Button>
        <Button v-if="userStore.isAdmin" variant="ghost" as-child>
          <RouterLink to="/dashboard">Admin</RouterLink>
        </Button>
      </nav>
    </div>

    <div class="w-full order-last md:order-0 md:flex-1 md:max-w-md lg:max-w-lg mx-0 md:mx-4">
      <form class="w-full" @submit.prevent="handleSearch">
        <InputGroup class="w-full">
          <InputGroupAddon>
            <SearchIcon class="w-4 h-4 text-muted-foreground" />
          </InputGroupAddon>
          <InputGroupInput v-model="searchId" placeholder="Enter item ID to view details" />
          <InputGroupAddon align="inline-end">
            <InputGroupButton variant="default" type="submit">Find</InputGroupButton>
          </InputGroupAddon>
        </InputGroup>
      </form>
    </div>

    <div class="flex items-center gap-1 md:gap-2 shrink-0">
      <Button variant="outline" class="hidden sm:flex" as-child>
        <RouterLink to="/inventory">
          <Plus class="w-4 h-4" />
          New item
        </RouterLink>
      </Button>

      <Button variant="ghost" size="icon" @click="toggleTheme">
        <Sun v-if="isDark" class="w-5 h-5" />
        <Moon v-else class="w-5 h-5" />
      </Button>

      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <Button variant="ghost" class="relative h-8 w-8 rounded-full ml-1 md:ml-2">
            <Avatar class="h-8 w-8">
              <AvatarImage v-if="avatarUrl" :src="avatarUrl" />
              <AvatarFallback>{{ avatarFallback }}</AvatarFallback>
            </Avatar>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" :side-offset="8" class="w-48">
          <DropdownMenuLabel>Tài khoản</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem as-child>
            <RouterLink to="/settings/profile"> Hồ sơ cá nhân </RouterLink>
          </DropdownMenuItem>
          <DropdownMenuItem as-child>
            <RouterLink to="/settings/change-password"> Đổi mật khẩu </RouterLink>
          </DropdownMenuItem>
          <DropdownMenuItem as-child>
            <RouterLink to="/transactions">Don hang cua toi</RouterLink>
          </DropdownMenuItem>
          <DropdownMenuItem as-child>
            <RouterLink to="/ai-chat">AI tro ly</RouterLink>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem @click="handleLogout" class="text-red-600 font-medium cursor-pointer">
            Đăng xuất
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  </header>
</template>
