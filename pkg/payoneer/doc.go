/*
Package payoneer provides a comprehensive Go SDK for the Payoneer API.

It follows a service-oriented client design with functional options for configuration
and robust support for OAuth 2.0 authentication (Client Credentials and Authorization Code flows).

Features:
  - Payee Management: Generate onboarding registration links and track payee status.
  - Financial Operations: Submit mass payouts with mandatory idempotency and validation.
  - Webhooks (IPCN): Securely receive and validate HMAC SHA-256 signed notifications.
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

	err := client.Authenticate(ctx)

For more details on specific services, refer to the corresponding service types:
  - PayeesService for registration links and status.
  - PayoutsService for mass payouts.
  - WebhookValidator for HTTP middleware and signature validation.
*/
package payoneer
