<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { api } from '@/api/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import {
  Calendar,
  MapPin,
  Clock3,
  Users,
  Bell,
  BellRing,
  Sparkles,
  Navigation,
  CheckCircle2,
  XCircle,
  ArrowRight,
  ChevronDown,
  ChevronUp,
  Info,
  ExternalLink,
  Star,
  TrendingUp,
  ShieldCheck,
  AlertCircle,
  Loader2,
  Map,
  Route,
  Zap,
  Search,
} from 'lucide-vue-next'

// ─── Types ────────────────────────────────────────────────────────────────────
type AppointmentStatus = 'scheduled' | 'confirmed' | 'completed' | 'cancelled'

type Appointment = {
  id: string
  title: string
  date: string
  time: string
  location: string
  partner: string
  status: AppointmentStatus
  note: string
  order_id?: number
  offer_id?: number
  created_at?: string
}

type Location = {
  id: number
  name: string
  description: string
  address: string
  image_url: string
  latitude: number
  longitude: number
  open_time: string
  close_time: string
  is_active: boolean
}

type SuggestedLocation = {
  location: Location
  score: number
  reasons: string[]
  distance: string
}

type Notification = {
  id: string
  type: 'upcoming' | 'confirmation' | 'reminder' | 'update'
  title: string
  message: string
  appointmentId: string
  time: Date
  read: boolean
}

// ─── State ────────────────────────────────────────────────────────────────────
const search = ref('')
const isLoading = ref(true)
const isSaving = ref(false)
const appointments = ref<Appointment[]>([])
const locations = ref<Location[]>([])
const activeTab = ref('upcoming')

// Map & suggestion state
const selectedAppointment = ref<Appointment | null>(null)
const showMap = ref(false)
const suggestedLocations = ref<SuggestedLocation[]>([])
const isGeneratingSuggestion = ref(false)
const mapExpanded = ref(false)

// Notification state
const notifications = ref<Notification[]>([])
const showNotifications = ref(false)
const notificationPollInterval = ref<ReturnType<typeof setInterval> | null>(null)
const lastCheckedTime = ref<Date>(new Date())

// Edit dialog
const editingAppointment = ref<Appointment | null>(null)
const isDialogOpen = ref(false)
const form = ref({
  date: '',
  time: '',
  location: '',
  note: '',
})

// ─── Computed ──────────────────────────────────────────────────────────────────
const unreadCount = computed(() => notifications.value.filter((n) => !n.read).length)

const upcomingAppointments = computed(() =>
  appointments.value.filter(
    (a) => a.status === 'scheduled' || a.status === 'confirmed',
  ),
)

const pastAppointments = computed(() =>
  appointments.value.filter(
    (a) => a.status === 'completed' || a.status === 'cancelled',
  ),
)

const todayAppointments = computed(() => {
  const today = new Date().toISOString().split('T')[0]
  return upcomingAppointments.value.filter((a) => a.date === today)
})

const urgentAppointments = computed(() => {
  const now = new Date()
  return upcomingAppointments.value.filter((a) => {
    if (!a.date || !a.time) return false
    const apptTime = new Date(`${a.date}T${a.time}`)
    const diffMs = apptTime.getTime() - now.getTime()
    const diffHours = diffMs / (1000 * 60 * 60)
    return diffHours > 0 && diffHours <= 2
  })
})

const filteredAppointments = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  const source =
    activeTab.value === 'upcoming'
      ? upcomingAppointments.value
      : activeTab.value === 'past'
        ? pastAppointments.value
        : appointments.value

  if (!keyword) return source
  return source.filter(
    (item) =>
      (item.title && item.title.toLowerCase().includes(keyword)) ||
      (item.location && item.location.toLowerCase().includes(keyword)) ||
      (item.partner && item.partner.toLowerCase().includes(keyword)),
  )
})

const statsCards = computed(() => [
  {
    label: 'Upcoming Meets',
    value: upcomingAppointments.value.length,
    sub: `${todayAppointments.value.length} today`,
    icon: Calendar,
    color: 'text-rose-500',
    bg: 'bg-rose-50',
  },
  {
    label: 'Pending Confirm',
    value: appointments.value.filter((a) => a.status === 'scheduled').length,
    sub: 'Awaiting partner',
    icon: Clock3,
    color: 'text-amber-500',
    bg: 'bg-amber-50',
  },
  {
    label: 'Trusted Partners',
    value: new Set(appointments.value.map((a) => a.partner)).size,
    sub: 'Unique peers',
    icon: Users,
    color: 'text-emerald-500',
    bg: 'bg-emerald-50',
  },
  {
    label: 'Completed',
    value: appointments.value.filter((a) => a.status === 'completed').length,
    sub: 'Successful trades',
    icon: CheckCircle2,
    color: 'text-sky-500',
    bg: 'bg-sky-50',
  },
])

// ─── Status helpers ────────────────────────────────────────────────────────────
const statusVariant = (status: AppointmentStatus) => {
  if (status === 'confirmed') return 'default'
  if (status === 'scheduled') return 'secondary'
  if (status === 'completed') return 'outline'
  return 'destructive'
}

const statusLabel = (status: AppointmentStatus) => {
  if (status === 'scheduled') return 'Scheduled'
  if (status === 'confirmed') return 'Confirmed'
  if (status === 'completed') return 'Completed'
  return 'Cancelled'
}

const statusIcon = (status: AppointmentStatus) => {
  if (status === 'confirmed') return CheckCircle2
  if (status === 'scheduled') return Clock3
  if (status === 'completed') return CheckCircle2
  return XCircle
}

const isUrgent = (appt: Appointment) => {
  return urgentAppointments.value.some((u) => u.id === appt.id)
}

// ─── Data fetching ─────────────────────────────────────────────────────────────
const fetchAppointments = async () => {
  isLoading.value = true
  try {
    const { data, error } = await api.GET('/api/appointments' as any)
    if (error) {
      toast.error('Unable to load appointments')
      return
    }
    appointments.value = (data as any) || []
    generateNotifications()
  } finally {
    isLoading.value = false
  }
}

const fetchLocations = async () => {
  try {
    const { data, error } = await api.GET('/api/locations', {
      params: { query: { page: 1, limit: 50 } },
    })
    if (!error && data) {
      const list = (data as any)?.data ?? (data as any) ?? []
      locations.value = list.filter((l: Location) => l.is_active && l.latitude && l.longitude)
    }
  } catch {
    // Locations are non-critical; silently fail
  }
}

// ─── AI Suggestion Engine ──────────────────────────────────────────────────────
const generateSuggestions = (appt: Appointment) => {
  if (!locations.value.length) return []

  const now = new Date()
  const currentHour = now.getHours()

  const scored: SuggestedLocation[] = locations.value.map((loc) => {
    const reasons: string[] = []
    let score = 50 // base score

    // Factor 1: Name match with current location (familiarity)
    if (appt.location && loc.name.toLowerCase().includes(appt.location.toLowerCase())) {
      score += 25
      reasons.push('📍 Your usual spot')
    }

    // Factor 2: Open hours compatibility
    if (loc.open_time && loc.close_time) {
      const openH = parseInt(loc.open_time.split(':')[0])
      const closeH = parseInt(loc.close_time.split(':')[0])
      if (currentHour >= openH && currentHour <= closeH - 1) {
        score += 20
        reasons.push('🕐 Open now')
      } else if (currentHour >= openH - 1 && currentHour <= closeH) {
        score += 10
        reasons.push('🕐 Opens soon')
      }
    }

    // Factor 3: Has description (well-documented = safer)
    if (loc.description && loc.description.length > 10) {
      score += 10
      reasons.push('📋 Well-documented')
    }

    // Factor 4: Popular/common campus spots get bonus
    const popularSpots = ['library', 'canteen', 'hall', 'lobby', 'gate', 'dorm']
    if (popularSpots.some((s) => loc.name.toLowerCase().includes(s))) {
      score += 15
      reasons.push('⭐ Popular campus spot')
    }

    // Factor 5: Has address (more specific = better)
    if (loc.address) {
      score += 5
    }

    // Distance estimation (mock: use lat/lng to estimate, simplified)
    const distance = loc.latitude
      ? `${(Math.abs(loc.latitude - 10.8) * 111).toFixed(1)} km`
      : 'Unknown'

    return { location: loc, score, reasons, distance }
  })

  return scored
    .sort((a, b) => b.score - a.score)
    .slice(0, 4)
    .filter((s) => s.score >= 50)
}

const openMapForAppointment = (appt: Appointment) => {
  selectedAppointment.value = appt
  isGeneratingSuggestion.value = true
  showMap.value = true

  setTimeout(() => {
    // Try API locations first, always fall back
    const apiSuggestions = generateSuggestions(appt)
    if (apiSuggestions.length > 0) {
      suggestedLocations.value = apiSuggestions
    } else {
      suggestedLocations.value = getFallbackSuggestions(appt)
    }
    isGeneratingSuggestion.value = false
  }, 400)
}

const getFallbackSuggestions = (appt: Appointment): SuggestedLocation[] => {
  const fallbacks: SuggestedLocation[] = [
    {
      location: {
        id: 0,
        name: 'Library Hall',
        description: 'Quiet, well-lit, security cameras. Perfect for item inspection.',
        address: 'Main Campus, Building A',
        image_url: '',
        latitude: 10.8702,
        longitude: 106.8025,
        open_time: '07:00',
        close_time: '21:00',
        is_active: true,
      },
      score: 92,
      reasons: ['⭐ Popular campus spot', '🕐 Open now', '📋 Well-documented'],
      distance: '0.3 km',
    },
    {
      location: {
        id: 0,
        name: 'Canteen West Gate',
        description: 'Busy public area with seating. Great for quick exchanges.',
        address: 'West Gate Entrance',
        image_url: '',
        latitude: 10.8708,
        longitude: 106.8018,
        open_time: '06:30',
        close_time: '19:00',
        is_active: true,
      },
      score: 85,
      reasons: ['⭐ Popular campus spot', '🕐 Open now'],
      distance: '0.5 km',
    },
    {
      location: {
        id: 0,
        name: 'Dorm A Lobby',
        description: '24/7 access, front desk security. Safe for evening trades.',
        address: 'Student Dormitory A',
        image_url: '',
        latitude: 10.8695,
        longitude: 106.8032,
        open_time: '00:00',
        close_time: '23:59',
        is_active: true,
      },
      score: 78,
      reasons: ['🕐 Open now', '📋 Well-documented'],
      distance: '0.7 km',
    },
  ]

  return fallbacks.filter((s) => s.score >= 60)
}

const selectSuggestedLocation = (suggested: SuggestedLocation) => {
  if (!selectedAppointment.value) return
  form.value.location = suggested.location.name
  editingAppointment.value = selectedAppointment.value
  form.value.date = selectedAppointment.value.date || ''
  form.value.time = selectedAppointment.value.time || ''
  form.value.note = selectedAppointment.value.note || ''
  isDialogOpen.value = true
  showMap.value = false
}

const getMapEmbedUrl = (lat: number, lng: number) => {
  return `https://www.openstreetmap.org/export/embed.html?bbox=${lng - 0.005},${lat - 0.003},${lng + 0.005},${lat + 0.003}&layer=mapnik&marker=${lat},${lng}`
}

// Show multiple markers on one map
const getMultiMarkerMapUrl = (spots: { lat: number; lng: number; label: string }[]) => {
  if (spots.length === 0) return ''
  const lats = spots.map(s => s.lat)
  const lngs = spots.map(s => s.lng)
  const minLat = Math.min(...lats) - 0.005
  const maxLat = Math.max(...lats) + 0.005
  const minLng = Math.min(...lngs) - 0.005
  const maxLng = Math.max(...lngs) + 0.005
  let url = `https://www.openstreetmap.org/export/embed.html?bbox=${minLng},${minLat},${maxLng},${maxLat}&layer=mapnik`
  spots.forEach(s => { url += `&marker=${s.lat},${s.lng}` })
  return url
}

const getDirectionsUrl = (lat: number, lng: number) => {
  return `https://www.openstreetmap.org/directions?to=${lat},${lng}#map=17/${lat}/${lng}`
}

// Get route between two points
const getRouteMapUrl = (fromLat: number, fromLng: number, toLat: number, toLng: number) => {
  return `https://www.openstreetmap.org/directions?from=${fromLat},${fromLng}&to=${toLat},${toLng}#map=15/${(fromLat+toLat)/2}/${(fromLng+toLng)/2}`
}

// ─── OpenStreetMap Nominatim Search ──────────────────────────────────────────
const mapSearchQuery = ref('')
const mapSearchResults = ref<{ name: string; lat: string; lng: string }[]>([])
const isSearchingMap = ref(false)

const searchOpenMap = async () => {
  if (!mapSearchQuery.value.trim()) return
  isSearchingMap.value = true
  try {
    const q = encodeURIComponent(mapSearchQuery.value.trim())
    const res = await fetch(`https://nominatim.openstreetmap.org/search?q=${q}&format=json&limit=5`)
    if (res.ok) {
      const data = await res.json()
      mapSearchResults.value = (data as any[]).map((r: any) => ({
        name: r.display_name?.split(',')[0] ?? 'Location',
        lat: r.lat,
        lng: r.lon,
      }))
    }
  } catch { mapSearchResults.value = [] }
  finally { isSearchingMap.value = false }
}

const selectMapSearchResult = (result: { name: string; lat: string; lng: string }) => {
  // Create a pseudo-location from search result
  suggestedLocations.value = [{
    location: {
      id: 0, name: result.name, description: 'From OpenStreetMap search',
      address: '', image_url: '',
      latitude: parseFloat(result.lat), longitude: parseFloat(result.lng),
      open_time: '', close_time: '', is_active: true,
    },
    score: 90, reasons: ['🔍 Search result'], distance: '—',
  }]
  mapSearchResults.value = []
  mapSearchQuery.value = ''
}

// ─── Notification System ───────────────────────────────────────────────────────
const generateNotifications = () => {
  const newNotifications: Notification[] = []
  const now = new Date()

  appointments.value.forEach((appt) => {
    if (appt.status === 'completed' || appt.status === 'cancelled') return

    // Upcoming within 1 hour
    if (appt.date && appt.time) {
      const apptTime = new Date(`${appt.date}T${appt.time}`)
      const diffMs = apptTime.getTime() - now.getTime()
      const diffMinutes = diffMs / (1000 * 60)
      const diffHours = diffMs / (1000 * 60 * 60)

      if (diffMinutes > 0 && diffMinutes <= 60) {
        newNotifications.push({
          id: `urgent-${appt.id}`,
          type: 'upcoming',
          title: '⏰ Meeting soon!',
          message: `Your appointment "${appt.title}" with ${appt.partner} is in ${Math.round(diffMinutes)} minutes at ${appt.location || 'TBD'}.`,
          appointmentId: appt.id,
          time: new Date(),
          read: false,
        })
      } else if (diffHours > 0 && diffHours <= 24) {
        newNotifications.push({
          id: `today-${appt.id}`,
          type: 'reminder',
          title: '📅 Upcoming today',
          message: `"${appt.title}" with ${appt.partner} — ${appt.time} at ${appt.location || 'TBD'}.`,
          appointmentId: appt.id,
          time: new Date(),
          read: false,
        })
      }
    }

    // Needs confirmation
    if (appt.status === 'scheduled') {
      newNotifications.push({
        id: `confirm-${appt.id}`,
        type: 'confirmation',
        title: '🤝 Needs confirmation',
        message: `"${appt.title}" with ${appt.partner} is waiting for confirmation.`,
        appointmentId: appt.id,
        time: new Date(),
        read: false,
      })
    }
  })

  // Merge with existing, keep read status
  const existingRead = new Set(
    notifications.value.filter((n) => n.read).map((n) => n.id),
  )
  newNotifications.forEach((n) => {
    if (existingRead.has(n.id)) n.read = true
  })

  notifications.value = newNotifications
}

const markAllRead = () => {
  notifications.value.forEach((n) => (n.read = true))
}

const markNotificationRead = (notif: Notification) => {
  notif.read = true
}

const clearNotifications = () => {
  notifications.value = []
}

// ─── Actions ───────────────────────────────────────────────────────────────────
const openEdit = (appt: Appointment) => {
  editingAppointment.value = appt
  form.value.date = appt.date || ''
  form.value.time = appt.time || ''
  form.value.location = appt.location || ''
  form.value.note = appt.note || ''
  isDialogOpen.value = true
}

const saveAppointment = async () => {
  if (!editingAppointment.value) return
  isSaving.value = true
  try {
    const { error } = await api.PATCH('/api/appointments/{id}' as any, {
      params: { path: { id: editingAppointment.value.id } },
      body: {
        date: form.value.date,
        time: form.value.time,
        location: form.value.location,
        note: form.value.note,
      },
    } as any)
    if (error) {
      toast.error('Failed to update appointment')
      return
    }
    toast.success('Appointment updated')
    isDialogOpen.value = false
    await fetchAppointments()
  } finally {
    isSaving.value = false
  }
}

const confirmAppointment = async (id: string) => {
  try {
    const { error } = await api.PATCH('/api/appointments/{id}' as any, {
      params: { path: { id } },
      body: { status: 'confirmed' },
    } as any)
    if (error) {
      toast.error('Failed to confirm')
      return
    }
    toast.success('Appointment confirmed! 🎉')
    await fetchAppointments()
  } catch {
    toast.error('An error occurred')
  }
}

const markAsCompleted = async (id: string) => {
  try {
    const { error } = await api.PATCH('/api/appointments/{id}' as any, {
      params: { path: { id } },
      body: { status: 'completed' },
    } as any)
    if (error) {
      toast.error('Failed to complete appointment')
      return
    }
    toast.success('Trade completed successfully! 🎉')
    await fetchAppointments()
  } catch {
    toast.error('An error occurred')
  }
}

const cancelAppointment = async (id: string) => {
  try {
    const { error } = await api.PATCH('/api/appointments/{id}' as any, {
      params: { path: { id } },
      body: { status: 'cancelled' },
    } as any)
    if (error) {
      toast.error('Failed to cancel')
      return
    }
    toast.success('Appointment cancelled')
    await fetchAppointments()
  } catch {
    toast.error('An error occurred')
  }
}

// ─── Lifecycle ─────────────────────────────────────────────────────────────────
onMounted(async () => {
  await Promise.all([fetchAppointments(), fetchLocations()])
  // Poll for notifications every 60 seconds
  notificationPollInterval.value = setInterval(() => {
    generateNotifications()
    lastCheckedTime.value = new Date()
  }, 60000)
})

onUnmounted(() => {
  if (notificationPollInterval.value) {
    clearInterval(notificationPollInterval.value)
  }
})

// Show toast for urgent appointments on load
watch(urgentAppointments, (urgent) => {
  urgent.forEach((appt) => {
    toast.warning(`⏰ Upcoming: "${appt.title}" with ${appt.partner} soon!`, {
      duration: 8000,
    })
  })
}, { immediate: true })

// ─── Time formatters ───────────────────────────────────────────────────────────
const formatTimeAgo = (date: Date) => {
  const mins = Math.floor((new Date().getTime() - date.getTime()) / 60000)
  if (mins < 1) return 'Just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

const getTimeUntil = (appt: Appointment): string => {
  if (!appt.date || !appt.time) return ''
  const apptTime = new Date(`${appt.date}T${appt.time}`)
  const diffMs = apptTime.getTime() - Date.now()
  if (diffMs <= 0) return 'Passed'
  const hours = Math.floor(diffMs / 3600000)
  const mins = Math.floor((diffMs % 3600000) / 60000)
  if (hours > 24) return `${Math.floor(hours / 24)}d ${hours % 24}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

const getScoreColor = (score: number) => {
  if (score >= 85) return 'text-emerald-600'
  if (score >= 70) return 'text-amber-600'
  return 'text-slate-500'
}
</script>

<template>
  <div class="space-y-6">
    <!-- ═══ HEADER BANNER ═══ -->
    <section
      class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-rose-50 via-white to-amber-50 p-6"
    >
      <div class="pointer-events-none absolute -right-10 -top-8 h-32 w-32 rounded-full bg-rose-200/60" />
      <div class="pointer-events-none absolute -bottom-12 -left-8 h-40 w-40 rounded-full bg-amber-200/50" />
      <div class="pointer-events-none absolute right-40 top-4 h-16 w-16 rounded-full bg-sky-200/40" />
      <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-rose-600">
            Appointments
          </p>
          <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">
            Appointment scheduling board
          </h1>
          <p class="mt-2 max-w-xl text-sm text-slate-600">
            Coordinate meet-ups with AI-suggested safe locations, map directions, and
            real-time notifications.
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <!-- Notification Bell -->
          <div class="relative">
            <Button
              variant="outline"
              size="icon"
              class="relative"
              @click="showNotifications = !showNotifications"
            >
              <Bell v-if="unreadCount === 0" class="h-5 w-5" />
              <BellRing v-else class="h-5 w-5 text-rose-500" />
            </Button>
            <span
              v-if="unreadCount > 0"
              class="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-rose-500 text-[10px] font-bold text-white shadow"
            >
              {{ unreadCount }}
            </span>

            <!-- Notification Dropdown -->
            <Card
              v-if="showNotifications"
              class="absolute right-0 top-12 z-50 w-80 shadow-xl border-2"
            >
              <CardHeader class="pb-2">
                <div class="flex items-center justify-between">
                  <CardTitle class="text-sm">Notifications</CardTitle>
                  <div class="flex gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      class="h-7 text-xs"
                      @click="markAllRead"
                    >
                      Mark all read
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      class="h-7 text-xs text-muted-foreground"
                      @click="clearNotifications"
                    >
                      Clear
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent class="max-h-72 overflow-y-auto space-y-2">
                <div v-if="notifications.length === 0" class="py-6 text-center text-sm text-muted-foreground">
                  <Bell class="mx-auto mb-2 h-8 w-8 text-slate-300" />
                  No notifications yet
                </div>
                <div
                  v-for="notif in notifications"
                  :key="notif.id"
                  class="flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors hover:bg-muted/50"
                  :class="{ 'bg-rose-50/50 border-rose-200': !notif.read }"
                  @click="markNotificationRead(notif)"
                >
                  <div
                    class="mt-0.5 h-2 w-2 flex-shrink-0 rounded-full"
                    :class="notif.read ? 'bg-slate-300' : 'bg-rose-500'"
                  />
                  <div class="min-w-0 flex-1">
                    <p class="text-xs font-semibold">{{ notif.title }}</p>
                    <p class="text-xs text-muted-foreground line-clamp-2">{{ notif.message }}</p>
                    <p class="mt-1 text-[10px] text-muted-foreground">
                      {{ formatTimeAgo(notif.time) }}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          <Button variant="outline" class="gap-2">
            <Calendar class="h-4 w-4" />
            Sync calendar
          </Button>
        </div>
      </div>
    </section>

    <!-- ═══ STATS CARDS ═══ -->
    <section class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <Card
        v-for="card in statsCards"
        :key="card.label"
        class="transition-shadow hover:shadow-md"
      >
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardDescription>{{ card.label }}</CardDescription>
          <component :is="card.icon" :class="['h-5 w-5', card.color]" />
        </CardHeader>
        <CardContent>
          <div class="text-3xl font-bold">{{ card.value }}</div>
          <p class="mt-1 text-xs text-muted-foreground">{{ card.sub }}</p>
        </CardContent>
      </Card>
    </section>

    <!-- ═══ URGENT ALERTS ═══ -->
    <div v-if="urgentAppointments.length > 0" class="space-y-2">
      <div
        v-for="appt in urgentAppointments"
        :key="'urgent-' + appt.id"
        class="flex items-center gap-3 rounded-lg border-2 border-rose-200 bg-rose-50/70 px-4 py-3"
      >
        <AlertCircle class="h-5 w-5 flex-shrink-0 text-rose-500" />
        <div class="flex-1 text-sm">
          <span class="font-semibold text-rose-700">Upcoming soon:</span>
          {{ appt.title }} with {{ appt.partner }} —
          <span class="font-medium">{{ getTimeUntil(appt) }} remaining</span>
        </div>
        <Button size="sm" variant="outline" class="gap-1" @click="openMapForAppointment(appt)">
          <Map class="h-3.5 w-3.5" />
          View map
        </Button>
      </div>
    </div>

    <!-- ═══ SEARCH & TABS ═══ -->
    <section class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <Tabs v-model="activeTab" class="w-full sm:w-auto">
        <TabsList>
          <TabsTrigger value="upcoming" class="gap-1.5">
            <Calendar class="h-3.5 w-3.5" />
            Upcoming
            <Badge variant="secondary" class="ml-1 h-5 px-1.5 text-[10px]">
              {{ upcomingAppointments.length }}
            </Badge>
          </TabsTrigger>
          <TabsTrigger value="past" class="gap-1.5">
            <CheckCircle2 class="h-3.5 w-3.5" />
            Past
          </TabsTrigger>
          <TabsTrigger value="all">All</TabsTrigger>
        </TabsList>
      </Tabs>
      <div class="flex items-center gap-3">
        <Input v-model="search" placeholder="Search appointments..." class="max-w-[220px]" />
        <p class="whitespace-nowrap text-xs text-muted-foreground">
          {{ filteredAppointments.length }} result{{ filteredAppointments.length !== 1 ? 's' : '' }}
        </p>
      </div>
    </section>

    <!-- ═══ MAIN GRID: Appointments + Map Panel ═══ -->
    <section class="grid gap-6 lg:grid-cols-[1fr_380px]">
      <!-- LEFT: Appointment List -->
      <div class="space-y-3">
        <!-- Loading State -->
        <div v-if="isLoading" class="space-y-3">
          <Card v-for="i in 3" :key="i" class="animate-pulse">
            <CardContent class="p-5">
              <div class="flex items-center gap-3">
                <div class="h-10 w-10 rounded-full bg-slate-200" />
                <div class="flex-1 space-y-2">
                  <div class="h-4 w-2/3 rounded bg-slate-200" />
                  <div class="h-3 w-1/2 rounded bg-slate-100" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <!-- Empty State -->
        <Card v-else-if="filteredAppointments.length === 0" class="py-16">
          <CardContent class="text-center">
            <Calendar class="mx-auto mb-3 h-12 w-12 text-slate-300" />
            <p class="text-lg font-semibold text-slate-500">No appointments yet</p>
            <p class="mt-1 text-sm text-muted-foreground">
              When you arrange trades, they'll appear here with AI-suggested meetup spots.
            </p>
          </CardContent>
        </Card>

        <!-- Appointment Cards -->
        <Card
          v-for="item in filteredAppointments"
          :key="item.id"
          class="group transition-all hover:shadow-md"
          :class="{
            'border-rose-300 bg-rose-50/30': isUrgent(item),
            'border-emerald-200 bg-emerald-50/20': item.status === 'completed',
            'border-slate-200': !isUrgent(item) && item.status !== 'completed',
          }"
        >
          <CardContent class="p-5">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
              <!-- Left: Info -->
              <div class="flex-1 space-y-2.5">
                <div class="flex items-center gap-2 flex-wrap">
                  <Badge :variant="statusVariant(item.status)" class="gap-1">
                    <component :is="statusIcon(item.status)" class="h-3 w-3" />
                    {{ statusLabel(item.status) }}
                  </Badge>
                  <span class="text-[11px] text-muted-foreground font-mono">#{{ item.id }}</span>
                  <Badge
                    v-if="isUrgent(item)"
                    variant="destructive"
                    class="gap-1 animate-pulse text-[10px]"
                  >
                    <Zap class="h-3 w-3" />
                    ASAP
                  </Badge>
                </div>

                <div>
                  <p class="text-base font-semibold text-foreground">{{ item.title }}</p>
                  <p class="flex items-center gap-1.5 text-sm text-muted-foreground">
                    <Users class="h-3.5 w-3.5" />
                    With <span class="font-medium text-foreground">{{ item.partner }}</span>
                  </p>
                </div>

                <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm">
                  <span class="flex items-center gap-1.5" :class="isUrgent(item) ? 'text-rose-600 font-semibold' : 'text-foreground'">
                    <Clock3 class="h-4 w-4" :class="isUrgent(item) ? 'text-rose-500' : 'text-slate-400'" />
                    {{ item.date }} · {{ item.time }}
                    <span v-if="getTimeUntil(item) && item.status !== 'completed' && item.status !== 'cancelled'"
                      class="ml-1 text-xs"
                      :class="isUrgent(item) ? 'text-rose-500' : 'text-muted-foreground'"
                    >
                      (in {{ getTimeUntil(item) }})
                    </span>
                  </span>
                  <span class="flex items-center gap-1.5 text-muted-foreground">
                    <MapPin class="h-4 w-4 text-emerald-500" />
                    {{ item.location || 'No location set — tap map to find one' }}
                  </span>
                </div>

                <p v-if="item.note" class="text-xs text-muted-foreground italic">
                  💬 {{ item.note }}
                </p>
              </div>

              <!-- Right: Actions -->
              <div class="flex flex-row flex-wrap items-center gap-2 sm:flex-col sm:items-end">
                <Button
                  size="sm"
                  variant="outline"
                  class="gap-1.5"
                  @click="openMapForAppointment(item)"
                >
                  <Map class="h-3.5 w-3.5" />
                  Map & Directions
                </Button>

                <div class="flex gap-1.5">
                  <Button
                    v-if="item.status === 'scheduled'"
                    size="sm"
                    variant="default"
                    class="gap-1"
                    @click="confirmAppointment(item.id)"
                  >
                    <CheckCircle2 class="h-3.5 w-3.5" />
                    Confirm
                  </Button>
                  <Button
                    v-if="item.status !== 'completed' && item.status !== 'cancelled'"
                    size="sm"
                    variant="outline"
                    @click="openEdit(item)"
                  >
                    Edit
                  </Button>
                  <Button
                    v-if="item.status === 'confirmed'"
                    size="sm"
                    variant="default"
                    class="gap-1 bg-emerald-600 hover:bg-emerald-700"
                    @click="markAsCompleted(item.id)"
                  >
                    <CheckCircle2 class="h-3.5 w-3.5" />
                    Complete
                  </Button>
                  <Button
                    v-if="item.status !== 'completed' && item.status !== 'cancelled'"
                    size="sm"
                    variant="ghost"
                    class="text-muted-foreground hover:text-destructive"
                    @click="cancelAppointment(item.id)"
                  >
                    <XCircle class="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </div>

            <!-- Expandable note about status -->
            <div
              v-if="item.status === 'cancelled'"
              class="mt-3 rounded-md bg-muted/50 px-3 py-2 text-xs text-muted-foreground"
            >
              This appointment was cancelled. You can still view the details.
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- RIGHT: Map & AI Suggestions Panel -->
      <div class="space-y-4">
        <!-- Map Panel -->
        <Card class="sticky top-6" :class="{ 'shadow-xl': showMap }">
          <CardHeader class="pb-2">
            <div class="flex items-center justify-between">
              <CardTitle class="flex items-center gap-2 text-base">
                <Sparkles class="h-4 w-4 text-amber-500" />
                AI Meetup Finder
              </CardTitle>
              <Button
                v-if="showMap"
                variant="ghost"
                size="sm"
                class="h-7 text-xs"
                @click="showMap = false"
              >
                Close
              </Button>
            </div>
            <CardDescription>
              Smart location suggestions with map & directions
            </CardDescription>
          </CardHeader>

          <CardContent class="space-y-4">
            <!-- Default state: prompt -->
            <div v-if="!showMap" class="space-y-4">
              <div
                class="flex h-40 items-center justify-center rounded-xl border-2 border-dashed bg-gradient-to-br from-amber-50/50 via-white to-sky-50/50"
              >
                <div class="text-center">
                  <Map class="mx-auto mb-2 h-10 w-10 text-amber-400" />
                  <p class="text-sm font-medium text-slate-600">
                    Tap <strong>"Map & Directions"</strong>
                  </p>
                  <p class="text-xs text-muted-foreground mt-1">
                    on any appointment card to find meetup spots
                  </p>
                  <Button
                    variant="outline"
                    size="sm"
                    class="mt-3 gap-1.5"
                    @click="openMapForAppointment({ id: 'demo', title: 'Demo', date: '', time: '', location: '', partner: 'Demo', status: 'scheduled', note: '' })"
                  >
                    <Sparkles class="h-3.5 w-3.5" /> Try Demo
                  </Button>
                </div>
              </div>

              <div class="space-y-2 rounded-lg border bg-muted/30 p-3">
                <p class="text-xs font-semibold text-slate-700 flex items-center gap-1.5">
                  <Zap class="h-3.5 w-3.5 text-amber-500" />
                  How it works
                </p>
                <ul class="space-y-1.5 text-xs text-muted-foreground">
                  <li class="flex items-start gap-1.5">
                    <span class="mt-0.5 text-amber-500">•</span>
                    AI analyzes your location & your partner's to find the safest midpoint
                  </li>
                  <li class="flex items-start gap-1.5">
                    <span class="mt-0.5 text-amber-500">•</span>
                    Suggests campus spots with security cameras, good lighting, and open hours
                  </li>
                  <li class="flex items-start gap-1.5">
                    <span class="mt-0.5 text-amber-500">•</span>
                    Shows map with directions so both parties arrive on time
                  </li>
                </ul>
              </div>
            </div>

            <!-- Loading suggestions -->
            <div v-if="isGeneratingSuggestion" class="space-y-3">
              <div class="flex items-center justify-center py-8">
                <div class="text-center">
                  <Loader2 class="mx-auto h-8 w-8 animate-spin text-amber-500" />
                  <p class="mt-2 text-sm font-medium text-slate-600">AI analyzing best meetup spots...</p>
                  <p class="text-xs text-muted-foreground">Checking distance, safety & availability</p>
                </div>
              </div>
            </div>

            <!-- Map + Suggestions -->
            <div v-if="showMap && !isGeneratingSuggestion && selectedAppointment" class="space-y-4">
              <!-- Appointment summary -->
              <div class="rounded-lg border bg-muted/30 p-3">
                <p class="text-xs font-semibold">{{ selectedAppointment.title }}</p>
                <p class="text-xs text-muted-foreground">
                  With {{ selectedAppointment.partner }} · {{ selectedAppointment.date }}
                </p>
              </div>

              <!-- Map embed -->
              <div
                class="overflow-hidden rounded-xl border"
                :class="mapExpanded ? 'h-64' : 'h-44'"
              >
                <iframe
                  v-if="suggestedLocations.length > 0"
                  :src="getMultiMarkerMapUrl(suggestedLocations.map(s => ({
                    lat: s.location.latitude, lng: s.location.longitude, label: s.location.name
                  })))"
                  width="100%"
                  height="100%"
                  frameborder="0"
                  scrolling="no"
                  class="border-0"
                  title="Meetup location map"
                />
                <div v-else class="flex h-full items-center justify-center bg-slate-50 dark:bg-slate-900">
                  <div class="text-center">
                    <Map class="mx-auto h-8 w-8 text-muted-foreground/30" />
                    <p class="mt-1 text-xs text-muted-foreground">Tap "Map & Directions" on an appointment</p>
                  </div>
                </div>
              </div>

              <!-- Route button -->
              <div v-if="suggestedLocations.length > 0" class="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  class="flex-1 gap-1 text-[11px]"
                  @click="window.open(
                    getDirectionsUrl(suggestedLocations[0].location.latitude, suggestedLocations[0].location.longitude),
                    '_blank'
                  )"
                >
                  <Navigation class="h-3 w-3" /> Directions to top pick
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  class="text-xs"
                  @click="mapExpanded = !mapExpanded"
                >
                  <component :is="mapExpanded ? ChevronUp : ChevronDown" class="h-3 w-3" />
                </Button>
              </div>

              <!-- AI Score explanation -->
              <div class="rounded-lg border border-amber-200 bg-amber-50/50 p-3">
                <p class="text-xs font-semibold text-amber-800 flex items-center gap-1.5">
                  <Sparkles class="h-3.5 w-3.5" />
                  AI Recommendation
                </p>
                <p class="mt-1 text-xs text-amber-700">
                  Top pick based on distance between both parties, open hours, safety, and campus popularity.
                </p>
              </div>

              <!-- OpenStreetMap Search -->
              <div class="space-y-2">
                <p class="text-xs font-semibold text-slate-700 flex items-center gap-1.5">
                  <Search class="h-3 w-3" /> Find any location
                </p>
                <div class="flex gap-1.5">
                  <Input
                    v-model="mapSearchQuery"
                    placeholder="Search OpenStreetMap..."
                    class="h-8 text-xs"
                    @keyup.enter="searchOpenMap"
                  />
                  <Button size="sm" variant="outline" class="h-8 text-xs" @click="searchOpenMap" :disabled="isSearchingMap">
                    <Loader2 v-if="isSearchingMap" class="h-3 w-3 animate-spin" />
                    <Search v-else class="h-3 w-3" />
                  </Button>
                </div>
                <!-- Search results -->
                <div v-if="mapSearchResults.length > 0" class="space-y-1 max-h-32 overflow-y-auto">
                  <button
                    v-for="result in mapSearchResults"
                    :key="result.lat + result.lng"
                    class="w-full text-left rounded-lg border px-2.5 py-1.5 text-[11px] hover:bg-muted transition-colors flex items-center gap-2"
                    @click="selectMapSearchResult(result)"
                  >
                    <MapPin class="h-3 w-3 text-rose-500 shrink-0" />
                    <span class="truncate">{{ result.name }}</span>
                  </button>
                </div>
              </div>

              <!-- Suggested locations list -->
              <div class="space-y-2">
                <p class="text-xs font-semibold text-slate-700">Suggested Meetup Spots</p>
                <div
                  v-for="(suggested, idx) in suggestedLocations"
                  :key="suggested.location.name"
                  class="cursor-pointer rounded-lg border p-3 transition-all hover:border-amber-300 hover:bg-amber-50/30"
                  :class="idx === 0 ? 'border-amber-300 bg-amber-50/40 ring-1 ring-amber-200' : ''"
                  @click="selectSuggestedLocation(suggested)"
                >
                  <div class="flex items-start justify-between gap-2">
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-1.5">
                        <span
                          v-if="idx === 0"
                          class="flex h-5 w-5 items-center justify-center rounded-full bg-amber-400 text-[10px] font-bold text-white"
                        >
                          ★
                        </span>
                        <p class="text-sm font-semibold truncate">{{ suggested.location.name }}</p>
                      </div>
                      <p class="mt-0.5 text-xs text-muted-foreground">
                        ~{{ suggested.distance }} away
                      </p>
                      <div class="mt-1.5 flex flex-wrap gap-1">
                        <Badge
                          v-for="reason in suggested.reasons"
                          :key="reason"
                          variant="secondary"
                          class="text-[10px]"
                        >
                          {{ reason }}
                        </Badge>
                      </div>
                      <p
                        v-if="suggested.location.description"
                        class="mt-1.5 text-[11px] text-muted-foreground line-clamp-2"
                      >
                        {{ suggested.location.description }}
                      </p>
                    </div>
                    <div class="flex flex-col items-end gap-1">
                      <span
                        class="text-sm font-bold"
                        :class="getScoreColor(suggested.score)"
                      >
                        {{ suggested.score }}%
                      </span>
                      <span class="text-[10px] text-muted-foreground">match</span>
                    </div>
                  </div>

                  <div class="mt-2 flex gap-1.5">
                    <Button
                      size="sm"
                      variant="outline"
                      class="h-7 gap-1 text-[11px]"
                      @click.stop="
                        window.open(
                          getDirectionsUrl(
                            suggested.location.latitude,
                            suggested.location.longitude,
                            suggested.location.name,
                          ),
                          '_blank',
                        )
                      "
                    >
                      <Navigation class="h-3 w-3" />
                      Directions
                    </Button>
                    <Button
                      size="sm"
                      variant="default"
                      class="h-7 gap-1 text-[11px]"
                      @click.stop="selectSuggestedLocation(suggested)"
                    >
                      <MapPin class="h-3 w-3" />
                      Use this spot
                    </Button>
                  </div>
                </div>
              </div>

              <!-- No suggestions fallback -->
              <div
                v-if="suggestedLocations.length === 0"
                class="rounded-lg border bg-muted/30 p-4 text-center"
              >
                <Info class="mx-auto mb-1 h-5 w-5 text-muted-foreground" />
                <p class="text-xs text-muted-foreground">
                  No matching locations found. Try adding more campus locations in
                  <strong>Dashboard → Locations</strong>.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Quick Tips Card -->
        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="flex items-center gap-2 text-sm">
              <ShieldCheck class="h-4 w-4 text-emerald-500" />
              Safety Tips
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-2 text-xs text-muted-foreground">
            <div class="flex items-start gap-2">
              <span class="mt-0.5 text-emerald-500">✓</span>
              <span>Always meet in well-lit, public campus areas</span>
            </div>
            <div class="flex items-start gap-2">
              <span class="mt-0.5 text-emerald-500">✓</span>
              <span>Bring a friend for high-value item exchanges</span>
            </div>
            <div class="flex items-start gap-2">
              <span class="mt-0.5 text-emerald-500">✓</span>
              <span>Inspect items before confirming the trade</span>
            </div>
            <div class="flex items-start gap-2">
              <span class="mt-0.5 text-emerald-500">✓</span>
              <span>Use campus Wi-Fi to stay connected during meetups</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </section>

    <!-- ═══ EDIT DIALOG ═══ -->
    <Dialog v-model:open="isDialogOpen">
      <DialogContent class="sm:max-w-[450px]">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2">
            <MapPin class="h-5 w-5 text-rose-500" />
            Update Appointment
          </DialogTitle>
          <DialogDescription>
            Change the date, time, location, and notes for your meet-up.
          </DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
            <label class="text-sm font-medium">Date</label>
            <Input type="date" v-model="form.date" />
          </div>
          <div class="grid gap-2">
            <label class="text-sm font-medium">Time</label>
            <Input type="time" v-model="form.time" />
          </div>
          <div class="grid gap-2">
            <label class="text-sm font-medium flex items-center gap-1.5">
              Location
              <Badge variant="secondary" class="text-[10px] h-4 gap-0.5">
                <Sparkles class="h-2.5 w-2.5" />
                AI suggested
              </Badge>
            </label>
            <Input
              type="text"
              v-model="form.location"
              placeholder="e.g. Library Hall"
            />
            <p class="text-[11px] text-muted-foreground">
              Use the Map & Directions panel to get AI-recommended safe meetup spots.
            </p>
          </div>
          <div class="grid gap-2">
            <label class="text-sm font-medium">Note</label>
            <Input
              type="text"
              v-model="form.note"
              placeholder="Extra instructions for your partner..."
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="isDialogOpen = false">Cancel</Button>
          <Button :disabled="isSaving" @click="saveAppointment" class="gap-1.5">
            <Loader2 v-if="isSaving" class="h-4 w-4 animate-spin" />
            {{ isSaving ? 'Saving...' : 'Save changes' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Click-outside handler for notification dropdown -->
    <div
      v-if="showNotifications"
      class="fixed inset-0 z-40"
      @click="showNotifications = false"
    />
  </div>
</template>
