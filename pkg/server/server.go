package server

import (
	"html/template"
	"log"
	"net/http"

	"gowarp/pkg/warp"
)

type Server struct {
	warpHandle *warp.Warp
	mux        *http.ServeMux

	rateLimiter http.Handler

	homeTmpl   *template.Template
	configTmpl *template.Template
	keyTmpl    *template.Template
}

func New() *Server {
	server := &Server{
		warpHandle: warp.New(),
		mux:        http.NewServeMux(),
	}

	server.setupTemplates()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	server.mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	server.mux.HandleFunc("/", server.home())
	server.mux.HandleFunc("/key/generate", server.generateKey())
	server.mux.HandleFunc("/config/update", server.updateConfig())

	server.rateLimiter = newRateLimiter().Limits(server.mux)

	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.rateLimiter.ServeHTTP(w, r)
}

func (s *Server) setupTemplates() {
	ts, err := template.ParseFiles([]string{
		"./ui/html/home.page.tmpl.html",
		"./ui/html/base.layout.tmpl.html",
		"./ui/html/footer.partial.tmpl.html",
	}...)
	if err != nil {
		panic(err)
	}

	s.homeTmpl = ts

	ts, err = template.ParseFiles([]string{
		"./ui/html/config.page.tmpl.html",
		"./ui/html/base.layout.tmpl.html",
		"./ui/html/footer.partial.tmpl.html",
	}...)
	if err != nil {
		panic(err)
	}

	s.configTmpl = ts

	ts, err = template.ParseFiles([]string{
		"./ui/html/key.page.tmpl.html",
		"./ui/html/base.layout.tmpl.html",
		"./ui/html/footer.partial.tmpl.html",
	}...)
	if err != nil {
		panic(err)
	}

	s.keyTmpl = ts
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

		if err := s.homeTmpl.Execute(w, nil); err != nil {
			log.Println(err)
			errorWithCode(w, http.StatusInternalServerError)
		}
	}
}

func (s *Server) updateConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		message := "finished config update"
		if err := s.warpHandle.UpdateConfig(); err != nil {
			message = "failed to update config"
		}

		if err := s.configTmpl.Execute(w, message); err != nil {
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

		generatedKey, err := s.warpHandle.GetKey()
		if err != nil {
			log.Println(err)
			errorWithCode(w, http.StatusInternalServerError)
			return
		}

		if err := s.keyTmpl.Execute(w, generatedKey); err != nil {
			log.Println(err)
			errorWithCode(w, http.StatusInternalServerError)
		}
	}
}
