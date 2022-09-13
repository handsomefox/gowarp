package server

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"gowarp/pkg/warp"
)

const (
	requestLimit  = 500
	requestPeriod = time.Hour * 6
)

type Server struct {
	mux *http.ServeMux

	warpHandle     *warp.Warp
	requestCounter *IPRequestCounter

	homeTmpl   *template.Template
	configTmpl *template.Template
	keyTmpl    *template.Template
}

func New() *Server {
	server := &Server{
		mux:        http.NewServeMux(),
		warpHandle: warp.New(),
		requestCounter: &IPRequestCounter{
			ips: make(map[string]int),
		},
	}

	server.initTemplates()
	server.setupRoutes()

	go server.clearBlockedIPs()

	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) setupRoutes() {
	fileServer := http.FileServer(http.Dir("./ui/static"))

	s.mux.HandleFunc("/", s.home())
	s.mux.HandleFunc("/key/generate", s.generateKey())
	s.mux.HandleFunc("/config/update", s.updateConfig())
	s.mux.Handle("/static/", http.StripPrefix("/static", fileServer))
}

func (s *Server) initTemplates() {
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

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")

	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}

	if IPAddress == "" {
		IPAddress = strings.Split(r.RemoteAddr, ":")[0]
	}

	return IPAddress
}

func (s *Server) Limiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipAddr := readUserIP(r)

		s.requestCounter.Inc(ipAddr)

		cv := s.requestCounter.Value(ipAddr)

		// log.Println("Counter: " + strconv.Itoa(cv) + ", IP: " + ipAddr)

		if cv >= requestLimit {
			errorWithCode(w, http.StatusTooManyRequests)

			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) clearBlockedIPs() {
	for {
		s.requestCounter.mu.Lock()
		s.requestCounter.ips = make(map[string]int)
		s.requestCounter.mu.Unlock()
		time.Sleep(requestPeriod)
	}
}
