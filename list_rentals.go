package fetchsms

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

// ListRentalsRequest defaults to active and the server's default limit (100).
type ListRentalsRequest struct {
	Tab   Tab
	Limit int
}

func (c *Client) ListRentals(ctx context.Context, r ListRentalsRequest) ([]Rental, error) {
	q, e := tabQuery(r.Tab)
	if e != nil {
		return nil, e
	}
	if r.Limit < 0 || r.Limit > 200 {
		return nil, errors.New("fetchsms: limit must be 1–200 or zero for default")
	}
	if r.Limit != 0 {
		q.Set("limit", strconv.Itoa(r.Limit))
	}
	var v []Rental
	e = c.do(ctx, http.MethodGet, "/rentals", q, nil, &v, true)
	return v, e
}
