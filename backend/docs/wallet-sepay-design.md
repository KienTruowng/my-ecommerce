# Wallet + SePay Top-up Design

Tài liệu này phác thảo hướng thiết kế cho tính năng ví nội bộ, nạp tiền bằng SePay webhook, và luồng mua online kết hợp barter + cash trong backend Go hiện tại.

## Mục tiêu

- Tách wallet thành một domain riêng, không nhét balance trực tiếp vào `User`.
- Dùng SePay webhook làm nguồn sự thật cho giao dịch nạp tiền.
- Cho phép user nạp tiền, xem số dư, và xem lịch sử giao dịch.
- Cho phép luồng mua online trừ tiền từ wallet khi phần cash amount phát sinh trong offer.
- Giữ logic trao đổi item hiện tại ổn định, chỉ mở rộng ở nhánh cash.

## Non-goals giai đoạn đầu

- Không triển khai frontend ngay.
- Không tích hợp thêm gateway khác ngoài SePay.
- Không xử lý payout/rút tiền.
- Không làm settlement phức tạp kiểu escrow đa bên ở giai đoạn đầu.

## Kiến trúc đề xuất

Tách một module mới theo cùng kiểu với các module hiện có:

- `internal/wallet/handler.go`
- `internal/wallet/service.go`
- `internal/wallet/repository.go`
- `internal/wallet/dto/`
- `internal/models/wallet_account.go`
- `internal/models/wallet_transaction.go`

Luồng chính sẽ đi theo mô hình hiện có của repo:

`Handler -> Service -> Repository -> GORM`

## Mô hình dữ liệu

### WalletAccount

Đại diện cho ví của một user.

Trường đề xuất:

- `ID`
- `UserID` unique
- `Balance` bigint / integer
- `Currency` mặc định `VND`
- `Status` active / frozen / closed
- `CreatedAt`, `UpdatedAt`, `DeletedAt`

### WalletTransaction

Nhật ký bất biến cho mọi biến động số dư.

Trường đề xuất:

- `ID`
- `WalletAccountID`
- `UserID`
- `Type` = `topup`, `debit`, `refund`, `adjustment`
- `Amount`
- `Direction` = `in`, `out`
- `ReferenceType` = `sepay_webhook`, `offer_cash`, `admin_adjustment`
- `ReferenceID`
- `ExternalID` unique, dùng cho SePay transaction `id`
- `Status` = `pending`, `completed`, `failed`, `reversed`
- `Metadata` jsonb
- `CreatedAt`, `UpdatedAt`

### WalletTopupRequest

Nếu cần chủ động tạo mã nạp tiền trước webhook, có thể thêm bảng phụ:

- `ID`
- `UserID`
- `ExpectedAmount`
- `PaymentCode`
- `Status` = `pending`, `matched`, `expired`
- `ExpiresAt`

Ở giai đoạn đầu có thể chưa cần bảng này nếu SePay content/code đã đủ để đối soát.

## Luồng SePay top-up

### 1. Nhận webhook

SePay gửi payload đến endpoint public, ví dụ:

- `POST /api/payments/sepay/webhook`

Yêu cầu xử lý:

- Parse payload nhanh.
- Xác định giao dịch `in` thôi.
- Ở test mode thì không cần xác thực HMAC.
- Nếu sau này chuyển production, mới bật xác thực HMAC ở lớp webhook.
- Chặn trùng bằng `SePay.id` hoặc `referenceCode` tùy cấu hình, nhưng `id` là khóa chính chống trùng.
- Kiểm tra `transferAmount > 0`.
- Tìm user/wallet đích theo `code`, `subAccount`, hoặc rule mapping trong `content`.
- Ghi `WalletTransaction` trong một DB transaction.
- Cộng `WalletAccount.Balance` trong cùng transaction.

### 2. Phản hồi cho SePay

Trả ngay `200` hoặc `201` với body chính xác:

```json
{"success": true}
```

Phần xử lý nặng phải hoàn tất trong một transaction ngắn hoặc đẩy vào job nội bộ nếu sau này cần mở rộng.

## Luồng mua online bằng cash

### Biến thể giữ offer hiện tại

`Offer` đã hỗ trợ:

- barter item for item
- barter + cash
- cash only

Đề xuất ở giai đoạn đầu:

- Giữ `Offer` như luồng thương lượng.
- Khi seller accept offer và `CashAmount > 0`, hệ thống tạo một bản ghi thanh toán từ ví của buyer.
- Nếu buyer không đủ số dư, offer chuyển sang trạng thái chờ thanh toán hoặc bị từ chối theo rule sản phẩm.

### Một transaction chuẩn khi accept offer

1. Lock offer + target item + wallet account liên quan.
2. Nếu `CashAmount > 0`, tạo `WalletTransaction` dạng debit.
3. Trừ balance của buyer.
4. Chuyển item sang `traded` hoặc `reserved` tùy business rule hiện tại.
5. Ghi log sang `ExchangeLog` nếu cần trace hành vi.

## API đề xuất

### Public / user

- `GET /api/wallet` - xem số dư và trạng thái ví
- `GET /api/wallet/transactions` - xem lịch sử giao dịch
- `POST /api/wallet/topup/request` - tạo yêu cầu nạp tiền nếu cần sinh mã đối soát trước

### Webhook

- `POST /api/payments/sepay/webhook` - nhận callback từ SePay

### Admin

- `GET /api/admin/wallets` - tra cứu ví
- `POST /api/admin/wallets/{id}/adjust` - điều chỉnh số dư thủ công

## Env vars đề xuất

Thêm vào `internal/config.Config`:

- `SEPAY_WEBHOOK_SECRET` (optional, chỉ dùng khi bật HMAC ở production)
- `SEPAY_MERCHANT_ID` hoặc biến tương đương nếu cần map tài khoản
- `SEPAY_ENABLED`

Nếu SePay dùng nhiều tài khoản nhận tiền, có thể thêm bảng mapping thay vì nhét toàn bộ vào env.

## Quy tắc idempotency

- Mỗi webhook SePay chỉ được xử lý một lần.
- `WalletTransaction.ExternalID` nên unique.
- Nếu ghi trùng, service trả thành công nhưng không cộng tiền lại.
- Tất cả thay đổi balance phải nằm trong cùng transaction với bản ghi transaction ledger.

## Bảo mật và kiểm soát

- Chỉ chấp nhận `transferType = in` cho top-up.
- Không tin `content` nếu không có mapping rõ ràng.
- Bỏ qua HMAC trong test mode; chỉ verify khi production yêu cầu.
- Log payload tối thiểu, không lưu thông tin nhạy cảm thừa.
- Nếu cần xử lý queue sau này, webhook vẫn phải trả response nhanh trước.

## Cách nối với code hiện có

Điểm cần đụng nhẹ nhất:

- `cmd/api/main.go`: đăng ký module wallet và webhook route.
- `internal/config/config.go`: thêm biến env cho SePay.
- `internal/models/`: thêm wallet models.
- `internal/offer/`: chèn debit wallet khi accept offer có cash.
- `internal/exchangelog/`: log các sự kiện wallet nếu muốn thống nhất audit trail.

## Thứ tự triển khai đề xuất

1. Thêm wallet models + migration.
2. Thêm wallet repository/service cơ bản.
3. Thêm SePay webhook handler tối thiểu, idempotent.
4. Gắn top-up vào wallet ledger.
5. Gắn debit cash vào offer accept flow.
6. Bổ sung API đọc số dư/lịch sử giao dịch.

## Câu hỏi còn mở

- Mã đối soát SePay sẽ map user bằng `content`, `code`, hay bảng request riêng?
- Cash top-up trong offer có cần giữ trạng thái pending khi ví chưa đủ tiền không?
- Có cần cho admin cộng/trừ thủ công ngay từ bản đầu không?
