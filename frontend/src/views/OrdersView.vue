<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import {
  Package, ShoppingBag, CheckCircle2, XCircle, Clock3,
  Loader2, ImageOff, CreditCard,
} from 'lucide-vue-next'

type OrderResponse = {
  id: number
  buyer_id: number
  seller_id: number
  item_id: number
  trade_in_item_id: number | null
  cash_amount: number
  payment_method: string
  payment_status: string
  order_status: string
  item_name: string
  trade_in_item_name?: string
  buyer_name: string
  seller_name: string
  created_at: string
}

const router = useRouter()
const orders = ref<OrderResponse[]>([])
const isLoading = ref(false)
const isCancelling = ref<number | null>(null)

const receivedOrders = computed(() =>
  orders.value.filter(() => {
    // All orders shown in received tab are those where current user is seller
    // This is determined server-side in the component logic
    return true
  }),
)

const placedOrders = computed(() => orders.value)

const orderStatusVariant = (status: string) => {
  if (status === 'confirmed') return 'default'
  if (status === 'completed') return 'outline'
  return 'destructive'
}

const orderStatusIcon = (status: string) => {
  if (status === 'confirmed') return Clock3
  if (status === 'completed') return CheckCircle2
  return XCircle
}

const fetchOrders = async () => {
  isLoading.value = true
  try {
    const token = localStorage.getItem('access_token') ?? ''
    const res = await fetch('/api/orders', {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (res.ok) {
      const json = await res.json()
      orders.value = json.data ?? []
    }
  } catch {
    toast.error('Unable to load orders')
  } finally {
    isLoading.value = false
  }
}

const cancelOrder = async (orderId: number) => {
  isCancelling.value = orderId
  try {
    const token = localStorage.getItem('access_token') ?? ''
    const res = await fetch(`/api/orders/${orderId}/cancel`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    })
    if (res.ok) {
      toast.success('Order cancelled')
      await fetchOrders()
    } else {
      const err = await res.json().catch(() => ({}))
      toast.error(err.message || err.error || 'Unable to cancel')
    }
  } catch {
    toast.error('Unable to cancel order')
  } finally {
    isCancelling.value = null
  }
}

onMounted(fetchOrders)
</script>

<template>
  <div class="space-y-6">
    <section class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-emerald-50 via-white to-sky-50 p-6">
      <div class="pointer-events-none absolute -right-10 -top-8 h-32 w-32 rounded-full bg-emerald-200/50" />
      <div class="pointer-events-none absolute -bottom-12 -left-8 h-40 w-40 rounded-full bg-sky-200/40" />
      <div class="relative">
        <p class="text-xs font-semibold uppercase tracking-widest text-emerald-600">Orders</p>
        <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">My Orders</h1>
        <p class="mt-2 max-w-xl text-sm text-slate-600">
          Track your placed orders and manage incoming orders on your items.
        </p>
      </div>
    </section>

    <Tabs default-value="placed" class="space-y-4">
      <TabsList>
        <TabsTrigger value="placed" class="gap-1.5">
          <ShoppingBag class="h-3.5 w-3.5" /> Placed Orders
        </TabsTrigger>
      </TabsList>

      <TabsContent value="placed">
        <div v-if="isLoading" class="py-12 text-center">
          <Loader2 class="mx-auto h-6 w-6 animate-spin text-muted-foreground" />
          <p class="mt-2 text-sm text-muted-foreground">Loading orders...</p>
        </div>

        <div v-else-if="orders.length === 0" class="space-y-4">
          <Card class="bg-card/80 py-12">
            <CardContent class="text-center">
              <ShoppingBag class="mx-auto h-12 w-12 text-muted-foreground/30" />
              <p class="mt-3 text-lg font-semibold text-muted-foreground">No orders yet</p>
              <p class="mt-1 text-sm text-muted-foreground">
                Browse the marketplace and place your first order.
              </p>
              <Button class="mt-4" @click="router.push('/')">
                Browse Marketplace
              </Button>
            </CardContent>
          </Card>
        </div>

        <div v-else class="space-y-3">
          <Card v-for="order in orders" :key="order.id"
            class="bg-card/80 overflow-hidden transition-all hover:shadow-md"
            :class="order.order_status === 'completed' ? 'border-emerald-300/60' : (order.order_status === 'cancelled') ? 'border-red-200/40 opacity-70' : ''">
            <div class="flex flex-col sm:flex-row">
              <div class="flex-1 p-4 space-y-3">
                <div class="flex items-center justify-between gap-2 flex-wrap">
                  <Badge :variant="orderStatusVariant(order.order_status)" class="gap-1">
                    <component :is="orderStatusIcon(order.order_status)" class="h-3 w-3" />
                    {{ order.order_status }}
                  </Badge>
                  <span v-if="order.cash_amount > 0" class="text-sm font-semibold text-amber-600">
                    {{ new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(order.cash_amount) }}
                  </span>
                </div>

                <div>
                  <p class="font-semibold">{{ order.item_name }}</p>
                  <p class="text-xs text-muted-foreground mt-0.5">
                    <span v-if="order.trade_in_item_name" class="text-emerald-600">
                      Trade-in: {{ order.trade_in_item_name }}
                    </span>
                  </p>
                </div>

                <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
                  <span class="flex items-center gap-1">
                    From <span class="font-medium text-foreground">{{ order.seller_name }}</span>
                  </span>
                  <span class="flex items-center gap-1">
                    <CreditCard class="h-3 w-3" />
                    {{ order.payment_method === 'wallet' ? 'Wallet' : 'COD' }}
                    — {{ order.payment_status }}
                  </span>
                  <span>{{ new Date(order.created_at).toLocaleDateString() }}</span>
                </div>

                <div v-if="order.order_status === 'confirmed'" class="flex gap-2 pt-1">
                  <Button
                    size="sm"
                    variant="ghost"
                    class="text-xs text-muted-foreground hover:text-destructive underline underline-offset-2"
                    :disabled="isCancelling === order.id"
                    @click="cancelOrder(order.id)"
                  >
                    <Loader2 v-if="isCancelling === order.id" class="h-3 w-3 animate-spin mr-1" />
                    Cancel Order
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    class="text-xs"
                    @click="router.push('/appointments')"
                  >
                    View Appointment
                  </Button>
                </div>

                <div v-else-if="order.order_status === 'completed'" class="pt-1">
                  <p class="text-xs text-emerald-600 font-medium">Trade completed successfully!</p>
                </div>
                <div v-else-if="order.order_status === 'cancelled'" class="pt-1">
                  <p class="text-xs text-muted-foreground">This order has been cancelled.</p>
                </div>
              </div>
            </div>
          </Card>
        </div>
      </TabsContent>
    </Tabs>
  </div>
</template>
