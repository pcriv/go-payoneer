# Context: Phase 4 - Webhooks & Payee Management

## Phase Goal
Complete the SDK with asynchronous event notifications (webhooks) and payee onboarding/lifecycle management, building on the authentication and financial foundations established in prior phases.

## Implementation Decisions

### [Webhooks] Signature Validation & Middleware
- **Decision**: **Standard Middleware with Standalone Helper**.
- **Pattern**: 
    - Provide a standalone `ValidateSignature(payload []byte, signature string, secret string) bool` function for flexibility.
    - Provide an `http.Handler` middleware (e.g., `WebhookValidator(secret string) func(http.Handler) http.Handler`) that automatically validates the `X-Payoneer-Signature` header (HMAC SHA-256).
- **Rationale**: Ensures security is easy to implement for standard Go HTTP servers while remaining compatible with serverless or custom router environments.

### [Webhooks] Payload Dispatching
- **Decision**: **Event Interface / Type Switch**.
- **Pattern**: 
    - Implement a `ParseWebhook(r *http.Request) (WebhookEvent, error)` function.
    - `WebhookEvent` is an interface implemented by specific event structs (e.g., `PayoutStatusChanged`, `PayeeRegistrationCompleted`).
    - Users use a standard Go `switch e := event.(type)` to handle specific business logic for each event.
- **Rationale**: Most idiomatic Go approach for polymorphic data, providing full type safety for event-specific fields.

### [Payees] Registration Link Strategy
- **Decision**: **Simple URL Generator**.
- **Pattern**: Provide a method `c.Payees.RegistrationURL(payeeID string, opts ...LinkOption)` that returns a `string`.
- **Options**: Support functional options for `WithRedirectURL(url)`, `WithLanguage(lang)`, etc.
- **Rationale**: Developers need a flexible way to generate these links for emails, SMS, or frontend "Connect Payoneer" buttons without SDK-enforced redirection logic.

### [Payees] Management Service Depth
- **Decision**: **Resource-Centric Lifecycle**.
- **Pattern**: Implement a standard service with `Get(ctx, id)`, `List(ctx, opts)`, and `Create(ctx, req)` methods.
- **Scope**: Focus on the core onboarding lifecycle (Creation, Status Tracking, Retrieval). Advanced document management or bank account updates are deferred to v2 unless explicitly required by the v1 Payouts flow.

## Code Context
- **Base Pattern**: Continue using Functional Options for optional request parameters and URL generation.
- **Security**: Webhook logic should reside in a new `pkg/payoneer/webhooks` package or similar to isolate HTTP/Crypto concerns.
- **Data Types**: Use the `Optional[T]` generic wrapper for nullable fields in Payee and Webhook payloads.
- **Service**: Implement logic within `PayeesService` in `pkg/payoneer/payees.go` (new file).

## Deferred Ideas (Out of Scope)
- **Automatic Webhook Registration**: Users must manage their webhook URLs through the Payoneer portal; the SDK only handles *receiving* events.
- **Payee Document Uploads**: High-complexity KYC flows are deferred to a future phase.
- **Email Triggers**: The SDK generates the link; the user is responsible for the delivery mechanism (email/SMS).
