# Phase 2: Observability, Robustness & Account Services - Research

**Researched:** 2026-03-15
**Domain:** Go API Client, OpenTelemetry, Retry Logic, Financial APIs
**Confidence:** HIGH

## Summary
Phase 2 focuses on enhancing the `go-payoneer` SDK with production-grade reliability (retries, rate limiting) and observability (OpenTelemetry), while implementing the core Account Services (Balances and Transactions).

**Primary recommendation:** Use `hashicorp/go-retryablehttp` for robustness and `otelhttp` for automated transport-level tracing, supplemented by manual spans for business operations.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Library**: Use `github.com/hashicorp/go-retryablehttp`.
- **Configuration**: Simple exposure via `WithRetries(max int)` functional option.
- **Integration**: Internal wrapping; `Client.HTTPClient` remains a standard `*http.Client` pointing to the retryable client.
- **Logging**: Unified; route `go-retryablehttp` logs through the SDK's existing redacted `slog.Logger`.
- **Control**: Fully automatic retries based on status codes (e.g., 429, 5xx), leveraging Payoneer's idempotency.
- **Instrumentation**: Hybrid approach (Transport-level `otelhttp` + Service-level semantic spans).
- **Metadata**: Minimal; capture standard HTTP attributes to avoid leaking PII.
- **Metrics**: Basic (counts, durations, error rates).
- **Configuration**: Explicit `WithTracerProvider(tp)` and `WithMeterProvider(mp)`.
- **Money**: Idiomatic Cents (`int64`).
- **IDs**: Plain `string`.
- **Optionality**: Use a modern `Optional[T]` generic wrapper.
- **Pattern**: Unified Functional Options for pagination and filtering.

### Claude's Discretion
- OpenTelemetry implementation details and `Optional[T]` design.

### Deferred Ideas (OUT OF SCOPE)
- **Auto-Paging/Iterators**: Not planned for this phase.
- **Advanced Retry Logic**: Custom `CheckRetry` functions are not exposed.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| OBS-03 | OpenTelemetry (OTel) integration | Use `otelhttp` for transport and `otel.Tracer` for services. |
| HTTP-01 | Robust retry logic with exponential backoff | `go-retryablehttp` provides this out-of-the-box. |
| HTTP-02 | Proper handling of Rate Limiting (429) | `go-retryablehttp` handles 429 retries automatically. |
| ACC-01 | Retrieve account balances | `GET /v2/accounts/{id}/balances` endpoint identified. |
| ACC-02 | Retrieve transaction history | `GET /v2/accounts/{id}/transactions` endpoint identified. |
| ACC-03 | Retrieve specific transaction details | Part of the transactions/accounts API set. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `go-retryablehttp` | v0.7.7 | Automatic retries | De facto standard for Go HTTP clients needing reliability. |
| `otelhttp` | v0.49.0 | HTTP Instrumentation | Official OTel contribution for net/http. |
| `otel` | v1.24.0 | Tracing/Metrics API | Standard observability framework. |

## Architecture Patterns

### Pattern 1: Retryable Transport Wrapper
The SDK will use `retryablehttp` internally but expose a standard `*http.Client`.
```go
retryClient := retryablehttp.NewClient()
retryClient.Logger = logger // slog.Logger implements LeveledLogger
client := retryClient.StandardClient()
```

### Pattern 2: Service-Level Instrumentation
Wrap business logic in spans using the tracer provided via functional options.
```go
ctx, span := p.tracer.Start(ctx, "payoneer.account.list_balances", trace.WithAttributes(...))
defer span.End()
```

### Pattern 3: Optional[T] Wrapper
```go
type Optional[T any] struct {
    Value T
    Valid bool
}
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Retry Backoff | Custom loops with `time.Sleep` | `go-retryablehttp` | Handles jitter, backoff, and 429/5xx logic correctly. |
| HTTP Tracing | Manual header injection/extraction | `otelhttp` | Implements W3C Trace Context and standard attributes automatically. |

## Common Pitfalls

### Pitfall 1: Leaking PII in Traces
**What goes wrong:** Adding `AccountID` or `Amount` to span attributes in public/shared collectors.
**How to avoid:** Use "Minimal Metadata" strategy. Only log generic HTTP attributes and anonymized IDs.

### Pitfall 2: Double Wrapping Transports
**What goes wrong:** Nesting `otelhttp.Transport` and `retryablehttp.RoundTripper` incorrectly.
**How to avoid:** Use `retryablehttp.StandardClient()` first, then wrap its `Transport` with `otelhttp.NewTransport`.

## Code Examples

### Optional[T] Implementation
```go
type Optional[T any] struct {
	value T
	ok    bool
}

func Some[T any](v T) Optional[T] { return Optional[T]{value: v, ok: true} }
func None[T any]() Optional[T]    { return Optional[T]{} }

func (o Optional[T]) Get() (T, bool) { return o.value, o.ok }
func (o Optional[T]) OrDefault(d T) T {
	if o.ok { return o.value }
	return d
}
```

## Sources
- [Official Payoneer Services API v2 Docs] - Account, Balance, Transaction endpoints.
- [Hashicorp go-retryablehttp GitHub] - Logger interfaces and StandardClient usage.
- [OpenTelemetry Go Documentation] - otelhttp and tracer usage.
