package warp

import (
	"fmt"
	"log"
	"sync"
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

func (s *Storage) Fill(config *Config) {
	for {
		var (
			progressChan = make(chan int)
			wg           = new(sync.WaitGroup)
			key          *AccountData
			err          error
		)

		wg.Add(1)

		go func(*Config, chan int) {
			defer wg.Done()
			defer close(progressChan)

			key, err = Generate(config, progressChan)
		}(config, progressChan)

		wg.Add(1)

		go func(chan int) {
			defer wg.Done()

			for progress := range progressChan {
				log.Printf("current key generation progress: %v", progress)
			}
		}(progressChan)

		wg.Wait()

		if err != nil {
			log.Printf("error when generating key: %v", err)
		} else {
			log.Println("added key to storage")
			s.keyChan <- *key
		}

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
