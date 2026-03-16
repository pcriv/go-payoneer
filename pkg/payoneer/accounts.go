package payoneer

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Balance represents a Payoneer account balance.
type Balance struct {
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"` // in cents
}

type balanceItem struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Type     string  `json:"type"`
}

type balanceResponse struct {
	Items []balanceItem `json:"items"`
}

// ListBalances retrieves the balances for a specific Payoneer account.
func (s *AccountsService) ListBalances(ctx context.Context, accountID string) ([]Balance, error) {
	if accountID == "" {
		return nil, ErrAccountIDRequired
	}

	path := fmt.Sprintf("/v2/accounts/%s/balances", accountID)

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.account.list_balances",
			trace.WithAttributes(attribute.String("account_id", accountID)))
		defer span.End()
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var apiResp balanceResponse
	err = s.client.Do(req, &apiResp)
	if err != nil {
		return nil, err
	}

	balances := make([]Balance, len(apiResp.Items))
	for i, item := range apiResp.Items {
		balances[i] = Balance{
			Currency: item.Currency,
			Amount:   int64(math.Round(item.Balance * 100)),
		}
	}

	return balances, nil
}
