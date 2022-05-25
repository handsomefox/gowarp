package warpgen

import (
	"fmt"
	"gowarp/pkg/config"
	"math/rand"
	"sync"
	"time"
)

type Stashed struct {
	acc    *accountData
	filled bool
	mutex  sync.Mutex
}

const (
	waitTime = 1*time.Minute + 30*time.Second
)

var (
	stash [config.RingSize]*Stashed
)

func refillStash() {
	for i := 0; i < config.RingSize; i++ {
		refillAtIndex(int64(i), waitTime)
	}
	fmt.Println("Refilled stash")
}

func refillAtIndex(index int64, sleepTime time.Duration) {
	fmt.Printf("Refilling stashed key at index %v\n", index)

	if stash[index] != nil {
		fmt.Println("This key is already filled")
		return
	}

	start := time.Now()
	time.Sleep(sleepTime)

	stash[index] = &Stashed{}

	stash[index].mutex.Lock()
	defer stash[index].mutex.Unlock()

	data, err := generateToStash()
	if err != nil {
		fmt.Println("Error when refilling a key")
		stash[index] = nil
		go refillAtIndex(index, waitTime+randomAdditionalTime())
	} else {
		gbs, _ := data.RefCount.Int64()
		if gbs < int64(100000) {
			fmt.Println("Account limit was to small when refilling the key")
			stash[index] = nil
			go refillAtIndex(index, waitTime+randomAdditionalTime())
		} else {
			fmt.Println("Refilled successfully")
			stash[index].acc = data
			stash[index].filled = true
		}
	}
	fmt.Printf("Filled entry %v in %v\n", index, time.Since(start))
}

func getFromStash() *accountData {
	for i := 0; i < config.RingSize; i++ {
		if stash[i] != nil {
			stash[i].mutex.Lock()
			defer stash[i].mutex.Unlock()

			data := stash[i].acc
			fmt.Printf("Getting entry from index %v\n", i)

			stash[i] = nil

			go refillAtIndex(int64(i), waitTime+randomAdditionalTime())
			return data
		}
	}
	return nil
}

func randomAdditionalTime() time.Duration {
	return time.Duration(rand.Intn(180)) * time.Second
}

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
