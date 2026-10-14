import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { toast } from 'vue-sonner'

export type ServerCartItem = {
  item_id: number
  offer_item_id: number | null
  cash_amount: number
  item?: {
    id: number
    name: string
    description: string
    status: string
    type: string
    images?: { id: number; url: string }[]
  }
  trade_in_item?: {
    id: number
    name: string
    status: string
    images?: { id: number; url: string }[]
  } | null
}

function normalizeItem(item: any): ServerCartItem['item'] | undefined {
  if (!item) return undefined
  const images = (item?.ItemImage ?? item?.item_image ?? item?.images ?? item?.Images ?? [])
  return {
    id: item?.ID ?? item?.id ?? item?.Id ?? 0,
    name: item?.Name ?? item?.name ?? '',
    description: item?.Description ?? item?.description ?? '',
    status: item?.Status ?? item?.status ?? '',
    type: item?.Type ?? item?.type ?? '',
    images: Array.isArray(images)
      ? images.map((img: any) => ({
          id: img?.ID ?? img?.id ?? img?.Id ?? 0,
          url: img?.URL ?? img?.url ?? img?.Url ?? '',
        }))
      : [],
  }
}

function normalizeCart(raw: any): ServerCartItem | null {
  if (!raw) return null
  return {
    item_id: raw?.item_id ?? raw?.ItemID ?? raw?.itemId ?? raw?.ID ?? 0,
    offer_item_id: raw?.offer_item_id ?? raw?.TradeInItemID ?? raw?.trade_in_item_id ?? null,
    cash_amount: raw?.cash_amount ?? raw?.CashAmount ?? 0,
    item: normalizeItem(raw?.item ?? raw?.Item),
    trade_in_item: normalizeItem(raw?.trade_in_item ?? raw?.TradeInItem ?? raw?.tradeInItem),
  }
}

export const useCartStore = defineStore('cart', () => {
  const cart = ref<ServerCartItem | null>(null)
  const isLoading = ref(false)
  const isOpen = ref(false)

  const count = computed(() => (cart.value ? 1 : 0))
  const isEmpty = computed(() => cart.value === null)

  async function fetchCart() {
    isLoading.value = true
    try {
      const token = localStorage.getItem('access_token') ?? ''
      const res = await fetch('/api/cart', {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (res.ok) {
        const json = await res.json()
        const normalized = normalizeCart(json?.data ?? json)
        if (normalized !== null) {
          cart.value = normalized
        }
      }
    } catch {
      // keep existing cart data on network error
    } finally {
      isLoading.value = false
    }
  }

  async function addToCart(itemId: number, tradeInItemId?: number | null, cashAmount?: number) {
    isLoading.value = true
    try {
      const token = localStorage.getItem('access_token') ?? ''
      const body: Record<string, unknown> = { item_id: itemId }
      if (tradeInItemId !== undefined) body.offer_item_id = tradeInItemId
      if (cashAmount !== undefined) body.cash_amount = cashAmount
      const res = await fetch('/api/cart', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      })
      if (res.ok) {
        const json = await res.json()
        cart.value = normalizeCart(json?.data ?? json)
        toast.success('Added to cart')
      } else {
        const err = await res.json().catch(() => ({}))
        toast.error(err.message || err.error || 'Unable to add to cart')
      }
    } catch {
      toast.error('Unable to add to cart')
    } finally {
      isLoading.value = false
    }
  }

  async function updateCart(tradeInItemId: number | null, cashAmount: number) {
    isLoading.value = true
    try {
      const token = localStorage.getItem('access_token') ?? ''
      const res = await fetch('/api/cart', {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          offer_item_id: tradeInItemId,
          cash_amount: cashAmount,
        }),
      })
      if (res.ok) {
        const json = await res.json()
        cart.value = normalizeCart(json?.data ?? json)
      }
    } catch {
      // ignore
    } finally {
      isLoading.value = false
    }
  }

  async function clearCart() {
    try {
      const token = localStorage.getItem('access_token') ?? ''
      await fetch('/api/cart', {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      })
    } catch {
      // ignore
    }
    cart.value = null
  }

  async function checkout(paymentMethod: string) {
    try {
      const token = localStorage.getItem('access_token') ?? ''
      const res = await fetch('/api/checkout', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ payment_method: paymentMethod }),
      })
      const json = await res.json()
      if (!res.ok) {
        throw new Error(json.message || json.error || 'Checkout failed')
      }
      cart.value = null
      return json.data
    } catch (err: any) {
      throw err
    }
  }

  function isInCart(itemId: number): boolean {
    return cart.value?.item_id === itemId
  }

  function toggleOpen() {
    isOpen.value = !isOpen.value
  }

  return {
    cart,
    isLoading,
    isOpen,
    count,
    isEmpty,
    isInCart,
    fetchCart,
    addToCart,
    updateCart,
    clearCart,
    checkout,
    toggleOpen,
  }
})
