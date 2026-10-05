package fetchsms

import (
	"context"
	"net/http"
)

type Balance struct {
	BalanceCents int64 `json:"balance_cents"`
}

func (c *Client) Balance(ctx context.Context) (Balance, error) {
	var v Balance
	e := c.do(ctx, http.MethodGet, "/wallet/balance", nil, nil, &v, true)
	return v, e
}
