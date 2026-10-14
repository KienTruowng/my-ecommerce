<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { toast } from 'vue-sonner'
import { Package, User, ArrowLeftRight, ImageOff, CheckCircle2, XCircle, Clock3 } from 'lucide-vue-next'

type ItemImage = { id: number; url: string }
type Item = {
  id: number
  name: string
  description: string
  status: string
  type?: string
  images?: ItemImage[]
}

const sentOffers = ref<any[]>([])
const receivedOffers = ref<any[]>([])
const myItems = ref<Item[]>([])
const selectedItemId = ref('')
const pendingCounts = ref<Record<string, number>>({})
const isLoadingSent = ref(false)
const isLoadingReceived = ref(false)

const receivedTitle = computed(() => {
  const item = myItems.value.find((entry) => String(entry.id) === selectedItemId.value)
  return item ? `Offers for "${item.name}"` : 'Select an item to load offers'
})

// Helpers
const getImageUrl = (images?: any[]) => {
  if (!images || images.length === 0) return ''
  const img = images[0]
  return img?.URL ?? img?.url ?? img?.Url ?? ''
}

const getItemName = (obj: any) => obj?.Name ?? obj?.name ?? '—'
const getUserName = (obj: any) => obj?.Name ?? obj?.name ?? 'Unknown'
const getOfferId = (offer: any) => offer?.ID ?? offer?.id
const getStatus = (offer: any) => (offer?.Status ?? offer?.status ?? 'pending') as string

const statusBadgeVariant = (status: string) => {
  if (status === 'accepted') return 'default'
  if (status === 'rejected' || status === 'cancelled') return 'destructive'
  return 'secondary'
}

const statusIcon = (status: string) => {
  if (status === 'accepted') return CheckCircle2
  if (status === 'rejected' || status === 'cancelled') return XCircle
  return Clock3
}

const fetchSentOffers = async () => {
  isLoadingSent.value = true
  try {
    const { data, error } = await api.GET('/api/users/me/offers')
    if (error) { toast.error('Unable to load sent offers'); return }
    sentOffers.value = data?.data ?? []
  } finally { isLoadingSent.value = false }
}

const fetchMyItems = async () => {
  const { data, error } = await api.GET('/api/users/me/items')
  if (error) { toast.error('Unable to load inventory'); return }
  myItems.value = (data as any) ?? []
  if (!selectedItemId.value && myItems.value.length) {
    selectedItemId.value = String(myItems.value[0].id)
  }
  fetchPendingCounts()
}

const fetchPendingCounts = async () => {
  try {
    const token = localStorage.getItem('access_token') ?? ''
    const res = await fetch(`/api/offers/pending-counts`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (res.ok) pendingCounts.value = await res.json()
  } catch { /* ignore */ }
}

const fetchReceivedOffers = async () => {
  const id = Number(selectedItemId.value)
  if (!id) return
  isLoadingReceived.value = true
  try {
    const { data, error } = await api.GET('/api/items/{id}/offers', { params: { path: { id } } })
    if (error) { toast.error('Unable to load received offers'); return }
    receivedOffers.value = (data as any)?.data ?? (data as any) ?? []
  } finally { isLoadingReceived.value = false }
}

const updateOfferStatus = async (offer: any, status: 'accepted' | 'rejected') => {
  const offerId = getOfferId(offer)
  if (!offerId) return
  const { error } = await api.PATCH('/api/offers/{id}/status', {
    params: { path: { id: offerId } },
    body: { status },
  })
  if (error) { toast.error('Unable to update'); return }
  if (status === 'accepted') toast.success('✅ Offer accepted! Trade confirmed.')
  else toast('❌ Offer rejected. Buyer will be notified.', { duration: 5000 })
  await Promise.all([fetchReceivedOffers(), fetchSentOffers()])
}

const cancelOffer = async (offer: any) => {
  const offerId = getOfferId(offer)
  if (!offerId) return
  const token = localStorage.getItem('access_token') ?? ''
  const res = await fetch(`/api/offers/${offerId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) { toast.error('Unable to cancel offer'); return }
  toast.success('Offer cancelled')
  await fetchSentOffers()
}

onMounted(async () => {
  await Promise.all([fetchSentOffers(), fetchMyItems()])
  await fetchReceivedOffers()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Banner -->
    <section class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-violet-50 via-white to-sky-50 p-6">
      <div class="pointer-events-none absolute -right-10 -top-8 h-32 w-32 rounded-full bg-violet-200/50" />
      <div class="pointer-events-none absolute -bottom-12 -left-8 h-40 w-40 rounded-full bg-sky-200/40" />
      <div class="relative">
        <p class="text-xs font-semibold uppercase tracking-widest text-violet-600">Offers</p>
        <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">Trade Offers</h1>
        <p class="mt-2 max-w-xl text-sm text-slate-600">
          Review incoming offers on your items and track the offers you've sent.
        </p>
      </div>
    </section>

    <Tabs default-value="received" class="space-y-4">
      <TabsList>
        <TabsTrigger value="received" class="gap-1.5">
          <Package class="h-3.5 w-3.5" /> Received
          <Badge variant="secondary" class="ml-1 h-5 px-1.5 text-[10px]">{{ receivedOffers.length }}</Badge>
        </TabsTrigger>
        <TabsTrigger value="sent" class="gap-1.5">
          <ArrowLeftRight class="h-3.5 w-3.5" /> Sent
          <Badge variant="secondary" class="ml-1 h-5 px-1.5 text-[10px]">{{ sentOffers.length }}</Badge>
        </TabsTrigger>
      </TabsList>

      <!-- RECEIVED -->
      <TabsContent value="received">
        <!-- Item grid to select which item's offers to view -->
        <Card class="bg-card/80 mb-4">
          <CardContent class="pt-4 space-y-3">
            <p class="text-xs font-medium text-muted-foreground">Select an item to view its offers</p>
            <!-- Item image grid -->
            <div v-if="myItems.length === 0" class="text-center py-4">
              <Package class="mx-auto h-8 w-8 text-muted-foreground/30" />
              <p class="text-xs text-muted-foreground mt-1">No items in your inventory yet.</p>
            </div>
            <div v-else class="flex gap-2 overflow-x-auto pb-1 scrollbar-thin">
              <button
                v-for="item in myItems"
                :key="item.id"
                class="flex flex-col items-center gap-1 shrink-0 rounded-xl border-2 p-2 transition-all w-24 relative"
                :class="selectedItemId === String(item.id)
                  ? 'border-emerald-400 bg-emerald-50 ring-1 ring-emerald-200 dark:bg-emerald-950'
                  : 'border-border hover:border-emerald-200 hover:bg-muted/50'"
                @click="selectedItemId = String(item.id); fetchReceivedOffers()"
              >
                <div class="h-14 w-14 rounded-lg border bg-background overflow-hidden">
                  <img v-if="getImageUrl(item.images)" :src="getImageUrl(item.images)"
                    class="h-full w-full object-cover" :alt="item.name" />
                  <div v-else class="flex h-full items-center justify-center">
                    <Package class="h-4 w-4 text-muted-foreground/30" />
                  </div>
                </div>
                <p class="text-[10px] font-medium text-center line-clamp-2 leading-tight">{{ item.name }}</p>
                <span
                  v-if="pendingCounts[String(item.id)]"
                  class="absolute -top-1.5 -right-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold text-white shadow"
                >{{ pendingCounts[String(item.id)] }}</span>
              </button>
            </div>
          </CardContent>
        </Card>

        <div v-if="isLoadingReceived" class="text-sm text-muted-foreground">Loading...</div>
        <div v-else class="space-y-3">
          <!-- Empty -->
          <Card v-if="!receivedOffers.length" class="bg-card/80 py-10">
            <CardContent class="text-center">
              <Package class="mx-auto h-10 w-10 text-muted-foreground/30" />
              <p class="mt-2 font-semibold text-muted-foreground">No offers received</p>
              <p class="text-xs text-muted-foreground mt-1">Share your item ID to start receiving trade proposals.</p>
            </CardContent>
          </Card>

          <!-- Offer cards -->
          <Card v-for="offer in receivedOffers" :key="getOfferId(offer)"
            class="bg-card/80 overflow-hidden transition-all hover:shadow-md"
            :class="getStatus(offer) === 'accepted' ? 'border-emerald-300/60' : (getStatus(offer) === 'rejected' || getStatus(offer) === 'cancelled') ? 'border-red-200/40 opacity-70' : ''">
            <div class="flex flex-col sm:flex-row">
              <!-- Thumbnail area -->
              <div class="flex flex-row sm:flex-col gap-2 p-4 sm:w-40 shrink-0 bg-muted/30 sm:border-r">
                <!-- Target item image -->
                <div class="h-16 w-16 sm:h-20 sm:w-20 shrink-0 rounded-lg border bg-background overflow-hidden">
                  <img v-if="getImageUrl(offer.OfferItem?.ItemImage ?? offer.offer_item?.ItemImage ?? offer.offer_item?.images)"
                    :src="getImageUrl(offer.OfferItem?.ItemImage ?? offer.offer_item?.ItemImage ?? offer.offer_item?.images)"
                    class="h-full w-full object-cover" alt="Offer item" />
                  <div v-else class="flex h-full items-center justify-center">
                    <ImageOff class="h-5 w-5 text-muted-foreground/30" />
                  </div>
                </div>
                <div class="flex-1 sm:text-center min-w-0">
                  <p class="text-[10px] text-muted-foreground uppercase tracking-wider">From</p>
                  <p class="text-xs font-semibold truncate">{{ getUserName(offer.User ?? offer.user) }}</p>
                  <p class="text-[10px] text-muted-foreground mt-1">Trading</p>
                  <p class="text-xs font-semibold line-clamp-2">
                    {{ getItemName(offer.OfferItem ?? offer.offer_item) || 'Cash offer' }}
                  </p>
                </div>
              </div>

              <!-- Info area -->
              <div class="flex-1 p-4 space-y-3">
                <div class="flex items-center justify-between gap-2 flex-wrap">
                  <Badge :variant="statusBadgeVariant(getStatus(offer))" class="gap-1">
                    <component :is="statusIcon(getStatus(offer))" class="h-3 w-3" />
                    {{ getStatus(offer) }}
                  </Badge>
                  <span v-if="offer.CashAmount ?? offer.cash_amount" class="text-sm font-semibold text-amber-600">
                    {{ new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(offer.CashAmount ?? offer.cash_amount ?? 0) }}
                  </span>
                </div>
                <p class="text-sm text-muted-foreground line-clamp-2">
                  "{{ offer.OfferMessage ?? offer.offer_message ?? 'No message' }}"
                </p>

                <!-- Actions (only for pending) -->
                <div v-if="getStatus(offer) === 'pending'" class="flex gap-2 pt-1">
                  <Button size="sm" variant="default" class="gap-1.5 bg-emerald-600 hover:bg-emerald-700" @click="updateOfferStatus(offer, 'accepted')">
                    <CheckCircle2 class="h-3.5 w-3.5" /> Accept
                  </Button>
                  <Button size="sm" variant="destructive" class="gap-1.5" @click="updateOfferStatus(offer, 'rejected')">
                    <XCircle class="h-3.5 w-3.5" /> Reject
                  </Button>
                </div>
                <!-- Already resolved -->
                <div v-else class="pt-1">
                  <p v-if="getStatus(offer) === 'accepted'" class="text-xs text-emerald-600 font-medium">✅ Trade confirmed — schedule a meetup in Appointments.</p>
                  <p v-else class="text-xs text-muted-foreground">This offer has been rejected.</p>
                </div>
              </div>
            </div>
          </Card>
        </div>
      </TabsContent>

      <!-- SENT -->
      <TabsContent value="sent">
        <div v-if="isLoadingSent" class="text-sm text-muted-foreground">Loading...</div>
        <div v-else class="space-y-3">
          <Card v-if="!sentOffers.length" class="bg-card/80 py-10">
            <CardContent class="text-center">
              <ArrowLeftRight class="mx-auto h-10 w-10 text-muted-foreground/30" />
              <p class="mt-2 font-semibold text-muted-foreground">No offers sent yet</p>
              <p class="text-xs text-muted-foreground mt-1">Browse the marketplace and send offers to items you want.</p>
            </CardContent>
          </Card>

          <Card v-for="offer in sentOffers" :key="getOfferId(offer)"
            class="bg-card/80 overflow-hidden transition-all hover:shadow-md"
            :class="getStatus(offer) === 'accepted' ? 'border-emerald-300/60' : (getStatus(offer) === 'rejected' || getStatus(offer) === 'cancelled') ? 'border-red-200/40 opacity-70' : ''">
            <div class="flex flex-col sm:flex-row">
              <!-- Thumbnail area -->
              <div class="flex flex-row sm:flex-col gap-2 p-4 sm:w-40 shrink-0 bg-muted/30 sm:border-r">
                <div class="h-16 w-16 sm:h-20 sm:w-20 shrink-0 rounded-lg border bg-background overflow-hidden">
                  <img v-if="getImageUrl(offer.TargetItem?.ItemImage ?? offer.target_item?.ItemImage ?? offer.target_item?.images)"
                    :src="getImageUrl(offer.TargetItem?.ItemImage ?? offer.target_item?.ItemImage ?? offer.target_item?.images)"
                    class="h-full w-full object-cover" alt="Target item" />
                  <div v-else class="flex h-full items-center justify-center">
                    <ImageOff class="h-5 w-5 text-muted-foreground/30" />
                  </div>
                </div>
                <div class="flex-1 sm:text-center min-w-0">
                  <p class="text-[10px] text-muted-foreground uppercase tracking-wider">You want</p>
                  <p class="text-xs font-semibold line-clamp-2">
                    {{ getItemName(offer.TargetItem ?? offer.target_item) }}
                  </p>
                  <p class="text-[10px] text-muted-foreground mt-1">You offered</p>
                  <p class="text-xs font-semibold line-clamp-2">
                    {{ getItemName(offer.OfferItem ?? offer.offer_item) || 'Cash' }}
                  </p>
                </div>
              </div>

              <!-- Info area -->
              <div class="flex-1 p-4 space-y-3">
                <div class="flex items-center justify-between gap-2 flex-wrap">
                  <Badge :variant="statusBadgeVariant(getStatus(offer))" class="gap-1">
                    <component :is="statusIcon(getStatus(offer))" class="h-3 w-3" />
                    {{ getStatus(offer) }}
                  </Badge>
                  <span v-if="offer.CashAmount ?? offer.cash_amount" class="text-sm font-semibold text-amber-600">
                    {{ new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(offer.CashAmount ?? offer.cash_amount ?? 0) }}
                  </span>
                </div>
                <p class="text-sm text-muted-foreground line-clamp-2">
                  "{{ offer.OfferMessage ?? offer.offer_message ?? 'No message' }}"
                </p>
                <p v-if="getStatus(offer) === 'accepted'" class="text-xs text-emerald-600 font-medium">✅ Accepted! Check Appointments.</p>
                <p v-else-if="getStatus(offer) === 'cancelled'" class="text-xs text-muted-foreground">❌ Cancelled by you.</p>
                <p v-else-if="getStatus(offer) === 'rejected'" class="text-xs text-muted-foreground">This offer was rejected by the seller.</p>
                <div v-else class="flex items-center gap-3">
                  <p class="text-xs text-muted-foreground">⏳ Waiting for seller's response...</p>
                  <button
                    class="text-xs text-muted-foreground hover:text-destructive transition-colors underline underline-offset-2"
                    @click="cancelOffer(offer)"
                  >Cancel</button>
                </div>
              </div>
            </div>
          </Card>
        </div>
      </TabsContent>
    </Tabs>
  </div>
</template>
