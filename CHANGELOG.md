# Changelog

## v0.2.0

- Added `New(key) *SimpleClient`, `GetNumber(service)`, and `GetCode(id)` without caller contexts.
- Internal per-call deadlines: 30 seconds for purchasing, 15 minutes for code polling; default polling delay remains three seconds and each HTTP request has a 30-second timeout.
- Preserved every v0.1.0 API, including `NewClient`, context-aware `GetNumber`, and `WaitForCode`.
- Simplified README quickstart and guarded simple example; documented external SMS triggering, polling lifetime, and purchase reconciliation.
- Added local tests for purchase payloads, polling, API/input errors, deadlines, no mutation retries, and API compatibility. No live API calls or purchases were made for this change.

## v0.1.0

Initial public release of the independent FetchSMS Go SDK.

- Simple `GetNumber(ctx, serviceSlug)` and `WaitForCode(ctx, verificationID)` workflow.
- Typed public catalog, area codes, quotes, wallet balance, verification, and rental methods.
- Context-aware polling with cancellable rate-limit backoff.
- Webhook HMAC verification and typed event decoding.
- Standard library only; MIT licensed.
- Documentation, examples, mock-server tests, and GitHub Actions CI.

### Validation and limitations

Local tests cover request/response handling, errors, polling, and webhooks. Live read-only checks passed for services, area codes, price quotes, and authenticated wallet balance. Live purchases, SMS delivery, rentals, and webhook delivery have not been validated. Treat this initial v0 release as an evolving API.
