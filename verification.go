package fetchsms

import (
	"context"
	"errors"
	"net/http"
	"regexp"
)

// Status is extensible so future server statuses remain visible.
type Status string

const (
	StatusWaiting   Status = "waiting"
	StatusReceived  Status = "received"
	StatusExpired   Status = "expired"
	StatusCancelled Status = "cancelled"
)

type Verification struct {
	ID             string    `json:"id"`
	Ref            string    `json:"ref"`
	Service        string    `json:"service"`
	ServiceName    string    `json:"service_name"`
	Number         string    `json:"number"`
	Code           *string   `json:"code"`
	CodeCount      int       `json:"code_count"`
	Status         Status    `json:"status"`
	CostCents      int64     `json:"cost_cents"`
	CreatedAt      Timestamp `json:"created_at"`
	ExpiresAt      Timestamp `json:"expires_at"`
	Reusable       bool      `json:"reusable"`
	ReuseCostCents *int64    `json:"reuse_cost_cents"`
	ReuseNote      *string   `json:"reuse_note"`
}
type CreateVerificationRequest struct {
	Service  ServiceRef `json:"service"`
	AreaCode string     `json:"area_code,omitempty"`
}

var areaPattern = regexp.MustCompile(`^[0-9]{3}$`)

func validArea(s string) bool { return s == "" || areaPattern.MatchString(s) }

// CreateVerification charges the wallet. It is never automatically retried.
func (c *Client) CreateVerification(ctx context.Context, r CreateVerificationRequest) (Verification, error) {
	var v Verification
	if !r.Service.valid() || !validArea(r.AreaCode) {
		return v, errors.New("fetchsms: invalid verification request")
	}
	err := c.do(ctx, http.MethodPost, "/verifications", nil, r, &v, true)
	return v, err
}
