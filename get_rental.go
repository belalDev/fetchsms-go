package fetchsms

import (
	"context"
	"net/http"
)

func (c *Client) GetRental(ctx context.Context, id string) (Rental, error) {
	var v Rental
	p, e := resourcePath("rentals", id, "")
	if e != nil {
		return v, e
	}
	e = c.do(ctx, http.MethodGet, p, nil, nil, &v, true)
	return v, e
}
