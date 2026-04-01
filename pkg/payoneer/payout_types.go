package payoneer

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

// PayoutItem represents a single payment in a mass payout request.
type PayoutItem struct {
	ClientReferenceID string `json:"client_reference_id"`
	PayeeID           string `json:"payee_id"`
	Amount            int64  `json:"-"` // Internal use in cents
	Currency          string `json:"currency,omitempty"`
	Description       string `json:"description"`
	PayoutDate        string `json:"payout_date,omitempty"`
	GroupID           string `json:"group_id,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface for PayoutItem.
func (p PayoutItem) MarshalJSON() ([]byte, error) {
	type Alias PayoutItem

	return json.Marshal(&struct {
		Amount float64 `json:"amount"`
		Alias
	}{
		Amount: float64(p.Amount) / 100.0,
		Alias:  Alias(p),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for PayoutItem.
func (p *PayoutItem) UnmarshalJSON(data []byte) error {
	type Alias PayoutItem
	aux := &struct {
		Amount float64 `json:"amount"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Amount = int64(math.Round(aux.Amount * 100))

	return nil
}

// MassPayoutRequest contains a batch of payments to be processed.
type MassPayoutRequest struct {
	Payments []PayoutItem `json:"Payments"`
}

// Validate ensures that the mass payout request is structurally valid.
func (r *MassPayoutRequest) Validate() error {
	if len(r.Payments) == 0 {
		return errors.New("at least one payment is required")
	}
	if len(r.Payments) > 500 {
		return errors.New("maximum 500 payments allowed per batch")
	}
	for i, p := range r.Payments {
		if p.ClientReferenceID == "" {
			return fmt.Errorf("payment %d: client_reference_id is required", i)
		}
		if p.Amount <= 0 {
			return fmt.Errorf("payment %d: amount must be greater than zero", i)
		}
		if p.Description == "" {
			return fmt.Errorf("payment %d: description is required", i)
		}
	}

	return nil
}

// MassPayoutResult is the response from submitting a mass payout.
// The API returns the result as a plain string (e.g. "Payments Created").
type MassPayoutResult struct {
	Result string `json:"result"`
}

// PayoutStatusResult represents the status of a single payout.
type PayoutStatusResult struct {
	PayoutDate              Optional[string]  `json:"payout_date"`
	Amount                  Optional[float64] `json:"amount"`
	Currency                Optional[string]  `json:"currency"`
	Status                  string            `json:"status"`
	TargetAmount            Optional[float64] `json:"target_amount"`
	TargetCurrency          Optional[string]  `json:"target_currency"`
	PayeeID                 Optional[string]  `json:"payee_id"`
	PayoutID                Optional[string]  `json:"payout_id"`
	ScheduledPayoutDate     Optional[string]  `json:"scheduled_payout_date"`
	LoadDate                Optional[string]  `json:"load_date"`
	ReasonCode              Optional[string]  `json:"reason_code"`
	ReasonDescription       Optional[string]  `json:"reason_description"`
	CancelReasonCode        Optional[int]     `json:"cancel_reason_code"`
	CancelReasonDescription Optional[string]  `json:"cancel_reason_description"`
}

// CancelResult represents the response from cancelling a payout.
type CancelResult struct {
	Description string `json:"description"`
}
