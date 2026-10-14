<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '@/api/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { toast } from 'vue-sonner'
import { Truck, Plus, Pencil, Trash2 } from 'lucide-vue-next'

type Shipping = {
  id: number
  name: string
  description: string
  base_price: number
  price_per_km: number
  min_distance: number
  estimated_days: string
  icon_url: string
  is_active: boolean
}

const services = ref<Shipping[]>([])
const isLoading = ref(false)
const editing = ref<Shipping | null>(null)
const form = ref({
  name: '',
  description: '',
  base_price: 0,
  price_per_km: 0,
  min_distance: 0,
  estimated_days: '',
  icon_url: '',
})

const fetchServices = async () => {
  isLoading.value = true
  try {
    const { data, error } = await api.GET('/api/admin/shipping')
    if (!error && data) services.value = (data as any) || []
  } catch { toast.error('Failed to load shipping services') }
  finally { isLoading.value = false }
}

const resetForm = () => {
  form.value = { name: '', description: '', base_price: 0, price_per_km: 0, min_distance: 0, estimated_days: '', icon_url: '' }
  editing.value = null
}

const startEdit = (s: Shipping) => {
  editing.value = s
  form.value = { name: s.name, description: s.description, base_price: s.base_price, price_per_km: s.price_per_km, min_distance: s.min_distance, estimated_days: s.estimated_days, icon_url: s.icon_url }
}

const save = async () => {
  if (!form.value.name.trim()) { toast.error('Name is required'); return }
  try {
    if (editing.value) {
      await api.PATCH('/api/admin/shipping/{id}' as any, { params: { path: { id: editing.value.id } }, body: form.value } as any)
      toast.success('Updated')
    } else {
      await api.POST('/api/admin/shipping', { body: form.value } as any)
      toast.success('Created')
    }
    resetForm()
    await fetchServices()
  } catch { toast.error('Failed to save') }
}

const remove = async (id: number) => {
  try {
    await api.DELETE('/api/admin/shipping/{id}' as any, { params: { path: { id } } } as any)
    toast.success('Deleted')
    await fetchServices()
  } catch { toast.error('Failed to delete') }
}

const formatVnd = (v: number) => new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(v)

onMounted(fetchServices)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold">Shipping Services</h1>
        <p class="text-sm text-muted-foreground">Manage shipping providers for long-distance transactions.</p>
      </div>
    </div>

    <!-- Form -->
    <Card>
      <CardHeader><CardTitle class="text-base">{{ editing ? 'Edit' : 'Add' }} Shipping Service</CardTitle></CardHeader>
      <CardContent class="space-y-3">
        <div class="grid gap-3 sm:grid-cols-2">
          <div class="grid gap-1.5">
            <label class="text-xs font-medium">Name</label>
            <Input v-model="form.name" placeholder="Giao hàng nhanh" />
          </div>
          <div class="grid gap-1.5">
            <label class="text-xs font-medium">Estimated Days</label>
            <Input v-model="form.estimated_days" placeholder="1-3" />
          </div>
          <div class="grid gap-1.5">
            <label class="text-xs font-medium">Base Price (VND)</label>
            <Input v-model.number="form.base_price" type="number" min="0" />
          </div>
          <div class="grid gap-1.5">
            <label class="text-xs font-medium">Price per km (VND)</label>
            <Input v-model.number="form.price_per_km" type="number" min="0" />
          </div>
          <div class="grid gap-1.5">
            <label class="text-xs font-medium">Min Distance (km)</label>
            <Input v-model.number="form.min_distance" type="number" min="0" />
          </div>
          <div class="grid gap-1.5">
            <label class="text-xs font-medium">Description</label>
            <Input v-model="form.description" placeholder="Brief description" />
          </div>
        </div>
        <div class="flex gap-2">
          <Button size="sm" @click="save" class="gap-1">
            <Plus v-if="!editing" class="h-3.5 w-3.5" />
            <Pencil v-else class="h-3.5 w-3.5" />
            {{ editing ? 'Update' : 'Add' }}
          </Button>
          <Button v-if="editing" size="sm" variant="outline" @click="resetForm">Cancel</Button>
        </div>
      </CardContent>
    </Card>

    <!-- List -->
    <div v-if="isLoading" class="text-sm text-muted-foreground">Loading...</div>
    <div v-else class="space-y-2">
      <Card v-for="s in services" :key="s.id">
        <CardContent class="flex items-center justify-between p-4">
          <div class="flex items-center gap-3">
            <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-amber-100 text-amber-600">
              <Truck class="h-5 w-5" />
            </div>
            <div>
              <p class="text-sm font-semibold">{{ s.name }}</p>
              <p class="text-xs text-muted-foreground">
                Base {{ formatVnd(s.base_price) }} + {{ formatVnd(s.price_per_km) }}/km
                · Min {{ s.min_distance }}km · {{ s.estimated_days }} days
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <Badge :variant="s.is_active ? 'default' : 'secondary'">{{ s.is_active ? 'Active' : 'Inactive' }}</Badge>
            <Button size="sm" variant="ghost" @click="startEdit(s)"><Pencil class="h-3.5 w-3.5" /></Button>
            <Button size="sm" variant="ghost" class="text-destructive" @click="remove(s.id)"><Trash2 class="h-3.5 w-3.5" /></Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
