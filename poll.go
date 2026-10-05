package fetchsms

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// WithPollInterval sets the delay between polls (default three seconds).
func WithPollInterval(d time.Duration) Option {
	return func(c *Client) error {
		if d <= 0 {
			return errors.New("fetchsms: poll interval must be positive")
		}
		c.pollInterval = d
		return nil
	}
}

// TerminalStatusError means polling cannot continue, including an unknown status.
type TerminalStatusError struct {
	ResourceID string
	Status     Status
}

func (e *TerminalStatusError) Error() string {
	return fmt.Sprintf("fetchsms: polling stopped with status %q", e.Status)
}
func wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// WaitForCode polls immediately, then at the configured interval. Always supply a deadline in production.
func (c *Client) WaitForCode(ctx context.Context, id string) (string, error) {
	return c.poll(ctx, id, func(ctx context.Context) (Status, *string, error) {
		v, e := c.GetVerification(ctx, id)
		return v.Status, v.Code, e
	})
}

// WaitForRentalCode returns the latest existing code, not necessarily a new one.
// Use RentalMessages or webhooks to track subsequent deliveries.
func (c *Client) WaitForRentalCode(ctx context.Context, id string) (string, error) {
	return c.poll(ctx, id, func(ctx context.Context) (Status, *string, error) {
		v, e := c.GetRental(ctx, id)
		return v.Status, v.LastCode, e
	})
}
func (c *Client) poll(ctx context.Context, id string, fetch func(context.Context) (Status, *string, error)) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		if e := ctx.Err(); e != nil {
			return "", e
		}
		status, code, e := fetch(ctx)
		if e != nil {
			var api *APIError
			if errors.As(e, &api) && api.StatusCode == 429 {
				d := c.pollInterval
				if api.RetryAfter > d {
					d = api.RetryAfter
				}
				if e := wait(ctx, d); e != nil {
					return "", e
				}
				continue
			}
			return "", e
		}
		switch status {
		case StatusReceived:
			if code != nil && *code != "" {
				return *code, nil
			}
		case StatusWaiting:
		default:
			return "", &TerminalStatusError{ResourceID: id, Status: status}
		}
		if e := wait(ctx, c.pollInterval); e != nil {
			return "", e
		}
	}
}
