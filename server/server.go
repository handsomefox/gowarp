package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/handsomefox/gowarp/client"
	"github.com/handsomefox/gowarp/client/cfg/pastebin"
	"github.com/handsomefox/gowarp/client/cfg/serdar"
	"github.com/handsomefox/gowarp/storage"
)

// Server is the main gowarp http.Handler.
type Server struct {
	handler   http.Handler
	templates *TemplateStorage

	client  *client.Client
	storage *storage.Storage
}

func NewHandler() (*Server, error) {
	// Create storage for templates
	ts, err := NewTemplateStorage()
	if err != nil {
		return nil, fmt.Errorf("error creating the server: %w", err)
	}

	// Create server
	server := &Server{
		templates: ts,
		client:    client.NewClient(nil),
		storage:   storage.NewStorage(),
	}

	go server.storage.Fill(server.client)
	go func() {
		for {
			config, err := serdar.GetConfig(context.Background())
			if err != nil {
				time.Sleep(1 * time.Minute)
			}
			server.client.UpdateConfig(config)

			time.Sleep(1 * time.Hour)
		}
	}()

	mux := http.NewServeMux()
	// Setup routes
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./ui/static"))))
	mux.HandleFunc("/", server.home())
	mux.HandleFunc("/config/update", server.updateConfig())
	mux.HandleFunc("/key/generate", server.generateKey())

	// Apply ratelimiting, logging, else...
	server.handler = Decorate(mux, NewRateLimiter())

	return server, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func errorWithCode(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (s *Server) home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if err := s.templates.Home().Execute(w, nil); err != nil {
			log.Println(err)
			errorWithCode(w, http.StatusInternalServerError)
		}
	}
}

func (s *Server) updateConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		message := "finished config update"

		newConfig, err := pastebin.GetConfig(r.Context())
		if err != nil {
			message = "failed to update config"
		}
		log.Println(newConfig)
		s.client.UpdateConfig(newConfig)

		if err := s.templates.Config().Execute(w, message); err != nil {
			log.Println(err)
			errorWithCode(w, http.StatusInternalServerError)
		}
	}
}

func (s *Server) generateKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errorWithCode(w, http.StatusMethodNotAllowed)
			return
		}

		key, err := s.storage.GetKey(r.Context(), s.client)
		if err != nil {
			log.Println(err)
			errorWithCode(w, http.StatusInternalServerError)
			return
		}

		if err := s.templates.Key().Execute(w, key); err != nil {
			log.Println(err)
			errorWithCode(w, http.StatusInternalServerError)
		}
	}
}
