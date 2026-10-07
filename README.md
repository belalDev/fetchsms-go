# FetchSMS Go

A standard-library-only, typed Go client for Fetch SMS API v1. Includes the public catalog, API-key verification/rental endpoints, balance, cancellable polling, and signed webhook decoding. This is an independent SDK, not an official Fetch SMS release.

## Get a number, then get its code

No `os`, `context`, `log`, or timer setup required. The package handles timeouts,
polling, and response decoding internally.

```go
fetchClient := fetchsms.New("FETCHSMS_API_KEY")

number, err := fetchClient.GetNumber("whatsapp")
if err != nil { return err }

// Request the SMS using number.Number.

code, err := fetchClient.GetCode(number.ID)
if err != nil { return err }
```

Import `fetchsms "github.com/belaldev/fetchsms-go"` and use this snippet inside
an application function that returns `error`; use `code` in the rest of that
function. Replace `"FETCHSMS_API_KEY"` with your actual key—it is a placeholder,
not an automatic environment-variable lookup. Never commit a real key.

**`GetNumber` charges wallet funds.** Choose a slug from the current catalog
and review prices/account limits before running. `New` makes no requests.
`GetNumber` makes one paid POST, with no quote/balance preflight and no retry.
A timeout or failed response may follow a successful charge: reconcile via
list/get/dashboard before purchasing again. Save the returned order ID.

You must enter the number into the external service and request its SMS there.
`GetCode` only retrieves the code: it polls immediately, then waits three seconds
between polls, for up to **15 minutes from the call**. Each HTTP request has a
30-second timeout; `GetNumber` also has its own **30-second operation deadline**.
The polling deadline does not extend the server's verification lifetime, cancel
an order, or guarantee a refund. HTTP 429 polling backoff honors `Retry-After`;
other errors and terminal statuses can stop polling earlier. Codes remain strings,
preserving leading zeroes.

## Requirements and installation

Go 1.26.5 or newer (the existing module/toolchain requirement is preserved).

```sh
go get github.com/belaldev/fetchsms-go@v0.2.0
```

For local development, clone the repository or use a Go workspace.

## Examples and advanced API

The runnable [simple example](examples/simple/main.go) uses `New`, `GetNumber`,
and `GetCode`, with an explicit `-spend` guard and an SMS-trigger prompt.
The [advanced verification example](examples/verification/main.go) and
[rental example](examples/rental/main.go) retain context/configuration control.
Examples are compiled in tests, not executed against live APIs.

```sh
# These commands SPEND MONEY; do not run as an installation check.
go run ./examples/simple -spend -service telegram
go run ./examples/verification -spend -service 4
go run ./examples/rental -spend
```

All v0.1.0 APIs remain unchanged. Use `NewClient(key, ...options)` for custom
contexts, HTTP settings, polling intervals, catalog/quote/balance calls, numeric
service IDs, area codes, rentals, and webhooks. Its existing
`GetNumber(ctx, service)` and `WaitForCode(ctx, id)` remain available.
`Quote` is not a price reservation or spending cap.

### Public catalog: no purchase

```go
package main

import (
    "context"
    "fmt"
    "time"

    fetchsms "github.com/belaldev/fetchsms-go"
)

func main() {
    client, err := fetchsms.NewClient("")
    if err != nil { fmt.Println(err); return }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    services, err := client.Services(ctx)
    if err != nil { fmt.Println(err); return }
    for _, service := range services {
        fmt.Printf("%d %s: %d cents\n", service.ID, service.Name, service.PriceCents)
    }
}
```

A verification has a fixed 15-minute window. `VerificationMessages` returns all matching codes newest first; `Verification.Code` is nullable and is the latest. Rental delivery is limited to supported services. `WaitForRentalCode` returns the latest existing code, **not necessarily a new delivery**; track message IDs through `RentalMessages` or webhooks for subsequent SMS.

## Advanced configuration and shared behavior

```go
client, err := fetchsms.NewClient(key,
    fetchsms.WithHTTPClient(&http.Client{Timeout: 20 * time.Second}),
    fetchsms.WithPollInterval(3 * time.Second),
    fetchsms.WithBaseURL("https://api.fetchsms.com/v1"),
)
```

- Default request timeout: 30 seconds; polling interval: 3 seconds. An injected HTTP client supplies its own timeout. Its configuration is copied, not its transport/jar. Do not mutate shared configuration concurrently.
- Every `Client` network method takes `context.Context`. Nil means `context.Background()`; prefer explicit deadlines, especially for polling. `SimpleClient` instead supplies its own per-call deadlines and exposes no settings.
- Public catalog calls do not send the key. Authenticated calls require a nonblank key. All redirects are rejected, even with an injected client's redirect callback, preventing credential forwarding and redirected purchases.
- Use HTTPS in production. HTTP base URLs exist for local mock servers; a custom base URL receives your credentials on authenticated calls. Only configure trusted hosts/transports.
- No SDK retries for ordinary requests, especially purchases. Only polling GETs retry HTTP 429, waiting at least the greater of the poll interval and `Retry-After` (seconds or HTTP date). Waiting is cancellable. Other HTTP/transport failures return immediately.
- **A timeout, cancellation, malformed response, or connection failure after a purchase may mean the server already charged you.** Reconcile via list/get/dashboard before retrying. The API documents no idempotency key; this SDK does not invent one. A custom transport must not add automatic mutation retries.
- Polling returns `TerminalStatusError` for expired, cancelled, or unknown statuses. Received with a null/empty code continues waiting until a deadline. Polling does not cancel orders or promise refunds.
- Responses are bounded to 4 MiB, including errors; oversized successful responses return `ErrResponseTooLarge`. Unknown JSON fields are accepted for forward compatibility.
- Monetary fields are `int64` integer US cents. Optional values use pointers. `Timestamp` accepts offset-free UTC (including fractional seconds), RFC3339 offsets, and null; parsed timestamps normalize to UTC. Null maps to zero; marshaling zero emits null.
- IDs accept only ASCII letters, digits, `_`, and `-`; blank/path/query/escape tricks are rejected locally. UUIDs are supported without unnecessarily enforcing a UUID version.

## Errors

Use `errors.As(err, &apiErr)` with `var apiErr *fetchsms.APIError`. Inspect `StatusCode`, `Detail` (`json.RawMessage`, string or structured JSON), `Body` (bounded raw bytes), `Truncated`, `RetryAfter`, and `RetryAfterRaw`. Non-JSON error bodies are retained. `Error()` only includes HTTP status, intentionally avoiding sensitive server text. Do not log raw detail/body without sanitization. Context errors work with `errors.Is`.

## Reference and webhooks

- [Endpoint reference](docs/API.md)
- [Webhook verification and replay precautions](docs/webhooks.md)
- [Authentication-only webhook server example](examples/webhook/main.go)

Dashboard session routes (deposits, account/API-key/webhook management, support, admin) are intentionally excluded. Webhook endpoints are configured in the Fetch SMS dashboard, not through invented API-key methods.

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

Tests use local `httptest` servers and do not call live APIs or spend money. Implementation follows the upstream API reference retrieved October 5, 2026: https://fetchsms.com/docs/api. The small client/quickstart style was inspired by https://github.com/capsolver/capsolver-go; no implementation copied. Live read-only checks passed for services, area codes, quotes, and authenticated balance. Live purchases, SMS delivery, rentals, and webhook delivery remain untested.

## License

Licensed under the MIT License. See [LICENSE](LICENSE).
