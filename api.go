package main

import (
	"errors"
	"html/template"
	"net/http"
	"time"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error

var homeTmpl = template.Must(template.ParseFiles(
	"./resources/html/home.page.tmpl.html",
	"./resources/html/base.layout.tmpl.html",
	"./resources/html/footer.partial.tmpl.html",
))

var errTmpl = template.Must(template.ParseFiles(
	"./resources/html/err.page.tmpl.html",
	"./resources/html/base.layout.tmpl.html",
	"./resources/html/footer.partial.tmpl.html",
))

var configTmpl = template.Must(template.ParseFiles(
	"./resources/html/config.page.tmpl.html",
	"./resources/html/base.layout.tmpl.html",
	"./resources/html/footer.partial.tmpl.html",
))

var keyTmpl = template.Must(template.ParseFiles(
	"./resources/html/key.page.tmpl.html",
	"./resources/html/base.layout.tmpl.html",
	"./resources/html/footer.partial.tmpl.html",
))

func (s *Server) makeHTTPHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			s.Println("Handler took: ", time.Since(start))
		}(time.Now())

		if err := f(w, r); err != nil {
			var ae *apiError
			if errors.As(err, &ae) {
				if err := s.writeError(w, ae); err != nil {
					s.Println(err)
				}
				return
			}
			if err := s.writeError(w, &apiError{
				Err:    err.Error(),
				Status: http.StatusInternalServerError,
			}); err != nil {
				s.Println(err)
			}
		}
	}
}

func (s *Server) GetHomePage() http.HandlerFunc {
	return s.makeHTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		if err := homeTmpl.Execute(w, nil); err != nil {
			s.Println("Failed to execute template: ", err)
			return ErrExecTmpl
		}

		return nil
	})
}

func (s *Server) GetUpdateConfigPage() http.HandlerFunc {
	return s.makeHTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		message := "finished config update"
		s.Println("Updating the config")

		if err := s.UpdateConfiguration(r.Context()); err != nil {
			message = "failed to update config"
			s.Println("Failed to update config: ", err)
		}

		if err := configTmpl.Execute(w, message); err != nil {
			s.Println("Failed to execute template: ", err)
			return ErrExecTmpl
		}

		return nil
	})
}

func (s *Server) GetGeneratedKey() http.HandlerFunc {
	return s.makeHTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		key, err := s.GetKey(r.Context())
		if err != nil {
			s.Println("Error when getting key: ", err)
			return ErrGetKey
		}

		if err := keyTmpl.Execute(w, key); err != nil {
			s.Println("Failed to execute template: ", err)
			return ErrExecTmpl
		}

		return nil
	})
}

func (s *Server) writeError(w http.ResponseWriter, e *apiError) error {
	w.WriteHeader(e.Status)
	if err := errTmpl.Execute(w, e); err != nil {
		s.Println("Failed to execute template: ", err)
		return ErrExecTmpl
	}

	return nil
}
