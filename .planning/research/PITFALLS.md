# Domain Pitfalls: Go SDK & Payoneer Integration

**Domain:** Fintech / Payment Processing (Payoneer API)
**Researched:** 2024-05-24
**Overall Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: "False Positive" HTTP 200 OK Responses
**What goes wrong:** The API returns an `HTTP 200 OK` status, but the business operation (e.g., a payout or payment) has actually failed or been aborted.
**Why it happens:** Payoneer's API gateway separates communication success from application-level logic. If the request reached the server, it may return 200 even if the transaction was declined.
**Consequences:** The SDK user assumes a payment was successful when it wasn't, leading to reconcilation nightmares and missed payments.
**Prevention:**
- **Always** parse the response body regardless of the HTTP status.
- For Checkout API: Check `interaction.code == "PROCEED"`.
- For Mass Payout API: Check `status == 2` (or "success") and look for internal `error_code` fields.
**Detection:** Transactions marked as "Success" in the local database but "Failed" or "Pending" in the Payoneer dashboard.

### Pitfall 2: Credential Leaks via Request/Response Logging
**What goes wrong:** Sensitive credentials (`client_secret`, OAuth `Bearer` tokens) are written to application logs or APM systems.
**Why it happens:** Standard practice is to log full request/response bodies for debugging, especially during early integration phases.
**Consequences:** Unauthorized access to the Payoneer account, potential for massive financial loss through fraudulent payouts.
**Prevention:**
- Implement a custom `RoundTripper` in Go that redacts the `Authorization` header.
- Use structured logging (e.g., `slog`) and explicitly mask fields named `client_secret`, `access_token`, or `refresh_token`.
- Never log the raw body of token exchange requests.
**Detection:** Auditing logs for "Bearer" strings or 64-character hex strings.

### Pitfall 3: Idempotency Neglect (Double Payouts)
**What goes wrong:** A payout is sent twice due to a timeout or a 5xx error that was retried without an idempotency key.
**Why it happens:** Developers may not realize that a "timed out" request might have actually succeeded on the Payoneer side.
**Consequences:** Paying the same recipient twice, which is often difficult to recover.
**Prevention:**
- **Always** include a unique `Client Reference ID` (or `Idempotency-Key` if supported by the specific endpoint) for every `POST` request.
- If a request fails with a 5xx error or a timeout, **retry with the exact same ID**.
**Detection:** Duplicate transaction entries in the Payoneer Partner Admin console with different Payoneer IDs but similar amounts/recipients.

---

## Moderate Pitfalls

### Pitfall 1: Thread-Unsafe Token Management
**What goes wrong:** Multiple goroutines attempt to refresh an expired OAuth token simultaneously, causing race conditions or invalidating the token for other routines.
**Why it happens:** Using a shared `Client` struct across goroutines without proper synchronization for the internal token state.
**Prevention:**
- Use `sync.RWMutex` to protect the token field within the client.
- Alternatively, use the `golang.org/x/oauth2` package, which handles thread-safe token refreshing automatically via its `TokenSource` interface.
- Ensure the client receiver is always a pointer: `func (c *Client) Request(...)`.

### Pitfall 2: Static Redirect URI Mismatch
**What goes wrong:** The OAuth 2.0 flow fails with an "Invalid Redirect URI" error during the Authorization Code flow.
**Why it happens:** Payoneer requires the `redirect_uri` to be **exactly** matched against what is pre-configured in the Developer Portal. It does not allow dynamic query parameters or subdomains.
**Prevention:**
- Hardcode the redirect URI in the SDK or configuration to match the portal exactly.
- If dynamic state is needed, use the OAuth `state` parameter instead of appending query strings to the URI.

### Pitfall 3: Timezone and Format Misalignment
**What goes wrong:** Transactions are rejected or reported with incorrect dates.
**Why it happens:** Payoneer expects timestamps in **UTC** (ISO 8601) or sometimes Unix epoch in **milliseconds**. Local machine time (non-UTC) is a common source of errors.
**Prevention:**
- Always use `time.Now().UTC()` when generating timestamps.
- Use `time.RFC3339` for ISO 8601 formatting, but check if Payoneer requires the specific `+0000` offset format rather than `Z`.
**Detection:** "Invalid date format" error responses or offsets in reporting dashboards.

---

## Minor Pitfalls

### Pitfall 1: Missing Correlation-ID for Support
**What goes wrong:** When an API call fails and you contact Payoneer support, they cannot find the request in their logs.
**Why it happens:** Payoneer provides a `Correlation-ID` (or similar) in the response headers of every request. If this isn't logged, support has no way to trace the specific internal failure.
**Prevention:** Always log the `Correlation-ID` from the response headers whenever an error (even a 2xx business error) occurs.

### Pitfall 2: Sandbox vs. Production Compliance Divergence
**What goes wrong:** A flow that works perfectly in Sandbox is blocked in Production.
**Why it happens:** The Sandbox environment does not enforce the same "Risk/Compliance" holds or manual document verification steps as Production.
**Prevention:** Build the system to handle "Pending" or "Under Review" statuses as standard states, not just "Success" or "Failure".

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| **Auth Implementation** | Token refresh race conditions | Use `golang.org/x/oauth2` or `sync.Mutex` for token state. |
| **Payout Flow** | Double Payouts on retry | Implement mandatory `Client Reference ID` for all payout methods. |
| **Error Handling** | Missing business failures | Implement a `ResponseValidator` that checks the JSON body for `interaction` or `status` fields regardless of HTTP code. |
| **Webhook Integration** | Replay attacks / Spoofing | Mandatory HMAC SHA-256 signature verification for all ICPN callbacks. |

## Sources

- [Payoneer Checkout Documentation - Interaction Codes](https://developer.payoneer.com/docs/checkout/interaction-codes)
- [Payoneer Mass Payouts API Reference](https://developer.payoneer.com/docs/mass-payouts/api-reference)
- [Go SDK Design Patterns (Community Best Practices)](https://github.com/golang/go/wiki/CodeReviewComments)
- [Developer Community Discussions (Postman Collections & GitHub Issues)](https://www.postman.com/payoneer-dev/workspace/payoneer-public-workspace)
