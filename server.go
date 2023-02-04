package main

import (
	"bufio"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/handsomefox/gowarp/client"
	"github.com/handsomefox/gowarp/models"
	"github.com/handsomefox/gowarp/models/mongo"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	http.Handler
	*mongo.AccountModel
	c          *client.Configuration
	listenAddr string
}

// NewServer returns a *Server with all the required setup done.
func NewServer(ctx context.Context, addr, connStr string) (*Server, error) {
	db, err := mongo.NewAccountModel(ctx, connStr)
	if err != nil {
		return nil, ErrConnStr
	}

	config, err := GetClientConfiguration(ctx)
	if err != nil {
		return nil, ErrFetchingConfiguration
	}

	c := client.NewConfiguration()
	c.Update(config)

	server := &Server{
		AccountModel: db,
		c:            c,
		listenAddr:   addr,
	}

	server.initRoutes()

	// Start a goroutine to automatically update the config.
	go func(s *Server) {
		for {
			time.Sleep(1 * time.Hour) // update config every hour.
			if err := server.UpdateConfiguration(ctx); err != nil {
				log.Err(err).Send()
			}
		}
	}(server)

	// Start a goroutine to generate keys in the background.
	go server.Fill()

	return server, nil
}

func (s *Server) initRoutes() {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./resources/static"))))
	mux.HandleFunc("/", s.GetHomePage())
	mux.HandleFunc("/config/update", s.GetUpdateConfigPage())
	mux.HandleFunc("/key/generate", RateLimiter(s.GetGeneratedKey(), 20, 1*time.Hour))

	s.Handler = mux
}

func (s *Server) UpdateConfiguration(ctx context.Context) error {
	config, err := GetClientConfiguration(ctx)
	if err != nil {
		return ErrFetchingConfiguration
	}

	s.c.Update(config)

	return nil
}

// Fill fills (forever) the internal storage with correctly generated keys.
func (s *Server) Fill() {
	for {
		ctx := context.Background()
		if s.Len(ctx) > 250 {
			time.Sleep(10 * time.Second)
			continue
		}

		s.newKeyToDB()
		log.Info().Int64("current_key_count", s.Len(ctx))
		time.Sleep(30 * time.Second)
	}
}

// GetKey either returns a key that is already stored or creates a new one.
func (s *Server) GetKey(ctx context.Context) (*models.Account, error) {
	item, err := s.GetAny(ctx)
	if err != nil {
		c := client.NewClient(s.c)
		key, err := c.NewAccountWithLicense(ctx)
		if err != nil {
			log.Err(err).Send()
			return nil, ErrCreateKey
		}

		return key, nil
	}

	if err := s.Delete(ctx, item.ID); err != nil {
		log.Err(err).Msg("failed to remove key from the database")
	}

	log.Info().Int64("current_key_count", s.Len(ctx))
	return item, nil
}

// newKeyToDB wraps the client.NewAccountWithLicense and stores the key inside database.
func (s *Server) newKeyToDB() {
	var (
		errg       = new(errgroup.Group)
		ctx        = context.Background()
		createdKey *models.Account
	)
	errg.Go(func() error {
		c := client.NewClient(s.c)
		key, err := c.NewAccountWithLicense(ctx)
		if err != nil {
			return ErrGetKey
		}
		createdKey = key
		return nil
	})

	if err := errg.Wait(); err != nil {
		log.Err(err).Send()
		return
	}

	i, err := createdKey.RefCount.Int64()
	if err != nil {
		log.Err(err).Msg("couldn't get generated key size")
		return
	}
	if i < 1000 {
		log.Error().Int64("key_size", i).Msg("generated key was too small to use")
		return
	}

	id, err := s.Insert(ctx, createdKey)
	if err != nil {
		log.Err(err).Msg("failed to add key to the database")
	}

	log.Info().Any("id", id).Msg("added key to the database")
}

func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:              s.listenAddr,
		Handler:           s,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
		ReadHeaderTimeout: 1 * time.Minute,
	}

	return srv.ListenAndServe()
}

// GetClientConfiguration returns a new configuration from the hardcoded pastebin url.
func GetClientConfiguration(ctx context.Context) (*client.ConfigurationData, error) {
	const pastebinURL = "https://pastebin.com/raw/pwtQLBiK"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pastebinURL, http.NoBody)
	if err != nil {
		return nil, ErrFetchingConfiguration
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrFetchingConfiguration
	}
	defer res.Body.Close()

	config := &client.ConfigurationData{}

	scanner := bufio.NewScanner(res.Body)

	for scanner.Scan() {
		text := scanner.Text()
		split := strings.Split(text, "=")

		if len(split) < 2 { // it should be a key=value pair
			return nil, ErrUnexpectedBody
		}

		key, value := split[0], split[1]

		switch key {
		case "CfClientVersion":
			config.CFClientVersion = value
		case "UserAgent":
			config.UserAgent = value
		case "Host":
			config.Host = value
		case "BaseURL":
			config.BaseURL = value
		case "Keys":
			if keys := strings.Split(value, ","); len(keys) > 0 {
				config.Keys = keys
			}
		}
	}

	return config, nil
}
