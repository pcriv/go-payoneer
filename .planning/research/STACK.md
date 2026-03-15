# Technology Stack

**Project:** go-payoneer
**Researched:** 2025-03-20
**Overall Confidence:** HIGH

## Recommended Stack

### Core Framework
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.21+ | Language | Required for native `log/slog` support and modern concurrency features. |

### HTTP Client & Networking
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `hashicorp/go-retryablehttp` | v0.7.x | Robust HTTP Retries | Handles complex retry logic, specifically request body rewinding for POST/PUT requests which is difficult to implement manually. |
| `golang.org/x/oauth2` | Latest | OAuth 2.0 Auth | Industry standard for handling OAuth2 flows (Authorization Code, Client Credentials) used by Payoneer. |

### Observability
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `go.opentelemetry.io/otel` | Latest (API only) | Tracing/Metrics | Standard for cloud-native observability. Library should only depend on API to avoid version conflicts with consumers. |
| `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` | Latest | HTTP Instrumentation | Automatic instrumentation for HTTP client calls within the SDK. |

### Logging
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `log/slog` | Stdlib (1.21+) | Structured Logging | Use for all internal SDK logging to ensure compatibility with modern Go ecosystems. |

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| HTTP Client | `go-retryablehttp` | Custom Wrapper | Custom wrappers often fail to handle request body rewinding correctly for retries on non-idempotent methods. |
| Logging | `log/slog` | `zap` / `logrus` | Public libraries should avoid third-party logging dependencies to prevent dependency bloat for consumers. |
| Observability | `OTel` | `OpenCensus` / None | OTel is the modern successor and standard for Go observability. |

## Installation

```bash
# Core Dependencies
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/trace
go get github.com/hashicorp/go-retryablehttp
go get golang.org/x/oauth2

# Dev Dependencies for Testing
go get go.opentelemetry.io/otel/sdk/trace/tracetest
```

## Implementation Patterns

### 1. Functional Options (Configuration)
```go
type Client struct {
    logger *slog.Logger
    httpClient *http.Client
}

type Option func(*Client)

func WithLogger(l *slog.Logger) Option {
    return func(c *Client) { c.logger = l }
}

func NewClient(opts ...Option) *Client {
    c := &Client{logger: slog.Default()} // Default to global
    for _, opt := range opts { opt(c) }
    return c
}
```

### 2. OTel Integration (API Only)
```go
// Use the global tracer provider via the API
var tracer = otel.Tracer("github.com/pablocrivella/go-payoneer")

func (c *Client) Execute(ctx context.Context, req *Request) error {
    ctx, span := tracer.Start(ctx, "PayoneerAPI.Execute")
    defer span.End()
    // ...
}
```

## Sources

- [OpenTelemetry Go Best Practices](https://opentelemetry.io/docs/languages/go/instrumentation/)
- [Slog Idiomatic Usage](https://go.dev/blog/slog)
- [Hashicorp go-retryablehttp](https://github.com/hashicorp/go-retryablehttp)
- [Payoneer API Documentation](https://developer.payoneer.com/docs/)
