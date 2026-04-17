/*
Package payoneer provides a comprehensive Go SDK for the Payoneer API.

It follows a service-oriented client design with functional options for configuration
and robust support for OAuth 2.0 authentication (Client Credentials and Authorization Code flows).

Features:
  - Payee Management: Generate onboarding registration links and track payee status.
  - Financial Operations: Submit mass payouts with mandatory idempotency and validation.
  - Webhooks (IPCN): Verify the `Authorization: hmacauth` header, including HMAC-SHA256
    signature, timestamp skew, and optional nonce-based replay protection.
  - Observability: Built-in support for slog structured logging and OpenTelemetry.
  - Robustness: Automatic retries with exponential backoff and 429 rate-limit handling.

Initialization:

	client := payoneer.NewClient(
	    payoneer.WithSandbox(),
	    payoneer.WithProgramID("my-program-id"),
	    payoneer.WithClientCredentials("my-client-id", "my-client-secret"),
	    payoneer.WithLogger(myLogger),
	)

Authentication:

The client authenticates lazily on the first API call, so in typical usage
no explicit step is required. Call Authenticate to surface credential errors
at startup instead of on first request:

	if err := client.Authenticate(ctx); err != nil {
	    // credentials are wrong; fail fast
	}

For more details on specific services, refer to the corresponding service types:
  - PayeesService for registration links and status.
  - PayoutsService for mass payouts.
  - WebhookValidator for HTTP middleware and signature validation.
*/
package payoneer
