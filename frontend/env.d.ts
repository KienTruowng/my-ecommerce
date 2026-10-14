/// <reference types="vite/client" />

// Augment vue-router to include custom route meta fields
declare module 'vue-router' {
  interface RouteMeta {
    layout?: string
    requiresAuth?: boolean
    requiresAdmin?: boolean
  }
}
