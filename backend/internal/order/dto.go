package order

import "time"

type CheckoutRequest struct {
	PaymentMethod string `json:"payment_method"`
}

type OrderResponse struct {
	ID              uint      `json:"id"`
	BuyerID         uint      `json:"buyer_id"`
	SellerID        uint      `json:"seller_id"`
	ItemID          uint      `json:"item_id"`
	TradeInItemID   *uint     `json:"trade_in_item_id,omitempty"`
	CashAmount      uint      `json:"cash_amount"`
	PaymentMethod   string    `json:"payment_method"`
	PaymentStatus   string    `json:"payment_status"`
	OrderStatus     string    `json:"order_status"`
	ItemName        string    `json:"item_name"`
	TradeInItemName string    `json:"trade_in_item_name,omitempty"`
	BuyerName       string    `json:"buyer_name"`
	SellerName      string    `json:"seller_name"`
	CreatedAt       time.Time `json:"created_at"`
}
