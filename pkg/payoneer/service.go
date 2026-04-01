package payoneer

// service is the base struct for all Payoneer API resource services.
type service struct {
	client *Client
}

// PayoutsService handles operations related to Payoneer payouts.
type PayoutsService service

// PayeesService handles operations related to Payoneer payees.
type PayeesService service
