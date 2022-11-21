package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/handsomefox/gowarp/pkg/models/mongo"
	"github.com/handsomefox/gowarp/pkg/server/ratelimit"
	"github.com/handsomefox/gowarp/pkg/server/storage"
	"github.com/handsomefox/gowarp/pkg/warp"
	"github.com/handsomefox/gowarp/pkg/warp/pastebin"
)

// Server is the main gowarp http.Handler.
type Server struct {
	handler   http.Handler
	templates *storage.TemplateStorage
	service   *warp.Service
	storage   *storage.Storage
}

// NewServer returns a *Server with all the required setup done.
// If mongoURI is nil, it uses the env variable MONGO_URI.
func NewServer(useProxy bool, mongoURI *string) (*Server, error) {
	var uri string
	if mongoURI != nil {
		uri = *mongoURI
	} else {
		uri = os.Getenv("MONGO_URI")
	}

	// Connect to database
	db, err := mongo.NewAccountModel(context.TODO(), uri)
	if err != nil {
		panic(err) // nowhere to store keys :/
	}

	log.Println("Connected to the database")

	// Create storage for templates
	ts, err := storage.NewTemplateStorage()
	if err != nil {
		return nil, fmt.Errorf("error creating the server: %w", err)
	}

	// Create server
	service := warp.NewService(nil, useProxy)
	config, err := pastebin.GetConfig(context.Background())
	if err != nil {
		panic(err) // we probably have outdated keys anyway, no point in continuing.
	}
	service.UpdateConfig(config)

	store := &storage.Storage{AM: db}

	server := &Server{
		templates: ts,
		service:   service,
		storage:   store,
	}

	// Start a goroutine to automatically update the config.
	go func() {
		for {
			time.Sleep(1 * time.Hour) // update config every hour.

			config, err := pastebin.GetConfig(context.Background())
			if err != nil {
				continue
			}
			server.service.UpdateConfig(config)
		}
	}()

	// Start a goroutine to generate keys in the background.
	go server.storage.Fill(server.service)

	// Setup routes
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./resources/static"))))
	mux.HandleFunc("/", server.home())
	mux.HandleFunc("/config/update", server.updateConfig())
	mux.HandleFunc("/key/generate", server.generateKey())

	// Apply ratelimiting, logging, else...
	server.handler = Decorate(mux, ratelimit.New().Decorate, timerMiddleware())

	return server, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// home return a HandlerFunc for the home page.
func (s *Server) home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if err := s.templates.Home().Execute(w, nil); err != nil {
			log.Println("Failed to execute template: ", err)
			errorWithCode(w, http.StatusInternalServerError)
		}
	}
}

// updateConfig returns a HandlerFunc for the config update address.
func (s *Server) updateConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		message := "finished config update"
		log.Println("Updating the config")

		newConfig, err := pastebin.GetConfig(r.Context())
		if err != nil {
			message = "failed to update config"
			log.Println("Failed to update config: ", err)
		}
		s.service.UpdateConfig(newConfig)
		log.Println("Using new config: ", newConfig)

		if err := s.templates.Config().Execute(w, message); err != nil {
			log.Println("Failed to execute template: ", err)
			errorWithCode(w, http.StatusInternalServerError)
		}
	}
}

// generateKey returns a HandlerFunc for the generated key page.
func (s *Server) generateKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorWithCode(w, http.StatusMethodNotAllowed)
			return
		}

		log.Println("Getting key")
		start := time.Now()

		key, err := s.storage.GetKey(r.Context(), s.service)
		if err != nil {
			log.Println("Error when getting key: ", err)
			errorWithCode(w, http.StatusInternalServerError)
			return
		}

		if err := s.templates.Key().Execute(w, key); err != nil {
			log.Println("Failed to execute template: ", err)
			errorWithCode(w, http.StatusInternalServerError)
		}
		log.Println("Getting key took: ", time.Since(start).Milliseconds(), "ms")
	}
}

func errorWithCode(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
