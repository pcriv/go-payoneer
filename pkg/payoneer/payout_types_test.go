package payoneer_test

import (
	"encoding/json"
	"testing"

	"github.com/pcriv/go-payoneer/pkg/payoneer"
)

func TestPayoutModels_Serialization(t *testing.T) {
	t.Run("MassPayoutRequest serialization", func(t *testing.T) {
		req := payoneer.MassPayoutRequest{
			Payments: []payoneer.PayoutItem{
				{
					ClientReferenceID: "ref123",
					Amount:            1050, // 10.50
					PayeeID:           "payee456",
					Currency:          "USD",
					Description:       "Test Payment",
				},
			},
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("failed to marshal MassPayoutRequest: %v", err)
		}

		var got, want any
		if uerr := json.Unmarshal(data, &got); uerr != nil {
			t.Fatalf("failed to unmarshal actual JSON: %v", uerr)
		}

		expected := `{"Payments":[{"client_reference_id":"ref123","amount":10.5,"payee_id":"payee456","currency":"USD","description":"Test Payment"}]}`
		if uerr := json.Unmarshal([]byte(expected), &want); uerr != nil {
			t.Fatalf("failed to unmarshal expected JSON: %v", uerr)
		}

		gotStr, _ := json.Marshal(got)
		wantStr, _ := json.Marshal(want)

		if string(gotStr) != string(wantStr) {
			t.Errorf("expected JSON %s, got %s", string(wantStr), string(gotStr))
		}
	})

	t.Run("PayoutStatusResult deserialization", func(t *testing.T) {
		jsonData := `{
			"payout_date": "2021-03-17T10:47:00-04:00",
			"amount": 5.10,
			"currency": "USD",
			"status": "Pending",
			"payee_id": "payee123",
			"payout_id": "1636595702",
			"reason_code": "5001",
			"reason_description": "Processing"
		}`

		var result payoneer.PayoutStatusResult
		if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
			t.Fatalf("failed to unmarshal PayoutStatusResult: %v", err)
		}

		if result.Status != "Pending" {
			t.Errorf("expected Status Pending, got %s", result.Status)
		}

		reasonDesc, ok := result.ReasonDescription.Get()
		if !ok || reasonDesc != "Processing" {
			t.Errorf("expected ReasonDescription Processing, got %s (ok: %v)", reasonDesc, ok)
		}
	})
}

func TestMassPayoutRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     payoneer.MassPayoutRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: payoneer.MassPayoutRequest{
				Payments: []payoneer.PayoutItem{
					{
						ClientReferenceID: "ref1",
						Amount:            100,
						PayeeID:           "payee1",
						Currency:          "USD",
						Description:       "test",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty payments slice",
			req: payoneer.MassPayoutRequest{
				Payments: []payoneer.PayoutItem{},
			},
			wantErr: true,
		},
		{
			name: "too many payments",
			req: payoneer.MassPayoutRequest{
				Payments: make([]payoneer.PayoutItem, 501),
			},
			wantErr: true,
		},
		{
			name: "missing client reference id",
			req: payoneer.MassPayoutRequest{
				Payments: []payoneer.PayoutItem{
					{
						ClientReferenceID: "",
						Amount:            100,
						PayeeID:           "payee1",
						Currency:          "USD",
						Description:       "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "zero amount",
			req: payoneer.MassPayoutRequest{
				Payments: []payoneer.PayoutItem{
					{
						ClientReferenceID: "ref1",
						Amount:            0,
						PayeeID:           "payee1",
						Currency:          "USD",
						Description:       "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			req: payoneer.MassPayoutRequest{
				Payments: []payoneer.PayoutItem{
					{
						ClientReferenceID: "ref1",
						Amount:            -10,
						PayeeID:           "payee1",
						Currency:          "USD",
						Description:       "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing description",
			req: payoneer.MassPayoutRequest{
				Payments: []payoneer.PayoutItem{
					{
						ClientReferenceID: "ref1",
						Amount:            100,
						PayeeID:           "payee1",
						Currency:          "USD",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("MassPayoutRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
