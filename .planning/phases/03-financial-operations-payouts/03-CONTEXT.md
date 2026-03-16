# Context: Phase 3 - Financial Operations (Payouts)

## Phase Goal
Enable core payout capabilities with business validation, building on the observability and account services established in Phase 2.

## Implementation Decisions

### [Payouts] Idempotency Strategy
- **Decision**: **User-Controlled / Mandatory**.
- **Pattern**: The `client_reference_id` (or similar idempotency key) must be a required field in the payout request struct.
- **Rationale**: Financial operations require explicit correlation IDs from the caller's system to prevent accidental double-payouts during retries or logic errors.

### [Payouts] Mass Payout Request Modeling
- **Decision**: **Plain Structs with Validation**.
- **Pattern**: Use a `MassPayoutRequest` struct containing a slice of `PayoutItem`.
- **Validation**: Implement a `(r *MassPayoutRequest) Validate() error` method to perform pre-flight checks (e.g., non-zero amounts, required payee fields) before sending the request to the API.

### [Payouts] Business Status Validation
- **Decision**: **Post-Process (Result + Error)**.
- **Pattern**: 
    - The SDK's `Do()` method handles transport errors (4xx/5xx) and top-level API "Failure" statuses via `ValidateResponse`.
    - The `PayoutsService.CreateMassPayout` method returns a `PayoutBatchResult` struct and an `error`.
    - The `error` is reserved for API/network failures.
    - The `PayoutBatchResult` contains the status of individual payout items, which the user must inspect for "partial success" scenarios.

### [Payouts] Payout Cancellation
- **Decision**: **Explicit Method**.
- **Pattern**: Provide a dedicated `Cancel(ctx, payoutID string)` method in the `PayoutsService`.
- **Rationale**: Maximizes discoverability and provides a clear intent-based API for users.

## Code Context
- **Base Pattern**: Continue using Functional Options for optional request parameters.
- **Data Types**: Use `int64` (cents/minor units) for all payout amounts, consistent with Phase 2 decisions.
- **Optionality**: Use the `Optional[T]` generic wrapper for nullable fields in payout responses (e.g., `Reason`, `ReleaseDate`).
- **Service**: Implement logic within `PayoutsService` in `pkg/payoneer/payouts.go` (new file) or extending `service.go`.

## Deferred Ideas (Out of Scope)
- **Payout Batch Builder**: A fluent/builder API for constructing batches is deferred; plain structs are the priority.
- **Auto-generated Idempotency**: The SDK will not automatically generate reference IDs to ensure user accountability for transaction mapping.
