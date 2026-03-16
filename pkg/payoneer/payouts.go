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
func (s *PayoutsService) CreateMassPayout(ctx context.Context, req *MassPayoutRequest) (*PayoutBatchResult, error) {
	if s.client.ProgramID == "" {
		return nil, fmt.Errorf("program_id is required")
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

	var result PayoutBatchResult
	err = s.client.Do(httpReq, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPayoutStatus retrieves the status of a specific payout.
func (s *PayoutsService) GetPayoutStatus(ctx context.Context, clientReferenceID string) (*PayoutStatusResult, error) {
	if s.client.ProgramID == "" {
		return nil, fmt.Errorf("program_id is required")
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

	var result PayoutStatusResult
	err = s.client.Do(httpReq, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CancelPayout cancels a pending payout.
func (s *PayoutsService) CancelPayout(ctx context.Context, clientReferenceID string) error {
	if s.client.ProgramID == "" {
		return fmt.Errorf("program_id is required")
	}

	path := fmt.Sprintf("/v4/programs/%s/payouts/%s/cancel", s.client.ProgramID, url.PathEscape(clientReferenceID))

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.payout.cancel",
			trace.WithAttributes(attribute.String("client_reference_id", clientReferenceID)))
		defer span.End()
	}

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	err = s.client.Do(httpReq, nil)

	return err
}
