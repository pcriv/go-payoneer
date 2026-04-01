package payoneer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CreateMassPayout submits a batch of payout requests.
// On success the API returns HTTP 201 with {"result": "Payments Created"}.
func (s *PayoutsService) CreateMassPayout(ctx context.Context, req *MassPayoutRequest) (*MassPayoutResult, error) {
	if s.client.ProgramID == "" {
		return nil, ErrProgramIDRequired
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v4/programs/%s/masspayouts", s.client.ProgramID)

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.payout.create_mass_payout",
			trace.WithAttributes(attribute.Int("payment_count", len(req.Payments))))
		defer span.End()
	}

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}

	var resp MassPayoutResult
	err = s.client.Do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetPayoutStatus retrieves the status of a specific payout.
func (s *PayoutsService) GetPayoutStatus(ctx context.Context, clientReferenceID string) (*PayoutStatusResult, error) {
	if s.client.ProgramID == "" {
		return nil, ErrProgramIDRequired
	}
	if clientReferenceID == "" {
		return nil, ErrClientReferenceIDRequired
	}

	path := fmt.Sprintf("/v4/programs/%s/payouts/%s/status", s.client.ProgramID, url.PathEscape(clientReferenceID))

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.payout.get_status",
			trace.WithAttributes(attribute.String("client_reference_id", clientReferenceID)))
		defer span.End()
	}

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp apiResult[PayoutStatusResult]
	err = s.client.Do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// CancelPayout cancels a pending payout.
func (s *PayoutsService) CancelPayout(ctx context.Context, clientReferenceID string) (*CancelPayoutResult, error) {
	if s.client.ProgramID == "" {
		return nil, ErrProgramIDRequired
	}
	if clientReferenceID == "" {
		return nil, ErrClientReferenceIDRequired
	}

	path := fmt.Sprintf("/v4/programs/%s/payouts/%s/cancel", s.client.ProgramID, url.PathEscape(clientReferenceID))

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.payout.cancel",
			trace.WithAttributes(attribute.String("client_reference_id", clientReferenceID)))
		defer span.End()
	}

	httpReq, err := s.client.NewRequest(ctx, http.MethodPut, path, nil)
	if err != nil {
		return nil, err
	}

	var resp apiResult[CancelPayoutResult]
	err = s.client.Do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Result, nil
}
