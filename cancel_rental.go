package fetchsms

import (
	"context"
	"net/http"
)

func (c *Client) CancelRental(ctx context.Context, id string) (Rental, error) {
	var v Rental
	p, e := resourcePath("rentals", id, "/cancel")
	if e != nil {
		return v, e
	}
	e = c.do(ctx, http.MethodPost, p, nil, nil, &v, true)
	return v, e
}
