package warp

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const storageSize = 20 // length for the buffered channel where the generated keys are stored

type Storage struct {
	keyChan chan AccountData
}

func NewStorage() *Storage {
	return &Storage{
		keyChan: make(chan AccountData, storageSize),
	}
}

func (store *Storage) Fill(config *Config) {
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
		}(config)

		wg.Wait()

		if err != nil {
			log.Printf("error when generating key: %v", err)
		} else {
			log.Println("added key to storage")
			store.keyChan <- *key
		}

		time.Sleep(time.Minute + randomTime())
	}
}

func (store *Storage) GetKey(cdata *ConfigData) (AccountData, error) {
	select {
	case accountData, ok := <-store.keyChan:
		if !ok {
			return AccountData{}, fmt.Errorf("channel is closed, can't get the key")
		}
		log.Println("got key from storage")
		return accountData, nil
	default:
		return AccountData{}, fmt.Errorf("no key was found")
	}
}
