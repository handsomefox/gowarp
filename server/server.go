package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/handsomefox/gowarp/models/mongo"
	"github.com/handsomefox/gowarp/server/ratelimit"
	"github.com/handsomefox/gowarp/server/storage"
	"github.com/handsomefox/gowarp/warp"
	"github.com/handsomefox/gowarp/warp/cfg/pastebin"
)

// Server is the main gowarp http.Handler.
type Server struct {
	handler   http.Handler
	templates *storage.TemplateStorage
	service   *warp.Service
	storage   *storage.Storage
}

func NewHandler(useProxy bool) (*Server, error) {
	// Create storage for templates
	ts, err := storage.NewTemplateStorage()
	if err != nil {
		return nil, fmt.Errorf("error creating the server: %w", err)
	}

	// Connect to databse
	db, err := mongo.NewAccountModel(context.TODO(), os.Getenv("MONGO_URI"))
	if err != nil {
		panic(err) // nowhere to store keys :/
	}

	// Create server
	server := &Server{
		templates: ts,
		service:   warp.NewService(nil, useProxy),
		storage: &storage.Storage{
			AM: db,
		},
	}

	config, err := pastebin.GetConfig(context.Background())
	if err != nil {
		panic(err) // we probably have outdated keys anyway, no point in continuing.
	}
	server.service.UpdateConfig(config)

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

	go server.storage.Fill(server.service)

	mux := http.NewServeMux()
	// Setup routes
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./server/resources/static"))))
	mux.HandleFunc("/", server.home())
	mux.HandleFunc("/config/update", server.updateConfig())
	mux.HandleFunc("/key/generate", server.generateKey())

	// Apply ratelimiting, logging, else...
	server.handler = Decorate(mux, ratelimit.New())

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
		s.service.UpdateConfig(newConfig)

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

		key, err := s.storage.GetKey(r.Context(), s.service)
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
