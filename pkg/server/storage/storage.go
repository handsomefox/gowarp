// package storage is a stack for the keys
package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/handsomefox/gowarp/pkg/models"
	"github.com/handsomefox/gowarp/pkg/models/mongo"
	"github.com/handsomefox/gowarp/pkg/warp"
)

type Storage struct {
	AM *mongo.AccountModel
}

// Fill fills (forever) the internal storage with correctly generated keys.
func (store *Storage) Fill(s *warp.Service) {
	for {
		if store.AM.Len(context.Background()) > 250 {
			time.Sleep(10 * time.Second)
			continue
		}

		store.makeKey(s)

		log.Println("Currently storing: ", store.AM.Len(context.Background()), " keys")
		time.Sleep(s.WaitTime())
	}
}

// makeKey wraps the warp.MakeKey and stores the key inside database.
func (store *Storage) makeKey(s *warp.Service) {
	start := time.Now()

	var wg errgroup.Group
	var createdKey *models.Account
	wg.Go(func() error {
		key, err := warp.MakeKey(context.Background(), s)
		if err != nil {
			return fmt.Errorf("error generating key: %w", err)
		}
		createdKey = key
		return nil
	})

	if err := wg.Wait(); err != nil {
		log.Println("Error when generating key: ", err)
		return
	}

	log.Println("Generating key took: ", time.Since(start))

	i, err := createdKey.RefCount.Int64()
	if err != nil {
		log.Println("Couldn't get generated key size: ", err)
		return
	}
	if i < 1000 {
		log.Println("Generated key was too small to use: ", i)
		return
	}

	id, err := store.AM.Insert(context.Background(), createdKey)
	if err != nil {
		log.Println("Failed to add key to the database: ", err)
	}

	log.Println("Added key to database, id: ", id)
}

// GetKey either returns a key that is already stored or creates a new one.
func (store *Storage) GetKey(ctx context.Context, s *warp.Service) (*models.Account, error) {
	item, err := store.AM.GetAny(ctx)
	if err != nil {
		key, err := warp.MakeKey(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("error generating key: %w", err)
		}
		return key, nil
	}
	if err := store.AM.Delete(ctx, item.ID); err != nil {
		log.Println("Failed to remove key from database: ", err)
	}

	log.Println("Currently storing: ", store.AM.Len(ctx), " keys")
	return item, nil
}
