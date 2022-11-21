// package storage is a stack for the keys
package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/handsomefox/gowarp/client"
	"github.com/handsomefox/gowarp/client/account"
	"github.com/handsomefox/gowarp/client/keygen"
)

type Storage struct {
	stack *Stack
}

func NewStorage() *Storage {
	return &Storage{
		stack: NewStack(),
	}
}

// Fill fills the internal storage with correctly generated keys.
func (store *Storage) Fill(s *client.WarpService) {
	for {
		if store.stack.Len() > 40 {
			time.Sleep(10 * time.Second)
			continue
		}
		var wg errgroup.Group
		var createdKey *account.Data

		wg.Go(func() error {
			key, err := keygen.MakeKey(context.Background(), s)
			if err != nil {
				return fmt.Errorf("error generating key: %w", err)
			}
			createdKey = key

			return nil
		})

		if err := wg.Wait(); err != nil {
			log.Printf("Error when generating key: %s", err)
		} else {
			store.stack.Push(*createdKey)
			log.Println("Added key to storage")
		}
		log.Println("Currently stored key size: ", store.stack.Len())
	}
}

// GetKey either returns a key that is already stored or creates a new one.
func (store *Storage) GetKey(ctx context.Context, s *client.WarpService) (*account.Data, error) {
	item, err := store.stack.Pop()
	if err != nil {
		key, err := keygen.MakeKey(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("error generating key: %w", err)
		}
		return key, nil
	}
	log.Println("Currently stored key size: ", store.stack.Len())
	return item, nil
}
