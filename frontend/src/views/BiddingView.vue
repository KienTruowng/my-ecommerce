<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useUserStore } from '@/stores/user'
import ImageLightbox from '@/components/ImageLightbox.vue'
import {
  Gavel,
  Clock,
  TrendingUp,
  Trophy,
  ArrowLeft,
  Users,
  ChevronRight,
  AlertCircle,
  CheckCircle2,
  Loader2,
  ImageOff,
  ShieldCheck,
} from 'lucide-vue-next'

// ─── Types ────────────────────────────────────────────────────────────────────
type Bidder = { id: number; Name: string; avatarUrl?: string }
type Bid = {
  ID: number
  amount: number
  rank: number
  status: 'active' | 'won' | 'lost' | 'forfeited'
  deposit_paid: boolean
  Bidder: Bidder
  CreatedAt: string
}
type AuctionConfig = {
  ID: number
  item_id: number
  base_price: number
  step_price: number
  deposit_amount: number
  deadline: string
  status: 'open' | 'settled' | 'cancelled'
  Bids: Bid[]
  Item?: {
    id: number
    name: string
    description: string
    status: string
    images?: { id: number; url: string }[]
  }
}

// ─── State ───────────────────────────────────────────────────────────────────
const route      = useRoute()
const router     = useRouter()
const userStore  = useUserStore()

const itemId     = computed(() => Number(route.params.id))
const auction    = ref<AuctionConfig | null>(null)
const isLoading  = ref(false)
const isBidding  = ref(false)
const bidAmount  = ref(0)

// For seller configuration panel
const showConfig    = ref(false)
const configForm    = ref({
  base_price: 0,
  step_price: 0,
  deposit_amount: 0,
  deadline: '',
})
const isConfiguring = ref(false)

// Countdown
const now        = ref(Date.now())
let   timer: ReturnType<typeof setInterval> | null = null

// ─── Computed ─────────────────────────────────────────────────────────────────
const deadline = computed(() =>
  auction.value ? new Date(auction.value.deadline).getTime() : 0,
)

const timeLeft = computed(() => {
  const diff = deadline.value - now.value
  if (diff <= 0) return null
  const h = Math.floor(diff / 3_600_000)
  const m = Math.floor((diff % 3_600_000) / 60_000)
  const s = Math.floor((diff % 60_000) / 1_000)
  return { h, m, s, total: diff }
})

const isExpired = computed(() => !timeLeft.value && !!auction.value)

const sortedBids = computed(() =>
  [...(auction.value?.Bids ?? [])].sort((a, b) => a.rank - b.rank),
)

const highestBid = computed(() => sortedBids.value[0] ?? null)

const currentPrice = computed(() => {
  if (!auction.value) return 0
  return highestBid.value ? highestBid.value.amount : auction.value.base_price
})

const minimumBid = computed(() => {
  if (!auction.value) return 0
  return highestBid.value
    ? highestBid.value.amount + auction.value.step_price
    : auction.value.base_price
})

const isMyAuction = computed(() => {
  if (!auction.value?.Item || !userStore.profile) return false
  return (auction.value.Item as any).UserID === userStore.profile.id ||
         (auction.value.Item as any).user_id === userStore.profile.id
})

const canBid = computed(() =>
  !!auction.value &&
  auction.value.status === 'open' &&
  !isExpired.value &&
  !isMyAuction.value &&
  !!userStore.profile,
)

// Lightbox
const lightboxOpen = ref(false)
const lightboxImages = ref<{ src: string; alt: string }[]>([])
const lightboxStartIdx = ref(0)

const openLightbox = (idx: number) => {
  const images = auction.value?.Item?.images ?? []
  lightboxImages.value = images.map((img) => ({
    src: img.url,
    alt: auction.value?.Item?.name ?? 'Auction image',
  }))
  lightboxStartIdx.value = idx
  lightboxOpen.value = true
}

// ─── API helpers ──────────────────────────────────────────────────────────────
const BASE = ''

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('access_token') ?? ''
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      Authorization: token ? `Bearer ${token}` : '',
      ...(options?.headers ?? {}),
    },
  })
  const json = await res.json()
  if (!res.ok) throw new Error(json?.error ?? `HTTP ${res.status}`)
  return json
}

// ─── Data fetching ────────────────────────────────────────────────────────────
const fetchAuction = async () => {
  isLoading.value = true
  try {
    const json = await apiFetch<{ data: AuctionConfig }>(`/api/items/${itemId.value}/auction`)
    auction.value = json.data
    // Pre-fill bid amount
    bidAmount.value = minimumBid.value
  } catch (e: any) {
    if (e.message?.includes('404') || e.message?.includes('không tồn tại')) {
      auction.value = null
    } else {
      toast.error('Không thể tải thông tin đấu giá')
    }
  } finally {
    isLoading.value = false
  }
}

// ─── Actions ──────────────────────────────────────────────────────────────────
const placeBid = async () => {
  if (!canBid.value) return
  if (bidAmount.value < minimumBid.value) {
    toast.error(`Giá đặt tối thiểu là ${fmtVND(minimumBid.value)}`)
    return
  }
  isBidding.value = true
  try {
    await apiFetch(`/api/items/${itemId.value}/bids`, {
      method: 'POST',
      body: JSON.stringify({ amount: bidAmount.value }),
    })
    toast.success('🎉 Đặt giá thành công!')
    await fetchAuction()
    bidAmount.value = minimumBid.value
  } catch (e: any) {
    toast.error(e.message ?? 'Đặt giá thất bại')
  } finally {
    isBidding.value = false
  }
}

const configureAuction = async () => {
  isConfiguring.value = true
  try {
    const payload = {
      ...configForm.value,
      deadline: new Date(configForm.value.deadline).toISOString(),
    }
    await apiFetch(`/api/items/${itemId.value}/auction`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    toast.success('Cấu hình đấu giá thành công!')
    showConfig.value = false
    await fetchAuction()
  } catch (e: any) {
    toast.error(e.message ?? 'Cấu hình thất bại')
  } finally {
    isConfiguring.value = false
  }
}

// ─── Formatting ───────────────────────────────────────────────────────────────
const fmtVND = (v: number) =>
  new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(v)

const fmtTime = (n: number) => String(n).padStart(2, '0')

const rankBadge = (rank: number) => {
  if (rank === 1) return '🥇'
  if (rank === 2) return '🥈'
  if (rank === 3) return '🥉'
  return `#${rank}`
}

const statusColor = (status: string) => {
  const map: Record<string, string> = {
    open: 'bg-emerald-500/15 text-emerald-600 border-emerald-300/40',
    settled: 'bg-sky-500/15 text-sky-600 border-sky-300/40',
    cancelled: 'bg-zinc-500/15 text-zinc-500 border-zinc-300/40',
  }
  return map[status] ?? 'bg-muted text-muted-foreground'
}

// ─── Lifecycle ────────────────────────────────────────────────────────────────
onMounted(async () => {
  await userStore.fetchProfile().catch(() => {})
  await fetchAuction()
  timer = setInterval(() => { now.value = Date.now() }, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

watch(itemId, fetchAuction)
</script>

<template>
  <div class="space-y-6 pb-12">
    <!-- ── Back ── -->
    <button
      @click="router.back()"
      class="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
    >
      <ArrowLeft class="h-4 w-4" />
      Back
    </button>

    <!-- ── Loading ── -->
    <div v-if="isLoading" class="flex items-center justify-center py-24">
      <Loader2 class="h-10 w-10 animate-spin text-muted-foreground/50" />
    </div>

    <!-- ── No auction configured yet ── -->
    <div
      v-else-if="!auction"
      class="flex flex-col items-center gap-4 py-20 text-center text-muted-foreground"
    >
      <Gavel class="h-14 w-14 opacity-20" />
      <div class="space-y-1">
        <p class="text-base font-medium text-foreground">Chưa có phiên đấu giá</p>
        <p class="text-sm">Sản phẩm này chưa được cấu hình đấu giá.</p>
      </div>

      <!-- Seller config CTA -->
      <div v-if="userStore.profile">
        <Button @click="showConfig = !showConfig" variant="outline" class="gap-2">
          <Gavel class="h-4 w-4" />
          Tạo phiên đấu giá cho sản phẩm này
        </Button>
      </div>

      <!-- Config form for seller -->
      <Transition name="fade-slide">
        <form
          v-if="showConfig"
          @submit.prevent="configureAuction"
          class="mt-4 w-full max-w-sm rounded-2xl border bg-card p-5 text-left space-y-4 shadow-lg"
        >
          <h3 class="font-semibold text-base">Cấu hình đấu giá</h3>
          <div class="space-y-1">
            <label class="text-xs font-medium text-muted-foreground">Giá khởi điểm (VND)</label>
            <Input v-model.number="configForm.base_price" type="number" min="0" placeholder="500000" />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-muted-foreground">Bước giá (≤ 20% giá khởi điểm)</label>
            <Input v-model.number="configForm.step_price" type="number" min="0" placeholder="50000" />
            <p class="text-[11px] text-muted-foreground">
              Tối đa: {{ fmtVND(Math.floor(configForm.base_price * 0.2)) }}
            </p>
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-muted-foreground">Tiền cọc (VND)</label>
            <Input v-model.number="configForm.deposit_amount" type="number" min="0" placeholder="100000" />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-medium text-muted-foreground">Thời hạn đấu giá</label>
            <Input v-model="configForm.deadline" type="datetime-local" />
          </div>
          <Button type="submit" class="w-full gap-2" :disabled="isConfiguring">
            <Loader2 v-if="isConfiguring" class="h-4 w-4 animate-spin" />
            <Gavel v-else class="h-4 w-4" />
            {{ isConfiguring ? 'Đang lưu…' : 'Bắt đầu đấu giá' }}
          </Button>
        </form>
      </Transition>
    </div>

    <!-- ── Main auction layout ── -->
    <template v-else>
      <!-- Hero band -->
      <section
        class="relative overflow-hidden rounded-3xl border p-6 md:p-8"
        :class="
          auction.status === 'open' && !isExpired
            ? 'bg-gradient-to-br from-amber-500/20 via-orange-400/10 to-transparent'
            : 'bg-gradient-to-br from-zinc-500/10 via-zinc-400/5 to-transparent'
        "
      >
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <!-- Left: item info -->
          <div class="flex gap-4">
            <!-- Image gallery -->
            <div v-if="auction.Item?.images?.length" class="flex gap-2 overflow-x-auto pb-1 scrollbar-thin max-w-[320px]">
              <img
                v-for="(img, idx) in auction.Item.images"
                :key="img.id"
                :src="img.url"
                :alt="`${auction.Item.name} - ${idx + 1}`"
                class="h-20 w-20 shrink-0 rounded-xl border object-cover shadow-sm cursor-pointer transition-transform hover:scale-105"
                @click="openLightbox(idx)"
              />
            </div>
            <div v-else class="flex h-20 w-20 shrink-0 items-center justify-center rounded-xl border bg-muted">
              <ImageOff class="h-7 w-7 text-muted-foreground/30" />
            </div>

            <div class="space-y-1">
              <div class="flex items-center gap-2 flex-wrap">
                <h1 class="text-xl font-bold">{{ auction.Item?.name ?? `Sản phẩm #${auction.item_id}` }}</h1>
                <span
                  class="rounded-full border px-2.5 py-0.5 text-[11px] font-semibold"
                  :class="statusColor(auction.status)"
                >
                  {{ auction.status === 'open' ? '🔨 Đang đấu giá' : auction.status === 'settled' ? '✅ Đã kết thúc' : '❌ Đã huỷ' }}
                </span>
              </div>
              <p class="text-sm text-muted-foreground line-clamp-2">
                {{ auction.Item?.description ?? '' }}
              </p>
              <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
                <Users class="h-3.5 w-3.5" />
                <span>{{ auction.Bids?.length ?? 0 }} lượt đặt giá</span>
              </div>
            </div>
          </div>

          <!-- Right: countdown -->
          <div
            class="shrink-0 rounded-2xl border bg-background/70 px-5 py-4 text-center backdrop-blur-sm"
            :class="isExpired ? 'border-zinc-300/40' : 'border-amber-300/40'"
          >
            <div v-if="timeLeft && !isExpired">
              <p class="text-[11px] font-semibold uppercase tracking-widest text-amber-600 mb-2">
                <Clock class="inline h-3 w-3 mr-1" />Thời gian còn lại
              </p>
              <div class="flex items-center gap-1 text-3xl font-mono font-bold tabular-nums">
                <span>{{ fmtTime(timeLeft.h) }}</span>
                <span class="text-amber-400 animate-pulse">:</span>
                <span>{{ fmtTime(timeLeft.m) }}</span>
                <span class="text-amber-400 animate-pulse">:</span>
                <span>{{ fmtTime(timeLeft.s) }}</span>
              </div>
              <p class="text-[10px] text-muted-foreground mt-1.5">
                Kết thúc: {{ new Date(auction.deadline).toLocaleString('vi-VN') }}
              </p>
            </div>
            <div v-else class="text-sm font-medium text-muted-foreground">
              <Clock class="inline h-4 w-4 mr-1" />
              {{ auction.status === 'settled' ? 'Đã kết thúc' : 'Hết hạn' }}
            </div>
          </div>
        </div>

        <!-- Decorative blurs -->
        <div class="absolute -right-10 -top-10 h-36 w-36 rounded-full bg-amber-400/20 blur-3xl pointer-events-none" />
        <div class="absolute right-24 bottom-0 h-24 w-24 rounded-full bg-orange-400/15 blur-2xl pointer-events-none" />
      </section>

      <div class="grid gap-6 lg:grid-cols-[1fr_360px]">
        <!-- ── Bids list ── -->
        <section class="space-y-3">
          <div class="flex items-center justify-between">
            <h2 class="text-base font-semibold flex items-center gap-2">
              <TrendingUp class="h-4 w-4 text-amber-500" />
              Lịch sử đặt giá
            </h2>
            <Badge variant="outline" class="text-xs">{{ sortedBids.length }} lượt</Badge>
          </div>

          <!-- Empty bids -->
          <div
            v-if="sortedBids.length === 0"
            class="flex flex-col items-center gap-3 rounded-2xl border border-dashed py-12 text-muted-foreground"
          >
            <Gavel class="h-10 w-10 opacity-20" />
            <p class="text-sm">Chưa có ai đặt giá. Hãy là người đầu tiên!</p>
          </div>

          <!-- Bid rows -->
          <div v-else class="rounded-2xl border bg-card/80 overflow-hidden divide-y">
            <div
              v-for="bid in sortedBids"
              :key="bid.ID"
              class="flex items-center gap-4 px-4 py-3 transition-colors"
              :class="bid.rank === 1 ? 'bg-amber-500/5' : 'hover:bg-muted/40'"
            >
              <!-- Rank -->
              <span class="text-lg w-8 text-center shrink-0">{{ rankBadge(bid.rank) }}</span>

              <!-- Avatar placeholder -->
              <div class="h-8 w-8 rounded-full bg-gradient-to-br from-emerald-400 to-cyan-500 shrink-0 flex items-center justify-center text-[11px] font-bold text-white">
                {{ (bid.Bidder?.Name ?? '?').charAt(0).toUpperCase() }}
              </div>

              <!-- Name + time -->
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate">{{ bid.Bidder?.Name ?? 'Ẩn danh' }}</p>
                <p class="text-[11px] text-muted-foreground">
                  {{ new Date(bid.CreatedAt).toLocaleString('vi-VN') }}
                </p>
              </div>

              <!-- Amount -->
              <div class="text-right shrink-0">
                <p class="font-bold text-base" :class="bid.rank === 1 ? 'text-amber-600' : ''">
                  {{ fmtVND(bid.amount) }}
                </p>
                <div class="flex items-center justify-end gap-1">
                  <ShieldCheck
                    v-if="bid.deposit_paid"
                    class="h-3 w-3 text-emerald-500"
                    title="Đã đặt cọc"
                  />
                  <span
                    class="text-[10px] font-medium"
                    :class="{
                      'text-emerald-600': bid.status === 'won',
                      'text-muted-foreground': bid.status === 'active' || bid.status === 'lost',
                      'text-red-500': bid.status === 'forfeited',
                    }"
                  >
                    {{ bid.status === 'won' ? '🏆 Thắng' : bid.status === 'forfeited' ? 'Bỏ cuộc' : bid.status === 'active' ? '' : 'Thua' }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- ── Right panel ── -->
        <aside class="space-y-4">
          <!-- Price summary card -->
          <div class="rounded-2xl border bg-card/80 p-5 space-y-4">
            <h3 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">Thông tin giá</h3>

            <div class="space-y-2.5">
              <div class="flex justify-between text-sm">
                <span class="text-muted-foreground">Giá khởi điểm</span>
                <span class="font-medium">{{ fmtVND(auction.base_price) }}</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-muted-foreground">Bước giá tối thiểu</span>
                <span class="font-medium">{{ fmtVND(auction.step_price) }}</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-muted-foreground">Tiền cọc yêu cầu</span>
                <span class="font-medium text-amber-600">{{ fmtVND(auction.deposit_amount) }}</span>
              </div>
              <div class="h-px bg-border" />
              <div class="flex justify-between">
                <span class="text-sm font-semibold">Giá hiện tại</span>
                <span class="text-lg font-bold text-amber-600">{{ fmtVND(currentPrice) }}</span>
              </div>
              <div v-if="highestBid" class="flex justify-between text-xs text-muted-foreground">
                <span>Đặt bởi</span>
                <span class="font-medium text-foreground">{{ highestBid.Bidder?.Name ?? 'Ẩn danh' }}</span>
              </div>
            </div>
          </div>

          <!-- Bid input card -->
          <div v-if="canBid" class="rounded-2xl border bg-card/80 p-5 space-y-4">
            <div>
              <h3 class="text-sm font-semibold">Đặt giá của bạn</h3>
              <p class="text-xs text-muted-foreground mt-0.5">
                Giá đặt phải <strong>≥ giá khởi điểm</strong> và <strong>≥ giá cao nhất + bước giá</strong>
              </p>
            </div>

            <!-- Bid rules summary -->
            <div class="grid gap-2 rounded-xl bg-muted/50 p-3 text-xs">
              <div class="flex justify-between">
                <span class="text-muted-foreground">Giá khởi điểm:</span>
                <span class="font-semibold">{{ fmtVND(auction.base_price) }}</span>
              </div>
              <div v-if="highestBid" class="flex justify-between">
                <span class="text-muted-foreground">Giá cao nhất hiện tại:</span>
                <span class="font-semibold text-amber-600">{{ fmtVND(highestBid.amount) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">Bước giá:</span>
                <span class="font-semibold">{{ fmtVND(auction.step_price) }}</span>
              </div>
              <div class="h-px bg-border" />
              <div class="flex justify-between">
                <span class="font-semibold text-foreground">Giá đặt tối thiểu của bạn:</span>
                <span class="text-base font-bold text-amber-600">{{ fmtVND(minimumBid) }}</span>
              </div>
            </div>

            <!-- Live validation -->
            <div
              v-if="bidAmount > 0 && bidAmount < minimumBid"
              class="flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-400"
            >
              <AlertCircle class="h-3.5 w-3.5 shrink-0" />
              <span>
                Giá đặt quá thấp! Mức tối thiểu là <strong>{{ fmtVND(minimumBid) }}</strong>
                (thiếu {{ fmtVND(minimumBid - bidAmount) }})
              </span>
            </div>
            <div
              v-else-if="bidAmount >= minimumBid"
              class="flex items-center gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-400"
            >
              <CheckCircle2 class="h-3.5 w-3.5 shrink-0" />
              <span>Hợp lệ — giá của bạn cao hơn mức tối thiểu {{ fmtVND(bidAmount - minimumBid) }}</span>
            </div>

            <!-- Deposit notice -->
            <div
              class="flex items-start gap-2 rounded-xl bg-amber-500/10 border border-amber-300/40 p-3 text-xs text-amber-700 dark:text-amber-400"
            >
              <ShieldCheck class="h-4 w-4 shrink-0 mt-0.5" />
              <div>
                Đặt giá yêu cầu đặt cọc <strong>{{ fmtVND(auction.deposit_amount) }}</strong>.
                Nếu bạn thắng mà không hoàn tất giao dịch, tiền cọc sẽ chuyển cho người đứng thứ hai.
              </div>
            </div>

            <div class="space-y-2">
              <label class="text-xs font-medium text-muted-foreground">Mức giá của bạn (VND)</label>
              <Input
                v-model.number="bidAmount"
                type="number"
                :min="minimumBid"
                :step="auction.step_price"
                class="font-mono text-lg"
                :class="bidAmount > 0 && bidAmount < minimumBid ? 'border-red-300 focus-visible:ring-red-300' : ''"
              />
            </div>

            <!-- Quick bid buttons -->
            <div>
              <p class="mb-2 text-[10px] font-medium text-muted-foreground uppercase tracking-wider">Đặt nhanh</p>
              <div class="flex flex-wrap gap-2">
                <button
                  @click="bidAmount = minimumBid"
                  class="rounded-full border-2 border-amber-300 bg-amber-50 px-3 py-1.5 text-xs font-semibold text-amber-700 transition-all hover:bg-amber-100 hover:border-amber-400 hover:scale-105 dark:bg-amber-950 dark:text-amber-300 dark:border-amber-700"
                >
                  🔥 Tối thiểu
                  <span class="ml-1 font-mono">{{ fmtVND(minimumBid) }}</span>
                </button>
                <button
                  v-for="step in [1, 2, 3]"
                  :key="step"
                  @click="bidAmount = minimumBid + auction.step_price * step"
                  class="rounded-full border px-3 py-1.5 text-xs font-medium transition-all hover:bg-muted hover:scale-105"
                >
                  +{{ step }} bước
                  <span class="ml-0.5 font-mono text-[10px] text-muted-foreground">
                    {{ fmtVND(minimumBid + auction.step_price * step) }}
                  </span>
                </button>
              </div>
            </div>

            <Button
              class="w-full gap-2 bg-amber-500 hover:bg-amber-600 text-white font-semibold"
              size="lg"
              @click="placeBid"
              :disabled="isBidding || bidAmount < minimumBid"
            >
              <Loader2 v-if="isBidding" class="h-5 w-5 animate-spin" />
              <Gavel v-else class="h-5 w-5" />
              {{ isBidding ? 'Đang đặt giá…' : `Xác nhận đặt ${fmtVND(bidAmount || minimumBid)}` }}
            </Button>
          </div>

          <!-- Auction ended state -->
          <div
            v-else-if="isExpired || auction.status === 'settled'"
            class="rounded-2xl border bg-card/80 p-5 text-center space-y-2"
          >
            <Trophy class="mx-auto h-10 w-10 text-amber-400" />
            <p class="font-semibold text-base">
              {{ auction.status === 'settled' ? 'Phiên đấu giá đã kết thúc' : 'Hết thời hạn đặt giá' }}
            </p>
            <p v-if="highestBid" class="text-sm text-muted-foreground">
              Người thắng: <strong>{{ highestBid.Bidder?.Name }}</strong>
            </p>
            <p v-else class="text-sm text-muted-foreground">Không có ai đặt giá.</p>
          </div>

          <!-- My auction notice -->
          <div
            v-else-if="isMyAuction"
            class="rounded-2xl border border-amber-300/50 bg-amber-500/10 p-4 text-sm text-amber-700 dark:text-amber-400"
          >
            <AlertCircle class="inline h-4 w-4 mr-1.5" />
            Đây là sản phẩm của bạn. Bạn không thể tự đặt giá.
          </div>

          <!-- Not logged in -->
          <div
            v-else-if="!userStore.profile"
            class="rounded-2xl border bg-muted/40 p-4 text-sm text-muted-foreground text-center"
          >
            <RouterLink to="/login" class="text-amber-600 font-medium hover:underline">Đăng nhập</RouterLink>
            để tham gia đấu giá.
          </div>

          <!-- View item detail link -->
          <RouterLink
            :to="`/items/${itemId}`"
            class="flex items-center justify-between rounded-xl border bg-card/80 px-4 py-3 text-sm font-medium hover:bg-muted/60 transition-colors"
          >
            <span>Xem chi tiết sản phẩm</span>
            <ChevronRight class="h-4 w-4 text-muted-foreground" />
          </RouterLink>
        </aside>
      </div>
    </template>

    <ImageLightbox
      v-model="lightboxOpen"
      :images="lightboxImages"
      :start-index="lightboxStartIdx"
    />
  </div>
</template>

<style scoped>
.fade-slide-enter-active { transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1); }
.fade-slide-leave-active { transition: all 0.18s ease; }
.fade-slide-enter-from  { opacity: 0; transform: translateY(-8px); }
.fade-slide-leave-to    { opacity: 0; transform: translateY(-4px); }
</style>
