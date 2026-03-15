# Requirements: go-payoneer

**Defined:** 2026-03-15
**Core Value:** Provide a high-quality, type-safe, and observable Go SDK for Payoneer that minimizes boilerplate and maximizes reliability.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Authentication & Client Foundation (Phase 1)

- [x] **AUTH-01**: Support OAuth 2.0 Client Credentials flow for application-level access
- [x] **AUTH-02**: Support OAuth 2.0 Authorization Code flow for user-level access
- [x] **AUTH-03**: Automatic token refreshing using `golang.org/x/oauth2` or similar thread-safe mechanism
- [x] **CLNT-01**: Functional options pattern for client configuration (BaseURL, Timeout, UserAgent, etc.)
- [x] **CLNT-02**: Support for Sandbox and Production environments
- [x] **CLNT-03**: Context-first API design (all methods accept `context.Context`)
- [x] **ERR-01**: Custom `APIError` type that captures Payoneer-specific error codes and messages from response bodies (even on 200 OK)

### Observability & Robustness (Phase 1/2)

- [x] **OBS-01**: Integrated structured logging using `log/slog` with support for custom loggers
- [x] **OBS-02**: Redaction of sensitive information (tokens, secrets) in logs via custom `RedactionHandler`
- [ ] **OBS-03**: OpenTelemetry (OTel) integration for tracing and metrics
- [ ] **HTTP-01**: Robust retry logic with exponential backoff (e.g., using `go-retryablehttp`)
- [ ] **HTTP-02**: Proper handling of Rate Limiting (429) status codes and headers

### Account Services (Phase 2)

- [ ] **ACC-01**: Retrieve account balances for all supported currencies
- [ ] **ACC-02**: Retrieve transaction history with filtering and pagination
- [ ] **ACC-03**: Retrieve specific transaction details by ID

### Financial Operations (Phase 3)

- [ ] **PAY-01**: Submit Mass Payout requests with mandatory idempotency/reference IDs
- [ ] **PAY-02**: Retrieve payout status and details
- [ ] **PAY-03**: Cancel pending payouts (if supported by endpoint)
- [ ] **PAY-04**: Handle "False Positive" 200 OK responses by validating business status in body

### Webhooks & Payees (Phase 4)

- [ ] **WEB-01**: Middleware/Handler for receiving and parsing Payoneer IPCN (webhooks)
- [ ] **WEB-02**: Mandatory HMAC SHA-256 signature verification for all incoming webhooks
- [ ] **PYE-01**: Create and manage Payee registration links
- [ ] **PYE-02**: Retrieve payee status and information

## v2 Requirements

### Advanced Features

- **DIST-01**: Out-of-the-box support for distributed token storage (e.g., Redis)
- **MOCK-01**: Built-in mock client/server for testing user applications without hitting Sandbox
- **CLI-01**: A companion CLI tool for common Payoneer operations and debugging

## Out of Scope

| Feature | Reason |
|---------|--------|
| UI Components | This is a backend SDK/library |
| Mobile SDK | Out of scope for a Go library |
| Direct Database Integration | Users should manage their own persistence |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 1 | Complete |
| AUTH-02 | Phase 1 | Complete |
| AUTH-03 | Phase 1 | Complete |
| CLNT-01 | Phase 1 | Complete |
| CLNT-02 | Phase 1 | Complete |
| CLNT-03 | Phase 1 | Complete |
| ERR-01 | Phase 1 | Complete |
| OBS-01 | Phase 1 | Complete |
| OBS-02 | Phase 1 | Complete |
| OBS-03 | Phase 2 | Pending |
| HTTP-01 | Phase 2 | Pending |
| HTTP-02 | Phase 2 | Pending |
| ACC-01 | Phase 2 | Pending |
| ACC-02 | Phase 2 | Pending |
| ACC-03 | Phase 2 | Pending |
| PAY-01 | Phase 3 | Pending |
| PAY-02 | Phase 3 | Pending |
| PAY-03 | Phase 3 | Pending |
| PAY-04 | Phase 3 | Pending |
| WEB-01 | Phase 4 | Pending |
| WEB-02 | Phase 4 | Pending |
| PYE-01 | Phase 4 | Pending |
| PYE-02 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 23 total
- Mapped to phases: 23
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-15*
*Last updated: 2026-03-15 after Phase 1 completion*
