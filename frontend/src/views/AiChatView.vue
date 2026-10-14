<script setup lang="ts">
import { computed, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { toast } from 'vue-sonner'
import { Bot, Flame, Sparkles, Wand2, Send, Loader2 } from 'lucide-vue-next'

type ChatMessage = {
  id: string
  role: 'assistant' | 'user'
  text: string
  time: string
}

const prompt = ref('')
const isThinking = ref(false)
const modelStatus = ref('DeepSeek Chat')

const messages = ref<ChatMessage[]>([
  {
    id: 'm1',
    role: 'assistant',
    text: 'Hi! I\'m your CampusTrade AI assistant powered by DeepSeek. I can help you craft listings, negotiate trades, suggest fair prices, and recommend safe meetup spots. What can I help with?',
    time: '',
  },
])

const suggestions = [
  'Draft a listing for my used textbook',
  'Suggest a fair price for a used laptop',
  'Write a polite counter-offer message',
  'What are safe meetup spots on campus?',
]

const canSend = computed(() => prompt.value.trim().length > 0 && !isThinking.value)

const BASE = ''

const sendPrompt = async () => {
  if (!canSend.value) return
  const now = new Date()
  const time = now.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
  const userText = prompt.value.trim()

  messages.value.push({
    id: `user-${messages.value.length + 1}`,
    role: 'user', text: userText, time,
  })
  prompt.value = ''
  isThinking.value = true

  try {
    const token = localStorage.getItem('access_token') ?? ''
    const apiMessages = messages.value
      .filter(m => m.id !== 'm1') // skip initial greeting
      .map(m => ({ role: m.role === 'assistant' ? 'assistant' : 'user', content: m.text }))

    const res = await fetch(`${BASE}/api/ai/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ messages: apiMessages }),
    })

    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      throw new Error((err as any)?.error ?? `HTTP ${res.status}`)
    }

    const data = await res.json()
    messages.value.push({
      id: `assistant-${messages.value.length + 1}`,
      role: 'assistant', text: data.reply, time,
    })
  } catch (e: any) {
    messages.value.push({
      id: `assistant-${messages.value.length + 1}`,
      role: 'assistant',
      text: 'Sorry, I couldn\'t reach the AI service. Make sure DEEPSEEK_API_KEY is set in the backend .env file.',
      time,
    })
    toast.error('AI service unavailable')
  } finally {
    isThinking.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <section
      class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-amber-50 via-white to-sky-50 p-6"
    >
      <div class="pointer-events-none absolute -right-10 -top-10 h-32 w-32 rounded-full bg-amber-200/60" />
      <div class="pointer-events-none absolute -bottom-10 -left-8 h-36 w-36 rounded-full bg-sky-200/50" />
      <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-amber-600">AI Chat</p>
          <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">
            Campus AI assistant
          </h1>
          <p class="mt-2 max-w-xl text-sm text-slate-600">
            Placeholder
          </p>
        </div>
        <div class="flex flex-wrap gap-2">
          <Badge class="gap-1" variant="secondary">
            <Sparkles class="h-3 w-3" />
            {{ modelStatus }}
          </Badge>
          <Button variant="outline" class="gap-2">
            <Flame class="h-4 w-4" />
            Generate briefing
          </Button>
        </div>
      </div>
    </section>

    <section class="grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
      <Card class="bg-card/90">
        <CardHeader class="border-b">
          <CardTitle class="flex items-center gap-2 text-lg">
            <Bot class="h-5 w-5 text-amber-500" />
            Live conversation
          </CardTitle>
          <CardDescription>Mock messages only. No requests are sent.</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4 p-5">
          <div class="space-y-3">
            <div
              v-for="msg in messages"
              :key="msg.id"
              class="flex items-start gap-3"
              :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
            >
              <div
                class="max-w-[80%] rounded-2xl border px-4 py-3 text-sm"
                :class="
                  msg.role === 'user'
                    ? 'bg-slate-900 text-white border-slate-900'
                    : 'bg-white text-slate-700 border-slate-200'
                "
              >
                <p class="font-medium text-xs uppercase tracking-widest text-muted-foreground">
                  {{ msg.role === 'user' ? 'You' : 'AI' }} · {{ msg.time }}
                </p>
                <p class="mt-2 text-sm leading-relaxed">{{ msg.text }}</p>
              </div>
            </div>
          </div>

          <div class="space-y-3 border-t pt-4">
            <p class="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
              Quick prompts
            </p>
            <div class="flex flex-wrap gap-2">
              <Button
                v-for="item in suggestions"
                :key="item"
                size="sm"
                variant="outline"
                class="rounded-full"
                @click="prompt = item"
              >
                {{ item }}
              </Button>
            </div>
          </div>

          <div class="space-y-2">
            <Textarea
              v-model="prompt"
              placeholder="Ask for listing copy, negotiation tips, or a safety checklist"
            />
            <div class="flex items-center justify-between">
              <p class="text-xs text-muted-foreground">{{ isThinking ? 'AI is thinking...' : 'Powered by DeepSeek' }}</p>
              <Button :disabled="!canSend" class="gap-2" @click="sendPrompt">
                <Loader2 v-if="isThinking" class="h-4 w-4 animate-spin" />
                <Send v-else class="h-4 w-4" />
                {{ isThinking ? 'Thinking...' : 'Send' }}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <div class="space-y-4">
        <Card class="bg-card/90">
          <CardHeader>
            <CardTitle class="text-lg">Prompt templates</CardTitle>
            <CardDescription>Reusable scripts for common trade tasks.</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3 text-sm text-muted-foreground">
            <div class="flex items-start gap-2">
              <Wand2 class="mt-0.5 h-4 w-4 text-sky-500" />
              <p>Write listing descriptions in friendly, student-first tone.</p>
            </div>
            <div class="flex items-start gap-2">
              <Wand2 class="mt-0.5 h-4 w-4 text-sky-500" />
              <p>Generate safe meetup proposals using approved locations.</p>
            </div>
            <div class="flex items-start gap-2">
              <Wand2 class="mt-0.5 h-4 w-4 text-sky-500" />
              <p>Create polite counter-offers and decline messages.</p>
            </div>
          </CardContent>
        </Card>

      </div>
    </section>
  </div>
</template>
