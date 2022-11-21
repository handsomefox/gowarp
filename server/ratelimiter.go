package server

import (
	"net/http"
	"sync"
	"time"
)

var (
	RequestLimit  = 500
	RequestPeriod = time.Hour * 6
)

type IPRequestCount struct {
	mu  sync.Mutex
	ips map[string]int
}

func (sc *IPRequestCount) Increment(key string) {
	sc.mu.Lock()
	sc.ips[key]++
	sc.mu.Unlock()
}

func (sc *IPRequestCount) Get(key string) int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.ips[key]
}

type RateLimiter struct {
	requestCounter *IPRequestCount
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requestCounter: &IPRequestCount{
			ips: make(map[string]int, 0),
		},
	}
	go rl.Clear()

	return rl
}

func (rl *RateLimiter) Decorate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipAddr := ReadUserIP(r)
		rl.requestCounter.Increment(ipAddr)
		cv := rl.requestCounter.Get(ipAddr)

		if cv >= RequestLimit {
			errorWithCode(w, http.StatusTooManyRequests)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) Clear() {
	for {
		rl.requestCounter.mu.Lock()
		rl.requestCounter.ips = make(map[string]int)
		rl.requestCounter.mu.Unlock()
		time.Sleep(RequestPeriod)
	}
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
