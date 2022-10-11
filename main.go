package main

import (
	"net/http"
	"os"
	"time"

	"gowarp/server"
)

func main() {
	sh, err := server.NewHandler()
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		Handler:           sh,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
