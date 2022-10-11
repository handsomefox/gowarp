package warp

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var ErrGetKey = errors.New("error getting the key from storage")

const storageSize = 20 // length for the buffered channel where the generated keys are stored

// Storage is the internal storage for keys with automatically generates and stores them.
// Use (*Storage).Fill(&warp.Config) on a goroutine.
type Storage struct {
	keyChan chan *AccountData
}

// NewStorage returns a new instance of *Storage with a buffered channel.
func NewStorage() *Storage {
	return &Storage{
		keyChan: make(chan *AccountData, storageSize),
	}
}

// Fill fills the internal storage with correctly generated keys.
func (store *Storage) Fill(cfg *Config) {
	for {
		var (
			key *AccountData
			err error
			wg  sync.WaitGroup
		)

		progress := make(chan int)

		wg.Add(1)
		go func(config *Config) {
			defer wg.Done()
			defer close(progress)

			key, err = Generate(config)
		}(cfg)

		wg.Wait()

		if err != nil {
			log.Printf("error when generating key: %s", err)
		} else {
			log.Println("added key to storage")
			store.keyChan <- key
		}

		time.Sleep(2 * time.Minute)
	}
}

// GetKey returns a valid key from the storage or an error.
func (store *Storage) GetKey(cfg *ConfigData) (*AccountData, error) {
	select {
	case accountData, ok := <-store.keyChan:
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrGetKey, "channel is closed, can't get the key")
		}
		log.Println("Got key from storage", accountData)
		return accountData, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrGetKey, "no key was found")
	}
}
