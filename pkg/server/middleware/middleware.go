package middleware

import (
	"log"
	"net/http"
	"time"
)

func RateLimiter(requestLimit int, requestPeriod time.Duration) DecoratorFunc {
	rl := newRateLimiter(requestLimit, requestPeriod)
	return func(h http.Handler) http.Handler {
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
}

func RequestTimer() DecoratorFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			log.Printf("[%s] request from %s to %s took %dms", r.Method, readUserIP(r), r.URL, time.Since(start).Milliseconds())
		})
	}
}
