package main

import (
	"log"
	"net/http"
	"os"

	"gowarp/ratelimiter"
	"gowarp/warp"
)

func main() {
	mux := http.NewServeMux()

	limiter := ratelimiter.New(1, 4)
	limiterMiddleware := ratelimiter.NewMiddleware(limiter)
	warpHandle := warp.New()

	mux.Handle("/", limiterMiddleware(warpHandle))
	mux.Handle("/config/update", limiterMiddleware(warpHandle.GetConfig()))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), mux))
}
