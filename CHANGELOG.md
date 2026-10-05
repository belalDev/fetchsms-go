# Changelog

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
