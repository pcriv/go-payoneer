package payoneer

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Payee represents a Payoneer payee's registration details.
type Payee struct {
	PayeeID      string           `json:"payee_id"`
	FirstName    Optional[string] `json:"first_name"`
	LastName     Optional[string] `json:"last_name"`
	Email        Optional[string] `json:"email"`
	Status       Optional[string] `json:"status"`
	AuditID      Optional[int64]  `json:"audit_id"`
	RegisterTime Optional[string] `json:"register_time"`
}

// PayeeStatus represents the current standing of a payee.
type PayeeStatus struct {
	PayeeID           string `json:"payee_id"`
	StatusCode        int    `json:"status_code"`
	StatusDescription string `json:"status_description"`
}

// RegistrationLinkRequest is the payload for generating an onboarding link.
type RegistrationLinkRequest struct {
	PayeeID              string           `json:"payee_id"`
	AlreadyHaveAnAccount bool             `json:"already_have_an_account"`
	RedirectURL          string           `json:"redirect_url,omitempty"`
	Language             string           `json:"language,omitempty"`
	FirstName            Optional[string] `json:"first_name"`
	LastName             Optional[string] `json:"last_name"`
	Email                Optional[string] `json:"email"`
}

// RegistrationLinkResponse wraps the generated onboarding URL.
type RegistrationLinkResponse struct {
	RegistrationLink string `json:"registration_link"`
}

// RegistrationOption defines functional options for registration links.
type RegistrationOption func(*RegistrationLinkRequest)

// WithRedirectURL sets the URL the payee is redirected to after registration.
func WithRedirectURL(url string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.RedirectURL = url
	}
}

// WithAlreadyHaveAnAccount sets whether the payee is prompted to link an existing account.
func WithAlreadyHaveAnAccount(v bool) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.AlreadyHaveAnAccount = v
	}
}

// WithLanguage sets the language for the registration page.
func WithLanguage(lang string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.Language = lang
	}
}

// WithPayeeDetails pre-populates the registration form with payee info.
func WithPayeeDetails(firstName, lastName, email string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.FirstName = Some(firstName)
		r.LastName = Some(lastName)
		r.Email = Some(email)
	}
}

// CreateRegistrationURL generates a unique onboarding link for a payee.
func (s *PayeesService) CreateRegistrationURL(ctx context.Context, payeeID string, opts ...RegistrationOption) (string, error) {
	if s.client.ProgramID == "" {
		return "", ErrProgramIDRequired
	}
	if payeeID == "" {
		return "", ErrPayeeIDRequired
	}

	reqBody := &RegistrationLinkRequest{
		PayeeID: payeeID,
	}

	for _, opt := range opts {
		opt(reqBody)
	}

	path := fmt.Sprintf("/v4/programs/%s/payees/registration-link", s.client.ProgramID)

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.payee.create_registration_url",
			trace.WithAttributes(attribute.String("payee_id", payeeID)))
		defer span.End()
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		return "", err
	}

	var resp RegistrationLinkResponse
	if err = s.client.Do(req, &resp); err != nil {
		return "", err
	}

	return resp.RegistrationLink, nil
}

// GetStatus retrieves the current standing of a payee.
func (s *PayeesService) GetStatus(ctx context.Context, payeeID string) (*PayeeStatus, error) {
	if s.client.ProgramID == "" {
		return nil, ErrProgramIDRequired
	}
	if payeeID == "" {
		return nil, ErrPayeeIDRequired
	}

	path := fmt.Sprintf("/v4/programs/%s/payees/%s/status", s.client.ProgramID, payeeID)

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.payee.get_status",
			trace.WithAttributes(attribute.String("payee_id", payeeID)))
		defer span.End()
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var status PayeeStatus
	if err = s.client.Do(req, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// Get retrieves full details for a specific payee.
func (s *PayeesService) Get(ctx context.Context, payeeID string) (*Payee, error) {
	if s.client.ProgramID == "" {
		return nil, ErrProgramIDRequired
	}
	if payeeID == "" {
		return nil, ErrPayeeIDRequired
	}

	path := fmt.Sprintf("/v4/programs/%s/payees/%s", s.client.ProgramID, payeeID)

	if s.client.tracer != nil {
		var span trace.Span
		ctx, span = s.client.tracer.Start(ctx, "payoneer.payee.get",
			trace.WithAttributes(attribute.String("payee_id", payeeID)))
		defer span.End()
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var payee Payee
	if err = s.client.Do(req, &payee); err != nil {
		return nil, err
	}

	return &payee, nil
}
