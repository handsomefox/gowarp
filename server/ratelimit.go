package server

import (
	"net/http"
	"sync"
	"time"
)

func RateLimit(h http.Handler, requestLimit int, requestPeriod time.Duration) http.HandlerFunc {
	rl := &rateLimiter{
		requestCounter: &ipRequestCount{ips: make(map[string]int, 0), mu: sync.Mutex{}},
		requestPeriod:  requestPeriod,
		requestLimit:   requestLimit,
	}
	go rl.clear()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipAddr := readUserIP(r)
		rl.requestCounter.increment(ipAddr)
		cv := rl.requestCounter.get(ipAddr)

		if cv >= rl.requestLimit {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		h.ServeHTTP(w, r)
	})
}

type rateLimiter struct {
	requestCounter *ipRequestCount
	requestPeriod  time.Duration
	requestLimit   int
}

func (rl *rateLimiter) clear() {
	for {
		rl.requestCounter.mu.Lock()
		rl.requestCounter.ips = make(map[string]int)
		rl.requestCounter.mu.Unlock()
		time.Sleep(rl.requestPeriod)
	}
}

type ipRequestCount struct {
	mu  sync.Mutex
	ips map[string]int
}

func (sc *ipRequestCount) increment(key string) {
	sc.mu.Lock()
	sc.ips[key]++
	sc.mu.Unlock()
}

func (sc *ipRequestCount) get(key string) int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.ips[key]
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
