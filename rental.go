package fetchsms

import (
	"context"
	"errors"
	"net/http"
)

type Rental struct {
	ID             string    `json:"id"`
	Ref            string    `json:"ref"`
	Service        string    `json:"service"`
	ServiceName    string    `json:"service_name"`
	Number         string    `json:"number"`
	DurationDays   int       `json:"duration_days"`
	ExtendedDays   int       `json:"extended_days"`
	CustomAreaCode bool      `json:"custom_area_code"`
	Status         Status    `json:"status"`
	LastCode       *string   `json:"last_code"`
	CostCents      int64     `json:"cost_cents"`
	Refunded       bool      `json:"refunded"`
	CreatedAt      Timestamp `json:"created_at"`
	ExpiresAt      Timestamp `json:"expires_at"`
	CancelDeadline Timestamp `json:"cancel_deadline"`
}
type CreateRentalRequest struct {
	Service  ServiceRef `json:"service"`
	Days     int        `json:"days"`
	AreaCode string     `json:"area_code,omitempty"`
}

// CreateRental charges the entire term up front; it is never automatically retried.
func (c *Client) CreateRental(ctx context.Context, r CreateRentalRequest) (Rental, error) {
	var v Rental
	if (r.Service.String() != "1" && r.Service.String() != "unlimited-services") || !validDays(r.Days) || !validArea(r.AreaCode) {
		return v, errors.New("fetchsms: invalid rental request")
	}
	e := c.do(ctx, http.MethodPost, "/rentals", nil, r, &v, true)
	return v, e
}
