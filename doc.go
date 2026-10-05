// Package fetchsms provides simple and context-aware clients for the Fetch SMS
// v1 public catalog and API-key endpoints. Money is integer US cents. Creating or reusing
// verifications and creating or extending rentals spends wallet funds.
//
// New(key) returns a SimpleClient: GetNumber(service) buys a number with a
// 30-second per-call deadline; GetCode(id) polls immediately, then every three
// seconds, for up to 15 minutes from the call. Each HTTP request has a 30-second
// timeout. Neither method triggers an external SMS. Request that SMS at the
// chosen service using the purchased number. Polling does not extend the server's
// verification lifetime or cancel/refund an order on timeout.
//
// NewClient(key, options...) retains the full context-aware API and configuration.
// Its GetNumber(ctx, service) buys a number; it is not a read-only lookup.
// Use CreateVerification for numeric service IDs or an area-code preference,
// GetVerification to read an order, and WaitForCode after triggering an SMS
// at the external service.
//
// No mutating call is automatically retried. After a transport error, reconcile
// account state before deciding whether to repeat a purchase. WaitForCode and
// WaitForRentalCode retry only HTTP 429 responses, respecting Retry-After.
// With Client, supply a deadline for polling; nil contexts mean context.Background.
//
// Verify webhook signatures on the raw request body before DecodeWebhook.
// Signatures alone do not prevent replay.
package fetchsms
