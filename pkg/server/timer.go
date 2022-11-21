package server

import (
	"log"
	"net/http"
	"time"
)

func timerMiddleware() DecoratorFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)

			log.Printf("[%s] request to %s took %dms", r.Method, r.URL, time.Since(start).Milliseconds())
		})
	}
}
