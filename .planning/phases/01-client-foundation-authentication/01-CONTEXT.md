# Phase 1: Client Foundation & Authentication - Context

**Gathered:** 2026-03-15
**Status:** Ready for planning

<domain>
## Phase Boundary

Establishing the core SDK structure and secure OAuth 2.0 integration. This includes the base Client, functional options for configuration, OAuth 2.0 flows (Client Credentials & Authorization Code), custom error handling for Payoneer's response patterns, and structured logging with redaction.

</domain>

<decisions>
## Implementation Decisions

### Client Interface Design
- **Service Access**: Resource services (e.g., Payouts, Accounts) will be pre-initialized on the main Client struct to improve discoverability via autocomplete (e.g., `client.Payouts.List(...)`).
- **Client Type**: Use a concrete struct for the `Client`. This is standard Go practice for SDKs, allowing users to define their own interfaces for mocking if required.

### Authentication Lifecycle
- **Credential Injection**: Credentials (client_id, client_secret) will be passed via Functional Options during client initialization.
- **Token Management**: Use an interface for token storage with an in-memory implementation provided by default. This satisfies the "thread-safe mechanism" requirement and allows for future distributed storage.

### Error Handling
- **APIError Structure**: Map known Payoneer error codes to Go constants to enable easy `errors.Is` checks.
- **Business Logic Validation**: Implement response validation in the base client's transport layer to handle "False Positive" 200 OK responses where the business operation actually failed.

### Logging & Redaction
- **Default Level**: Set the default log level to `Info`.
- **Redaction**: Provide a robust default redaction list for sensitive headers (Authorization) and body fields (tokens, secrets), with an option for users to extend or override this list.

### Claude's Discretion
- Exact package naming (e.g., `payoneer`, `client`, etc.)
- Specific struct field names for internal state.
- Internal helper function signatures.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Core
- `.planning/PROJECT.md` — Project vision and core value
- `.planning/REQUIREMENTS.md` — V1 requirements checklist
- `.planning/ROADMAP.md` — Phase structure and success criteria

### Research Findings
- `.planning/research/STACK.md` — Recommended Go libraries (slog, OTel, oauth2)
- `.planning/research/ARCHITECTURE.md` — Service-oriented client design
- `.planning/research/PITFALLS.md` — Critical warnings on "False Positives" and Credential leaks

### External Documentation
- [Payoneer OAuth 2.0 Documentation](https://developer.payoneer.com/docs/mass-payouts/authentication)
- [Payoneer Error Codes Reference](https://developer.payoneer.com/docs/mass-payouts/error-codes)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- None (Greenfield project).

### Established Patterns
- **Functional Options**: Decided as the primary configuration pattern.
- **Context-First**: All public methods must accept `context.Context`.

### Integration Points
- This phase establishes the integration point for all future resource-specific services (Accounts, Payouts, etc.).

</code_context>

<specifics>
## Specific Ideas
- Use `golang.org/x/oauth2` for the heavy lifting of token exchange and refreshing.
- Ensure `log/slog` integration allows users to pass their own `*slog.Logger`.

</specifics>

<deferred>
## Deferred Ideas
- Distributed token storage (Redis/DB) — Moved to v2.
- Mock client for testing — Moved to v2.

</deferred>

---

*Phase: 01-client-foundation-authentication*
*Context gathered: 2026-03-15*
