package warp

import (
	"context"
	"crypto/rand"
	"math/big"

	"github.com/handsomefox/gowarp/pkg/models"
)

// CreateAccountError is returned if something during MakeKey() fails.
type CreateAccountError struct {
	reason error
}

func (e CreateAccountError) Error() string {
	return "error creating account: " + e.reason.Error()
}

func newCreateAccountError(reason error) CreateAccountError {
	return CreateAccountError{
		reason: reason,
	}
}

// MakeKey creates models.Account with random license.
func MakeKey(ctx context.Context, s *Service) (*models.Account, error) {
	c := s.GetRequestClient(ctx)

	acc1, err := NewAccount(ctx, c)
	if err != nil {
		return nil, newCreateAccountError(err)
	}

	acc2, err := NewAccount(ctx, c)
	if err != nil {
		return nil, newCreateAccountError(err)
	}

	if err := acc1.AddReferrer(ctx, c, acc2); err != nil {
		return nil, newCreateAccountError(err)
	}

	if err := acc2.RemoveDevice(ctx, c); err != nil {
		return nil, newCreateAccountError(err)
	}

	keys := s.Keys()

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(keys)))) // Range=[0; Length)
	if err != nil {
		n = big.NewInt(0)
	}

	key := keys[n.Int64()]
	if err := acc1.ApplyKey(ctx, c, key); err != nil {
		return nil, newCreateAccountError(err)
	}

	if err := acc1.ApplyKey(ctx, c, acc1.Account.License); err != nil {
		return nil, newCreateAccountError(err)
	}

	accData, err := acc1.GetAccountData(ctx, c)
	if err != nil {
		return nil, newCreateAccountError(err)
	}

	if err := acc1.RemoveDevice(ctx, c); err != nil {
		return nil, newCreateAccountError(err)
	}

	return accData, nil
}
