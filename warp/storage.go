package warp

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

const storeSize = 20

type Storage struct {
	keyChan chan AccountData
}

func NewStorage() *Storage {
	return &Storage{
		keyChan: make(chan AccountData, storeSize),
	}
}

func (s *Storage) Fill(config *ConfigData) {
	for i := 0; i < storeSize; i++ {
		key, err := Generate(config)
		if err != nil {
			continue
		}

		log.Println("added key to storage")
		s.keyChan <- *key

		time.Sleep(time.Minute + randomTime())
	}
}

func (s *Storage) GetKey(config *ConfigData) (AccountData, error) {
	select {
	case v, ok := <-s.keyChan:
		if ok {
			log.Println("got key from storage")

			return v, nil
		}

		return AccountData{}, fmt.Errorf("channel is closed")
	default:
		return AccountData{}, fmt.Errorf("no key was found")
	}
}

// Generate handles generating a key for user.
func Generate(config *ConfigData) (*AccountData, error) {
	client := createClient()

	acc1, err := registerAccount(config, client)
	if err != nil {
		return nil, err
	}

	acc2, err := registerAccount(config, client)
	if err != nil {
		return nil, err
	}

	if err := acc1.addReferrer(config, client, acc2); err != nil {
		return nil, err
	}

	if err := acc2.removeDevice(config, client); err != nil {
		return nil, err
	}

	keys := config.Keys

	if err := acc1.setKey(config, client, keys[rand.Intn(len(keys))]); err != nil {
		return nil, err
	}

	if err := acc1.setKey(config, client, acc1.Account.License); err != nil {
		return nil, err
	}

	accData, err := acc1.fetchAccountData(config, client)
	if err != nil {
		return nil, err
	}

	if err := acc1.removeDevice(config, client); err != nil {
		return nil, err
	}

	return accData, nil
}
