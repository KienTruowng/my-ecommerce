package appointment

import "time"

type AppointmentResponse struct {
	ID       uint   `json:"id"`
	OfferID  *uint  `json:"offer_id,omitempty"`
	OrderID  *uint  `json:"order_id,omitempty"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Time     string `json:"time"`
	Location string `json:"location"`
	Note     string `json:"note"`
	Status   string `json:"status"`

	PartnerName string    `json:"partner"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpdateAppointmentRequest struct {
	Date     *string `json:"date,omitempty"`
	Time     *string `json:"time,omitempty"`
	Location *string `json:"location,omitempty"`
	Note     *string `json:"note,omitempty"`
	Status   *string `json:"status,omitempty"`
}
