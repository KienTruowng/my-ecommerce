<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { api } from '@/api/api'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import router from '@/router'
import { toast } from 'vue-sonner'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import ImageLightbox from '@/components/ImageLightbox.vue'
import {
  Search,
  Package,
  Tag,
  ChevronLeft,
  ChevronRight,
  X,
  ArrowLeftRight,
  Loader2,
  ImageOff,
  Gavel,
  Bookmark,
  BookmarkCheck,
  ShoppingCart,
  ArrowRight,
} from 'lucide-vue-next'

// ─── Types ────────────────────────────────────────────────────────────────────
type ItemImage = { id: number; url: string }
type Category = { id: number; name: string }
type Item = {
  id: number
  name: string
  description: string
  status: string
  type: string
  owner_id: number
  created_at: string
  images: ItemImage[]
  categories: Category[]
}

type AuctionMeta = {
  itemId: number
  basePrice: number
  stepPrice: number
  depositAmount: number
  deadline: string
  bidsCount: number
  highestBid: number
  status: string
}

// ─── State ────────────────────────────────────────────────────────────────────
const userStore = useUserStore()
const cartStore = useCartStore()

const items = ref<Item[]>([])
const categories = ref<Category[]>([])
const isLoading = ref(false)
const auctionLoading = ref(false)
const auctionMeta = ref<Record<number, AuctionMeta>>({})
const search = ref('')
const selectedCat = ref<number | null>(null)
const page = ref(1)
const totalPages = ref(1)
const totalItems = ref(0)
const LIMIT = 20

const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | null = null

const selectedItem = ref<Item | null>(null)
const panelOpen = ref(false)

// Lightbox
const lightboxOpen = ref(false)
const lightboxImages = ref<{ src: string; alt: string }[]>([])
const lightboxStartIdx = ref(0)

const openLightbox = (images: { src: string; alt: string }[], idx: number) => {
  lightboxImages.value = images
  lightboxStartIdx.value = idx
  lightboxOpen.value = true
}

// ─── Computed ─────────────────────────────────────────────────────────────────
const filteredItems = computed(() => {
  if (!search.value.trim()) return items.value
  const q = search.value.toLowerCase()
  return items.value.filter(
    (i) => i.name.toLowerCase().includes(q) || i.description?.toLowerCase().includes(q),
  )
})

const allAuctionItems = computed(() => items.value.filter((item) => item.type === 'auction'))
const auctionItems = computed(() => filteredItems.value.filter((item) => item.type === 'auction'))
// Filter out settled/cancelled/expired auctions
const liveAuctionItems = computed(() =>
  auctionItems.value.filter((item) => {
    const meta = auctionMeta.value[item.id]
    if (!meta) return true // still loading, show it
    if (meta.status !== 'open') return false
    if (meta.deadline && new Date(meta.deadline).getTime() < now.value) return false
    return true
  }),
)
const tradeItems = computed(() => filteredItems.value.filter((item) => item.type !== 'auction'))

const formatVnd = (value: number) =>
  new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(value)

const formatCountdown = (deadline?: string) => {
  if (!deadline) return 'No deadline'
  const diff = new Date(deadline).getTime() - now.value
  if (diff <= 0) return 'Ended'
  const hours = Math.floor(diff / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)
  return `${hours}h ${minutes}m left`
}

const getAuctionMeta = (itemId: number) => auctionMeta.value[itemId]

const isMyItem = computed(() => {
  if (!selectedItem.value || !userStore.profile) return false
  return selectedItem.value.owner_id === userStore.profile.id
})

const canOffer = computed(
  () => !isMyItem.value && selectedItem.value?.status === 'available' && !!userStore.profile,
)

const isInSelectedCart = computed(() => {
  if (!selectedItem.value) return false
  return cartStore.isInCart(selectedItem.value.id)
})

// ─── Data fetching ────────────────────────────────────────────────────────────
const fetchMarketplace = async () => {
  isLoading.value = true
  try {
    const url = new URL(`/api/items`, window.location.origin)
    url.searchParams.set('page', String(page.value))
    url.searchParams.set('limit', String(LIMIT))
    if (selectedCat.value) url.searchParams.set('category_id', String(selectedCat.value))

    const [itemsRes, catsRes] = await Promise.all([
      fetch(url.toString()).then((r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`)
        return r.json()
      }),
      api.GET('/api/categories', { params: { query: { page: 1, limit: 50 } } }),
    ])

    items.value = itemsRes?.data ?? []
    totalPages.value = itemsRes?.meta?.total_pages ?? 1
    totalItems.value = itemsRes?.meta?.total_records ?? 0

    await loadAuctionMeta(allAuctionItems.value)

    if (!catsRes.error && Array.isArray(catsRes.data?.data)) {
      categories.value = catsRes.data?.data as Category[]
    }
  } catch (e) {
    console.error('Marketplace fetch error:', e)
    toast.error('Unable to load marketplace')
  } finally {
    isLoading.value = false
  }
}

watch([selectedCat, page], fetchMarketplace)

const loadAuctionMeta = async (auctionItemsList: Item[]) => {
  if (!auctionItemsList.length) {
    auctionMeta.value = {}
    return
  }

  auctionLoading.value = true
  const targets = auctionItemsList.slice(0, 6)
  const requests = targets.map((item) =>
    fetch(`/api/items/${item.id}/auction`)
      .then((res) => (res.ok ? res.json() : null))
      .then((json) => {
        const data = json?.data
        if (!data) return null
        const bids = Array.isArray(data.Bids) ? data.Bids : []
        const highestBid = bids.reduce(
          (max: number, bid: { amount?: number }) => Math.max(max, bid.amount ?? 0),
          0,
        )
        return {
          itemId: item.id,
          basePrice: data.base_price ?? 0,
          stepPrice: data.step_price ?? 0,
          depositAmount: data.deposit_amount ?? 0,
          deadline: data.deadline ?? '',
          bidsCount: bids.length,
          highestBid,
          status: data.status ?? 'open',
        } as AuctionMeta
      })
      .catch(() => null),
  )

  const results = await Promise.allSettled(requests)
  const nextMeta: Record<number, AuctionMeta> = {}
  results.forEach((result) => {
    if (result.status === 'fulfilled' && result.value) {
      nextMeta[result.value.itemId] = result.value
    }
  })
  auctionMeta.value = nextMeta
  auctionLoading.value = false
}

// ─── Actions ──────────────────────────────────────────────────────────────────
const selectCategory = (id: number | null) => {
  selectedCat.value = id
  page.value = 1
}

const openItem = (item: Item) => {
  if (item.type === 'auction') {
    router.push(`/items/${item.id}/auction`)
    return
  }
  selectedItem.value = item
  panelOpen.value = true
}

const closePanel = () => {
  panelOpen.value = false
  selectedItem.value = null
}

const goToDetail = () => {
  if (selectedItem.value) router.push(`/items/${selectedItem.value.id}`)
}

const addToCart = async (item: Item) => {
  if (cartStore.isInCart(item.id)) {
    await cartStore.clearCart()
    toast('Removed from your cart')
  } else {
    await cartStore.addToCart(item.id)
  }
}

const goToCart = () => {
  router.push('/cart')
}

const statusColor = (status: string) => {
  const map: Record<string, string> = {
    available: 'bg-emerald-500/15 text-emerald-600',
    reserved: 'bg-amber-500/15 text-amber-600',
    traded: 'bg-sky-500/15 text-sky-600',
    hidden: 'bg-zinc-500/15 text-zinc-500',
  }
  return map[status] ?? 'bg-muted text-muted-foreground'
}

const getPrimaryImageUrl = (images?: ItemImage[]) => images?.[0]?.url ?? ''

onMounted(async () => {
  await userStore.fetchProfile().catch(() => { })
  cartStore.fetchCart()
  fetchMarketplace()

  timer = setInterval(() => {
    now.value = Date.now()
  }, 60000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  timer = null
})
</script>

<template>
  <div class="space-y-8 pb-12">
    <!-- ── Hero ── -->
    <section
      class="relative overflow-hidden rounded-3xl border bg-gradient-to-br from-emerald-500/20 via-cyan-500/10 to-transparent p-8 md:p-10">
      <div class="relative z-10 max-w-2xl space-y-4">
        <Badge class="bg-emerald-600 text-white gap-1">
          <Package class="h-3 w-3" />
          Campus Marketplace
        </Badge>
        <h1 class="text-3xl font-semibold md:text-4xl">Browse what students are trading</h1>
        <p class="text-sm text-muted-foreground md:text-base max-w-xl">
          Discover items listed by verified students. <strong class="text-foreground">Save to your cart</strong> to plan trades,
          compare offers, and coordinate meetups — all in one place.
        </p>
        <div class="flex flex-wrap items-center gap-3 pt-2">
          <span class="inline-flex items-center gap-1.5 rounded-full bg-emerald-100 px-3 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300">
            <Bookmark class="h-3 w-3" />
            {{ cartStore.count }} in cart
          </span>
          <span class="text-xs text-muted-foreground">
            🛒 Use <strong>Your Cart</strong> in the sidebar to plan trades
          </span>
        </div>
      </div>
      <div class="absolute -right-16 -top-16 h-40 w-40 rounded-full bg-emerald-400/30 blur-3xl" />
      <div class="absolute right-20 bottom-0 h-32 w-32 rounded-full bg-cyan-400/20 blur-2xl" />
    </section>

    <!-- ── Quick Stats ── -->
    <div class="grid gap-3 sm:grid-cols-3">
      <div class="flex items-center gap-3 rounded-xl border bg-card/80 px-4 py-3">
        <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-100 text-emerald-600 dark:bg-emerald-900 dark:text-emerald-400">
          <Package class="h-5 w-5" />
        </div>
        <div>
          <p class="text-lg font-bold">{{ totalItems }}</p>
          <p class="text-xs text-muted-foreground">Items available</p>
        </div>
      </div>
      <div class="flex items-center gap-3 rounded-xl border bg-card/80 px-4 py-3">
        <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-amber-100 text-amber-600 dark:bg-amber-900 dark:text-amber-400">
          <Gavel class="h-5 w-5" />
        </div>
        <div>
          <p class="text-lg font-bold">{{ liveAuctionItems.length }}</p>
          <p class="text-xs text-muted-foreground">Live auctions</p>
        </div>
      </div>
      <div class="flex items-center gap-3 rounded-xl border bg-card/80 px-4 py-3">
        <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-sky-100 text-sky-600 dark:bg-sky-900 dark:text-sky-400">
          <BookmarkCheck class="h-5 w-5" />
        </div>
        <div>
          <p class="text-lg font-bold">{{ cartStore.count }}</p>
          <p class="text-xs text-muted-foreground">Items in cart</p>
        </div>
      </div>
    </div>

    <!-- ── Search + Filter bar ── -->
    <section class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <!-- Search -->
      <div class="relative flex-1 max-w-sm">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input v-model="search" placeholder="Search items…" class="pl-9" />
      </div>

      <!-- Category pills -->
      <div class="flex flex-wrap gap-2">
        <button @click="selectCategory(null)" :class="[
          'inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium border transition-colors',
          selectedCat === null
            ? 'bg-emerald-600 text-white border-emerald-600'
            : 'bg-background border-border hover:bg-muted',
        ]">
          <Tag class="h-3 w-3" /> All
        </button>
        <button v-for="cat in categories" :key="cat.id" @click="selectCategory(cat.id)" :class="[
          'inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium border transition-colors',
          selectedCat === cat.id
            ? 'bg-emerald-600 text-white border-emerald-600'
            : 'bg-background border-border hover:bg-muted',
        ]">
          {{ cat.name }}
        </button>
      </div>
    </section>

    <!-- ── Results count ── -->
    <div class="flex items-center justify-between text-sm text-muted-foreground">
      <span>
        <span class="font-semibold text-foreground">{{ totalItems }}</span> items available
        <span v-if="selectedCat">in <span class="font-semibold text-foreground">{{categories.find(c => c.id ===
            selectedCat)?.name }}</span></span>
      </span>
      <span v-if="totalPages > 1">Page {{ page }} / {{ totalPages }}</span>
    </div>

    <!-- ── Auction Spotlight ── -->
    <section class="space-y-4">
      <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h2 class="text-xl font-semibold">Live bidding arena</h2>
          <p class="text-sm text-muted-foreground">
            Jump into active auctions and place bids in real time.
          </p>
        </div>
        <Button variant="outline" class="gap-2" @click="selectCategory(null)">
          <Gavel class="h-4 w-4" /> Browse all auctions
        </Button>
      </div>

      <div v-if="isLoading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="n in 3" :key="n" class="rounded-2xl border bg-card animate-pulse overflow-hidden">
          <div class="h-40 bg-muted" />
          <div class="p-4 space-y-2">
            <div class="h-4 bg-muted rounded w-3/4" />
            <div class="h-3 bg-muted rounded w-1/2" />
          </div>
        </div>
      </div>

      <div v-else-if="liveAuctionItems.length === 0"
        class="rounded-2xl border border-dashed bg-muted/30 p-6 text-sm text-muted-foreground">
        No auctions right now. Create an auction listing to start bidding.
      </div>

      <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <article v-for="item in liveAuctionItems.slice(0, 6)" :key="item.id" @click="openItem(item)"
          class="group relative overflow-hidden rounded-2xl border border-amber-200/60 bg-gradient-to-br from-amber-50 via-white to-amber-100/30 transition-all hover:-translate-y-0.5 hover:shadow-md">
          <div class="relative h-40 overflow-hidden">
            <img v-if="getPrimaryImageUrl(item.images)" :src="getPrimaryImageUrl(item.images)" :alt="item.name"
              class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105" />
            <div v-else class="flex h-full items-center justify-center bg-amber-50">
              <ImageOff class="h-10 w-10 text-amber-200" />
            </div>
            <span
              class="absolute left-3 top-3 inline-flex items-center gap-1 rounded-full bg-amber-500 px-2 py-0.5 text-[11px] font-bold text-white shadow">
              <Gavel class="h-3 w-3" /> Live auction
            </span>
          </div>
          <div class="space-y-2 p-4">
            <h3 class="text-base font-semibold line-clamp-1">{{ item.name }}</h3>
            <p class="text-xs text-muted-foreground line-clamp-2">
              {{ item.description || 'Auction listing' }}
            </p>
            <div class="grid gap-2 rounded-xl border border-amber-200/60 bg-white/80 p-3 text-xs">
              <div class="flex items-center justify-between">
                <span class="text-muted-foreground">Base</span>
                <span class="font-semibold text-foreground">
                  {{ getAuctionMeta(item.id) ? formatVnd(getAuctionMeta(item.id)!.basePrice) : '-' }}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-muted-foreground">Step</span>
                <span class="font-semibold text-foreground">
                  {{ getAuctionMeta(item.id) ? formatVnd(getAuctionMeta(item.id)!.stepPrice) : '-' }}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-muted-foreground">Top bid</span>
                <span class="font-semibold text-foreground">
                  {{ getAuctionMeta(item.id) ? formatVnd(getAuctionMeta(item.id)!.highestBid ||
                    getAuctionMeta(item.id)!.basePrice) : '-' }}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-muted-foreground">Bids</span>
                <span class="font-semibold text-foreground">
                  {{ getAuctionMeta(item.id) ? getAuctionMeta(item.id)!.bidsCount : '-' }}
                </span>
              </div>
            </div>
            <div class="flex items-center justify-between text-xs text-muted-foreground">
              <span v-if="auctionLoading">Loading schedule...</span>
              <span v-else>{{ formatCountdown(getAuctionMeta(item.id)?.deadline) }}</span>
              <Badge variant="secondary" class="text-[10px]">
                {{ getAuctionMeta(item.id)?.status ?? 'open' }}
              </Badge>
            </div>
            <Button size="sm" class="w-full bg-amber-500 text-white hover:bg-amber-600">
              Enter bidding room
            </Button>
          </div>
        </article>
      </div>
    </section>

    <!-- ── Item Grid ── -->
    <section>
      <!-- Skeleton -->
      <div v-if="isLoading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <div v-for="n in 8" :key="n" class="rounded-2xl border bg-card animate-pulse overflow-hidden">
          <div class="h-44 bg-muted" />
          <div class="p-4 space-y-2">
            <div class="h-4 bg-muted rounded w-3/4" />
            <div class="h-3 bg-muted rounded w-1/2" />
            <div class="h-8 bg-muted rounded mt-4" />
          </div>
        </div>
      </div>

      <!-- Empty state -->
      <div v-else-if="filteredItems.length === 0" class="flex flex-col items-center gap-4 py-20 text-muted-foreground">
        <Package class="h-14 w-14 opacity-30" />
        <p class="text-base">No items found. Try a different category or search term.</p>
      </div>

      <!-- Cards -->
      <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <article v-for="item in tradeItems" :key="item.id" @click="openItem(item)"
          class="group relative rounded-2xl border bg-card/80 overflow-hidden cursor-pointer transition-all hover:shadow-md hover:-translate-y-0.5 hover:border-emerald-400/60">
          <!-- Image -->
          <div class="relative h-44 overflow-hidden bg-muted">
            <img v-if="getPrimaryImageUrl(item.images)" :src="getPrimaryImageUrl(item.images)" :alt="item.name"
              class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105" />
            <div v-else class="h-full flex items-center justify-center">
              <ImageOff class="h-10 w-10 text-muted-foreground/30" />
            </div>
            <!-- Status badge -->
            <span class="absolute top-2 right-2 rounded-full px-2 py-0.5 text-[11px] font-semibold backdrop-blur-sm"
              :class="statusColor(item.status)">
              {{ item.status }}
            </span>
            <!-- Cart button -->
            <button
              class="absolute top-2 left-2 flex h-7 w-7 items-center justify-center rounded-full bg-black/30 backdrop-blur-sm text-white transition-all hover:bg-black/50 hover:scale-110"
              :class="cartStore.isInCart(item.id) ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'"
              @click.stop="addToCart(item)"
              :title="cartStore.isInCart(item.id) ? 'Remove from cart' : 'Add to cart'"
            >
              <BookmarkCheck v-if="cartStore.isInCart(item.id)" class="h-4 w-4 text-amber-400" />
              <Bookmark v-else class="h-4 w-4" />
            </button>
          </div>

          <!-- Content -->
          <div class="p-4 space-y-2">
            <h3 class="font-semibold leading-tight line-clamp-1">{{ item.name }}</h3>
            <p class="text-xs text-muted-foreground line-clamp-2 min-h-[2rem]">
              {{ item.description || 'No description provided.' }}
            </p>
            <!-- Categories -->
            <div v-if="item.categories?.length" class="flex flex-wrap gap-1 pt-1">
              <span v-for="cat in item.categories.slice(0, 2)" :key="cat.id"
                class="rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
                {{ cat.name }}
              </span>
              <span v-if="item.categories.length > 2"
                class="rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
                +{{ item.categories.length - 2 }}
              </span>
            </div>
            <div class="flex gap-1.5">
              <Button size="sm" class="flex-1 mt-1" variant="outline">View & Trade</Button>
              <Button
                size="sm"
                variant="ghost"
                class="mt-1 px-2"
                :class="cartStore.isInCart(item.id) ? 'text-amber-500' : ''"
                @click.stop="addToCart(item)"
              >
                <BookmarkCheck v-if="cartStore.isInCart(item.id)" class="h-4 w-4" />
                <Bookmark v-else class="h-4 w-4" />
              </Button>
            </div>
          </div>
        </article>
      </div>
    </section>

    <!-- ── Pagination ── -->
    <div v-if="totalPages > 1" class="flex items-center justify-center gap-2">
      <Button variant="outline" size="sm" :disabled="page === 1" @click="page--">
        <ChevronLeft class="h-4 w-4" />
      </Button>
      <span class="text-sm text-muted-foreground">{{ page }} / {{ totalPages }}</span>
      <Button variant="outline" size="sm" :disabled="page >= totalPages" @click="page++">
        <ChevronRight class="h-4 w-4" />
      </Button>
    </div>

    <!-- ── Item Detail Slide-over Panel ── -->
    <Teleport to="body">
      <Transition name="panel">
        <div v-if="panelOpen && selectedItem" class="fixed inset-0 z-50 flex justify-end" @click.self="closePanel">
          <!-- Backdrop -->
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closePanel" />

          <!-- Panel -->
          <aside class="relative z-10 flex h-full w-full max-w-md flex-col bg-background shadow-2xl overflow-y-auto">
            <!-- Header -->
            <div class="flex items-center justify-between border-b px-5 py-4 shrink-0">
              <h2 class="text-base font-semibold">Item Detail</h2>
              <div class="flex items-center gap-1">
                <Button
                  size="sm"
                  variant="ghost"
                  class="gap-1.5"
                  :class="cartStore.isInCart(selectedItem.id) ? 'text-amber-500' : ''"
                  @click="addToCart(selectedItem)"
                >
                  <BookmarkCheck v-if="cartStore.isInCart(selectedItem.id)" class="h-4 w-4" />
                  <Bookmark v-else class="h-4 w-4" />
                  {{ cartStore.isInCart(selectedItem.id) ? 'In Cart' : 'Add to Cart' }}
                </Button>
                <Button variant="ghost" size="icon" @click="closePanel">
                  <X class="h-4 w-4" />
                </Button>
              </div>
            </div>

            <!-- Image gallery -->
            <div class="relative bg-muted shrink-0">
              <!-- Multiple images: horizontal scroll -->
              <div v-if="selectedItem.images && selectedItem.images.length > 1" class="flex gap-2 overflow-x-auto p-3 pb-2 scrollbar-thin">
                <img
                  v-for="(image, idx) in selectedItem.images"
                  :key="image.id"
                  :src="image.url"
                  :alt="`${selectedItem.name} - image ${idx + 1}`"
                  class="h-52 w-auto max-w-[85%] shrink-0 rounded-xl border object-cover shadow-sm cursor-pointer transition-transform hover:scale-105"
                  @click="selectedItem && openLightbox(
                    (selectedItem.images ?? []).map(img => ({ src: img.url, alt: selectedItem!.name })),
                    idx
                  )"
                />
              </div>
              <!-- Single image -->
              <div v-else-if="getPrimaryImageUrl(selectedItem.images)" class="h-64">
                <img
                  :src="getPrimaryImageUrl(selectedItem.images)"
                  :alt="selectedItem.name"
                  class="h-full w-full object-cover cursor-pointer"
                  @click="selectedItem && getPrimaryImageUrl(selectedItem.images) && openLightbox(
                    [{ src: getPrimaryImageUrl(selectedItem.images)!, alt: selectedItem!.name }],
                    0
                  )"
                />
              </div>
              <!-- No image fallback -->
              <div v-else class="flex h-64 items-center justify-center">
                <ImageOff class="h-16 w-16 text-muted-foreground/30" />
              </div>
              <!-- Image count badge -->
              <span
                v-if="selectedItem.images && selectedItem.images.length > 1"
                class="absolute bottom-4 right-4 rounded-full bg-black/60 px-2.5 py-0.5 text-xs text-white backdrop-blur-sm"
              >
                {{ selectedItem.images.length }} images — scroll →
              </span>
            </div>

            <!-- Body -->
            <div class="flex-1 p-5 space-y-4">
              <!-- Title + status -->
              <div class="flex items-start gap-3">
                <div class="flex-1 space-y-1">
                  <h3 class="text-xl font-bold">{{ selectedItem.name }}</h3>
                  <p class="text-xs text-muted-foreground">Item #{{ selectedItem.id }}</p>
                </div>
                <span class="mt-1 rounded-full px-3 py-1 text-xs font-semibold"
                  :class="statusColor(selectedItem.status)">
                  {{ selectedItem.status }}
                </span>
              </div>

              <!-- Description -->
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-1">Description</p>
                <p class="text-sm leading-relaxed">
                  {{ selectedItem.description || 'No description provided.' }}
                </p>
              </div>

              <!-- Categories -->
              <div v-if="selectedItem.categories?.length">
                <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-2">Categories</p>
                <div class="flex flex-wrap gap-2">
                  <Badge v-for="cat in selectedItem.categories" :key="cat.id" variant="secondary">
                    {{ cat.name }}
                  </Badge>
                </div>
              </div>

              <!-- Type -->
              <div class="flex gap-6 text-sm">
                <div>
                  <p class="text-xs text-muted-foreground">Trade type</p>
                  <p class="font-medium capitalize">{{ selectedItem.type || 'barter' }}</p>
                </div>
                <div>
                  <p class="text-xs text-muted-foreground">Posted</p>
                  <p class="font-medium">{{ new Date(selectedItem.created_at).toLocaleDateString() }}</p>
                </div>
              </div>

              <!-- Own item notice -->
              <div v-if="isMyItem"
                class="rounded-xl border border-amber-300/50 bg-amber-500/10 p-3 text-sm text-amber-700 dark:text-amber-400">
                This is your item. <button class="underline font-medium" @click="goToDetail">Manage it here.</button>
              </div>

              <!-- Offer section (not own item, status available) -->
              <div v-else-if="canOffer" class="space-y-3">
                <div class="rounded-xl border bg-muted/20 p-3 space-y-2">
                  <p class="text-xs font-semibold text-muted-foreground">Add to cart to proceed with checkout</p>
                  <p class="text-xs text-muted-foreground">
                    You can select a trade-in item and set a cash amount during checkout.
                  </p>
                </div>

                <Button
                  v-if="isInSelectedCart"
                  class="w-full gap-2 bg-emerald-600 hover:bg-emerald-700"
                  @click="goToCart"
                >
                  View in Cart
                  <ArrowRight class="h-4 w-4" />
                </Button>
                <Button
                  v-else
                  class="w-full gap-2"
                  @click="addToCart(selectedItem!)"
                >
                  <ShoppingCart class="h-4 w-4" />
                  Add to Cart
                </Button>
              </div>

              <!-- Not logged in -->
              <div v-else-if="!userStore.profile && selectedItem?.status === 'available'"
                class="rounded-xl border bg-muted/40 p-3 text-sm text-muted-foreground text-center">
                <RouterLink to="/login" class="text-emerald-600 font-medium hover:underline">Log in</RouterLink> to send
                a trade offer.
              </div>

              <!-- Not available -->
              <div v-else-if="selectedItem?.status !== 'available' && !isMyItem"
                class="rounded-xl border bg-muted/40 p-3 text-sm text-muted-foreground text-center">
                This item is no longer available for trade.
              </div>
            </div>

            <!-- Footer -->
            <div class="border-t px-5 py-4 shrink-0">
              <Button variant="outline" class="w-full" @click="goToDetail">
                Open full detail page
              </Button>
            </div>
          </aside>
        </div>
      </Transition>
    </Teleport>

    <ImageLightbox
      v-model="lightboxOpen"
      :images="lightboxImages"
      :start-index="lightboxStartIdx"
    />
  </div>
</template>

<style scoped>
.panel-enter-active,
.panel-leave-active {
  transition: opacity 0.2s ease;
}

.panel-enter-active aside,
.panel-leave-active aside {
  transition: transform 0.25s ease;
}

.panel-enter-from {
  opacity: 0;
}

.panel-enter-from aside {
  transform: translateX(100%);
}

.panel-leave-to {
  opacity: 0;
}

.panel-leave-to aside {
  transform: translateX(100%);
}

/* Inventory accordion */
.inv-enter-active {
  transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1);
}

.inv-leave-active {
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
}

.inv-enter-from,
.inv-leave-to {
  opacity: 0;
  transform: translateY(-6px) scaleY(0.97);
  transform-origin: top;
}

/* Checkmark pop */
.check-enter-active {
  transition: transform 0.15s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.1s;
}

.check-leave-active {
  transition: all 0.1s ease;
}

.check-enter-from {
  transform: scale(0.4);
  opacity: 0;
}

.check-leave-to {
  opacity: 0;
}

/* Preview strip slide up */
.slide-up-enter-active {
  transition: all 0.2s ease;
}

.slide-up-leave-active {
  transition: all 0.15s ease;
}

.slide-up-enter-from {
  opacity: 0;
  transform: translateY(6px);
}

.slide-up-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
</style>
