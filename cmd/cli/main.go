package main

import (
	"context"
	"fmt"
	"os"

	"github.com/handsomefox/gowarp/internal/server"
	"github.com/handsomefox/gowarp/pkg/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.Logger.Level(zerolog.DebugLevel)

	ctx := context.Background()

	config := client.NewConfiguration()
	cdata, err := server.GetClientConfiguration(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("failed to load the latest config, using fallback")
	} else {
		log.Debug().Msg("using the latest config")
		config.Update(cdata)
	}

	c := client.NewClient(config)

	acc, err := c.NewAccountWithLicense(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create an account")
	}

	rc, err := acc.RefCount.Int64()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse referall count")
	}

	if rc < 1000 {
		log.Fatal().Int64("ref_count", rc).Msg("generated key is too small to use")
	}

	log.Info().Msg("Generated a new account successfully!")
	log.Info().Str("License             ", acc.License).Send()
	log.Info().Str("Referral count (GB) ", fmt.Sprint(rc)).Send()
	log.Info().Str("License Type        ", acc.Type).Send()
}
