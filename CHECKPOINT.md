# 🏷️ Project Checkpoint — Campus Marketplace (AiEmStylix/ecommerce)

> **Last updated:** 2026-05-11  
> **Stack:** Go (backend) + Vue 3 + TypeScript (frontend) + PostgreSQL

---

## ✅ Completed Sprints

| Sprint | Feature | Status |
|--------|---------|--------|
| SP1 | User Management (Student & Admin CRUD) | ✅ Done |
| SP2 | Location/Zone Management | ✅ Done |
| SP3 | Category Management | ✅ Done |
| SP4 | Attributes & Tags | ⚠️ Partial (item type: barter/auction; status: available/reserved/traded/hidden) |
| SP5 | Product/Item Management | ✅ Done |
| SP6 | Transactions & Appointments | ⚠️ Partial (offers/barter done; appointments list/edit functional with AI map suggestions & notifications; wallet/transactions placeholder only) |
| SP7 | Statistics & Rating | ⚠️ Partial (admin dashboard basic stats; rating not started) |
| SP8 | Auth & Authorization | ✅ Done (JWT + RBAC: admin/user roles) |
| SP9 | AI & Chatbot | ⚠️ Partial — **Bidding/Auction feature: ✅ Done** (2026-05-11); AI chat + recommendations mock UI |

---

## 🗂️ Architecture Overview

### Backend (Go — `backend/`)
- **Framework:** stdlib `net/http` + `gorm.io/gorm` (PostgreSQL)
- **Entry:** `cmd/api/main.go`
- **Modules:** `internal/{auth,user,items,categories,offer,media,storage,models,middleware,appctx,utils}`
- **Auth:** JWT Bearer token, middleware in `internal/auth/`
- **Storage:** Cloudflare R2 (S3-compatible) via `internal/storage/`
- **Port:** `3000`

### Frontend (Vue 3 — `frontend/`)
- **Framework:** Vue 3 + TypeScript + Vite
- **Routing:** `vue-router` (`src/router/index.ts`)
- **State:** Pinia (`src/stores/`)
- **UI library:** shadcn-vue components (`src/components/ui/`)
- **API client:** `src/api/api.ts` (openapi-fetch typed client)
- **Port (dev):** `5173`

---

## 📌 Implemented: Bidding Feature (2026-05-11)

### Backend: New files
- `internal/models/auction.go` — `AuctionConfig` + `Bid` models
- `internal/auction/repository.go` — GORM-backed data access (rank recalculation, settlement TX)
- `internal/auction/service.go` — Business logic (step ≤ 20% base, deadline, self-bid checks)
- `internal/auction/handler.go` — HTTP handlers (5 endpoints)

### Backend: New routes
| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| POST | `/api/items/{id}/auction` | user | Seller configures auction |
| GET | `/api/items/{id}/auction` | public | Get auction info + bids |
| POST | `/api/items/{id}/bids` | user | Buyer places a bid |
| GET | `/api/items/{id}/bids` | user | List bids |
| POST | `/api/admin/auctions/{id}/settle` | admin | Settle auction |

### Frontend: New/modified files
- `src/views/BiddingView.vue` — Full auction page (countdown, bid list, place-bid form, seller config)
- `src/views/HomeView.vue` — Auction badge on cards, routes auction items to BiddingView
- `src/views/ItemDetailView.vue` — Auction CTA panel for auction-type items
- `src/router/index.ts` — Route `/items/:id/auction` → BiddingView

---

## 🔑 Key Files Reference

| File | Purpose |
|------|---------|
| `backend/cmd/api/main.go` | DI wiring + route registration |
| `backend/internal/models/item.go` | Item model (ItemStatus, ItemType enums) |
| `backend/internal/models/offer.go` | Offer model (barter/cash flows) |
| `backend/internal/items/handler.go` | Items CRUD + image upload handlers |
| `backend/internal/offer/handler.go` | Offer create/list/accept/reject handlers |
| `frontend/src/router/index.ts` | Route definitions + auth guards |
| `frontend/src/views/HomeView.vue` | Marketplace browse page (main landing) |
| `frontend/src/views/ItemDetailView.vue` | Item detail + send offer UI |
| `frontend/src/views/dashboard/DashboardHome.vue` | Admin stats dashboard |
| `frontend/src/views/OffersView.vue` | Buyer's sent offers history |

---

## 🧩 DB Schema (Current)

```
users           — id, name, email, password, avatar_url, role, is_verified, university_id
universities    — id, name, domain
categories      — id, name, parent_id (self-referential)
items           — id, name, description, type (barter|auction), status, user_id
item_images     — id, item_id, url
item_categories — item_id, category_id (join table)
offers          — id, offer_user_id, target_item_id, offer_item_id, cash_amount, offer_message, status
auction_configs — id, item_id, base_price, step_price, deposit_amount, deadline, status (open|settled|cancelled)
bids            — id, auction_config_id, bidder_id, amount, deposit_paid, status (active|won|lost|forfeited), rank
```

---

## 🔧 Dev Setup

```bash
# Backend
cd backend && go run cmd/api/main.go   # or: air (hot-reload)

# Frontend  
cd frontend && npm run dev             # dev server on :5173

# DB
docker-compose -f docker-compose.dev.yml up -d
```

---

## 🚧 Known Issues / TODOs

- `ItemDetailView.vue` image rendering shows URL text instead of `<img>` — needs polish
- Rating/review system not started (SP7)
- Wallet/payment is placeholder only (SP6)
- Appointment creation endpoint not built (SP6 — only GET + PATCH exist)
- AI chat & recommendation logic not built (SP9)
- AI Vision & Smart Match are placeholders (SP9)
