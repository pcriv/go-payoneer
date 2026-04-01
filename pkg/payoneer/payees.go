package payoneer

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PayeeStatus represents the registration status of a payee.
type PayeeStatus struct {
	AccountID        string             `json:"account_id"`
	Status           PayeeStatusDetail  `json:"status"`
	RegistrationDate string             `json:"registration_date"`
	PayoutMethod     *PayeePayoutMethod `json:"payout_method,omitempty"`
}

// PayeeStatusDetail holds the status type code and description.
type PayeeStatusDetail struct {
	Type        int    `json:"type"`
	Description string `json:"description"`
}

// PayeePayoutMethod holds the payee's configured payout method.
type PayeePayoutMethod struct {
	Type     string `json:"type"`
	Currency string `json:"currency"`
}

// RegistrationLinkRequest is the payload for generating an onboarding link.
type RegistrationLinkRequest struct {
	PayeeID               string                    `json:"payee_id"`
	ClientSessionID       string                    `json:"client_session_id,omitempty"`
	RedirectURL           string                    `json:"redirect_url,omitempty"`
	RedirectTime          *int                      `json:"redirect_time,omitempty"`
	PayoutMethods         []string                  `json:"payout_methods,omitempty"`
	LockType              string                    `json:"lock_type,omitempty"`
	PayeeDataMatchingType string                    `json:"payee_data_matching_type,omitempty"`
	AlreadyHaveAnAccount  string                    `json:"already_have_an_account,omitempty"`
	Payee                 *RegistrationPayee        `json:"payee,omitempty"`
	LanguageID            string                    `json:"language_id,omitempty"`
	PayoutMethod          *RegistrationPayoutMethod `json:"payout_method,omitempty"`
}

// RegistrationPayee describes the payee in a registration link request.
type RegistrationPayee struct {
	Type       string                  `json:"type,omitempty"`
	Company    *RegistrationCompany    `json:"company,omitempty"`
	Contact    *RegistrationContact    `json:"contact,omitempty"`
	Address    *RegistrationAddress    `json:"address,omitempty"`
	IDDocument *RegistrationIDDocument `json:"id_document,omitempty"`
}

// RegistrationContact holds the payee's contact information.
type RegistrationContact struct {
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Mobile      string `json:"mobile,omitempty"`
}

// RegistrationCompany holds company details for a company-type payee.
type RegistrationCompany struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// RegistrationAddress holds the payee's address.
type RegistrationAddress struct {
	AddressLine1 string `json:"address_line_1,omitempty"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	ZipCode      string `json:"zip_code,omitempty"`
	Country      string `json:"country,omitempty"`
}

// RegistrationIDDocument holds identity document details.
type RegistrationIDDocument struct {
	Type                     string `json:"type,omitempty"`
	Number                   string `json:"number,omitempty"`
	IssueCountry             string `json:"issue_country,omitempty"`
	NameOnID                 string `json:"name_on_id,omitempty"`
	ExpirationDate           string `json:"expiration_date,omitempty"`
	IssueDate                string `json:"IssueDate,omitempty"`
	FirstNameInLocalLanguage string `json:"first_name_in_local_language,omitempty"`
	LastNameInLocalLanguage  string `json:"last_name_in_local_language,omitempty"`
}

// RegistrationPayoutMethod holds payout method details for bank pre-population.
type RegistrationPayoutMethod struct {
	Type             string                        `json:"type,omitempty"`
	BankAccountType  string                        `json:"bank_account_type,omitempty"`
	Country          string                        `json:"country,omitempty"`
	Currency         string                        `json:"currency,omitempty"`
	BankFieldDetails []RegistrationBankFieldDetail `json:"bank_field_details,omitempty"`
}

// RegistrationBankFieldDetail holds a single bank field key-value pair.
type RegistrationBankFieldDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RegistrationLinkResult contains the registration link and token.
type RegistrationLinkResult struct {
	RegistrationLink string `json:"registration_link"`
	Token            string `json:"token"`
}

// RegistrationOption defines functional options for registration links.
type RegistrationOption func(*RegistrationLinkRequest)

// WithRedirectURL sets the URL the payee is redirected to after registration.
func WithRedirectURL(url string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.RedirectURL = url
	}
}

// WithRedirectTime sets the interval in seconds before redirecting after registration.
func WithRedirectTime(seconds int) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.RedirectTime = &seconds
	}
}

// WithClientSessionID sets the session identifier referenced in webhooks.
func WithClientSessionID(id string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.ClientSessionID = id
	}
}

// WithPayoutMethods sets the allowed payout methods (e.g. "BankTransfer", "PrepaidCard").
func WithPayoutMethods(methods ...string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.PayoutMethods = methods
	}
}

// WithLockType sets the lock type for pre-populated registration fields.
func WithLockType(lockType string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.LockType = lockType
	}
}

// WithPayeeDataMatchingType sets the entity matching type for the payee.
func WithPayeeDataMatchingType(matchType string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.PayeeDataMatchingType = matchType
	}
}

// WithAlreadyHaveAnAccount sets whether the payee is prompted to link an existing account.
func WithAlreadyHaveAnAccount(v bool) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		if v {
			r.AlreadyHaveAnAccount = "true"
		} else {
			r.AlreadyHaveAnAccount = "false"
		}
	}
}

// WithLanguage sets the language for the registration page.
func WithLanguage(lang string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.LanguageID = lang
	}
}

// WithPayee sets the full payee object for the registration request.
func WithPayee(payee *RegistrationPayee) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.Payee = payee
	}
}

// WithPayeeContact pre-populates the registration form with payee contact info.
func WithPayeeContact(firstName, lastName, email string) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		if r.Payee == nil {
			r.Payee = &RegistrationPayee{}
		}
		r.Payee.Contact = &RegistrationContact{
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
		}
	}
}

// WithPayoutMethod sets the payout method details for bank pre-population.
func WithPayoutMethod(method *RegistrationPayoutMethod) RegistrationOption {
	return func(r *RegistrationLinkRequest) {
		r.PayoutMethod = method
	}
}

// CreateRegistrationLink generates a unique onboarding link for a payee.
func (s *PayeesService) CreateRegistrationLink(ctx context.Context, payeeID string, opts ...RegistrationOption) (*RegistrationLinkResult, error) {
	if s.client.ProgramID == "" {
		return nil, ErrProgramIDRequired
	}
	if payeeID == "" {
		return nil, ErrPayeeIDRequired
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
		ctx, span = s.client.tracer.Start(ctx, "payoneer.payee.create_registration_link",
			trace.WithAttributes(attribute.String("payee_id", payeeID)))
		defer span.End()
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		return nil, err
	}

	var resp apiResult[RegistrationLinkResult]
	if err = s.client.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
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

	var resp apiResult[PayeeStatus]
	if err = s.client.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}
