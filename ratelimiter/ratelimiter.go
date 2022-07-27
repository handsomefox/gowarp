package ratelimiter

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips   map[string]*rate.Limiter
	mutex *sync.RWMutex
	rate  rate.Limit
	size  int
}

func New(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:   make(map[string]*rate.Limiter),
		mutex: &sync.RWMutex{},
		rate:  r,
		size:  b,
	}
}

// addIP creates a new rate limiter and adds it to the ips map,
// using the IP address as the key.
func (i *IPRateLimiter) addIP(ip string) *rate.Limiter {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	limiter := rate.NewLimiter(i.rate, i.size)

	i.ips[ip] = limiter

	return limiter
}

// getLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise calls AddIP to add IP address to the map.
func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	limiter, exists := i.ips[ip]

	if !exists {
		return i.addIP(ip)
	}

	return limiter
}

func NewMiddleware(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiter := limiter.getLimiter(readUserIP(r))
			if !limiter.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)

				return
			}
			next.ServeHTTP(w, r)
		})
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
