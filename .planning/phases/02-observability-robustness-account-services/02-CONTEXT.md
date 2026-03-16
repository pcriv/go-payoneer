# Context: Phase 2 - Observability, Robustness & Account Services

## Phase Goal
Implement reliable account data retrieval with full telemetry, building on the Phase 1 client foundation.

## Implementation Decisions

### [Robustness] Retry Policy & Configuration
- **Library**: Use `github.com/hashicorp/go-retryablehttp`.
- **Configuration**: Simple exposure via `WithRetries(max int)` functional option.
- **Integration**: Internal wrapping; `Client.HTTPClient` remains a standard `*http.Client` pointing to the retryable client.
- **Logging**: Unified; route `go-retryablehttp` logs through the SDK's existing redacted `slog.Logger`.
- **Control**: Fully automatic retries based on status codes (e.g., 429, 5xx), leveraging Payoneer's idempotency.

### [Observability] OpenTelemetry Strategy
- **Instrumentation**: Hybrid approach.
    - **Transport-level**: Automatic spans/metrics for all HTTP activity via `otelhttp`.
    - **Service-level**: Semantic spans for business operations (e.g., `payoneer.account.list_balances`).
- **Metadata**: Minimal; capture standard HTTP attributes (method, URL, status code) to avoid leaking PII in traces.
- **Metrics**: Basic; track request counts, durations, and error rates.
- **Configuration**: Explicit; provide `WithTracerProvider(tp)` and `WithMeterProvider(mp)` functional options.

### [Account Services] API Data Modeling
- **Money**: Idiomatic Cents; represent amounts as `int64` minor units.
- **IDs**: Use plain `string` for `AccountID` and related identifiers.
- **List Envelope**: Metadata-rich; return structs (e.g., `AccountList`) containing both `Items []Account` and metadata like `Total`.
- **Optionality**: Use a modern `Optional[T]` generic wrapper for nullable/optional API fields.

### [Account Services] Pagination & Filtering
- **Pattern**: Unified Functional Options for both pagination and filtering (e.g., `WithPage`, `WithFrom`, `WithStatus`).
- **Date Handling**: Strict `time.Time`; the SDK handles conversion to/from Payoneer's ISO 8601 strings.
- **Result Sets**: Manual pagination (Limit/Offset); users must explicitly request subsequent pages to manage API costs and rate limits.

## Code Context
- **Base Pattern**: Continue using Functional Options (`Option` type in `pkg/payoneer/options.go`).
- **Transport**: Middleware should be added to `LoggingTransport` or chained in `NewClient`.
- **Services**: `AccountsService` in `pkg/payoneer/service.go` is the primary implementation target.

## Deferred Ideas (Out of Scope)
- **Auto-Paging/Iterators**: Not planned for this phase; may be added as a convenience layer later.
- **Advanced Retry Logic**: Custom `CheckRetry` functions are not exposed to keep the API simple for now.
