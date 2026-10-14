<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/api'
import { useCartStore } from '@/stores/cart'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { toast } from 'vue-sonner'
import {
  Package,
  Trash2,
  ShoppingCart,
  CreditCard,
  Banknote,
  Loader2,
  ImageOff,
} from 'lucide-vue-next'

type ItemImage = { id: number; url: string }
type Item = {
  id: number
  name: string
  description: string
  status: string
  type?: string
  images?: ItemImage[]
}

const router = useRouter()
const cartStore = useCartStore()

const myItems = ref<Item[]>([])
const selectedTradeItemId = ref<number | null>(null)
const cashAmount = ref(0)
const paymentMethod = ref('wallet')
const isCheckingOut = ref(false)
const fetchError = ref('')

const hasCart = computed(() => !!cartStore.cart)

const getPrimaryImageUrl = (images?: ItemImage[]) => images?.[0]?.url ?? ''

const fetchMyItems = async () => {
  const { data, error } = await api.GET('/api/users/me/items')
  if (!error) {
    const list = Array.isArray(data) ? (data as Item[]) : []
    myItems.value = list.filter((i) => i.status === 'available')
  }
}

const loadCart = async () => {
  try {
    await cartStore.fetchCart()
    if (cartStore.cart) {
      selectedTradeItemId.value = cartStore.cart.offer_item_id ?? null
      cashAmount.value = cartStore.cart.cash_amount ?? 0
      if (!cartStore.cart.item?.name && cartStore.cart.item_id) {
        const { data } = await api.GET('/api/items/{id}', {
          params: { path: { id: cartStore.cart.item_id } },
        })
        if (data && !(data as any).error) {
          const item: any = (data as any)?.data ?? data
          cartStore.cart.item = {
            id: item?.ID ?? item?.id ?? cartStore.cart.item_id,
            name: item?.Name ?? item?.name ?? 'Item #' + cartStore.cart.item_id,
            description: item?.Description ?? item?.description ?? '',
            status: item?.Status ?? item?.status ?? 'unknown',
            type: item?.Type ?? item?.type ?? 'barter',
            images: (item?.ItemImage ?? item?.images ?? []).map((img: any) => ({
              id: img?.ID ?? img?.id ?? 0,
              url: img?.URL ?? img?.url ?? '',
            })),
          }
        }
      }
    }
    fetchError.value = ''
  } catch (e: any) {
    fetchError.value = e?.message || 'Failed to load cart'
  }
}

onMounted(async () => {
  await loadCart()
  await fetchMyItems()
})

const handleRemoveFromCart = async () => {
  await cartStore.clearCart()
  toast.success('Removed from cart')
}

const handleUpdateCart = async () => {
  await cartStore.updateCart(selectedTradeItemId.value, cashAmount.value)
  await loadCart()
  toast.success('Cart updated')
}

const handleCheckout = async () => {
  isCheckingOut.value = true
  try {
    await cartStore.checkout(paymentMethod.value)
    toast.success('Order placed successfully!')
    router.push('/orders')
  } catch (err: any) {
    toast.error(err.message || 'Checkout failed')
  } finally {
    isCheckingOut.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <header class="space-y-2">
      <h1 class="text-2xl font-semibold">Shopping Cart</h1>
      <p class="text-sm text-muted-foreground">Review your item and proceed to checkout.</p>
    </header>

    <div v-if="cartStore.isLoading" class="py-12 text-center">
      <Loader2 class="mx-auto h-6 w-6 animate-spin text-muted-foreground" />
      <p class="mt-3 text-sm text-muted-foreground">Loading cart...</p>
    </div>

    <div v-else-if="fetchError" class="space-y-4">
      <Card class="bg-card/80 py-12">
        <CardContent class="text-center">
          <p class="text-red-500">{{ fetchError }}</p>
          <Button class="mt-4" variant="outline" @click="loadCart">Retry</Button>
        </CardContent>
      </Card>
    </div>

    <div v-else-if="!hasCart" class="space-y-4">
      <Card class="bg-card/80 py-12">
        <CardContent class="text-center">
          <ShoppingCart class="mx-auto h-12 w-12 text-muted-foreground/30" />
          <p class="mt-3 text-lg font-semibold text-muted-foreground">Your cart is empty</p>
          <p class="mt-1 text-sm text-muted-foreground">
            Browse the marketplace and add items to get started.
          </p>
          <Button class="mt-4" @click="router.push('/')">
            Browse Marketplace
          </Button>
        </CardContent>
      </Card>
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-[1fr_380px]">
      <Card class="bg-card/80">
        <CardHeader>
          <CardTitle>Cart Item</CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="flex items-start gap-4">
            <div class="h-24 w-24 shrink-0 rounded-xl border bg-background overflow-hidden">
              <img
                v-if="cartStore.cart?.item?.images?.length"
                :src="getPrimaryImageUrl(cartStore.cart.item.images)"
                :alt="cartStore.cart.item.name"
                class="h-full w-full object-cover"
              />
              <div v-else class="flex h-full items-center justify-center">
                <ImageOff class="h-8 w-8 text-muted-foreground/30" />
              </div>
            </div>
            <div class="flex-1 min-w-0">
              <h3 class="text-lg font-semibold">{{ cartStore.cart?.item?.name ?? 'Unknown item' }}</h3>
              <p class="text-sm text-muted-foreground line-clamp-2 mt-1">
                {{ cartStore.cart?.item?.description ?? 'No description' }}
              </p>
              <Badge variant="secondary" class="mt-2">{{ cartStore.cart?.item?.status ?? 'unknown' }}</Badge>
            </div>
          </div>

          <Button variant="outline" class="w-full gap-2 text-muted-foreground hover:text-destructive" @click="handleRemoveFromCart">
            <Trash2 class="h-4 w-4" /> Remove from cart
          </Button>
        </CardContent>
      </Card>

      <div class="space-y-4">
        <Card class="bg-card/80">
          <CardHeader>
            <CardTitle class="text-base">Trade-in (optional)</CardTitle>
            <CardDescription>Select one of your items to trade for a discount.</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <select
              v-model="selectedTradeItemId"
              class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
            >
              <option :value="null">No trade-in — cash only</option>
              <option v-for="mi in myItems" :key="mi.id" :value="mi.id">
                {{ mi.name }}
              </option>
            </select>

            <div class="space-y-2">
              <label class="text-sm font-medium">Cash amount (VND)</label>
              <Input v-model.number="cashAmount" type="number" min="0" placeholder="0" />
            </div>

            <Button variant="outline" class="w-full" @click="handleUpdateCart">
              Update cart
            </Button>
          </CardContent>
        </Card>

        <Card class="bg-card/80">
          <CardHeader>
            <CardTitle class="text-base">Payment</CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="flex flex-col gap-2">
              <label class="flex items-center gap-3 rounded-lg border p-3 cursor-pointer transition-all"
                :class="paymentMethod === 'wallet' ? 'border-indigo-400 bg-indigo-50 ring-1 ring-indigo-200 dark:bg-indigo-950' : 'hover:border-indigo-200'">
                <input type="radio" v-model="paymentMethod" value="wallet" class="sr-only" />
                <CreditCard class="h-5 w-5 text-indigo-600" />
                <div>
                  <p class="text-sm font-semibold">Wallet</p>
                  <p class="text-xs text-muted-foreground">Pay from your balance</p>
                </div>
              </label>
              <label class="flex items-center gap-3 rounded-lg border p-3 cursor-pointer transition-all"
                :class="paymentMethod === 'cod' ? 'border-indigo-400 bg-indigo-50 ring-1 ring-indigo-200 dark:bg-indigo-950' : 'hover:border-indigo-200'">
                <input type="radio" v-model="paymentMethod" value="cod" class="sr-only" />
                <Banknote class="h-5 w-5 text-amber-600" />
                <div>
                  <p class="text-sm font-semibold">Cash on Meetup</p>
                  <p class="text-xs text-muted-foreground">Pay in person when you meet</p>
                </div>
              </label>
            </div>

            <div class="rounded-lg bg-muted/40 p-3 text-sm">
              <div class="flex justify-between">
                <span class="text-muted-foreground">Cash payment</span>
                <span class="font-semibold">{{ cashAmount ? new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(cashAmount) : 'Free trade' }}</span>
              </div>
              <div v-if="selectedTradeItemId" class="flex justify-between mt-1">
                <span class="text-muted-foreground">Trade-in</span>
                <span class="font-semibold text-emerald-600">{{ myItems.find(i => i.id === selectedTradeItemId)?.name ?? 'Item' }}</span>
              </div>
            </div>

            <Button class="w-full gap-2" size="lg" :disabled="isCheckingOut" @click="handleCheckout">
              <Loader2 v-if="isCheckingOut" class="h-4 w-4 animate-spin" />
              <ShoppingCart v-else class="h-4 w-4" />
              {{ isCheckingOut ? 'Processing...' : 'Place Order' }}
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
