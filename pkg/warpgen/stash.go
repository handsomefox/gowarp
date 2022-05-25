package warpgen

import (
	"fmt"
	"gowarp/pkg/config"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Stashed struct {
	acc    *accountData
	filled bool
	mtx    sync.Mutex
}

const (
	waitTime = 1*time.Minute + 30*time.Second
)

var (
	stash [config.RingSize]*Stashed
	index int64
)

func refillStash() {
	for i := 0; i < config.RingSize; i++ {
		refillAtIndex(int64(i))
	}
	fmt.Println("Refilled stash")
}

func refillAtIndex(index int64) {
	fmt.Printf("Refilling stashed key at index %v\n", index)

	if stash[index] != nil {
		fmt.Println("This key is already filled")
		return
	}

	start := time.Now()
	time.Sleep(waitTime)

	stash[index] = &Stashed{}

	stash[index].mtx.Lock()
	defer stash[index].mtx.Unlock()

	data, err := generateToStash()
	if err != nil {
		stash[index] = nil
	} else {
		gbs, _ := data.RefCount.Int64()
		if gbs < int64(100000) {
			stash[index] = nil
		} else {
			stash[index].acc = data
			stash[index].filled = true
		}
	}
	fmt.Printf("Filled entry %v in %v\n", index, time.Since(start))
}

func getFromStash() *accountData {
	if stash[index] == nil {
		return nil
	}

	if stash[index].filled && stash[index].acc == nil {
		stash[index] = nil
		go refillAtIndex(index)
		atomic.AddInt64(&index, 1)
		return nil
	}

	data := stash[index].acc

	fmt.Printf("Getting entry from index %v\n", index)
	stash[index] = nil
	go refillAtIndex(index)
	handleIndex()
	return data
}

func handleIndex() {
	atomic.AddInt64(&index, 1)
	if index == config.RingSize-1 {
		atomic.StoreInt64(&index, 0)
	}
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
