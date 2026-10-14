# Ecommerce

Dự án ecommerce dạng monorepo gồm:

- Backend viết bằng Go, dùng `net/http`, `GORM`, PostgreSQL, JWT và Cloudflare R2.
- Frontend viết bằng Vue 3 + Vite + TypeScript, dùng Pinia, Vue Router và các component UI theo kiểu shadcn-vue.

## Tính năng chính

- Đăng ký, đăng nhập bằng email và mật khẩu.
- Xác thực bằng JWT.
- Đổi mật khẩu cho người dùng đã đăng nhập.
- Lấy và cập nhật thông tin profile người dùng.
- Upload avatar qua presigned URL lên Cloudflare R2.
- API quản trị người dùng: danh sách, xem chi tiết, cập nhật, xóa, khôi phục.
- Swagger JSON được phục vụ sẵn từ backend.
- Seed dữ liệu sẵn một số trường đại học.

## Cấu trúc thư mục

```text
backend/
frontend/
docker-compose.dev.yml
```

### Backend

- `cmd/api/main.go`: khởi động server, kết nối database, đăng ký route.
- `internal/auth`: đăng ký, đăng nhập, đổi mật khẩu, JWT middleware.
- `internal/user`: profile người dùng, admin user management, avatar flow.
- `internal/media`: tạo presigned URL và xác thực file upload.
- `internal/storage`: tích hợp Cloudflare R2.
- `internal/seeders`: seed dữ liệu khởi tạo.
- `docs/swagger.json`: tài liệu Swagger phục vụ trực tiếp.

### Frontend

- `src/router/index.ts`: định tuyến và guard đăng nhập.
- `src/api/api.ts`: OpenAPI client, tự gắn `Authorization` token từ `localStorage`.
- `src/views`: các trang login, signup, settings và dashboard.
- `src/components`: UI components, form, sidebar, data table.

## Yêu cầu môi trường

- Go `1.25.7` hoặc tương thích.
- Node.js `^20.19.0 || >=22.12.0`.
- pnpm.
- PostgreSQL.
- Cloudflare R2 credentials.

## Biến môi trường backend

Backend đọc biến môi trường từ `.env` hoặc biến hệ thống.

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ecommerce
DB_SSLMODE=disable
JWT_SECRET=your-secret-key

R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY=your-access-key
R2_SECRET_KEY=your-secret-key
R2_BUCKET=your-bucket
R2_PUBLIC_URL=https://your-public-r2-url
```

## Chạy bằng Docker

File `docker-compose.dev.yml` hiện chỉ dựng backend và PostgreSQL.

```bash
docker compose -f docker-compose.dev.yml up --build
```

Mặc định:

- Backend: `http://localhost:3000`
- PostgreSQL container: port `5431` trên máy host

## Chạy backend local

```bash
cd backend
go mod download
go run ./cmd/api
```

Backend sẽ:

- load `.env` nếu có
- migrate bảng `universities` và `users`
- seed danh sách trường đại học
- chạy HTTP server tại port `3000`

## Chạy frontend local

```bash
cd frontend
pnpm install
pnpm dev
```

Mặc định Vite sẽ chạy ở `http://localhost:5173`.

## Build và kiểm tra frontend

```bash
cd frontend
pnpm build
pnpm test:unit
pnpm lint
```

## API backend

### Auth

| Method | Route                       | Mô tả                                    |
| ------ | --------------------------- | ---------------------------------------- |
| POST   | `/api/auth/register`        | Đăng ký tài khoản mới                    |
| POST   | `/api/auth/login`           | Đăng nhập bằng email và mật khẩu         |
| POST   | `/api/auth/change-password` | Đổi mật khẩu cho người dùng đã đăng nhập |

### User

| Method | Route                          | Mô tả                              |
| ------ | ------------------------------ | ---------------------------------- |
| GET    | `/api/users/profile`           | Lấy thông tin profile hiện tại     |
| PATCH  | `/api/users/profile`           | Cập nhật profile                   |
| POST   | `/api/users/avatar/upload-url` | Lấy presigned URL để upload avatar |
| POST   | `/api/users/avatar/confirm`    | Xác nhận upload avatar thành công  |

### Admin user

| Method | Route                           | Mô tả                            |
| ------ | ------------------------------- | -------------------------------- |
| GET    | `/api/admin/users`              | Lấy danh sách user có phân trang |
| GET    | `/api/admin/users/{id}`         | Lấy chi tiết một user            |
| PATCH  | `/api/admin/users/{id}`         | Cập nhật user                    |
| DELETE | `/api/admin/users/{id}`         | Xóa user                         |
| POST   | `/api/admin/users/{id}/restore` | Khôi phục user đã xóa            |

### Swagger

| Method | Route                | Mô tả             |
| ------ | -------------------- | ----------------- |
| GET    | `/docs/swagger.json` | File Swagger JSON |

## Frontend routes

| Route                       | Mô tả                  |
| --------------------------- | ---------------------- |
| `/login`                    | Trang đăng nhập        |
| `/signup`                   | Trang đăng ký          |
| `/`                         | Trang chủ              |
| `/settings/profile`         | Cập nhật profile       |
| `/settings/change-password` | Đổi mật khẩu           |
| `/dashboard`                | Dashboard quản trị     |
| `/dashboard/users`          | Quản lý danh sách user |

Các route có `requiresAuth` sẽ tự chuyển về trang đăng nhập nếu chưa có token trong `localStorage`.

## Ghi chú

- Frontend hiện trỏ API về `http://localhost:3000` trong `src/api/api.ts`.
- Nếu đổi port hoặc domain backend, cập nhật lại `baseUrl` trong file đó.
- Trang home hiện tại là placeholder, còn phần quản trị user đã có bảng, sửa và xóa ở dashboard.

## Tổng quan kỹ thuật

- Backend dùng `http.ServeMux` và inject service theo từng module.
- User model có liên kết với `University`.
- Avatar upload đi theo flow: lấy presigned URL -> upload lên R2 -> confirm upload -> lưu URL vào database.
