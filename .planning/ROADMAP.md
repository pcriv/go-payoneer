# Roadmap: go-payoneer

## Phases

- [x] **Phase 1: Client Foundation & Authentication** - Establish the core SDK structure and secure OAuth 2.0 integration. (Completed 2026-03-15)
- [ ] **Phase 2: Observability, Robustness & Account Services** - Implement reliable account data retrieval with full telemetry.
- [ ] **Phase 3: Financial Operations (Payouts)** - Enable core payout capabilities with business validation.
- [ ] **Phase 4: Webhooks & Payee Management** - Complete the SDK with asynchronous notifications and payee onboarding.

## Phase Details

### Phase 1: Client Foundation & Authentication
**Goal**: Establish the core SDK structure and secure OAuth 2.0 integration.
**Depends on**: Nothing
**Requirements**: AUTH-01, AUTH-02, AUTH-03, CLNT-01, CLNT-02, CLNT-03, ERR-01, OBS-01, OBS-02
**Success Criteria**:
1. ✓ Developer can initialize the client using functional options for Sandbox/Production.
2. ✓ Client successfully performs OAuth 2.0 Client Credentials and Authorization Code flows.
3. ✓ Tokens are automatically refreshed when expired without user intervention.
4. ✓ All outgoing requests include proper logging (with sensitive data redacted).
5. ✓ Non-2xx responses are parsed into custom `APIError` types.

**Plans:** 3/3 plans executed
- [x] 01-01-PLAN.md — SDK Foundation & Client Initialization
- [x] 01-02-PLAN.md — OAuth 2.0 & Token Management
- [x] 01-03-PLAN.md — Error Handling & Observability Middleware

### Phase 2: Observability, Robustness & Account Services
**Goal**: Implement reliable account data retrieval with full telemetry.
**Depends on**: Phase 1
**Requirements**: OBS-03, HTTP-01, HTTP-02, ACC-01, ACC-02, ACC-03
**Success Criteria**:
1. Client automatically retries failed requests with exponential backoff and handles 429 rate limits.
2. Outgoing requests and incoming responses generate OpenTelemetry traces and metrics.
3. User can retrieve account balances for multiple currencies.
4. User can list, filter, and paginate transaction history.
**Plans**: TBD

### Phase 3: Financial Operations (Payouts)
**Goal**: Enable core payout capabilities with business validation.
**Depends on**: Phase 2
**Requirements**: PAY-01, PAY-02, PAY-03, PAY-04
**Success Criteria**:
1. User can submit a mass payout request with mandatory idempotency keys.
2. Payout status and details can be retrieved by ID or cancelled if pending.
3. SDK correctly identifies "False Positive" 200 OK responses by checking the business status in the response body.
**Plans**: TBD

### Phase 4: Webhooks & Payee Management
**Goal**: Complete the SDK with asynchronous notifications and payee onboarding.
**Depends on**: Phase 3
**Requirements**: WEB-01, WEB-02, PYE-01, PYE-02
**Success Criteria**:
1. Developer can register a webhook handler that validates HMAC SHA-256 signatures.
2. Webhook payloads are automatically parsed into typed Go structures.
3. User can generate payee registration links and track payee status.
**Plans**: TBD

## Progress Table

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Client Foundation & Authentication | 3/3 | Complete | 2026-03-15 |
| 2. Observability, Robustness & Account Services | 0/1 | Not started | - |
| 3. Financial Operations (Payouts) | 0/1 | Not started | - |
| 4. Webhooks & Payee Management | 0/1 | Not started | - |

---
*Roadmap generated: 2026-03-15*
*Last updated: 2026-03-15 after Phase 1 completion*
