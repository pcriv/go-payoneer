package payoneer

import (
	"context"
	"errors"
	"testing"
)

func TestAccountsService_Validation(t *testing.T) {
	client := NewClient()

	t.Run("ListBalances requires accountID", func(t *testing.T) {
		_, err := client.Accounts.ListBalances(context.Background(), "")
		if !errors.Is(err, ErrAccountIDRequired) {
			t.Errorf("got %v, want ErrAccountIDRequired", err)
		}
	})

	t.Run("ListTransactions requires accountID", func(t *testing.T) {
		_, err := client.Accounts.ListTransactions(context.Background(), "")
		if !errors.Is(err, ErrAccountIDRequired) {
			t.Errorf("got %v, want ErrAccountIDRequired", err)
		}
	})

	t.Run("GetTransaction requires accountID", func(t *testing.T) {
		_, err := client.Accounts.GetTransaction(context.Background(), "", "tx1")
		if !errors.Is(err, ErrAccountIDRequired) {
			t.Errorf("got %v, want ErrAccountIDRequired", err)
		}
	})

	t.Run("GetTransaction requires transactionID", func(t *testing.T) {
		_, err := client.Accounts.GetTransaction(context.Background(), "acc1", "")
		if !errors.Is(err, ErrTransactionIDRequired) {
			t.Errorf("got %v, want ErrTransactionIDRequired", err)
		}
	})
}

func TestPayeesService_Validation(t *testing.T) {
	client := NewClient(WithProgramID("prog1"))

	t.Run("CreateRegistrationURL requires payeeID", func(t *testing.T) {
		_, err := client.Payees.CreateRegistrationURL(context.Background(), "")
		if !errors.Is(err, ErrPayeeIDRequired) {
			t.Errorf("got %v, want ErrPayeeIDRequired", err)
		}
	})

	t.Run("CreateRegistrationURL requires programID", func(t *testing.T) {
		c := NewClient()
		_, err := c.Payees.CreateRegistrationURL(context.Background(), "payee1")
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

	t.Run("Get requires payeeID", func(t *testing.T) {
		_, err := client.Payees.Get(context.Background(), "")
		if !errors.Is(err, ErrPayeeIDRequired) {
			t.Errorf("got %v, want ErrPayeeIDRequired", err)
		}
	})

	t.Run("Get requires programID", func(t *testing.T) {
		c := NewClient()
		_, err := c.Payees.Get(context.Background(), "payee1")
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
		err := client.Payouts.CancelPayout(context.Background(), "")
		if !errors.Is(err, ErrClientReferenceIDRequired) {
			t.Errorf("got %v, want ErrClientReferenceIDRequired", err)
		}
	})
}
