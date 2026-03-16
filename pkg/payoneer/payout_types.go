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
	Currency          string `json:"currency"`
	Description       string `json:"description,omitempty"`
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
	Payments []PayoutItem `json:"payments"`
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
	}

	return nil
}

// PayoutBatchResult represents the API response for mass payouts.
type PayoutBatchResult struct {
	BatchID     string           `json:"batch_id"`
	Status      string           `json:"status"`
	Reason      Optional[string] `json:"reason"`
	ReleaseDate Optional[string] `json:"release_date"`
}

// PayoutStatusResult represents the status of a single payout.
type PayoutStatusResult struct {
	PayoutID          string           `json:"payout_id"`
	ClientReferenceID string           `json:"client_reference_id"`
	Status            string           `json:"status"`
	Reason            Optional[string] `json:"reason"`
	ReleaseDate       Optional[string] `json:"release_date"`
}
