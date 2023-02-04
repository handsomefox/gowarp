package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func (s *Server) makeHTTPHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			log.Trace().Dur("handler_took", time.Since(start))
		}(time.Now())

		if err := f(w, r); err != nil {
			var ae *apiError
			if errors.As(err, &ae) {
				if err := s.writeError(w, ae); err != nil {
					log.Err(err).Send()
				}
				return
			}
			if err := s.writeError(w, &apiError{
				Err:    err.Error(),
				Status: http.StatusInternalServerError,
			}); err != nil {
				log.Err(err).Send()
			}
		}
	}
}

func (s *Server) GetHomePage() http.HandlerFunc {
	return s.makeHTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		if err := HomeTemplate.Execute(w, nil); err != nil {
			log.Err(err).Msg("failed to exec template")
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

		ctx := r.Context()

		message := "finished config update"
		log.Info().Msg("started config update")

		if err := s.UpdateConfiguration(ctx); err != nil {
			message = "failed to update config"
			log.Err(err).Msg("failed to update the config")
		}

		if err := ConfigTemplate.Execute(w, message); err != nil {
			log.Err(err).Msg("failed to exec template")
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

		ctx := r.Context()

		key, err := s.GetKey(ctx)
		if err != nil {
			log.Err(err).Msg("error getting the key")
			return ErrGetKey
		}

		if err := KeyTemplate.Execute(w, key); err != nil {
			log.Err(err).Msg("failed to exec template")
			return ErrExecTmpl
		}

		return nil
	})
}

func (s *Server) writeError(w http.ResponseWriter, e *apiError) error {
	w.WriteHeader(e.Status)
	if err := ErrorTemplate.Execute(w, e); err != nil {
		log.Err(err).Msg("failed to exec template")
		return ErrExecTmpl
	}

	return nil
}
