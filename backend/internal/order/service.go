package order

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
)

var (
	ErrCartEmpty           = errors.New("Giỏ hàng trống")
	ErrItemNotFound        = errors.New("Sản phẩm không tồn tại hoặc đã bị xóa")
	ErrItemNotAvailable    = errors.New("Sản phẩm không còn sẵn sàng")
	ErrSelfPurchase        = errors.New("Bạn không thể mua sản phẩm của chính mình")
	ErrTradeInNotOwned     = errors.New("Sản phẩm trao đổi phải thuộc về bạn")
	ErrTradeInNotAvail     = errors.New("Sản phẩm trao đổi không còn sẵn sàng")
	ErrOrderNotFound       = errors.New("Không tìm thấy đơn hàng")
	ErrForbidden           = errors.New("Bạn không có quyền thực hiện hành động này")
	ErrOrderNotCancellable = errors.New("Đơn hàng không thể hủy ở trạng thái này")
	ErrWalletNotFound      = errors.New("Không tìm thấy ví")
	ErrInsufficientBalance = errors.New("Số dư ví không đủ")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Checkout(ctx context.Context, userID uint, paymentMethod string) (*models.Order, error) {
	if paymentMethod == "" {
		paymentMethod = "wallet"
	}
	return s.repo.CheckoutTx(ctx, userID, paymentMethod)
}

func (s *Service) GetOrders(ctx context.Context, userID uint) ([]*models.Order, error) {
	buyerOrders, err := s.repo.FindByBuyerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sellerOrders, err := s.repo.FindBySellerID(ctx, userID)
	if err != nil {
		return nil, err
	}

	all := make([]*models.Order, 0, len(buyerOrders)+len(sellerOrders))
	all = append(all, buyerOrders...)
	all = append(all, sellerOrders...)
	return all, nil
}

func (s *Service) GetOrderByID(ctx context.Context, userID uint, orderID uint) (*models.Order, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, ErrForbidden
	}
	return order, nil
}

func (s *Service) CancelOrder(ctx context.Context, userID uint, orderID uint) error {
	return s.repo.CancelOrderTx(ctx, userID, orderID)
}

func ToOrderResponse(o *models.Order) *OrderResponse {
	resp := &OrderResponse{
		ID:            o.ID,
		BuyerID:       o.BuyerID,
		SellerID:      o.SellerID,
		ItemID:        o.ItemID,
		TradeInItemID: o.TradeInItemID,
		CashAmount:    o.CashAmount,
		PaymentMethod: o.PaymentMethod,
		PaymentStatus: string(o.PaymentStatus),
		OrderStatus:   string(o.OrderStatus),
		CreatedAt:     o.CreatedAt,
	}
	if o.Item.ID != 0 {
		resp.ItemName = o.Item.Name
	}
	if o.TradeInItem != nil && o.TradeInItem.ID != 0 {
		resp.TradeInItemName = o.TradeInItem.Name
	}
	if o.Buyer.ID != 0 {
		resp.BuyerName = o.Buyer.Name
	}
	if o.Seller.ID != 0 {
		resp.SellerName = o.Seller.Name
	}
	return resp
}
