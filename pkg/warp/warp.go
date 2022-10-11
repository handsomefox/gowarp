// warp is a package used for generating cfwarp+ keys
package warp

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"
)

const pastebinURL = "https://pastebin.com/raw/pwtQLBiK" // this is where the plaintext config is stored right now

type Warp struct {
	config  *Config
	storage *Storage
}

func (warp *Warp) GetKey(ctx context.Context) (*AccountData, error) {
	storageKey, err := warp.storage.GetKey(&warp.GetConfig().cdata)
	if err != nil {
		log.Printf("Couldn't get key from storage: %s", err)
	} else {
		return storageKey, nil
	}

	log.Println("Going the slow path")
	generatedKey, err := warp.Generate(ctx)
	if err != nil {
		return nil, fmt.Errorf("error when creating keys: %w", err)
	}

	return generatedKey, nil
}

func (warp *Warp) UpdateConfig(ctx context.Context) error {
	return warp.config.Update(ctx, pastebinURL)
}

func New() *Warp {
	cfg := NewConfig()

	if err := cfg.Update(context.TODO(), pastebinURL); err != nil {
		log.Printf("error updating config: %v", err)
	}

	storage := NewStorage()

	// start a goroutine that will be actively trying to update the storage
	go storage.Fill(cfg)

	// start a goroutine that will be trying to update the config something like every 6 hours
	go func() {
		for {
			time.Sleep(6 * time.Hour)

			if err := cfg.Update(context.TODO(), pastebinURL); err != nil {
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

func (warp *Warp) Update(ctx context.Context) {
	if err := warp.config.Update(ctx, pastebinURL); err != nil {
		log.Println(err)
	}
}

func (warp *Warp) Generate(ctx context.Context) (*AccountData, error) {
	cfg := warp.GetConfig().Get()
	key, err := Generate(ctx, &cfg)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Generate handles generating a key for user.
func Generate(ctx context.Context, cfg *ConfigData) (*AccountData, error) {
	log.Println("Started generating key")
	start := time.Now()

	acc1, err := NewAccount(ctx, cfg)
	if err != nil {
		return nil, err
	}

	acc2, err := NewAccount(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := acc1.addReferrer(ctx, cfg, acc2); err != nil {
		return nil, err
	}

	if err := acc2.removeDevice(ctx, cfg); err != nil {
		return nil, err
	}

	keys := cfg.Keys

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(keys)))) // Range=[0; Length)
	if err != nil {
		n = big.NewInt(0)
	}

	if err := acc1.setKey(ctx, cfg, keys[n.Int64()]); err != nil {
		return nil, err
	}

	if err := acc1.setKey(ctx, cfg, acc1.Account.License); err != nil {
		return nil, err
	}

	accData, err := acc1.fetchAccountData(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := acc1.removeDevice(ctx, cfg); err != nil {
		return nil, err
	}

	log.Printf("Generating key took: %vms", time.Since(start).Milliseconds())
	return accData, nil
}
