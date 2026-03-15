# Architecture Patterns

**Domain:** Payment SDK (Payoneer)
**Researched:** 2025-03-20

## Recommended Architecture

The SDK follows a "Client-centric" architecture where a main `Client` struct coordinates authentication, transport (retries/OTel), and provides access to specific API service domains (e.g., `Client.Payouts`, `Client.Accounts`).

### Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| `Client` | Coordination, configuration, and public entry point. | `Option`, `Service`, `Transport` |
| `Authenticator` | Managing OAuth2 tokens and refreshing. | Payoneer Auth API, `http.Client` |
| `Transport` | Handles retries, logging, and tracing. | `http.Client`, `OTel`, `slog` |
| `Services` | Domain-specific logic (Payouts, Balances). | `Client`, Payoneer REST API |
| `Models` | Request/Response data structures. | `Services`, Consumer App |

### Data Flow

```
User → Client.Payouts.Create(ctx, req) 
  → Client.execute(ctx, method, path, body)
    → Transport.RoundTrip(http.Request)
      → OTel Span Start
      → slog.InfoContext("starting request")
      → go-retryablehttp.Do(http.Request)
        → Authenticator (Inject Bearer Token)
        → Payoneer API
      → slog.InfoContext("request completed")
      → OTel Span End
  → Decode JSON to Response Model
→ Return to User
```

## Patterns to Follow

### Pattern 1: Functional Options for Extensibility
**What:** Define a set of functions that modify a configuration struct.
**When:** Creating a new `Client` or complex request objects.
**Example:**
```typescript
// Go example in pseudo-code
func NewClient(apiKey string, opts ...Option) *Client {
    c := &Client{apiKey: apiKey, logger: slog.Default()}
    for _, opt := range opts {
        opt(c)
    }
    return c
}
```

### Pattern 2: Context-aware Operations
**What:** Every API call should accept `context.Context` as the first argument.
**When:** Performing any network I/O or long-running work.
**Rationale:** Standard Go practice for cancellation and trace propagation.

## Anti-Patterns to Avoid

### Anti-Pattern 1: Global Singleton Client
**What:** Providing a package-level "DefaultClient".
**Why bad:** Prevents testing in isolation and makes it impossible to use multiple Payoneer accounts in a single process.
**Instead:** Require users to instantiate a `Client` via `NewClient()`.

### Anti-Pattern 2: Log-and-Return-Error
**What:** Logging an error inside the library and then returning it.
**Why bad:** Causes duplicate logs (the caller will likely log it too) and pollutes the application logs.
**Instead:** Return the error wrapped with context. Only log internal SDK events if necessary, using `Debug` or `Info` level.

## Scalability Considerations

| Concern | At 100 users | At 10K users | At 1M users |
|---------|--------------|--------------|-------------|
| Rate Limiting | Basic 429 handling | Exponential backoff + jitter | Multi-instance token bucket synchronization |
| Memory usage | Minimal | Avoid unnecessary allocations in JSON decoding | Use `sync.Pool` for byte buffers |
| Concurrency | Thread-safe `Client` | Use `http.DefaultTransport` for connection pooling | Advanced tuning of `MaxIdleConns` |

## Sources

- [Stripe Go SDK Architecture](https://github.com/stripe/stripe-go)
- [AWS SDK for Go V2 Architecture](https://github.com/aws/aws-sdk-go-v2)
