<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { storeToRefs } from 'pinia'
import {
  LayoutDashboard,
  Package,
  Calendar,
  Bot,
  Compass,
  Tags,
  Shield,
  ShoppingBag,
  ShoppingCart,
  Sun,
  Moon,
  ChevronLeft,
  ChevronRight,
  Bookmark,
  Bell,
  Search,
  X,
  ClipboardList,
  ArrowLeftRight,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()
const cartStore = useCartStore()
const { profile } = storeToRefs(userStore)

const collapsed = ref(false)
const isDark = ref(false)
const searchId = ref('')
const notifCount = ref(0)

const fetchNotifCount = async () => {
  try {
    const token = localStorage.getItem('access_token') ?? ''
    const res = await fetch(`/api/notifications`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (res.ok) {
      const data = await res.json()
      notifCount.value = data.unread_count ?? 0
    }
  } catch { /* ignore */ }
}

// Poll notifications every 30s
let notifTimer: ReturnType<typeof setInterval> | null = null

const avatarUrl = computed(() => (profile.value as any)?.avatar_url || '')
const avatarFallback = computed(() => {
  const name = (profile.value as any)?.name || (profile.value as any)?.email || 'U'
  return String(name).trim().charAt(0).toUpperCase()
})

const applyTheme = (value: boolean) => {
  isDark.value = value
  document.documentElement.classList.toggle('dark', value)
  localStorage.setItem('theme', value ? 'dark' : 'light')
}

const toggleTheme = () => applyTheme(!isDark.value)

const handleSearch = () => {
  if (!searchId.value.trim()) return
  router.push(`/items/${searchId.value.trim()}`)
  searchId.value = ''
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

const navItems = computed(() => [
  { to: '/', label: 'Marketplace', icon: ShoppingBag, exact: true },
  { to: '/inventory', label: 'My Inventory', icon: Package },
  { to: '/cart', label: 'Cart', icon: ShoppingCart },
  { to: '/orders', label: 'Orders', icon: ClipboardList },
  { to: '/appointments', label: 'Appointments', icon: Calendar },
  { to: '/wallet', label: 'Wallet', icon: ArrowLeftRight },
  { to: '/ai-chat', label: 'AI Assistant', icon: Bot },
  { to: '/recommendations', label: 'For You', icon: Compass },
  { to: '/categories', label: 'Categories', icon: Tags },
  ...(userStore.isAdmin
    ? [{ to: '/dashboard', label: 'Admin Panel', icon: Shield }]
    : []),
])

const isActive = (to: string, exact?: boolean) => {
  if (exact) return route.path === to
  return route.path.startsWith(to)
}

onMounted(() => {
  const stored = localStorage.getItem('theme')
  if (stored) {
    applyTheme(stored === 'dark')
  } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    applyTheme(true)
  }
  fetchNotifCount()
  cartStore.fetchCart()
  notifTimer = setInterval(fetchNotifCount, 30000)
})
</script>

<template>
  <!-- Mobile hamburger -->
  <button
    class="fixed left-3 top-3 z-50 flex h-9 w-9 items-center justify-center rounded-lg border bg-background shadow-sm lg:hidden"
    @click="collapsed = !collapsed"
  >
    <ChevronRight v-if="collapsed" class="h-5 w-5" />
    <ChevronLeft v-else class="h-5 w-5" />
  </button>

  <!-- Mobile overlay -->
  <div
    v-if="!collapsed"
    class="fixed inset-0 z-40 bg-black/40 backdrop-blur-sm lg:hidden"
    @click="collapsed = true"
  />

  <!-- Sidebar -->
  <aside
    class="fixed left-0 top-0 z-40 flex h-full flex-col border-r bg-background/95 backdrop-blur transition-all duration-300"
    :class="[
      collapsed ? '-translate-x-full lg:translate-x-0 lg:w-[72px]' : 'translate-x-0 w-[280px]',
      'lg:relative lg:z-0',
    ]"
  >
    <!-- Logo + collapse toggle -->
    <div class="flex h-14 items-center justify-between border-b px-4 shrink-0">
      <RouterLink
        to="/"
        class="flex items-center gap-2 overflow-hidden whitespace-nowrap"
      >
        <div class="flex h-7 w-7 items-center justify-center rounded-lg bg-emerald-600 text-xs font-bold text-white shrink-0">
          CT
        </div>
        <span
          class="text-base font-semibold tracking-tight transition-all duration-300"
          :class="collapsed ? 'lg:opacity-0 lg:w-0 lg:overflow-hidden' : 'opacity-100'"
        >KIEN STORE</span>
      </RouterLink>
      <Button
        variant="ghost"
        size="icon"
        class="hidden h-8 w-8 lg:flex shrink-0"
        @click="collapsed = !collapsed"
      >
        <ChevronLeft v-if="!collapsed" class="h-4 w-4" />
        <ChevronRight v-else class="h-4 w-4" />
      </Button>
    </div>

    <!-- Quick search -->
    <div
      class="px-3 py-2 border-b shrink-0"
      :class="collapsed ? 'lg:hidden' : ''"
    >
      <form @submit.prevent="handleSearch" class="relative">
        <Search class="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
        <Input
          v-model="searchId"
          placeholder="Search by item ID..."
          class="h-8 pl-8 pr-8 text-xs"
        />
        <button
          v-if="searchId"
          type="button"
          class="absolute right-1.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          @click="searchId = ''"
        >
          <X class="h-3.5 w-3.5" />
        </button>
      </form>
    </div>

    <!-- Notification bell -->
    <RouterLink
      to="/orders"
      class="flex items-center gap-3 mx-2 mb-1 rounded-lg px-3 py-2 text-sm font-medium transition-colors text-muted-foreground hover:bg-muted hover:text-foreground"
      :class="collapsed ? 'lg:justify-center lg:px-0' : ''"
      :title="collapsed ? `Orders (${notifCount} new)` : undefined"
    >
      <div class="relative shrink-0">
        <Bell class="h-4 w-4" :class="notifCount > 0 ? 'text-rose-500' : ''" />
        <span
          v-if="notifCount > 0"
          class="absolute -right-1.5 -top-1.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-0.5 text-[9px] font-bold text-white shadow"
        >{{ notifCount > 99 ? '99+' : notifCount }}</span>
      </div>
      <span
        class="transition-opacity whitespace-nowrap"
        :class="collapsed ? 'lg:opacity-0 lg:w-0 lg:overflow-hidden' : 'opacity-100'"
      >{{ notifCount > 0 ? `${notifCount} new offer${notifCount > 1 ? 's' : ''}` : 'No new offers' }}</span>
    </RouterLink>

    <!-- Navigation -->
    <nav class="flex-1 overflow-y-auto px-2 py-3 space-y-1">
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors"
        :class="isActive(item.to, item.exact)
          ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
        :title="collapsed ? item.label : undefined"
        @click="collapsed = true"
      >
        <component :is="item.icon" class="h-4 w-4 shrink-0" />
        <span
          class="transition-opacity whitespace-nowrap"
          :class="collapsed ? 'lg:opacity-0 lg:w-0 lg:overflow-hidden' : 'opacity-100'"
        >{{ item.label }}</span>
      </RouterLink>
    </nav>

    <!-- ═══ YOUR CART SECTION ═══ -->
    <div class="border-t shrink-0" :class="collapsed ? 'lg:hidden' : ''">
      <RouterLink
        to="/cart"
        class="flex w-full items-center justify-between px-4 py-2.5 text-sm font-semibold hover:bg-muted/50 transition-colors"
      >
        <span class="flex items-center gap-2">
          <ShoppingCart class="h-4 w-4 text-amber-500" />
          <span>Your Cart</span>
        </span>
        <div class="flex items-center gap-2">
          <Badge
            v-if="cartStore.count > 0"
            variant="secondary"
            class="h-5 px-1.5 text-[10px]"
          >
            {{ cartStore.count }}
          </Badge>
        </div>
      </RouterLink>

      <div
        class="border-t px-3 py-2 space-y-2 max-h-[200px] overflow-y-auto scrollbar-thin"
      >
        <!-- Empty -->
        <div
          v-if="cartStore.isEmpty"
          class="py-3 text-center"
        >
          <Bookmark class="mx-auto h-5 w-5 text-muted-foreground/40" />
          <p class="mt-1 text-[11px] text-muted-foreground">
            Add items from MarketPlace
          </p>
          <RouterLink to="/" class="text-[11px] text-emerald-600 hover:underline mt-0.5 inline-block">
            Browse items →
          </RouterLink>
        </div>

        <!-- Cart item -->
        <div v-else class="rounded-lg border bg-card/60 p-2.5">
          <div class="flex items-start gap-2">
            <img
              v-if="(cartStore.cart?.item as any)?.images?.[0]?.url"
              :src="(cartStore.cart?.item as any)?.images?.[0]?.url"
              class="h-10 w-10 shrink-0 rounded-md border object-cover"
              :alt="(cartStore.cart?.item as any)?.name"
            />
            <div v-else class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md border bg-muted">
              <Package class="h-4 w-4 text-muted-foreground/50" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-semibold truncate">{{ (cartStore.cart?.item as any)?.name ?? 'Item' }}</p>
              <p v-if="cartStore.cart?.cash_amount" class="text-[10px] text-muted-foreground">
                {{ new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(cartStore.cart.cash_amount) }}
              </p>
              <button
                class="text-[10px] text-muted-foreground hover:text-destructive mt-1 underline underline-offset-2"
                @click="cartStore.clearCart()"
              >
                Remove
              </button>
            </div>
          </div>
        </div>

        <!-- Checkout button -->
        <RouterLink
          v-if="!cartStore.isEmpty"
          to="/cart"
          class="flex w-full items-center justify-center gap-1.5 rounded-md bg-emerald-600 py-1.5 text-[11px] font-medium text-white hover:bg-emerald-700 transition-colors"
        >
          <ShoppingCart class="h-3 w-3" />
          Go to Cart
        </RouterLink>
      </div>
    </div>

    <!-- User + Theme footer -->
    <div
      class="flex items-center gap-3 border-t px-4 py-3 shrink-0"
      :class="collapsed ? 'lg:justify-center lg:px-2' : ''"
    >
      <RouterLink
        to="/settings/profile"
        class="flex items-center gap-3 min-w-0 flex-1"
        :class="collapsed ? 'lg:flex-initial' : ''"
      >
        <Avatar class="h-8 w-8 shrink-0">
          <AvatarImage v-if="avatarUrl" :src="avatarUrl" />
          <AvatarFallback class="text-xs">{{ avatarFallback }}</AvatarFallback>
        </Avatar>
        <div
          class="min-w-0 transition-opacity"
          :class="collapsed ? 'lg:opacity-0 lg:w-0 lg:overflow-hidden' : 'opacity-100'"
        >
          <p class="text-sm font-medium truncate">{{ (profile as any)?.name || 'User' }}</p>
          <p class="text-[11px] text-muted-foreground truncate">{{ (profile as any)?.email || '' }}</p>
        </div>
      </RouterLink>

      <div class="flex items-center gap-1 shrink-0">
        <Button variant="ghost" size="icon" class="h-8 w-8" @click="toggleTheme">
          <Sun v-if="isDark" class="h-4 w-4" />
          <Moon v-else class="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          class="h-8 w-8 text-muted-foreground hover:text-destructive"
          @click="handleLogout"
          title="Sign out"
        >
          <ChevronRight class="h-4 w-4 rotate-180" />
        </Button>
      </div>
    </div>
  </aside>
</template>
