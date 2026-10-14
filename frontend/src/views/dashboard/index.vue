<script setup lang="ts">
import AppSidebar from '@/components/dashboard/AppSidebar.vue'
import SiteHeader from '@/components/dashboard/SiteHeader.vue'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { useMagicKeys } from '@vueuse/core'
</script>
<template>
  <SidebarProvider
    :style="{
      '--header-height': 'calc(var(--spacing) * 12)',
    }"
  >
    <AppSidebar />

    <SidebarInset>
      <SiteHeader />
      <RouterView />
    </SidebarInset>
  </SidebarProvider>
</template>
