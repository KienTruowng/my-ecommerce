# Ecommerce Backend API

Go REST API cho hệ thống trao đổi/barter dành cho sinh viên.

## Tech Stack

| Thành phần | Công nghệ |
|---|---|
| Ngôn ngữ | Go 1.22+ |
| HTTP Router | `net/http.ServeMux` (native, Go 1.22+) |
| Database | PostgreSQL |
| ORM | `gorm.io/gorm` |
| Type-safe Query | `gorm.io/gen` (package `internal/query`) |
| Auth | JWT (`golang-jwt/jwt`) |
| File Storage | Cloudflare R2 (S3-compatible) |
| API Docs | OpenAPI **3.1** (`swaggo/swag v2` + `@ID` annotations) |
| Hot Reload | `air` |

## Kiến trúc

```
cmd/api/main.go          ← Entry point, DI wiring, route registration
cmd/generate/            ← gorm/gen code generator
internal/
  appctx/               ← Context helpers (GetUserID, WithUserID)
  auth/                 ← JWT service + auth middleware + handlers
  categories/           ← Category CRUD
  config/               ← Env config loader
  items/                ← Item CRUD + image upload (R2)
  media/                ← R2 presign & verify helper
  middleware/           ← CORS middleware
  models/               ← GORM models (shared)
  offer/                ← Offer/Barter module
  query/                ← Auto-generated type-safe query (gorm/gen)
  seeders/              ← DB seed data
  storage/              ← R2 storage adapter
  university/           ← University data
  user/                 ← User profile management
  utils/                ← WriteJSON / WriteError helpers
docs/                    ← OpenAPI 3.1 spec (auto-generated, DO NOT EDIT)
```

Mỗi module tuân theo kiến trúc 3 tầng: **Handler → Service → Repository**.

---

## Chạy local

```bash
# Copy env
cp .env.dev .env

# Chạy với hot-reload (air)
air

# Hoặc chạy thẳng
go run ./cmd/api/
```

Server khởi động tại `http://localhost:3000`.

---

## API Endpoints

> Tất cả routes có prefix `/api`. Routes đánh dấu 🔒 yêu cầu header `Authorization: Bearer <token>`.

### Auth

| Method | Route | `operationId` | Mô tả |
|---|---|---|---|
| `POST` | `/api/auth/register` | `auth-register` | Đăng ký tài khoản |
| `POST` | `/api/auth/login` | `auth-login` | Đăng nhập, nhận JWT |
| `POST` | `/api/auth/change-password` | `auth-change-password` | 🔒 Đổi mật khẩu |

### User

| Method | Route | `operationId` | Mô tả |
|---|---|---|---|
| `GET` | `/api/users/profile` | `user-get-profile` | 🔒 Xem profile cá nhân |
| `PATCH` | `/api/users/profile` | `user-update-profile` | 🔒 Cập nhật profile |
| `POST` | `/api/users/avatar/upload-url` | `user-get-avatar-upload-url` | 🔒 Lấy presigned URL upload avatar |
| `POST` | `/api/users/avatar/confirm` | `user-confirm-avatar-upload` | 🔒 Xác nhận upload avatar |

### Admin – User Management

| Method | Route | `operationId` | Mô tả |
|---|---|---|---|
| `GET` | `/api/admin/users` | `admin-list-users` | 🔒 Lấy danh sách tất cả users (phân trang) |
| `GET` | `/api/admin/users/{id}` | `admin-get-user` | 🔒 Lấy chi tiết user |
| `PATCH` | `/api/admin/users/{id}` | `admin-update-user` | 🔒 Cập nhật user |
| `DELETE` | `/api/admin/users/{id}` | `admin-delete-user` | 🔒 Xóa user (soft delete) |
| `POST` | `/api/admin/users/{id}/restore` | `admin-restore-user` | 🔒 Khôi phục user đã xóa |

### Categories

| Method | Route | `operationId` | Mô tả |
|---|---|---|---|
| `GET` | `/api/categories` | `category-list` | Lấy tất cả danh mục (phân trang) |
| `GET` | `/api/categories/{id}` | `category-get` | Lấy chi tiết danh mục |
| `POST` | `/api/categories` | `category-create` | Tạo danh mục mới |
| `PATCH` | `/api/categories/{id}` | `category-update` | Cập nhật danh mục |
| `DELETE` | `/api/categories/{id}` | `category-delete` | Xóa danh mục |

### Items

| Method | Route | `operationId` | Mô tả |
|---|---|---|---|
| `GET` | `/api/items/{id}` | `item-get` | Xem chi tiết sản phẩm |
| `POST` | `/api/items` | `item-create` | 🔒 Đăng sản phẩm mới |
| `PATCH` | `/api/items/{id}` | `item-update` | Cập nhật sản phẩm |
| `DELETE` | `/api/items/{id}` | `item-delete` | Xóa sản phẩm |
| `GET` | `/api/users/me/items` | `item-list-mine` | 🔒 Xem kho đồ của tôi (`?status=available\|traded\|...`) |
| `POST` | `/api/items/images/presign` | `item-presign-image` | Lấy presigned URL upload ảnh lên R2 |
| `POST` | `/api/items/{id}/images` | `item-confirm-image` | Xác nhận upload ảnh, lưu vào DB |
| `DELETE` | `/api/item-images/{imageID}` | `item-delete-image` | Xóa ảnh sản phẩm |

### Offers (Barter / Trao đổi)

| Method | Route | `operationId` | Actor | Mô tả |
|---|---|---|---|---|
| `POST` | `/api/items/{id}/offers` | `offer-create` | 🔒 Buyer | Gửi đề nghị trao đổi cho 1 item |
| `GET` | `/api/users/me/offers` | `offer-list-sent` | 🔒 Buyer | Xem lịch sử offer đã gửi |
| `GET` | `/api/items/{id}/offers` | `offer-list-received` | 🔒 Seller | Xem tất cả offer đến item của mình |
| `PATCH` | `/api/offers/{id}/status` | `offer-update-status` | 🔒 Seller | Duyệt hoặc từ chối offer |

#### Offer – Business Rules

- **Tạo offer:** Item đích phải ở trạng thái `available`; không thể offer cho item của chính mình.
- **Accept offer (transaction):**
  1. Offer được chọn → `accepted`
  2. Tất cả offers `pending` còn lại của cùng TargetItem → `rejected` (tự động)
  3. TargetItem → `traded`
  4. OfferItem (nếu có, trường hợp barter) → `traded`
- **Payload PATCH status:** `{ "status": "accepted" }` hoặc `{ "status": "rejected" }`

#### Offer – Các kịch bản barter

| Kịch bản | `offer_item_id` | `cash_amount` |
|---|---|---|
| Đổi item lấy item | `!= null` | `0` |
| Đổi item + thêm tiền | `!= null` | `> 0` |
| Mua bằng tiền mặt | `null` | `> 0` |

---

## OpenAPI 3.1 & Type Generation

### Tái sinh API docs

```bash
# Cần cài swag v2 (lần đầu)
go install github.com/swaggo/swag/v2/cmd/swag@latest

# Tái sinh từ godoc annotations
go run github.com/swaggo/swag/v2/cmd/swag@latest init \
  -g cmd/api/main.go \
  --v3.1 \
  --output docs
```

Output: `docs/swagger.json` (OpenAPI 3.1), `docs/swagger.yaml`, `docs/docs.go`.

> **Lưu ý:** Mỗi handler function phải có `@ID` annotation riêng (ví dụ `// @ID auth-login`)
> để `operationId` xuất hiện trong spec. Thiếu `@ID` → `openapi-typescript` sẽ tạo ra
> `operations: Record<string, never>` thay vì các types cụ thể.

### Tái sinh TypeScript types (frontend)

```bash
# Chạy từ thư mục frontend/
pnpm dlx openapi-typescript ../backend/docs/swagger.json -o ./src/schemas/types.d.ts
```

Sau khi gen, dùng type-safe operations trong frontend:

```typescript
import type { operations, components } from '@/schemas/types.d.ts'

// Request/Response types theo operationId
type LoginBody     = operations['auth-login']['requestBody']['content']['application/json']
type LoginResponse = operations['auth-login']['responses'][200]['content']['application/json']

// Schema types
type ItemResponse = components['schemas']['dto.ItemResponse']
```

### Tái sinh gorm/gen query

```bash
go run ./cmd/generate/
```

File output được tạo tự động vào `internal/query/` (DO NOT EDIT thủ công).

---

## Swagger UI

Swagger JSON được serve tại: `GET /docs/swagger.json`

Import vào **Postman** hoặc mở bằng **Swagger UI** / **Scalar** để xem interactive docs.
