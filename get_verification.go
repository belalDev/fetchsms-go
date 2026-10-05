package fetchsms

import (
	"context"
	"net/http"
)

func (c *Client) GetVerification(ctx context.Context, id string) (Verification, error) {
	var v Verification
	p, e := resourcePath("verifications", id, "")
	if e != nil {
		return v, e
	}
	e = c.do(ctx, http.MethodGet, p, nil, nil, &v, true)
	return v, e
}
