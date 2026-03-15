# Feature Landscape

**Domain:** Cross-border Payments (Payoneer API)
**Researched:** 2024-05-24
**Overall Confidence:** HIGH

## Table Stakes

Features users expect from a payment provider API.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Account Management** | Essential for identifying accounts and status. | Low | `/v4/accounts/{account_id}` |
| **Balance Inquiries** | Need to know available funds before initiating payouts. | Low | `/v4/accounts/{account_id}/balances` |
| **Mass Payouts** | Core functionality for business-to-many payments. | High | `/v4/programs/{program_id}/payouts` |
| **Transaction History** | Required for reconciliation and reporting. | Medium | `/v4/accounts/{account_id}/transactions` |
| **Payee Onboarding** | Necessary to register new recipients. | Medium | `/v4/programs/{program_id}/payees/registration-link` |

## Differentiators

Features that set the Payoneer API apart or provide extra value.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Sandbox Environment** | Allows full testing without real funds. | Low | `api.sandbox.payoneer.com` |
| **Webhooks (IPCN)** | Real-time notifications for payment status and KYC updates. | Medium | Prevents polling and improves UX. |
| **OAuth 2.0 Auth** | Industry standard security and granular permissions. | Medium | Uses Application and Access tokens via `/v2/oauth2/token`. |
| **Payee Status Tracking** | Detailed visibility into payee registration and approval. | Low | `/v4/programs/{program_id}/payees/{payee_id}` |
| **API Versioning** | Modern `v4` for resources with stable `v2` for auth. | Low | Predictable deprecation cycles (6-18 months). |

## Anti-Features

Features to explicitly NOT build or avoid in the context of this wrapper.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Consumer P2P** | Payoneer API is primarily for B2B/B2C mass payouts. | Focus on Program/Account payouts. |
| **Crypto Support** | Usually handled via separate compliance/flow. | Use traditional fiat cross-border flows. |
| **Front-end UI** | This is a backend Go wrapper. | Provide solid data models for others to build UI. |

## Feature Dependencies

```
Auth Token (OAuth2 v2) → Account/Payment API Calls (v4)
Payee Registration → Payout Submission
Account ID → Balance/Transaction History
```

## MVP Recommendation

Prioritize:
1. **OAuth2 Authentication Flow**: Foundation for all other calls.
2. **Account Balances**: Simplest GET endpoint to verify connectivity.
3. **Mass Payout Submission**: The core value proposition of the API.
4. **Transaction History**: Essential for audit trails.

Defer:
- **Webhooks (IPCN)**: Can be added after core API parity is achieved.
- **Payee Management**: Manual onboarding via Payoneer portal can be used initially if necessary.

## Sources

- [Payoneer Developer Portal](https://developer.payoneer.com/)
- [Payoneer API v4 Reference](https://developer.payoneer.com/docs/api-reference.html)
- [Postman Collection for Payoneer v4](https://developer.payoneer.com/docs/postman-collection.html)
