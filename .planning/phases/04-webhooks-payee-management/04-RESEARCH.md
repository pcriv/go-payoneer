# Research: Phase 4 - Webhooks & Payee Management

## Webhooks (IPCN)

### Signature Verification
Payoneer uses HMAC SHA-256 to sign IPCN (Instant Payment Confirmation Notification) webhooks.

- **Algorithm**: HMAC SHA-256
- **Secret**: Merchant Secret (configured in Payoneer Portal).
- **Header**: `X-Payoneer-Signature` (Hex-encoded).
- **Payload**: The signature is calculated over the **raw request body**.

**Go Implementation Strategy:**
```go
func ValidateSignature(payload []byte, signature string, secret string) bool {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(payload)
    expected := hex.EncodeToString(h.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Webhook Event Types
Common events to support:
- `PAYOUT_COMPLETED`
- `PAYOUT_FAILED`
- `PAYEE_STATUS_CHANGED`
- `PAYEE_KYC_REQUIRED`

### Webhook Dispatcher Design
A `WebhookHandler` will be provided to parse and validate requests.
```go
type WebhookEvent struct {
    EventType string          `json:"event_type"`
    EventID   string          `json:"event_id"`
    Timestamp string          `json:"timestamp"`
    Content   json.RawMessage `json:"content"`
}

func (c *Client) ParseWebhook(r *http.Request, secret string) (*WebhookEvent, error) {
    // 1. Read raw body
    // 2. Validate signature
    // 3. Unmarshal to WebhookEvent
}
```

## Payee Management

### Registration Links
Endpoint: `POST /v4/programs/{program_id}/payees/registration-link`

**Request Payload:**
```json
{
  "payee_id": "internal_id",
  "already_have_an_account": false,
  "redirect_url": "https://myapp.com/callback",
  "payee": {
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com"
  }
}
```

### Payee Status
Endpoint: `GET /v4/programs/{program_id}/payees/{payee_id}/status`

**Statuses:**
- `ACTIVE` (Code 1)
- `PENDING`
- `DECLINED`
- `INCOMPLETE`

## Proposed Service Design

### PayeesService
```go
type PayeesService service

func (s *PayeesService) RegistrationURL(ctx context.Context, payeeID string, opts ...RegistrationOption) (string, error)
func (s *PayeesService) GetStatus(ctx context.Context, payeeID string) (*PayeeStatus, error)
func (s *PayeesService) Get(ctx context.Context, payeeID string) (*Payee, error)
```

### Webhook Package
To avoid cluttering the main client, webhook logic might reside in `pkg/payoneer/webhooks.go` or a sub-package if it requires many types.

## Success Criteria Checklist
- [ ] `ValidateSignature` helper implemented and tested.
- [ ] `ParseWebhook` handles payload extraction and validation.
- [ ] `RegistrationURL` generates valid onboarding links.
- [ ] `GetStatus` retrieves current payee standing.
- [ ] All new types use `Optional[T]` where appropriate.
