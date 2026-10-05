package fetchsms

import (
	"context"
	"net/http"
)

type VerificationMessage struct {
	ID         string    `json:"id"`
	Code       string    `json:"code"`
	ReceivedAt Timestamp `json:"received_at"`
}

func (c *Client) VerificationMessages(ctx context.Context, id string) ([]VerificationMessage, error) {
	p, e := resourcePath("verifications", id, "/messages")
	if e != nil {
		return nil, e
	}
	var v []VerificationMessage
	e = c.do(ctx, http.MethodGet, p, nil, nil, &v, true)
	return v, e
}
