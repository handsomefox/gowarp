package server

import (
	"context"
	"errors"
	"html/template"
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
	config     *client.Configuration
	client     *client.Client
	db         *mongo.AccountModel
	handler    http.Handler
	templates  map[templates.TemplateID]*template.Template
	listenAddr string
}

// New returns a *Server with all the required setup done.
func New(ctx context.Context, addr, connStr, dbname, colname string, tmpl map[templates.TemplateID]*template.Template) (*Server, error) {
	db, err := mongo.NewAccountModel(ctx, connStr, dbname, colname)
	if err != nil {
		return nil, ErrConnStr
	}

	config, err := client.GetConfiguration(ctx, pastebinURL)
	if err != nil {
		return nil, ErrFetchingConfiguration
	}

	c := client.NewConfiguration()
	c.Update(config)

	server := &Server{
		db:         db,
		config:     c,
		client:     client.NewClient(c, true),
		listenAddr: addr,
		templates:  tmpl,
	}

	server.initRoutes()

	// Start a goroutine to automatically update the config.
	go func(s *Server) {
		for {
			time.Sleep(1 * time.Hour) // update config every hour.
			if err := s.UpdateConfiguration(ctx); err != nil {
				log.Err(err).Send()
			}
		}
	}(server)

	// Start a goroutine to generate keys in the background.
	go server.Fill(ctx, 200, 20*time.Minute)

	return server, nil
}

func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:              s.listenAddr,
		Handler:           s.handler,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
		ReadHeaderTimeout: 1 * time.Minute,
	}

	return srv.ListenAndServe()
}

func (s *Server) initRoutes() {
	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("./assets/static"))))
	r.Get("/", s.HandleHomePage())
	r.Get("/config/update", s.HandleUpdateConfig())
	r.HandleFunc("/key/generate", ratelimiter.New(s.HandleGenerateKey(), 20, 1*time.Hour))

	s.handler = r
}

func (s *Server) UpdateConfiguration(ctx context.Context) error {
	config, err := client.GetConfiguration(ctx, pastebinURL)
	if err != nil {
		return ErrFetchingConfiguration
	}
	s.config.Update(config)
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
