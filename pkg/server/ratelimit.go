package server

import (
	"sync"
)

type IPRequestCounter struct {
	ips map[string]int
	mu  sync.Mutex
}

func (sc *IPRequestCounter) Inc(key string) {
	sc.mu.Lock()
	sc.ips[key]++
	sc.mu.Unlock()
}

func (sc *IPRequestCounter) Value(key string) int {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return sc.ips[key]
}
