package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

type AppConfiguration struct {
	DatabaseURI string `env:"DB_URI"`
	Port        string `env:"PORT"`
}

func main() {
	ctx := context.Background()

	var c AppConfiguration
	if err := envconfig.Process(ctx, &c); err != nil {
		log.Fatal().Err(err).Send()
	}

	if c.DatabaseURI == "" {
		log.Fatal().Msg("no connection string provided")
	}
	if c.Port == "" {
		log.Info().Msg("no port specified, using fallback (8080)")
		c.Port = "8080"
	}

	s, err := NewServer(ctx, ":"+c.Port, c.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Send()
	}
}
