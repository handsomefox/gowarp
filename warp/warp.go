// warp is a package used for generating cfwarp+ keys
package warp

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"gowarp/progressbar"
)

const pastebinURL = "https://pastebin.com/raw/pwtQLBiK"

type Warp struct {
	config  *Config
	storage *Storage
}

func (warp *Warp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	fstr := "\nAccount type: %v\nData available: %vGB\nLicense: %v\n"

	// fast path
	storageKey, err := warp.storage.GetKey(&warp.GetConfig().data)
	if err == nil {
		log.Println("fast path, got key from stash")
		fmt.Fprintf(w, fstr, storageKey.Type, storageKey.RefCount, storageKey.License)

		return
	}

	// slow path
	log.Println("slow path, generating key")

	key, err := warp.Generate(w, r)
	if err != nil {
		fmt.Fprintln(w, fmt.Errorf("\nError when creating keys: %w", err))

		return
	}

	fmt.Fprintf(w, fstr, key.Type, key.RefCount, key.License)
}

func New() *Warp {
	cfg := NewConfig()

	if err := cfg.Update(pastebinURL); err != nil {
		log.Printf("error updating config: %v", err)
	}

	store := NewStorage()

	// start a goroutine that will be actively trying to update the storage
	go store.Fill(cfg)

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
		storage: store,
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

func (warp *Warp) Generate(w http.ResponseWriter, r *http.Request) (*AccountData, error) {
	var (
		progressChan = make(chan int)
		config       = warp.GetConfig()
		wg           = new(sync.WaitGroup)

		key *AccountData
		err error
	)

	handleBrowsers(w, r)

	wg.Add(1)

	go func(*Config, chan int) {
		defer wg.Done()
		defer close(progressChan)

		key, err = Generate(config, progressChan)
	}(config, progressChan)

	wg.Add(1)

	go func(chan int) {
		defer wg.Done()

		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("no flusher")
		}

		pb := progressbar.New(w, flusher)

		for progress := range progressChan {
			pb.Update(progress)
		}
	}(progressChan)

	wg.Wait()

	if err != nil {
		return nil, err
	}

	return key, nil
}

// Generate handles generating a key for user.
func Generate(config *Config, progressChan chan int) (*AccountData, error) {
	var (
		client   = createClient()
		cfg      = config.Get()
		progress = 0
	)
	progressChan <- progress

	progress += 10
	progressChan <- progress

	acc1, err := registerAccount(&cfg, client)
	if err != nil {
		return nil, err
	}

	progress += 10
	progressChan <- progress

	acc2, err := registerAccount(&cfg, client)
	if err != nil {
		return nil, err
	}

	progress += 10
	progressChan <- progress

	if err := acc1.addReferrer(&cfg, client, acc2); err != nil {
		return nil, err
	}

	progress += 10
	progressChan <- progress

	if err := acc2.removeDevice(&cfg, client); err != nil {
		return nil, err
	}

	progress += 10
	progressChan <- progress

	keys := cfg.Keys

	if err := acc1.setKey(&cfg, client, keys[rand.Intn(len(keys))]); err != nil {
		return nil, err
	}

	progress += 10
	progressChan <- progress

	if err := acc1.setKey(&cfg, client, acc1.Account.License); err != nil {
		return nil, err
	}

	progress += 10
	progressChan <- progress

	accData, err := acc1.fetchAccountData(&cfg, client)
	if err != nil {
		return nil, err
	}

	progress += 10
	progressChan <- progress

	if err := acc1.removeDevice(&cfg, client); err != nil {
		return nil, err
	}

	progress += 10
	progressChan <- progress

	return accData, nil
}
