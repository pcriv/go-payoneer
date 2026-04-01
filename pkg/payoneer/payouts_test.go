package payoneer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayoutsService_CreateMassPayout(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithProgramID("12345"),
	)

	mux.HandleFunc("/v4/programs/12345/masspayouts", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var body map[string]any
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			t.Fatal(err)
		}

		// Verify the key is capital "Payments"
		_, ok := body["Payments"]
		assert.True(t, ok, "expected capital 'Payments' key in request body")

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"result":"Payments Created"}`)
	})

	req := &MassPayoutRequest{
		Payments: []PayoutItem{
			{
				ClientReferenceID: "ref-1",
				PayeeID:           "payee-1",
				Amount:            1000,
				Currency:          "USD",
				Description:       "Test payout",
			},
		},
	}

	result, err := client.Payouts.CreateMassPayout(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Payments Created", result.Result)
}

func TestPayoutsService_GetPayoutStatus(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithProgramID("12345"),
	)

	mux.HandleFunc("/v4/programs/12345/payouts/ref-1/status", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(
			w,
			`{"result":{"payout_date":"2021-03-17T10:47:00-04:00","amount":5.10,"currency":"USD","status":"Transferred","payee_id":"payee-1","payout_id":"1636595702","load_date":"2021-03-17T14:09:39-04:00"}}`,
		)
	})

	result, err := client.Payouts.GetPayoutStatus(context.Background(), "ref-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Transferred", result.Status)

	payoutID, ok := result.PayoutID.Get()
	assert.True(t, ok)
	assert.Equal(t, "1636595702", payoutID)

	amount, ok := result.Amount.Get()
	assert.True(t, ok)
	assert.InDelta(t, 5.10, amount, 0.001)

	currency, ok := result.Currency.Get()
	assert.True(t, ok)
	assert.Equal(t, "USD", currency)
}

func TestPayoutsService_GetPayoutStatus_Cancelled(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithProgramID("12345"),
	)

	mux.HandleFunc("/v4/programs/12345/payouts/ref-cancelled/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(
			w,
			`{"result":{"payout_date":"2021-03-17T10:47:00-04:00","amount":5.10,"currency":"USD","status":"Cancelled","payee_id":"payee-1","payout_id":"123","cancel_reason_code":30013,"cancel_reason_description":"Bank details - Invalid Branch Code"}}`,
		)
	})

	result, err := client.Payouts.GetPayoutStatus(context.Background(), "ref-cancelled")
	require.NoError(t, err)
	assert.Equal(t, "Cancelled", result.Status)

	code, ok := result.CancelReasonCode.Get()
	assert.True(t, ok)
	assert.Equal(t, 30013, code)

	desc, ok := result.CancelReasonDescription.Get()
	assert.True(t, ok)
	assert.Equal(t, "Bank details - Invalid Branch Code", desc)
}

func TestPayoutsService_CancelPayout(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithProgramID("12345"),
	)

	mux.HandleFunc("/v4/programs/12345/payouts/ref-1/cancel", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"result":{"description":"The request was received successfully. The cancel action has not yet been performed."}}`)
	})

	result, err := client.Payouts.CancelPayout(context.Background(), "ref-1")
	require.NoError(t, err)
	assert.Equal(t, "The request was received successfully. The cancel action has not yet been performed.", result.Description)
}

func TestPayoutsService_ProgramIDRequired(t *testing.T) {
	client := NewClient()

	_, err := client.Payouts.CreateMassPayout(context.Background(), &MassPayoutRequest{
		Payments: []PayoutItem{{ClientReferenceID: "ref-1", PayeeID: "p1", Amount: 100, Currency: "USD", Description: "test"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "program_id is required")

	_, err = client.Payouts.GetPayoutStatus(context.Background(), "ref-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "program_id is required")

	_, err = client.Payouts.CancelPayout(context.Background(), "ref-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "program_id is required")
}

func TestPayoutErrorHandling(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithProgramID("12345"),
	)

	t.Run("404 Payout Not Found (2306)", func(t *testing.T) {
		mux.HandleFunc("/v4/programs/12345/payouts/ref-404/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error_code":"2306","description":"Payout not found","status":"Failure"}`)
		})

		_, err := client.Payouts.GetPayoutStatus(context.Background(), "ref-404")
		require.Error(t, err)

		apiErr, ok := errors.AsType[*APIError](err)
		if assert.True(t, ok, "expected APIError, got %T", err) {
			assert.Equal(t, ErrCodePayoutNotFound, apiErr.Code)
			assert.Equal(t, http.StatusNotFound, apiErr.HTTPStatusCode)
		}
	})

	t.Run("400 Validation Error", func(t *testing.T) {
		mux.HandleFunc("/v4/programs/12345/masspayouts", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error_code":"400","description":"Invalid payout data"}`)
		})

		req := &MassPayoutRequest{
			Payments: []PayoutItem{{ClientReferenceID: "ref-1", PayeeID: "p1", Amount: 100, Currency: "USD", Description: "test"}},
		}
		_, err := client.Payouts.CreateMassPayout(context.Background(), req)
		require.Error(t, err)
	})
}

func TestPayoutsService_Validation(t *testing.T) {
	client := NewClient(WithProgramID("12345"))

	t.Run("Empty payments list", func(t *testing.T) {
		req := &MassPayoutRequest{Payments: []PayoutItem{}}
		_, err := client.Payouts.CreateMassPayout(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one payment is required")
	})

	t.Run("Missing client_reference_id", func(t *testing.T) {
		req := &MassPayoutRequest{
			Payments: []PayoutItem{{PayeeID: "p1", Amount: 100, Currency: "USD", Description: "test"}},
		}
		_, err := client.Payouts.CreateMassPayout(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client_reference_id is required")
	})

	t.Run("Zero amount", func(t *testing.T) {
		req := &MassPayoutRequest{
			Payments: []PayoutItem{{ClientReferenceID: "ref-1", PayeeID: "p1", Amount: 0, Currency: "USD", Description: "test"}},
		}
		_, err := client.Payouts.CreateMassPayout(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be greater than zero")
	})

	t.Run("Missing description", func(t *testing.T) {
		req := &MassPayoutRequest{
			Payments: []PayoutItem{{ClientReferenceID: "ref-1", PayeeID: "p1", Amount: 100, Currency: "USD"}},
		}
		_, err := client.Payouts.CreateMassPayout(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "description is required")
	})
}
