# Ecommerce Checkout Refactoring — Implementation Plan

## Summary

Replace the peer-to-peer offer/swap system with an ecommerce-style cart → checkout → order flow.
Single-item cart, optional barter trade-in at checkout, wallet payment integration.

---

## Phase 1: New Models

### 1.1 Create `backend/internal/models/cart.go`
- Single-item cart (UserID unique), holds ItemID, optional TradeInItemID, CashAmount
- GORM model with foreign keys to User, Item

### 1.2 Create `backend/internal/models/order.go`
- OrderStatus enum: confirmed, completed, cancelled
- PaymentStatus enum: pending, completed, failed, refunded
- Order struct: BuyerID, SellerID, ItemID, TradeInItemID (nullable), CashAmount, PaymentMethod ("wallet"/"cod"), PaymentStatus, OrderStatus
- HasOne Appointment

### 1.3 Update `backend/internal/models/appointment.go`
- Change `OfferID uint` → `OrderID uint`
- Change `Offer Offer` → `Order Order`
- Add `gorm:"foreignKey:OrderID"` tag on Order

### 1.4 Update `backend/internal/models/item.go`
- Add `Orders []*Order` field (has-many, for query preloading)

---

## Phase 2: Cart Module

### 2.1 Create `backend/internal/cart/repository.go`
```go
type Repository interface {
    GetCart(ctx, userID) (*models.Cart, error)        // preload Item, TradeInItem, item owner
    UpsertCart(ctx, cart) error                        // upsert by user_id (single-item)
    ClearCart(ctx, userID) error                       // delete by user_id
}
```
Implementation: raw GORM (no query gen — Cart isn't in gen)

### 2.2 Create `backend/internal/cart/service.go`
- Validate item exists, is available, not self-cart
- Validate trade-in item belongs to buyer and is available
- GetCart (for checkout page display)

### 2.3 Create `backend/internal/cart/handler.go`
- `POST /api/cart` — AddToCart (body: {item_id, offer_item_id?, cash_amount?})
- `GET /api/cart` — GetCart
- `DELETE /api/cart` — RemoveFromCart
- `PATCH /api/cart` — UpdateCart (change trade-in item / cash amount)

Sentinel errors: ErrItemNotFound, ErrItemNotAvailable, ErrSelfCart, ErrCartEmpty

---

## Phase 3: Order Module

### 3.1 Create `backend/internal/order/dto.go`
- CheckoutRequest: PaymentMethod string
- OrderResponse: full order with buyer/seller/item/trade-in item details

### 3.2 Create `backend/internal/order/repository.go`
```go
type Repository interface {
    CheckoutTx(ctx, userID, paymentMethod) (*models.Order, error)
    FindByID(ctx, orderID) (*models.Order, error)
    FindByBuyerID(ctx, userID) ([]*models.Order, error)
    FindBySellerID(ctx, userID) ([]*models.Order, error)
    CancelOrderTx(ctx, userID, orderID) error
    CompleteOrderTx(ctx, tx/gorm.DB, order) error   // called from appointment completion
}
```

**CheckoutTx flow (single DB transaction):**
1. Load user's cart (preload Item with owner)
2. Validate item still available, not self-purchase
3. If TradeInItemID set, validate it belongs to buyer and is available
4. Create Order: BuyerID=buyer, SellerID=item.UserID, ItemID, TradeInItemID, CashAmount, OrderStatus=confirmed, PaymentStatus=pending
5. If CashAmount > 0 and PaymentMethod == "wallet":
   - Find buyer's wallet (SELECT FOR UPDATE lock)
   - Check sufficient balance
   - Deduct balance
   - Create WalletTransaction (type=debit, direction=out, reference_type=order_payment, external_id=order-{id}-payment)
   - Set PaymentStatus = completed
6. If CashAmount == 0: PaymentStatus = completed (no payment needed)
7. Mark Item status → "reserved" (and TradeInItem if present)
8. Auto-create Appointment linked to Order (status: scheduled)
9. Delete cart
10. Return created Order

**CancelOrderTx flow:**
1. Load order, verify buyer is canceller and status is "confirmed"
2. Mark order cancelled
3. Restore item statuses → "available"
4. If CashAmount > 0 and payment was completed:
   - Refund buyer's wallet (credit + create refund transaction)
   - Set PaymentStatus = refunded
5. Cancel linked appointment

### 3.3 Create `backend/internal/order/service.go`
- Checkout: call CheckoutTx, send notification to seller
- GetOrders: combine buyer+seller orders for user
- GetOrder: single order detail
- CancelOrder: validate, call CancelOrderTx, notify seller

### 3.4 Create `backend/internal/order/handler.go`
- `POST /api/checkout` — Checkout (body: {payment_method})
- `GET /api/orders` — GetOrders (my orders as buyer and seller)
- `GET /api/orders/{id}` — GetOrder
- `POST /api/orders/{id}/cancel` — CancelOrder

Sentinel errors: ErrCartEmpty, ErrItemNotAvailable, ErrSelfPurchase, ErrOrderNotFound, ErrOrderNotCancellable, ErrInsufficientBalance

---

## Phase 4: Wallet Route Registration

### 4.1 Add `setupWalletModule` in `backend/cmd/api/main.go`
```go
func setupWalletModule(db *gorm.DB, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
    walletRepo := wallet.NewRepository(db)
    walletService := wallet.NewService(walletRepo)
    walletHandler := wallet.NewHandler(walletService)
    requireAuth := authMid.RequireAuth

    mux.HandleFunc("GET /api/wallet", requireAuth(walletHandler.GetWalletSummary))
    mux.HandleFunc("GET /api/wallet/transactions", requireAuth(walletHandler.GetWalletTransactions))
    mux.HandleFunc("POST /api/wallet/topup/request", requireAuth(walletHandler.CreateTopupRequest))
    mux.HandleFunc("POST /api/payments/sepay/webhook", walletHandler.SePayWebhook)
}
```

### 4.2 Add wallet models to AutoMigrate
Add `&models.WalletAccount{}`, `&models.WalletTransaction{}`, `&models.WalletTopupRequest{}` to the AutoMigrate call.

---

## Phase 5: Update Appointment Module

### 5.1 Update `backend/internal/models/appointment.go`
```go
type Appointment struct {
    gorm.Model
    OrderID  uint  `gorm:"not null;index"`
    Order    Order `gorm:"foreignKey:OrderID"`
    // ... rest unchanged
}
```

### 5.2 Update `backend/internal/appointment/dto.go`
- `AppointmentResponse.OfferID` → `OrderID`
- Partner name logic: use Order.BuyerID / Order.SellerID instead of Offer

### 5.3 Update `backend/internal/appointment/repository.go`
**Repository interface:** Update to reference Order instead of Offer.
- `FindByID`: preload "Order" + "Order.Item" + "Order.TradeInItem" + "Order.Buyer" + "Order.Seller"
- `FindByUserID`: JOIN orders instead of offers, check buyer_id OR seller_id
- `CompleteAppointmentTx`: 
  1. Mark appointment completed
  2. Transfer Item ownership (user_id → buyer, status → traded)
  3. If TradeInItem exists: transfer to seller (status → traded)
  4. If CashAmount > 0: credit seller's wallet, create WalletTransaction
  5. Mark Order as completed

### 5.4 Update `backend/internal/appointment/service.go`
- Use Order fields instead of Offer fields throughout
- Pass correct references for Order-based logic

### 5.5 Update `backend/internal/appointment/handler.go`
- Partner name logic: use order.Buyer / order.Seller

---

## Phase 6: Update Notification Service

### 6.1 Add to `backend/internal/notify/service.go`
```go
func (s *Service) OrderPlaced(ctx context.Context, order *models.Order) error     // notify seller
func (s *Service) OrderCancelled(ctx context.Context, order *models.Order) error  // notify seller
func (s *Service) OrderCompleted(ctx context.Context, order *models.Order) error  // notify buyer
```

### 6.2 Update `backend/internal/notify/handler.go`
- `PendingOfferCounts` → remove or rename (depends on UI needs)
- Or replace with `PendingOrderCounts` tracking confirmed orders

---

## Phase 7: Update main.go

### 7.1 AutoMigrate additions
Add: `&models.Cart{}`, `&models.Order{}`, `&models.WalletAccount{}`, `&models.WalletTransaction{}`, `&models.WalletTopupRequest{}`

### 7.2 New module setup functions
- `setupCartModule(db, authMid, mux)`
- `setupOrderModule(db, authMid, notifySvc, mux)`
- `setupWalletModule(db, authMid, mux)`

### 7.3 Remove offer module setup
- Delete `setupOfferModule` call
- Keep `setupAppointmentModule` (updated)

### 7.4 Update AdminStats
Replace `PendingOffers int64` with `PendingOrders int64` (count confirmed orders).
Keep `CompletedTrades` but query orders instead of offers.

### 7.5 Update imports
Add: `cart`, `order`, `wallet` packages
Remove or keep: `offer` (just remove its setup call)

---

## Phase 8: Cleanup

### 8.1 Remove offer module files
- `backend/internal/offer/handler.go`
- `backend/internal/offer/service.go`
- `backend/internal/offer/repository.go`

### 8.2 Verify build
```bash
cd backend && go build ./cmd/api
```
Fix any compilation errors.

### 8.3 Notes
- Offer model stays in `models/offer.go` (historical data, auto-generated query gen depends on it)
- Offer table stays in DB (existing data preserved)
- Generated query code (`query/gen.go`, `query/offers.gen.go`) stays as-is
