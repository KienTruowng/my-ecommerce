<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api/api'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Clock,
  Image as ImageIcon,
  MapPin,
  Plus,
  RefreshCw,
  Search,
  Trash2,
  PencilLine,
} from 'lucide-vue-next'

type Location = {
  id: number
  name: string
  description?: string
  address?: string
  image_url?: string
  latitude?: number
  longitude?: number
  open_time?: string
  close_time?: string
  is_active?: boolean
  created_at?: string
  updated_at?: string
}

type LocationListResponse = {
  data: Location[]
  meta: {
    current_page: number
    page_size: number
    total_pages: number
    total_records: number
  }
}

const locations = ref<Location[]>([])
const isLoading = ref(false)
const isSaving = ref(false)
const isSheetOpen = ref(false)
const isDeleteDialogOpen = ref(false)
const editingLocation = ref<Location | null>(null)
const deletingLocation = ref<Location | null>(null)

const page = ref(1)
const limit = ref(9)
const totalPages = ref(1)
const totalRecords = ref(0)

const searchQuery = ref('')
const appliedSearch = ref('')
const activeFilter = ref<'all' | 'active' | 'inactive'>('all')

const formState = ref({
  name: '',
  description: '',
  address: '',
  image_url: '',
  latitude: '',
  longitude: '',
  open_time: '',
  close_time: '',
  is_active: true,
})

const activeFilterValue = computed(() => {
  if (activeFilter.value === 'active') return true
  if (activeFilter.value === 'inactive') return false
  return undefined
})

const imagePreview = computed(() => formState.value.image_url.trim())

const activeCount = computed(() => locations.value.filter((loc) => loc.is_active).length)
const inactiveCount = computed(() => locations.value.filter((loc) => !loc.is_active).length)

const formatCoords = (location: Location) => {
  if (typeof location.latitude === 'number' && typeof location.longitude === 'number') {
    return `${location.latitude.toFixed(5)}, ${location.longitude.toFixed(5)}`
  }
  return 'Chua co toa do'
}

const formatHours = (location: Location) => {
  if (location.open_time && location.close_time) {
    return `${location.open_time} - ${location.close_time}`
  }
  return 'Chua co khung gio'
}

const resetForm = () => {
  formState.value = {
    name: '',
    description: '',
    address: '',
    image_url: '',
    latitude: '',
    longitude: '',
    open_time: '',
    close_time: '',
    is_active: true,
  }
}

const openCreate = () => {
  editingLocation.value = null
  resetForm()
  isSheetOpen.value = true
}

const openEdit = (location: Location) => {
  editingLocation.value = location
  formState.value = {
    name: location.name ?? '',
    description: location.description ?? '',
    address: location.address ?? '',
    image_url: location.image_url ?? '',
    latitude: location.latitude !== undefined ? String(location.latitude) : '',
    longitude: location.longitude !== undefined ? String(location.longitude) : '',
    open_time: location.open_time ?? '',
    close_time: location.close_time ?? '',
    is_active: location.is_active ?? true,
  }
  isSheetOpen.value = true
}

const buildPayload = () => {
  const payload: Record<string, unknown> = {
    name: formState.value.name.trim(),
    is_active: formState.value.is_active,
  }

  const description = formState.value.description.trim()
  if (description) payload.description = description

  const address = formState.value.address.trim()
  if (address) payload.address = address

  const imageUrl = formState.value.image_url.trim()
  if (imageUrl) payload.image_url = imageUrl

  const openTime = formState.value.open_time.trim()
  if (openTime) payload.open_time = openTime

  const closeTime = formState.value.close_time.trim()
  if (closeTime) payload.close_time = closeTime

  const latitude = Number(formState.value.latitude)
  if (!Number.isNaN(latitude) && formState.value.latitude !== '') payload.latitude = latitude

  const longitude = Number(formState.value.longitude)
  if (!Number.isNaN(longitude) && formState.value.longitude !== '') payload.longitude = longitude

  return payload
}

const fetchLocations = async () => {
  isLoading.value = true
  try {
    const { data, error } = await api.GET('/api/locations', {
      params: {
        query: {
          page: page.value,
          limit: limit.value,
          search: appliedSearch.value || undefined,
          active: activeFilterValue.value,
        },
      },
    })

    if (error) {
      toast.error('Khong the tai danh sach dia diem')
      return
    }

    const response = data as LocationListResponse | undefined
    locations.value = response?.data ?? []
    totalPages.value = response?.meta?.total_pages ?? 1
    totalRecords.value = response?.meta?.total_records ?? locations.value.length
  } finally {
    isLoading.value = false
  }
}

const applySearch = async () => {
  appliedSearch.value = searchQuery.value.trim()
  page.value = 1
  await fetchLocations()
}

const clearSearch = async () => {
  searchQuery.value = ''
  appliedSearch.value = ''
  page.value = 1
  await fetchLocations()
}

const setFilter = async (value: 'all' | 'active' | 'inactive') => {
  activeFilter.value = value
  page.value = 1
  await fetchLocations()
}

const submitForm = async () => {
  const name = formState.value.name.trim()
  if (!name) {
    toast.error('Ten dia diem khong duoc de trong')
    return
  }

  isSaving.value = true
  try {
    const payload = buildPayload()
    if (editingLocation.value) {
      const { error } = await api.PATCH('/api/admin/locations/{id}', {
        params: { path: { id: editingLocation.value.id } },
        body: payload,
      })

      if (error) {
        toast.error('Khong the cap nhat dia diem')
        return
      }

      toast.success('Da cap nhat dia diem')
    } else {
      const { error } = await api.POST('/api/admin/locations', {
        body: payload,
      })

      if (error) {
        toast.error('Khong the tao dia diem')
        return
      }

      toast.success('Da tao dia diem moi')
    }

    isSheetOpen.value = false
    await fetchLocations()
  } finally {
    isSaving.value = false
  }
}

const confirmDelete = (location: Location) => {
  deletingLocation.value = location
  isDeleteDialogOpen.value = true
}

const deleteLocation = async () => {
  if (!deletingLocation.value) return

  try {
    const { error } = await api.DELETE('/api/admin/locations/{id}', {
      params: { path: { id: deletingLocation.value.id } },
    })

    if (error) {
      toast.error('Khong the xoa dia diem')
      return
    }

    toast.success('Da xoa dia diem')
    isDeleteDialogOpen.value = false
    deletingLocation.value = null
    await fetchLocations()
  } catch (error) {
    console.error(error)
  }
}

onMounted(fetchLocations)
</script>

<template>
  <div class="space-y-6">
    <section
      class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-amber-50 via-white to-emerald-50 p-6"
    >
      <div class="pointer-events-none absolute -right-10 -top-10 h-32 w-32 rounded-full bg-amber-100/60" />
      <div class="pointer-events-none absolute -bottom-12 -left-12 h-40 w-40 rounded-full bg-emerald-100/50" />
      <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-emerald-600">
            Campus Location Management
          </p>
          <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">
            Quan ly dia diem & khu vuc giao dich
          </h1>
          <p class="mt-2 max-w-xl text-sm text-slate-600">
            Dinh nghia cac khu vuc duoc phep giao dich trong khuon vien truong, gan toa do, hinh
            anh va khung gio hoat dong.
          </p>
          <div class="mt-4 flex flex-wrap gap-2 text-xs text-slate-600">
            <span class="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1">
              Dang hien thi: {{ locations.length }} dia diem
            </span>
            <span class="rounded-full border border-amber-200 bg-amber-50 px-3 py-1">
              Dang hoat dong: {{ activeCount }}
            </span>
            <span class="rounded-full border border-slate-200 bg-white px-3 py-1">
              Tam dung: {{ inactiveCount }}
            </span>
          </div>
        </div>
        <div class="flex flex-wrap gap-2">
          <Button variant="outline" class="gap-2" @click="fetchLocations">
            <RefreshCw class="h-4 w-4" />
            Lam moi
          </Button>
          <Button class="gap-2" @click="openCreate">
            <Plus class="h-4 w-4" />
            Them dia diem
          </Button>
        </div>
      </div>
    </section>

    <section class="grid gap-4 md:grid-cols-3">
      <Card class="bg-card/80">
        <CardHeader>
          <CardTitle class="text-lg">Tong quan</CardTitle>
          <CardDescription>Thong ke nhanh cho khu vuc giao dich.</CardDescription>
        </CardHeader>
        <CardContent class="space-y-2 text-sm text-muted-foreground">
          <div class="flex items-center justify-between">
            <span>Tong so dia diem</span>
            <span class="font-semibold text-foreground">{{ totalRecords }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span>Trang hien tai</span>
            <span class="font-semibold text-foreground">{{ page }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span>Tong so trang</span>
            <span class="font-semibold text-foreground">{{ totalPages }}</span>
          </div>
        </CardContent>
      </Card>
      <Card class="bg-card/80">
        <CardHeader>
          <CardTitle class="text-lg">Goi y van hanh</CardTitle>
          <CardDescription>Bo sung thong tin de nguoi dung de chon diem giao.</CardDescription>
        </CardHeader>
        <CardContent class="space-y-2 text-sm text-muted-foreground">
          <p>+ Them anh dai dien cho khu vuc de tang do nhan dien.</p>
          <p>+ Neu co toa do, ban do se chinh xac hon.</p>
          <p>+ Dat khung gio de nguoi dung hen gap dung luc.</p>
        </CardContent>
      </Card>
      <Card class="bg-card/80">
        <CardHeader>
          <CardTitle class="text-lg">Loc nhanh</CardTitle>
          <CardDescription>Quan ly danh sach theo trang thai.</CardDescription>
        </CardHeader>
        <CardContent class="flex flex-wrap gap-2">
          <Button
            size="sm"
            :variant="activeFilter === 'all' ? 'default' : 'outline'"
            @click="setFilter('all')"
          >
            Tat ca
          </Button>
          <Button
            size="sm"
            :variant="activeFilter === 'active' ? 'default' : 'outline'"
            @click="setFilter('active')"
          >
            Dang hoat dong
          </Button>
          <Button
            size="sm"
            :variant="activeFilter === 'inactive' ? 'default' : 'outline'"
            @click="setFilter('inactive')"
          >
            Tam dung
          </Button>
        </CardContent>
      </Card>
    </section>

    <section class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex flex-1 flex-wrap items-center gap-2">
        <Input
          v-model="searchQuery"
          placeholder="Tim theo ten hoac dia chi"
          class="max-w-xs"
        />
        <Button variant="outline" class="gap-2" @click="applySearch">
          <Search class="h-4 w-4" />
          Tim kiem
        </Button>
        <Button variant="ghost" @click="clearSearch">Xoa loc</Button>
      </div>
      <div class="flex items-center gap-2 text-xs text-muted-foreground">
        <span>Trang {{ page }} / {{ totalPages }}</span>
        <Button
          size="sm"
          variant="outline"
          :disabled="page <= 1"
          @click="page = Math.max(1, page - 1); fetchLocations()"
        >
          Truoc
        </Button>
        <Button
          size="sm"
          variant="outline"
          :disabled="page >= totalPages"
          @click="page = Math.min(totalPages, page + 1); fetchLocations()"
        >
          Sau
        </Button>
      </div>
    </section>

    <section v-if="isLoading" class="text-sm text-muted-foreground">Dang tai danh sach...</section>

    <section v-else class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <Card v-for="location in locations" :key="location.id" class="overflow-hidden">
        <div class="relative aspect-[16/9]">
          <img
            v-if="location.image_url"
            :src="location.image_url"
            :alt="location.name"
            class="h-full w-full object-cover"
          />
          <div
            v-else
            class="flex h-full w-full items-center justify-center bg-gradient-to-br from-slate-100 via-white to-emerald-100"
          >
            <ImageIcon class="h-10 w-10 text-slate-400" />
          </div>
          <Badge
            class="absolute right-3 top-3"
            :variant="location.is_active ? 'default' : 'secondary'"
          >
            {{ location.is_active ? 'Dang hoat dong' : 'Tam dung' }}
          </Badge>
        </div>
        <CardHeader>
          <CardTitle class="text-lg">{{ location.name }}</CardTitle>
          <CardDescription>
            {{ location.description || 'Chua co mo ta chi tiet.' }}
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-3 text-sm text-muted-foreground">
          <div class="flex items-start gap-2">
            <MapPin class="mt-0.5 h-4 w-4 text-emerald-600" />
            <div>
              <p class="font-medium text-foreground">
                {{ location.address || 'Chua cap nhat dia chi.' }}
              </p>
              <p class="text-xs">{{ formatCoords(location) }}</p>
            </div>
          </div>
          <div class="flex items-start gap-2">
            <Clock class="mt-0.5 h-4 w-4 text-amber-500" />
            <div>
              <p class="font-medium text-foreground">Khung gio hoat dong</p>
              <p class="text-xs">{{ formatHours(location) }}</p>
            </div>
          </div>
          <div class="flex flex-wrap gap-2 pt-2">
            <Button size="sm" variant="outline" class="gap-2" @click="openEdit(location)">
              <PencilLine class="h-4 w-4" />
              Chinh sua
            </Button>
            <Button size="sm" variant="destructive" class="gap-2" @click="confirmDelete(location)">
              <Trash2 class="h-4 w-4" />
              Xoa
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card v-if="!locations.length" class="bg-card/80">
        <CardHeader>
          <CardTitle>Chua co dia diem nao</CardTitle>
          <CardDescription>Them dia diem dau tien cho campus marketplace.</CardDescription>
        </CardHeader>
      </Card>
    </section>
  </div>

  <Sheet v-model:open="isSheetOpen">
    <SheetContent class="flex flex-col">
      <SheetHeader>
        <SheetTitle>{{ editingLocation ? 'Cap nhat dia diem' : 'Them dia diem moi' }}</SheetTitle>
        <SheetDescription>
          Cung cap ten khu vuc, anh dai dien va thong tin hoat dong.
        </SheetDescription>
      </SheetHeader>

      <div class="mt-6 flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
        <div class="grid gap-2">
          <Label for="location-name">Ten khu vuc</Label>
          <Input id="location-name" v-model="formState.name" placeholder="VD: Thu vien A" />
        </div>

        <div class="grid gap-2">
          <Label for="location-description">Mo ta ngan</Label>
          <Textarea
            id="location-description"
            v-model="formState.description"
            placeholder="Mo ta dac diem, thong tin huong dan..."
          />
        </div>

        <div class="grid gap-2">
          <Label for="location-address">Dia chi / Vi tri</Label>
          <Input
            id="location-address"
            v-model="formState.address"
            placeholder="VD: Tang 1, toa nha A"
          />
        </div>

        <div class="grid gap-2">
          <Label for="location-image">Anh dai dien (URL)</Label>
          <Input
            id="location-image"
            v-model="formState.image_url"
            placeholder="https://"
          />
        </div>

        <div class="overflow-hidden rounded-xl border">
          <div v-if="imagePreview" class="aspect-[16/9] w-full">
            <img :src="imagePreview" alt="Preview" class="h-full w-full object-cover" />
          </div>
          <div
            v-else
            class="flex aspect-[16/9] w-full items-center justify-center bg-gradient-to-br from-slate-50 via-white to-emerald-50"
          >
            <div class="text-center text-xs text-muted-foreground">
              <ImageIcon class="mx-auto mb-2 h-6 w-6" />
              Chua co anh xem truoc
            </div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="grid gap-2">
            <Label for="location-lat">Vi do</Label>
            <Input
              id="location-lat"
              v-model="formState.latitude"
              type="number"
              step="0.000001"
              placeholder="10.762622"
            />
          </div>
          <div class="grid gap-2">
            <Label for="location-lng">Kinh do</Label>
            <Input
              id="location-lng"
              v-model="formState.longitude"
              type="number"
              step="0.000001"
              placeholder="106.660172"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="grid gap-2">
            <Label for="location-open">Gio mo cua</Label>
            <Input id="location-open" v-model="formState.open_time" type="time" />
          </div>
          <div class="grid gap-2">
            <Label for="location-close">Gio dong cua</Label>
            <Input id="location-close" v-model="formState.close_time" type="time" />
          </div>
        </div>

        <label class="flex items-center gap-2 text-sm font-medium text-foreground">
          <input
            v-model="formState.is_active"
            type="checkbox"
            class="h-4 w-4 rounded border-slate-300 text-emerald-600 focus:ring-emerald-500"
          />
          Khu vuc dang hoat dong
        </label>
      </div>

      <SheetFooter class="mt-6">
        <Button variant="outline" @click="isSheetOpen = false">Huy</Button>
        <Button :disabled="isSaving" @click="submitForm">
          {{ isSaving ? 'Dang luu...' : 'Luu thong tin' }}
        </Button>
      </SheetFooter>
    </SheetContent>
  </Sheet>

  <Dialog v-model:open="isDeleteDialogOpen">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Xoa dia diem</DialogTitle>
        <DialogDescription>
          Ban chac chan muon xoa {{ deletingLocation?.name }}? Thao tac nay se an dia diem khoi
          he thong.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button variant="outline" @click="isDeleteDialogOpen = false">Huy</Button>
        <Button variant="destructive" @click="deleteLocation">Xoa</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
