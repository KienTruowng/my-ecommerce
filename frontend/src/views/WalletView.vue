<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/api'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const router = useRouter()

const summary = ref<{ balance?: number; currency?: string; status?: string } | null>(null)
const transactions = ref<Array<any>>([])
const amount = ref<number | null>(null)
const loading = ref(false)
const code = ref<string | null>(null)

async function fetchSummary() {
  try {
    const { data, error } = await api.GET('/api/wallet')
    if (error) {
      console.error(error)
      return
    }
    summary.value = data ?? null
  } catch (err) {
    console.error(err)
  }
}

async function fetchTransactions() {
  try {
    const { data, error } = await api.GET('/api/wallet/transactions', { query: { limit: 20 } })
    if (error) {
      console.error(error)
      return
    }
    transactions.value = data?.transactions ?? []
  } catch (err) {
    console.error(err)
  }
}

async function createTopupRequest() {
  if (!amount.value || amount.value <= 0) return toast('Nhập số tiền hợp lệ', { type: 'error' })
  loading.value = true
  try {
    const { data, error } = await api.POST('/api/wallet/topup/request', { body: { amount: amount.value } })
    if (error) {
      toast('Tạo yêu cầu nạp thất bại', { type: 'error' })
      console.error(error)
      return
    }
    code.value = data?.code ?? null
    if (data?.code) {
      const query = new URLSearchParams({
        amount: String(amount.value),
        expires_at: (data as any)?.expires_at ?? new Date(Date.now() + 86400000).toISOString(),
      })
      router.push(`/wallet/topup/${data.code}?${query}`)
    } else {
      toast('Yêu cầu nạp đã tạo — copy mã và dùng SePay test', { type: 'success' })
    }
    await fetchTransactions()
  } catch (err) {
    console.error(err)
    toast('Lỗi khi tạo yêu cầu nạp', { type: 'error' })
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  if (!userStore.profile) {
    await userStore.fetchProfile()
  }
  await fetchSummary()
  await fetchTransactions()
})
</script>

<template>
  <div class="min-h-screen px-6 py-8 bg-gradient-to-b from-indigo-50 via-white to-rose-50">
    <div class="max-w-6xl mx-auto">
      <div class="mb-6">
        <h1 class="text-4xl font-extrabold tracking-tight text-slate-800">Ví của tôi</h1>
        <p class="text-sm text-muted-foreground mt-1">Quản lý số dư, lịch sử giao dịch và nạp tiền qua SePay (test mode).</p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="md:col-span-1 bg-white/80 backdrop-blur rounded-2xl p-6 shadow-lg">
          <div class="text-sm text-muted-foreground">Số dư khả dụng</div>
          <div class="mt-4 flex items-baseline gap-3">
            <div class="text-3xl font-bold text-emerald-600">{{ summary?.balance ?? 0 }}</div>
            <div class="text-sm text-muted-foreground">{{ summary?.currency ?? 'VND' }}</div>
          </div>
          <div class="mt-4 text-xs text-muted-foreground">Trạng thái: {{ summary?.status ?? 'active' }}</div>

          <div class="mt-6">
            <label class="block text-xs text-muted-foreground mb-2">Số tiền nạp</label>
            <input v-model.number="amount" type="number" min="1" class="w-full px-3 py-2 rounded-md border" placeholder="Số tiền (ví dụ: 10000)" />
            <div class="mt-3 flex gap-2">
              <Button :class="'w-full'" :loading="loading" @click="createTopupRequest">Tạo mã nạp</Button>
            </div>
            <div v-if="code" class="mt-4 p-3 bg-amber-50 border-l-4 border-amber-400 rounded"> 
              <div class="text-xs text-muted-foreground">Mã nạp (copy và dùng SePay test - chuyển nội dung bằng mã này)</div>
              <div class="mt-2 font-mono text-sm break-all">{{ code }}</div>
            </div>
          </div>
        </div>

        <div class="md:col-span-2 bg-white/85 backdrop-blur rounded-2xl p-6 shadow-lg">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-semibold">Lịch sử giao dịch</h2>
            <Button variant="ghost" @click="fetchTransactions">Làm mới</Button>
          </div>

          <div class="divide-y">
            <div v-if="transactions.length === 0" class="py-8 text-center text-muted-foreground">Không có giao dịch nào</div>
            <div v-for="tx in transactions" :key="tx.id" class="py-3 flex items-start justify-between">
              <div>
                <div class="font-medium">{{ tx.type }} — {{ tx.direction }}</div>
                <div class="text-sm text-muted-foreground">{{ tx.metadata?.note || tx.external_id || tx.id }}</div>
              </div>
              <div class="text-right">
                <div :class="['font-semibold', tx.direction === 'in' ? 'text-emerald-600' : 'text-rose-600']">{{ tx.amount }}</div>
                <div class="text-xs text-muted-foreground">{{ new Date(tx.created_at).toLocaleString() }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.font-mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, 'Roboto Mono', 'Courier New', monospace; }
</style>
