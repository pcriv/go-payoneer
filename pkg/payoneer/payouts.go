package payoneer

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

var (
	SubmitMassPayoutsEndpoint = EndpointDef{
		Method:   http.MethodPost,
		Path:     "/v4/programs/{programID}/masspayouts",
		SpanName: "payoneer.payout.create_mass_payout",
	}
	GetPayoutStatusEndpoint = EndpointDef{
		Method:   http.MethodGet,
		Path:     "/v4/programs/{programID}/payouts/{clientReferenceID}/status",
		SpanName: "payoneer.payout.get_status",
	}
	CancelPayoutEndpoint = EndpointDef{
		Method:   http.MethodPut,
		Path:     "/v4/programs/{programID}/payouts/{clientReferenceID}/cancel",
		SpanName: "payoneer.payout.cancel",
	}
)

// SubmitMany submits a batch of payout requests.
// On success the API returns HTTP 201 with {"result": "Payments Created"}.
func (s *PayoutsService) SubmitMany(ctx context.Context, req *MassPayoutRequest) (*MassPayoutResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var resp MassPayoutResult
	err := s.client.execute(ctx, SubmitMassPayoutsEndpoint, req, &resp,
		WithSpanAttr(attribute.Int("payment_count", len(req.Payments))),
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetStatus retrieves the status of a specific payout.
func (s *PayoutsService) GetStatus(ctx context.Context, clientReferenceID string) (*PayoutStatusResult, error) {
	if clientReferenceID == "" {
		return nil, ErrClientReferenceIDRequired
	}

	var resp apiResult[PayoutStatusResult]
	err := s.client.execute(ctx, GetPayoutStatusEndpoint, nil, &resp,
		WithPathArg("clientReferenceID", clientReferenceID),
		WithSpanAttr(attribute.String("client_reference_id", clientReferenceID)),
	)
	if err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// Cancel cancels a pending payout.
func (s *PayoutsService) Cancel(ctx context.Context, clientReferenceID string) (*CancelResult, error) {
	if clientReferenceID == "" {
		return nil, ErrClientReferenceIDRequired
	}

	var resp apiResult[CancelResult]
	err := s.client.execute(ctx, CancelPayoutEndpoint, nil, &resp,
		WithPathArg("clientReferenceID", clientReferenceID),
		WithSpanAttr(attribute.String("client_reference_id", clientReferenceID)),
	)
	if err != nil {
		return nil, err
	}

	return &resp.Result, nil
}
