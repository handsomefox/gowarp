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

func New() *Warp {
	cfg := NewConfig()

	if err := cfg.Update(pastebinURL); err != nil {
		log.Printf("error updating config: %v", err)
	}

	store := &Storage{
		mu:   sync.Mutex{},
		keys: [20]*AccountData{},
	}

	// start a goroutine that will be actively trying to update the storage
	go func() {
		for {
			config := cfg.Get()
			store.Fill(&config)
			time.Sleep(12 * time.Hour)
		}
	}()

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

// Generate handles generating a key for user.
func (warp *Warp) Generate(w http.ResponseWriter, r *http.Request) error {
	var (
		client   = createClient()
		config   = warp.config.Get()
		progress = 0
	)

	handleBrowsers(w, r)

	key, err := warp.storage.GetKey(&config)
	if err == nil {
		log.Println("Got key from stash")

		str := fmt.Sprintf("Account type: %v\nData available: %vGB\nLicense: %v\n",
			key.Type, key.RefCount, key.License)

		fmt.Fprint(w, str)

		return nil
	}

	log.Println("Couldn't get key from stash, going the slow path")

	flusher, _ := w.(http.Flusher)
	pb := progressbar.New(w, flusher)

	progress += 10
	pb.Update(progress)

	acc1, err := registerAccount(&config, client)
	if err != nil {
		return err
	}

	progress += 10
	pb.Update(progress)

	acc2, err := registerAccount(&config, client)
	if err != nil {
		return err
	}

	progress += 10
	pb.Update(progress)

	if err := acc1.addReferrer(&config, client, acc2); err != nil {
		return err
	}

	progress += 10
	pb.Update(progress)

	if err := acc2.removeDevice(&config, client); err != nil {
		return nil
	}

	progress += 10
	pb.Update(progress)

	keys := config.Keys

	if err := acc1.setKey(&config, client, keys[rand.Intn(len(keys))]); err != nil {
		return err
	}

	progress += 10
	pb.Update(progress)

	if err := acc1.setKey(&config, client, acc1.Account.License); err != nil {
		return err
	}

	progress += 10
	pb.Update(progress)

	accData, err := acc1.fetchAccountData(&config, client)
	if err != nil {
		return err
	}

	progress += 10
	pb.Update(progress)

	if err := acc1.removeDevice(&config, client); err != nil {
		return err
	}

	progress += 10
	pb.Update(progress)

	str := fmt.Sprintf("\n\nAccount type: %v\nData available: %vGB\nLicense: %v\n",
		accData.Type, accData.RefCount, accData.License)

	fmt.Fprint(w, str)

	return nil
}

func (warp *Warp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	if err := warp.Generate(w, r); err != nil {
		fmt.Fprintln(w, fmt.Errorf("\nError when creating keys: %w", err))
	}
}
