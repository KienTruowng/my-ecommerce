package appointment

import (
	"context"
	"errors"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
)

var (
	ErrNotFound     = errors.New("Không tìm thấy lịch hẹn")
	ErrForbidden    = errors.New("Bạn không có quyền thực hiện hành động này")
	ErrInvalidState = errors.New("Không thể cập nhật lịch hẹn ở trạng thái này")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMyAppointments(ctx context.Context, userID uint) ([]*models.Appointment, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *Service) UpdateAppointment(ctx context.Context, userID uint, id uint, req *UpdateAppointmentRequest) (*models.Appointment, error) {
	appt, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	var isBuyer, isSeller bool
	if appt.Order != nil {
		isBuyer = appt.Order.BuyerID == userID
		isSeller = appt.Order.SellerID == userID
	} else if appt.Offer != nil {
		isBuyer = appt.Offer.OfferUserID == userID
		isSeller = appt.Offer.TargetItem.UserID == userID
	}

	if !isBuyer && !isSeller {
		return nil, ErrForbidden
	}

	if appt.Status == models.ApptStatusCompleted || appt.Status == models.ApptStatusCancelled {
		return nil, ErrInvalidState
	}

	if req.Date != nil {
		appt.Date = *req.Date
	}
	if req.Time != nil {
		appt.Time = *req.Time
	}
	if req.Location != nil {
		appt.Location = *req.Location
	}
	if req.Note != nil {
		appt.Note = *req.Note
	}

	if req.Status != nil {
		newStatus := models.AppointmentStatus(*req.Status)
		if newStatus == models.ApptStatusCompleted {
			if err := s.repo.CompleteAppointmentTx(ctx, appt); err != nil {
				return nil, err
			}
			appt.Status = newStatus
			return appt, nil
		}
		appt.Status = newStatus
	}

	if err := s.repo.Update(ctx, appt); err != nil {
		return nil, err
	}

	return appt, nil
}
