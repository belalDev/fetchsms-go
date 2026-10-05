package fetchsms

import (
	"context"
	"time"
)

// SimpleClient provides number purchase and code polling without caller contexts.
// It is safe for concurrent use. Create it with New; use NewClient for configuration.
type SimpleClient struct {
	client                     *Client
	numberTimeout, codeTimeout time.Duration
}

// New creates a client with default settings and makes no network requests.
// A blank key is rejected locally when an authenticated method is called.
func New(key string) *SimpleClient {
	// With no options, NewClient cannot return a configuration error.
	client, _ := NewClient(key)
	return &SimpleClient{client: client, numberTimeout: 30 * time.Second, codeTimeout: 15 * time.Minute}
}

// GetCode polls immediately, then every three seconds, for up to 15 minutes
// from this call. Each HTTP request has a 30-second timeout. HTTP 429 backoff
// honors Retry-After; other errors return immediately. The server may expire
// the verification sooner. It does not trigger an external SMS, extend the
// verification, cancel it on timeout, or guarantee a refund.
func (c *SimpleClient) GetCode(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.codeTimeout)
	defer cancel()
	return c.client.WaitForCode(ctx, id)
}

// GetNumber buys a verification number for a service slug, charging wallet funds.
// Each call has an internal 30-second context timeout and never retries.
// Reconcile an ambiguous failure before purchasing again. It does not fetch
// a quote/balance or trigger an SMS at the external service.
func (c *SimpleClient) GetNumber(service string) (Verification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.numberTimeout)
	defer cancel()
	return c.client.GetNumber(ctx, service)
}
