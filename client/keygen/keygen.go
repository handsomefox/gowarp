package keygen

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"time"

	"github.com/handsomefox/gowarp/client"
	"github.com/handsomefox/gowarp/client/account"
)

type CreateAccountError struct {
	reason string
}

func (e CreateAccountError) Error() string {
	return "error creating account: " + e.reason
}

func NewCreateAccountError(reason string) CreateAccountError {
	return CreateAccountError{
		reason: reason,
	}
}

func MakeKey(ctx context.Context, c *client.Client) (*account.Data, error) {
	log.Println("Started generating key")
	start := time.Now()

	c.UseProxy(ctx)
	defer c.UnuseProxy()

	acc1, err := account.NewAccount(ctx, c)
	if err != nil {
		return nil, NewCreateAccountError(err.Error())
	}

	acc2, err := account.NewAccount(ctx, c)
	if err != nil {
		return nil, NewCreateAccountError(err.Error())
	}

	if err := acc1.AddReferrer(ctx, c, acc2); err != nil {
		return nil, NewCreateAccountError(err.Error())
	}

	if err := acc2.RemoveDevice(ctx, c); err != nil {
		return nil, NewCreateAccountError(err.Error())
	}

	keys := c.Keys()

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(keys)))) // Range=[0; Length)
	if err != nil {
		n = big.NewInt(0)
	}

	key := keys[n.Int64()]
	log.Printf("Used key: %s", key)
	if err := acc1.ApplyKey(ctx, c, key); err != nil {
		return nil, NewCreateAccountError(err.Error())
	}

	if err := acc1.ApplyKey(ctx, c, acc1.Account.License); err != nil {
		return nil, NewCreateAccountError(err.Error())
	}

	accData, err := acc1.GetAccountData(ctx, c)
	if err != nil {
		return nil, NewCreateAccountError(err.Error())
	}

	if err := acc1.RemoveDevice(ctx, c); err != nil {
		return nil, NewCreateAccountError(err.Error())
	}

	log.Printf("Generating key took: %vms", time.Since(start).Milliseconds())
	log.Printf("Generated key size: %v", accData.RefCount.String())
	return accData, nil
}
