package fetchsms

import (
	"context"
	"net/http"
)

// CancelVerification accepts any bounded successful body: the upstream return shape is undocumented.
func (c *Client) CancelVerification(ctx context.Context, id string) error {
	p, e := resourcePath("verifications", id, "/cancel")
	if e != nil {
		return e
	}
	return c.do(ctx, http.MethodPost, p, nil, nil, nil, true)
}
