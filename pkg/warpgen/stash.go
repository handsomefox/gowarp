package warpgen

import (
	"fmt"
	"gowarp/pkg/config"
	"math/rand"
	"sync"
	"time"
)

// StashedValue represents a value that is saved in stash
type StashedValue struct {
	acc    accountData
	filled bool
}

// Stash is is a storage space which is used for storing cached keys for users
// to not have to wait for generation every time they try to get a key
type Stash struct {
	store [config.RingSize]*StashedValue
	mutex sync.Mutex
}

var (
	stash = Stash{}
)

// refillStash goes through the whole stash and calls refillAtIndex(index)
func refillStash() {
	for i := 0; i < config.RingSize; i++ {
		refillAtIndex(int64(i), config.WaitTime)
	}
	fmt.Println("Refilled stash")
}

// refillAtIndex checks if the value at a given index is valid, if not, it regenerates a key.
// If generation fails or the limit is too small, it launches a goroutine which will have a
// longer sleepTime and will try to update the key in some time in the future
func refillAtIndex(index int64, sleepTime time.Duration) {
	fmt.Printf("Refilling stashed key at index %v\n", index)

	if stash.store[index] != nil {
		fmt.Println("This key is already filled")
		return
	}

	start := time.Now()
	time.Sleep(sleepTime)

	stash.store[index] = &StashedValue{}

	stash.mutex.Lock()
	defer stash.mutex.Unlock()

	data, err := generateToStash()
	if err != nil {
		fmt.Println("Error when refilling a key")
		stash.store[index] = nil
		go refillAtIndex(index, config.WaitTime+randomAdditionalTime())
	} else {
		gbs, _ := data.RefCount.Int64()
		if gbs < int64(100000) {
			fmt.Println("Account limit was to small when refilling the key")
			stash.store[index] = nil
			go refillAtIndex(index, config.WaitTime+randomAdditionalTime())
		} else {
			fmt.Println("Refilled successfully")
			stash.store[index].acc = *data
			stash.store[index].filled = true
		}
	}
	fmt.Printf("Filled entry %v in %v\n", index, time.Since(start))
}

// getFromStash is a function that handles getting a key from stash
// if it returns a key, it will launch a goroutine to refill the key at the index
// at which the key was given.
func getFromStash() *accountData {

	for i := 0; i < config.RingSize; i++ {
		if stash.store[i] != nil {
			stash.mutex.Lock()
			defer stash.mutex.Unlock()

			data := stash.store[i].acc
			fmt.Printf("Getting entry from index %v\n", i)

			stash.store[i] = nil

			go refillAtIndex(int64(i), config.WaitTime+randomAdditionalTime())
			return &data
		}
	}
	return nil
}

// randomAdditionalTime is a helper function that generates a random time.Duration
// in range 0-180 seconds
func randomAdditionalTime() time.Duration {
	return time.Duration(rand.Intn(180)) * time.Second
}

// generateToStash is basically the same as warpgen.Generate, but only for generating keys
// it handles the whole process of getting a new key for warp+
func generateToStash() (*accountData, error) {
	client := createClient()
	acc1, err := registerAccount(client)
	if err != nil {
		return nil, err
	}
	acc2, err := registerAccount(client)
	if err != nil {
		return nil, err
	}
	if err := acc1.addReferrer(client, acc2); err != nil {
		return nil, err
	}
	if err := acc2.removeDevice(client); err != nil {
		return nil, err
	}
	if err := acc1.setKey(client, config.KeyStorage.Keys[rand.Intn(len(config.KeyStorage.Keys))]); err != nil {
		return nil, err
	}
	if err := acc1.setKey(client, acc1.Account.License); err != nil {
		return nil, err
	}
	accData, err := acc1.fetchAccountData(client)
	if err != nil {
		return nil, err
	}
	return accData, nil
}
