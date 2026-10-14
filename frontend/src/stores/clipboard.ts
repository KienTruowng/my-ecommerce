import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export type CartItem = {
  id: number
  name: string
  description: string
  type: string
  status: string
  owner_id: number
  imageUrl?: string
  // User's personal notes
  note: string
  meetupLocation: string
  offerItem: string // what you plan to trade
  addedAt: number // timestamp
}

export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>(loadFromStorage())
  const isOpen = ref(false)

  const count = computed(() => items.value.length)
  const isEmpty = computed(() => items.value.length === 0)

  function loadFromStorage(): CartItem[] {
    try {
      const raw = localStorage.getItem('cart_items')
      return raw ? JSON.parse(raw) : []
    } catch {
      return []
    }
  }

  function saveToStorage() {
    localStorage.setItem('cart_items', JSON.stringify(items.value))
  }

  function isInCart(itemId: number): boolean {
    return items.value.some((i) => i.id === itemId)
  }

  function addItem(item: {
    id: number
    name: string
    description: string
    type: string
    status: string
    owner_id: number
    imageUrl?: string
  }) {
    if (isInCart(item.id)) return
    items.value.push({
      ...item,
      note: '',
      meetupLocation: '',
      offerItem: '',
      addedAt: Date.now(),
    })
    saveToStorage()
  }

  function removeItem(itemId: number) {
    items.value = items.value.filter((i) => i.id !== itemId)
    saveToStorage()
  }

  function updateNote(itemId: number, note: string) {
    const item = items.value.find((i) => i.id === itemId)
    if (item) {
      item.note = note
      saveToStorage()
    }
  }

  function updateMeetupLocation(itemId: number, location: string) {
    const item = items.value.find((i) => i.id === itemId)
    if (item) {
      item.meetupLocation = location
      saveToStorage()
    }
  }

  function updateOfferItem(itemId: number, offerItem: string) {
    const item = items.value.find((i) => i.id === itemId)
    if (item) {
      item.offerItem = offerItem
      saveToStorage()
    }
  }

  function toggleItem(item: {
    id: number
    name: string
    description: string
    type: string
    status: string
    owner_id: number
    imageUrl?: string
  }) {
    if (isInCart(item.id)) {
      removeItem(item.id)
    } else {
      addItem(item)
    }
  }

  function clearAll() {
    items.value = []
    saveToStorage()
  }

  function toggleOpen() {
    isOpen.value = !isOpen.value
  }

  return {
    items,
    count,
    isEmpty,
    isOpen,
    isInCart,
    addItem,
    removeItem,
    updateNote,
    updateMeetupLocation,
    updateOfferItem,
    toggleItem,
    clearAll,
    toggleOpen,
  }
})
