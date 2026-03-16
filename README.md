# go-payoneer

A high-quality, type-safe, and observable Go SDK for the Payoneer API.

[![Go Reference](https://pkg.go.dev/badge/github.com/pcriv/go-payoneer.svg)](https://pkg.go.dev/github.com/pcriv/go-payoneer)
[![Go Report Card](https://goreportcard.com/badge/github.com/pcriv/go-payoneer)](https://goreportcard.com/report/github.com/pcriv/go-payoneer)

## Features

- **Full Service Coverage**: Accounts, Payouts, Webhooks, and Payee Management.
- **Robust Authentication**: Secure OAuth 2.0 implementation with automatic token refreshing.
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
    payoneer.WithRetryMax(3),
)
```

### 2. Authenticate

```go
ctx := context.Background()

// OAuth 2.0 Client Credentials flow (recommended for backend-to-backend)
err := client.Authenticate(ctx, payoneer.WithClientCredentials(
    "your-client-id",
    "your-client-secret",
))
```

## Usage Examples

### Account Balances

```go
balances, err := client.Accounts.ListBalances(ctx)
if err != nil {
    // Handle error (SDK provides custom APIError with status code mapping)
}

for _, balance := range balances {
    fmt.Printf("%s: %d\n", balance.Currency, balance.Amount)
}
```

### Submit a Mass Payout

```go
request := &payoneer.MassPayoutRequest{
    ClientReferenceID: "unique-payout-ref-123", // Mandatory idempotency
    Payouts: []payoneer.Payout{
        {
            PayeeID:  "payee-456",
            Amount:   15000, // $150.00 (in cents/minor units)
            Currency: "USD",
        },
    },
}

result, err := client.Payouts.Create(ctx, request)
```

### Handle Webhooks (IPCN)

Protect your webhook endpoint using the built-in middleware.

```go
secret := "your-merchant-secret"
mux := http.NewServeMux()

// Middleware validates HMAC signature and restores the body for the handler
mux.Handle("/webhooks", payoneer.WebhookValidator(secret)(
    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        event, _ := payoneer.ParseWebhook(r, secret)
        fmt.Printf("Received %s for event ID %s\n", event.EventType, event.EventID)
        w.WriteHeader(http.StatusOK)
    }),
))
```

### Payee Onboarding

```go
// Generate a link for a new payee to register
link, err := client.Payees.RegistrationURL(ctx, "payee-789", 
    payoneer.WithRedirectURL("https://myapp.com/onboarded"),
    payoneer.WithLanguage("en"),
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
