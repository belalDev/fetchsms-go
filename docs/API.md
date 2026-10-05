# API reference

Base URL: `https://api.fetchsms.com/v1`. All methods take a context first and return `(typedValue, error)`, except `CancelVerification`, which returns only `error`. The client is created with `NewClient(key, ...options)` and returns `(*Client, error)`.

## Public catalog (no Authorization header)

| Method | HTTP endpoint | Parameters / result |
|---|---|---|
| `Services(ctx)` | GET `/services` | `[]Service`: ID, slug, name, price cents, long-term price map, short/long stock |
| `AreaCodes(ctx, AreaCodesRequest)` | GET `/services/area-codes` | Mode defaults short; zero Service omits filter; `[]string` |
| `Quote(ctx, QuoteRequest)` | GET `/services/quote` | Service + Mode required; Days required for long; CustomAreaCode boolean; `Quote` |

`ServiceID(positiveInt)` and `ServiceSlug(lowercase-hyphenated-slug)` return `(ServiceRef, error)`. Numeric references marshal as JSON numbers, slugs as strings. The zero reference is invalid for required fields. `ModeShort` / `ModeLong` are typed constants. Long-term service is ID 1 / `unlimited-services`. Long terms accepted: 1, 7, 30, 90, 365 days.

Quote fields: `Service`, `Mode`, `Days *int`, `BaseDailyCents int64`, `DiscountPct int`, `PerDayCents *int64`, `SurchargeCents int64`, `TotalCents int64`. Do not calculate final prices from rounded per-day figures; use TotalCents. Quotes do not lock price/stock.

## API-key endpoints (Bearer authentication)

| Method | HTTP endpoint | Parameters / result |
|---|---|---|
| `Balance(ctx)` | GET `/wallet/balance` | `Balance{BalanceCents int64}` |
| `GetNumber(ctx, service)` | POST `/verifications` | Validated service slug string; `Verification`; **spends funds**, not a lookup |
| `CreateVerification(ctx, CreateVerificationRequest)` | POST `/verifications` | Service required, AreaCode optional; `Verification`; **spends funds** |
| `GetVerification(ctx, id)` | GET `/verifications/{id}` | `Verification` |
| `ListVerifications(ctx, tab)` | GET `/verifications?tab=…` | `[]Verification`; TabActive / TabHistory; zero defaults active |
| `VerificationMessages(ctx, id)` | GET `/verifications/{id}/messages` | `[]VerificationMessage`, newest first, no SMS body |
| `CancelVerification(ctx, id)` | POST `/verifications/{id}/cancel` | Only `error`; successful empty or bounded nonempty responses accepted because upstream shape is undocumented |
| `ReuseVerification(ctx, id)` | POST `/verifications/{id}/reuse` | New `Verification`; **spends funds** |
| `CreateRental(ctx, CreateRentalRequest)` | POST `/rentals` | Service=1/Unlimited, Days required, AreaCode optional; `Rental`; **spends funds** |
| `GetRental(ctx, id)` | GET `/rentals/{id}` | `Rental` |
| `ListRentals(ctx, ListRentalsRequest)` | GET `/rentals?tab=…&limit=…` | `[]Rental`; active default, limit 1–200; zero omits limit (server default 100) |
| `RentalMessages(ctx, id)` | GET `/rentals/{id}/messages` | `[]RentalMessage`, newest first, supported-service messages only |
| `ExtendRental(ctx, id, days)` | POST `/rentals/{id}/extend` | Body `{"days": n}`; updated `Rental`; **spends funds** |
| `CancelRental(ctx, id)` | POST `/rentals/{id}/cancel` | Updated `Rental`, cancelled/refunded on success |

Cancel/reuse calls send no invented request body. Optional AreaCode is a three-digit string, omitted when empty. The server remains authoritative on stock, spend limits, eligibility, timing, and refunds. All money is integer US cents.

### Verification

For simple purchases, call `GetNumber(ctx, "telegram")`. It wraps `ServiceSlug`
and then `CreateVerification`, returning `(Verification, error)`. Invalid slugs
fail locally without sending a request; API errors propagate unchanged. It does
not fetch a quote/balance, select an area code, trigger an external SMS, or wait
for a code. Use `CreateVerification` for numeric IDs or an area-code preference,
`GetVerification` to read an existing order, and `WaitForCode` after triggering SMS.

**GetNumber charges wallet funds and never automatically retries.** A timeout,
cancellation, connection failure, or malformed response may occur after a charge;
reconcile via list/get/dashboard before repeating a purchase. There is no SDK
idempotency guarantee.

Fields: `ID`, `Ref`, `Service`, `ServiceName`, `Number` strings; `Code *string`; `CodeCount int`; `Status`; `CostCents int64`; `CreatedAt`, `ExpiresAt` timestamps; `Reusable bool`, `ReuseCostCents *int64`, `ReuseNote *string`. Absent/null optional reuse price/note remain nil. Reuse needs a finished code-receiving verification and a currently eligible unchanged/free number; it is not guaranteed by a prior list response.

`VerificationMessage`: `ID string`, `Code string`, `ReceivedAt Timestamp`.

### Rental

Fields: `ID`, `Ref`, `Service`, `ServiceName`, `Number`; `DurationDays`, `ExtendedDays`; `CustomAreaCode bool`; `Status`; `LastCode *string`; `CostCents int64`; `Refunded bool`; `CreatedAt`, `ExpiresAt`, `CancelDeadline` timestamps. Extension totals are cumulative. Received rentals remain active until expiry. Server limits extensions to 730 days ahead. Custom area-code surcharges recur on extensions.

`RentalMessage`: `ID`, `Sender`, `Body`, `Code` strings and `ReceivedAt Timestamp`.

Statuses: `StatusWaiting`, `StatusReceived`, `StatusExpired`, `StatusCancelled`. Unknown statuses remain visible as strings. Cancellation eligibility is enforced by the server: verifications cannot be cancelled after a code; rentals require no code and cancellation before the 15-minute deadline. Do not infer refund guarantees from a local timeout.

## Polling

`WaitForCode(ctx, verificationID)` and `WaitForRentalCode(ctx, rentalID)` return `(string, error)`. Poll immediately; success requires received status and a nonempty code. Rental polling may return an old latest code. A shared polling loop retries only 429 GET responses; terminal/unknown statuses yield `*TerminalStatusError` with ResourceID/Status. Use context deadlines. No automatic cancellation, mutation retry, or delivery deduplication.

## HTTP errors

`*APIError`: StatusCode, raw Detail JSON, bounded Body bytes, RetryAfter duration, RetryAfterRaw header, Truncated flag. Typical statuses: 400 input/business rule, 401 credentials, 402 funds, 403 permission/spend cap, 404 resource/service, 409 stock/state/cancellation conflict, 422 parameters, 429 rate or active-number limit. A 429 does not necessarily mean request-rate exhaustion.

The SDK accepts successful 2xx responses, rejects redirects, bounds bodies, and returns decoding errors for invalid success JSON. No pagination cursor or idempotency field is documented, so none is invented. Lists expose the documented filters only.

## Scope and provenance

Public/API-key scope only; no dashboard tokens, wallet deposits/history, webhook registration, API-key management, support or admin routes. Source: https://fetchsms.com/docs/api, retrieved in full October 5, 2026. Tests use mock servers. Live read-only checks passed for services, area codes, quotes, and authenticated balance; live purchases, SMS delivery, rentals, and webhook delivery remain untested.
