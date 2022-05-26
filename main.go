package main

import (
	"fmt"
	"gowarp/pkg/warpgen"
	"log"
	"net/http"
	"os"
	"sync"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips   map[string]*rate.Limiter
	mutex *sync.RWMutex
	rate  rate.Limit
	size  int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips:   make(map[string]*rate.Limiter),
		mutex: &sync.RWMutex{},
		rate:  r,
		size:  b,
	}

	return i
}

// AddIP creates a new rate limiter and adds it to the ips map,
// using the IP address as the key
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	limiter := rate.NewLimiter(i.rate, i.size)

	i.ips[ip] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise calls AddIP to add IP address to the map
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mutex.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		i.mutex.Unlock()
		return i.AddIP(ip)
	}

	i.mutex.Unlock()

	return limiter
}

// warp is an http.HandleFunc that generates a warp+ key and writes it to the http.ResponseWriter
func warp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := warpgen.Generate(w, r); err != nil {
		fmt.Fprintf(w, "\nError when creating keys: %v\n", err)
	}
}

// configUpdate is an endpoint that triggers the configuration update
// allowing you to manually trigger the updates
func configUpdate(w http.ResponseWriter, r *http.Request) {
	warpgen.TriggerUpdate()
	fmt.Fprintf(w, "Updated config!\n")
}

var limiter = NewIPRateLimiter(1, 1)

func main() {
	mux := http.NewServeMux()

	warpHandler := http.HandlerFunc(warp)
	configHandler := http.HandlerFunc(configUpdate)

	mux.Handle("/", limitMiddleware(warpHandler))
	mux.Handle("/config/update", limitMiddleware(configHandler))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), mux))
}

func limitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := limiter.GetLimiter(r.RemoteAddr)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
