package payoneer

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Transaction represents a Payoneer transaction.
type Transaction struct {
	ID          string    `json:"id"`
	Amount      int64     `json:"amount"` // in cents
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	Direction   string    `json:"direction"`
}

type transactionItem struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
	CreatedAt   string  `json:"created_at"`
	Direction   string  `json:"direction"`
}

type transactionListResponse struct {
	Items []transactionItem `json:"items"`
}

// TransactionListOption is a functional option for listing transactions.
type TransactionListOption func(*url.Values)

// ListTransactions retrieves transaction history for a specific Payoneer account.
func (s *AccountsService) ListTransactions(ctx context.Context, accountID string, opts ...TransactionListOption) ([]Transaction, error) {
	path := fmt.Sprintf("/v2/accounts/%s/transactions", accountID)

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.account.list_transactions",
			trace.WithAttributes(attribute.String("account_id", accountID)))
		defer span.End()
	}

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	// Apply options to query parameters
	q := req.URL.Query()
	for _, opt := range opts {
		opt(&q)
	}
	req.URL.RawQuery = q.Encode()

	var apiResp transactionListResponse
	err = s.client.Do(req, &apiResp)
	if err != nil {
		return nil, err
	}

	transactions := make([]Transaction, len(apiResp.Items))
	for i, item := range apiResp.Items {
		tx, terr := mapTransaction(item)
		if terr != nil {
			return nil, fmt.Errorf("failed to map transaction %s: %w", item.ID, terr)
		}
		transactions[i] = tx
	}

	return transactions, nil
}

// GetTransaction retrieves a specific transaction by ID.
func (s *AccountsService) GetTransaction(ctx context.Context, accountID string, transactionID string) (*Transaction, error) {
	path := fmt.Sprintf("/v2/accounts/%s/transactions/%s", accountID, transactionID)

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.account.get_transaction",
			trace.WithAttributes(
				attribute.String("account_id", accountID),
				attribute.String("transaction_id", transactionID),
			))
		defer span.End()
	}

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var item transactionItem
	err = s.client.Do(req, &item)
	if err != nil {
		return nil, err
	}

	tx, err := mapTransaction(item)
	if err != nil {
		return nil, fmt.Errorf("failed to map transaction %s: %w", item.ID, err)
	}

	return &tx, nil
}

func mapTransaction(item transactionItem) (Transaction, error) {
	createdAt, err := time.Parse(time.RFC3339, item.CreatedAt)
	if err != nil {
		return Transaction{}, err
	}

	return Transaction{
		ID:          item.ID,
		Amount:      int64(math.Round(item.Amount * 100)),
		Currency:    item.Currency,
		Description: item.Description,
		Status:      item.Status,
		Type:        item.Type,
		CreatedAt:   createdAt,
		Direction:   item.Direction,
	}, nil
}
