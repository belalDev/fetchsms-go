package fetchsms

import "context"

// GetNumber buys a verification number for a service slug, charging wallet funds.
// It validates service with ServiceSlug, then calls CreateVerification without an
// area-code preference. Use CreateVerification for numeric IDs or an area code.
// It never automatically retries. A failed response may still mean a charge;
// reconcile account state before repeating a purchase. It does not trigger an
// SMS at the external service or wait for a code; use WaitForCode after doing so.
func (c *Client) GetNumber(ctx context.Context, service string) (Verification, error) {
	ref, err := ServiceSlug(service)
	if err != nil {
		return Verification{}, err
	}
	return c.CreateVerification(ctx, CreateVerificationRequest{Service: ref})
}
