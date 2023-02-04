package main

import (
	"errors"
	"net/http"
)

type APIError struct {
	Err    string
	Status int
}

func (e *APIError) Error() string {
	return e.Err
}

var (
	ErrExecTmpl              = &APIError{Err: "failed to exec tmpl", Status: http.StatusInternalServerError}
	ErrMethodNotAllowed      = &APIError{Err: "method not allowed", Status: http.StatusMethodNotAllowed}
	ErrGetKey                = errors.New("server: failed to get the key")
	ErrConnStr               = errors.New("server: invalid connection string")
	ErrFetchingConfiguration = errors.New("server: error fetching configuration")
	ErrCreateKey             = errors.New("server: failed to create a key on the fly")
	ErrUnexpectedBody        = errors.New("server: unexpected configuration response body")
)
