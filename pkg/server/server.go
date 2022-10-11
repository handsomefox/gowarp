package server

import (
	"fmt"
	"log"
	"net/http"

	"gowarp/pkg/warp"
)

// Server is the main gowarp http.Handler.
type Server struct {
	handler   http.Handler
	templates *TemplateStorage

	w *warp.Warp
}

func NewHandler() (*Server, error) {
	// Create storage for templates
	ts, err := NewTemplateStorage()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error creating the server", err)
	}

	// Create warp instance
	wh := warp.New()

	// Create server
	server := &Server{
		w:         wh,
		templates: ts,
	}
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
		if err := s.w.UpdateConfig(r.Context()); err != nil {
			message = "failed to update config"
		}

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

		generatedKey, err := s.w.GetKey(r.Context())
		if err != nil {
			log.Println(err)
			errorWithCode(w, http.StatusInternalServerError)
			return
		}

		if err := s.templates.Key().Execute(w, generatedKey); err != nil {
			log.Println(err)
			errorWithCode(w, http.StatusInternalServerError)
		}
	}
}
