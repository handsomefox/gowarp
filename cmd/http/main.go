package main

import (
	"context"

	"github.com/handsomefox/gowarp/cmd/http/server"
	"github.com/handsomefox/gowarp/cmd/http/server/templates"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

type AppConfiguration struct {
	DatabaseURI    string `env:"DB_URI"`
	Port           string `env:"PORT"`
	DatabaseName   string `env:"DATABASE_NAME"`
	CollectionName string `env:"COLLECTION_NAME"`
}

func main() {
	log.Logger = log.Logger.Level(zerolog.DebugLevel)

	if err := godotenv.Load(); err != nil {
		log.Err(err).Msg("failed to load .env file")
	}

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

	tmpls, err := templates.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load templates")
	}

	dbParams := server.DBParams{
		DBConnString: c.DatabaseURI,
		DBName:       c.DatabaseName,
		DBCollName:   c.CollectionName,
	}

	s, err := server.New(ctx, dbParams, tmpls)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Info().Str("addr", "localhost").Str("port", c.Port).Msg("server started on http://localhost:" + c.Port)
	if err := s.ListenAndServe(":" + c.Port); err != nil {
		log.Fatal().Err(err).Send()
	}
}
