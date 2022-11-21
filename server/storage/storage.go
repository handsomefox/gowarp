// package storage is a stack for the keys
package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/handsomefox/gowarp/client"
	"github.com/handsomefox/gowarp/client/keygen"
	"github.com/handsomefox/gowarp/models"
	"github.com/handsomefox/gowarp/models/mongo"
)

type Storage struct {
	AM *mongo.AccountModel
}

// Fill fills the internal storage with correctly generated keys.
func (store *Storage) Fill(s *client.WarpService) {
	for {
		if store.AM.Len(context.Background()) > 250 {
			time.Sleep(10 * time.Second)
			continue
		}
		var wg errgroup.Group
		var createdKey *models.Account

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
			i, err := createdKey.RefCount.Int64()
			if err != nil {
				log.Println("couldn't get generated key size")
				continue
			}
			if i < 1000 {
				log.Println("generated key was too small to use")
				continue
			}

			id, err := store.AM.Insert(context.Background(), createdKey)
			if err != nil {
				log.Println("Failed to add key to the database", err)
			}
			log.Println("Added key to database, id: ", id)
		}
		log.Println("Currently stored keys: ", store.AM.Len(context.Background()))
		time.Sleep(s.WaitTime())
	}
}

// GetKey either returns a key that is already stored or creates a new one.
func (store *Storage) GetKey(ctx context.Context, s *client.WarpService) (*models.Account, error) {
	item, err := store.AM.GetAny(ctx)
	if err != nil {
		key, err := keygen.MakeKey(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("error generating key: %w", err)
		}
		return key, nil
	}
	if err := store.AM.Delete(ctx, item.ID); err != nil {
		log.Println("Failed to remove key from database: ", err)
	}

	log.Println("Currently stored key size: ", store.AM.Len(ctx))
	return item, nil
}
