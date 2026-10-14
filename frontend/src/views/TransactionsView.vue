<script setup lang="ts">
import { computed, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  ArrowDownRight,
  ArrowUpRight,
  Filter,
  Landmark,
  Receipt,
  Wallet,
} from 'lucide-vue-next'

type TransactionStatus = 'completed' | 'pending' | 'failed'

type Transaction = {
  id: string
  title: string
  counterparty: string
  amount: number
  status: TransactionStatus
  type: 'credit' | 'debit'
  date: string
  method: string
}

const search = ref('')
const statusFilter = ref<'all' | TransactionStatus>('all')

const transactions = ref<Transaction[]>([
  {
    id: 'TX-2401',
    title: 'Barter top-up (Wallet placeholder)',
    counterparty: 'Campus Wallet',
    amount: 150000,
    status: 'completed',
    type: 'credit',
    date: '2026-05-10 14:20',
    method: 'Internal credit',
  },
  {
    id: 'TX-2398',
    title: 'Offer deposit - Dorm A Lobby',
    counterparty: 'Nguyen Thi Mai',
    amount: 50000,
    status: 'pending',
    type: 'debit',
    date: '2026-05-10 10:45',
    method: 'QR placeholder',
  },
  {
    id: 'TX-2392',
    title: 'Auction settlement refund',
    counterparty: 'Campus Auction',
    amount: 120000,
    status: 'completed',
    type: 'credit',
    date: '2026-05-09 19:18',
    method: 'Auto reversal',
  },
  {
    id: 'TX-2389',
    title: 'Marketplace payout',
    counterparty: 'Canteen West Gate',
    amount: 230000,
    status: 'failed',
    type: 'credit',
    date: '2026-05-09 08:30',
    method: 'Wallet pending',
  },
  {
    id: 'TX-2385',
    title: 'Item exchange fee',
    counterparty: 'KIEN STORE',
    amount: 20000,
    status: 'completed',
    type: 'debit',
    date: '2026-05-08 16:05',
    method: 'Platform fee',
  },
])

const summary = computed(() => {
  const completed = transactions.value.filter((t) => t.status === 'completed')
  const pending = transactions.value.filter((t) => t.status === 'pending')
  const failed = transactions.value.filter((t) => t.status === 'failed')

  return {
    completed: completed.length,
    pending: pending.length,
    failed: failed.length,
    volume: completed.reduce((total, t) => total + t.amount, 0),
  }
})

const filteredTransactions = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  return transactions.value.filter((item) => {
    const matchesStatus = statusFilter.value === 'all' || item.status === statusFilter.value
    if (!matchesStatus) return false
    if (!keyword) return true
    return (
      item.title.toLowerCase().includes(keyword) ||
      item.counterparty.toLowerCase().includes(keyword) ||
      item.id.toLowerCase().includes(keyword)
    )
  })
})

const formatAmount = (amount: number) =>
  new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount)

const statusBadge = (status: TransactionStatus) => {
  if (status === 'completed') return 'default'
  if (status === 'pending') return 'secondary'
  return 'destructive'
}

const statusLabel = (status: TransactionStatus) => {
  if (status === 'completed') return 'Completed'
  if (status === 'pending') return 'Pending'
  return 'Failed'
}
</script>

<template>
  <div class="space-y-6">
    <section
      class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-sky-50 via-white to-emerald-50 p-6"
    >
      <div class="pointer-events-none absolute -right-6 -top-10 h-32 w-32 rounded-full bg-sky-200/50" />
      <div class="pointer-events-none absolute -bottom-10 -left-8 h-36 w-36 rounded-full bg-emerald-200/40" />
      <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-sky-600">Transactions</p>
          <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">
            Wallet & transaction snapshots
          </h1>
          <p class="mt-2 max-w-xl text-sm text-slate-600">
            Mock UI for payment flows. The wallet system is placeholder-only, but this view
            illustrates balances, deposits, and settlement trails.
          </p>
        </div>
        <div class="flex flex-wrap gap-2">
          <Button variant="outline" class="gap-2">
            <Filter class="h-4 w-4" />
            Export CSV
          </Button>
          <Button class="gap-2">
            <Wallet class="h-4 w-4" />
            Add funds
          </Button>
        </div>
      </div>
    </section>

    <section class="grid gap-4 md:grid-cols-3">
      <Card class="bg-card/80">
        <CardHeader>
          <CardTitle class="text-lg">Wallet volume</CardTitle>
          <CardDescription>Completed transactions only.</CardDescription>
        </CardHeader>
        <CardContent class="flex items-center justify-between">
          <div>
            <p class="text-2xl font-semibold text-foreground">{{ formatAmount(summary.volume) }}</p>
            <p class="text-xs text-muted-foreground">Mock data · VND</p>
          </div>
          <Landmark class="h-9 w-9 text-sky-500" />
        </CardContent>
      </Card>
      <Card class="bg-card/80">
        <CardHeader>
          <CardTitle class="text-lg">Pending holds</CardTitle>
          <CardDescription>Deposits waiting for settlement.</CardDescription>
        </CardHeader>
        <CardContent class="flex items-center justify-between">
          <div>
            <p class="text-2xl font-semibold text-foreground">{{ summary.pending }}</p>
            <p class="text-xs text-muted-foreground">Needs confirmation</p>
          </div>
          <Receipt class="h-9 w-9 text-amber-500" />
        </CardContent>
      </Card>
      <Card class="bg-card/80">
        <CardHeader>
          <CardTitle class="text-lg">Completed payouts</CardTitle>
          <CardDescription>Approved trades & refunds.</CardDescription>
        </CardHeader>
        <CardContent class="flex items-center justify-between">
          <div>
            <p class="text-2xl font-semibold text-foreground">{{ summary.completed }}</p>
            <p class="text-xs text-muted-foreground">Settlement records</p>
          </div>
          <Wallet class="h-9 w-9 text-emerald-500" />
        </CardContent>
      </Card>
    </section>

    <section class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex flex-1 flex-wrap items-center gap-2">
        <Input v-model="search" placeholder="Search by ID, title, or name" class="max-w-xs" />
        <div class="flex flex-wrap gap-2">
          <Button
            size="sm"
            :variant="statusFilter === 'all' ? 'default' : 'outline'"
            @click="statusFilter = 'all'"
          >
            All
          </Button>
          <Button
            size="sm"
            :variant="statusFilter === 'completed' ? 'default' : 'outline'"
            @click="statusFilter = 'completed'"
          >
            Completed
          </Button>
          <Button
            size="sm"
            :variant="statusFilter === 'pending' ? 'default' : 'outline'"
            @click="statusFilter = 'pending'"
          >
            Pending
          </Button>
          <Button
            size="sm"
            :variant="statusFilter === 'failed' ? 'default' : 'outline'"
            @click="statusFilter = 'failed'"
          >
            Failed
          </Button>
        </div>
      </div>
      <div class="text-xs text-muted-foreground">Showing {{ filteredTransactions.length }} results</div>
    </section>

    <section class="space-y-4">
      <Card v-for="item in filteredTransactions" :key="item.id">
        <CardContent class="grid gap-4 p-5 sm:grid-cols-[1.2fr_auto] sm:items-center">
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <Badge :variant="statusBadge(item.status)">{{ statusLabel(item.status) }}</Badge>
              <span class="text-xs text-muted-foreground">{{ item.id }}</span>
            </div>
            <div>
              <p class="text-base font-semibold text-foreground">{{ item.title }}</p>
              <p class="text-sm text-muted-foreground">{{ item.counterparty }}</p>
            </div>
            <div class="flex flex-wrap gap-3 text-xs text-muted-foreground">
              <span>{{ item.date }}</span>
              <span>Method: {{ item.method }}</span>
            </div>
          </div>
          <div class="flex flex-col items-end gap-2 text-right">
            <div class="flex items-center gap-2 text-lg font-semibold">
              <ArrowUpRight v-if="item.type === 'credit'" class="h-4 w-4 text-emerald-500" />
              <ArrowDownRight v-else class="h-4 w-4 text-rose-500" />
              <span>{{ formatAmount(item.amount) }}</span>
            </div>
            <Button size="sm" variant="outline">View receipt</Button>
          </div>
        </CardContent>
      </Card>

      <Card v-if="!filteredTransactions.length" class="bg-card/80">
        <CardHeader>
          <CardTitle>No transactions found</CardTitle>
          <CardDescription>Try changing the filter or search keyword.</CardDescription>
        </CardHeader>
      </Card>
    </section>
  </div>
</template>
