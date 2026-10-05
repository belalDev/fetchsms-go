# FetchSMS Go

A standard-library-only, typed Go client for Fetch SMS API v1. Includes the public catalog, API-key verification/rental endpoints, balance, cancellable polling, and signed webhook decoding. This is an independent SDK, not an official Fetch SMS release.

## Requirements and installation

Go 1.26.5 or newer (the existing module/toolchain requirement is preserved).

```sh
go get github.com/belaldev/fetchsms-go@v0.1.0
```

For local development, clone the repository or use a Go workspace.

## Quickstart: no purchase

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    fetchsms "github.com/belaldev/fetchsms-go"
)

func main() {
    client, err := fetchsms.NewClient("") // Catalog calls need no key.
    if err != nil { log.Fatal(err) }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    services, err := client.Services(ctx)
    if err != nil { log.Fatal(err) }
    for _, service := range services {
        fmt.Printf("%d %s: %d cents\n", service.ID, service.Name, service.PriceCents)
    }
}
```

## Receive a verification code

**Creating/reusing a verification or creating/extending a rental spends real wallet funds.** Read the current quote and account limits first. Examples require an explicit `-spend` flag and read credentials from `FETCHSMS_API_KEY`; never put credentials in source control.

1. Fetch `Services` and choose a service slug, such as `telegram`.
2. Use `Quote` and `Balance` before purchasing. Quotes are not reservations or spending caps: prices and stock can change.
3. Call `GetNumber(ctx, "telegram")` once to buy a number. Save its ID immediately.
4. **Enter the returned number into the chosen external service and trigger its SMS there.** Fetch SMS reserves/receives numbers; creating a verification does not request a login SMS from that service.
5. Call `WaitForCode(ctx, id)` with a deadline. The returned string preserves leading zeroes. Deliver it securely; do not log codes or full SMS bodies.

### Simple purchase and polling

`GetNumber(ctx, service)` returns `(Verification, error)`. It validates the slug
with `ServiceSlug` and calls `CreateVerification` with no area-code preference.
Despite its name, **this is a paid POST, not a read-only lookup**. It makes no
quote/balance preflight and never automatically retries. A timeout, cancellation,
or failed response can leave the purchase outcome unknown: reconcile your account
before trying again. Use `GetVerification(ctx, id)` to read an existing order.

This complete program requires explicit purchase authorization:

```go
package main

import (
 "bufio"
 "context"
 "flag"
 "fmt"
 "log"
 "os"
 "time"

 fetchsms "github.com/belaldev/fetchsms-go"
)

func main() {
 spend := flag.Bool("spend", false, "authorize one paid verification")
 service := flag.String("service", "telegram", "service slug from the current catalog")
 flag.Parse()
 if !*spend {
  log.Fatal("This example spends wallet funds. Review it, then explicitly pass -spend.")
 }
 key := os.Getenv("FETCHSMS_API_KEY")
 if key == "" { log.Fatal("set FETCHSMS_API_KEY") }
 client, err := fetchsms.NewClient(key)
 if err != nil { log.Fatal(err) }
 ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
 defer cancel()
 // Paid purchase: never blindly retry an ambiguous failure.
 v, err := client.GetNumber(ctx, *service)
 if err != nil { log.Fatal(err) }
 fmt.Printf("Verification %s: enter %s into the selected service and request an SMS there.\nPress Enter after triggering the SMS.\n", v.ID, v.Number)
 if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil { log.Fatal(err) }
 code, err := client.WaitForCode(ctx, v.ID)
 if err != nil { log.Fatal(err) }
 // Deliver code to your application securely; do not log it.
 _ = code
 fmt.Println("Code received (not printed).")
}
```

Use `CreateVerification` with `ServiceID` or `ServiceSlug` for advanced requests,
including a numeric service ID or `AreaCode`; all existing APIs remain available.
See the compiling [simple example](examples/simple/main.go),
[advanced verification example](examples/verification/main.go), and
[rental example](examples/rental/main.go). They are compiled in CI but never
executed by tests. Review prices and account limits before running:

```sh
# These commands SPEND MONEY; do not run as an installation check.
go run ./examples/simple -spend -service telegram
# Advanced numeric-ID workflow:
go run ./examples/verification -spend -service 4
go run ./examples/rental -spend
```

A verification has a fixed 15-minute window. `VerificationMessages` returns all matching codes newest first; `Verification.Code` is nullable and is the latest. Rental delivery is limited to supported services. `WaitForRentalCode` returns the latest existing code, **not necessarily a new delivery**; track message IDs through `RentalMessages` or webhooks for subsequent SMS.

## Configuration and behavior

```go
client, err := fetchsms.NewClient(key,
    fetchsms.WithHTTPClient(&http.Client{Timeout: 20 * time.Second}),
    fetchsms.WithPollInterval(3 * time.Second),
    fetchsms.WithBaseURL("https://api.fetchsms.com/v1"),
)
```

- Default request timeout: 30 seconds; polling interval: 3 seconds. An injected HTTP client supplies its own timeout. Its configuration is copied, not its transport/jar. Do not mutate shared configuration concurrently.
- Every network method takes `context.Context`. Nil means `context.Background()`; prefer explicit deadlines, especially for polling.
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
