package warp

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Storage struct {
	keys [20]*AccountData
	mu   sync.Mutex
}

func (s *Storage) Fill(config *ConfigData) {
	for i := 0; i < len(s.keys); i++ {
		s.mu.Lock()

		key := s.keys[i]
		if key == nil {
			s.UpdateIndex(config, i)
		}

		s.mu.Unlock()

		time.Sleep(time.Minute + randomTime())
	}
}

func (s *Storage) GetKey(config *ConfigData) (AccountData, error) {
	for i := 0; i < len(s.keys); i++ {
		s.mu.Lock()

		key := s.keys[i]
		if key != nil {
			returnedKey := *key
			s.keys[i] = nil

			go func(config *ConfigData, index int) {
				time.Sleep(time.Minute + randomTime())
				s.UpdateIndex(config, index)
			}(config, i)

			return returnedKey, nil
		}

		s.mu.Unlock()
	}

	return AccountData{}, fmt.Errorf("no key was found")
}

func (s *Storage) UpdateIndex(config *ConfigData, index int) {
	if index <= 0 || index > 20 {
		return
	}

	log.Printf("updating key at index %d", index)

	s.mu.Lock()
	defer s.mu.Unlock()

	key := &s.keys[index]
	if key != nil {
		return
	}

	newKey, err := Generate(config)
	if err != nil {
		return
	}

	log.Printf("updated key at index %d", index)

	s.keys[index] = newKey
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
