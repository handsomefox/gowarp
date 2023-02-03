package main

import (
	"context"
	"log"

	"github.com/sethvargo/go-envconfig"
)

type AppConfiguration struct {
	DatabaseURI string `env:"DB_URI"`
	Port        string `env:"PORT"`
}

func main() {
	var (
		ctx    = context.Background()
		logger = log.Default()
	)

	var c AppConfiguration
	if err := envconfig.Process(ctx, &c); err != nil {
		log.Fatal(err)
	}

	if c.DatabaseURI == "" {
		logger.Fatal("no connection string provided")
	}
	if c.Port == "" {
		logger.Println("no port specified, falling back to 8080")
		c.Port = "8080"
	}

	s, err := NewServer(ctx, ":"+c.Port, c.DatabaseURI, logger)
	if err != nil {
		logger.Fatal(err)
	}

	if err := s.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
}
