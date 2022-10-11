package server

import (
	"net/http"
	"sync"
	"time"
)

const (
	requestLimit  = 500
	requestPeriod = time.Hour * 6
)

type ipRequestCounter struct {
	ips map[string]int
	mu  sync.Mutex
}

func (sc *ipRequestCounter) Increment(key string) {
	sc.mu.Lock()
	sc.ips[key]++
	sc.mu.Unlock()
}

func (sc *ipRequestCounter) Get(key string) int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.ips[key]
}

type RateLimiter struct {
	requestCounter *ipRequestCounter
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requestCounter: &ipRequestCounter{
			ips: make(map[string]int, 0),
		},
	}
	go rl.clearBlockedIPs()

	return rl
}

func (rl *RateLimiter) Decorate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipAddr := readUserIP(r)
		rl.requestCounter.Increment(ipAddr)
		cv := rl.requestCounter.Get(ipAddr)

		if cv >= requestLimit {
			errorWithCode(w, http.StatusTooManyRequests)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) clearBlockedIPs() {
	for {
		rl.requestCounter.mu.Lock()
		rl.requestCounter.ips = make(map[string]int)
		rl.requestCounter.mu.Unlock()
		time.Sleep(requestPeriod)
	}
}

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
