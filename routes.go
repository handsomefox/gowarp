package main

import (
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
)

type APIError struct {
	Err    string
	Status int
}

func (e *APIError) Error() string {
	return e.Err
}

var ErrExecTmpl = &APIError{Err: "failed to exec tmpl", Status: http.StatusInternalServerError}

func (s *Server) HandleHomePage() http.HandlerFunc {
	return s.WrapHandlerFuncErr(func(w http.ResponseWriter, _ *http.Request) error {
		if err := s.templates[HomeID].Execute(w, nil); err != nil {
			log.Err(err).Msg("failed to exec template")
			return ErrExecTmpl
		}
		return nil
	})
}

func (s *Server) HandleUpdateConfig() http.HandlerFunc {
	return s.WrapHandlerFuncErr(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		log.Info().Msg("started config update")
		if err := s.UpdateConfiguration(ctx); err != nil {
			log.Err(err).Msg("failed to update config")
			return err
		}

		if err := s.templates[ConfigID].Execute(w, "finished config update"); err != nil {
			log.Err(err).Msg("failed to exec template")
			return ErrExecTmpl
		}

		return nil
	})
}

func (s *Server) HandleGenerateKey() http.HandlerFunc {
	return s.WrapHandlerFuncErr(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		key, err := s.GetKey(ctx)
		if err != nil {
			log.Err(err).Msg("error getting the key")
			return ErrGetKey
		}

		if err := s.templates[KeyID].Execute(w, key); err != nil {
			log.Err(err).Msg("failed to exec template")
			return ErrExecTmpl
		}

		return nil
	})
}

type HandlerFuncErr func(w http.ResponseWriter, r *http.Request) error

func (s *Server) WrapHandlerFuncErr(f HandlerFuncErr) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			var ae *APIError
			if errors.As(err, &ae) {
				if err := s.WriteErr(w, ae); err != nil {
					log.Err(err).Send()
				}
				return
			}
			if err := s.WriteErr(w, &APIError{
				Err:    err.Error(),
				Status: http.StatusInternalServerError,
			}); err != nil {
				log.Err(err).Send()
			}
		}
	}
}

func (s *Server) WriteErr(w http.ResponseWriter, e *APIError) error {
	w.WriteHeader(e.Status)
	if err := s.templates[ErrorID].Execute(w, e); err != nil {
		log.Err(err).Msg("failed to exec template")
		return ErrExecTmpl
	}

	return nil
}
