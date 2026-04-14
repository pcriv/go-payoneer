# go-payoneer

A high-quality, type-safe, and observable Go SDK for the Payoneer API.

[![Go Reference](https://pkg.go.dev/badge/github.com/pcriv/go-payoneer.svg)](https://pkg.go.dev/github.com/pcriv/go-payoneer)
[![Go Report Card](https://goreportcard.com/badge/github.com/pcriv/go-payoneer)](https://goreportcard.com/report/github.com/pcriv/go-payoneer)
[![codecov](https://codecov.io/gh/pcriv/go-payoneer/graph/badge.svg)](https://codecov.io/gh/pcriv/go-payoneer)

## Features

- **Full Service Coverage**: Accounts, Payouts, Webhooks, and Payee Management.
- **Robust Authentication**: Secure OAuth 2.0 with eager credential validation and automatic token refreshing.
- **Observability**: First-class support for `slog` structured logging and OpenTelemetry (Tracing/Metrics).
- **Resiliency**: Built-in exponential backoff retries and rate-limit handling (429s).
- **Type-Safety**: Clean Go structs for all API resources, using generics for optional/nullable fields.
- **Secure Webhooks**: Mandatory HMAC SHA-256 signature validation with an easy-to-use middleware.

## Installation

```bash
go get github.com/pcriv/go-payoneer
```

## Quick Start

### 1. Initialize the Client

The SDK uses functional options for clean and flexible configuration.

```go
import "github.com/pcriv/go-payoneer/pkg/payoneer"

client := payoneer.NewClient(
    payoneer.WithSandbox(), // Use Sandbox for development
    payoneer.WithProgramID("your-program-id"),
    payoneer.WithClientCredentials("your-client-id", "your-client-secret"),
    payoneer.WithRetries(3),
)
```

### 2. Authenticate

`Authenticate` eagerly fetches a token to validate your credentials. If the credentials or token endpoint are invalid, it fails immediately with a clear error.

```go
ctx := context.Background()

err := client.Authenticate(ctx)
if err != nil {
    // errors.Is(err, payoneer.ErrAuthenticationFailed) can be used
    // to programmatically detect auth failures.
    log.Fatal(err)
}
```

## Usage Examples

### Submit a Mass Payout

```go
request := &payoneer.MassPayoutRequest{
    Payments: []payoneer.PayoutItem{
        {
            ClientReferenceID: "unique-payout-ref-123",
            PayeeID:           "payee-456",
            Amount:            15000, // $150.00 (in cents, converted to 150.00 on the wire)
            Currency:          "USD",
            Description:       "March payout",
        },
    },
}

result, err := client.Payouts.SubmitMany(ctx, request)
// result.Result — e.g. "Payments Created"
```

### Handle Webhooks (IPCN)

Protect your webhook endpoint using the built-in middleware.

```go
cfg := payoneer.WebhookConfig{
    Secret:          "your-shared-secret",
    ExpectedAppName: payoneer.AppNameProduction, // or AppNameSandbox
    // MaxClockSkew defaults to 5m; set to -1 to disable the timestamp check.
    // NonceStore: yourStore, // optional replay protection
}

mux := http.NewServeMux()

// Middleware parses `Authorization: hmacauth <AppName>:<Signature>:<Nonce>:<Timestamp>`,
// verifies the HMAC-SHA256 over payload+nonce+timestamp, and restores the body.
mux.Handle("/webhooks", payoneer.WebhookValidator(cfg)(
    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        event, _ := payoneer.ParseWebhook(r, cfg)
        fmt.Printf("Received %s for event ID %s\n", event.EventType, event.EventID)
        w.WriteHeader(http.StatusOK)
    }),
))
```

### Payee Onboarding

```go
// Generate a link for a new payee to register
result, err := client.Payees.CreateRegistrationLink(ctx, "payee-789",
    payoneer.WithRedirectURL("https://myapp.com/onboarded"),
    payoneer.WithLanguage("en"),
)
// result.RegistrationLink — the URL to redirect the payee to
// result.Token           — unique token for this registration session
```

## Configuration

### Environments

`WithSandbox()` configures both the API base URL (`api.sandbox.payoneer.com`) and the OAuth2 base URL (`login.sandbox.payoneer.com`) in one call. For production, the defaults point to `api.payoneer.com` and `login.payoneer.com` respectively.

You can override them independently if needed:

```go
client := payoneer.NewClient(
    payoneer.WithBaseURL("https://api.payoneer.com"),
    payoneer.WithAuthBaseURL("https://login.payoneer.com"),
)
```

## Advanced Patterns

### Optional Fields

The SDK uses a generic `Optional[T]` type to distinguish between empty values and fields omitted by the API.

```go
if val, ok := payee.FirstName.Get(); ok {
    fmt.Println("First Name:", val)
}
```

### Redacting Logs

Sensitive information is automatically redacted in logs.

```go
client := payoneer.NewClient(
    payoneer.WithLogger(slog.Default()), // Integrated with transport redaction
)
```

## License

MIT
