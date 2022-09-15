package main

import (
	"log"
	"net/http"
	"os"

	"gowarp/pkg/server"
)

func main() {
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), server.New()))
}
