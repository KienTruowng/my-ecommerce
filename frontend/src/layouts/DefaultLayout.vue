<script setup lang="ts">
import AppSidebar from '@/components/layout/AppSidebar.vue'
</script>

<template>
  <div class="flex h-screen overflow-hidden">
    <AppSidebar />
    <main class="flex-1 overflow-y-auto">
      <div class="mx-auto w-full max-w-6xl px-4 sm:px-6 lg:px-8 py-6 lg:py-8">
        <slot />
      </div>
      <footer class="border-t bg-background/80 backdrop-blur">
        <div
          class="mx-auto w-full max-w-6xl px-4 sm:px-6 lg:px-8 py-5 flex flex-col gap-2 md:flex-row md:items-center md:justify-between"
        >
          <div class="text-sm font-semibold">CampusTrade</div>
          <div class="text-xs text-muted-foreground">
            Built for student-to-student trading within your university.
          </div>
        </div>
      </footer>
    </main>
  </div>
</template>
