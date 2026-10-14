<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { toast } from 'vue-sonner'
import {
  QrCode,
  Copy,
  Clock,
  Banknote,
  ArrowLeft,
  ShieldCheck,
  Loader2,
  CheckCircle2,
  AlertCircle,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()

const code = ref((route.params.code as string) || '')
const amount = ref<number>(0)
const expiresAt = ref<string>('')
const isLoading = ref(true)
const timeLeft = ref('')
const isExpired = ref(false)
const isCopied = ref(false)
const isVerifying = ref(false)
const isPaid = ref(false)
const initialBalance = ref(0)

const verifyPayment = async () => {
  isVerifying.value = true
  try {
    const token = localStorage.getItem('access_token') ?? ''
    const res = await fetch('/api/wallet', {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (res.ok) {
      const json = await res.json()
      const currentBalance = json?.balance ?? 0
      const increased = currentBalance > initialBalance.value
      if (increased) {
        isPaid.value = true
        toast.success('Thanh toán đã được ghi nhận!')
      } else {
        toast('Chưa nhận được thanh toán. Vui lòng thử lại sau vài phút.', {
          duration: 5000,
        })
      }
    }
  } catch {
    toast.error('Không thể kiểm tra trạng thái thanh toán')
  } finally {
    isVerifying.value = false
  }
}

const startTimer = () => {
  const updateTimeLeft = () => {
    if (!expiresAt.value) return
    const expiry = new Date(expiresAt.value).getTime()
    const now = Date.now()
    const diff = expiry - now

    if (diff <= 0) {
      timeLeft.value = 'Hết hạn'
      isExpired.value = true
      if (timer) clearInterval(timer)
      return
    }

    const hours = Math.floor(diff / 3600000)
    const minutes = Math.floor((diff % 3600000) / 60000)
    const seconds = Math.floor((diff % 60000) / 1000)
    timeLeft.value = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  }

  updateTimeLeft()
  timer = setInterval(updateTimeLeft, 1000)
}

const copyCode = async () => {
  try {
    await navigator.clipboard.writeText(code.value)
    isCopied.value = true
    toast.success('Đã sao chép mã nạp')
    setTimeout(() => { isCopied.value = false }, 2000)
  } catch {
    toast.error('Không thể sao chép')
  }
}

const copyAccountNumber = async () => {
  try {
    await navigator.clipboard.writeText(accountNumber)
    toast.success('Đã sao chép số tài khoản')
  } catch {
    toast.error('Không thể sao chép')
  }
}

onMounted(async () => {
  if (!code.value) {
    router.replace('/wallet')
    return
  }

  try {
    const token = localStorage.getItem('access_token') ?? ''
    const res = await fetch(`/api/wallet`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (res.ok) {
      const json = await res.json()
      initialBalance.value = json?.balance ?? 0
    }
  } catch {
    // non-critical
  }

  amount.value = Number(route.query.amount) || 0
  expiresAt.value = (route.query.expires_at as string) || new Date(Date.now() + 86400000).toISOString()
  isLoading.value = false
  startTimer()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="min-h-screen px-4 py-8">
    <div class="max-w-lg mx-auto space-y-6">
      <button
        class="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
        @click="router.push('/wallet')"
      >
        <ArrowLeft class="h-4 w-4" />
        Quay lại Ví
      </button>

      <section
        v-if="isPaid"
        class="rounded-2xl border-2 border-emerald-300 bg-emerald-50/60 p-8 text-center space-y-4"
      >
        <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-emerald-100">
          <CheckCircle2 class="h-8 w-8 text-emerald-600" />
        </div>
        <h2 class="text-xl font-bold text-emerald-700">Thanh toán thành công!</h2>
        <p class="text-sm text-emerald-600">
          Số dư của bạn đã được cập nhật. Hãy kiểm tra ví để xác nhận.
        </p>
        <Button class="mt-2" @click="router.push('/wallet')">
          Xem số dư
        </Button>
      </section>

      <template v-else>
        <div class="text-center space-y-2">
          <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-indigo-100">
            <QrCode class="h-7 w-7 text-indigo-600" />
          </div>
          <h1 class="text-2xl font-bold">Nạp tiền qua SePay</h1>
          <p class="text-sm text-muted-foreground">
            Quét mã QR hoặc chuyển khoản với mã nạp bên dưới
          </p>
        </div>

        <Card class="overflow-hidden">
          <div class="bg-white p-6 flex justify-center">
            <div v-if="isLoading" class="py-12">
              <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
            <img
              v-else
              :src="qrUrl"
              :alt="'QR Code for ' + code"
              class="h-64 w-64 rounded-xl border"
              @error="($event.target as HTMLImageElement).style.display = 'none'"
            />
          </div>
        </Card>

        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="flex items-center gap-2 text-base">
              <Banknote class="h-4 w-4 text-indigo-500" />
              Thông tin chuyển khoản
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex justify-between items-center py-1.5 border-b">
              <span class="text-sm text-muted-foreground">Ngân hàng</span>
              <span class="text-sm font-semibold">{{ bankName }}</span>
            </div>
            <div class="flex justify-between items-center py-1.5 border-b">
              <span class="text-sm text-muted-foreground">Số tài khoản</span>
              <button
                class="flex items-center gap-1.5 text-sm font-mono font-semibold hover:text-indigo-600 transition-colors"
                @click="copyAccountNumber"
              >
                {{ accountNumber }}
                <Copy class="h-3 w-3" />
              </button>
            </div>
            <div class="flex justify-between items-center py-1.5 border-b">
              <span class="text-sm text-muted-foreground">Chủ tài khoản</span>
              <span class="text-sm font-semibold">{{ accountName }}</span>
            </div>
            <div class="flex justify-between items-center py-1.5 border-b">
              <span class="text-sm text-muted-foreground">Số tiền</span>
              <span class="text-sm font-bold text-indigo-600 font-mono">
                {{ new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount) }}
              </span>
            </div>
            <div class="flex justify-between items-center py-1.5">
              <span class="text-sm text-muted-foreground">Nội dung CK</span>
              <button
                class="flex items-center gap-1.5 text-sm font-mono font-bold text-indigo-600 hover:text-indigo-700 transition-colors"
                @click="copyCode"
              >
                {{ code }}
                <Copy v-if="!isCopied" class="h-3 w-3" />
                <CheckCircle2 v-else class="h-3 w-3 text-emerald-500" />
              </button>
            </div>
          </CardContent>
        </Card>

        <Card
          :class="isExpired ? 'border-red-200 bg-red-50/60' : 'border-amber-200 bg-amber-50/40'"
        >
          <CardContent class="flex items-center justify-between py-4">
            <div class="flex items-center gap-3">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-full"
                :class="isExpired ? 'bg-red-100' : 'bg-amber-100'"
              >
                <Clock
                  class="h-5 w-5"
                  :class="isExpired ? 'text-red-500' : 'text-amber-600'"
                />
              </div>
              <div>
                <p class="text-sm font-semibold" :class="isExpired ? 'text-red-600' : 'text-amber-700'">
                  {{ isExpired ? 'Đã hết hạn' : 'Thời gian còn lại' }}
                </p>
                <p class="text-xs text-muted-foreground">
                  {{ isExpired ? 'Vui lòng tạo mã nạp mới' : 'Hoàn tất chuyển khoản trước khi hết hạn' }}
                </p>
              </div>
            </div>
            <span
              class="text-xl font-bold font-mono tabular-nums"
              :class="isExpired ? 'text-red-600' : 'text-amber-700'"
            >
              {{ timeLeft || '--:--:--' }}
            </span>
          </CardContent>
        </Card>

        <div class="flex gap-3">
          <Button
            variant="outline"
            class="flex-1"
            :disabled="isVerifying || isExpired"
            @click="verifyPayment"
          >
            <Loader2 v-if="isVerifying" class="h-4 w-4 animate-spin mr-2" />
            <ShieldCheck v-else class="h-4 w-4 mr-2" />
            Kiểm tra thanh toán
          </Button>
          <Button
            class="flex-1"
            :disabled="isExpired"
            @click="router.push('/wallet')"
          >
            <Banknote class="h-4 w-4 mr-2" />
            Về ví
          </Button>
        </div>

        <Card>
          <CardContent class="py-3 space-y-2 text-xs text-muted-foreground">
            <div class="flex items-start gap-2">
              <AlertCircle class="h-3.5 w-3.5 text-amber-500 mt-0.5 shrink-0" />
              <span>Sau khi chuyển khoản, hệ thống sẽ tự động cập nhật số dư trong vài phút.</span>
            </div>
            <div class="flex items-start gap-2">
              <AlertCircle class="h-3.5 w-3.5 text-amber-500 mt-0.5 shrink-0" />
              <span>Vui lòng nhập chính xác <strong>Nội dung CK</strong> để hệ thống ghi nhận giao dịch.</span>
            </div>
          </CardContent>
        </Card>
      </template>
    </div>
  </div>
</template>

<style scoped>
.font-mono {
  font-family: 'JetBrains Mono', ui-monospace, monospace;
}
</style>
