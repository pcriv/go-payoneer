# Phase 3: Financial Operations (Payouts) - Research

**Researched:** 2026-03-15
**Domain:** Payoneer Mass Payouts API (v4)
**Confidence:** HIGH

## Summary

Phase 3 focuses on implementing core payout capabilities using the Payoneer Mass Payouts API v4. This API is inherently asynchronous and batch-oriented. A key characteristic is its "all-or-nothing" validation for batch submission: if one item in a batch fails basic field validation, the entire batch is rejected with a `400 Bad Request`. However, once a batch is accepted (`200 OK`), individual payments are processed asynchronously, and failures at this stage (e.g., insufficient funds, blocked payee) are reported via webhooks or must be discovered via status polling.

**Primary recommendation:** Implement `PayoutsService` with dedicated methods for mass payout submission, status retrieval, and cancellation. Mandatory idempotency via `client_reference_id` must be enforced at the struct level.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Idempotency Strategy**: User-Controlled / Mandatory. The `client_reference_id` must be a required field in the payout request struct.
- **Mass Payout Request Modeling**: Plain Structs with Validation. Use `MassPayoutRequest` containing a slice of `PayoutItem`. Implement a `Validate() error` method.
- **Business Status Validation**: Post-Process (Result + Error). The SDK handles transport errors; `CreateMassPayout` returns `PayoutBatchResult` and `error`. User must inspect `PayoutBatchResult` for individual item statuses.
- **Payout Cancellation**: Dedicated `Cancel(ctx, payoutID string)` method.
- **Data Types**: Use `int64` (cents/minor units) for all payout amounts.
- **Optionality**: Use the `Optional[T]` generic wrapper for nullable fields in payout responses.
- **Service**: Implement logic within `PayoutsService` in `pkg/payoneer/payouts.go`.

### Claude's Discretion
- Base Pattern: Continue using Functional Options for optional request parameters.

### Deferred Ideas (OUT OF SCOPE)
- **Payout Batch Builder**: A fluent/builder API for constructing batches is deferred.
- **Auto-generated Idempotency**: The SDK will not automatically generate reference IDs.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| PAY-01 | Submit Mass Payout requests with mandatory idempotency/reference IDs | Verified `POST /v4/programs/{program_id}/masspayouts` endpoint and `client_reference_id` requirement. |
| PAY-02 | Retrieve payout status and details | Verified `GET /v4/programs/{program_id}/payouts/{client_reference_id}/status` endpoint. |
| PAY-03 | Cancel pending payouts | Verified `POST /v4/programs/{program_id}/payouts/{client_reference_id}/cancel` endpoint. |
| PAY-04 | Handle "False Positive" 200 OK responses by validating business status in body | Identified that 200 OK only means batch acceptance; async failures require polling/webhooks. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go Standard Library | 1.21+ | Core logic and networking | Reliability and performance |
| `github.com/hashicorp/go-retryablehttp` | v0.7 | Robust HTTP retries | Essential for financial operations to handle transient failures |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/stretchr/testify` | v1.8+ | Testing assertions | Unit and integration tests |

## Architecture Patterns

### Recommended Project Structure
```
pkg/payoneer/
├── payouts.go       # PayoutsService implementation
└── payout_types.go  # Structs for requests and responses
```

### Pattern 1: Mandatory Idempotency
**What:** Require `client_reference_id` in the constructor or struct fields for all payout operations.
**When to use:** All `PayoutItem` and `Cancel` requests.
**Example:**
```typescript
type PayoutItem struct {
    ClientReferenceID string `json:"client_reference_id"`
    PayeeID           string `json:"payee_id"`
    Amount            int64  `json:"-"` // Internal use (cents)
    Currency          string `json:"currency"`
    Description       string `json:"description,omitempty"`
}
```

### Anti-Patterns to Avoid
- **Auto-generating IDs:** The user must provide the ID to ensure they can correlate the payout with their own internal database state.
- **Assuming 200 OK means Paid:** Always check the status or wait for webhooks.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Money Formatting | Manual string formatting | `fmt.Sprintf("%.2f", float64(cents)/100.0)` | Payoneer expects decimal strings. |
| ISO Currency Codes | Custom validation | Simple string constants or user-provided strings | Currency list is stable but Payoneer's accepted list may vary by program. |
| Idempotency Keys | UUID generators | User-provided IDs | User accountability for transaction mapping. |

## Common Pitfalls

### Pitfall 1: Batch Rejection
**What goes wrong:** A single invalid item (e.g., negative amount) causes the entire batch of 500 to fail.
**How to avoid:** Implement strict `Validate()` logic in the SDK to catch formatting errors before sending.

### Pitfall 2: Async Failure Discovery
**What goes wrong:** A payout is accepted (200 OK) but fails due to "Insufficient Funds" seconds later.
**How to avoid:** Instruct users to use the `GetStatus` method or implement a webhook listener (Phase 4).

### Pitfall 3: Payout Not Found (404/2306)
**What goes wrong:** Querying status for an item that failed initial async processing returns 404.
**How to avoid:** Treat 404/2306 as a terminal failure of the payout creation attempt.

## Code Examples

### Mass Payout Request Validation
```go
func (r *MassPayoutRequest) Validate() error {
    if len(r.Payments) == 0 {
        return errors.New("at least one payment is required")
    }
    if len(r.Payments) > 500 {
        return errors.New("maximum 500 payments allowed per batch")
    }
    for i, p := range r.Payments {
        if p.ClientReferenceID == "" {
            return fmt.Errorf("payment %d: client_reference_id is required", i)
        }
        if p.Amount <= 0 {
            return fmt.Errorf("payment %d: amount must be greater than zero", i)
        }
    }
    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Payouts v2/v3 | Payouts v4 (Mass Payouts) | 2021+ | Improved batching and mandatory OAuth2. |

## Open Questions

1. **Max length of `client_reference_id`?**
   - What we know: It must be unique.
   - What's unclear: Official docs don't specify a hard limit, but 50-100 chars is typical for Payoneer.
   - Recommendation: Allow arbitrary string but document that 50 is a safe limit.

2. **Program ID availability?**
   - What we know: All v4 endpoints require `program_id`.
   - What's unclear: Currently not in the `Client` struct.
   - Recommendation: Add `WithProgramID(string)` option to the client.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | `testing` + `testify` |
| Config file | None — Standard Go |
| Quick run command | `go test -v ./pkg/payoneer/...` |
| Full suite command | `go test -v ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PAY-01 | Submit Mass Payout | integration | `go test ./pkg/payoneer/ -run TestCreateMassPayout` | ❌ Wave 0 |
| PAY-02 | Get Payout Status | integration | `go test ./pkg/payoneer/ -run TestGetPayoutStatus` | ❌ Wave 0 |
| PAY-03 | Cancel Payout | integration | `go test ./pkg/payoneer/ -run TestCancelPayout` | ❌ Wave 0 |
| PAY-04 | Handle 404/2306 | unit | `go test ./pkg/payoneer/ -run TestPayoutErrorHandling` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./pkg/payoneer/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `pkg/payoneer/payouts_test.go` — Mocked and integration tests for payout operations.
- [ ] `WithProgramID` option in `pkg/payoneer/options.go`.

## Sources

### Primary (HIGH confidence)
- Payoneer Mass Payouts API v4 Documentation - Endpoints and JSON structures.
- Official Postman Collection for Payoneer v4 - Request/Response examples.

### Secondary (MEDIUM confidence)
- Community forums (Reddit/StackOverflow) - Behavior of 404/2306 errors.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH
- Architecture: HIGH
- Pitfalls: HIGH

**Research date:** 2026-03-15
**Valid until:** 2026-06-15
