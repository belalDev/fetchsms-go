package fetchsms

import (
	"context"
	"net/http"
)

// ReuseVerification purchases a new window on the original number. No automatic retries.
func (c *Client) ReuseVerification(ctx context.Context, id string) (Verification, error) {
	var v Verification
	p, e := resourcePath("verifications", id, "/reuse")
	if e != nil {
		return v, e
	}
	e = c.do(ctx, http.MethodPost, p, nil, nil, &v, true)
	return v, e
}
