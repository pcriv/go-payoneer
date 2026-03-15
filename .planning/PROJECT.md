# go-payoneer

## What This Is

A Go client/SDK for the Payoneer API, designed for developers who need a robust, performant, and observable way to integrate Payoneer services. It leverages modern Go idioms like functional options, slog for logging, and OpenTelemetry (OTel) for observability.

## Core Value

Provide a high-quality, type-safe, and observable Go SDK for Payoneer that minimizes boilerplate and maximizes reliability.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Functional options pattern for client and request configuration
- [ ] Comprehensive coverage of Payoneer API endpoints
- [ ] Integrated logging using slog
- [ ] Observability with OpenTelemetry (tracing/metrics)
- [ ] Robust retryable HTTP client with backoff
- [ ] Proper error handling and type-safe responses

### Out of Scope

- [ ] Mobile app integration (outside the scope of a Go SDK)
- [ ] Direct UI components (this is an SDK/library)

## Context

The SDK will be used in backend Go applications. It needs to handle authentication, rate limiting, and various Payoneer API versions if applicable.

## Constraints

- **Language**: Go 1.21+ (for slog support)
- **Dependencies**: OpenTelemetry Go SDK, slog, retryable-http (or custom)
- **Style**: Idiomatic Go, functional options

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Functional Options | Standard Go pattern for configurable clients | — Pending |
| slog | Built-in Go logging library (Go 1.21+) | — Pending |
| OTel | Standard for cloud-native observability | — Pending |

---
*Last updated: 2026-03-15 after project initialization*
