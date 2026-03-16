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

	payoneererrors "github.com/pcriv/go-payoneer/pkg/payoneer/errors"
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

		var req MassPayoutRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, req.Payments, 1)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"batch_id":"batch-123","status":"SUBMITTED"}`)
	})

	req := &MassPayoutRequest{
		Payments: []PayoutItem{
			{
				ClientReferenceID: "ref-1",
				PayeeID:           "payee-1",
				Amount:            1000,
				Currency:          "USD",
			},
		},
	}

	result, err := client.Payouts.CreateMassPayout(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "batch-123", result.BatchID)
	assert.Equal(t, "SUBMITTED", result.Status)
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
		fmt.Fprint(w, `{"payout_id":"payout-123","client_reference_id":"ref-1","status":"COMPLETED"}`)
	})

	result, err := client.Payouts.GetPayoutStatus(context.Background(), "ref-1")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "payout-123", result.PayoutID)
	assert.Equal(t, "COMPLETED", result.Status)
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
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.Payouts.CancelPayout(context.Background(), "ref-1")
	require.NoError(t, err)
}

func TestPayoutsService_ProgramIDRequired(t *testing.T) {
	client := NewClient()

	_, err := client.Payouts.CreateMassPayout(context.Background(), &MassPayoutRequest{
		Payments: []PayoutItem{{ClientReferenceID: "ref-1", PayeeID: "p1", Amount: 100, Currency: "USD"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "program_id is required")

	_, err = client.Payouts.GetPayoutStatus(context.Background(), "ref-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "program_id is required")

	err = client.Payouts.CancelPayout(context.Background(), "ref-1")
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

		var apiErr *payoneererrors.APIError
		if errors.As(err, &apiErr) {
			assert.Equal(t, payoneererrors.ErrCodePayoutNotFound, apiErr.Code)
			assert.Equal(t, http.StatusNotFound, apiErr.HTTPStatusCode)
		} else {
			t.Errorf("expected APIError, got %T", err)
		}
	})

	t.Run("400 Validation Error", func(t *testing.T) {
		mux.HandleFunc("/v4/programs/12345/masspayouts", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error_code":"400","description":"Invalid payout data"}`)
		})

		req := &MassPayoutRequest{
			Payments: []PayoutItem{{ClientReferenceID: "ref-1", PayeeID: "p1", Amount: 100, Currency: "USD"}},
		}
		_, err := client.Payouts.CreateMassPayout(context.Background(), req)
		require.Error(t, err)
	})

	t.Run("200 OK with business failure", func(t *testing.T) {
		// If status is "Failure", it should be an error according to our ValidateResponse.
		// If status is "REJECTED", it should be returned in result.
		mux.HandleFunc("/v4/programs/12345/payouts/ref-rejected/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"payout_id":"payout-123","client_reference_id":"ref-rejected","status":"REJECTED","reason":"Insufficient funds"}`)
		})

		result, err := client.Payouts.GetPayoutStatus(context.Background(), "ref-rejected")
		require.NoError(t, err)
		assert.Equal(t, "REJECTED", result.Status)
		assert.Equal(t, "Insufficient funds", result.Reason.OrDefault(""))
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
			Payments: []PayoutItem{{PayeeID: "p1", Amount: 100, Currency: "USD"}},
		}
		_, err := client.Payouts.CreateMassPayout(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client_reference_id is required")
	})

	t.Run("Zero amount", func(t *testing.T) {
		req := &MassPayoutRequest{
			Payments: []PayoutItem{{ClientReferenceID: "ref-1", PayeeID: "p1", Amount: 0, Currency: "USD"}},
		}
		_, err := client.Payouts.CreateMassPayout(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be greater than zero")
	})
}
