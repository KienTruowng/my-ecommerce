<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { toast } from 'vue-sonner'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import {
  Package,
  Gavel,
  MapPin,
  ImageOff,
  ShoppingCart,
  ArrowRight,
  Loader2,
} from 'lucide-vue-next'
import ImageLightbox from '@/components/ImageLightbox.vue'
import { RouterLink } from 'vue-router'

type ItemImage = { id: number; url: string }
type Item = {
  id: number
  name: string
  description: string
  status: string
  type?: string
  images?: ItemImage[]
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()
const itemIdInput = ref(String(route.params.id || ''))
const itemDetail = ref<any | null>(null)
const isLoading = ref(false)
const isAddingToCart = ref(false)

const getPrimaryImageUrl = (images?: ItemImage[]) => images?.[0]?.url ?? ''

const resolvedImages = computed(() => {
  return (
    itemDetail.value?.ItemImage ??
    itemDetail.value?.images ??
    itemDetail.value?.item_images ??
    []
  )
})

const lightboxOpen = ref(false)
const lightboxImages = ref<{ src: string; alt: string }[]>([])
const lightboxStartIdx = ref(0)

const openLightbox = (idx: number) => {
  lightboxImages.value = resolvedImages.value.map((img: any) => ({
    src: img.URL ?? img.url,
    alt: itemDetail.value?.Name ?? itemDetail.value?.name ?? 'Item image',
  }))
  lightboxStartIdx.value = idx
  lightboxOpen.value = true
}

const isAuction = computed(() => {
  const t = itemDetail.value?.type ?? itemDetail.value?.Type ?? ''
  return t === 'auction'
})

const isMyItem = computed(() => {
  if (!itemDetail.value || !userStore.profile) return false
  const ownerId = itemDetail.value?.UserID ?? itemDetail.value?.user_id ?? itemDetail.value?.owner_id
  return ownerId === userStore.profile.id
})

const canAddToCart = computed(() => {
  if (!itemDetail.value || isAuction.value || isMyItem.value) return false
  const status = itemDetail.value?.Status ?? itemDetail.value?.status ?? ''
  return status === 'available'
})

const isInCart = computed(() => {
  if (!itemDetail.value) return false
  const id = itemDetail.value?.ID ?? itemDetail.value?.id
  return cartStore.isInCart(id)
})

type OfferTypeInfo = {
  type: string
  distance_km: number
  message: string
  suggested_locations?: { ID: number; Name: string; Description: string; Address: string; Latitude: number; Longitude: number }[]
  shipping_options?: { service: { id: number; name: string; description: string; base_price: number; price_per_km: number; estimated_days: string }; cost: number }[]
}
const offerType = ref<OfferTypeInfo | null>(null)
const isLoadingOfferType = ref(false)

const fetchOfferType = async () => {
  if (!itemDetail.value || isMyItem.value || isAuction.value) return
  const profile = userStore.profile as any
  if (!profile?.latitude || !profile?.longitude) return

  const sellerLat = itemDetail.value?.User?.Latitude ?? itemDetail.value?.user?.latitude
  const sellerLng = itemDetail.value?.User?.Longitude ?? itemDetail.value?.user?.longitude
  if (!sellerLat || !sellerLng) return

  isLoadingOfferType.value = true
  try {
    const params = new URLSearchParams({
      buyer_lat: String(profile.latitude),
      buyer_lng: String(profile.longitude),
      seller_lat: String(sellerLat),
      seller_lng: String(sellerLng),
      has_item: 'false',
    })
    const res = await fetch(`/api/offers/type?${params}`)
    if (res.ok) {
      offerType.value = await res.json()
    }
  } finally {
    isLoadingOfferType.value = false
  }
}

const fetchItem = async (id: string) => {
  if (!id) return
  isLoading.value = true
  try {
    const numericId = Number(id)
    if (Number.isNaN(numericId)) {
      toast.error('Invalid item ID')
      return
    }
    const { data, error } = await api.GET('/api/items/{id}', {
      params: { path: { id: numericId } },
    })
    if (error) {
      toast.error('Unable to load item')
      return
    }
    itemDetail.value = data?.data ?? data
    await fetchOfferType()
  } finally {
    isLoading.value = false
  }
}

const handleLookup = () => {
  if (!itemIdInput.value.trim()) return
  router.push(`/items/${itemIdInput.value.trim()}`)
}

const handleAddToCart = async () => {
  if (!itemDetail.value) return
  const id: number = itemDetail.value?.ID ?? itemDetail.value?.id
  if (!id) return
  isAddingToCart.value = true
  try {
    await cartStore.addToCart(id)
  } finally {
    isAddingToCart.value = false
  }
}

const handleGoToCart = () => {
  router.push('/cart')
}

watch(
  () => route.params.id,
  (value) => {
    itemIdInput.value = String(value || '')
    fetchItem(itemIdInput.value)
  },
)

onMounted(async () => {
  await cartStore.fetchCart()
  await fetchItem(itemIdInput.value)
})
</script>

<template>
  <div class="space-y-6">
    <header class="space-y-2">
      <h1 class="text-2xl font-semibold">Item detail</h1>
      <p class="text-sm text-muted-foreground">
        Look up any item by ID and add it to your cart.
      </p>
    </header>

    <Card class="bg-card/80">
      <CardHeader>
        <CardTitle>Lookup item</CardTitle>
        <CardDescription>Use the item ID shared by another student.</CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col gap-3 sm:flex-row">
        <Input v-model="itemIdInput" placeholder="Item ID" />
        <Button @click="handleLookup">Open item</Button>
      </CardContent>
    </Card>

    <section v-if="isLoading" class="text-sm text-muted-foreground">Loading item...</section>

    <section v-else-if="itemDetail" class="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
      <Card class="bg-card/80">
        <CardHeader>
          <div class="flex flex-wrap items-center justify-between gap-2">
            <CardTitle>{{ itemDetail.Name ?? itemDetail.name }}</CardTitle>
            <div class="flex items-center gap-2">
              <Badge
                v-if="isAuction"
                class="bg-amber-500 text-white gap-1"
              >
                <Gavel class="h-3 w-3" /> Auction
              </Badge>
              <Badge variant="secondary">{{ itemDetail.Status ?? itemDetail.status }}</Badge>
            </div>
          </div>
          <CardDescription>{{ itemDetail.Description ?? itemDetail.description }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="flex flex-wrap gap-2">
            <Badge
              v-for="category in itemDetail.Categories ?? itemDetail.categories ?? []"
              :key="category.ID ?? category.id"
              variant="outline"
            >
              {{ category.Name ?? category.name }}
            </Badge>
          </div>
          <div v-if="resolvedImages.length > 0" class="space-y-2">
            <p class="text-xs font-medium text-muted-foreground">
              {{ resolvedImages.length }} image{{ resolvedImages.length !== 1 ? 's' : '' }}
            </p>
            <div class="flex gap-3 overflow-x-auto pb-2 scrollbar-thin">
              <img
                v-for="(image, imgIdx) in resolvedImages"
                :key="image.ID ?? image.id"
                :src="image.URL ?? image.url"
                :alt="itemDetail.Name ?? itemDetail.name"
                class="h-48 w-auto max-w-xs shrink-0 rounded-xl border object-cover shadow-sm cursor-pointer transition-transform hover:scale-105"
                @click="openLightbox(imgIdx)"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Auction CTA -->
      <Card v-if="isAuction" class="bg-card/80 border-amber-300/40">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Gavel class="h-5 w-5 text-amber-500" />
            Auction
          </CardTitle>
          <CardDescription>
            This item is sold via auction. Join to place bids and win the item.
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <Button
            class="w-full gap-2 bg-amber-500 hover:bg-amber-600 text-white"
            @click="router.push(`/items/${itemDetail.ID ?? itemDetail.id}/auction`)"
          >
            <Gavel class="h-4 w-4" />
            Join Auction
          </Button>
        </CardContent>
      </Card>

      <!-- Own item notice -->
      <Card v-else-if="isMyItem" class="bg-card/80">
        <CardHeader>
          <CardTitle class="text-base">Your own item</CardTitle>
          <CardDescription>
            This item belongs to you. You cannot purchase your own listing.
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="rounded-xl bg-muted/60 border p-3 text-sm text-muted-foreground">
            <p>Want to manage this item?</p>
            <Button variant="outline" size="sm" class="mt-2 gap-1.5" as-child>
              <RouterLink to="/inventory">
                <Package class="h-3.5 w-3.5" />
                Go to My Inventory
              </RouterLink>
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- Add to Cart (available items) -->
      <Card v-else-if="canAddToCart" class="bg-card/80">
        <CardHeader class="pb-3">
          <CardTitle>Add to Cart</CardTitle>
          <CardDescription>
            <span v-if="offerType && !isLoadingOfferType" class="flex items-center gap-1">
              <MapPin class="h-3 w-3 text-emerald-500" />
              {{ offerType.distance_km.toFixed(1) }} km away
            </span>
            <span v-else>Add this item to your cart to proceed to checkout.</span>
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <!-- Offer type banner -->
          <div v-if="offerType && !isLoadingOfferType" class="space-y-2">
            <div v-if="offerType.type === 'meetup'" class="rounded-xl border border-emerald-200 bg-emerald-50/60 p-3 dark:border-emerald-800 dark:bg-emerald-950/40">
              <p class="text-xs font-semibold text-emerald-700 dark:text-emerald-300">Meet in Person</p>
              <p class="text-[11px] text-emerald-600 dark:text-emerald-400 mt-1">{{ offerType.message }}</p>
              <div v-if="offerType.suggested_locations?.length" class="mt-2 flex flex-wrap gap-1.5">
                <span v-for="loc in offerType.suggested_locations.slice(0, 3)" :key="loc.ID"
                  class="rounded-full bg-white/80 dark:bg-white/10 px-2 py-0.5 text-[10px] border">
                  {{ loc.Name }}
                </span>
              </div>
            </div>
            <div v-else-if="offerType.type === 'campus_trade'" class="rounded-xl border border-sky-200 bg-sky-50/60 p-3 dark:border-sky-800 dark:bg-sky-950/40">
              <p class="text-xs font-semibold text-sky-700 dark:text-sky-300">Campus Exchange</p>
              <p class="text-[11px] text-sky-600 dark:text-sky-400 mt-1">{{ offerType.message }}</p>
            </div>
            <div v-else-if="offerType.type === 'shipping'" class="rounded-xl border border-amber-200 bg-amber-50/60 p-3 dark:border-amber-800 dark:bg-amber-950/40">
              <p class="text-xs font-semibold text-amber-700 dark:text-amber-300">Shipping Required</p>
              <p class="text-[11px] text-amber-600 dark:text-amber-400 mt-1">{{ offerType.message }}</p>
              <div v-if="offerType.shipping_options?.length" class="mt-2 space-y-1.5">
                <div v-for="opt in offerType.shipping_options.slice(0, 2)" :key="opt.service.id" class="flex items-center justify-between rounded-lg bg-white/60 dark:bg-white/10 px-2 py-1 text-[10px]">
                  <span>{{ opt.service.name }} ({{ opt.service.estimated_days }} days)</span>
                  <span class="font-semibold text-amber-600">{{ new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(opt.cost) }}</span>
                </div>
              </div>
            </div>
            <div v-else-if="offerType.type === 'too_far'" class="rounded-xl border border-red-200 bg-red-50/60 p-3 dark:border-red-800 dark:bg-red-950/40">
              <p class="text-xs font-semibold text-red-700 dark:text-red-300">Too Far</p>
              <p class="text-[11px] text-red-600 dark:text-red-400 mt-1">{{ offerType.message }}</p>
            </div>
          </div>

          <Button
            v-if="isInCart"
            class="w-full gap-2 bg-emerald-600 hover:bg-emerald-700"
            @click="handleGoToCart"
          >
            <ShoppingCart class="h-4 w-4" />
            View in Cart
            <ArrowRight class="h-4 w-4" />
          </Button>
          <Button
            v-else
            class="w-full gap-2"
            :disabled="isAddingToCart || offerType?.type === 'too_far'"
            @click="handleAddToCart"
          >
            <Loader2 v-if="isAddingToCart" class="h-4 w-4 animate-spin" />
            <ShoppingCart v-else class="h-4 w-4" />
            Add to Cart
          </Button>
        </CardContent>
      </Card>

      <!-- Item not available -->
      <Card v-else class="bg-card/80">
        <CardHeader>
          <CardTitle class="text-base">Item not available</CardTitle>
          <CardDescription>
            This item is no longer available. It may have been reserved or traded already.
          </CardDescription>
        </CardHeader>
      </Card>
    </section>

    <section v-else class="text-sm text-muted-foreground">Item not found.</section>
  </div>

  <ImageLightbox
    v-model="lightboxOpen"
    :images="lightboxImages"
    :start-index="lightboxStartIdx"
  />
</template>
