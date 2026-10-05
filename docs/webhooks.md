# Webhooks

Register an endpoint and its endpoint-specific secret in the Fetch SMS dashboard. Do not use the API key as the webhook secret. Deliveries are POST JSON with `X-FetchSMS-Signature` (hex HMAC-SHA256 of the **raw body**) and `X-FetchSMS-Event`.

## Safe processing order

1. Require POST and HTTPS (directly or via a trusted TLS-terminating proxy).
2. Bound the body before reading (`http.MaxBytesReader`, for example 1 MiB); set server timeouts. Reject oversized or unreadable bodies.
3. Call `VerifyWebhookSignature(raw, signatureHeader, endpointSecret)` on the exact bytes read. Do not trim whitespace, decode/re-encode JSON, or normalize line endings first. Empty secrets and malformed signatures fail closed; comparison uses `hmac.Equal`.
4. Only after authentication, call `DecodeWebhook(raw)`. This decoder is not an authenticator.
5. Dispatch on the signed body's `Event`. The event header is not separately signed; if present, check it agrees with the body rather than trusting it for authorization.
6. Persist and atomically deduplicate/enqueue before acknowledging success. Restrict processing to your expected resource IDs/accounts. Do not expose SMS text/codes in logs or error responses.

The [example server](../examples/webhook/main.go) demonstrates verification and typed parsing only. It deliberately performs no business effects; add durable replay protection and processing before production. It binds loopback, expects TLS termination, reads `FETCHSMS_WEBHOOK_SECRET`, limits bodies, and configures HTTP timeouts.

## Typed envelope

`DecodeWebhook` returns `WebhookEnvelope` with Event, CreatedAt, raw Data, and one populated typed pointer for recognized events:

- `EventSMSReceived` (`sms.received`): `SMSReceived *SMSReceivedEvent` with VerificationID or RentalID, Ref, Number, Sender, Code, DetectedService, ReceivedAt, and optional Body. Short-term events have no message body; rental events include the full supported-service SMS text.
- `EventRentalExpired` (`rental.expired`): `RentalExpired *RentalExpiredEvent` with RentalID, Ref, Number.
- Unknown events retain raw `Data` without failing merely because the event is new. Both typed pointers are nil. Safely ignore or persist unknown events for later handling; never treat them as known business actions.

Use `DecodeWebhook`, not direct `json.Unmarshal`, to populate typed pointers. Raw Data is preserved for all events, including unknown fields in known events. Null or missing data/envelope event names are rejected. UTC timestamp behavior is shared with REST models.

## Replay and duplicate delivery

**A valid HMAC proves authenticity/integrity, not freshness.** The documented signature has no signed delivery timestamp header or unique event ID. Captured valid requests can be replayed unchanged; do not claim HMAC alone prevents this.

Persist a digest of the authenticated raw body with a suitable retention period and use a database uniqueness constraint/transaction for deduplication across workers/restarts. Treat CreatedAt freshness windows as an application policy: delayed legitimate deliveries and clock skew can cause false rejection. For strong business guarantees, combine idempotent resource processing and reconciliation with the API; two byte-identical legitimate events cannot be distinguished solely by a raw-body digest. Never deduplicate on code alone (codes may repeat), and do not invent an upstream event ID.

Rotate endpoint secrets carefully; if accepting an old secret during a brief overlap, apply the same replay store. Store secrets in a secret manager/environment, not source or logs. Delivery retry schedules and ordering guarantees are not specified by the retrieved API reference; make no assumptions about exactly-once delivery.

Source: https://fetchsms.com/docs/api (retrieved October 5, 2026).
