# 🎨 Design Theme Standard — Campus Marketplace

> **Last updated:** 2026-06-09  
> **Purpose:** Single source of truth for all UI/UX design decisions.  
> **When changing the theme:** Update this file FIRST, then propagate to code.

---

## 0. 🧭 Layout Architecture

### 0.1 Sidebar Navigation (replaced top header bar)

The app uses a **collapsible sidebar** layout instead of a top navigation bar.

```
┌──────────┬──────────────────────────────────────┐
│          │                                      │
│  Sidebar │         Main Content Area            │
│  280px   │         (scrollable)                 │
│          │                                      │
│  ─────── │                                      │
│  Logo    │   Page Header / Banner               │
│  ─────── │                                      │
│  Search  │   Content Grid                       │
│  ─────── │                                      │
│  Nav     │                                      │
│  Links   │                                      │
│  ─────── │                                      │
│  Clip-   │                                      │
│  board   │                                      │
│  ─────── │                                      │
│  User    │   Footer                             │
│  Profile │                                      │
└──────────┴──────────────────────────────────────┘
```

**Sidebar specs:**
- Width: `280px` expanded, `72px` collapsed (on `lg:` screens)
- Position: `fixed` on mobile (with overlay backdrop), `relative` on desktop
- Collapse toggle: Chevron button in sidebar header, hamburger on mobile
- Background: `bg-background/95 backdrop-blur`
- Sections: Logo → Quick Search → Navigation → Clipboard → User Profile
- Transition: `transition-all duration-300`

**Sidebar nav states:**
- Active: `bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300`
- Inactive: `text-muted-foreground hover:bg-muted hover:text-foreground`
- Items: `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium`

### 0.2 Your Cart System

A Pinia store (`useCartStore`) allows users to save items from the marketplace to plan trades:

**Store location:** `src/stores/clipboard.ts` (stores/cart.ts)

**Features:**
- `addItem()` / `removeItem()` / `toggleItem()` — manage cart items
- `updateNote()` — personal notes per item
- `updateMeetupLocation()` — planned meeting spot
- `updateOfferItem()` — what you plan to trade in return
- `clearAll()` — remove all cart items
- `isInCart(id)` — check if an item is in the cart
- Persisted to `localStorage` as `cart_items`

**Cart UI (in sidebar):**
- Collapsible "Your Cart" section at bottom of nav
- Shows item thumbnail, name, type badge
- Inline editable fields: Note, Meetup Location, Offer Item
- "Clear cart" button
- Empty state with helper text

**Cart button on cards:**
- Floating button top-left of card images (appears on hover)
- `Bookmark` icon (not in cart) / `BookmarkCheck` (in cart, amber color)
- Also in slide-over panel header as "Add to Cart" / "In Cart" button
- Toast feedback: 🛒 "Added to your cart" / "Removed from your cart"


| Library | Version | Notes |
|---------|---------|-------|
| **shadcn-vue** | latest | All UI primitives (`Button`, `Card`, `Dialog`, `Badge`, `Input`, `Tabs`, etc.) |
| **style** | `reka-nova` | shadcn style variant |
| **Icon library** | `lucide-vue-next` | All icons throughout the app |
| **CSS framework** | Tailwind CSS v4 | Via `@import 'tailwindcss'` |
| **Animations** | `tw-animate-css` | Tailwind animation utilities |
| **Toasts** | `vue-sonner` | Toast notification system |

**Rule:** Never import icons from other libraries. Always use `lucide-vue-next`.

---

## 2. 🎨 Color System

### 2.1 Semantic Tokens (Light Mode)

| Token | Value (oklch) | Visual | Usage |
|-------|---------------|--------|-------|
| `--background` | `0.99 0.01 90` | Off-white warm | Page background |
| `--foreground` | `0.145 0 0` | Near-black | Body text |
| `--card` | `1 0 0` | Pure white | Card surfaces |
| `--primary` | `0.62 0.14 154` | **Emerald green** | Primary buttons, links, active states |
| `--primary-foreground` | `0.98 0.02 160` | White-green tint | Text on primary |
| `--secondary` | `0.965 0.01 120` | Warm gray-green | Secondary buttons, tags |
| `--muted` | `0.97 0 0` | Light gray | Muted backgrounds |
| `--muted-foreground` | `0.556 0 0` | Medium gray | Secondary text, descriptions |
| `--accent` | `0.96 0.02 70` | Warm amber | Accent backgrounds |
| `--destructive` | `0.577 0.245 27.325` | Red | Delete buttons, error states |
| `--border` | `0.922 0 0` | Light gray | Borders, dividers |
| `--ring` | `0.708 0 0` | Medium gray | Focus rings |

### 2.2 Dark Mode

All tokens invert. Key changes:
- `--background` → `0.15 0.01 240` (dark blue-gray)
- `--foreground` → `0.985 0 0` (near white)
- `--card` → `0.205 0 0` (dark card)
- `--border` → `1 0 0 / 10%` (translucent white)

### 2.3 Semantic Color Palette (Banners & Decorations)

Each feature/sprint page uses a distinct gradient pair for its banner:

| Page/Sprint | From (left) | Via (center) | To (right) | Blob 1 (top-right) | Blob 2 (bottom-left) |
|-------------|-------------|---------------|------------|---------------------|----------------------|
| **Home** | `emerald-500/20` | `cyan-500/10` | `transparent` | (hero layout) | (hero layout) |
| **Appointments** | `rose-50` | `white` | `amber-50` | `rose-200/60` | `amber-200/50` |
| **Transactions** | `sky-50` | `white` | `emerald-50` | `sky-200/50` | `emerald-200/40` |
| **AI Chat** | `amber-50` | `white` | `sky-50` | `amber-200/60` | `sky-200/50` |
| **Recommendations** | `emerald-50` | `white` | `orange-50` | `orange-200/60` | `emerald-200/50` |

**Rule for new pages:** Pick a gradient pair from this palette or create a new complementary pair. Document it here.

### 2.4 Status/Functional Colors

| Context | Light | Dark | Usage |
|---------|-------|------|-------|
| Success / Completed | `emerald-500` | `emerald-400` | Completed badges, success toasts |
| Warning / Pending | `amber-500` | `amber-400` | Pending badges, scheduled states |
| Danger / Error | `destructive` token | `destructive` token | Delete, cancel, errors |
| Info / Neutral | `sky-500` | `sky-400` | Info badges |
| Auction accent | `amber-500` | `amber-400` | Auction badges, bidding UI |
| Primary action | `primary` token | `primary` token | Confirm, save, submit |

---

## 3. 🔤 Typography

### 3.1 Font Stack

| Role | Font | Weight | CSS Variable |
|------|------|--------|-------------|
| **Body / UI** | Manrope | 400, 500, 600, 700 | `--font-sans` |
| **Headings** | Space Grotesk | 400, 500, 600, 700 | `--font-heading` |

### 3.2 Typography Scale

```
Page title (h1):        text-2xl font-semibold tracking-tight
Section title (h2):     text-xl font-semibold
Card title (h3):        text-lg font-semibold / text-base font-semibold
Body text:              text-sm text-muted-foreground
Small text / meta:      text-xs text-muted-foreground
Tiny / badges:          text-[11px] / text-[10px]
Sprint label:           text-xs font-semibold uppercase tracking-widest
```

### 3.3 Heading Rule

```css
h1, h2, h3, h4 {
  font-family: var(--font-heading);  /* Space Grotesk */
  letter-spacing: -0.02em;
}
```

All headings use `tracking-tight` (Tailwind class: `tracking-tight`).

---

## 4. 📏 Spacing & Sizing

| Context | Value | Tailwind |
|---------|-------|----------|
| Page padding | 24px (1.5rem) | `p-6` |
| Section gap | 24px | `space-y-6` |
| Card gap (grid) | 16px | `gap-4` |
| Card internal padding | 20px | `p-5` |
| Form field gap | 8px / 16px | `gap-2` / `gap-4` |
| Button gap (icon+text) | 8px | `gap-2` |
| Inline element gap | 8px | `gap-2` |
| Dialog max-width | 450px / 672px | `sm:max-w-[450px]` / `max-w-2xl` |

---

## 5. 🔲 Border Radius

| Element | Radius | Tailwind |
|---------|--------|----------|
| **CSS variable base** | `0.625rem` (10px) | `--radius` |
| Cards | `rounded-2xl` (16px) | Standard card |
| Banner sections | `rounded-2xl` (16px) | Page headers |
| Buttons | default (6px) | shadcn default |
| Inputs | default (6px) | shadcn default |
| Images (gallery) | `rounded-lg` (8px) / `rounded-xl` (12px) | Thumbnails |
| Image lightbox | `rounded-lg` (8px) | Fullscreen image |
| Badges | `rounded-full` | Pills |
| Dialogs | `rounded-xl` (12px) | shadcn default |
| Hero banner | `rounded-3xl` (24px) | Home page only |

**Rule:** Never use `rounded-none` or square corners. Everything is rounded.

---

## 6. 🎪 Banner / Page Header Pattern

Every page MUST use this structure at the top:

```html
<section class="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-[color1] via-white to-[color2] p-6">
  <!-- Decorative blobs -->
  <div class="pointer-events-none absolute -right-10 -top-8 h-32 w-32 rounded-full bg-[color1-blob]" />
  <div class="pointer-events-none absolute -bottom-12 -left-8 h-40 w-40 rounded-full bg-[color2-blob]" />

  <div class="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
    <div>
      <p class="text-xs font-semibold uppercase tracking-widest text-[accent-color]">
        {Page Name}
      </p>
      <h1 class="mt-2 text-2xl font-semibold tracking-tight text-slate-900">
        {Page Title}
      </h1>
      <p class="mt-2 max-w-xl text-sm text-slate-600">
        {Page Description}
      </p>
    </div>
    <!-- Action buttons go here -->
  </div>
</section>
```

**Rules:**
- Always use `overflow-hidden` on the section
- Blobs must be `pointer-events-none`
- Blob 1: top-right, `h-32 w-32`, opacity `/60`
- Blob 2: bottom-left, `h-36 w-36` or `h-40 w-40`, opacity `/50`
- Optional blob 3: center-right, `h-16 w-16`, opacity `/40`
- Sprint label color matches the first gradient color

---

## 7. 🃏 Card Patterns

### 7.1 Standard Card

```html
<Card class="bg-card/80">
  <CardHeader>
    <CardTitle>{title}</CardTitle>
    <CardDescription>{description}</CardDescription>
  </CardHeader>
  <CardContent>
    <!-- content -->
  </CardContent>
</Card>
```

### 7.2 Stat Card

```html
<Card class="transition-shadow hover:shadow-md">
  <CardHeader class="flex flex-row items-center justify-between pb-2">
    <CardDescription>{label}</CardDescription>
    <component :is="Icon" class="h-5 w-5 text-{color}-500" />
  </CardHeader>
  <CardContent>
    <div class="text-3xl font-bold">{value}</div>
    <p class="mt-1 text-xs text-muted-foreground">{subtitle}</p>
  </CardContent>
</Card>
```

### 7.3 Item Card (Marketplace Grid)

```html
<article class="group relative rounded-2xl border bg-card/80 overflow-hidden cursor-pointer
                transition-all hover:shadow-md hover:-translate-y-0.5 hover:border-emerald-400/60">
  <!-- Image: h-44, object-cover, group-hover:scale-105 -->
  <!-- Status badge: absolute top-2 right-2, rounded-full, text-[11px], backdrop-blur-sm -->
  <!-- Content: p-4 space-y-2 -->
  <!-- Categories: rounded-full bg-muted px-2 py-0.5 text-[10px] -->
</article>
```

### 7.4 Auction Card

Same as item card but with:
- Border: `border-amber-200/60`
- Background: `bg-gradient-to-br from-amber-50 via-white to-amber-100/30`
- "Live auction" badge: `bg-amber-500 text-white`
- Bid info table: `rounded-xl border border-amber-200/60 bg-white/80 p-3 text-xs`

---

## 8. 🔘 Button Patterns

All buttons use shadcn `Button` component. Variants:

| Variant | Usage |
|---------|-------|
| `default` | Primary actions: Save, Confirm, Complete |
| `default` + `bg-emerald-600` | Complete trade button |
| `default` + `bg-amber-500` | Auction/Bidding actions |
| `outline` | Secondary: Edit, Cancel, Reschedule |
| `ghost` | Subtle: Delete icon, navigation |
| `destructive` | Delete, Remove |
| `size="sm"` | Compact: card actions, inline buttons |

**Button with icon pattern:**
```html
<Button size="sm" variant="outline" class="gap-1.5">
  <IconName class="h-3.5 w-3.5" />
  Label
</Button>
```

Icon sizes in buttons:
- `size="default"` → icon `h-4 w-4`
- `size="sm"` → icon `h-3.5 w-3.5`

---

## 9. 🏷️ Badge Patterns

| Context | Variant | Class |
|---------|---------|-------|
| Status: available/active | `default` | — |
| Status: scheduled/pending | `secondary` | — |
| Status: completed | `outline` | — |
| Status: cancelled/error | `destructive` | — |
| Auction badge | custom | `bg-amber-500 text-white` |
| Sprint label | custom | `bg-emerald-600 text-white` |
| Count badge (tabs) | `secondary` | `ml-1 h-5 px-1.5 text-[10px]` |
| Notification dot | custom | `bg-rose-500 text-white h-5 w-5 text-[10px]` |

---

## 10. 🖼️ Image Gallery Patterns

### 10.1 Horizontal Scroll Gallery

```html
<div class="flex gap-3 overflow-x-auto pb-2 scrollbar-thin">
  <img
    v-for="(img, idx) in images"
    :key="img.id"
    :src="img.url"
    class="h-28 w-28 shrink-0 rounded-lg border object-cover shadow-sm
           cursor-pointer transition-transform hover:scale-105"
    @click="openLightbox(idx)"
  />
</div>
```

**Rules:**
- Always use `shrink-0` to prevent images from squishing
- Always use `overflow-x-auto` with `scrollbar-thin`
- Always `cursor-pointer` + `hover:scale-105`
- Thumbnail sizes by context:
  - Dialog form: `h-28 w-28`
  - Card list: `h-20 w-20`
  - Detail view: `h-48 w-auto max-w-xs`
  - Panel: `h-52 w-auto max-w-[85%]`

### 10.2 Single Image (Hero)

```html
<div class="relative h-44 overflow-hidden bg-muted">
  <img :src="url" class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105" />
  <div v-else class="h-full flex items-center justify-center">
    <ImageOff class="h-10 w-10 text-muted-foreground/30" />
  </div>
</div>
```

### 10.3 Image Lightbox

Use the `<ImageLightbox>` component everywhere:
```html
<ImageLightbox v-model="lightboxOpen" :images="lightboxImages" :start-index="lightboxStartIdx" />
```

See `src/components/ImageLightbox.vue` for implementation details.

---

## 11. 🎯 Dialog / Modal Pattern

```html
<Dialog v-model:open="isDialogOpen">
  <DialogContent class="sm:max-w-[450px]">
    <DialogHeader>
      <DialogTitle class="flex items-center gap-2">
        <Icon class="h-5 w-5 text-{accent}-500" />
        {Title}
      </DialogTitle>
      <DialogDescription>{Description}</DialogDescription>
    </DialogHeader>
    <div class="grid gap-4 py-4">
      <!-- Form fields: each wrapped in <div class="grid gap-2"> -->
      <!-- Labels: <label class="text-sm font-medium"> -->
    </div>
    <DialogFooter>
      <Button variant="outline" @click="isDialogOpen = false">Cancel</Button>
      <Button :disabled="isSaving" @click="save">
        <Loader2 v-if="isSaving" class="h-4 w-4 animate-spin" />
        {{ isSaving ? 'Saving...' : 'Save changes' }}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

**Rules:**
- Cancel always `variant="outline"`, left side
- Save/Confirm always `variant="default"`, right side
- Show `<Loader2>` spinner when saving

---

## 12. 📊 Slide-over Panel (HomeView)

Used only in HomeView for the item detail slide-over. Key specs:
- Width: `max-w-md` (448px)
- Backdrop: `bg-black/40 backdrop-blur-sm`
- Transition: slides from right, 300ms ease
- Image gallery: horizontal scroll inside panel
- Header: `border-b px-5 py-4` with title + close button
- Body: `p-5 space-y-4`

---

## 13. 📜 Scrollbar Styling

Custom thin scrollbar for horizontal galleries:
```css
.scrollbar-thin {
  scrollbar-width: thin;
  scrollbar-color: oklch(0.85 0 0) transparent;
}
.scrollbar-thin::-webkit-scrollbar { height: 5px; }
.scrollbar-thin::-webkit-scrollbar-track { background: transparent; border-radius: 8px; }
.scrollbar-thin::-webkit-scrollbar-thumb { background: oklch(0.85 0 0); border-radius: 8px; }
.scrollbar-thin::-webkit-scrollbar-thumb:hover { background: oklch(0.7 0 0); }
```

Use the `scrollbar-thin` class on any `overflow-x-auto` container.

---

## 14. 🎬 Animations & Transitions

| Element | Animation |
|---------|-----------|
| Item cards hover | `hover:-translate-y-0.5 hover:shadow-md` |
| Card images hover | `group-hover:scale-105 transition-transform duration-300` |
| Gallery thumbnails hover | `hover:scale-105 transition-transform` |
| Lightbox open/close | `opacity 0.25s ease` (fade) |
| Slide-over panel | `transition: all 300ms ease` (slide + backdrop) |
| Notification pulse | `animate-pulse` on urgent badges |
| Loading spinner | `animate-spin` on `<Loader2>` |
| Skeleton loading | `animate-pulse` on placeholder cards |

---

## 15. 🔲 Background Texture

The body has a subtle dual-radial-gradient background:

**Light mode:**
```css
radial-gradient(circle at 10% 10%, oklch(0.94 0.05 150 / 0.35), transparent 35%),
radial-gradient(circle at 90% 0%, oklch(0.96 0.04 70 / 0.25), transparent 40%)
```

**Dark mode:**
```css
radial-gradient(circle at 10% 10%, oklch(0.35 0.08 150 / 0.25), transparent 40%),
radial-gradient(circle at 90% 0%, oklch(0.4 0.06 210 / 0.2), transparent 45%)
```

**Rule:** This is set globally in `main.css`. Do not override per page.

---

## 16. 🧭 Icon Usage Guidelines

| Icon | Where |
|------|-------|
| `Search` | Search inputs |
| `Package` | Items, inventory, empty states |
| `Tag` | Categories |
| `Gavel` | Auction features |
| `MapPin` | Locations, addresses |
| `Clock3` | Time, schedules, deadlines |
| `Calendar` | Dates, appointments |
| `Users` | Partners, people |
| `Bell` / `BellRing` | Notifications |
| `Sparkles` | AI features |
| `Loader2` | Loading states (spinner) |
| `ImageOff` | Missing images fallback |
| `CheckCircle2` | Completed, confirmed |
| `XCircle` | Cancelled, close |
| `ChevronLeft` / `ChevronRight` | Pagination, galleries |
| `X` | Close buttons |
| `Plus` | Add/create |
| `ArrowLeft` / `ArrowRight` | Navigation |
| `Map` | Map features |
| `Navigation` | Directions |
| `ShieldCheck` | Safety, verified |
| `AlertCircle` | Warnings, urgent |
| `Zap` | Urgent indicators |
| `ExternalLink` | External links |
| `Star` / `TrendingUp` | Popular, trending |
| `Route` | Directions/routes |
| `Info` | Information |

**Icon size rules:**
- In page headers: `h-4 w-4` or `h-5 w-5`
- In buttons: `h-3.5 w-3.5` (sm), `h-4 w-4` (default)
- In stat cards: `h-5 w-5`
- In badges: `h-3 w-3`
- Standalone decorative: `h-8 w-8` to `h-16 w-16`

---

## 17. 📐 Layout Grid Patterns

| Context | Grid |
|---------|------|
| Item cards (marketplace) | `grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4` |
| Stat cards | `grid gap-4 sm:grid-cols-2 lg:grid-cols-4` |
| Main + sidebar | `grid gap-6 lg:grid-cols-[1fr_380px]` |
| Appointments list + map | `grid gap-6 lg:grid-cols-[1fr_380px]` |
| Item detail + offer form | `grid gap-6 lg:grid-cols-[1.2fr_0.8fr]` |
| Auction card grid | `grid gap-4 sm:grid-cols-2 lg:grid-cols-3` |

---

## 18. 📝 Form Patterns

```html
<div class="grid gap-2">
  <label class="text-sm font-medium">{Label}</label>
  <Input v-model="field" placeholder="{placeholder}" />
</div>
```

For textareas:
```html
<div class="grid gap-2">
  <label class="text-sm font-medium">{Label}</label>
  <Textarea v-model="field" rows="3" placeholder="{placeholder}" />
</div>
```

For selects (native):
```html
<select v-model="field" class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm">
  <option value="">{Default}</option>
</select>
```

For radio groups:
```html
<div class="grid gap-2 sm:grid-cols-2">
  <label class="flex items-center gap-2 text-sm">
    <input type="radio" value="option1" v-model="field" class="size-4 rounded border-input" />
    Label
  </label>
</div>
```

For checkboxes:
```html
<label class="flex items-center gap-2 text-sm">
  <input type="checkbox" :value="id" v-model="selectedIds" class="size-4 rounded border-input" />
  Label
</label>
```

---

## 19. 🔔 Notification / Toast Pattern

Use `vue-sonner` for all toasts:
```typescript
import { toast } from 'vue-sonner'

toast.success('Action completed! 🎉')
toast.error('Something went wrong')
toast.warning('Heads up!', { duration: 8000 })
```

**Rules:**
- Success toasts include emoji: `🎉`
- Error toasts are plain text
- Warning toasts use longer duration (8s+)
- Never use `alert()` or `confirm()`

---

## 20. 🔄 Loading States

### 20.1 Skeleton Cards

```html
<div v-if="isLoading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
  <div v-for="n in 3" :key="n" class="rounded-2xl border bg-card animate-pulse overflow-hidden">
    <div class="h-40 bg-muted" />
    <div class="p-4 space-y-2">
      <div class="h-4 bg-muted rounded w-3/4" />
      <div class="h-3 bg-muted rounded w-1/2" />
    </div>
  </div>
</div>
```

### 20.2 Inline Loading

```html
<span v-if="isLoading" class="text-sm text-muted-foreground">Loading items...</span>
```

### 20.3 Button Loading

```html
<Button :disabled="isSaving">
  <Loader2 v-if="isSaving" class="h-4 w-4 animate-spin" />
  {{ isSaving ? 'Saving...' : 'Save' }}
</Button>
```

---

## 21. 🗂️ File Organization

```
src/
├── assets/
│   └── main.css              # Global styles, CSS variables, @layer base
├── components/
│   ├── ui/                    # shadcn-vue components (auto-generated)
│   ├── ImageLightbox.vue      # Custom: fullscreen image viewer
│   └── LoginForm.vue          # Custom: auth form
├── views/
│   ├── HomeView.vue           # Marketplace browse (main landing)
│   ├── ItemDetailView.vue     # Item detail + offer
│   ├── BiddingView.vue        # Auction room
│   ├── InventoryView.vue      # My items CRUD
│   ├── OffersView.vue         # Sent/received offers
│   ├── CategoriesView.vue     # Category management
│   ├── AppointmentsView.vue   # Appointment scheduling + map
│   ├── TransactionsView.vue   # Wallet transactions (placeholder)
│   ├── AiChatView.vue         # AI chat (placeholder)
│   ├── RecommendationsView.vue # AI recommendations (placeholder)
│   ├── LoginView.vue          # Login page
│   ├── SignUpView.vue         # Registration page
│   ├── settings/              # User settings sub-pages
│   └── dashboard/             # Admin dashboard sub-pages
├── api/
│   └── api.ts                 # openapi-fetch typed client
├── stores/
│   ├── auth.ts                # Pinia: auth state (JWT)
│   └── user.ts                # Pinia: user profile
├── router/
│   └── index.ts               # vue-router with auth guards
└── lib/
    └── utils.ts               # cn() utility for class merging
```

---

## 22. 📋 Checklist: Adding a New Page

1. [ ] Pick gradient colors for banner (see §2.3 — add to table)
2. [ ] Use the banner pattern from §6
3. [ ] Use standard card pattern from §7
4. [ ] Use shadcn Button variants from §8
5. [ ] Use `lucide-vue-next` icons only (§14)
6. [ ] Add loading state (§20)
7. [ ] Add empty state with icon + message
8. [ ] Use `scrollbar-thin` for any horizontal scroll areas (§13)
9. [ ] Use `vue-sonner` toast for user feedback (§19)
10. [ ] Add route in `src/router/index.ts` with auth guards
11. [ ] Update this document if new patterns are introduced

---

## 📝 Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-06-09 | Initial standardization document | — |
| 2026-06-09 | Replaced top header bar with collapsible sidebar layout | — |
| 2026-06-09 | Added Your Cart system (Pinia store + sidebar UI + card buttons) | — |
