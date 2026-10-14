<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useRouter } from 'vue-router'
import { Compass, Sparkles, Star, Tag, Package, ImageOff } from 'lucide-vue-next'

type Item = {
  id: number; name: string; description: string; status: string; type: string
  images?: { id: number; url: string }[]; categories?: { id: number; name: string }[]
}

const router = useRouter()
const items = ref<Item[]>([])
const isLoading = ref(true)
const search = ref('')
const activeCat = ref('all')

const categories = computed(() => {
  const cats = new Set<string>()
  items.value.forEach(i => i.categories?.forEach(c => cats.add(c.name)))
  return ['all', ...Array.from(cats)]
})

const filteredItems = computed(() => {
  const q = search.value.trim().toLowerCase()
  return items.value.filter(i => {
    if (activeCat.value !== 'all' && !i.categories?.some(c => c.name === activeCat.value)) return false
    if (q && !i.name.toLowerCase().includes(q) && !i.description?.toLowerCase().includes(q)) return false
    return true
  })
})

const getImageUrl = (images?: { id: number; url: string }[]) => images?.[0]?.url ?? ''

const fetchItems = async () => {
  isLoading.value = true
  try {
    const res = await fetch('/api/items?limit=50')
    if (res.ok) {
      const data = await res.json()
      items.value = data?.data ?? []
    }
  } catch { /* ignore */ }
  finally { isLoading.value = false }
}

const goToItem = (id: number) => router.push(`/items/${id}`)

onMounted(fetchItems)
</script>

<template>
  <div class="space-y-6">
    <section
      class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-emerald-50 via-white to-orange-50 p-6"
    >
      <div class="pointer-events-none absolute -right-10 -top-6 h-32 w-32 rounded-full bg-orange-200/60" />
      <div class="pointer-events-none absolute -bottom-10 -left-8 h-36 w-36 rounded-full bg-emerald-200/50" />
      <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-emerald-600">
            Recommendations
          </p>
          <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">Discover Items</h1>
          <p class="mt-2 max-w-xl text-sm text-slate-600">Browse all available marketplace items. Tap any item to view details and make an offer.</p>
        </div>
        <Button variant="outline" class="gap-2" @click="fetchItems">
          <Compass class="h-4 w-4" /> Refresh
        </Button>
      </div>
    </section>

    <section class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <Input v-model="search" placeholder="Search items..." class="max-w-xs" />
      <div class="flex flex-wrap gap-2">
        <Button v-for="cat in categories" :key="cat" size="sm" :variant="activeCat === cat ? 'default' : 'outline'" class="rounded-full" @click="activeCat = cat">{{ cat }}</Button>
      </div>
    </section>

    <div v-if="isLoading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card v-for="n in 6" :key="n" class="animate-pulse"><CardContent class="p-6 space-y-3"><div class="h-32 bg-muted rounded-xl"/><div class="h-4 bg-muted rounded w-3/4"/><div class="h-3 bg-muted rounded w-1/2"/></CardContent></Card>
    </div>

    <section v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card v-for="item in filteredItems" :key="item.id" class="group bg-card/80 overflow-hidden cursor-pointer transition-all hover:shadow-md hover:-translate-y-0.5" @click="goToItem(item.id)">
        <div class="relative h-40 overflow-hidden bg-muted">
          <img v-if="getImageUrl(item.images)" :src="getImageUrl(item.images)" :alt="item.name" class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105" />
          <div v-else class="flex h-full items-center justify-center"><ImageOff class="h-10 w-10 text-muted-foreground/30" /></div>
          <Badge class="absolute top-2 right-2 text-[10px] capitalize">{{ item.status }}</Badge>
          <Badge v-if="item.type === 'auction'" class="absolute top-2 left-2 text-[10px] bg-amber-500 text-white">Auction</Badge>
        </div>
        <CardContent class="p-4 space-y-2">
          <h3 class="font-semibold line-clamp-1">{{ item.name }}</h3>
          <p class="text-xs text-muted-foreground line-clamp-2 min-h-[2rem]">{{ item.description || 'No description' }}</p>
          <div v-if="item.categories?.length" class="flex flex-wrap gap-1">
            <Badge v-for="cat in item.categories.slice(0, 3)" :key="cat.id" variant="secondary" class="text-[10px]">{{ cat.name }}</Badge>
          </div>
          <Button size="sm" variant="outline" class="w-full mt-1">View Details</Button>
        </CardContent>
      </Card>

      <Card v-if="!filteredItems.length" class="bg-card/80 py-10 col-span-full">
        <CardContent class="text-center">
          <Package class="mx-auto h-10 w-10 text-muted-foreground/30" />
          <p class="mt-2 font-semibold text-muted-foreground">No items found</p>
          <p class="text-xs text-muted-foreground mt-1">Try a different search or category.</p>
        </CardContent>
      </Card>
    </section>
  </div>
</template>
