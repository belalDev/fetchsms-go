// Package fetchsms provides a context-aware client for the Fetch SMS v1 public
// catalog and API-key endpoints. Money is integer US cents. Creating or reusing
// verifications and creating or extending rentals spends wallet funds.
//
// GetNumber buys a number using a service slug; it is not a read-only lookup.
// Use CreateVerification for numeric service IDs or an area-code preference,
// GetVerification to read an order, and WaitForCode after triggering an SMS
// at the external service.
//
// No mutating call is automatically retried. After a transport error, reconcile
// account state before deciding whether to repeat a purchase. WaitForCode and
// WaitForRentalCode retry only HTTP 429 responses, respecting Retry-After.
// Always use a deadline for polling. Nil contexts mean context.Background.
//
// Verify webhook signatures on the raw request body before DecodeWebhook.
// Signatures alone do not prevent replay.
package fetchsms
