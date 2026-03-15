# Validation: Phase 3 - Financial Operations (Payouts)

**Phase Goal:** Enable core payout capabilities with business validation.

## Observable Truths

- [ ] "Developer can configure a ProgramID on the client"
- [ ] "Developer can submit a mass payout request with multiple items"
- [ ] "SDK rejects invalid payout requests locally (e.g., missing reference ID)"
- [ ] "Developer can retrieve the status of a specific payout"
- [ ] "Developer can cancel a pending payout"
- [ ] "SDK identifies business failures in 200 OK responses"

## Required Artifacts

| Artifact | Purpose | Status |
|----------|---------|--------|
| `pkg/payoneer/client.go` | Client struct has `ProgramID` field. | [ ] |
| `pkg/payoneer/options.go` | `WithProgramID` option exists. | [ ] |
| `pkg/payoneer/payout_types.go` | Defines `MassPayoutRequest` and `PayoutItem`. | [ ] |
| `pkg/payoneer/payout_types_test.go` | Tests for payout models and validation. | [ ] |
| `pkg/payoneer/payouts.go` | `PayoutsService` implementation. | [ ] |
| `pkg/payoneer/payouts_test.go` | Comprehensive test suite for Payout operations. | [ ] |

## Key Links

| From | To | Via | Pattern |
|------|----|-----|---------|
| `pkg/payoneer/payouts.go` | `/v4/programs/{program_id}/masspayouts` | `POST` request | `post.*masspayouts` |
| `pkg/payoneer/payouts.go` | `/v4/programs/{program_id}/payouts/{id}/status` | `GET` request | `get.*payouts.*status` |
| `pkg/payoneer/payouts.go` | `/v4/programs/{program_id}/payouts/{id}/cancel` | `POST` request | `post.*payouts.*cancel` |

## Automated Verification Suite

Run the full payout test suite to ensure all behaviors are correctly implemented:

```bash
go test -v ./pkg/payoneer/ -run TestPayoutsService
go test -v ./pkg/payoneer/ -run TestMassPayoutRequest_Validate
go test -v ./pkg/payoneer/ -run TestPayoutErrorHandling
```

## Manual Verification (if needed)

If integration with Sandbox is available:
1. Initialize client with a valid Sandbox `ProgramID`.
2. Submit a `MassPayoutRequest` with dummy data.
3. Observe successful batch acceptance.
4. Retrieve status for the submitted `client_reference_id`.
