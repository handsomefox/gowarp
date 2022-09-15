// warp is a package used for generating cfwarp+ keys
package warp

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const pastebinURL = "https://pastebin.com/raw/pwtQLBiK" // this is where the plaintext config is stored right now

type Warp struct {
	config  *Config
	storage *Storage
}

func (warp *Warp) GetKey() (*AccountData, error) {
	storageKey, err := warp.storage.GetKey(&warp.GetConfig().cdata)
	if err == nil {
		log.Println("fast path, got key from stash")
		return &storageKey, nil
	}

	log.Println("slow path, generating key")
	generatedKey, err := warp.Generate()
	if err != nil {
		return nil, fmt.Errorf("\nError when creating keys: %w", err)
	}

	return generatedKey, nil
}

func (warp *Warp) UpdateConfig() error {
	return warp.config.Update(pastebinURL)
}

func New() *Warp {
	cfg := NewConfig()

	if err := cfg.Update(pastebinURL); err != nil {
		log.Printf("error updating config: %v", err)
	}

	storage := NewStorage()

	// start a goroutine that will be actively trying to update the storage
	go storage.Fill(cfg)

	// start a goroutine that will be trying to update the config something like every 6 hours
	go func() {
		for {
			time.Sleep(6 * time.Hour)

			if err := cfg.Update(pastebinURL); err != nil {
				log.Printf("error updating config: %v", err)
			}
		}
	}()

	return &Warp{
		config:  cfg,
		storage: storage,
	}
}

func (warp *Warp) GetConfig() *Config {
	return warp.config
}

func (warp *Warp) Update() {
	if err := warp.config.Update(pastebinURL); err != nil {
		log.Println(err)
	}
}

func (warp *Warp) Generate() (*AccountData, error) {
	var (
		config = warp.GetConfig()
		wg     = new(sync.WaitGroup)

		key *AccountData
		err error
	)

	wg.Add(1)
	go func(*Config) {
		defer wg.Done()

		key, err = Generate(config)
	}(config)

	wg.Wait()

	if err != nil {
		return nil, err
	}

	return key, nil
}

// Generate handles generating a key for user.
func Generate(config *Config) (*AccountData, error) {
	var (
		client = newClient()
		cfg    = config.Get()
	)

	acc1, err := registerAccount(client, &cfg)
	if err != nil {
		return nil, err
	}

	acc2, err := registerAccount(client, &cfg)
	if err != nil {
		return nil, err
	}

	if err := acc1.addReferrer(client, &cfg, acc2); err != nil {
		return nil, err
	}

	if err := acc2.removeDevice(client, &cfg); err != nil {
		return nil, err
	}

	keys := cfg.Keys

	if err := acc1.setKey(client, &cfg, keys[rand.Intn(len(keys))]); err != nil {
		return nil, err
	}

	if err := acc1.setKey(client, &cfg, acc1.Account.License); err != nil {
		return nil, err
	}

	accData, err := acc1.fetchAccountData(client, &cfg)
	if err != nil {
		return nil, err
	}

	if err := acc1.removeDevice(client, &cfg); err != nil {
		return nil, err
	}

	return accData, nil
}
