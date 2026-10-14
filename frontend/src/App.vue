<script setup lang="ts">
import 'vue-sonner/style.css'
import { Toaster } from '@/components/ui/sonner'
import { useRoute } from 'vue-router'
import { computed, onMounted, type Component } from 'vue'
import DefaultLayout from './layouts/DefaultLayout.vue'
import AuthLayout from './layouts/AuthLayout.vue'
import { useAuthStore } from './stores/auth'
import { useUserStore } from './stores/user'
import EmptyLayout from './layouts/EmptyLayout.vue'

const route = useRoute()

const layouts: Record<string, Component> = {
  default: DefaultLayout,
  auth: AuthLayout,
  empty: EmptyLayout,
}

const currentLayout = computed(() => {
  const layoutName = (route.meta.layout as string) || 'default'
  return layouts[layoutName]
})

const userStore = useUserStore()

onMounted(() => {
  userStore.fetchProfile()
})
</script>

<template>
  <component :is="currentLayout">
    <RouterView />
  </component>
  <Toaster />
</template>

<style scoped></style>
