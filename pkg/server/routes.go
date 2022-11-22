package server

import (
	"log"
	"net/http"
	"time"

	"github.com/handsomefox/gowarp/pkg/server/middleware"
	"github.com/handsomefox/gowarp/pkg/warp/pastebin"
)

func (s *Server) SetupRoutes() {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./resources/static"))))
	mux.HandleFunc("/", s.home())
	mux.HandleFunc("/config/update", s.updateConfig())
	mux.HandleFunc("/key/generate", s.generateKey())

	// Apply ratelimiting, logging, else...
	s.handler = middleware.Decorate(
		mux,
		middleware.RateLimiter(500, time.Hour*6),
		middleware.RequestTimer(),
	)
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
		log.Println("Getting key took: ", time.Since(start))
	}
}

func errorWithCode(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
