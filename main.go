package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"gowarp/pkg/server"

	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
)

func main() {
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   10,
		Interval: time.Minute,
	})
	if err != nil {
		panic(err)
	}

	middleware, err := httplimit.NewMiddleware(store, httplimit.IPKeyFunc())
	if err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(":8080"+os.Getenv("PORT"), middleware.Handle(server.New())))
}
