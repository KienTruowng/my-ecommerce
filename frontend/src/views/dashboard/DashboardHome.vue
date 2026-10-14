<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Users, Package, Gavel, Handshake, Calendar, TrendingUp, Loader2 } from 'lucide-vue-next'

type AdminStats = {
  total_users: number
  total_items: number
  active_auctions: number
  pending_offers: number
  total_appointments: number
  completed_trades: number
}

const stats = ref<AdminStats>({
  total_users: 0, total_items: 0, active_auctions: 0,
  pending_offers: 0, total_appointments: 0, completed_trades: 0,
})
const isLoading = ref(true)
const error = ref('')

const fetchStats = async () => {
  isLoading.value = true
  error.value = ''
  try {
    const token = localStorage.getItem('access_token') ?? ''
    const res = await fetch('/api/admin/stats', {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    stats.value = data
  } catch (e: any) {
    error.value = e.message || 'Failed to load stats'
  } finally {
    isLoading.value = false
  }
}

onMounted(fetchStats)
</script>

<template>
  <div class="flex flex-1 flex-col">
    <div class="@container/main flex flex-1 flex-col gap-4 py-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-semibold">Admin Dashboard</h1>
          <p class="text-sm text-muted-foreground mt-1">Overview of all marketplace activity.</p>
        </div>
        <button
          class="text-xs text-muted-foreground hover:text-foreground transition-colors"
          @click="fetchStats"
          :disabled="isLoading"
        >
          <Loader2 v-if="isLoading" class="h-3.5 w-3.5 inline animate-spin mr-1" />
          Refresh
        </button>
      </div>

      <!-- Error -->
      <div v-if="error" class="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-400">
        ⚠️ {{ error }} — make sure the backend is running.
      </div>

      <!-- Stats grid -->
      <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card class="bg-card/80 transition-shadow hover:shadow-md">
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardDescription>Total Users</CardDescription>
            <Users class="h-5 w-5 text-sky-500" />
          </CardHeader>
          <CardContent>
            <div class="text-3xl font-bold">{{ stats.total_users }}</div>
            <p class="mt-1 text-xs text-muted-foreground">Registered accounts</p>
          </CardContent>
        </Card>

        <Card class="bg-card/80 transition-shadow hover:shadow-md">
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardDescription>Total Items</CardDescription>
            <Package class="h-5 w-5 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div class="text-3xl font-bold">{{ stats.total_items }}</div>
            <p class="mt-1 text-xs text-muted-foreground">Listed items</p>
          </CardContent>
        </Card>

        <Card class="bg-card/80 transition-shadow hover:shadow-md">
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardDescription>Active Auctions</CardDescription>
            <Gavel class="h-5 w-5 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div class="text-3xl font-bold">{{ stats.active_auctions }}</div>
            <p class="mt-1 text-xs text-muted-foreground">Live bidding</p>
          </CardContent>
        </Card>

        <Card class="bg-card/80 transition-shadow hover:shadow-md">
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardDescription>Pending Offers</CardDescription>
            <Handshake class="h-5 w-5 text-rose-500" />
          </CardHeader>
          <CardContent>
            <div class="text-3xl font-bold">{{ stats.pending_offers }}</div>
            <p class="mt-1 text-xs text-muted-foreground">Awaiting response</p>
          </CardContent>
        </Card>

        <Card class="bg-card/80 transition-shadow hover:shadow-md">
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardDescription>Appointments</CardDescription>
            <Calendar class="h-5 w-5 text-violet-500" />
          </CardHeader>
          <CardContent>
            <div class="text-3xl font-bold">{{ stats.total_appointments }}</div>
            <p class="mt-1 text-xs text-muted-foreground">Scheduled meetups</p>
          </CardContent>
        </Card>

        <Card class="bg-card/80 transition-shadow hover:shadow-md">
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardDescription>Completed Trades</CardDescription>
            <TrendingUp class="h-5 w-5 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div class="text-3xl font-bold">{{ stats.completed_trades }}</div>
            <p class="mt-1 text-xs text-muted-foreground">Accepted offers</p>
          </CardContent>
        </Card>
      </div>

      <!-- Quick links -->
      <div class="grid gap-4 sm:grid-cols-3 mt-2">
        <RouterLink to="/dashboard/users" class="rounded-xl border bg-card/80 p-4 transition-all hover:shadow-md hover:border-emerald-300/60">
          <div class="flex items-center gap-3">
            <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-sky-100 text-sky-600 dark:bg-sky-900 dark:text-sky-400">
              <Users class="h-5 w-5" />
            </div>
            <div>
              <p class="text-sm font-semibold">Manage Users</p>
              <p class="text-xs text-muted-foreground">View, edit, delete accounts</p>
            </div>
          </div>
        </RouterLink>
        <RouterLink to="/dashboard/locations" class="rounded-xl border bg-card/80 p-4 transition-all hover:shadow-md hover:border-emerald-300/60">
          <div class="flex items-center gap-3">
            <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-rose-100 text-rose-600 dark:bg-rose-900 dark:text-rose-400">
              <Gavel class="h-5 w-5" />
            </div>
            <div>
              <p class="text-sm font-semibold">Manage Locations</p>
              <p class="text-xs text-muted-foreground">Campus meetup spots</p>
            </div>
          </div>
        </RouterLink>
        <RouterLink to="/dashboard/shipping" class="rounded-xl border bg-card/80 p-4 transition-all hover:shadow-md hover:border-emerald-300/60">
          <div class="flex items-center gap-3">
            <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-amber-100 text-amber-600 dark:bg-amber-900 dark:text-amber-400">
              <Package class="h-5 w-5" />
            </div>
            <div>
              <p class="text-sm font-semibold">Manage Shipping</p>
              <p class="text-xs text-muted-foreground">Delivery services</p>
            </div>
          </div>
        </RouterLink>
      </div>
    </div>
  </div>
</template>
