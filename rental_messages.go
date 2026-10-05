package fetchsms

import (
	"context"
	"net/http"
)

type RentalMessage struct {
	ID         string    `json:"id"`
	Sender     string    `json:"sender"`
	Body       string    `json:"body"`
	Code       string    `json:"code"`
	ReceivedAt Timestamp `json:"received_at"`
}

func (c *Client) RentalMessages(ctx context.Context, id string) ([]RentalMessage, error) {
	p, e := resourcePath("rentals", id, "/messages")
	if e != nil {
		return nil, e
	}
	var v []RentalMessage
	e = c.do(ctx, http.MethodGet, p, nil, nil, &v, true)
	return v, e
}
