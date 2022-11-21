package main

import (
	"net/http"
	"os"
	"time"

	"github.com/handsomefox/gowarp/pkg/server"
)

func main() {
	sh, err := server.NewServer(false, nil)
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		Handler:           sh,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
		ReadHeaderTimeout: 1 * time.Minute,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
