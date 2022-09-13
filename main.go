package main

import (
	"log"
	"net/http"
	"os"

	"gowarp/pkg/server"
)

func main() {
	s := server.New()
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), s.Limiter(s)))
}
