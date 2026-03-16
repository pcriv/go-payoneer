# Summary: Phase 4 - Plan 01 (Webhook Foundation)

## Objective
Implement the foundation for receiving and validating Payoneer webhooks (IPCN), including HMAC SHA-256 signature verification and an HTTP middleware.

## Deliverables
- `pkg/payoneer/webhooks.go`: Core logic for signature validation and payload parsing.
- `pkg/payoneer/webhooks_test.go`: Comprehensive tests for validation, parsing, and middleware.

## Key Features
- **HMAC SHA-256 Validation**: Securely verifies `X-Payoneer-Signature` using the Merchant Secret.
- **Webhook Models**: Typed `WebhookEvent` for standard IPCN payloads.
- **Validation Middleware**: Easy-to-use `WebhookValidator` that protects endpoints and ensures the request body remains readable for downstream handlers.

## Verification
- [x] `TestValidateSignature`: Verified algorithm correctness.
- [x] `TestParseWebhook`: Verified header extraction and JSON unmarshaling.
- [x] `TestWebhookMiddleware`: Verified unauthorized access is blocked and valid requests proceed with body restored.

All tests passed. Phase 4 Plan 02 (Payee Management) can now proceed.
