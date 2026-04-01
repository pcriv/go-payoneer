package payoneer

import (
	"context"
	"errors"
	"testing"
)

func TestPayeesService_Validation(t *testing.T) {
	client := NewClient(WithProgramID("prog1"))

	t.Run("CreateRegistrationLink requires payeeID", func(t *testing.T) {
		_, err := client.Payees.CreateRegistrationLink(context.Background(), "")
		if !errors.Is(err, ErrPayeeIDRequired) {
			t.Errorf("got %v, want ErrPayeeIDRequired", err)
		}
	})

	t.Run("CreateRegistrationLink requires programID", func(t *testing.T) {
		c := NewClient()
		_, err := c.Payees.CreateRegistrationLink(context.Background(), "payee1")
		if !errors.Is(err, ErrProgramIDRequired) {
			t.Errorf("got %v, want ErrProgramIDRequired", err)
		}
	})

	t.Run("GetStatus requires payeeID", func(t *testing.T) {
		_, err := client.Payees.GetStatus(context.Background(), "")
		if !errors.Is(err, ErrPayeeIDRequired) {
			t.Errorf("got %v, want ErrPayeeIDRequired", err)
		}
	})

	t.Run("GetStatus requires programID", func(t *testing.T) {
		c := NewClient()
		_, err := c.Payees.GetStatus(context.Background(), "payee1")
		if !errors.Is(err, ErrProgramIDRequired) {
			t.Errorf("got %v, want ErrProgramIDRequired", err)
		}
	})

}

func TestPayoutsService_ClientReferenceIDValidation(t *testing.T) {
	client := NewClient(WithProgramID("prog1"))

	t.Run("GetPayoutStatus requires clientReferenceID", func(t *testing.T) {
		_, err := client.Payouts.GetPayoutStatus(context.Background(), "")
		if !errors.Is(err, ErrClientReferenceIDRequired) {
			t.Errorf("got %v, want ErrClientReferenceIDRequired", err)
		}
	})

	t.Run("CancelPayout requires clientReferenceID", func(t *testing.T) {
		_, err := client.Payouts.CancelPayout(context.Background(), "")
		if !errors.Is(err, ErrClientReferenceIDRequired) {
			t.Errorf("got %v, want ErrClientReferenceIDRequired", err)
		}
	})
}
