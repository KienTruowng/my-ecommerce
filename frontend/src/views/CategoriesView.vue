<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '@/api/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { toast } from 'vue-sonner'
import { Tags, Pencil, Trash2, Plus, Layers, Loader2 } from 'lucide-vue-next'
type Category = { id: number; name: string }

const categories = ref<Category[]>([])
const newCategory = ref('')
const editingId = ref<number | null>(null)
const editingName = ref('')
const isLoading = ref(false)
const page = ref(1)
const limit = ref(50)
const total = ref(0)

const categoryIcons = ['📱', '🏠', '📚', '🪑', '🚲', '🎮', '👕', '💄', '🍳', '🎸', '🧸', '💻']
const getCategoryIcon = (id: number) => categoryIcons[id % categoryIcons.length]
const getCategoryColor = (id: number) => {
  const colors = ['bg-rose-100 text-rose-600', 'bg-sky-100 text-sky-600', 'bg-emerald-100 text-emerald-600', 'bg-amber-100 text-amber-600', 'bg-violet-100 text-violet-600', 'bg-cyan-100 text-cyan-600']
  return colors[id % colors.length]
}

const fetchCategories = async () => {
  isLoading.value = true
  try {
    const { data, error } = await api.GET('/api/categories', {
      params: { query: { page: page.value, limit: limit.value } },
    })
    if (error) { toast.error('Unable to load categories'); return }
    categories.value = (data as any)?.data ?? []
    total.value = (data as any)?.total ?? 0
  } finally { isLoading.value = false }
}

const createCategory = async () => {
  if (!newCategory.value.trim()) return
  const { error } = await api.POST('/api/categories', { body: { name: newCategory.value.trim() } })
  if (error) { toast.error('Unable to create category'); return }
  toast.success('Category created')
  newCategory.value = ''
  await fetchCategories()
}

const startEdit = (category: Category) => { editingId.value = category.id; editingName.value = category.name }
const cancelEdit = () => { editingId.value = null; editingName.value = '' }

const saveEdit = async () => {
  if (!editingId.value) return
  const { error } = await api.PATCH('/api/categories/{id}', {
    params: { path: { id: editingId.value } }, body: { name: editingName.value.trim() },
  })
  if (error) { toast.error('Unable to update category'); return }
  toast.success('Category updated')
  cancelEdit()
  await fetchCategories()
}

const deleteCategory = async (id: number) => {
  const { error } = await api.DELETE('/api/categories/{id}', { params: { path: { id } } })
  if (error) { toast.error('Unable to delete category'); return }
  toast.success('Category deleted')
  categories.value = categories.value.filter((cat) => cat.id !== id)
}

onMounted(fetchCategories)
</script>

<template>
  <div class="space-y-6">
    <!-- Banner -->
    <section class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-cyan-50 via-white to-violet-50 p-6">
      <div class="pointer-events-none absolute -right-8 -top-6 h-28 w-28 rounded-full bg-cyan-200/50" />
      <div class="pointer-events-none absolute -bottom-10 -left-6 h-36 w-36 rounded-full bg-violet-200/40" />
      <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-cyan-600">Categories</p>
          <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">Category Management</h1>
          <p class="mt-2 max-w-xl text-sm text-slate-600">Organize marketplace items into clear categories for easy browsing.</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <Input v-model="newCategory" placeholder="New category name..." class="w-52" @keyup.enter="createCategory" />
          <Button @click="createCategory" class="gap-1.5">
            <Plus class="h-4 w-4" /> Create
          </Button>
        </div>
      </div>
    </section>

    <!-- Stats -->
    <div class="flex items-center gap-2 text-sm">
      <Layers class="h-4 w-4 text-cyan-500" />
      <span class="font-semibold">{{ categories.length }}</span>
      <span class="text-muted-foreground">categories total</span>
    </div>

    <!-- Loading -->
    <div v-if="isLoading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card v-for="n in 6" :key="n" class="animate-pulse"><CardContent class="p-6 space-y-3"><div class="h-4 bg-muted rounded w-2/3"/><div class="h-3 bg-muted rounded w-1/3"/></CardContent></Card>
    </div>

    <!-- Grid -->
    <section v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card v-for="category in categories" :key="category.id" class="group bg-card/80 overflow-hidden transition-all hover:shadow-md hover:-translate-y-0.5">
        <CardHeader class="pb-3">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl" :class="getCategoryColor(category.id)">
              <span class="text-lg">{{ getCategoryIcon(category.id) }}</span>
            </div>
            <div class="min-w-0">
              <CardTitle v-if="editingId !== category.id" class="text-base truncate">{{ category.name }}</CardTitle>
              <Input v-else v-model="editingName" class="h-8 text-sm" @keyup.enter="saveEdit" />
            </div>
          </div>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="flex items-center gap-2 text-xs text-muted-foreground">
            <Tags class="h-3 w-3" /> ID: {{ category.id }}
          </div>
          <!-- Actions -->
          <div v-if="editingId === category.id" class="flex gap-2">
            <Button size="sm" @click="saveEdit" class="gap-1"><Loader2 class="h-3 w-3"/>Save</Button>
            <Button size="sm" variant="outline" @click="cancelEdit">Cancel</Button>
          </div>
          <div v-else class="flex gap-2">
            <Button size="sm" variant="outline" class="gap-1.5" @click="startEdit(category)">
              <Pencil class="h-3 w-3" /> Edit
            </Button>
            <Button size="sm" variant="ghost" class="gap-1.5 text-muted-foreground hover:text-destructive" @click="deleteCategory(category.id)">
              <Trash2 class="h-3 w-3" /> Delete
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- Empty -->
      <Card v-if="!categories.length" class="bg-card/80 py-10 col-span-full">
        <CardContent class="text-center">
          <Tags class="mx-auto h-10 w-10 text-muted-foreground/30" />
          <p class="mt-2 font-semibold text-muted-foreground">No categories yet</p>
          <p class="text-xs text-muted-foreground mt-1">Create your first category above.</p>
        </CardContent>
      </Card>
    </section>
  </div>
</template>
