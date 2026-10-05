package fetchsms

import (
	"context"
	"net/http"
)

func (c *Client) ListVerifications(ctx context.Context, tab Tab) ([]Verification, error) {
	q, e := tabQuery(tab)
	if e != nil {
		return nil, e
	}
	var v []Verification
	e = c.do(ctx, http.MethodGet, "/verifications", q, nil, &v, true)
	return v, e
}
