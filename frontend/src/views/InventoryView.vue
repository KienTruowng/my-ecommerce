<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { Package, Pencil, Trash2, Plus, ImagePlus, Tag, FileText, ShoppingBag, Upload, X, Loader2, ArrowRightLeft, Gavel } from 'lucide-vue-next'
import ImageLightbox from '@/components/ImageLightbox.vue'
type ItemImage = { id: number; url: string }
type Item = {
  id: number
  name: string
  description: string
  status: string
  type?: string
  images?: ItemImage[]
}

type Category = { id: number; name: string }

const items = ref<Item[]>([])
const categories = ref<Category[]>([])
const statusFilter = ref('')
const isLoading = ref(false)
const isSaving = ref(false)
const isDialogOpen = ref(false)
const editingId = ref<number | null>(null)

// Lightbox state
const lightboxOpen = ref(false)
const lightboxImages = ref<{ src: string; alt: string }[]>([])
const lightboxStartIdx = ref(0)

const openLightbox = (images: { src: string; alt: string }[], idx: number) => {
  lightboxImages.value = images
  lightboxStartIdx.value = idx
  lightboxOpen.value = true
}

const form = ref({
  name: '',
  description: '',
  categoryId: null as number | null,
  existingImageUrls: [] as string[],
  imageFiles: [] as { file: File; preview: string }[],
  type: 'barter',
  auction: {
    basePrice: '',
    stepPrice: '',
    depositAmount: '',
    deadline: '',
  },
})

const dialogTitle = computed(() => (editingId.value ? 'Edit item' : 'Create item'))

const resetForm = () => {
  form.value = {
    name: '',
    description: '',
    categoryId: null,
    existingImageUrls: [],
    imageFiles: [],
    type: 'barter',
    auction: {
      basePrice: '',
      stepPrice: '',
      depositAmount: '',
      deadline: '',
    },
  }
}

const isEditing = computed(() => editingId.value !== null)
const showAuctionConfig = computed(() => !isEditing.value && form.value.type === 'auction')

const isDragOver = ref(false)

const selectedCategoryName = computed(() => {
  if (!form.value.categoryId) return null
  return categories.value.find((c) => c.id === form.value.categoryId)?.name ?? null
})

const totalImages = computed(() => form.value.existingImageUrls.length + form.value.imageFiles.length)

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files) {
    Array.from(target.files).forEach((file) => {
      form.value.imageFiles.push({
        file,
        preview: URL.createObjectURL(file),
      })
    })
  }
  target.value = ''
}

const removeFile = (index: number) => {
  const [removed] = form.value.imageFiles.splice(index, 1)
  if (removed) {
    URL.revokeObjectURL(removed.preview)
  }
}

const removeExistingImage = (index: number) => {
  form.value.existingImageUrls.splice(index, 1)
}

const handleDragOver = (e: DragEvent) => {
  e.preventDefault()
  isDragOver.value = true
}

const handleDragLeave = () => {
  isDragOver.value = false
}

const handleDrop = (e: DragEvent) => {
  e.preventDefault()
  isDragOver.value = false
  if (e.dataTransfer?.files) {
    Array.from(e.dataTransfer.files).forEach((file) => {
      if (file.type.startsWith('image/')) {
        form.value.imageFiles.push({
          file,
          preview: URL.createObjectURL(file),
        })
      }
    })
  }
}

const openCreate = () => {
  resetForm()
  editingId.value = null
  isDialogOpen.value = true
}

const openEdit = async (item: Item) => {
  editingId.value = item.id
  resetForm()
  form.value.name = item.name
  form.value.description = item.description
  form.value.existingImageUrls = item.images?.map((img) => img.url) ?? []
  form.value.imageFiles = []
  form.value.type = item.type ?? 'barter'

  try {
    const { data } = await api.GET('/api/items/{id}', {
      params: { path: { id: item.id } },
    })
    const raw = data?.data ?? data
    const categoryIds =
      raw?.Categories?.map((cat: { ID: number }) => cat.ID) ??
      raw?.categories?.map((cat: { id: number }) => cat.id) ??
      []
    form.value.categoryId = categoryIds[0] ?? null
  } catch (error) {
    console.error(error)
  }

  isDialogOpen.value = true
}

const fetchCategories = async () => {
  const { data, error } = await api.GET('/api/categories', {
    params: { query: { page: 1, limit: 50 } },
  })

  if (error) {
    toast.error('Unable to load categories')
    return
  }

  categories.value = data?.data ?? []
}

const fetchItems = async () => {
  isLoading.value = true
  try {
    const { data, error } = await api.GET('/api/users/me/items', {
      params: { query: statusFilter.value ? { status: statusFilter.value } : {} },
    })

    if (error) {
      toast.error('Unable to load your inventory')
      return
    }

    items.value = data ?? []
  } finally {
    isLoading.value = false
  }
}

const saveItem = async () => {
  isSaving.value = true
  const isAuction = form.value.type === 'auction'
  const auctionBase = Number(form.value.auction.basePrice)
  const auctionStep = Number(form.value.auction.stepPrice)
  const auctionDeposit = Number(form.value.auction.depositAmount)
  const auctionDeadline = form.value.auction.deadline

  if (!form.value.name.trim()) {
    toast.error('Name is required')
    isSaving.value = false
    return
  }

  if (isAuction && !editingId.value) {
    if (!auctionDeadline) {
      toast.error('Auction deadline is required')
      isSaving.value = false
      return
    }
    if (!auctionBase || auctionBase <= 0) {
      toast.error('Base price must be greater than 0')
      isSaving.value = false
      return
    }
    if (!auctionStep || auctionStep <= 0) {
      toast.error('Step price must be greater than 0')
      isSaving.value = false
      return
    }
    if (auctionStep > Math.floor(auctionBase / 5)) {
      toast.error('Step price cannot exceed 20% of base price')
      isSaving.value = false
      return
    }
  }

  const payload = {
    name: form.value.name,
    description: form.value.description,
    category_ids: form.value.categoryId ? [form.value.categoryId] : [],
    image_urls: [...form.value.existingImageUrls],
  }

  try {
    let currentItemId = editingId.value

    if (currentItemId) {
      const { error } = await api.PATCH('/api/items/{id}', {
        params: { path: { id: currentItemId } },
        body: payload,
      })

      if (error) {
        toast.error('Unable to update item')
        return
      }
    } else {
      const { data, error } = await api.POST('/api/items', {
        body: payload,
      })

      if (error) {
        toast.error('Unable to create item')
        return
      }

      currentItemId = (data as any)?.data?.id ?? (data as any)?.id

      if (isAuction && currentItemId) {
        const { error: auctionError } = await api.POST('/api/items/{id}/auction', {
          params: { path: { id: currentItemId } },
          body: {
            base_price: auctionBase,
            step_price: auctionStep,
            deposit_amount: auctionDeposit || 0,
            deadline: new Date(auctionDeadline).toISOString(),
          },
        })

        if (auctionError) {
          toast.error('Item created, but auction config failed')
        }
      }
    }

    if (currentItemId && form.value.imageFiles.length > 0) {
      for (const imgFile of form.value.imageFiles) {
        const { file } = imgFile
        const { data: presignData, error: presignError } = await api.POST('/api/items/images/presign', {
          body: {
            filename: file.name,
            file_type: file.type || 'application/octet-stream',
          },
        })

        if (presignError) {
          toast.error(`Failed to get presigned URL for ${file.name}`)
          continue
        }

        const { upload_url, object_key } = presignData as any

        try {
          const uploadRes = await fetch(upload_url, {
            method: 'PUT',
            body: file,
            headers: {
              'Content-Type': file.type || 'application/octet-stream',
            },
          })

          if (!uploadRes.ok) {
            toast.error(`Failed to upload ${file.name}`)
            continue
          }

          const { error: confirmError } = await api.POST('/api/items/{id}/images', {
            params: { path: { id: currentItemId } },
            body: { object_key },
          })

          if (confirmError) {
            toast.error(`Failed to confirm image ${file.name}`)
          }
        } catch (e) {
          toast.error(`Error uploading ${file.name}`)
        }
      }
    }

    toast.success(editingId.value ? 'Item updated' : 'Item created')
    isDialogOpen.value = false
    await fetchItems()
  } finally {
    isSaving.value = false
  }
}

const deleteItem = async (id: number) => {
  const { error } = await api.DELETE('/api/items/{id}', {
    params: { path: { id } },
  })

  if (error) {
    toast.error('Unable to delete item')
    return
  }

  toast.success('Item deleted')
  items.value = items.value.filter((item) => item.id !== id)
}

onMounted(async () => {
  await Promise.all([fetchCategories(), fetchItems()])
})
</script>

<template>
  <div class="space-y-8">
    <!-- Page Banner -->
    <section class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-emerald-50 via-white to-teal-50 p-6 lg:p-8">
      <div class="pointer-events-none absolute -right-8 -top-10 h-36 w-36 rounded-full bg-emerald-200/50" />
      <div class="pointer-events-none absolute -bottom-14 -left-10 h-44 w-44 rounded-full bg-teal-200/40" />
      <div class="pointer-events-none absolute right-1/3 top-1/2 h-20 w-20 rounded-full bg-emerald-300/30" />

      <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-emerald-600">My Inventory</p>
          <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">Your listings</h1>
          <p class="mt-2 max-w-xl text-sm text-slate-600">
            Manage your items, tweak details, and keep your storefront looking sharp.
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <!-- Quick stats -->
          <div class="hidden sm:flex items-center gap-4 rounded-xl border bg-white/70 px-4 py-2 text-sm backdrop-blur">
            <span class="text-muted-foreground">
              <span class="font-semibold text-foreground">{{ items.filter(i => i.status === 'available').length }}</span> available
            </span>
            <span class="h-4 w-px bg-border" />
            <span class="text-muted-foreground">
              <span class="font-semibold text-foreground">{{ items.filter(i => i.type === 'auction').length }}</span> auctions
            </span>
          </div>
          <select v-model="statusFilter" class="h-10 rounded-lg border border-input bg-white/80 px-3 text-sm backdrop-blur"
            @change="fetchItems">
            <option value="">All statuses</option>
            <option value="available">Available</option>
            <option value="reserved">Reserved</option>
            <option value="traded">Traded</option>
            <option value="hidden">Hidden</option>
          </select>
          <Dialog v-model:open="isDialogOpen">
            <DialogTrigger as-child>
              <Button @click="openCreate" class="gap-1.5 shadow-sm">
                <Plus class="h-4 w-4" /> New item
              </Button>
            </DialogTrigger>
            <DialogContent class="w-[95vw] sm:max-w-5xl max-h-[90vh] overflow-y-auto">
              <DialogHeader>
                <DialogTitle class="flex items-center gap-2 text-xl">
                  <div class="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-100 text-emerald-600">
                    <Plus v-if="!isEditing" class="h-4 w-4" />
                    <Pencil v-else class="h-4 w-4" />
                  </div>
                  {{ dialogTitle }}
                </DialogTitle>
                <DialogDescription>
                  {{ isEditing ? 'Update your item details below.' : 'List a new item on the marketplace in seconds.' }}
                </DialogDescription>
              </DialogHeader>

              <div class="grid gap-6 py-2 lg:grid-cols-2">
                <!-- Left Column — Core details -->
                <div class="space-y-4">
                  <div class="grid gap-1.5">
                    <label class="flex items-center gap-1.5 text-sm font-medium">
                      <FileText class="h-3.5 w-3.5 text-emerald-500" />
                      Name <span class="text-destructive">*</span>
                    </label>
                    <Input v-model="form.name" placeholder="e.g. Macbook stand, Calculus textbook..." />
                  </div>

                  <div class="grid gap-1.5">
                    <label class="flex items-center gap-1.5 text-sm font-medium">
                      <FileText class="h-3.5 w-3.5 text-emerald-500" />
                      Description
                    </label>
                    <Textarea v-model="form.description" rows="3" placeholder="Condition, color, extras… anything a buyer should know" />
                    <p class="text-xs text-muted-foreground self-end">{{ form.description.length }} / 500</p>
                  </div>

                  <div class="grid gap-1.5">
                    <label class="flex items-center gap-1.5 text-sm font-medium">
                      <ShoppingBag class="h-3.5 w-3.5 text-emerald-500" />
                      Listing type
                    </label>
                    <div class="grid grid-cols-2 gap-3 mt-1">
                      <!-- Barter Option -->
                      <div
                        @click="!isEditing && (form.type = 'barter')"
                        :class="[
                          'relative flex cursor-pointer flex-col gap-2 rounded-xl border-2 p-4 transition-all',
                          form.type === 'barter' ? 'border-emerald-500 bg-emerald-50 shadow-sm' : 'border-muted hover:border-emerald-200 hover:bg-emerald-50/30',
                          isEditing ? 'opacity-50 cursor-not-allowed' : ''
                        ]"
                      >
                        <div class="flex items-center justify-between">
                          <div class="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-100 text-emerald-600">
                            <ArrowRightLeft class="h-4 w-4" />
                          </div>
                          <div v-if="form.type === 'barter'" class="h-4 w-4 rounded-full border-[4px] border-emerald-500 bg-white shadow-sm" />
                          <div v-else class="h-4 w-4 rounded-full border-2 border-muted" />
                        </div>
                        <div class="mt-1">
                          <p class="font-semibold text-foreground text-sm">Barter / Fixed</p>
                          <p class="text-[11px] text-muted-foreground mt-0.5 leading-tight">Direct exchange or setting a fixed cash price</p>
                        </div>
                      </div>

                      <!-- Auction Option -->
                      <div
                        @click="!isEditing && (form.type = 'auction')"
                        :class="[
                          'relative flex cursor-pointer flex-col gap-2 rounded-xl border-2 p-4 transition-all',
                          form.type === 'auction' ? 'border-amber-500 bg-amber-50 shadow-sm' : 'border-muted hover:border-amber-200 hover:bg-amber-50/30',
                          isEditing ? 'opacity-50 cursor-not-allowed' : ''
                        ]"
                      >
                        <div class="flex items-center justify-between">
                          <div class="flex h-8 w-8 items-center justify-center rounded-full bg-amber-100 text-amber-600">
                            <Gavel class="h-4 w-4" />
                          </div>
                          <div v-if="form.type === 'auction'" class="h-4 w-4 rounded-full border-[4px] border-amber-500 bg-white shadow-sm" />
                          <div v-else class="h-4 w-4 rounded-full border-2 border-muted" />
                        </div>
                        <div class="mt-1">
                          <p class="font-semibold text-foreground text-sm">Auction</p>
                          <p class="text-[11px] text-muted-foreground mt-0.5 leading-tight">Let others bid on your item with a base price</p>
                        </div>
                      </div>
                    </div>
                    <p v-if="isEditing" class="text-xs text-muted-foreground mt-1">Listing type is locked after creation.</p>
                  </div>

                  <Transition name="auction-collapse">
                    <div v-if="showAuctionConfig" class="grid gap-4 rounded-2xl border-2 border-amber-200 bg-amber-50 p-5 mt-2 shadow-sm">
                      <div class="flex items-center gap-3">
                        <div class="flex h-10 w-10 items-center justify-center rounded-full bg-amber-100 text-amber-600 shadow-inner">
                          <Gavel class="h-5 w-5" />
                        </div>
                        <div>
                          <p class="text-base font-semibold text-amber-950">Auction Configuration</p>
                          <p class="text-xs text-amber-700/80">Set your base price, bidding steps, and timeframe.</p>
                        </div>
                      </div>
                      <div class="grid gap-3 sm:grid-cols-2">
                        <div class="grid gap-1.5">
                          <label class="text-xs font-medium">Base price (VND)</label>
                          <Input v-model="form.auction.basePrice" type="number" min="0" placeholder="100 000" />
                        </div>
                        <div class="grid gap-1.5">
                          <label class="text-xs font-medium">Step price (VND)</label>
                          <Input v-model="form.auction.stepPrice" type="number" min="0" placeholder="5 000" />
                        </div>
                      </div>
                      <div class="grid gap-3 sm:grid-cols-2">
                        <div class="grid gap-1.5">
                          <label class="text-xs font-medium">Deposit <span class="text-muted-foreground">(optional)</span></label>
                          <Input v-model="form.auction.depositAmount" type="number" min="0" placeholder="0" />
                        </div>
                        <div class="grid gap-1.5">
                          <label class="text-xs font-medium">Deadline <span class="text-destructive">*</span></label>
                          <Input v-model="form.auction.deadline" type="datetime-local" />
                        </div>
                      </div>
                    </div>
                  </Transition>
                </div>

                <!-- Right Column — Category & Images -->
                <div class="space-y-4">
                  <div class="grid gap-1.5">
                    <label class="flex items-center gap-1.5 text-sm font-medium">
                      <Tag class="h-3.5 w-3.5 text-emerald-500" />
                      Category
                    </label>
                    <select
                      v-model="form.categoryId"
                      class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm"
                    >
                      <option :value="null" disabled>Select a category…</option>
                      <option v-for="category in categories" :key="category.id" :value="category.id">
                        {{ category.name }}
                      </option>
                    </select>
                    <p v-if="selectedCategoryName" class="inline-flex items-center gap-1 self-start rounded-full bg-emerald-50 px-2.5 py-0.5 text-xs font-medium text-emerald-700">
                      ✓ {{ selectedCategoryName }}
                    </p>
                  </div>

                  <div class="grid gap-1.5 min-w-0">
                    <label class="flex items-center gap-1.5 text-sm font-medium">
                      <ImagePlus class="h-3.5 w-3.5 text-emerald-500" />
                      Images
                      <Badge v-if="totalImages > 0" variant="secondary" class="ml-auto h-5 px-1.5 text-[10px]">{{ totalImages }}</Badge>
                    </label>

                    <div
                      v-if="totalImages > 0"
                      class="flex gap-3 overflow-x-auto pb-3 pt-2 px-0.5 scrollbar-thin"
                    >
                      <div v-for="(url, index) in form.existingImageUrls" :key="'ex-' + index" class="relative group shrink-0">
                        <img :src="url" class="h-28 w-28 rounded-lg border object-cover shadow-sm cursor-pointer transition-transform hover:scale-105"
                          @click="openLightbox(form.existingImageUrls.map((u, i) => ({ src: u, alt: `Image ${i + 1}` })), index)" />
                        <span class="absolute bottom-1.5 left-1.5 rounded bg-black/60 px-1.5 py-0.5 text-[10px] text-white">{{ index + 1 }}</span>
                        <button type="button" @click.stop="removeExistingImage(index)"
                          class="absolute -right-1.5 -top-1.5 z-10 flex h-5 w-5 items-center justify-center rounded-full bg-destructive text-white opacity-0 transition-opacity group-hover:opacity-100 shadow text-xs leading-none">✕</button>
                      </div>
                      <div v-for="(imgFile, index) in form.imageFiles" :key="'file-' + index" class="relative group shrink-0">
                        <img :src="imgFile.preview" class="h-28 w-28 rounded-lg border object-cover shadow-sm cursor-pointer transition-transform hover:scale-105"
                          @click="openLightbox(form.imageFiles.map((f, i) => ({ src: f.preview, alt: `New image ${i + 1}` })), index)" />
                        <span class="absolute bottom-1.5 left-1.5 rounded bg-emerald-600/80 px-1.5 py-0.5 text-[10px] text-white">New</span>
                        <button type="button" @click.stop="removeFile(index)"
                          class="absolute -right-1.5 -top-1.5 z-10 flex h-5 w-5 items-center justify-center rounded-full bg-destructive text-white opacity-0 transition-opacity group-hover:opacity-100 shadow text-xs leading-none">✕</button>
                      </div>
                    </div>

                    <div v-else class="rounded-xl border-2 border-dashed border-muted-foreground/20 bg-muted/20 px-4 py-8 text-center">
                      <ImagePlus class="mx-auto h-8 w-8 text-muted-foreground/30" />
                      <p class="mt-2 text-xs text-muted-foreground">No images yet — add some below</p>
                    </div>

                    <label
                      for="image-upload-input"
                      :class="[
                        'flex cursor-pointer flex-col items-center gap-2 rounded-xl border-2 border-dashed px-4 py-5 transition-all',
                        isDragOver
                          ? 'border-emerald-500 bg-emerald-50/60 scale-[1.02]'
                          : 'border-muted-foreground/25 bg-muted/20 hover:border-emerald-400/50 hover:bg-emerald-50/20',
                      ]"
                      @dragover="handleDragOver"
                      @dragleave="handleDragLeave"
                      @drop="handleDrop"
                    >
                      <div :class="[
                        'flex h-10 w-10 items-center justify-center rounded-full transition-colors',
                        isDragOver ? 'bg-emerald-500 text-white' : 'bg-emerald-100 text-emerald-600',
                      ]">
                        <Upload v-if="isDragOver" class="h-5 w-5" />
                        <ImagePlus v-else class="h-5 w-5" />
                      </div>
                      <p class="text-sm font-medium text-muted-foreground">
                        <template v-if="isDragOver">Release to upload</template>
                        <template v-else>Drop images here or <span class="text-emerald-600 font-semibold">browse</span></template>
                      </p>
                      <p class="text-xs text-muted-foreground">PNG, JPG, WEBP · up to 10 MB each</p>
                    </label>
                    <input id="image-upload-input" type="file" multiple accept="image/*" @change="handleFileChange" class="hidden" />
                  </div>
                </div>
              </div>

              <DialogFooter class="sticky bottom-0 -mx-4 -mb-4 rounded-b-xl border-t bg-background/95 backdrop-blur px-4 py-3">
                <Button type="button" variant="outline" @click="isDialogOpen = false">
                  Cancel
                </Button>
                <Button type="button" :disabled="isSaving" @click="saveItem">
                  <Loader2 v-if="isSaving" class="h-4 w-4 animate-spin" />
                  {{ isSaving ? 'Saving…' : isEditing ? 'Save changes' : 'Publish item' }}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>
    </section>

    <!-- Loading state -->
    <section v-if="isLoading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <div v-for="n in 4" :key="n" class="rounded-2xl border bg-card/80 animate-pulse overflow-hidden">
        <div class="h-44 bg-muted" />
        <div class="p-4 space-y-3">
          <div class="h-4 bg-muted rounded w-3/4" />
          <div class="h-3 bg-muted rounded w-full" />
          <div class="h-3 bg-muted rounded w-1/2" />
        </div>
      </div>
    </section>

    <!-- Items Grid -->
    <section v-else class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <Card
        v-for="item in items"
        :key="item.id"
        class="group relative bg-card/90 overflow-hidden transition-all duration-300 hover:shadow-lg hover:-translate-y-1 border-0 ring-1 ring-border/50 hover:ring-emerald-300/60"
      >
        <!-- Image area -->
        <div class="relative h-48 overflow-hidden bg-gradient-to-br from-slate-100 to-slate-50 dark:from-slate-800 dark:to-slate-900">
          <img
            v-if="(item.images ?? [])[0]?.url"
            :src="(item.images ?? [])[0]!.url"
            :alt="item.name"
            class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
          />
          <div v-else class="flex h-full items-center justify-center">
            <Package class="h-12 w-12 text-slate-300 dark:text-slate-600" />
          </div>

          <!-- Overlay gradient on hover -->
          <div class="absolute inset-0 bg-gradient-to-t from-black/30 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

          <!-- Status badges -->
          <div class="absolute top-3 right-3 flex gap-1.5">
            <Badge
              :variant="item.status === 'available' ? 'default' : item.status === 'reserved' ? 'secondary' : 'outline'"
              class="text-[10px] capitalize shadow-sm backdrop-blur-sm"
            >{{ item.status }}</Badge>
            <Badge v-if="item.type === 'auction'" class="text-[10px] bg-amber-500 text-white shadow-sm">Auction</Badge>
          </div>

          <!-- Image count -->
          <span v-if="(item.images ?? []).length > 1"
            class="absolute bottom-3 left-3 rounded-full bg-black/50 px-2 py-0.5 text-[10px] text-white backdrop-blur-sm">
            {{ (item.images ?? []).length }} photos
          </span>
        </div>

        <!-- Content -->
        <CardContent class="p-5 space-y-3">
          <div>
            <h3 class="font-semibold text-base line-clamp-1">{{ item.name }}</h3>
            <p class="text-xs text-muted-foreground line-clamp-2 mt-1.5 leading-relaxed">
              {{ item.description || 'No description' }}
            </p>
          </div>

          <!-- Thumbnail strip -->
          <div v-if="(item.images ?? []).length > 1" class="flex gap-2 overflow-x-auto pb-1 scrollbar-thin">
            <img
              v-for="(image, imgIdx) in (item.images ?? [])"
              :key="image.id"
              :src="image.url"
              class="h-12 w-12 shrink-0 rounded-lg border border-border/50 object-cover shadow-sm cursor-pointer transition hover:scale-105 hover:border-emerald-400/60"
              :alt="item.name"
              @click="openLightbox((item.images ?? []).map(img => ({ src: img.url, alt: item.name })), imgIdx)"
            />
          </div>

          <!-- Actions -->
          <div class="flex gap-2 pt-1">
            <Button size="sm" variant="outline" class="gap-1.5 flex-1 rounded-lg" @click="openEdit(item)">
              <Pencil class="h-3.5 w-3.5" /> Edit
            </Button>
            <Button size="sm" variant="ghost" class="text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg" @click="deleteItem(item.id)">
              <Trash2 class="h-3.5 w-3.5" />
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- Empty state -->
      <div v-if="!items.length" class="col-span-full py-16 text-center">
        <div class="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-emerald-50">
          <Package class="h-10 w-10 text-emerald-400" />
        </div>
        <p class="mt-5 text-lg font-semibold text-foreground">No items yet</p>
        <p class="mt-1 text-sm text-muted-foreground max-w-sm mx-auto">
          Your inventory is empty. Create your first listing to start trading with others on campus.
        </p>
        <Button class="mt-5 gap-1.5 shadow-sm" @click="openCreate">
          <Plus class="h-4 w-4"/> Create your first item
        </Button>
      </div>
    </section>
  </div>

  <!-- Lightbox -->
  <ImageLightbox
    v-model="lightboxOpen"
    :images="lightboxImages"
    :start-index="lightboxStartIdx"
  />
</template>

<style scoped>
.auction-collapse-enter-active,
.auction-collapse-leave-active {
  transition: all 0.25s ease;
}
.auction-collapse-enter-from,
.auction-collapse-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
