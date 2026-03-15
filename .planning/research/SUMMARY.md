# Research Summary: go-payoneer

**Domain:** Cross-border Payment API (Go SDK)
**Researched:** 2024-05-24
**Overall confidence:** HIGH

## Executive Summary

The Payoneer API v4 is a comprehensive REST-based system for managing cross-border financial operations. It uses industry-standard OAuth 2.0 (with OpenID Connect) for secure authorization via `v2` endpoints (e.g., `/api/v2/oauth2/token`), supporting both application-level and user-level context. Key resource features include mass payout submission, real-time balance inquiries, and detailed transaction reporting, all standard on the `v4` resource endpoints.

The API documentation is robust, offering both a Sandbox environment for testing and a Production environment for live operations. While no official Go SDK exists, the RESTful nature of the API and its use of standard OAuth2 make it straightforward to implement using Go's `net/http` and `golang.org/x/oauth2` libraries.

Webhooks (IPCN) are available for asynchronous notifications, such as payment status updates and KYC verification events. Rate limiting is enforced with a 429 status code, requiring clients to implement robust retry logic. Payoneer follows a URI-based versioning strategy with a typical 6-18 month deprecation policy.

## Key Findings

**Stack:** Go (latest) using `net/http` and `golang.org/x/oauth2`.
**Architecture:** Modular client-based SDK with service-specific separation.
**Critical pitfall:** OAuth2 token management and automatic refreshing.

## Implications for Roadmap

Based on research, suggested phase structure:

1. **Phase 1: Foundation (Auth & Core)** - Establish the `Client` structure and OAuth2 v2 integration.
   - Addresses: Client credentials and authorization code flows.
   - Avoids: Token expiration pitfalls.

2. **Phase 2: Account Services (v4)** - Implement balance and transaction history endpoints.
   - Addresses: Account-level information retrieval.
   - Rationale: High value, low complexity for initial verification.

3. **Phase 3: Financial Operations (Payouts)** - Core value proposition implementation.
   - Addresses: Mass payout submission and status tracking.
   - Rationale: Most critical feature for business users.

4. **Phase 4: Payee Management & Webhooks** - Secondary features for full parity.
   - Addresses: IPCN handling and payee registration.
   - Rationale: Essential for a "complete" SDK experience.

**Phase ordering rationale:**
- Authentication is a hard dependency for all API calls.
- Account-level GET requests are easier to test and verify connectivity than complex POST payout requests.
- Webhooks require a separate listener architecture, so they should follow the core API wrapper.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Go is ideal for this; standard libraries are sufficient. |
| Features | HIGH | Official documentation is detailed and accessible. |
| Architecture | HIGH | Standard API wrapper patterns apply perfectly. |
| Pitfalls | MEDIUM | Based on search results and common API patterns; specific edge cases may exist. |

## Gaps to Address

- **Specific Rate Limit Numbers:** Payoneer does not publicly state the exact numeric limits (e.g., 500 req/min). This will need to be discovered through testing or communication with their support.
- **Webhook Payload Schemas:** Detailed IPCN payload schemas are sometimes only accessible inside the Developer Portal's secure area.
