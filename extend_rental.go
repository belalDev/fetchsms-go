package fetchsms

import (
	"context"
	"errors"
	"net/http"
)

// ExtendRental charges for an additional term; it is never automatically retried.
func (c *Client) ExtendRental(ctx context.Context, id string, days int) (Rental, error) {
	var v Rental
	if !validDays(days) {
		return v, errors.New("fetchsms: invalid rental term")
	}
	p, e := resourcePath("rentals", id, "/extend")
	if e != nil {
		return v, e
	}
	e = c.do(ctx, http.MethodPost, p, nil, struct {
		Days int `json:"days"`
	}{days}, &v, true)
	return v, e
}
