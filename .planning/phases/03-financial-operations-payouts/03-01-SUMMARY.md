---
phase: 03-financial-operations-payouts
plan: 01
subsystem: Payouts
tags: [auth, models, validation]
dependency_graph:
  requires: [AUTH-01, AUTH-02, OBS-01]
  provides: [PAY-01]
  affects: [PayoutsService]
tech_stack:
  added: []
  patterns: [Functional Options, Validation Method, Mandatory Idempotency]
key_files:
  created:
    - pkg/payoneer/payout_types.go
    - pkg/payoneer/payout_types_test.go
  modified:
    - pkg/payoneer/client.go
    - pkg/payoneer/options.go
decisions:
  - id: PAY-IDEMP-01
    name: Mandatory Idempotency via ClientReferenceID
    summary: Enforced at the struct level by requiring it in all payout operations.
metrics:
  duration: 15m
  completed_date: "2026-03-15"
---

# Phase 03 Plan 01: Client Updates & Payout Models Summary

## Substantive Summary
Implemented core structural support for Payout operations. This includes adding `ProgramID` configuration to the `Client` and `WithProgramID` functional option. Defined robust request and response models for Mass Payouts with a focus on type safety and mandatory idempotency via `client_reference_id`. A local structural validation method `Validate()` was added to `MassPayoutRequest` to catch errors (empty payments, >500 items, non-positive amounts) before making API calls.

## Key Changes
- **Client Configuration**: Added `ProgramID` to the `Client` struct and provided a functional option `WithProgramID` for its initialization.
- **Payout Models**: Created `PayoutItem`, `MassPayoutRequest`, and `PayoutBatchResult` structs.
- **Custom JSON Marshaling**: Implemented `MarshalJSON` and `UnmarshalJSON` for `PayoutItem` to handle the conversion between cents (`int64`) and decimal strings (`float64` in JSON) as expected by the Payoneer v4 API.
- **Validation**: Added a `Validate()` method to `MassPayoutRequest` to ensure batch size and field integrity.

## Deviations from Plan
None - plan executed exactly as written. Task 1 and Task 2 were already partially implemented in the environment and were completed and verified along with Task 3.

## Verification Results
- All tests in `pkg/payoneer/payout_types_test.go` passed, covering serialization and validation.
- All tests in `pkg/payoneer/client_test.go` passed, verifying the new configuration option.
- Full project test suite passed without regressions.

## Self-Check: PASSED
- [x] Client struct contains ProgramID.
- [x] WithProgramID option works.
- [x] MassPayoutRequest.Validate() correctly identifies invalid requests.
- [x] Commits 7bef7d3 and 069e427 exist in history.
