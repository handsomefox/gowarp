package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/handsomefox/gowarp/cmd/http/server/ratelimiter"
	"github.com/handsomefox/gowarp/cmd/http/server/templates"
	"github.com/handsomefox/gowarp/internal/models"
	"github.com/handsomefox/gowarp/internal/models/mongo"
	"github.com/handsomefox/gowarp/pkg/client"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const pastebinURL = "https://pastebin.com/raw/pwtQLBiK"

var (
	ErrGetKey                = errors.New("server: failed to get the key")
	ErrConnStr               = errors.New("server: invalid connection string")
	ErrFetchingConfiguration = errors.New("server: error fetching configuration")
	ErrCreateKey             = errors.New("server: failed to create a key on the fly")
	ErrUnexpectedBody        = errors.New("server: unexpected configuration response body")
)

type Server struct {
	client *client.Client

	db *mongo.AccountModel

	mux *chi.Mux

	tmpls templates.Map
}

type DBParams struct {
	DBConnString string
	DBName       string
	DBCollName   string
}

// New returns a *Server with all the required setup done.
func New(ctx context.Context, dbParams DBParams, tmpls templates.Map) (*Server, error) {
	// Connect to the database
	db, err := mongo.NewAccountModel(ctx, dbParams.DBConnString, dbParams.DBName, dbParams.DBCollName)
	if err != nil {
		return nil, ErrConnStr
	}

	// Create the server
	server := &Server{
		client: client.NewClient(client.NewConfiguration(), true),
		db:     db,
		tmpls:  tmpls,
	}

	// Get and set the most recent configuration
	config, err := client.GetConfiguration(ctx, pastebinURL)
	if err != nil {
		return nil, ErrFetchingConfiguration
	}
	server.client.UpdateConfig(config)

	// Setup routing
	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Handle("/static/*",
		http.StripPrefix("/static", http.FileServer(http.Dir("./assets/static"))))

	r.Get("/",
		server.HandleHomePage())

	r.Get("/config/update",
		server.HandleUpdateConfig())

	r.HandleFunc("/key/generate",
		ratelimiter.New(server.HandleGenerateKey(), 20, 1*time.Hour))

	server.mux = r

	// Start a goroutine to periodically update the config.
	go func(s *Server) {
		for {
			time.Sleep(1 * time.Hour) // update config every hour.
			if err := s.UpdateConfiguration(ctx); err != nil {
				log.Err(err).Send()
			}
		}
	}(server)

	// Start a goroutine to generate keys in the background if necessary.
	go server.Fill(ctx, 200, 20*time.Minute)

	return server, nil
}

// ListenAndServe is a wrapper around (*http.Server).ListenAndServe().
func (s *Server) ListenAndServe(listenAddr string) error {
	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           s.mux,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
		ReadHeaderTimeout: 1 * time.Minute,
	}

	return srv.ListenAndServe()
}

// UpdateConfiguration fetches the most recent configuration and sets it.
func (s *Server) UpdateConfiguration(ctx context.Context) error {
	config, err := client.GetConfiguration(ctx, pastebinURL)
	if err != nil {
		return ErrFetchingConfiguration
	}
	s.client.UpdateConfig(config)
	return nil
}

// Fill fills the db to the maxCount.
func (s *Server) Fill(ctx context.Context, maxCount int64, sleepDuration time.Duration) {
	for {
		if s.db.Len(ctx) >= maxCount {
			time.Sleep(sleepDuration)
		}
		s.pushNewKeyToDatabase(ctx)
		log.Info().Int64("current_key_count", s.db.Len(ctx))
		time.Sleep(30 * time.Second)
	}
}

// GetKey either returns a key that is already stored or creates a new one.
func (s *Server) GetKey(ctx context.Context) (*models.Account, error) {
	item, err := s.db.GetAny(ctx)
	if err != nil {
		key, err := s.client.NewAccountWithLicense(ctx)
		if err != nil {
			log.Err(err).Send()
			return nil, ErrCreateKey
		}

		return key, nil
	}

	if err := s.db.Delete(ctx, item.ID); err != nil {
		log.Err(err).Msg("failed to remove key from the database")
	}

	log.Info().Int64("current_key_count", s.db.Len(ctx))
	return item, nil
}

// pushNewKeyToDatabase wraps the client.NewAccountWithLicense and stores the key inside database.
func (s *Server) pushNewKeyToDatabase(ctx context.Context) {
	var (
		errg       = new(errgroup.Group)
		createdKey *models.Account
	)
	errg.Go(func() error {
		key, err := s.client.NewAccountWithLicense(ctx)
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

	id, err := s.db.Insert(ctx, createdKey)
	if err != nil {
		log.Err(err).Msg("failed to add key to the database")
	}

	log.Info().Any("id", id).Msg("added key to the database")
}
