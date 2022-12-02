package main

import (
	"log"
	"os"
)

func main() {
	var (
		connStr = os.Getenv("DB_URI")
		port    = os.Getenv("PORT")
	)

	if connStr == "" {
		log.Fatal("no connection string provided")
	}
	if port == "" {
		log.Println("no port specified, falling back to 8080")
		port = "8080"
	}

	server, err := NewServer(":"+port, connStr, log.Default())
	if err != nil {
		log.Fatal(err)
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
