<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { X, ChevronLeft, ChevronRight } from 'lucide-vue-next'

const props = defineProps<{
  images: { src: string; alt?: string }[]
  modelValue: boolean
  startIndex?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const currentIdx = ref(0)

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      currentIdx.value = props.startIndex ?? 0
    }
  },
)

const currentImage = computed(() => props.images[currentIdx.value] ?? null)
const hasMultiple = computed(() => props.images.length > 1)
const show = computed(() => props.modelValue && props.images.length > 0)

const goNext = () => {
  if (currentIdx.value < props.images.length - 1) {
    currentIdx.value++
  }
}

const goPrev = () => {
  if (currentIdx.value > 0) {
    currentIdx.value--
  }
}

const goTo = (idx: number) => {
  currentIdx.value = idx
}

const close = () => emit('update:modelValue', false)

// Keyboard navigation
const handleKeydown = (e: KeyboardEvent) => {
  if (!props.modelValue) return
  if (e.key === 'Escape') close()
  if (e.key === 'ArrowRight') goNext()
  if (e.key === 'ArrowLeft') goPrev()
}

onMounted(() => document.addEventListener('keydown', handleKeydown))
onUnmounted(() => document.removeEventListener('keydown', handleKeydown))

// Prevent body scroll when open
watch(
  () => props.modelValue,
  (v) => {
    document.body.style.overflow = v ? 'hidden' : ''
  },
)
</script>

<template>
  <Teleport to="body">
    <Transition name="lightbox-fade">
      <div
        v-if="show"
        class="fixed inset-0 z-[100] flex items-center justify-center bg-black/92 backdrop-blur-sm"
        @click.self="close"
      >
        <!-- Close button -->
        <button
          @click="close"
          class="absolute right-4 top-4 z-10 flex h-10 w-10 items-center justify-center rounded-full bg-white/10 text-white backdrop-blur transition-colors hover:bg-white/20"
          aria-label="Close"
        >
          <X class="h-6 w-6" />
        </button>

        <!-- Counter -->
        <div
          class="absolute left-4 top-4 z-10 rounded-full bg-black/50 px-3 py-1.5 text-sm text-white backdrop-blur"
        >
          {{ currentIdx + 1 }} / {{ images.length }}
        </div>

        <!-- Previous -->
        <button
          v-if="currentIdx > 0"
          @click.stop="goPrev"
          class="absolute left-4 top-1/2 z-10 flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white backdrop-blur transition-colors hover:bg-white/20"
          aria-label="Previous image"
        >
          <ChevronLeft class="h-7 w-7" />
        </button>

        <!-- Main image -->
        <img
          :src="currentImage.src"
          :alt="currentImage.alt ?? 'Full size'"
          class="max-h-[90vh] max-w-[90vw] rounded-lg object-contain shadow-2xl select-none"
          @click.stop
        />

        <!-- Next -->
        <button
          v-if="currentIdx < images.length - 1"
          @click.stop="goNext"
          class="absolute right-4 top-1/2 z-10 flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white backdrop-blur transition-colors hover:bg-white/20"
          aria-label="Next image"
        >
          <ChevronRight class="h-7 w-7" />
        </button>

        <!-- Bottom thumbnail strip -->
        <div
          v-if="hasMultiple"
          class="absolute bottom-4 left-1/2 z-10 flex -translate-x-1/2 gap-2 overflow-x-auto rounded-xl bg-black/50 px-4 py-2.5 backdrop-blur max-w-[90vw]"
        >
          <button
            v-for="(img, idx) in images"
            :key="idx"
            @click="goTo(idx)"
            class="h-14 w-14 shrink-0 overflow-hidden rounded-lg border-2 transition-all"
            :class="
              idx === currentIdx
                ? 'border-white opacity-100 ring-2 ring-white/30'
                : 'border-transparent opacity-50 hover:opacity-80'
            "
          >
            <img :src="img.src" :alt="img.alt ?? `Thumb ${idx + 1}`" class="h-full w-full object-cover" />
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.lightbox-fade-enter-active,
.lightbox-fade-leave-active {
  transition: opacity 0.25s ease;
}
.lightbox-fade-enter-from,
.lightbox-fade-leave-to {
  opacity: 0;
}
</style>
